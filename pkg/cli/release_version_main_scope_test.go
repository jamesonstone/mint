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
