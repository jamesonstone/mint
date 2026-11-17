package cli

import (
	"bytes"
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

func TestAdapterCompletionNoticeSettlesOnlyAnEndedAdapterRun(t *testing.T) {
	for _, scenario := range []string{"settles", "string-id", "waits-then-settles", "still-active", "not-adapter", "foreign-run", "wrong-type", "no-run", "negative-run", "wrong-controller-event", "unverified"} {
		t.Run(scenario, func(t *testing.T) {
			policy := promotion.Policy{Schema: 2, Mode: "deployment", Repository: "owner/repo", DefaultBranch: "main", HumanLogin: "human", ControlWorkflow: ".github/workflows/control.yaml", Environments: map[string]promotion.EnvironmentPolicy{"stage": {Scope: "shared", Deploy: "automatic", Follow: "latest", Configuration: "sha256:" + strings.Repeat("a", 64), PromotionWorkflow: ".github/workflows/deploy.yaml"}}}
			state := promotion.NewState(policy.Repository, "stage")
			state.Schema = 2
			intent := promotion.Intent{ID: "frozen", Status: "deploying", DeploymentRunID: 17}
			state.Intents[intent.ID] = intent
			state.InFlight = intent.ID
			source := promotion.WorkflowRun{ID: 17, Path: ".github/workflows/deploy.yaml", Event: "workflow_dispatch", HeadBranch: "main", Status: "completed", Conclusion: "success"}
			source.Repository.FullName = policy.Repository
			source.HeadRepository.FullName = policy.Repository
			if scenario == "not-adapter" {
				source.Path = ".github/workflows/ci.yaml"
			}
			if scenario == "foreign-run" {
				source.Repository.FullName = "other/repo"
			}
			pending := 0
			if scenario == "waits-then-settles" {
				pending = 2
			}
			if scenario == "still-active" {
				pending = 100
			}
			controller := promotion.WorkflowRun{ID: 99, Path: policy.ControlWorkflow, Event: "repository_dispatch", HeadBranch: "main", Status: "in_progress"}
			controller.Repository.FullName = policy.Repository
			controller.HeadRepository.FullName = policy.Repository
			if scenario == "wrong-controller-event" {
				controller.Event = "workflow_dispatch"
			}
			polls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("unexpected mutation %s %s", r.Method, r.URL.Path)
					w.WriteHeader(500)
					return
				}
				switch {
				case strings.HasSuffix(r.URL.Path, "/runs/17"):
					polls++
					run := source
					if polls <= pending {
						run.Status, run.Conclusion = "in_progress", ""
					}
					_ = json.NewEncoder(w).Encode(run)
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
				default:
					t.Error("unexpected request", r.URL.Path)
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			t.Setenv("MINT_NOTICE_TEST_TOKEN", "test")
			var out bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetContext(t.Context())
			cmd.SetOut(&out)
			f := productionFlags{APIURL: server.URL, TokenEnv: "MINT_NOTICE_TEST_TOKEN", Event: "repository_dispatch", RunID: 99}
			operations := []string{}
			execute := func(cmd *cobra.Command, op string, child productionFlags) error {
				operations = append(operations, op)
				if op == "finish" {
					if child.RunID != 17 || child.IntentID != "frozen" {
						t.Fatal("finish lost the adapter run", child)
					}
					if scenario == "unverified" {
						return fmt.Errorf("unverified runtime")
					}
					current := state.Intents["frozen"]
					current.Status = "deployed"
					state.Intents["frozen"] = current
					state.InFlight = ""
					return nil
				}
				if op == "propose" && (child.RunID != 99 || child.Event != "repository_dispatch") {
					t.Fatal("reconciliation lost the controller identity", child)
				}
				return nil
			}
			payload := `{"action":"mint-adapter-completed","client_payload":{"run_id":17}}`
			switch scenario {
			case "string-id":
				payload = `{"action":"mint-adapter-completed","client_payload":{"run_id":"17"}}`
			case "wrong-type":
				payload = `{"action":"deploy-now","client_payload":{"run_id":17}}`
			case "no-run":
				payload = `{"action":"mint-adapter-completed","client_payload":{}}`
			case "negative-run":
				payload = `{"action":"mint-adapter-completed","client_payload":{"run_id":-17}}`
			}
			err := runAdapterCompletionNoticeWith(cmd, policy, f, []byte(payload), execute, noticeWait{Attempts: 4})
			want := map[string]string{"settles": "finish,propose", "string-id": "finish,propose", "waits-then-settles": "finish,propose", "unverified": "finish"}[scenario]
			if got := strings.Join(operations, ","); got != want {
				t.Fatalf("operations=%q want=%q", got, want)
			}
			wantError := scenario == "foreign-run" || scenario == "wrong-type" || scenario == "no-run" || scenario == "negative-run" || scenario == "wrong-controller-event" || scenario == "unverified"
			if (err != nil) != wantError {
				t.Fatalf("unexpected error %v", err)
			}
			if scenario == "still-active" && (polls != 4 || !strings.Contains(out.String(), "next producer callback settles it")) {
				t.Fatal("bounded wait did not end with a result", polls, out.String())
			}
			if scenario == "not-adapter" && !strings.Contains(out.String(), "does not name an adapter run") {
				t.Fatal("non-adapter notice result missing", out.String())
			}
			if scenario == "unverified" && state.InFlight != "frozen" {
				t.Fatal("an unverified run released the fence")
			}
		})
	}
}
