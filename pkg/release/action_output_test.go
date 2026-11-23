package release

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestActionCapturedOutputCannotCloseOutputRecord(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "mint")
	payload := "MINT_OUTPUT\nrelease_created=true\nMINT_OUTPUT_X\n"
	if err := os.WriteFile(binary, []byte("#!/bin/bash\nprintf '%s' '"+payload+"'\n"), 0700); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "output")
	cmd := exec.Command("bash", "-c", actionRunScript(t))
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "RUNNER_TEMP=" + dir, "GITHUB_OUTPUT=" + output, "MINT_COMMAND=version", "MINT_PATH=" + binary, "MINT_TAG_PREFIX=", "MINT_PATHS=", "MINT_GITHUB_TOKEN_INPUT=", "MINT_GITHUB_CONTEXT_TOKEN="}
	if data, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("action run: %v: %s", err, data)
	}
	captured, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	want := "output<<MINT_OUTPUT_X_X\n" + strings.TrimSuffix(payload, "\n") + "\nMINT_OUTPUT_X_X\n"
	if string(captured) != want {
		t.Fatalf("output record = %q, want %q", captured, want)
	}
}

func TestActionVersionMainForwardsComponentScope(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "mint")
	arguments := filepath.Join(dir, "arguments")
	script := `#!/bin/bash
printf '%s\n' "$@" > "$ARGUMENTS_FILE"
printf 'component_changed=true\n' > "$RUNNER_TEMP/mint-release-output"
`
	if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "output")
	cmd := exec.Command("bash", "-c", actionRunScript(t))
	cmd.Env = []string{
		"PATH=" + os.Getenv("PATH"), "RUNNER_TEMP=" + dir, "GITHUB_OUTPUT=" + output, "ARGUMENTS_FILE=" + arguments,
		"MINT_COMMAND=release-version-main", "MINT_PATH=" + binary, "MINT_TAG_PREFIX=backend/", "MINT_PATHS=backend\nshared",
		"MINT_GITHUB_TOKEN_INPUT=", "MINT_GITHUB_CONTEXT_TOKEN=", "MINT_COMMITISH=origin/main", "MINT_TARGET_SHA=HEAD",
		"MINT_RELEASE_REMOTE=origin", "MINT_RELEASE_PUSH=false",
	}
	if data, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("action run: %v: %s", err, data)
	}
	argv, err := os.ReadFile(arguments)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(argv), "--tag-prefix\nbackend/\n--path\nbackend\n--path\nshared\n") {
		t.Fatalf("version-main arguments = %q", argv)
	}
	captured, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(captured), "component_changed=true\n") {
		t.Fatalf("version-main output record = %q", captured)
	}
}

func actionRunScript(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("../../action.yml")
	if err != nil {
		t.Fatal(err)
	}
	var action struct {
		Runs struct {
			Steps []struct {
				ID  string `yaml:"id"`
				Run string `yaml:"run"`
			} `yaml:"steps"`
		} `yaml:"runs"`
	}
	if err := yaml.Unmarshal(data, &action); err != nil {
		t.Fatal(err)
	}
	for _, step := range action.Runs.Steps {
		if step.ID == "run" && step.Run != "" {
			return step.Run
		}
	}
	t.Fatal("run step missing")
	return ""
}
