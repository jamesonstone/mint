package promotion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func controlConfig() Config {
	return Config{Repository: "owner/repo", DefaultBranch: "main", HumanLogin: "jamesonstone"}
}
func TestControlAuthorizationUsesLiveServerIdentity(t *testing.T) {
	cfg := controlConfig()
	for _, event := range []string{"workflow_dispatch", "pull_request_target"} {
		for _, mutation := range []string{"valid", "event", "path", "branch", "actor", "repo", "head-repo", "status", "id"} {
			t.Run(event+"/"+mutation, func(t *testing.T) {
				run := WorkflowRun{ID: 7, Path: cfg.ControlWorkflowPath(), Event: event, Status: "in_progress", HeadBranch: "main"}
				run.Repository.FullName = "owner/repo"
				run.HeadRepository.FullName = "owner/repo"
				run.Actor.Login = "jamesonstone"
				switch mutation {
				case "event":
					run.Event = "push"
				case "path":
					run.Path = ".github/workflows/untrusted.yaml"
				case "branch":
					run.HeadBranch = "feature"
				case "actor":
					run.Actor.Login = AutomationLogin
				case "repo":
					run.Repository.FullName = "other/repo"
				case "head-repo":
					run.HeadRepository.FullName = "other/repo"
				case "status":
					run.Status = "completed"
				case "id":
					run.ID = 8
				}
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != "GET" || r.URL.Path != "/repos/owner/repo/actions/runs/7" {
						t.Error(r.Method, r.URL.Path)
					}
					_ = json.NewEncoder(w).Encode(run)
				}))
				defer server.Close()
				c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
				err := c.AuthorizeControlRun(context.Background(), cfg, 7, event)
				if (err == nil) != (mutation == "valid") {
					t.Fatal(mutation, err)
				}
			})
		}
	}
	c := Client{}
	for _, event := range []string{"push", "pull_request", ""} {
		if c.AuthorizeControlRun(context.Background(), cfg, 7, event) == nil {
			t.Fatal("unsupported event accepted", event)
		}
	}
	if c.AuthorizeControlRun(context.Background(), cfg, 0, "workflow_dispatch") == nil {
		t.Fatal("missing run accepted")
	}
}

func TestControlPrefixUsesAuthoritativeMergedPR(t *testing.T) {
	cfg := controlConfig()
	for _, scenario := range []string{"hotfix", "payload-only", "unmerged", "generated", "nondefault", "foreign"} {
		t.Run(scenario, func(t *testing.T) {
			_, pr := outcomeFixture()
			pr.Title = "hotfix(GH-123): :firetruck: repair login"
			pr.Body = "Reviewed fix"
			switch scenario {
			case "payload-only":
				pr.Title = "fix(GH-123): ordinary fix"
			case "unmerged":
				pr.Merged = false
				pr.MergedAt = nil
			case "generated":
				pr.Body = "<!-- mint:proposal:generated:hotfix -->"
			case "nondefault":
				pr.Base.Ref = "maintenance"
			case "foreign":
				pr.Head.Repo.FullName = "other/repo"
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(pr) }))
			defer server.Close()
			c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
			data := []byte(`{"action":"closed","repository":{"full_name":"owner/repo"},"pull_request":{"number":42,"title":"hotfix(GH-999): untrusted payload","merged":true}}`)
			got, err := c.ControlEvent(context.Background(), cfg, "pull_request_target", data)
			if scenario == "foreign" {
				if err == nil {
					t.Fatal("foreign PR accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if scenario == "hotfix" {
				if got.Operation != "hotfix" || got.FixPR != 42 || got.Reason != pr.Title {
					t.Fatal(got)
				}
			} else if got.Operation != "" {
				t.Fatal("noneligible PR triggered preparation", got)
			}
		})
	}
}
func TestControlManualOperationAndInputAllowlist(t *testing.T) {
	cfg := controlConfig()
	for _, op := range []string{"hotfix", "rollback", "deploy", "publish", "retry", ""} {
		data, _ := json.Marshal(map[string]any{"repository": map[string]string{"full_name": "owner/repo"}, "inputs": map[string]string{"operation": op, "reason": "repair", "fix_pr": "123", "issue": "456", "target_version": "v1.2.3"}})
		got, err := (Client{}).ControlEvent(context.Background(), cfg, "workflow_dispatch", data)
		valid := op == "hotfix" || op == "rollback"
		if (err == nil) != valid {
			t.Fatal(op, err)
		}
		if valid && (got.FixPR != 123 || got.Issue != 456 || got.Version != "v1.2.3" || got.Reason != "repair") {
			t.Fatal(got)
		}
	}
	for _, field := range []string{"fix_pr", "issue"} {
		for _, value := range []string{"0", "-1", "1.2", "abc", "123; echo bad", "99999999999999999999999999999"} {
			data, _ := json.Marshal(map[string]any{"repository": map[string]string{"full_name": "owner/repo"}, "inputs": map[string]string{"operation": "hotfix", field: value}})
			if _, err := (Client{}).ControlEvent(context.Background(), cfg, "workflow_dispatch", data); err == nil {
				t.Fatal("invalid number accepted", field, value)
			}
		}
	}
	if _, err := (Client{}).ControlEvent(context.Background(), cfg, "workflow_dispatch", []byte(`{"repository":{"full_name":"other/repo"},"inputs":{"operation":"hotfix"}}`)); err == nil {
		t.Fatal("foreign payload accepted")
	}
	if _, err := (Client{}).ControlEvent(context.Background(), cfg, "workflow_dispatch", []byte(`{`)); err == nil {
		t.Fatal("invalid JSON accepted")
	}
}

