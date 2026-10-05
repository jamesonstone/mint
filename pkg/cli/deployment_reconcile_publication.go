package cli

import (
	"fmt"
	"io"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

// A human reconciliation settles runtime before attempting publication. Failed
// publication stays retryable through the same consumed reconciliation request.
func completeReconciledPublication(cmd *cobra.Command, cfg promotion.Config, f productionFlags, value any) error {
	return completeReconciledPublicationWith(cmd, cfg, f, value, runProduction)
}

func completeReconciledPublicationWith(cmd *cobra.Command, cfg promotion.Config, f productionFlags, value any, execute func(*cobra.Command, string, productionFlags) error) error {
	intent, ok := value.(promotion.Intent)
	if !ok || intent.Status != "publication_pending" {
		return nil
	}
	if cfg.Schema != 2 || intent.Environment != cfg.Environment || intent.ID == "" {
		return fmt.Errorf("reconciled publication identity differs from environment")
	}
	if !cfg.Publish {
		return fmt.Errorf("verified deployment recorded; publication authority revoked; release remains pending")
	}
	child := f
	child.Environment, child.IntentID, child.Output = cfg.Environment, intent.ID, ""
	original := cmd.OutOrStdout()
	cmd.SetOut(io.Discard)
	defer cmd.SetOut(original)
	if err := execute(cmd, "publish", child); err != nil {
		return fmt.Errorf("verified deployment recorded; publication remains pending: %w", err)
	}
	return nil
}
