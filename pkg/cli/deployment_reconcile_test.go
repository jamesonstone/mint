package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jamesonstone/mint/pkg/promotion"
)

func TestExplicitReconciliationRequiresFreshExactRuntime(t *testing.T) {
	for _, scenario := range []string{"intent", "unchanged", "unknown", "drift", "stale", "foreign", "unauthorized", "running", "policy-change", "replay", "fallback"} {
		t.Run(scenario, func(t *testing.T) {
			cfg := promotion.Config{Schema: 2, Mode: "deployment", Scope: "shared", Deploy: "reviewed", Repository: "owner/repo", Environment: "stage", DefaultBranch: "main", HumanLogin: "human", ControlWorkflow: ".github/workflows/control.yaml", PromotionWorkflow: ".github/workflows/deploy.yaml", ObservationWorkflow: ".github/workflows/observe.yaml", Configuration: "sha256:" + strings.Repeat("c", 64)}
			s := promotion.NewState(cfg.Repository, cfg.Environment)
			s.Schema = 2
			s.Configuration = cfg.Configuration
			old := promotion.Candidate{Repository: cfg.Repository, Environment: "build", Kind: "normal", SourceSHA: strings.Repeat("a", 40), Version: "v1.0.0", Artifact: promotion.Artifact{Reference: "old", Digest: "sha256:" + strings.Repeat("b", 64), Configuration: "sha256:" + strings.Repeat("f", 64)}}
			next := old
			next.SourceSHA = strings.Repeat("d", 40)
			next.Version = "v1.0.1"
			next.Artifact.Reference = "new"
			next.Artifact.Digest = "sha256:" + strings.Repeat("e", 64)
			authority, _ := promotion.AuthorityDigest(cfg)
			intent := promotion.Intent{ID: "approved", ProposalID: "proposal", Kind: "normal", Status: "deploying", BaselineID: "baseline", DeploymentRunID: 17, Candidate: next, Configuration: cfg.Configuration, PolicyDigest: authority}
			s.Baseline = &promotion.Baseline{ID: "baseline", Candidate: old, Configuration: cfg.Configuration}
			s.Candidates[next.SourceSHA] = next
			s.Intents[intent.ID] = intent
			s.InFlight = intent.ID
			s.Proposals["normal"] = promotion.Proposal{ID: intent.ProposalID, Kind: "normal", State: "merged"}
			now := time.Now().UTC()
			format := func(d time.Duration) string { return now.Add(d).Format(time.RFC3339Nano) }
			observation := promotion.Observation{Repository: cfg.Repository, Environment: cfg.Environment, SourceSHA: next.SourceSHA, Version: next.Version, Artifact: next.Artifact, Configuration: cfg.Configuration, Status: "live", ObservedAt: format(-time.Minute), RunID: 31, RunURL: "https://run/31"}
			if scenario == "unchanged" {
				observation.SourceSHA = old.SourceSHA
				observation.Version = old.Version
				observation.Artifact = old.Artifact
			}
			if scenario == "unknown" {
				observation.Status = "deploying"
			}
			if scenario == "drift" {
				observation.SourceSHA = strings.Repeat("9", 40)
			}
			if scenario == "stale" {
				observation.ObservedAt = format(-3 * time.Minute)
			}
			if scenario == "foreign" {
				observation.Environment = "other"
			}
			if scenario == "policy-change" {
				intent.PolicyDigest = "sha256:" + strings.Repeat("8", 64)
				s.Intents[intent.ID] = intent
			}
			if scenario == "fallback" {
				saved := observation
				s.Observation = &saved
			}
			archive, digest := completionObservationArchive(t, observation)
			controller := promotion.WorkflowRun{ID: 99, Path: cfg.ControlWorkflow, HeadBranch: cfg.DefaultBranch, Event: "workflow_dispatch", Status: "in_progress"}
			controller.Actor.Login = "human"
			original := promotion.WorkflowRun{ID: 17, Path: cfg.PromotionWorkflow, HeadBranch: cfg.DefaultBranch, Event: "workflow_dispatch", Status: "completed", Conclusion: "failure", UpdatedAt: format(-2 * time.Minute), HTMLURL: "https://run/17"}
			observer := promotion.WorkflowRun{ID: 31, Path: cfg.ObservationWorkflow, HeadBranch: cfg.DefaultBranch, Event: "workflow_dispatch", Status: "completed", Conclusion: "success", CreatedAt: format(-4 * time.Minute), UpdatedAt: format(-30 * time.Second), HTMLURL: observation.RunURL}
			for _, run := range []*promotion.WorkflowRun{&controller, &original, &observer} {
				run.Repository.FullName = cfg.Repository
				run.HeadRepository.FullName = cfg.Repository
			}
			if scenario == "unauthorized" {
				controller.Actor.Login = "other"
			}
			if scenario == "running" {
				original.Status = "in_progress"
			}
			calls := map[string]int{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls[r.URL.Path]++
				if r.Method != "GET" {
					t.Error("unexpected mutation", r.Method)
					w.WriteHeader(500)
					return
				}
				switch {
				case strings.HasSuffix(r.URL.Path, "/runs/99"):
					_ = json.NewEncoder(w).Encode(controller)
				case strings.HasSuffix(r.URL.Path, "/runs/17"):
					_ = json.NewEncoder(w).Encode(original)
				case strings.HasSuffix(r.URL.Path, "/runs/31"):
					_ = json.NewEncoder(w).Encode(observer)
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
			op := productionOperation{config: cfg, snapshot: promotion.JournalSnapshot{State: s}, client: promotion.Client{APIURL: server.URL, Token: "test", Repository: cfg.Repository}, flags: productionFlags{RunID: 99}}
			intentID, observerID := intent.ID, int64(31)
			if scenario == "fallback" {
				intentID = ""
				observerID = 0
			}
			_, changed, err := op.reconcileDeployment(t.Context(), intentID, observerID)
			accepted := scenario == "intent" || scenario == "unchanged" || scenario == "replay" || scenario == "fallback"
			if accepted {
				if err != nil || !changed || op.snapshot.State.InFlight != "" {
					t.Fatal("exact runtime not reconciled", changed, err)
				}
				if op.snapshot.State.Observation == nil {
					t.Fatal("accepted runtime evidence lost")
				}
			} else if err == nil || changed || op.snapshot.State.InFlight != intent.ID || op.snapshot.State.Baseline.ID != "baseline" {
				t.Fatal("uncertain/unauthorized reconciliation mutated lifecycle", changed, err)
			}
			if scenario == "unchanged" && (op.snapshot.State.Baseline.ID != "baseline" || op.snapshot.State.Intents[intent.ID].Status != "deployment_failed") {
				t.Fatal("unchanged runtime advanced baseline")
			}
			if accepted && scenario != "unchanged" && op.snapshot.State.Baseline.Candidate.SourceSHA != next.SourceSHA {
				t.Fatal("verified intent did not become baseline")
			}
			if scenario == "replay" {
				op.snapshot.State.InFlight = "later-deployment"
				before := calls["/repos/owner/repo/actions/runs/17"]
				_, again, replayErr := op.reconcileDeployment(t.Context(), "", 0)
				if replayErr != nil || again || op.snapshot.State.InFlight != "later-deployment" || calls["/repos/owner/repo/actions/runs/17"] != before {
					t.Fatal("old recovery consumed a later fence", again, replayErr)
				}
			}
		})
	}
}