func TestControlWorkflowIsPinnedAndUsesTrustedCode(t *testing.T) {
	cfg := controlConfig()
	rendered, err := RenderControlWorkflow(cfg, sha(1))
	if err != nil {
		t.Fatal(err)
	}
	var workflow map[string]any
	if err := yaml.Unmarshal([]byte(rendered), &workflow); err != nil {
		t.Fatal(err)
	}
	on := workflow["on"].(map[string]any)
	inputs := on["workflow_dispatch"].(map[string]any)["inputs"].(map[string]any)
	operations := inputs["operation"].(map[string]any)
	if fmt.Sprint(operations["options"]) != "[hotfix rollback]" || operations["default"] != "hotfix" {
		t.Fatal(operations)
	}
	request := workflow["jobs"].(map[string]any)["request"].(map[string]any)
	guard := request["if"].(string)
	if !strings.Contains(guard, "MINT_RELEASE_ENABLED == 'true'") || !strings.Contains(guard, "github.actor == 'jamesonstone'") {
		t.Fatal(guard)
	}
	steps := request["steps"].([]any)
	checkout := steps[0].(map[string]any)
	if checkout["uses"] != "actions/checkout@11d5960a326750d5838078e36cf38b85af677262" {
		t.Fatal(checkout)
	}
	with := checkout["with"].(map[string]any)
	if with["ref"] != "${{ github.event.repository.default_branch }}" || with["fetch-depth"] != 0 || with["persist-credentials"] != false {
		t.Fatal(with)
	}
	action := steps[2].(map[string]any)
	if action["uses"] != "jamesonstone/mint@"+sha(1) || action["with"].(map[string]any)["command"] != "production-control" {
		t.Fatal(action)
	}
	for _, step := range steps {
		if run, ok := step.(map[string]any)["run"].(string); ok && strings.Contains(run, "${{") {
			t.Fatal("expression interpolated into shell", run)
		}
	}
	for _, pin := range []string{"main", "v0.2.1", ""} {
		if _, err := RenderControlWorkflow(cfg, pin); err == nil {
			t.Fatal("mutable ref accepted", pin)
		}
	}
	cfg.HumanLogin = "user' || true"
	if _, err := RenderControlWorkflow(cfg, sha(1)); err == nil {
		t.Fatal("injected actor accepted")
	}
}

func TestHotfixRequestIssueDeduplicatesAndRejectsStaleOrForeignMatches(t *testing.T) {
	_, pr := outcomeFixture()
	baseline := Baseline{ID: "production-1", Candidate: candidate(1, "normal")}
	marker := fmt.Sprintf("<!-- mint:hotfix-request:%d:%s -->", pr.Number, pr.MergeSHA)
	for _, scenario := range []string{"create", "existing", "stale", "foreign", "duplicate", "pr-marker"} {
		t.Run(scenario, func(t *testing.T) {
			issues := []map[string]any{}
			issue := map[string]any{"number": 17, "body": marker + "\n<!-- mint:hotfix-baseline:production-1 -->", "user": map[string]string{"login": AutomationLogin}}
			switch scenario {
			case "existing":
				issues = append(issues, issue)
			case "stale":
				issue["body"] = marker + "\n<!-- mint:hotfix-baseline:old -->"
				issues = append(issues, issue)
			case "foreign":
				issue["user"] = map[string]string{"login": "intruder"}
				issues = append(issues, issue)
			case "duplicate":
				other := map[string]any{"number": 18, "body": issue["body"], "user": issue["user"]}
				issues = append(issues, issue, other)
			case "pr-marker":
				issue["pull_request"] = map[string]string{"url": "pull/17"}
				issues = append(issues, issue)
			}
			writes := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "GET" {
					_ = json.NewEncoder(w).Encode(issues)
					return
				}
				if r.Method != "POST" || r.URL.Path != "/repos/owner/repo/issues" {
					t.Error(r.Method, r.URL.Path)
				}
				writes++
				var payload map[string]any
				_ = json.NewDecoder(r.Body).Decode(&payload)
				if !strings.HasPrefix(payload["body"].(string), marker+"\n") || fmt.Sprint(payload["assignees"]) != "[jamesonstone]" {
					t.Error(payload)
				}
				issues = append(issues, map[string]any{"number": 19, "body": payload["body"], "user": map[string]string{"login": AutomationLogin}})
				w.WriteHeader(201)
				_ = json.NewEncoder(w).Encode(map[string]int{"number": 19})
			}))
			defer server.Close()
			c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo", HumanLogin: "jamesonstone"}
			number, err := c.EnsureHotfixRequestIssue(context.Background(), pr, baseline, "Restore login")
			if scenario == "stale" || scenario == "foreign" || scenario == "duplicate" {
				if err == nil || writes != 0 {
					t.Fatal(number, err, writes)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if scenario == "existing" {
				if number != 17 || writes != 0 {
					t.Fatal(number, writes)
				}
			} else if number != 19 || writes != 1 {
				t.Fatal(number, writes)
			}
			before := writes
			again, err := c.EnsureHotfixRequestIssue(context.Background(), pr, baseline, "Restore login")
			if err != nil || again != number || writes != before {
				t.Fatal("duplicate callback created another issue", again, err, writes)
			}
		})
	}
}
