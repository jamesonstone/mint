package promotion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRepositoryWriteAuthorization(t *testing.T) {
	for _, tc := range []struct {
		name, permission, login, kind string
		status                        int
		allowed                       bool
	}{
		{"write", "write", "alice", "User", 200, true}, {"maintain", "maintain", "alice", "User", 200, true}, {"admin", "admin", "alice", "User", 200, true},
		{"read", "read", "alice", "User", 200, false}, {"triage", "triage", "alice", "User", 200, false}, {"bot", "write", "alice", "Bot", 200, false},
		{"foreign", "write", "other", "User", 200, false}, {"missingtype", "write", "alice", "", 200, false}, {"unknown", "custom", "alice", "User", 200, false},
		{"missing", "", "", "", 404, false}, {"forbidden", "", "", "", 403, false}, {"error", "", "", "", 500, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/repos/owner/repo/collaborators/alice/permission" || r.Method != "GET" {
					t.Errorf("unexpected permission lookup: %s", r.URL.Path)
				}
				w.WriteHeader(tc.status)
				if tc.status == 200 {
					_, _ = fmt.Fprintf(w, `{"permission":%q,"user":{"login":%q,"type":%q}}`, tc.permission, tc.login, tc.kind)
				}
			}))
			defer server.Close()
			c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo", Authorization: RepositoryWriteAuthorization}
			if got := c.AuthorizeHuman(context.Background(), "alice") == nil; got != tc.allowed {
				t.Fatalf("allowed=%v want %v", got, tc.allowed)
			}
		})
	}
}
func TestAuthorizationPolicyValidationAndAssignment(t *testing.T) {
	for _, tc := range []struct {
		mode, human string
		assignees   []string
		valid       bool
	}{
		{"", "human", nil, true}, {RepositoryWriteAuthorization, "", nil, true}, {RepositoryWriteAuthorization, "", []string{}, true},
		{"", "", nil, false}, {RepositoryWriteAuthorization, "human", nil, false}, {"team", "", nil, false}, {"", "a[bot]", nil, false},
		{RepositoryWriteAuthorization, "", []string{"alice", "Alice"}, false}, {RepositoryWriteAuthorization, "", []string{""}, false},
	} {
		if got := validateAuthorization(tc.mode, tc.human, tc.assignees) == nil; got != tc.valid {
			t.Errorf("%+v valid=%v", tc, got)
		}
	}
	legacy := Client{HumanLogin: "human"}
	if len(legacy.assignmentLogins()) != 1 {
		t.Fatal("legacy assignment lost")
	}
	legacy.Assignees = []string{}
	if len(legacy.assignmentLogins()) != 0 {
		t.Fatal("explicit empty assignment lost")
	}
	team := Client{Authorization: RepositoryWriteAuthorization}
	if _, ok := team.withAssignments(map[string]any{"assignees": nil})["assignees"]; ok {
		t.Fatal("empty assignments emitted")
	}
	if err := team.assign(context.Background(), 123); err != nil {
		t.Fatal(err)
	}
	if !team.IsReleaseAuthor(AutomationLogin) {
		t.Fatal("built-in automation author rejected")
	}
	if team.IsReleaseAuthor("other[bot]") {
		t.Fatal("arbitrary bot accepted")
	}
}
func TestAuthorizationStrictPolicyAndDigest(t *testing.T) {
	for _, data := range []string{"authorization: null", "authorization: 42", "assignees: null", "assignees: alice", "assignees: [42]"} {
		if _, err := policyDocument([]byte(data)); err == nil {
			t.Errorf("accepted %s", data)
		}
	}
	cfg := Config{HumanLogin: "human"}
	wire, _ := json.Marshal(cfg)
	if strings.Contains(string(wire), "Authorization") || strings.Contains(string(wire), "Assignees") {
		t.Fatal("legacy wire changed")
	}
	before, _ := AuthorityDigest(cfg)
	cfg.Assignees = []string{}
	after, _ := AuthorityDigest(cfg)
	if before == after {
		t.Fatal("explicit unassigned authority not protected")
	}
	cfg.Authorization = RepositoryWriteAuthorization
	cfg.HumanLogin = ""
	third, _ := AuthorityDigest(cfg)
	if third == after {
		t.Fatal("authorization not protected")
	}
	p := Policy{Assignees: []string{}}
	p.Schema = 2
	p.Environments = map[string]EnvironmentPolicy{"stage": {Scope: "shared"}}
	p.Repository = "owner/repo"
	p.DefaultBranch = "main"
	p.Authorization = RepositoryWriteAuthorization
	c, err := p.ForEnvironment("stage")
	if err != nil || c.Assignees == nil || c.Authorization != RepositoryWriteAuthorization {
		t.Fatalf("assignment policy lost: %+v %v", c, err)
	}
}
func TestTeamWorkflowHasNoPersonalActorGate(t *testing.T) {
	workflow, err := RenderControlWorkflow(Config{Authorization: RepositoryWriteAuthorization}, strings.Repeat("a", 40))
	if err != nil || strings.Contains(workflow, "github.actor") || !strings.Contains(workflow, "MINT_RELEASE_ENABLED") {
		t.Fatalf("unsafe team workflow: %v\n%s", err, workflow)
	}
}
