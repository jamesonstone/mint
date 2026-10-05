package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

// Explicit recovery can restore the reviewed queue after settling publication.
// Its replay cannot advance a newer baseline or an active/paused environment.
func advanceReconciledQueue(cmd *cobra.Command, cfg promotion.Config, f productionFlags, value any) error {
	return advanceReconciledQueueWith(cmd, cfg, f, value, runProduction)
}

func advanceReconciledQueueWith(cmd *cobra.Command, cfg promotion.Config, f productionFlags, value any, execute func(*cobra.Command, string, productionFlags) error) error {
	intent, ok := value.(promotion.Intent)
	if !ok || intent.Status != "deployed" || cfg.Schema != 2 || cfg.Mode != "deployment" || cfg.Scope != "shared" || (cfg.Deploy != "reviewed" && cfg.Deploy != "automatic") {
		return nil
	}
	if intent.ID == "" || intent.Environment != cfg.Environment {
		return fmt.Errorf("reconciled queue intent differs from environment")
	}
	client := promotion.Client{APIURL: f.APIURL, Token: os.Getenv(f.TokenEnv), Repository: cfg.Repository, HumanLogin: cfg.HumanLogin, Authorization: cfg.Authorization, Assignees: cfg.Assignees}
	snapshot, err := client.LoadJournal(cmd.Context(), cfg.Environment)
	if err != nil {
		return err
	}
	state := snapshot.State
	registered, ok := state.Intents[intent.ID]
	if !ok || registered.Status != "deployed" || state.Baseline == nil || state.Baseline.ID != intent.ID || state.Paused || state.InFlight != "" {
		return nil
	}
	child := f
	child.Environment, child.Output, child.Kind, child.Pin, child.Summary = cfg.Environment, "", "normal", "", ""
	child.IntentID = intent.ID
	original := cmd.OutOrStdout()
	cmd.SetOut(io.Discard)
	defer cmd.SetOut(original)
	if err := execute(cmd, "recovery-propose", child); err != nil {
		return fmt.Errorf("verified recovery recorded; queued release review could not be restored: %w", err)
	}
	return nil
}
