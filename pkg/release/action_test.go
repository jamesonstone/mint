package release

import (
	"os"
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
		"release-tag",
		"github-release",
		"release-publish",
		"current-ref:",
		"MINT_CURRENT_REF",
		"MINT_COMMITISH",
		"MINT_RELEASE_TAG",
		"MINT_TARGET_SHA",
		"MINT_RELEASE_REMOTE",
		"MINT_RELEASE_PUSH",
		"MINT_GITHUB_TOKEN_INPUT",
		"MINT_GITHUB_CONTEXT_TOKEN",
		"release resolve --commitish",
		"--current-ref \"$MINT_CURRENT_REF\"",
		"release tag --tag",
		"release github --owner",
		"release publish --commitish",
		"--github-output \"$release_output\"",
		"cat \"$release_output\" >> \"$GITHUB_OUTPUT\"",
		"version_tag:",
		"version_bump:",
		"base_tag:",
		"target_sha:",
		"short_sha:",
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
	} {
		if !strings.Contains(action, want) {
			t.Fatalf("action.yml missing %q:\n%s", want, action)
		}
	}

	if strings.Contains(action, "eval ") || strings.Contains(action, "bash -c \"$MINT_COMMAND\"") {
		t.Fatalf("action.yml appears to run arbitrary shell:\n%s", action)
	}
}

func TestSelfReleaseWorkflowUsesMintForGitHubRelease(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/release.yaml")
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}

	var decoded map[string]any
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("release workflow does not parse as YAML: %v", err)
	}

	workflow := string(data)
	for _, want := range []string{
		"name: Release",
		"contents: write",
		"uses: ./",
		"command: release-publish",
		"commitish: ${{ github.sha }}",
		"github-token: ${{ secrets.GITHUB_TOKEN }}",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("release workflow missing %q:\n%s", want, workflow)
		}
	}

	for _, reject := range []string{"docker buildx", "packages: write", "aws ecs", "amazon-ecr-login", "docker/login-action"} {
		if strings.Contains(workflow, reject) {
			t.Fatalf("release workflow contains container/deploy content %q:\n%s", reject, workflow)
		}
	}
}
