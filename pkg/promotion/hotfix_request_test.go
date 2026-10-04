package promotion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func hotfixRequestFixture(t *testing.T, strategy string) (string, Baseline, PullRequest, []string) {
	t.Helper()
	dir, baseline, _ := hotfixRepo(t)
	gitTest(t, dir, "reset", "--hard", "HEAD^") // Drop fixture fix; keep queued B/C.
	gitTest(t, dir, "checkout", "-b", "source-fix")
	writeTest(t, dir, "production.txt", "fixed\n")
	gitTest(t, dir, "add", "production.txt")
	gitTest(t, dir, "commit", "-m", "hotfix(GH-123): repair login")
	first := gitTest(t, dir, "rev-parse", "HEAD")
	writeTest(t, dir, "urgent.txt", "urgent correction\n")
	gitTest(t, dir, "add", "urgent.txt")
	gitTest(t, dir, "commit", "-m", "fix: complete correction")
	head := gitTest(t, dir, "rev-parse", "HEAD")
	gitTest(t, dir, "checkout", "main")
	writeTest(t, dir, "queued-d.txt", "more queued work\n")
	gitTest(t, dir, "add", "queued-d.txt")
	gitTest(t, dir, "commit", "-m", "feat: queued D")
	switch strategy {
	case "squash":
		gitTest(t, dir, "merge", "--squash", "source-fix")
		gitTest(t, dir, "commit", "-m", "hotfix(GH-123): reviewed squash")
	case "merge":
		gitTest(t, dir, "merge", "--no-ff", "source-fix", "-m", "hotfix(GH-123): reviewed merge")
	case "rebase":
		gitTest(t, dir, "cherry-pick", first, head)
	default:
		t.Fatal("unknown strategy")
	}
	pr := PullRequest{Number: 123, Title: "hotfix(GH-123): repair login", Merged: true, MergeSHA: gitTest(t, dir, "rev-parse", "HEAD")}
	mergedAt := "2026-10-04T00:00:00Z"
	pr.MergedAt = &mergedAt
	pr.Head.SHA, pr.Head.Repo.FullName = head, "owner/repo"
	pr.Base.Ref, pr.Base.Repo.FullName = "main", "owner/repo"
	pr.User.Login = "author"
	return dir, baseline, pr, []string{first, head}
}

