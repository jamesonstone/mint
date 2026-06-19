package release

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestActionYAMLSupportsReleaseResolveOutputs(t *testing.T) {
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
		"MINT_COMMITISH",
		"release resolve --commitish",
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
