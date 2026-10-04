package cli

import (
	"fmt"
	"os"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

func writeControlWorkflow(cmd *cobra.Command, cfg promotion.Config, f productionFlags) error {
	text, err := promotion.RenderControlWorkflow(cfg, f.MintRef)
	if err != nil {
		return err
	}
	if f.Output != "" {
		return os.WriteFile(f.Output, []byte(text), 0600)
	}
	_, err = fmt.Fprint(cmd.OutOrStdout(), text)
	return err
}
