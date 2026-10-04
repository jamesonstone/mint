package promotion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestHotfixConflictHasPushableCompleteRequest(t *testing.T) {
	dir, baseline, _ := hotfixRepo(t)
	writeTest(t, dir, "extra-fix.txt", "first requested correction\n")
	gitTest(t, dir, "add", "extra-fix.txt")
	gitTest(t, dir, "commit", "-m", "fix: first requested correction")
	first := gitTest(t, dir, "rev-parse", "HEAD")
	writeTest(t, dir, "production.txt", "conflicting correction\n")
	gitTest(t, dir, "add", "production.txt")
	gitTest(t, dir, "commit", "-m", "fix: conflicting correction")
	second := gitTest(t, dir, "rev-parse", "HEAD")
	writeTest(t, dir, "later-fix.txt", "later requested correction\n")
	gitTest(t, dir, "add", "later-fix.txt")
	gitTest(t, dir, "commit", "-m", "fix: later requested correction")
	third := gitTest(t, dir, "rev-parse", "HEAD")
	fixes := []string{first, second, third}
	source, err := PrepareHotfix(context.Background(), HotfixOptions{WorkDir: dir, Baseline: baseline, Issue: 789, Fixes: fixes, CommitterName: "Human", CommitterEmail: "human@example.com"})
	if source.WorkDir != "" {
		defer func() { _ = os.RemoveAll(source.WorkDir) }()
	}
	if err == nil || !source.Conflict || !shaPattern.MatchString(source.SourceSHA) {
		t.Fatal(source, err)
	}
	if gitTest(t, source.WorkDir, "rev-parse", "HEAD^") != baseline.Candidate.SourceSHA {
		t.Fatal("recoverable request did not start from verified production")
	}
	if got := gitTest(t, source.WorkDir, "diff", "--name-only", "HEAD^", "HEAD"); got != ".mint/hotfix.json" {
		t.Fatal("partial fixes entered pushable HEAD", got)
	}
	if got := gitTest(t, source.WorkDir, "diff", "--name-only", "--diff-filter=U"); got != "production.txt" {
		t.Fatal("local conflict not preserved", got)
	}
	data := gitTest(t, source.WorkDir, "show", "HEAD:.mint/hotfix.json")
	var request struct {
		Fixes    []string `json:"fixes"`
		PatchIDs []string `json:"patch_ids"`
	}
	if err := json.Unmarshal([]byte(data), &request); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(request.Fixes, fixes) || len(request.PatchIDs) != len(fixes) {
		t.Fatalf("incomplete durable request: %#v", request)
	}
	proof := GitProof{Context: context.Background(), WorkDir: dir}
	for i, fix := range fixes {
		id, err := proof.PatchID(fix)
		if err != nil || id != request.PatchIDs[i] {
			t.Fatal("request patch identity lost", fix, err)
		}
	}
	// Git pushes committed HEAD independently of the conflicted local index.
	remote := t.TempDir()
	gitTest(t, remote, "init", "--bare")
	gitTest(t, source.WorkDir, "push", remote, "HEAD:refs/heads/"+source.Branch)
	if got := gitTest(t, remote, "rev-parse", "refs/heads/"+source.Branch); got != source.SourceSHA {
		t.Fatal("conflict request could not survive runner loss", got)
	}
}

func TestConflictSourcePRContainsRemoteRecoveryInstructions(t *testing.T) {
	source := HotfixSource{Branch: "GH-789", BaselineID: "production-a", SourceSHA: sha(2), Conflict: true, Fixes: []string{sha(3), sha(4)}}
	baseline := Baseline{ID: source.BaselineID, Candidate: Candidate{Version: "v1.2.3", SourceSHA: sha(1)}}
	body := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/repos/owner/repo/git/ref/heads/mint-hotfix-base/GH-789":
			_ = json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": sha(1)}})
		case r.Method == "GET" && r.URL.Path == "/repos/owner/repo/git/ref/heads/GH-789":
			_ = json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": sha(2)}})
		case r.Method == "GET" && r.URL.Path == "/repos/owner/repo/pulls":
			_, _ = w.Write([]byte("[]"))
		case r.Method == "POST" && r.URL.Path == "/repos/owner/repo/pulls":
			var request struct{ Body string }
			_ = json.NewDecoder(r.Body).Decode(&request)
			body = request.Body
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(PullRequest{Number: 321})
		case r.Method == "POST" && r.URL.Path == "/repos/owner/repo/issues/321/assignees":
			w.WriteHeader(http.StatusCreated)
		default:
			t.Error("unexpected request", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	client := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo", HumanLogin: "human"}
	pr, err := client.PublishHotfixSource(context.Background(), source, baseline, 789)
	if err != nil || pr.Number != 321 {
		t.Fatal(pr, err)
	}
	for _, want := range []string{"https://github.com/owner/repo/tree/GH-789", "apply all requested fixes", "Preserve `.mint/hotfix.json`", "CI will reject the metadata-only branch", sha(3), sha(4)} {
		if !strings.Contains(body, want) {
			t.Fatalf("missing recovery instruction %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "mint-hotfix-") || strings.Contains(body, "/tmp/") {
		t.Fatal("runner-local paths leaked into recovery UX", body)
	}
}
