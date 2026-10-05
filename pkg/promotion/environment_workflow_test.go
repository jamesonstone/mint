package promotion

import (
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestEnvironmentWorkflowOffersGenericRequestsAndTrustedCompletionCallbacks(t *testing.T) {
	policy, err := ParsePolicy([]byte(environmentPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := RenderEnvironmentWorkflow(policy, sha(1), ".mint.yaml", "Build reusable artifact", "Observe running environment", "Promote exact artifact")
	if err != nil {
		t.Fatal(err)
	}
	var workflow map[string]any
	if err := yaml.Unmarshal([]byte(rendered), &workflow); err != nil {
		t.Fatal(err)
	}
	on := workflow["on"].(map[string]any)
	callbacks := on["workflow_run"].(map[string]any)
	names := callbacks["workflows"].([]any)
	if !reflect.DeepEqual(names, []any{"Build reusable artifact", "Observe running environment", "Promote exact artifact"}) || !reflect.DeepEqual(callbacks["types"], []any{"completed"}) {
		t.Fatal("adapter completion callbacks incomplete", callbacks)
	}
	inputs := on["workflow_dispatch"].(map[string]any)["inputs"].(map[string]any)
	environment := inputs["environment"].(map[string]any)
	if !reflect.DeepEqual(environment["options"], []any{"dev", "production"}) || environment["required"] != true {
		t.Fatal(environment)
	}
	operation := inputs["operation"].(map[string]any)
	if operation["default"] != "promote" || !reflect.DeepEqual(operation["options"], []any{"promote", "hotfix", "rollback", "resume", "observe", "reconcile"}) {
		t.Fatal(operation)
	}
	job := workflow["jobs"].(map[string]any)["control"].(map[string]any)
	guard := job["if"].(string)
	if !strings.Contains(guard, "MINT_RELEASE_ENABLED == 'true'") {
		t.Fatal("missing activation gate", guard)
	}
	steps := job["steps"].([]any)
	checkout := steps[0].(map[string]any)
	with := checkout["with"].(map[string]any)
	if checkout["uses"] != "actions/checkout@11d5960a326750d5838078e36cf38b85af677262" || with["ref"] != "${{ github.event.repository.default_branch }}" || with["fetch-depth"] != 0 || with["persist-credentials"] != false {
		t.Fatal("untrusted checkout or mutable action", checkout)
	}
	mint := steps[2].(map[string]any)
	if mint["uses"] != "jamesonstone/mint@"+sha(1) || mint["with"].(map[string]any)["command"] != "deployment-control" {
		t.Fatal(mint)
	}
	for _, value := range steps {
		if shell, ok := value.(map[string]any)["run"].(string); ok && strings.Contains(shell, "${{") {
			t.Fatal("event data interpolated into shell", shell)
		}
	}
}
func TestEnvironmentWorkflowRejectsMutablePinsAndUnsafeCallbackInputs(t *testing.T) {
	policy, err := ParsePolicy([]byte(environmentPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}
	for _, pin := range []string{"main", "v0.3.0", ""} {
		if _, err := RenderEnvironmentWorkflow(policy, pin, ".mint.yaml"); err == nil {
			t.Fatal("mutable Mint pin accepted", pin)
		}
	}
	for _, path := range []string{"../.mint.yaml", "/tmp/.mint.yaml", "${{ github.event.path }}"} {
		if _, err := RenderEnvironmentWorkflow(policy, sha(1), path); err == nil {
			t.Fatal("unsafe config path accepted", path)
		}
	}
	for _, name := range []string{"", "Build\nInjected", "${{ github.event.name }}"} {
		if _, err := RenderEnvironmentWorkflow(policy, sha(1), ".mint.yaml", name); err == nil {
			t.Fatal("unsafe callback name accepted", name)
		}
	}
	noCallbacks, err := RenderEnvironmentWorkflow(policy, sha(1), ".mint.yaml")
	if err != nil || strings.Contains(noCallbacks, "workflow_run:") {
		t.Fatal("optional callbacks not omitted", err)
	}
}
