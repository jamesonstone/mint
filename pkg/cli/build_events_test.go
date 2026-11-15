package cli

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

var chainedBuildEvents = []string{"push", "workflow_dispatch", "workflow_run"}

func TestRegistrationFollowsBuildEventsForChainedProducers(t *testing.T) {
	chain := func(_ *promotion.Candidate, r *promotion.WorkflowRun) { r.Event = "workflow_run" }
	cases := []struct {
		name       string
		events     []string
		mutate     func(*promotion.Candidate, *promotion.WorkflowRun)
		registered bool
		message    string
	}{
		{"strict-default-refuses-chained", nil, chain, false, "trusted producer run"},
		{"opt-in-registers-chained", chainedBuildEvents, chain, true, ""},
		{"opt-in-keeps-push", chainedBuildEvents, func(*promotion.Candidate, *promotion.WorkflowRun) {}, true, ""},
		{"chained-only-refuses-push", []string{"workflow_run"}, func(*promotion.Candidate, *promotion.WorkflowRun) {}, false, "trusted producer run"},
		{"opt-in-refuses-pull-request", chainedBuildEvents, func(_ *promotion.Candidate, r *promotion.WorkflowRun) { r.Event = "pull_request" }, false, "trusted producer run"},
		{"opt-in-refuses-foreign-head", chainedBuildEvents, func(_ *promotion.Candidate, r *promotion.WorkflowRun) {
			r.Event, r.HeadRepository.FullName = "workflow_run", "fork/repo"
		}, false, "trusted producer run"},
		{"opt-in-refuses-failed", chainedBuildEvents, func(_ *promotion.Candidate, r *promotion.WorkflowRun) {
			r.Event, r.Conclusion = "workflow_run", "failure"
		}, false, "trusted producer run"},
		{"opt-in-refuses-other-branch", chainedBuildEvents, func(_ *promotion.Candidate, r *promotion.WorkflowRun) {
			r.Event, r.HeadBranch = "workflow_run", "feature"
		}, false, "does not match successful build"},
		{"opt-in-refuses-other-source", chainedBuildEvents, func(_ *promotion.Candidate, r *promotion.WorkflowRun) {
			r.Event, r.HeadSHA = "workflow_run", strings.Repeat("a", 40)
		}, false, "does not match successful build"},
		{"opt-in-refuses-foreign-tag", chainedBuildEvents, func(c *promotion.Candidate, r *promotion.WorkflowRun) {
			r.Event, c.Version = "workflow_run", "v1.0.0"
		}, false, "version tag"},
		{"opt-in-refuses-chained-hotfix", chainedBuildEvents, func(c *promotion.Candidate, r *promotion.WorkflowRun) {
			r.Event, c.Kind, c.SourcePR = "workflow_run", "hotfix", 5
		}, false, "workflow_run"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			op, c, run := candidateFixture(t)
			op.config.BuildEvents = tc.events
			tc.mutate(&c, &run)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !strings.HasSuffix(r.URL.Path, "/pulls") {
					t.Error("unexpected provenance request", r.URL.Path)
					w.WriteHeader(500)
					return
				}
				_ = json.NewEncoder(w).Encode([]any{})
			}))
			defer server.Close()
			op.client = promotion.Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
			_, changed, err := op.registerCandidate(context.Background(), c, run)
			if tc.registered {
				if err != nil || !changed || op.snapshot.State.Candidates[c.SourceSHA].RunID != run.ID {
					t.Fatal("trusted producer refused", changed, err)
				}
				return
			}
			if err == nil || changed || len(op.snapshot.State.Candidates) != 0 {
				t.Fatal("untrusted producer registered", changed, err)
			}
			if !strings.Contains(err.Error(), tc.message) {
				t.Fatal("refused by another rule", err)
			}
		})
	}
}

