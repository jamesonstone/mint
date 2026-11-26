package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigCommand(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := os.WriteFile(".mint.yaml", []byte("artifact: binary\ndeployment: continuous\nenvironments: [dev, stg, prd]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	cmd := newConfigCommand()
	cmd.SetOut(&output)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	want := `{"artifact":"binary","deployment":"continuous","environments":["dev","stg","prd"]}`
	if got := strings.TrimSpace(output.String()); got != want {
		t.Fatalf("configuration = %s, want %s", got, want)
	}

	cmd = newConfigCommand()
	cmd.SetArgs([]string{"--file", filepath.Join(t.TempDir(), "missing.yaml")})
	if err := cmd.Execute(); err == nil {
		t.Fatal("missing configuration was accepted")
	}
}

func TestPipelineConfigRejectsInvalidConfiguration(t *testing.T) {
	valid := "artifact: binary\ndeployment: continuous\nenvironments: [dev, stg, prd]\n"
	for _, test := range []struct{ name, config, want string }{
		{"missing artifact", strings.TrimPrefix(valid, "artifact: binary\n"), "artifact must be binary or image"},
		{"wrong artifact", strings.Replace(valid, "binary", "archive", 1), "artifact must be binary or image"},
		{"old policy fields", valid + "schema_version: 2\n", "field schema_version not found"},
		{"wrong deployment", strings.Replace(valid, "continuous", "automatic", 1), "deployment must be continuous or manual"},
		{"empty stages", strings.Replace(valid, "[dev, stg, prd]", "[]", 1), "environments must name at least one stage"},
		{"invalid stage", strings.Replace(valid, "stg", "Staging", 1), "invalid environment name"},
		{"duplicate stage", strings.Replace(valid, "stg", "dev", 1), "duplicate environment"},
		{"trailing document", valid + "---\nartifact: image\n", "expected one YAML document"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := readPipelineConfig(strings.NewReader(test.config))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("configuration error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestConfigCommandWithPerEnvironmentDeployment(t *testing.T) {
	t.Chdir(t.TempDir())
	config := "artifact: image\nenvironments:\n  - name: stg\n    deployment: continuous\n  - name: prd\n    deployment: manual\n"
	if err := os.WriteFile(".mint.yaml", []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	cmd := newConfigCommand()
	cmd.SetOut(&output)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	want := `{"artifact":"image","environments":[{"name":"stg","deployment":"continuous"},{"name":"prd","deployment":"manual"}]}`
	if got := strings.TrimSpace(output.String()); got != want {
		t.Fatalf("configuration = %s, want %s", got, want)
	}
}

func TestConfigRejectsInvalidEnvironmentMappings(t *testing.T) {
	for _, test := range []struct {
		name, config, want string
	}{
		{"mixed modes", "deployment: manual\nenvironments: [{name: stg, deployment: continuous}]", "either a top-level deployment mode or a mode per environment"},
		{"mixed stage types", "environments: [dev, {name: stg, deployment: continuous}]", "for a list of environment names"},
		{"missing name", "environments: [{deployment: manual}]", "invalid environment name"},
		{"missing mode", "environments: [{name: stg}]", "deployment must be continuous or manual"},
		{"unknown key", "environments: [{name: stg, deployment: continuous, extra: true}]", "unknown environment field"},
		{"duplicate name", "environments: [{name: stg, deployment: continuous}, {name: stg, deployment: manual}]", "duplicate environment"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := readPipelineConfig(strings.NewReader("artifact: binary\n" + test.config))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("configuration error = %v, want %q", err, test.want)
			}
		})
	}
}
