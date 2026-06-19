package cli

import (
	"bytes"
	"context"
	"fmt"
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
	for _, want := range []string{"resolve", "workflow", "Resolve release metadata"} {
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
