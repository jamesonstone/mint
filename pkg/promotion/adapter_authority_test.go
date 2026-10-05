package promotion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdapterAuthorityBindsRoleToLiveWorkflow(t *testing.T) {
	cfg := Config{Schema: 1, Repository: "owner/repo", DefaultBranch: "main", Authorization: "repository-write", ControlWorkflow: ".github/workflows/control.yaml", BuildWorkflow: ".github/workflows/build.yaml", PromotionWorkflow: ".github/workflows/promote.yaml"}
	cases := []struct{ role, event, path string }{
		{"source", "push", cfg.BuildWorkflow}, {"source", "workflow_dispatch", cfg.BuildWorkflow},
		{"promote", "push", cfg.PromotionWorkflow}, {"control", "workflow_run", cfg.ControlWorkflow},
		{"control", "pull_request_target", cfg.ControlWorkflow}, {"control", "workflow_dispatch", cfg.ControlWorkflow},
	}
	for _, tc := range cases {
		for _, mutation := range []string{"valid", "id", "path", "event", "status", "branch", "repo", "head-repo", "permission"} {
			t.Run(tc.role+"/"+tc.event+"/"+mutation, func(t *testing.T) {
				run := WorkflowRun{ID: 7, Path: tc.path, Event: tc.event, Status: "in_progress", HeadBranch: "main"}
				run.Repository.FullName, run.HeadRepository.FullName, run.Actor.Login = "owner/repo", "owner/repo", "teammate"
				switch mutation {
				case "id":
					run.ID = 8
				case "path":
					run.Path = cfg.ValidationWorkflow
				case "event":
					run.Event = "pull_request"
				case "status":
					run.Status = "completed"
				case "branch":
					run.HeadBranch = "unreviewed"
				case "repo":
					run.Repository.FullName = "foreign/repo"
				case "head-repo":
					run.HeadRepository.FullName = "foreign/repo"
				}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/repos/owner/repo/actions/runs/7" {
						_ = json.NewEncoder(w).Encode(run)
						return
					}
					if r.URL.Path == "/repos/owner/repo/collaborators/teammate/permission" {
						permission := "write"
						if mutation == "permission" {
							permission = "read"
						}
						_ = json.NewEncoder(w).Encode(map[string]any{"permission": permission, "user": map[string]string{"login": "teammate", "type": "User"}})
						return
					}
					t.Errorf("unexpected API %s", r.URL.Path)
					w.WriteHeader(404)
				}))
				defer server.Close()
				err := (Client{APIURL: server.URL, Token: "test", Repository: cfg.Repository}).AuthorizeAdapterRun(context.Background(), cfg, 7, tc.role, tc.event)
				valid := mutation == "valid" || mutation == "permission" && tc.event != "workflow_dispatch" && tc.event != "pull_request_target"
				if (err == nil) != valid {
					t.Fatalf("valid=%v error=%v", valid, err)
				}
			})
		}
	}
	if (Client{}).AuthorizeAdapterRun(context.Background(), cfg, 7, "promote", "workflow_dispatch") == nil {
		t.Fatal("unsupported role/event allowed")
	}
	cfg.Schema = 2
	if (Client{}).AuthorizeAdapterRun(context.Background(), cfg, 7, "source", "push") == nil {
		t.Fatal("schema 2 compatibility allowed")
	}
}
