package promotion

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWholePolicyRequestSourceRejectsRebaseAuthoritySmuggling(t *testing.T) {
	original := environmentPolicyYAML
	targetOnly := strings.Replace(original, "target: v1.2.3", "target: v1.2.4", 1)
	weakened := strings.Replace(original, "deploy: reviewed", "deploy: automatic", 1)
	weakenedTarget := strings.Replace(weakened, "target: v1.2.3", "target: v1.2.4", 1)
	for _, scenario := range []string{"target-only", "weakening-first-target-last", "external-authority-drift", "wrong-environment", "ambiguous-first-parent", "empty-history", "same-policy"} {
		t.Run(scenario, func(t *testing.T) {
			sourceHead, actualBefore, expectedEnvironment := targetOnly, original, "production"
			if scenario == "weakening-first-target-last" {
				sourceHead, actualBefore = weakenedTarget, weakened
			}
			if scenario == "external-authority-drift" {
				actualBefore = weakened
			}
			if scenario == "wrong-environment" {
				expectedEnvironment = "dev"
			}
			if scenario == "same-policy" {
				sourceHead = original
			}
			before, err := ParsePolicy([]byte(actualBefore))
			if err != nil {
				t.Fatal(err)
			}
			pr := PullRequest{Number: 42}
			pr.Head.SHA = sha(3)
			reads := map[string]int{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Error("source proof performed mutation", r.Method, r.URL.Path)
					w.WriteHeader(500)
					return
				}
				switch r.URL.Path {
				case "/repos/owner/repo/pulls/42/commits":
					commits := []any{map[string]string{"sha": sha(1)}, map[string]string{"sha": sha(2)}, map[string]string{"sha": sha(3)}}
					if scenario == "empty-history" {
						commits = []any{}
					}
					_ = json.NewEncoder(w).Encode(commits)
				case "/repos/owner/repo/git/commits/" + sha(1):
					parents := []any{map[string]string{"sha": sha(0)}}
					if scenario == "ambiguous-first-parent" {
						parents = append(parents, map[string]string{"sha": sha(4)})
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"sha": sha(1), "parents": parents})
				case "/repos/owner/repo/contents/.mint.yaml":
					ref := r.URL.Query().Get("ref")
					reads[ref]++
					content := original
					if ref == sha(3) {
						content = sourceHead
					} else if ref != sha(0) {
						t.Error("source proof read mutable or final-merge base", ref)
					}
					_ = json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "content": base64.StdEncoding.EncodeToString([]byte(content))})
				default:
					t.Error("unexpected proof request", r.URL.Path)
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
			err = c.CheckPolicyRequestSource(context.Background(), pr, ".mint.yaml", before, expectedEnvironment)
			if scenario == "target-only" {
				if err != nil || reads[sha(0)] != 1 || reads[sha(3)] != 1 {
					t.Fatal(err, reads)
				}
			} else if err == nil {
				t.Fatal("unsafe or ambiguous whole-PR source accepted", scenario)
			}
		})
	}
}
