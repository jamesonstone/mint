package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestProductionHelpShowsOperatorWorkflow(t *testing.T) {
	cmd := newProductionCommand()
	var output bytes.Buffer
	cmd.SetOut(&output)
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	for _, name := range []string{"hotfix", "rollback", "status", "workflow"} {
		if !strings.Contains(text, "  "+name+" ") {
			t.Errorf("missing operator command %s: %s", name, text)
		}
	}
	for _, name := range []string{"candidate", "intent", "bootstrap", "prepare-hotfix", "control", "status-pr"} {
		if strings.Contains(text, "  "+name+" ") {
			t.Errorf("adapter command exposed in everyday help: %s", name)
		}
		found, _, err := cmd.Find([]string{name})
		if err != nil || found.Name() != name {
			t.Errorf("adapter removed: %s: %v", name, err)
		}
	}
}

func TestProductionRejectsIrrelevantOptionsBeforePolicyOrNetwork(t *testing.T) {
	for _, args := range [][]string{
		{"hotfix", "--to", "v1.2.3"},
		{"rollback", "--fix-pr", "123"},
		{"status", "--pin", "wrong"},
		{"candidate", "--input", "untrusted.json"},
		{"finish", "--outcome", "success"},
		{"workflow", "--token-env", "GH_TOKEN"},
	} {
		cmd := newProductionCommand()
		cmd.SilenceErrors, cmd.SilenceUsage = true, true
		cmd.SetArgs(args)
		err := cmd.Execute()
		if err == nil || !strings.Contains(err.Error(), "unknown flag") {
			t.Fatalf("irrelevant options were not rejected: %v: %v", args, err)
		}
	}
}

func TestProductionAdapterInputsRemainAvailable(t *testing.T) {
	group := newProductionCommand()
	for _, example := range []struct {
		name string
		args []string
	}{
		{"control", []string{"--event", "workflow_dispatch", "--input", "event.json", "--run-id", "42", "--token-env", "GITHUB_TOKEN"}},
		{"intent", []string{"--pr", "12", "--merge-sha", "commit", "--output", "intent.json"}},
		{"status-pr", []string{"--intent-id", "intent"}},
		{"propose", []string{"--kind", "hotfix", "--event", "candidate", "--pin", "source"}},
	} {
		cmd, _, err := group.Find([]string{example.name})
		if err != nil {
			t.Fatal(err)
		}
		if err := cmd.ParseFlags(example.args); err != nil {
			t.Fatalf("existing adapter invocation broken: %s: %v", example.name, err)
		}
	}
}
