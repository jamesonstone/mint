package promotion

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVerifyHotfixIssueAcceptsOnlyGovernedRequests(t *testing.T) {
	marker := "<!-- mint:hotfix-request:123:" + sha(1) + " -->\n<!-- mint:hotfix-baseline:production-0 -->\n\nPrepare hotfix."
	for _, tc := range []struct {
		name, author, body, state string
		assigned, isPR, accepted  bool
	}{
		{"human issue", "human", "Repair login", "open", true, false, true},
		{"generated tracking issue", AutomationLogin, marker, "open", true, false, true},
		{"foreign human marker spoof", "other-human", marker, "open", true, false, false},
		{"foreign bot marker spoof", "other[bot]", marker, "open", true, false, false},
		{"arbitrary bot issue", AutomationLogin, "Repair login", "open", true, false, false},
		{"missing baseline marker", AutomationLogin, "<!-- mint:hotfix-request:123:" + sha(1) + " -->\n", "open", true, false, false},
		{"malformed source marker", AutomationLogin, "<!-- mint:hotfix-request:123:main -->\n<!-- mint:hotfix-baseline:production-0 -->\n", "open", true, false, false},
		{"unassigned generated issue", AutomationLogin, marker, "open", false, false, false},
		{"closed generated issue", AutomationLogin, marker, "closed", true, false, false},
		{"pull request cannot be issue", AutomationLogin, marker, "open", true, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.URL.Path != "/repos/owner/repo/issues/456" {
					t.Error("unexpected request", r.Method, r.URL.Path)
				}
				issue := map[string]any{"number": 456, "body": tc.body, "state": tc.state, "user": map[string]string{"login": tc.author}, "assignees": []map[string]string{}}
				if tc.assigned {
					issue["assignees"] = []map[string]string{{"login": "human"}}
				}
				if tc.isPR {
					issue["pull_request"] = map[string]string{"url": "https://github.com/owner/repo/pull/456"}
				}
				_ = json.NewEncoder(w).Encode(issue)
			}))
			defer server.Close()
			client := Client{APIURL: server.URL, Token: "installation-token", Repository: "owner/repo", HumanLogin: "human"}
			err := client.VerifyIssue(t.Context(), 456, "human")
			if (err == nil) != tc.accepted {
				t.Fatalf("accepted=%v, want=%v: %v", err == nil, tc.accepted, err)
			}
		})
	}
}