func TestChainedProducerCandidateAndScanNeedPolicyOptIn(t *testing.T) {
	for _, optIn := range []bool{false, true} {
		for _, mode := range []string{"candidate", "scan"} {
			t.Run(fmt.Sprintf("%s-opt-in-%v", mode, optIn), func(t *testing.T) {
				op, c, run := candidateFixture(t)
				run.Event = "workflow_run"
				if optIn {
					op.config.BuildEvents = chainedBuildEvents
				}
				archive, digest := manifestArchive(t, c)
				downloads := 0
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch {
					case r.URL.Path == "/repos/owner/repo/actions/workflows/build.yaml/runs":
						_ = json.NewEncoder(w).Encode(map[string]any{"workflow_runs": []promotion.WorkflowRun{run}})
					case r.URL.Path == "/repos/owner/repo/actions/runs/7":
						_ = json.NewEncoder(w).Encode(run)
					case r.URL.Path == "/repos/owner/repo/actions/runs/7/artifacts":
						_ = json.NewEncoder(w).Encode(map[string]any{"total_count": 1, "artifacts": []any{map[string]any{"id": 9, "name": "mint-candidate", "digest": digest}}})
					case r.URL.Path == "/repos/owner/repo/actions/artifacts/9/zip":
						downloads++
						_, _ = w.Write(archive)
					case strings.HasSuffix(r.URL.Path, "/pulls"):
						_ = json.NewEncoder(w).Encode([]any{})
					default:
						t.Error("unexpected request", r.URL.Path)
						w.WriteHeader(500)
					}
				}))
				defer server.Close()
				op.client = promotion.Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
				_, changed, err := op.execute(context.Background(), mode)
				registered := len(op.snapshot.State.Candidates)
				switch {
				case optIn:
					if err != nil || !changed || downloads != 1 || registered != 1 {
						t.Fatal("policy-trusted chained producer not registered", changed, err, downloads, registered)
					}
				case mode == "candidate":
					if err == nil || changed || downloads != 0 || registered != 0 {
						t.Fatal("strict policy registered a chained producer", changed, err, downloads, registered)
					}
				default:
					// A scan skips runs outside the trust set; it reads no manifest.
					if err != nil || changed || downloads != 0 || registered != 0 {
						t.Fatal("strict scan read or registered a chained producer", changed, err, downloads, registered)
					}
				}
			})
		}
	}
}

func TestBootstrapSourceBuildFollowsBuildEventsAndVerifierStaysFixed(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) string {
		command := exec.Command("git", args...)
		command.Dir = dir
		data, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("git fixture: %v %s", err, data)
		}
		return strings.TrimSpace(string(data))
	}
	git("init", "-b", "main")
	git("config", "user.name", "Human")
	git("config", "user.email", "human@example.com")
	if err := os.WriteFile(filepath.Join(dir, "source.txt"), []byte("source"), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", "source.txt")
	git("commit", "-m", "feat: source")
	source := git("rev-parse", "HEAD")
	git("tag", "v1.0.0")
	cases := []struct {
		name                      string
		events                    []string
		buildEvent, verifierEvent string
		imported                  bool
	}{
		{"strict-default-refuses-chained-build", nil, "workflow_run", "workflow_dispatch", false},
		{"opt-in-accepts-chained-build", chainedBuildEvents, "workflow_run", "workflow_dispatch", true},
		{"opt-in-keeps-dispatched-build", chainedBuildEvents, "workflow_dispatch", "workflow_dispatch", true},
		{"opt-in-refuses-chained-verifier", chainedBuildEvents, "push", "workflow_run", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candidate := promotion.Candidate{Repository: "owner/repo", Environment: "production", SourceSHA: source, Version: "v1.0.0", Kind: "normal", Artifact: promotion.Artifact{Reference: "immutable/object", Digest: "sha256:" + strings.Repeat("a", 64), Configuration: "sha256:" + strings.Repeat("b", 64)}, RunID: 7, RunURL: "https://github.com/owner/repo/actions/runs/7"}
			baseline := promotion.Baseline{ID: "verified-runtime", Candidate: candidate, DeploymentURL: "https://github.com/owner/repo/actions/runs/8", Shipped: []promotion.Change{}}
			var archive bytes.Buffer
			zw := zip.NewWriter(&archive)
			file, err := zw.Create("mint-baseline.json")
			if err != nil {
				t.Fatal(err)
			}
			if err := json.NewEncoder(file).Encode(baseline); err != nil {
				t.Fatal(err)
			}
			if err := zw.Close(); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(archive.Bytes())
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/repos/owner/repo/actions/runs/8/artifacts":
					_ = json.NewEncoder(w).Encode(map[string]any{"total_count": 1, "artifacts": []any{map[string]any{"id": 9, "name": "mint-baseline", "digest": fmt.Sprintf("sha256:%x", sum)}}})
				case "/repos/owner/repo/actions/artifacts/9/zip":
					_, _ = w.Write(archive.Bytes())
				default:
					run := promotion.WorkflowRun{ID: 8, HeadSHA: strings.Repeat("c", 40), HeadBranch: "main", Event: tc.verifierEvent, Path: ".github/workflows/baseline.yaml", Status: "completed", Conclusion: "success"}
					if strings.HasSuffix(r.URL.Path, "/7") {
						run.ID, run.HeadSHA, run.Event, run.Path = 7, source, tc.buildEvent, ".github/workflows/build.yaml"
					}
					run.HTMLURL = fmt.Sprintf("https://github.com/owner/repo/actions/runs/%d", run.ID)
					run.Repository.FullName, run.HeadRepository.FullName = "owner/repo", "owner/repo"
					_ = json.NewEncoder(w).Encode(run)
				}
			}))
			defer server.Close()
			operation := productionOperation{config: promotion.Config{DefaultBranch: "main", BuildEvents: tc.events, BaselineWorkflow: ".github/workflows/baseline.yaml", BaselineBuildWorkflow: ".github/workflows/build.yaml"}, client: promotion.Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}, snapshot: promotion.JournalSnapshot{State: promotion.NewState("owner/repo", "production")}, flags: productionFlags{RunID: 8}, proof: promotion.GitProof{Context: context.Background(), WorkDir: dir}}
			_, changed, err := operation.bootstrap(context.Background())
			if tc.imported {
				if err != nil || !changed || operation.snapshot.State.Baseline == nil {
					t.Fatal("policy-trusted source build refused", changed, err)
				}
			} else if err == nil || changed || operation.snapshot.State.Baseline != nil {
				t.Fatal("bootstrap trusted a run outside its event set", changed, err)
			}
		})
	}
}

