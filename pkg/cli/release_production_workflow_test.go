package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

func TestWriteControlWorkflowCreatesConfiguredPath(t *testing.T) {
	t.Chdir(t.TempDir())
	cfg := promotion.Config{HumanLogin: "human", ControlWorkflow: ".github/workflows/recovery.yml"}
	f := productionFlags{Config: "config/release.yml", MintRef: strings.Repeat("a", 40), Output: cfg.ControlWorkflow}
	if err := writeControlWorkflow(&cobra.Command{}, cfg, f); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(cfg.ControlWorkflow)
	if err != nil || !strings.Contains(string(data), "production-config: 'config/release.yml'") {
		t.Fatal(string(data), err)
	}
	for _, output := range []string{".github/workflows/other.yml", "../recovery.yml", "/tmp/recovery.yml"} {
		f.Output = output
		if err := writeControlWorkflow(&cobra.Command{}, cfg, f); err == nil {
			t.Fatal("unexpected workflow path accepted", output)
		}
	}
}

func TestWriteControlWorkflowPrintsWithoutCreatingFile(t *testing.T) {
	t.Chdir(t.TempDir())
	var out bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&out)
	if err := writeControlWorkflow(cmd, promotion.Config{HumanLogin: "human"}, productionFlags{Config: ".mint.yaml", MintRef: strings.Repeat("a", 40)}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "production-config: '.mint.yaml'") {
		t.Fatal(out.String())
	}
	if _, err := os.Stat(".github"); !os.IsNotExist(err) {
		t.Fatal("stdout generation created repository files")
	}
}
