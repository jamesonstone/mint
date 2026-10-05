package cli

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/jamesonstone/mint/pkg/promotion"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFailedRunDoesNotUnlockWithoutVerifiedUnchangedRuntime(t *testing.T) {
	for _, scenario := range []string{"failure", "cancelled", "timed_out", "missing", "wrong-artifact", "unverified", "unchanged"} {
		t.Run(scenario, func(t *testing.T) {
			s := promotion.NewState("owner/repo", "production")
			baseline := promotion.Candidate{SourceSHA: strings.Repeat("a", 40), Artifact: promotion.Artifact{Digest: "sha256:" + strings.Repeat("b", 64), Configuration: "sha256:" + strings.Repeat("c", 64)}}
			s.Baseline = &promotion.Baseline{ID: "prior", Candidate: baseline}
			i := promotion.Intent{ID: "attempt", ProposalID: "proposal", Kind: "normal", MergeSHA: strings.Repeat("d", 40), Status: "deploying", DeploymentRunID: 17}
			s.Intents[i.ID] = i
			s.InFlight = i.ID
			m := promotion.DeploymentManifest{IntentID: i.ID, SourceSHA: baseline.SourceSHA, ArtifactDigest: baseline.Artifact.Digest, Configuration: baseline.Artifact.Configuration, Verified: true}
			if scenario == "unchanged" || scenario == "wrong-artifact" || scenario == "unverified" {
				m.Outcome = "unchanged"
			}
			if scenario == "wrong-artifact" {
				m.ArtifactDigest = "sha256:" + strings.Repeat("e", 64)
			}
			if scenario == "unverified" {
				m.Verified = false
			}
			var data bytes.Buffer
			z := zip.NewWriter(&data)
			file, err := z.Create("mint-deployment.json")
			if err != nil {
				t.Fatal(err)
			}
			if err := json.NewEncoder(file).Encode(m); err != nil {
				t.Fatal(err)
			}
			if err := z.Close(); err != nil {
				t.Fatal(err)
			}
			digest := fmt.Sprintf("sha256:%x", sha256.Sum256(data.Bytes()))
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasSuffix(r.URL.Path, "/runs/17"):
					conclusion := "failure"
					if scenario == "cancelled" || scenario == "timed_out" {
						conclusion = scenario
					}
					run := promotion.WorkflowRun{ID: 17, HeadSHA: i.MergeSHA, Path: ".github/workflows/deploy.yaml", Event: "push", Status: "completed", Conclusion: conclusion, HTMLURL: "https://run/17"}
					run.Repository.FullName = s.Repository
					run.HeadRepository.FullName = s.Repository
					_ = json.NewEncoder(w).Encode(run)
				case strings.HasSuffix(r.URL.Path, "/artifacts"):
					artifacts := []any{}
					if scenario != "missing" {
						artifacts = append(artifacts, map[string]any{"id": 9, "name": "mint-deployment", "digest": digest})
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"total_count": len(artifacts), "artifacts": artifacts})
				case strings.HasSuffix(r.URL.Path, "/zip"):
					_, _ = w.Write(data.Bytes())
				default:
					t.Error("unexpected request", r.URL.Path)
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			op := productionOperation{snapshot: promotion.JournalSnapshot{State: s}, client: promotion.Client{APIURL: server.URL, Token: "test", Repository: s.Repository}, config: promotion.Config{PromotionWorkflow: ".github/workflows/deploy.yaml"}, flags: productionFlags{RunID: 17, IntentID: i.ID}}
			_, changed, err := op.finish(t.Context())
			if scenario == "unchanged" {
				if err != nil || !changed || op.snapshot.State.InFlight != "" || op.snapshot.State.Intents[i.ID].Status != "deployment_failed" {
					t.Fatalf("verified unchanged failure not finalized: %v", err)
				}
			} else if err == nil || changed || op.snapshot.State.InFlight != i.ID || op.snapshot.State.Intents[i.ID].Status != "deploying" {
				t.Fatalf("uncertain runtime unlocked: %s %v", scenario, err)
			}
			if op.snapshot.State.Baseline.ID != "prior" {
				t.Fatal("failure advanced baseline")
			}
		})
	}
}