func TestBuildCallbackAuthenticatesChainedProducerByPolicy(t *testing.T) {
	cases := []struct {
		name    string
		events  []string
		mutate  func(*promotion.WorkflowRun)
		result  string
		message string
	}{
		{"strict-default-refuses", nil, func(*promotion.WorkflowRun) {}, "", "trusted build"},
		{"opt-in-authenticates", chainedBuildEvents, func(*promotion.WorkflowRun) {}, "not activated", ""},
		{"opt-in-reports-failed-producer", chainedBuildEvents, func(r *promotion.WorkflowRun) { r.Conclusion = "failure" }, "producer did not succeed; no candidate registered", ""},
		{"opt-in-refuses-pull-request", chainedBuildEvents, func(r *promotion.WorkflowRun) { r.Event = "pull_request" }, "", "trusted build"},
		{"opt-in-refuses-foreign-head", chainedBuildEvents, func(r *promotion.WorkflowRun) { r.HeadRepository.FullName = "fork/repo" }, "", "trusted build"},
		{"opt-in-refuses-other-branch", chainedBuildEvents, func(r *promotion.WorkflowRun) { r.HeadBranch = "feature" }, "", "configured default branch"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			policy := promotion.Policy{Schema: 2, Mode: "deployment", Repository: "owner/repo", DefaultBranch: "main", HumanLogin: "human", BuildWorkflow: ".github/workflows/build.yaml", BuildEvents: tc.events, ControlWorkflow: ".github/workflows/control.yaml", Environments: map[string]promotion.EnvironmentPolicy{"sandbox": {Scope: "shared", Deploy: "automatic", Follow: "latest", PromotionWorkflow: ".github/workflows/deploy.yaml"}}}
			producer := promotion.WorkflowRun{ID: 17, HeadSHA: strings.Repeat("a", 40), Path: policy.BuildWorkflow, Event: "workflow_run", HeadBranch: "main", Status: "completed", Conclusion: "success"}
			producer.Repository.FullName, producer.HeadRepository.FullName = policy.Repository, policy.Repository
			tc.mutate(&producer)
			controller := promotion.WorkflowRun{ID: 99, Path: policy.ControlWorkflow, Event: "workflow_run", HeadBranch: "main", Status: "in_progress"}
			controller.Repository.FullName, controller.HeadRepository.FullName = policy.Repository, policy.Repository
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method != "GET":
					t.Errorf("unexpected mutation %s %s", r.Method, r.URL.Path)
					w.WriteHeader(500)
				case strings.HasSuffix(r.URL.Path, "/runs/17"):
					_ = json.NewEncoder(w).Encode(producer)
				case strings.HasSuffix(r.URL.Path, "/runs/99"):
					_ = json.NewEncoder(w).Encode(controller)
				default:
					// An environment with no runtime configuration reads no journal.
					t.Error("unexpected request", r.URL.Path)
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			t.Setenv("MINT_BUILD_CALLBACK_TEST_TOKEN", "test")
			var out bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetContext(t.Context())
			cmd.SetOut(&out)
			f := productionFlags{APIURL: server.URL, TokenEnv: "MINT_BUILD_CALLBACK_TEST_TOKEN", Event: "workflow_run", RunID: 99}
			err := runBuildCallback(cmd, policy, f, []byte(`{"workflow_run":{"id":17}}`))
			if tc.message != "" {
				if err == nil || !strings.Contains(err.Error(), tc.message) || out.Len() != 0 {
					t.Fatal("producer outside the policy event set passed authentication", err, out.String())
				}
				return
			}
			var results map[string]string
			if err != nil || json.Unmarshal(out.Bytes(), &results) != nil || results["sandbox"] != tc.result {
				t.Fatal("authenticated producer callback result", err, out.String())
			}
		})
	}
}

