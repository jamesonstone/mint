package promotion

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const minimalProductionPolicy = `schema_version: 1
repository: owner/repo
environment: production
default_branch: trunk
human_login: human
build_workflow: .github/workflows/build.yml
validation_workflow: .github/workflows/check.yml
promotion_workflow: .github/workflows/deploy.yml
required_checks: [ci]
`

func loadTestPolicy(t *testing.T, policy string) (Config, error) {
	t.Helper()
	file := filepath.Join(t.TempDir(), "mint.yml")
	if err := os.WriteFile(file, []byte(policy), 0600); err != nil {
		t.Fatal(err)
	}
	return LoadConfig(file)
}

func TestConfigDefaultsControlsAndLimitsBootstrapRequirements(t *testing.T) {
	cfg, err := loadTestPolicy(t, minimalProductionPolicy)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg.ControlPaths, []string{"CHANGELOG.md", ".mint/proposal.json", ".mint/summary.md"}) {
		t.Fatal(cfg.ControlPaths)
	}
	if cfg.BaselineBuildWorkflow != cfg.BuildWorkflow || cfg.ValidateBootstrap() == nil {
		t.Fatal("bootstrap did not require its trusted verifier", cfg)
	}
	cfg, err = loadTestPolicy(t, minimalProductionPolicy+"baseline_workflow: .github/workflows/import.yml\n")
	if err != nil || cfg.ValidateBootstrap() != nil {
		t.Fatal(cfg, err)
	}
	cfg, err = loadTestPolicy(t, minimalProductionPolicy+"baseline_workflow: .github/workflows/import.yml\nbaseline_build_workflow: .github/workflows/legacy-build.yml\n")
	if err != nil || cfg.BaselineBuildWorkflow != ".github/workflows/legacy-build.yml" || cfg.ValidateBootstrap() != nil {
		t.Fatal(cfg, err)
	}
}

func TestConfigRejectsUnsafeOrPartialExplicitControls(t *testing.T) {
	for _, paths := range []string{"[]", "null", "[CHANGELOG.md]", "[CHANGELOG.md, .mint/proposal.json]", "[CHANGELOG.md, .mint/proposal.json, .mint/summary.md, source.go]", "[CHANGELOG.md, .mint/proposal.json, .mint/summary.md, CHANGELOG.md]"} {
		t.Run(paths, func(t *testing.T) {
			if _, err := loadTestPolicy(t, minimalProductionPolicy+"control_paths: "+paths+"\n"); err == nil {
				t.Fatal("incomplete or unsafe controls accepted", paths)
			}
		})
	}
	for _, paths := range []string{"[CHANGELOG.md, .mint/proposal.json, .mint/summary.md]", "[.mint/status.json, .mint/summary.md, CHANGELOG.md, .mint/proposal.json]"} {
		if _, err := loadTestPolicy(t, minimalProductionPolicy+"control_paths: "+paths+"\n"); err != nil {
			t.Fatal("complete control policy rejected", paths, err)
		}
	}
}

func TestConfigStillRequiresCoreReleasePolicy(t *testing.T) {
	for _, field := range []string{"repository: owner/repo\n", "build_workflow: .github/workflows/build.yml\n", "promotion_workflow: .github/workflows/deploy.yml\n", "required_checks: [ci]\n"} {
		if _, err := loadTestPolicy(t, strings.ReplaceAll(minimalProductionPolicy, field, "")); err == nil {
			t.Fatal("core policy missing", field)
		}
	}
	if _, err := loadTestPolicy(t, minimalProductionPolicy+"unrecognized: true\n"); err == nil {
		t.Fatal("unknown policy field accepted")
	}
	if _, err := loadTestPolicy(t, minimalProductionPolicy+"---\nrepository: other/repo\n"); err == nil {
		t.Fatal("second policy document accepted")
	}
	if _, err := loadTestPolicy(t, "<<: &policy\n  control_paths: [CHANGELOG.md]\n"+minimalProductionPolicy); err == nil {
		t.Fatal("merged control policy accepted")
	}
}
