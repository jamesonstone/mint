package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func (repo *cliTestRepo) commitAt(t *testing.T, relPath string, subject string, date string) {
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
}

func TestReleaseCommandsExposeScopeFlags(t *testing.T) {
	withPaths := map[string]*cobra.Command{"resolve": releaseResolveCmd, "publish": releasePublishCmd, "changelog": changelogCmd}
	prefixOnly := map[string]*cobra.Command{"select-tag": releaseSelectTagCmd, "tag": releaseTagCmd, "github": releaseGitHubCmd}
	for name, cmd := range withPaths {
		if cmd.Flags().Lookup("tag-prefix") == nil || cmd.Flags().Lookup("path") == nil {
			t.Errorf("%s is missing --tag-prefix or --path", name)
		}
	}
	for name, cmd := range prefixOnly {
		if cmd.Flags().Lookup("tag-prefix") == nil {
			t.Errorf("%s is missing --tag-prefix", name)
		}
	}
}

func TestRunReleaseResolveScoped(t *testing.T) {
	repo := newCLITestRepo(t)
	repo.commitAt(t, "backend/main.go", "feat: add api", "2024-01-01T00:00:00Z")
	repo.commitAt(t, "frontend/app.ts", "feat!: redesign", "2024-01-02T00:00:00Z")
	t.Chdir(repo.dir)

	var stdout bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetContext(context.Background())
	cmd.SetOut(&stdout)

	err := runReleaseResolve(cmd, releaseResolveFlags{
		commitish: "HEAD",
		scope:     scopeFlags{tagPrefix: "backend/", paths: []string{"backend"}},
	})
	if err != nil {
		t.Fatalf("runReleaseResolve() error = %v", err)
	}
	if got := strings.TrimSpace(stdout.String()); got != "backend/v0.1.0" {
		t.Fatalf("stdout = %q, want backend/v0.1.0 (frontend breaking change must not count)", got)
	}
}
