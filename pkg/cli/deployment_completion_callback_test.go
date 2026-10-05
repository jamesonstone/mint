package cli

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

func TestAdapterCompletionReconcilesOnlyVerifiedDeployment(t *testing.T) {
	for _, scenario := range []string{"deployed", "manual", "unknown", "unchanged-failure", "publish", "publish-revoked", "publication-failure", "publication-pending", "wrong-branch", "untrusted-event", "unknown-intent", "wrong-path", "observer", "observer-failure", "foreign-observation", "combined", "failure-replay", "rollback"} {
		t.Run(scenario, func(t *testing.T) {
			policy := promotion.Policy{Schema: 2, Mode: "deployment", Repository: "owner/repo", DefaultBranch: "main", HumanLogin: "human", ControlWorkflow: ".github/workflows/control.yaml", Environments: map[string]promotion.EnvironmentPolicy{"stage": {Scope: "shared", Deploy: "reviewed", Follow: "latest", Configuration: "sha256:" + strings.Repeat("a", 64), PromotionWorkflow: ".github/workflows/deploy.yaml", ObservationWorkflow: ".github/workflows/observe.yaml"}}}
			env := policy.Environments["stage"]
			if scenario == "manual" {
				env.Deploy = "manual"
			}
			if scenario == "publish" || scenario == "publication-failure" || scenario == "publication-pending" {
				env.Publish = true
			}
			if scenario == "combined" || scenario == "failure-replay" {
				env.ObservationWorkflow = env.PromotionWorkflow
			}
			policy.Environments["stage"] = env
			state := promotion.NewState(policy.Repository, "stage")
			state.Schema = 2
			intent := promotion.Intent{ID: "request", Status: "deploying", DeploymentRunID: 17}
			state.Intents[intent.ID] = intent
			state.InFlight = intent.ID
			if scenario == "unknown-intent" {
				state.Intents = map[string]promotion.Intent{}
			}
			source := promotion.WorkflowRun{ID: 17, Path: env.PromotionWorkflow, Event: "workflow_dispatch", HeadBranch: "main", Status: "completed", Conclusion: "success"}
			source.Repository.FullName = policy.Repository
			source.HeadRepository.FullName = policy.Repository
			if scenario == "wrong-branch" {
				source.HeadBranch = "feature"
			}
			if scenario == "untrusted-event" {
				source.Event = "pull_request"
			}
			if scenario == "wrong-path" {
				source.Path = ".github/workflows/untrusted.yaml"
			}
			if strings.Contains(scenario, "observer") || scenario == "foreign-observation" {
				source.Path = env.ObservationWorkflow
			}
			if scenario == "observer-failure" {
				source.Conclusion = "failure"
			}
			observedEnv := "stage"
			if scenario == "foreign-observation" {
				observedEnv = "other"
			}
			archive, digest := completionObservationArchive(t, promotion.Observation{Environment: observedEnv})
			controller := promotion.WorkflowRun{ID: 99, Path: policy.ControlWorkflow, Event: "workflow_run", HeadBranch: "main", Status: "in_progress"}
			controller.Repository.FullName = policy.Repository
			controller.HeadRepository.FullName = policy.Repository
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("unexpected mutation %s %s", r.Method, r.URL.Path)
					w.WriteHeader(500)
					return
				}
				switch {
				case strings.HasSuffix(r.URL.Path, "/runs/17"):
					_ = json.NewEncoder(w).Encode(source)
				case strings.HasSuffix(r.URL.Path, "/runs/99"):
					_ = json.NewEncoder(w).Encode(controller)
				case strings.Contains(r.URL.Path, "/git/ref/"):
					_ = json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": strings.Repeat("b", 40)}})
				case strings.Contains(r.URL.Path, "/git/commits/"):
					_ = json.NewEncoder(w).Encode(map[string]any{"tree": map[string]string{"sha": strings.Repeat("c", 40)}})
				case strings.Contains(r.URL.Path, "/git/trees/"):
					_ = json.NewEncoder(w).Encode(map[string]any{"tree": []any{map[string]string{"path": "release-state.json", "type": "blob", "sha": strings.Repeat("d", 40)}}})
				case strings.Contains(r.URL.Path, "/git/blobs/"):
					data, _ := json.Marshal(state)
					_ = json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "content": base64.StdEncoding.EncodeToString(data)})
				case strings.HasSuffix(r.URL.Path, "/artifacts"):
					_ = json.NewEncoder(w).Encode(map[string]any{"total_count": 1, "artifacts": []any{map[string]any{"id": 9, "name": "mint-observation", "digest": digest}}})
				case strings.HasSuffix(r.URL.Path, "/zip"):
					_, _ = w.Write(archive)
				default:
					t.Error("unexpected request", r.URL.Path)
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			t.Setenv("MINT_CALLBACK_TEST_TOKEN", "test")
			var out bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetContext(t.Context())
			cmd.SetOut(&out)
			f := productionFlags{APIURL: server.URL, TokenEnv: "MINT_CALLBACK_TEST_TOKEN", Event: "workflow_run", RunID: 99, Output: ""}
			operations := []string{}
			execute := func(cmd *cobra.Command, op string, child productionFlags) error {
				operations = append(operations, op)
				if child.Environment != "stage" || child.Output != "" {
					t.Fatal("child identity/output mismatch", child)
				}
				if op == "propose" {
					if child.RunID != 99 || child.Event != "workflow_run" || child.Kind != "normal" || child.IntentID != "" {
						t.Fatal("queued reconciliation lost controller identity", child)
					}
					return nil
				}
				if child.RunID != 17 {
					t.Fatal("completion did not use source run", child)
				}
				if op == "observe" {
					return nil
				}
				if child.IntentID != intent.ID {
					t.Fatal("wrong intent", child)
				}
				if op == "finish" {
					if scenario == "unknown" {
						return fmt.Errorf("unverified runtime")
					}
					current := state.Intents[intent.ID]
					current.Status = "deployed"
					state.InFlight = ""
					if scenario == "rollback" {
						state.Paused = true
					}
					if scenario == "unchanged-failure" || scenario == "failure-replay" {
						current.Status = "deployment_failed"
					}
					if strings.Contains(scenario, "publish") || strings.Contains(scenario, "publication") {
						current.Status = "publication_pending"
					}
					state.Intents[intent.ID] = current
					return nil
				}
				if op == "publish" {
					if scenario == "publication-failure" {
						return fmt.Errorf("publication rejected")
					}
					if scenario != "publication-pending" {
						current := state.Intents[intent.ID]
						current.Status = "deployed"
						state.Intents[intent.ID] = current
					}
					return nil
				}
				t.Fatal("unexpected operation", op)
				return nil
			}
			// Forged event fields cannot override the authoritative source path/state.
			handled, err := runAdapterCallbackWith(cmd, policy, f, []byte(`{"workflow_run":{"id":17,"conclusion":"success","path":".github/workflows/deploy.yaml"}}`), execute)
			if scenario == "combined" || scenario == "failure-replay" {
				replayHandled, replayErr := runAdapterCallbackWith(cmd, policy, f, []byte(`{"workflow_run":{"id":17}}`), execute)
				if !replayHandled || replayErr != nil {
					t.Fatal("terminal deployment replay became an observer", replayHandled, replayErr)
				}
			}
			want := map[string]string{"deployed": "finish,propose", "manual": "finish", "unknown": "finish", "unchanged-failure": "finish", "publish": "finish,publish,propose", "publish-revoked": "finish", "publication-failure": "finish,publish", "publication-pending": "finish,publish", "observer": "observe", "combined": "finish,propose,propose", "failure-replay": "finish", "rollback": "finish"}[scenario]
			if strings.Join(operations, ",") != want {
				t.Fatalf("operations=%v want=%s", operations, want)
			}
			wantError := scenario == "unknown" || scenario == "publish-revoked" || scenario == "publication-failure" || scenario == "publication-pending" || scenario == "wrong-branch" || scenario == "untrusted-event" || scenario == "observer-failure" || scenario == "foreign-observation"
			if (err != nil) != wantError {
				t.Fatalf("unexpected error %v", err)
			}
			if handled != (scenario != "wrong-path") {
				t.Fatal("routing mismatch", handled)
			}
			if scenario == "unknown" && state.InFlight != intent.ID {
				t.Fatal("unknown runtime lost fence")
			}
			_, _ = fmt.Fprint(cmd.OutOrStdout(), "restored")
			if !strings.HasSuffix(out.String(), "restored") {
				t.Fatal("command output was not restored")
			}
		})
	}
}

func completionObservationArchive(t *testing.T, observation promotion.Observation) ([]byte, string) {
	t.Helper()
	var data bytes.Buffer
	writer := zip.NewWriter(&data)
	file, err := writer.Create("mint-observation.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(file).Encode(observation); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return data.Bytes(), fmt.Sprintf("sha256:%x", sha256.Sum256(data.Bytes()))
}
