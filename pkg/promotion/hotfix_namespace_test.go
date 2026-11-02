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

func TestLinearHotfixSourceNamespace(t *testing.T) {
	for _, tc := range []struct{ component, ticket, want string }{
		{"", linearTestTicket, "mint-hotfix-source/ENG-42"},
		{"backend", linearTestTicket, "mint-hotfix-source/backend/ENG-42"},
		{"frontend", linearTestTicket, "mint-hotfix-source/frontend/ENG-42"},
		{"", "", "ENG-42"},
	} {
		if got := (Component{Name: tc.component}).hotfixSourceBranch("ENG-42", tc.ticket); got != tc.want {
			t.Errorf("got %q want %q", got, tc.want)
		}
	}
}

func TestLinearHotfixRecoveryPreservesOrdinaryBranchAndLegacySources(t *testing.T) {
	for _, tc := range []struct {
		name           string
		metadataStatus int
		legacy         bool
		wantError      bool
	}{
		{"ordinary development branch", 404, false, false},
		{"legacy prepared request", 200, true, false},
		{"API failure must fail closed", 500, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Error("attempted mutation", r.Method)
					w.WriteHeader(500)
					return
				}
				switch {
				case strings.Contains(r.URL.Path, "git/ref/heads/mint-hotfix-source/"):
					w.WriteHeader(404)
				case strings.HasSuffix(r.URL.Path, "git/ref/heads/ENG-42"):
					json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": sha(2)}})
				case strings.Contains(r.URL.Path, "contents/.mint/hotfix.json"):
					if tc.metadataStatus != 200 {
						w.WriteHeader(tc.metadataStatus)
						return
					}
					data, _ := json.Marshal(map[string]any{"baseline_id": "base", "baseline_sha": sha(1), "ticket_url": linearTestTicket, "fixes": []string{sha(3)}})
					json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "content": base64.StdEncoding.EncodeToString(data)})
				default:
					t.Error("unexpected request", r.URL.Path)
					w.WriteHeader(404)
				}
			}))
			defer server.Close()
			client := Client{APIURL: server.URL, Repository: "owner/repo", Token: "test"}
			found, err := client.RecoverHotfixSource(context.Background(), Baseline{ID: "base", Candidate: Candidate{SourceSHA: sha(1)}}, 0, []string{sha(3)}, linearTestTicket)
			if (err != nil) != tc.wantError {
				t.Fatalf("error=%v", err)
			}
			if tc.legacy {
				if found == nil || found.Branch != "ENG-42" || found.SourceSHA != sha(2) {
					t.Fatal("legacy request not recovered", found)
				}
			} else if found != nil {
				t.Fatal("ordinary branch reused", found)
			}
		})
	}
}
