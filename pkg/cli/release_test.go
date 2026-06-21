package cli

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestReleaseHelpIncludesSubcommands(t *testing.T) {
	var output bytes.Buffer
	releaseCmd.SetOut(&output)
	defer releaseCmd.SetOut(nil)

	if err := releaseCmd.Help(); err != nil {
		t.Fatalf("release help error = %v", err)
	}

	help := output.String()
	for _, want := range []string{"resolve", "tag", "github", "publish", "workflow", "Resolve, tag, and publish release state"} {
		if !strings.Contains(help, want) {
			t.Fatalf("release help missing %q:\n%s", want, help)
		}
	}
}

func TestRunReleaseResolvePrintsVersionAndWritesGitHubOutput(t *testing.T) {
	repo := newCLITestRepo(t)
	repo.commit(t, "feat: first release", "", "2024-01-01T00:00:00Z")
	t.Chdir(repo.dir)

	outputPath := filepath.Join(t.TempDir(), "github-output")
	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.SetOut(&stdout)

	err := runReleaseResolve(cmd, releaseResolveFlags{
		commitish:    "HEAD",
		githubOutput: outputPath,
	})
	if err != nil {
		t.Fatalf("runReleaseResolve() error = %v", err)
	}

	if got := strings.TrimSpace(stdout.String()); got != "v0.1.0" {
		t.Fatalf("stdout = %q, want v0.1.0", got)
	}

	output := readCLITestFile(t, outputPath)
	for _, want := range []string{
		"version_tag=v0.1.0",
		"version_bump=minor",
		"needs_git_tag=true",
		"commit_count=1",
		"release_notes<<",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("GitHub output missing %q:\n%s", want, output)
		}
	}
}

func TestRunReleaseWorkflowPrintsWorkflow(t *testing.T) {
	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.SetOut(&stdout)

	err := runReleaseWorkflow(cmd, releaseWorkflowFlags{
		images: []string{"name=api,uri=ghcr.io/jamesonstone/mint-api,dockerfile=Dockerfile.api"},
	})
	if err != nil {
		t.Fatalf("runReleaseWorkflow() error = %v", err)
	}

	workflow := stdout.String()
	for _, want := range []string{
		"name: Release Publish",
		"command: release-resolve",
		"docker buildx build",
		"ghcr.io/jamesonstone/mint-api",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("workflow output missing %q:\n%s", want, workflow)
		}
	}
}

func TestRunReleaseTagCreatesTagAndWritesGitHubOutput(t *testing.T) {
	repo := newCLITestRepo(t)
	repo.commit(t, "feat: tag release", "", "2024-01-01T00:00:00Z")
	target := repo.revParse(t, "HEAD")
	t.Chdir(repo.dir)

	notesPath := filepath.Join(t.TempDir(), "notes.md")
	if err := os.WriteFile(notesPath, []byte("Release notes\n"), 0o644); err != nil {
		t.Fatalf("write notes: %v", err)
	}
	outputPath := filepath.Join(t.TempDir(), "github-output")

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.SetOut(&stdout)

	err := runReleaseTag(cmd, releaseTagFlags{
		tag:          "v1.2.3",
		target:       "HEAD",
		notesFile:    notesPath,
		push:         false,
		githubOutput: outputPath,
	})
	if err != nil {
		t.Fatalf("runReleaseTag() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "created Git tag v1.2.3") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	output := readCLITestFile(t, outputPath)
	for _, want := range []string{
		"tag_name=v1.2.3",
		"tag_target_sha=" + target,
		"tag_created=true",
		"tag_reused=false",
		"tag_pushed=false",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("GitHub output missing %q:\n%s", want, output)
		}
	}
}

func TestRunReleaseGitHubCreatesRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/jamesonstone/mint/releases/tags/v1.2.3":
			http.NotFound(w, r)
		case r.Method == http.MethodPost && r.URL.Path == "/repos/jamesonstone/mint/releases":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"tag_name":"v1.2.3","html_url":"https://github.com/jamesonstone/mint/releases/tag/v1.2.3"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	notesPath := filepath.Join(t.TempDir(), "notes.md")
	if err := os.WriteFile(notesPath, []byte("Release notes\n"), 0o644); err != nil {
		t.Fatalf("write notes: %v", err)
	}
	outputPath := filepath.Join(t.TempDir(), "github-output")
	t.Setenv("MINT_TEST_GITHUB_TOKEN", "test-token")

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.SetOut(&stdout)

	err := runReleaseGitHub(cmd, releaseGitHubFlags{
		owner:        "jamesonstone",
		repo:         "mint",
		tag:          "v1.2.3",
		target:       "abc123",
		title:        "Mint v1.2.3",
		notesFile:    notesPath,
		tokenEnv:     "MINT_TEST_GITHUB_TOKEN",
		apiURL:       server.URL,
		githubOutput: outputPath,
	})
	if err != nil {
		t.Fatalf("runReleaseGitHub() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "created GitHub release v1.2.3") {
		t.Fatalf("stdout = %q", stdout.String())
	}

	output := readCLITestFile(t, outputPath)
	for _, want := range []string{
		"release_tag=v1.2.3",
		"release_url=https://github.com/jamesonstone/mint/releases/tag/v1.2.3",
		"release_created=true",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("GitHub output missing %q:\n%s", want, output)
		}
	}
}

