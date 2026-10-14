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

const backendPolicy = `schema_version: 2
mode: deployment
repository: owner/repo
default_branch: main
human_login: jamesonstone
build_workflow: .github/workflows/backend-build.yaml
control_workflow: .github/workflows/backend-mint.yaml
component:
  name: backend
  paths: [backend]
  changelog: backend/CHANGELOG.md
environments:
  production:
    scope: shared
    deploy: reviewed
    follow: latest
`

func versionMainCommand(t *testing.T) *cobra.Command {
	t.Helper()
	for _, cmd := range releaseCmd.Commands() {
		if cmd.Name() == "version-main" {
			return cmd
		}
	}
	t.Fatal("version-main command not registered")
	return nil
}

func TestVersionMainConfigScopesToComponent(t *testing.T) {
	repo := newCLITestRepo(t)
	repo.commitAt(t, "backend/main.go", "feat: api", "2026-01-01T00:00:00Z")
	repo.commitAt(t, "frontend/app.ts", "feat!: redesign", "2026-01-02T00:00:00Z")
	if err := os.MkdirAll(filepath.Join(repo.dir, ".mint"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo.dir, ".mint", "backend.yaml"), []byte(backendPolicy), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo.dir)

	cmd := versionMainCommand(t)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetContext(context.Background())
	for name, value := range map[string]string{"main-ref": "main", "push": "false", "config": ".mint/backend.yaml"} {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		_ = cmd.Flags().Set("config", "")
		_ = cmd.Flags().Set("push", "true")
		_ = cmd.Flags().Set("main-ref", "origin/main")
	})
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("version-main --config error = %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, `"version":"backend/v0.1.0"`) || strings.Contains(out, "v1.0.0") {
		t.Fatalf("output = %s; want only backend/v0.1.0", out)
	}
}

func TestVersionMainConfigRejectsExplicitControlPaths(t *testing.T) {
	cmd := versionMainCommand(t)
	cmd.SetContext(context.Background())
	_ = cmd.Flags().Set("config", ".mint/backend.yaml")
	_ = cmd.Flags().Set("control-paths", "CHANGELOG.md")
	t.Cleanup(func() {
		_ = cmd.Flags().Set("config", "")
		cmd.Flags().Lookup("control-paths").Changed = false
	})
	if err := cmd.RunE(cmd, nil); err == nil || !strings.Contains(err.Error(), "do not also pass --control-paths") {
		t.Fatalf("error = %v", err)
	}
}

func runVersionMain(t *testing.T, flags map[string]string) (string, string) {
	t.Helper()
	cmd := versionMainCommand(t)
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetContext(context.Background())
	output := filepath.Join(t.TempDir(), "github-output")
	flags["github-output"] = output
	flags["push"] = "false"
	for name, value := range flags {
		if err := cmd.Flags().Set(name, value); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		for name, value := range map[string]string{"config": "", "push": "true", "main-ref": "origin/main", "target-commitish": "", "github-output": ""} {
			_ = cmd.Flags().Set(name, value)
		}
	})
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("version-main %v error = %v", flags, err)
	}
	data, _ := os.ReadFile(output)
	return stdout.String(), string(data)
}

func writePolicy(t *testing.T, dir, rel, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, filepath.Dir(rel)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, rel), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestVersionMainComponentReportsWhetherTargetChangedIt(t *testing.T) {
	repo := newCLITestRepo(t)
	writePolicy(t, repo.dir, ".mint/backend.yaml", backendPolicy)
	repo.commitAt(t, "backend/main.go", "feat: api", "2026-01-01T00:00:00Z")
	backendSHA := repo.revParse(t, "HEAD")
	repo.commitAt(t, "frontend/app.ts", "feat: ui", "2026-01-02T00:00:00Z")
	frontendSHA := repo.revParse(t, "HEAD")
	t.Chdir(repo.dir)

	_, out := runVersionMain(t, map[string]string{"main-ref": "main", "config": ".mint/backend.yaml", "target-commitish": backendSHA})
	if !strings.Contains(out, "version_tag=backend/v0.1.0\n") || !strings.Contains(out, "component_changed=true\n") || !strings.Contains(out, "target_sha="+backendSHA) {
		t.Fatalf("changed output = %q", out)
	}
	_, out = runVersionMain(t, map[string]string{"main-ref": "main", "config": ".mint/backend.yaml", "target-commitish": frontendSHA})
	if !strings.Contains(out, "version_tag=\n") || !strings.Contains(out, "component_changed=false\n") || !strings.Contains(out, "target_sha="+frontendSHA) {
		t.Fatalf("unchanged output = %q", out)
	}
}

func TestVersionMainComponentWithoutHistoryStillWritesOutput(t *testing.T) {
	repo := newCLITestRepo(t)
	writePolicy(t, repo.dir, ".mint/backend.yaml", backendPolicy)
	repo.commitAt(t, "frontend/app.ts", "feat: ui", "2026-01-01T00:00:00Z")
	t.Chdir(repo.dir)
	stdout, out := runVersionMain(t, map[string]string{"main-ref": "main", "config": ".mint/backend.yaml"})
	if strings.TrimSpace(stdout) != "[]" || !strings.Contains(out, "component_changed=false\n") {
		t.Fatalf("stdout = %q, output = %q", stdout, out)
	}
}

func TestVersionMainUnscopedConfigKeepsStatusControlOnly(t *testing.T) {
	repo := newCLITestRepo(t)
	unscoped := strings.Replace(backendPolicy, "component:\n  name: backend\n  paths: [backend]\n  changelog: backend/CHANGELOG.md\n", "", 1)
	writePolicy(t, repo.dir, ".mint.yaml", unscoped)
	repo.commitAt(t, "app.go", "feat: app", "2026-01-01T00:00:00Z")
	repo.commitAt(t, ".mint/status.json", "chore: status", "2026-01-02T00:00:00Z")
	status := repo.revParse(t, "HEAD")
	t.Chdir(repo.dir)
	_, out := runVersionMain(t, map[string]string{"main-ref": "main", "config": ".mint.yaml", "target-commitish": status})
	if !strings.Contains(out, "control_only=true\n") || strings.Contains(out, "component_changed") {
		t.Fatalf("output = %q; status-only commit must be control-only and unscoped output unchanged", out)
	}
}
