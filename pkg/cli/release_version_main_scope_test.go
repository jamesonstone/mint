package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runVersionMain(t *testing.T, args ...string) (string, string) {
	t.Helper()
	cmd := newVersionMainCommand()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	output := filepath.Join(t.TempDir(), "github-output")
	cmd.SetArgs(append([]string{"--main-ref", "main", "--push=false", "--github-output", output}, args...))
	if err := cmd.ExecuteContext(t.Context()); err != nil {
		t.Fatalf("version-main %v error = %v", args, err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	return stdout.String(), string(data)
}

func TestVersionMainScopesToComponentAndReportsChangedTarget(t *testing.T) {
	repo := newCLITestRepo(t)
	repo.commitAt(t, "backend/main.go", "feat: api", "2026-01-01T00:00:00Z")
	repo.commitAt(t, "shared/log.go", "fix: logger", "2026-01-02T00:00:00Z")
	sharedSHA := repo.revParse(t, "HEAD")
	repo.commitAt(t, "frontend/app.ts", "feat!: ui", "2026-01-03T00:00:00Z")
	frontendSHA := repo.revParse(t, "HEAD")
	t.Chdir(repo.dir)

	scope := []string{"--tag-prefix", "backend/", "--path", "backend", "--path", "shared"}
	stdout, out := runVersionMain(t, append(scope, "--target-commitish", sharedSHA)...)
	if !strings.Contains(stdout, `"version":"backend/v0.1.0"`) || strings.Contains(stdout, "v1.0.0") {
		t.Fatalf("stdout = %s; want scoped versions", stdout)
	}
	if !strings.Contains(out, "version_tag=backend/v0.1.1\n") || !strings.Contains(out, "component_changed=true\n") || !strings.Contains(out, "target_sha="+sharedSHA) {
		t.Fatalf("changed output = %q", out)
	}
	_, out = runVersionMain(t, append(scope, "--target-commitish", frontendSHA)...)
	if !strings.Contains(out, "version_tag=\n") || !strings.Contains(out, "component_changed=false\n") || !strings.Contains(out, "target_sha="+frontendSHA) {
		t.Fatalf("unchanged output = %q", out)
	}
}

func TestVersionMainComponentWithoutHistoryStillWritesOutput(t *testing.T) {
	repo := newCLITestRepo(t)
	repo.commitAt(t, "frontend/app.ts", "feat: ui", "2026-01-01T00:00:00Z")
	t.Chdir(repo.dir)
	stdout, out := runVersionMain(t, "--tag-prefix", "backend/", "--path", "backend")
	if strings.TrimSpace(stdout) != "[]" || !strings.Contains(out, "component_changed=false\n") {
		t.Fatalf("stdout = %q, output = %q", stdout, out)
	}
}

func TestVersionMainReportsChangelogOnlyCommit(t *testing.T) {
	repo := newCLITestRepo(t)
	repo.commitAt(t, "app.go", "feat: app", "2026-01-01T00:00:00Z")
	repo.commitAt(t, "CHANGELOG.md", "chore: release notes", "2026-01-02T00:00:00Z")
	t.Chdir(repo.dir)
	_, out := runVersionMain(t)
	if !strings.Contains(out, "control_only=true\n") || strings.Contains(out, "component_changed") {
		t.Fatalf("unscoped changelog-only output = %q", out)
	}
}
