package promotion

import (
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTeamControlRunAuthenticatesRunBeforeActorPermission(t *testing.T) {
	for _, foreign := range []bool{false, true} {
		t.Run(fmt.Sprint(foreign), func(t *testing.T) {
			queries := 0
			cfg := Config{Repository: "owner/repo", DefaultBranch: "main", Authorization: RepositoryWriteAuthorization}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "collaborators") {
					queries++
					_, _ = fmt.Fprint(w, `{"permission":"write","user":{"login":"alice","type":"User"}}`)
					return
				}
				run := WorkflowRun{ID: 123, Path: cfg.ControlWorkflowPath(), HeadBranch: "main", Status: "in_progress", Event: "workflow_dispatch"}
				run.Actor.Login = "alice"
				run.Repository.FullName = cfg.Repository
				run.HeadRepository.FullName = cfg.Repository
				if foreign {
					run.HeadRepository.FullName = "foreign/repo"
				}
				_ = json.NewEncoder(w).Encode(run)
			}))
			defer server.Close()
			c := Client{APIURL: server.URL, Token: "test", Repository: cfg.Repository}
			err := c.AuthorizeControlRun(context.Background(), cfg, 123, "workflow_dispatch")
			if (err == nil) == foreign {
				t.Fatalf("foreign=%v err=%v", foreign, err)
			}
			if foreign && queries != 0 {
				t.Fatal("permission checked before run authentication")
			}
		})
	}
}
func TestTeamApprovalRequiresIndependentWriterAndIgnoresReadApproval(t *testing.T) {
	for _, writer := range []bool{false, true} {
		t.Run(fmt.Sprint(writer), func(t *testing.T) {
			sha := strings.Repeat("a", 40)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.Contains(r.URL.Path, "check-runs"):
					_, _ = fmt.Fprint(w, `{"total_count":0,"check_runs":[]}`)
				case strings.Contains(r.URL.Path, "reviews"):
					reviews := []map[string]any{{"state": "APPROVED", "commit_id": sha, "user": map[string]string{"login": "reader"}}}
					if writer {
						reviews = append(reviews, map[string]any{"state": "APPROVED", "commit_id": sha, "user": map[string]string{"login": "writer"}})
					}
					_ = json.NewEncoder(w).Encode(reviews)
				case strings.Contains(r.URL.Path, "collaborators/reader/"):
					_, _ = fmt.Fprint(w, `{"permission":"read","user":{"login":"reader","type":"User"}}`)
				case strings.Contains(r.URL.Path, "collaborators/writer/"):
					_, _ = fmt.Fprint(w, `{"permission":"write","user":{"login":"writer","type":"User"}}`)
				default:
					t.Errorf("unexpected %s", r.URL.Path)
				}
			}))
			defer server.Close()
			c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo", Authorization: RepositoryWriteAuthorization}
			pr := PullRequest{Number: 1}
			pr.Head.SHA = sha
			pr.User.Login = "author"
			if got := c.CheckHead(context.Background(), pr, nil) == nil; got != writer {
				t.Fatalf("approved=%v want %v", got, writer)
			}
		})
	}
}
func TestTeamIssueDoesNotRequireAssignment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "collaborators") {
			_, _ = fmt.Fprint(w, `{"permission":"write","user":{"login":"alice","type":"User"}}`)
			return
		}
		_, _ = fmt.Fprint(w, `{"number":123,"state":"open","user":{"login":"alice"},"assignees":[]}`)
	}))
	defer server.Close()
	c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo", Authorization: RepositoryWriteAuthorization}
	if err := c.VerifyIssue(context.Background(), 123, ""); err != nil {
		t.Fatal(err)
	}
}
func TestBothPolicySchemasAcceptTeamAuthorization(t *testing.T) {
	legacy := strings.Replace(minimalProductionPolicy, "human_login: human", "authorization: repository-write\nassignees: []", 1)
	cfg, err := loadTestPolicy(t, legacy)
	if err != nil || cfg.Assignees == nil || cfg.Authorization != RepositoryWriteAuthorization {
		t.Fatalf("legacy schema team: %+v %v", cfg, err)
	}
	data := `schema_version: 2
mode: package
repository: owner/repo
default_branch: main
authorization: repository-write
assignees: []
build_workflow: .github/workflows/build.yaml
environments:
  distribution:
    scope: shared
    deploy: manual
    follow: latest
    publish: true
`
	p, err := ParsePolicy([]byte(data))
	if err != nil || p.Assignees == nil {
		t.Fatalf("schema2 team: %+v %v", p, err)
	}
	for _, bad := range []string{strings.Replace(data, "authorization: repository-write", "authorization: repository-write\nhuman_login: alice", 1), strings.Replace(data, "assignees: []", "assignees: [alice, Alice]", 1)} {
		if _, err := ParsePolicy([]byte(bad)); err == nil {
			t.Fatal("accepted ambiguous policy")
		}
	}
	wire, err := yaml.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	roundtrip, err := ParsePolicy(wire)
	if err != nil || roundtrip.Assignees == nil || roundtrip.Authorization != p.Authorization {
		t.Fatalf("generated policy lost authorization or unassignment: %s %v", wire, err)
	}
	before := p
	after := p
	after.Environments = map[string]EnvironmentPolicy{"distribution": p.Environments["distribution"]}
	env := after.Environments["distribution"]
	env.Target = "v1.0.0"
	after.Environments["distribution"] = env
	after.Assignees = []string{"alice"}
	if _, err := PolicyRequestEnvironment(before, after); err == nil {
		t.Fatal("request changed assignment policy")
	}
}
