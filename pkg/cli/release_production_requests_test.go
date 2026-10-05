package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/mint/pkg/promotion"
)

func TestRollbackControlRerunDoesNotResolveAnotherTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/repos/owner/repo/actions/runs/789" {
			t.Error("rollback rerun unexpectedly touched proposal or history", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 789, "path": ".github/workflows/mint-recovery.yaml", "event": "workflow_dispatch", "status": "in_progress", "head_branch": "main", "repository": map[string]string{"full_name": "owner/repo"}, "head_repository": map[string]string{"full_name": "owner/repo"}, "actor": map[string]string{"login": "human"}})
	}))
	defer server.Close()
	event := filepath.Join(t.TempDir(), "event.json")
	if err := os.WriteFile(event, []byte(`{"repository":{"full_name":"owner/repo"},"inputs":{"operation":"rollback","reason":"restore login","target_version":""}}`), 0600); err != nil {
		t.Fatal(err)
	}
	s := promotion.NewState("owner/repo", "production")
	// Production already changed and no previous chain is available: re-resolving
	// an omitted target would fail (or incorrectly reverse the original rollback).
	s.Baseline = &promotion.Baseline{ID: "production-after-rollback"}
	prior := promotion.Proposal{ID: "original-rollback", Kind: "rollback", CandidateSHA: strings.Repeat("a", 40), BaselineID: "production-before-rollback", Summary: "restore login", State: "open", PR: 123}
	s.ControlRequests = map[string]promotion.Proposal{"rollback-control-789": prior}
	s.Intents["original-deployment"] = promotion.Intent{ProposalID: prior.ID, Status: "deployed"}
	op := productionOperation{snapshot: promotion.JournalSnapshot{State: s}, config: promotion.Config{Repository: "owner/repo", HumanLogin: "human", DefaultBranch: "main"}, client: promotion.Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}, flags: productionFlags{RunID: 789, Event: "workflow_dispatch", Input: event}}
	value, mutated, err := op.control(t.Context())
	prior.State = "deployed"
	if err != nil || mutated || value != prior {
		t.Fatalf("same-run replay did not return original deployment: value=%v mutated=%v err=%v", value, mutated, err)
	}
}

func TestRollbackReplayDoesNotModifyReviewedOrCompletedProposal(t *testing.T) {
	for _, state := range []string{"paused", "merged", "deploying", "publication_pending", "deployed"} {
		t.Run(state, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Error("replay unexpectedly touched GitHub", r.Method, r.URL.Path)
				w.WriteHeader(http.StatusInternalServerError)
			}))
			defer server.Close()
			s := promotion.NewState("owner/repo", "production")
			s.Baseline = &promotion.Baseline{ID: "production-1"}
			prior := promotion.Proposal{ID: "rollback-production-1", Kind: "rollback", CandidateSHA: strings.Repeat("a", 40), BaselineID: s.Baseline.ID, Summary: "restore login", State: state, PR: 123, Branch: "GH-456", HeadSHA: strings.Repeat("b", 40)}
			s.Proposals["rollback"] = prior
			op := productionOperation{snapshot: promotion.JournalSnapshot{State: s}, client: promotion.Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}, flags: productionFlags{Pin: prior.CandidateSHA, Summary: prior.Summary}}
			value, mutated, err := op.execute(t.Context(), "propose-rollback")
			if err != nil || mutated || value != prior {
				t.Fatalf("replay did not preserve prior reviewed state: value=%v mutated=%v err=%v", value, mutated, err)
			}
		})
	}
}
