package release

import (
	"os"
	"slices"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestActionYAMLSupportsReleaseCommands(t *testing.T) {
	data, err := os.ReadFile("../../action.yml")
	if err != nil {
		t.Fatalf("read action.yml: %v", err)
	}

	var decoded map[string]any
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("action.yml does not parse as YAML: %v", err)
	}

	action := string(data)
	for _, want := range []string{
		"release-resolve",
		"release-select-tag",
		"release-tag",
		"github-release",
		"release-publish",
		"release-version-main",
		"current-ref:",
		"requested-tag:",
		"MINT_CURRENT_REF",
		"MINT_COMMITISH",
		"MINT_REQUESTED_TAG",
		"MINT_RELEASE_TAG",
		"MINT_TARGET_SHA",
		"MINT_RELEASE_REMOTE",
		"MINT_RELEASE_PUSH",
		"MINT_GITHUB_TOKEN_INPUT",
		"MINT_GITHUB_CONTEXT_TOKEN",
		"release resolve --commitish",
		"release select-tag --commitish",
		"--requested-tag \"$MINT_REQUESTED_TAG\"",
		"--current-ref \"$MINT_CURRENT_REF\"",
		"release tag --tag",
		"release github --owner",
		"release publish --commitish",
		"release version-main --main-ref",
		"--github-output \"$release_output\"",
		"cat \"$release_output\" >> \"$GITHUB_OUTPUT\"",
		"version_tag:",
		"version_bump:",
		"base_tag:",
		"target_sha:",
		"short_sha:",
		"tag_source:",
		"needs_git_tag:",
		"commit_count:",
		"release_notes:",
		"release_tag:",
		"release_url:",
		"release_created:",
		"tag_created:",
		"tag_reused:",
		"tag_pushed:",
		"Unsupported mint action command",
		"tag-prefix:",
		"paths:",
		"MINT_TAG_PREFIX",
		"MINT_PATHS",
		"scope_paths+=(--path \"$scope_path\")",
	} {
		if !strings.Contains(action, want) {
			t.Fatalf("action.yml missing %q:\n%s", want, action)
		}
	}

	if strings.Contains(action, "eval ") || strings.Contains(action, "bash -c \"$MINT_COMMAND\"") {
		t.Fatalf("action.yml appears to run arbitrary shell:\n%s", action)
	}
	for _, removed := range []string{"production-control", "deployment-control", "production-config", "version-config", "mint-adapter"} {
		if strings.Contains(action, removed) {
			t.Fatalf("action.yml includes removed controller interface %q", removed)
		}
	}
}

func TestSelfReleaseWorkflowPromotesOneMintBinary(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/release.yaml")
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}

	type step struct {
		Run  string
		Uses string
		With map[string]string
	}
	type job struct {
		Needs       []string
		Permissions map[string]string
		If          string
		Steps       []step
	}
	var decoded struct {
		On          map[string]struct{ Branches []string } `yaml:"on"`
		Permissions map[string]string
		Jobs        map[string]job
	}
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("release workflow does not parse as YAML: %v", err)
	}

	if _, manual := decoded.On["workflow_dispatch"]; manual {
		t.Fatal("self-release workflow requires an automatic trigger")
	}
	if !slices.Equal(decoded.On["push"].Branches, []string{"main"}) {
		t.Fatal("self-release must run automatically on pushes to main")
	}
	if decoded.Permissions["contents"] != "read" {
		t.Fatal("release workflow must start with read permissions")
	}
	dev, stg, prd := decoded.Jobs["dev"], decoded.Jobs["stg"], decoded.Jobs["prd"]
	if len(decoded.Jobs) != 3 || len(dev.Needs) != 0 || !slices.Equal(stg.Needs, []string{"dev"}) ||
		!slices.Equal(prd.Needs, []string{"dev", "stg"}) {
		t.Fatalf("unexpected promotion graph: dev %v, stg %v, prd %v", dev.Needs, stg.Needs, prd.Needs)
	}
	if len(dev.Permissions) != 0 || len(stg.Permissions) != 0 ||
		prd.Permissions["contents"] != "write" || prd.If != "github.event_name == 'push' && github.ref == 'refs/heads/main'" {
		t.Fatal("only production on a main push may write release state")
	}
	if !slices.ContainsFunc(dev.Steps, func(s step) bool {
		return strings.HasPrefix(s.Uses, "jamesonstone/mint@") && s.With["command"] == "release-resolve"
	}) || !slices.ContainsFunc(dev.Steps, func(s step) bool { return strings.Contains(s.Run, "go build") }) {
		t.Fatal("dev must resolve the version with Mint and build the binary")
	}
	for _, stage := range []job{stg, prd} {
		if slices.ContainsFunc(stage.Steps, func(s step) bool { return strings.Contains(s.Run, "go build") }) {
			t.Fatal("a promotion stage rebuilds the artifact")
		}
		if !slices.ContainsFunc(stage.Steps, func(s step) bool {
			return strings.HasPrefix(s.Uses, "actions/download-artifact@") &&
				s.With["artifact-ids"] == "${{ needs.dev.outputs.artifact_id }}"
		}) {
			t.Fatal("a promotion stage does not download the exact dev artifact")
		}
	}
	if !slices.ContainsFunc(stg.Steps, func(s step) bool { return strings.Contains(s.Run, "verify-release-binary.sh") }) ||
		!slices.ContainsFunc(prd.Steps, func(s step) bool { return strings.Contains(s.Run, "publish-release-binary.sh") }) {
		t.Fatal("release workflow does not verify staging and publish production")
	}
}