func TestRunReleasePublishCreatesTagReleaseAndWritesGitHubOutput(t *testing.T) {
	repo := newCLITestRepo(t)
	repo.commit(t, "feat: publish release", "", "2024-01-01T00:00:00Z")
	t.Chdir(repo.dir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/jamesonstone/mint/releases/tags/v0.1.0":
			http.NotFound(w, r)
		case r.Method == http.MethodPost && r.URL.Path == "/repos/jamesonstone/mint/releases":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"tag_name":"v0.1.0","html_url":"https://github.com/jamesonstone/mint/releases/tag/v0.1.0"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "github-output")
	t.Setenv("MINT_TEST_GITHUB_TOKEN", "test-token")

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.SetOut(&stdout)

	err := runReleasePublish(cmd, releasePublishFlags{
		commitish:    "HEAD",
		owner:        "jamesonstone",
		repo:         "mint",
		push:         false,
		tokenEnv:     "MINT_TEST_GITHUB_TOKEN",
		apiURL:       server.URL,
		githubOutput: outputPath,
	})
	if err != nil {
		t.Fatalf("runReleasePublish() error = %v", err)
	}

	if !strings.Contains(stdout.String(), "published GitHub release v0.1.0") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	output := readCLITestFile(t, outputPath)
	for _, want := range []string{
		"version_tag=v0.1.0",
		"tag_name=v0.1.0",
		"tag_created=true",
		"release_tag=v0.1.0",
		"release_created=true",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("GitHub output missing %q:\n%s", want, output)
		}
	}
}

type cliTestRepo struct {
	dir     string
	counter int
}

func newCLITestRepo(t *testing.T) *cliTestRepo {
	t.Helper()

	repo := &cliTestRepo{dir: t.TempDir()}
	repo.git(t, nil, "init", "--initial-branch", "main")
	repo.git(t, nil, "config", "user.name", "Mint Test")
	repo.git(t, nil, "config", "user.email", "mint@example.com")
	return repo
}

func (repo *cliTestRepo) commit(t *testing.T, subject string, body string, date string) {
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
	repo.git(t, []string{
		"GIT_AUTHOR_DATE=" + date,
		"GIT_COMMITTER_DATE=" + date,
	}, args...)
}

func (repo *cliTestRepo) revParse(t *testing.T, ref string) string {
	t.Helper()

	cmd := exec.Command("git", "rev-parse", ref)
	cmd.Dir = repo.dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git rev-parse %s failed: %v\n%s", ref, err, string(out))
	}
	return strings.TrimSpace(string(out))
}

func (repo *cliTestRepo) git(t *testing.T, env []string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = repo.dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}

func readCLITestFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