// A build that completes while a deployment is in flight is registered and
// waits; it must not fail the controller run of a Continuous Delivery
// environment. A hotfix has no automatic continuation and still fails.
func TestBuildCallbackQueuesBehindDeploymentInFlight(t *testing.T) {
	inFlight := fmt.Errorf("production intent frozen is in flight: %w", promotion.ErrInFlight)
	waits := "candidate registered; a deployment is in flight, reconciliation waits for its outcome"
	// fence: the journal state of the deployment that refused reconciliation.
	// "running" and "released" are normal; every other state needs an operator.
	cases := []struct {
		name, deploy, kind, fence string
		propose                   error
		operations, result        string
		failed                    bool
	}{
		{"automatic-reconciled", "automatic", "normal", "released", nil, "candidate,propose", "eligible candidate reconciled", false},
		{"automatic-queued", "automatic", "normal", "running", inFlight, "candidate,propose", waits, false},
		{"automatic-queued-before-conclusion", "automatic", "normal", "queued-run", inFlight, "candidate,propose", waits, false},
		{"reviewed-queued", "reviewed", "normal", "running", inFlight, "candidate,propose", waits, false},
		{"released-since-refusal", "automatic", "normal", "released", inFlight, "candidate,propose", waits, false},
		{"ended-run-keeps-failure", "automatic", "normal", "ended", inFlight, "candidate,propose", "explicit reconcile request", true},
		{"fence-without-run-keeps-failure", "automatic", "normal", "no-run", inFlight, "candidate,propose", "explicit reconcile request", true},
		{"unknown-intent-keeps-failure", "automatic", "normal", "unknown-intent", inFlight, "candidate,propose", "explicit reconcile request", true},
		{"unreadable-run-keeps-failure", "automatic", "normal", "unreadable-run", inFlight, "candidate,propose", "deployment run state unavailable", true},
		{"hotfix-in-flight-fails", "automatic", "hotfix", "running", inFlight, "candidate,propose", inFlight.Error(), true},
		{"other-error-fails", "automatic", "normal", "running", fmt.Errorf("journal changed"), "candidate,propose", "journal changed", true},
		{"manual-collects", "manual", "normal", "running", inFlight, "candidate", "candidate registered; explicit promotion required", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			policy := promotion.Policy{Schema: 2, Mode: "deployment", Repository: "owner/repo", DefaultBranch: "main", HumanLogin: "human", BuildWorkflow: ".github/workflows/build.yaml", BuildEvents: chainedBuildEvents, ControlWorkflow: ".github/workflows/control.yaml", Environments: map[string]promotion.EnvironmentPolicy{"sandbox": {Scope: "shared", Deploy: tc.deploy, Follow: "latest", Configuration: "sha256:" + strings.Repeat("a", 64), PromotionWorkflow: ".github/workflows/deploy.yaml"}}}
			state := promotion.NewState(policy.Repository, "sandbox")
			state.Schema = 2
			state.Baseline = &promotion.Baseline{ID: "verified-baseline"}
			if tc.fence != "released" {
				state.InFlight = "frozen"
				intent := promotion.Intent{ID: "frozen", Status: "deploying", DeploymentRunID: 31}
				if tc.fence == "no-run" {
					intent.DeploymentRunID = 0
				}
				if tc.fence != "unknown-intent" {
					state.Intents[intent.ID] = intent
				}
			}
			deployment := promotion.WorkflowRun{ID: 31, Path: ".github/workflows/deploy.yaml", Event: "workflow_dispatch", HeadBranch: "main", Status: "in_progress"}
			deployment.Repository.FullName, deployment.HeadRepository.FullName = policy.Repository, policy.Repository
			if tc.fence == "queued-run" {
				deployment.Status = "queued"
			}
			if tc.fence == "ended" {
				deployment.Status, deployment.Conclusion = "completed", "success"
			}
			if tc.fence == "unreadable-run" {
				deployment.Repository.FullName = "other/repo"
			}
			candidate := promotion.Candidate{Repository: policy.Repository, Environment: "build", SourceSHA: strings.Repeat("a", 40), Version: "v1.0.1", Kind: tc.kind, BaselineID: "verified-baseline", RunID: 17}
			archive, digest := manifestArchive(t, candidate)
			producer := promotion.WorkflowRun{ID: 17, HeadSHA: candidate.SourceSHA, Path: policy.BuildWorkflow, Event: "workflow_run", HeadBranch: "main", Status: "completed", Conclusion: "success"}
			producer.Repository.FullName, producer.HeadRepository.FullName = policy.Repository, policy.Repository
			controller := promotion.WorkflowRun{ID: 99, Path: policy.ControlWorkflow, Event: "workflow_run", HeadBranch: "main", Status: "in_progress"}
			controller.Repository.FullName, controller.HeadRepository.FullName = policy.Repository, policy.Repository
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.Method != "GET":
					t.Errorf("unexpected mutation %s %s", r.Method, r.URL.Path)
					w.WriteHeader(500)
				case strings.HasSuffix(r.URL.Path, "/runs/17"):
					_ = json.NewEncoder(w).Encode(producer)
				case strings.HasSuffix(r.URL.Path, "/runs/99"):
					_ = json.NewEncoder(w).Encode(controller)
				case strings.HasSuffix(r.URL.Path, "/runs/31"):
					_ = json.NewEncoder(w).Encode(deployment)
				case strings.Contains(r.URL.Path, "/git/ref/"):
					_ = json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": strings.Repeat("b", 40)}})
				case strings.Contains(r.URL.Path, "/git/commits/"):
					_ = json.NewEncoder(w).Encode(map[string]any{"tree": map[string]string{"sha": strings.Repeat("c", 40)}})
				case strings.Contains(r.URL.Path, "/git/trees/"):
					_ = json.NewEncoder(w).Encode(map[string]any{"tree": []any{map[string]string{"path": "release-state.json", "type": "blob", "sha": strings.Repeat("d", 40)}}})
				case strings.Contains(r.URL.Path, "/git/blobs/"):
					data, _ := json.Marshal(state)
					_ = json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "content": base64.StdEncoding.EncodeToString(data)})
				case strings.HasSuffix(r.URL.Path, "/runs/17/artifacts"):
					_ = json.NewEncoder(w).Encode(map[string]any{"total_count": 1, "artifacts": []any{map[string]any{"id": 9, "name": "mint-candidate", "digest": digest}}})
				case strings.HasSuffix(r.URL.Path, "/artifacts/9/zip"):
					_, _ = w.Write(archive)
				default:
					t.Error("unexpected request", r.URL.Path)
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			t.Setenv("MINT_BUILD_CALLBACK_TEST_TOKEN", "test")
			var out bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetContext(t.Context())
			cmd.SetOut(&out)
			f := productionFlags{APIURL: server.URL, TokenEnv: "MINT_BUILD_CALLBACK_TEST_TOKEN", Event: "workflow_run", RunID: 99}
			operations := []string{}
			execute := func(_ *cobra.Command, operation string, child productionFlags) error {
				operations = append(operations, operation)
				if child.Environment != "sandbox" {
					t.Fatal("child lost the environment", child)
				}
				if operation == "candidate" {
					if child.RunID != 17 {
						t.Fatal("registration did not use the producer run", child)
					}
					return nil
				}
				if child.RunID != 99 || (tc.kind == "hotfix") != (child.Kind == "hotfix" && child.Pin == candidate.SourceSHA) {
					t.Fatal("reconciliation lost controller or hotfix identity", child)
				}
				return tc.propose
			}
			err := runBuildCallbackWith(cmd, policy, f, []byte(`{"workflow_run":{"id":17}}`), execute)
			var results map[string]string
			if json.Unmarshal(out.Bytes(), &results) != nil || !strings.Contains(results["sandbox"], tc.result) {
				t.Fatal("callback result", err, out.String())
			}
			// A failure keeps the refusal that names the fenced intent.
			if tc.failed && tc.propose == inFlight && !strings.Contains(results["sandbox"], "intent frozen is in flight") {
				t.Fatal("failure lost the fenced intent", results["sandbox"])
			}
			if strings.Join(operations, ",") != tc.operations || (err != nil) != tc.failed {
				t.Fatalf("operations=%v failed=%v (%v)", operations, err != nil, err)
			}
		})
	}
}
