package changelog

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratePrependsReleaseBlockFromConventionalCommits(t *testing.T) {
	repo := newTestRepo(t)
	repo.commit(t, "docs: initial docs", "", "2024-01-01T00:00:00Z")
	repo.tag(t, "v1.0.0")

	featHash := repo.commit(t, "feat(cli): add widget (#12)", "", "2024-01-02T00:00:00Z")
	fixHash := repo.commit(t, "fix: repair bug", "Fixes #34", "2024-01-03T00:00:00Z")
	breakingHash := repo.commit(t, "perf(api)!: speed up API", "BREAKING CHANGE: API contract changed\n\nCloses #56", "2024-01-04T00:00:00Z")
	repo.commit(t, "chore: update generated files", "", "2024-01-05T00:00:00Z")
	repo.tag(t, "v1.1.0")

	outputPath := filepath.Join(repo.dir, "CHANGELOG.md")
	result, err := Generate(context.Background(), Options{
		PrevTag:    "v1.0.0",
		CurrentTag: "v1.1.0",
		RepoOwner:  "jamesonstone",
		RepoName:   "kit",
		OutputFile: outputPath,
		WorkDir:    repo.dir,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if result.Path != outputPath {
		t.Fatalf("Generate() path = %q, want %q", result.Path, outputPath)
	}
	if result.Tag != "v1.1.0" || result.Version != "1.1.0" {
		t.Fatalf("Generate() result tag/version = %q/%q", result.Tag, result.Version)
	}
	if result.CommitCount != 3 || result.BreakingCount != 1 {
		t.Fatalf("Generate() counts = %d/%d, want 3/1", result.CommitCount, result.BreakingCount)
	}

	content := readFile(t, outputPath)
	if !strings.HasPrefix(content, "# Changelog\n\n") {
		t.Fatalf("CHANGELOG should start with top-level heading:\n%s", content)
	}

	wantHeader := "## [1.1.0](https://github.com/jamesonstone/kit/releases/tag/v1.1.0) - 2024-01-05"
	if !strings.Contains(content, wantHeader) {
		t.Fatalf("CHANGELOG missing header %q:\n%s", wantHeader, content)
	}

	breakingLine := "- **api:** speed up API\n  ([#56](https://github.com/jamesonstone/kit/issues/56))\n  ([" + breakingHash + "](https://github.com/jamesonstone/kit/commit/" + breakingHash + "))"
	featureLine := "- **cli:** add widget\n  ([#12](https://github.com/jamesonstone/kit/issues/12))\n  ([" + featHash + "](https://github.com/jamesonstone/kit/commit/" + featHash + "))"
	fixLine := "- repair bug\n  ([#34](https://github.com/jamesonstone/kit/issues/34))\n  ([" + fixHash + "](https://github.com/jamesonstone/kit/commit/" + fixHash + "))"
	for _, want := range []string{breakingLine, featureLine, fixLine} {
		if !strings.Contains(content, want) {
			t.Fatalf("CHANGELOG missing %q:\n%s", want, content)
		}
	}

	if strings.Contains(content, "update generated files") || strings.Contains(content, "initial docs") {
		t.Fatalf("CHANGELOG contains excluded commit:\n%s", content)
	}
	if strings.Contains(content, "add widget (#12)") {
		t.Fatalf("CHANGELOG should not duplicate subject issue suffix:\n%s", content)
	}
	if strings.Index(content, "### breaking changes") > strings.Index(content, "### features") {
		t.Fatalf("breaking changes section should render before features:\n%s", content)
	}
}

func TestGeneratePreservesExistingTopLevelHeadingWithoutTrailingNewline(t *testing.T) {
	repo := newTestRepo(t)
	repo.commit(t, "feat: first release", "", "2024-01-01T00:00:00Z")
	repo.tag(t, "v1.0.0")

	outputPath := filepath.Join(repo.dir, "CHANGELOG.md")
	if err := os.WriteFile(outputPath, []byte("# Changelog"), 0o644); err != nil {
		t.Fatalf("write CHANGELOG fixture: %v", err)
	}

	if _, err := Generate(context.Background(), Options{
		CurrentTag: "v1.0.0",
		RepoOwner:  "jamesonstone",
		RepoName:   "mint",
		OutputFile: outputPath,
		WorkDir:    repo.dir,
	}); err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	content := readFile(t, outputPath)
	if !strings.HasPrefix(content, "# Changelog\n\n## [1.0.0]") {
		t.Fatalf("CHANGELOG should preserve top-level heading first:\n%s", content)
	}
}

func TestGenerateFailsWhenVersionAlreadyExists(t *testing.T) {
	repo := newTestRepo(t)
	repo.commit(t, "feat: first release", "", "2024-01-01T00:00:00Z")
	repo.tag(t, "v1.0.0")

	outputPath := filepath.Join(repo.dir, "CHANGELOG.md")
	if _, err := Generate(context.Background(), Options{
		CurrentTag: "v1.0.0",
		RepoOwner:  "jamesonstone",
		RepoName:   "mint",
		OutputFile: outputPath,
		WorkDir:    repo.dir,
	}); err != nil {
		t.Fatalf("Generate() first call error = %v", err)
	}

	original := readFile(t, outputPath)
	_, err := Generate(context.Background(), Options{
		CurrentTag: "v1.0.0",
		RepoOwner:  "jamesonstone",
		RepoName:   "mint",
		OutputFile: outputPath,
		WorkDir:    repo.dir,
	})
	if err == nil || err.Error() != "version 1.0.0 already in CHANGELOG" {
		t.Fatalf("Generate() duplicate error = %v", err)
	}
	if got := readFile(t, outputPath); got != original {
		t.Fatalf("duplicate generation changed CHANGELOG:\n%s", got)
	}
}

func TestGenerateFailsForMissingTagAndEmptyRange(t *testing.T) {
	repo := newTestRepo(t)
	repo.commit(t, "feat: first release", "", "2024-01-01T00:00:00Z")
	repo.tag(t, "v1.0.0")

	_, err := Generate(context.Background(), Options{
		CurrentTag: "v2.0.0",
		RepoOwner:  "jamesonstone",
		RepoName:   "mint",
		OutputFile: filepath.Join(repo.dir, "CHANGELOG.md"),
		WorkDir:    repo.dir,
	})
	if err == nil || err.Error() != "tag not found: v2.0.0" {
		t.Fatalf("Generate() missing tag error = %v", err)
	}

	_, err = Generate(context.Background(), Options{
		PrevTag:    "v1.0.0",
		CurrentTag: "v1.0.0",
		RepoOwner:  "jamesonstone",
		RepoName:   "mint",
		OutputFile: filepath.Join(repo.dir, "CHANGELOG.md"),
		WorkDir:    repo.dir,
	})
	if err == nil || err.Error() != "no commits in range v1.0.0..v1.0.0" {
		t.Fatalf("Generate() empty range error = %v", err)
	}
}

func TestGenerateWarnsAndSkipsNonConventionalCommits(t *testing.T) {
	repo := newTestRepo(t)
	repo.commit(t, "feat: first release", "", "2024-01-01T00:00:00Z")
	repo.tag(t, "v1.0.0")
	repo.commit(t, "Update something manually", "", "2024-01-02T00:00:00Z")
	repo.commit(t, "fix: keep conventional commit", "", "2024-01-03T00:00:00Z")
	repo.tag(t, "v1.1.0")

	var warnings bytes.Buffer
	_, err := Generate(context.Background(), Options{
		PrevTag:       "v1.0.0",
		CurrentTag:    "v1.1.0",
		RepoOwner:     "jamesonstone",
		RepoName:      "mint",
		OutputFile:    filepath.Join(repo.dir, "CHANGELOG.md"),
		WorkDir:       repo.dir,
		WarningWriter: &warnings,
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if !strings.Contains(warnings.String(), "warning: skipping non-conventional commit") {
		t.Fatalf("expected non-conventional warning, got %q", warnings.String())
	}
}

func TestGenerateFailsForUnparseableChangelog(t *testing.T) {
	repo := newTestRepo(t)
	repo.commit(t, "feat: first release", "", "2024-01-01T00:00:00Z")
	repo.tag(t, "v1.0.0")

	outputPath := filepath.Join(repo.dir, "CHANGELOG.md")
	if err := os.WriteFile(outputPath, []byte("## unreleased\n"), 0o644); err != nil {
		t.Fatalf("write CHANGELOG fixture: %v", err)
	}

	_, err := Generate(context.Background(), Options{
		CurrentTag: "v1.0.0",
		RepoOwner:  "jamesonstone",
		RepoName:   "mint",
		OutputFile: outputPath,
		WorkDir:    repo.dir,
	})
	if err == nil || err.Error() != "cannot parse CHANGELOG.md" {
		t.Fatalf("Generate() unparseable changelog error = %v", err)
	}
}

type testRepo struct {
	dir     string
	counter int
}

func newTestRepo(t *testing.T) *testRepo {
	t.Helper()

	repo := &testRepo{dir: t.TempDir()}
	repo.git(t, nil, "init")
	repo.git(t, nil, "config", "user.name", "Mint Test")
	repo.git(t, nil, "config", "user.email", "mint@example.com")
	return repo
}

func (repo *testRepo) commit(t *testing.T, subject string, body string, date string) string {
	t.Helper()

	repo.counter++
	path := filepath.Join(repo.dir, fmt.Sprintf("file-%02d.txt", repo.counter))
	if err := os.WriteFile(path, []byte(subject+"\n"), 0o644); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}
	repo.git(t, nil, "add", ".")

	args := []string{"commit", "-m", subject}
	if body != "" {
		args = append(args, "-m", body)
	}
	env := []string{
		"GIT_AUTHOR_DATE=" + date,
		"GIT_COMMITTER_DATE=" + date,
	}
	repo.git(t, env, args...)
	return strings.TrimSpace(repo.gitOutput(t, nil, "rev-parse", "--short=7", "HEAD"))
}

func (repo *testRepo) tag(t *testing.T, tag string) {
	t.Helper()
	repo.git(t, nil, "tag", tag)
}

func (repo *testRepo) git(t *testing.T, env []string, args ...string) {
	t.Helper()
	_ = repo.gitOutput(t, env, args...)
}

func (repo *testRepo) gitOutput(t *testing.T, env []string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = repo.dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
	return string(out)
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
