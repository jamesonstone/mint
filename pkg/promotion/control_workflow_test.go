package promotion

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestControlWorkflowCarriesSafeConfigurationPath(t *testing.T) {
	for _, configPath := range []string{".mint.yaml", "config/release-policy.yml"} {
		text, err := RenderControlWorkflow(controlConfig(), sha(1), configPath)
		if err != nil {
			t.Fatal(err)
		}
		var workflow map[string]any
		if err := yaml.Unmarshal([]byte(text), &workflow); err != nil {
			t.Fatal(err)
		}
		steps := workflow["jobs"].(map[string]any)["request"].(map[string]any)["steps"].([]any)
		with := steps[2].(map[string]any)["with"].(map[string]any)
		if with["production-config"] != configPath {
			t.Fatal("custom policy was lost", with)
		}
	}
	for _, configPath := range []string{"", "../policy.yml", "/policy.yml", "config/../policy.yml", "${{ github.token }}", "policy.yml'\nrun: bad", "config\\policy.yml", "."} {
		if _, err := RenderControlWorkflow(controlConfig(), sha(1), configPath); err == nil {
			t.Fatal("unsafe policy path accepted", configPath)
		}
	}
}

func TestControlWorkflowRejectsUndiscoverableDestination(t *testing.T) {
	for _, destination := range []string{"README.md", ".github/workflows/nested/control.yml", ".github/workflows/control.json", "../control.yml", ".github/workflows/${{ github.ref }}.yml"} {
		cfg := controlConfig()
		cfg.ControlWorkflow = destination
		if _, err := RenderControlWorkflow(cfg, sha(1)); err == nil {
			t.Fatal("invalid workflow destination accepted", destination)
		}
	}
}
