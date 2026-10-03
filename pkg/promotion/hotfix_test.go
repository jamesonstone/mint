package promotion

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitTest(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	data, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, data)
	}
	return strings.TrimSpace(string(data))
}
func writeTest(t *testing.T, dir, path, value string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, path), []byte(value), 0644); err != nil {
		t.Fatal(err)
	}
}
func hotfixRepo(t *testing.T) (string, Baseline, string) {
	t.Helper()
	dir := t.TempDir()
	gitTest(t, dir, "init", "-b", "main")
	gitTest(t, dir, "config", "user.name", "Human")
	gitTest(t, dir, "config", "user.email", "human@example.com")
	writeTest(t, dir, "production.txt", "bug\n")
	gitTest(t, dir, "add", "production.txt")
	gitTest(t, dir, "commit", "-m", "fix: baseline")
	base := gitTest(t, dir, "rev-parse", "HEAD")
	baseline := Baseline{ID: "production-a", Candidate: candidate(0, "normal")}
	baseline.Candidate.SourceSHA = base
	for _, path := range []string{"queued-b.txt", "queued-c.txt"} {
		writeTest(t, dir, path, "queued feature\n")
		gitTest(t, dir, "add", path)
		gitTest(t, dir, "commit", "-m", "feat: queued feature")
	}
	writeTest(t, dir, "production.txt", "fixed\n")
	gitTest(t, dir, "add", "production.txt")
	gitTest(t, dir, "commit", "-m", "fix: isolated correction")
	fix := gitTest(t, dir, "rev-parse", "HEAD")
	return dir, baseline, fix
}
func TestHotfixFromProductionExcludesQueuedFeatures(t *testing.T) {
	dir, base, fix := hotfixRepo(t)
	source, err := PrepareHotfix(context.Background(), HotfixOptions{WorkDir: dir, Baseline: base, Issue: 123, Fixes: []string{fix}, HumanName: "Human", HumanEmail: "human@example.com"})
	if source.WorkDir != "" {
		defer func() { _ = os.RemoveAll(source.WorkDir) }()
	}
	if err != nil {
		t.Fatal(err)
	}
	for _, queued := range []string{"queued-b.txt", "queued-c.txt"} {
		if _, err := os.Stat(filepath.Join(source.WorkDir, queued)); !os.IsNotExist(err) {
			t.Fatal("queued change entered hotfix", queued)
		}
	}
	data, _ := os.ReadFile(filepath.Join(source.WorkDir, "production.txt"))
	if string(data) != "fixed\n" {
		t.Fatal("fix missing")
	}
	if gitTest(t, source.WorkDir, "rev-parse", "HEAD^") != base.Candidate.SourceSHA {
		t.Fatal("source did not start at production")
	}
	proof := GitProof{Context: context.Background(), WorkDir: dir}
	original, err := proof.PatchID(fix)
	if err != nil || len(source.PatchIDs) != 1 || source.PatchIDs[0] != original {
		t.Fatal("original provenance lost", err)
	}
	if gitTest(t, dir, "branch", "--show-current") != "main" {
		t.Fatal("user checkout changed")
	}
}
func TestHotfixConflictPreservesRecoverableClone(t *testing.T) {
	dir, base, _ := hotfixRepo(t)
	writeTest(t, dir, "production.txt", "another change\n")
	gitTest(t, dir, "add", "production.txt")
	gitTest(t, dir, "commit", "-m", "fix: conflicting fix")
	fix := gitTest(t, dir, "rev-parse", "HEAD")
	source, err := PrepareHotfix(context.Background(), HotfixOptions{WorkDir: dir, Baseline: base, Issue: 124, Fixes: []string{fix}, HumanName: "Human", HumanEmail: "human@example.com"})
	if source.WorkDir != "" {
		defer func() { _ = os.RemoveAll(source.WorkDir) }()
	}
	if err == nil || source.WorkDir == "" {
		t.Fatal("conflict was masked")
	}
	if !strings.Contains(err.Error(), source.WorkDir) {
		t.Fatal("recovery checkout not identified")
	}
}
func TestNewAuthoredHotfixStartsAtProduction(t *testing.T) {
	dir, base, _ := hotfixRepo(t)
	source, err := PrepareHotfix(context.Background(), HotfixOptions{WorkDir: dir, Baseline: base, Issue: 125, HumanName: "Human", HumanEmail: "human@example.com"})
	if source.WorkDir != "" {
		defer func() { _ = os.RemoveAll(source.WorkDir) }()
	}
	if err != nil {
		t.Fatal(err)
	}
	if gitTest(t, source.WorkDir, "rev-parse", "HEAD^") != base.Candidate.SourceSHA {
		t.Fatal("new authored hotfix started from main")
	}
}

func TestMergedAuthoredHotfixHasActualApplicationPatch(t *testing.T) {
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
	git("init", "-b", "main")
	git("config", "user.name", "Human")
	git("config", "user.email", "human@example.com")
	if err := os.WriteFile(filepath.Join(dir, "code.txt"), []byte("production\n"), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", "code.txt")
	git("commit", "-m", "fix: production")
	git("checkout", "-b", "hotfix")
	if err := os.WriteFile(filepath.Join(dir, "code.txt"), []byte("production\nfix\n"), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", "code.txt")
	git("commit", "-m", "fix: isolated code")
	original := git("rev-parse", "HEAD")
	git("checkout", "main")
	git("merge", "--no-ff", "hotfix", "-m", "fix: reviewed authored hotfix")
	merged := git("rev-parse", "HEAD")
	proof := GitProof{Context: context.Background(), WorkDir: dir}
	before, err := proof.PatchID(original)
	if err != nil {
		t.Fatal(err)
	}
	after, err := proof.PatchID(merged)
	if err != nil || after != before {
		t.Fatalf("merged hotfix lost patch provenance: %s != %s: %v", before, after, err)
	}
}
