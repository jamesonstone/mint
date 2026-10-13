package changelog

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// commitAt writes one file at relPath, commits it, and returns the short hash.
func (repo *testRepo) commitAt(t *testing.T, relPath string, subject string, date string) string {
	t.Helper()
	full := filepath.Join(repo.dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(subject+"\n"), 0o644); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}
	repo.git(t, nil, "add", ".")
	repo.git(t, []string{"GIT_AUTHOR_DATE=" + date, "GIT_COMMITTER_DATE=" + date}, "commit", "-m", subject)
	return strings.TrimSpace(repo.gitOutput(t, nil, "rev-parse", "--short=7", "HEAD"))
}

func TestGenerateScopedComponentChangelog(t *testing.T) {
	repo := newTestRepo(t)
	repo.commitAt(t, "backend/main.go", "feat: first api", "2024-01-01T00:00:00Z")
	repo.tag(t, "backend/v1.0.0")
	backendHash := repo.commitAt(t, "backend/orders.go", "feat(orders): place orders", "2024-01-02T00:00:00Z")
	repo.commitAt(t, "frontend/app.ts", "fix(ui): render home", "2024-01-03T00:00:00Z")

	outputPath := filepath.Join(repo.dir, "backend", "CHANGELOG.md")
	result, err := Generate(context.Background(), Options{
		PrevTag:    "backend/v1.0.0",
		CurrentTag: "backend/v1.1.0",
		CurrentRef: "HEAD",
		RepoOwner:  "acme",
		RepoName:   "app",
		OutputFile: outputPath,
		WorkDir:    repo.dir,
		TagPrefix:  "backend/",
		Paths:      []string{"backend"},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if result.Version != "1.1.0" || result.Tag != "backend/v1.1.0" || result.CommitCount != 1 {
		t.Fatalf("Generate() = %+v", result)
	}

	content := readFile(t, outputPath)
	wantHeader := "## [1.1.0](https://github.com/acme/app/releases/tag/backend/v1.1.0) - "
	if !strings.Contains(content, wantHeader) {
		t.Fatalf("CHANGELOG missing header %q:\n%s", wantHeader, content)
	}
	if !strings.Contains(content, backendHash) || strings.Contains(content, "render home") {
		t.Fatalf("CHANGELOG is not path-scoped:\n%s", content)
	}
}

func TestGenerateScopedRejectsUnprefixedTags(t *testing.T) {
	repo := newTestRepo(t)
	repo.commitAt(t, "backend/main.go", "feat: first api", "2024-01-01T00:00:00Z")
	repo.tag(t, "v1.0.0")
	repo.commitAt(t, "backend/orders.go", "feat: orders", "2024-01-02T00:00:00Z")

	base := Options{CurrentRef: "HEAD", RepoOwner: "acme", RepoName: "app", WorkDir: repo.dir,
		OutputFile: filepath.Join(repo.dir, "CHANGELOG.md"), TagPrefix: "backend/"}

	current := base
	current.CurrentTag = "v1.1.0"
	if _, err := Generate(context.Background(), current); err == nil || !strings.Contains(err.Error(), "prefix backend/") {
		t.Fatalf("Generate(unprefixed current) error = %v", err)
	}

	prev := base
	prev.CurrentTag = "backend/v1.1.0"
	prev.PrevTag = "v1.0.0"
	if _, err := Generate(context.Background(), prev); err == nil || !strings.Contains(err.Error(), "prefix backend/") {
		t.Fatalf("Generate(unprefixed previous) error = %v", err)
	}
}