func hotfixRequestServer(t *testing.T, pr PullRequest, commits []string, main string, approved bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Error("unexpected mutation", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		switch r.URL.Path {
		case "/repos/owner/repo/pulls/123":
			_ = json.NewEncoder(w).Encode(pr)
		case "/repos/owner/repo/pulls/123/commits":
			items := []map[string]string{}
			for _, sha := range commits {
				items = append(items, map[string]string{"sha": sha})
			}
			_ = json.NewEncoder(w).Encode(items)
		case "/repos/owner/repo/git/ref/heads/main":
			_ = json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": main}})
		case "/repos/owner/repo/commits/" + pr.Head.SHA + "/check-runs":
			_ = json.NewEncoder(w).Encode(map[string]any{"total_count": 1, "check_runs": []map[string]string{{"name": "ci", "status": "completed", "conclusion": "success"}}})
		case "/repos/owner/repo/pulls/123/reviews":
			reviews := []map[string]any{}
			if approved {
				reviews = append(reviews, map[string]any{"state": "APPROVED", "commit_id": pr.Head.SHA, "user": map[string]string{"login": "reviewer"}})
			}
			_ = json.NewEncoder(w).Encode(reviews)
		default:
			t.Error("unexpected request", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestHotfixFixesIsolatesReviewedMergeStrategies(t *testing.T) {
	for _, strategy := range []string{"squash", "merge", "rebase"} {
		t.Run(strategy, func(t *testing.T) {
			dir, baseline, pr, commits := hotfixRequestFixture(t, strategy)
			server := hotfixRequestServer(t, pr, commits, pr.MergeSHA, true)
			defer server.Close()
			client := Client{APIURL: server.URL, Repository: "owner/repo", Token: "test"}
			proof := GitProof{WorkDir: dir}
			got, fixes, err := client.HotfixFixes(context.Background(), Config{Repository: "owner/repo", DefaultBranch: "main", RequiredChecks: []string{"ci"}}, 123, proof)
			if err != nil || got.Title != pr.Title {
				t.Fatal(got, fixes, err)
			}
			if strategy == "merge" && !reflect.DeepEqual(fixes, commits) {
				t.Fatalf("merge fixes = %v, want reviewed source %v", fixes, commits)
			}
			if strategy == "squash" && !reflect.DeepEqual(fixes, []string{pr.MergeSHA}) {
				t.Fatalf("squash fixes = %v", fixes)
			}
			if strategy == "rebase" && len(fixes) != 2 {
				t.Fatalf("rebase fixes = %v", fixes)
			}
			source, err := PrepareHotfix(context.Background(), HotfixOptions{WorkDir: dir, Baseline: baseline, Issue: 456, Fixes: fixes, CommitterName: "Human", CommitterEmail: "human@example.com"})
			if source.WorkDir != "" {
				defer func() { _ = os.RemoveAll(source.WorkDir) }()
			}
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"queued-b.txt", "queued-c.txt", "queued-d.txt"} {
				if _, err := os.Stat(filepath.Join(source.WorkDir, name)); !os.IsNotExist(err) {
					t.Fatalf("unrelated queued change included: %s", name)
				}
			}
			for name, want := range map[string]string{"production.txt": "fixed\n", "urgent.txt": "urgent correction\n"} {
				got, err := os.ReadFile(filepath.Join(source.WorkDir, name))
				if err != nil || string(got) != want {
					t.Fatalf("isolated net diff %s = %q, %v", name, got, err)
				}
			}
		})
	}
}

func TestHotfixFixesRejectsUnreviewedAndAmbiguousSource(t *testing.T) {
	for _, mode := range []string{"unapproved", "foreign", "not-default", "unmerged", "missing-commit", "merge-resolution", "whitespace-resolution"} {
		t.Run(mode, func(t *testing.T) {
			dir, _, pr, commits := hotfixRequestFixture(t, "merge")
			switch mode {
			case "foreign":
				pr.Head.Repo.FullName = "foreign/repo"
			case "not-default":
				pr.Base.Ref = "other"
			case "unmerged":
				pr.Merged = false
			case "missing-commit":
				commits = commits[1:]
			case "whitespace-resolution":
				writeTest(t, dir, "production.txt", "fixed \n")
				gitTest(t, dir, "add", "production.txt")
				gitTest(t, dir, "commit", "--amend", "--no-edit")
				pr.MergeSHA = gitTest(t, dir, "rev-parse", "HEAD")
			case "merge-resolution":
				writeTest(t, dir, "unexpected.txt", "unreviewed merge resolution\n")
				gitTest(t, dir, "add", "unexpected.txt")
				gitTest(t, dir, "commit", "--amend", "--no-edit")
				pr.MergeSHA = gitTest(t, dir, "rev-parse", "HEAD")
			}
			server := hotfixRequestServer(t, pr, commits, pr.MergeSHA, mode != "unapproved")
			defer server.Close()
			client := Client{APIURL: server.URL, Repository: "owner/repo", Token: "test"}
			_, fixes, err := client.HotfixFixes(context.Background(), Config{Repository: "owner/repo", DefaultBranch: "main", RequiredChecks: []string{"ci"}}, 123, GitProof{WorkDir: dir})
			if mode == "missing-commit" {
				// A contiguous suffix alone cannot prove the full PR patch; the
				// merge's complete net diff must still reject it.
				if err == nil || !strings.Contains(err.Error(), "isolate") {
					t.Fatal(fixes, err)
				}
			} else if err == nil {
				t.Fatalf("unsafe hotfix accepted: %v", fixes)
			}
		})
	}
}
