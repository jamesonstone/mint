package cli

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/mint/pkg/promotion"
)

func TestTicketHotfixVersionRunsPreparationThroughNativeAllocation(t *testing.T) {
	for _, scenario := range []struct {
		name, ticket, component string
		issue                   int
	}{
		{name: "legacy", issue: 7},
		{name: "linear", ticket: "https://linear.app/acme/issue/ENG-42"},
		{name: "component-linear", ticket: "https://linear.app/acme/issue/ENG-42", component: "backend"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			op, source, origin, git, serve := ticketVersionFixture(t, scenario.issue, scenario.ticket, scenario.component)
			defer source()
			server := httptest.NewServer(serve)
			defer server.Close()
			op.client.APIURL = server.URL
			if _, _, err := op.review(context.Background()); err != nil {
				t.Fatal("source CI review rejected", err)
			}
			value, changed, err := op.versionHotfix(context.Background())
			if err != nil || changed {
				t.Fatal(value, changed, err)
			}
			selection := value.(promotion.MainVersionSelection)
			expectedVersion := "v1.2.4"
			if scenario.component != "" {
				expectedVersion = scenario.component + "/" + expectedVersion
			}
			if selection.Version != expectedVersion {
				t.Fatal(selection)
			}
			prefix := ""
			if scenario.component != "" {
				prefix = scenario.component + "/"
			}
			if got := git(origin, "rev-parse", prefix+"v1.2.4^{}"); got != selection.SourceSHA {
				t.Fatal("wrong published fixture tag", got, selection)
			}
			// Reruns reuse the same immutable tag rather than allocate another version.
			again, _, err := op.versionHotfix(context.Background())
			if err != nil || again.(promotion.MainVersionSelection).Version != selection.Version {
				t.Fatal(again, err)
			}
		})
	}
}

func ticketVersionFixture(t *testing.T, issue int, ticket, component string) (*productionOperation, func(), string, func(string, ...string) string, http.HandlerFunc) {
	t.Helper()
	origin := t.TempDir()
	git := func(dir string, args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		data, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, data)
		}
		return strings.TrimSpace(string(data))
	}
	git(origin, "init", "-b", "main")
	git(origin, "config", "user.name", "Human")
	git(origin, "config", "user.email", "human@example.com")
	file := filepath.Join(origin, "application.txt")
	os.WriteFile(file, []byte("before\n"), 0644)
	git(origin, "add", "application.txt")
	git(origin, "commit", "-m", "fix: baseline")
	base := git(origin, "rev-parse", "HEAD")
	prefix := ""
	scope := promotion.Component{}
	if component != "" {
		prefix = component + "/"
		scope = promotion.Component{Name: component, TagPrefix: prefix}
	}
	git(origin, "tag", prefix+"v1.2.3")
	os.WriteFile(file, []byte("after\n"), 0644)
	git(origin, "add", "application.txt")
	git(origin, "commit", "-m", "fix: restore application")
	fix := git(origin, "rev-parse", "HEAD")
	baseline := promotion.Baseline{ID: "verified", Candidate: promotion.Candidate{SourceSHA: base, Version: prefix + "v1.2.3"}}
	source, err := promotion.PrepareHotfix(context.Background(), promotion.HotfixOptions{WorkDir: origin, Baseline: baseline, Issue: issue, TicketURL: ticket, Fixes: []string{fix}, CommitterName: "Human", CommitterEmail: "human@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	cleanup := func() { os.RemoveAll(source.WorkDir) }
	metadata, err := os.ReadFile(filepath.Join(source.WorkDir, ".mint/hotfix.json"))
	if err != nil {
		cleanup()
		t.Fatal(err)
	}
	id := source.Branch
	marker := scope.HotfixSourceMarkerPrefix() + id + ":verified -->"
	baseRef := strings.TrimSuffix(scope.HotfixBasePrefix(), "GH-") + id
	pr := map[string]any{"number": 9, "merged": true, "merge_commit_sha": source.SourceSHA, "body": marker, "user": map[string]string{"login": "human"}, "head": map[string]any{"sha": source.SourceSHA, "ref": id, "repo": map[string]string{"full_name": "owner/repo"}}, "base": map[string]any{"ref": baseRef, "repo": map[string]string{"full_name": "owner/repo"}}}
	serve := func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/repos/owner/repo/pulls/9":
			json.NewEncoder(w).Encode(pr)
		case strings.HasSuffix(r.URL.Path, "/check-runs"):
			json.NewEncoder(w).Encode(map[string]any{"total_count": 1, "check_runs": []map[string]string{{"name": "native", "status": "completed", "conclusion": "success"}}})
		case strings.HasSuffix(r.URL.Path, "/reviews"):
			json.NewEncoder(w).Encode([]map[string]any{{"state": "APPROVED", "commit_id": source.SourceSHA, "user": map[string]string{"login": "reviewer"}}})
		case strings.Contains(r.URL.Path, "/commits/") && strings.HasSuffix(r.URL.Path, "/pulls"):
			_, _ = w.Write([]byte("[]"))
		case strings.Contains(r.URL.Path, "/contents/.mint/hotfix.json"):
			json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "content": base64.StdEncoding.EncodeToString(metadata)})
		default:
			t.Error("unexpected request", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}
	op := &productionOperation{config: promotion.Config{Component: scope, RequiredChecks: []string{"native"}}, client: promotion.Client{Repository: "owner/repo", HumanLogin: "human", Token: "test", Component: scope}, snapshot: promotion.JournalSnapshot{State: promotion.State{Baseline: &baseline}}, flags: productionFlags{PR: 9}, proof: promotion.GitProof{Context: context.Background(), WorkDir: source.WorkDir, DefaultBranch: "main"}}
	return op, cleanup, origin, git, serve
}

func TestTicketHotfixVersionStillRequiresChecksAndIndependentReview(t *testing.T) {
	for _, gate := range []string{"check", "review"} {
		t.Run(gate, func(t *testing.T) {
			op, cleanup, origin, git, serve := ticketVersionFixture(t, 0, "https://linear.app/acme/issue/ENG-42", "")
			defer cleanup()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if gate == "check" && strings.HasSuffix(r.URL.Path, "/check-runs") {
					json.NewEncoder(w).Encode(map[string]any{"total_count": 1, "check_runs": []map[string]string{{"name": "native", "status": "in_progress"}}})
					return
				}
				if gate == "review" && strings.HasSuffix(r.URL.Path, "/reviews") {
					w.Write([]byte("[]"))
					return
				}
				serve(w, r)
			}))
			defer server.Close()
			op.client.APIURL = server.URL
			if _, _, err := op.versionHotfix(context.Background()); err == nil {
				t.Fatal("missing", gate, "accepted")
			}
			if git(origin, "tag", "--list", "v1.2.4") != "" {
				t.Fatal("version published before required gate")
			}
		})
	}
}
