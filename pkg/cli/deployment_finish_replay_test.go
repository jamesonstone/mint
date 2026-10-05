package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jamesonstone/mint/pkg/promotion"
)

func TestHistoricalFinishReplayPreservesLaterVerifiedJournal(t *testing.T) {
	for _, status := range []string{"deployed", "publication_pending", "deployment_failed"} {
		t.Run(status, func(t *testing.T) {
			state := promotion.NewState("owner/repo", "stage")
			state.Schema = 2
			candidateA := promotion.Candidate{SourceSHA: strings.Repeat("a", 40), Version: "v1.0.0"}
			candidateB := promotion.Candidate{SourceSHA: strings.Repeat("b", 40), Version: "v1.1.0"}
			intentA := promotion.Intent{ID: "deploy-A", Status: status, DeploymentRunID: 17, Candidate: candidateA, DeploymentURL: "https://run/17"}
			intentB := promotion.Intent{ID: "deploy-B", Status: "deployed", DeploymentRunID: 19, Candidate: candidateB, DeploymentURL: "https://run/19"}
			baselineA := promotion.Baseline{ID: intentA.ID, Candidate: candidateA, VerifiedAt: "2026-10-05T11:00:00Z", DeploymentURL: intentA.DeploymentURL}
			baselineB := promotion.Baseline{ID: intentB.ID, Candidate: candidateB, PreviousID: baselineA.ID, VerifiedAt: "2026-10-05T12:00:00Z", DeploymentURL: intentB.DeploymentURL}
			state.Intents[intentA.ID] = intentA
			state.Intents[intentB.ID] = intentB
			state.History[baselineA.ID] = baselineA
			state.History[baselineB.ID] = baselineB
			state.Baseline = &baselineB
			state.Proposals["normal"] = promotion.Proposal{ID: "proposal-B", State: "deployed", Kind: "normal", CandidateSHA: candidateB.SourceSHA}
			state.Observation = &promotion.Observation{Environment: "stage", Version: candidateB.Version, ObservedAt: "2026-10-05T12:01:00Z", RunID: 23}
			before, err := json.Marshal(state)
			if err != nil {
				t.Fatal(err)
			}
			reads := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != "/repos/owner/repo/actions/runs/17" {
					t.Error("terminal replay fetched artifacts or mutated state", r.Method, r.URL.Path)
					w.WriteHeader(500)
					return
				}
				reads++
				run := promotion.WorkflowRun{ID: 17, Path: ".github/workflows/deploy.yaml", Event: "workflow_dispatch", HeadBranch: "main", Status: "completed", Conclusion: "success", UpdatedAt: "2026-10-05T11:00:00Z", HTMLURL: "https://run/17"}
				if status == "deployment_failed" {
					run.Conclusion = "failure"
				}
				run.Repository.FullName = state.Repository
				run.HeadRepository.FullName = state.Repository
				_ = json.NewEncoder(w).Encode(run)
			}))
			defer server.Close()
			op := productionOperation{snapshot: promotion.JournalSnapshot{State: state}, config: promotion.Config{Schema: 2, DefaultBranch: "main", PromotionWorkflow: ".github/workflows/deploy.yaml"}, client: promotion.Client{APIURL: server.URL, Token: "test", Repository: state.Repository}, flags: productionFlags{IntentID: intentA.ID, RunID: 17}}
			value, changed, err := op.finish(t.Context())
			if err != nil || changed || value.(promotion.Intent).ID != intentA.ID || reads != 1 {
				t.Fatal("historical replay did not authenticate and return unchanged", changed, err, reads)
			}
			after, err := json.Marshal(op.snapshot.State)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Fatal("historical finish changed baseline timestamp, history, live observation or another journal field")
			}
		})
	}
}
