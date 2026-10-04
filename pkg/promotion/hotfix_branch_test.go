package promotion

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthoredHotfixIsolationUsesConfiguredDefaultBranch(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git fixture: %v %s", err, output)
		}
		return strings.TrimSpace(string(output))
	}
	git("init", "-b", "trunk")
	git("config", "user.name", "Human")
	git("config", "user.email", "human@example.com")
	file := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(file, []byte("production"), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", "source.txt")
	git("commit", "-m", "feat: production")
	base := git("rev-parse", "HEAD")
	if err := os.WriteFile(file, []byte("queued"), 0600); err != nil {
		t.Fatal(err)
	}
	git("commit", "-am", "feat: queued change")
	queued := git("rev-parse", "HEAD")
	git("update-ref", "refs/remotes/origin/trunk", queued)
	git("checkout", "-b", "production-fix", base)
	if err := os.WriteFile(file, []byte("production fix"), 0600); err != nil {
		t.Fatal(err)
	}
	git("commit", "-am", "fix: isolated correction")
	isolated := git("rev-parse", "HEAD")
	metadata, err := json.Marshal(map[string]any{"baseline_id": "production-1", "baseline_sha": base, "fixes": []string{}, "patch_ids": []string{}})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/contents/.mint/hotfix.json") {
			t.Error("unexpected request", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"content": base64.StdEncoding.EncodeToString(metadata), "encoding": "base64"})
	}))
	defer server.Close()
	client := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
	state := NewState("owner/repo", "production")
	state.Baseline = &Baseline{ID: "production-1", Candidate: Candidate{SourceSHA: base}}
	proof := GitProof{Context: t.Context(), WorkDir: dir, DefaultBranch: "trunk"}
	candidate := Candidate{Kind: "hotfix", SourceSHA: isolated, BaselineID: state.Baseline.ID, Version: "v1.0.1"}
	changes, err := client.CollectChanges(t.Context(), proof, state, candidate, nil)
	if err != nil || len(changes) != 1 {
		t.Fatalf("custom branch isolated fix rejected: %v %v", changes, err)
	}
	candidate.SourceSHA = queued
	if _, err := client.CollectChanges(t.Context(), proof, state, candidate, nil); err == nil {
		t.Fatal("queued default branch accepted as isolated hotfix")
	}
	proof.DefaultBranch = ""
	candidate.SourceSHA = isolated
	if _, err := client.CollectChanges(t.Context(), proof, state, candidate, nil); err == nil {
		t.Fatal("unbound default branch accepted")
	}
}
