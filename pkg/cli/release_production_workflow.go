package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

func writeControlWorkflow(cmd *cobra.Command, cfg promotion.Config, f productionFlags) error {
	text, err := promotion.RenderControlWorkflow(cfg, f.MintRef, f.Config)
	if err != nil {
		return err
	}
	if f.Output != "" {
		if !promotion.SafeRepositoryPath(f.Output) || f.Output != cfg.ControlWorkflowPath() {
			return fmt.Errorf("workflow output must match configured control_workflow %s", cfg.ControlWorkflowPath())
		}
		if err := os.MkdirAll(filepath.Dir(f.Output), 0755); err != nil {
			return err
		}
		return os.WriteFile(f.Output, []byte(text), 0600)
	}
	_, err = fmt.Fprint(cmd.OutOrStdout(), text)
	return err
}
