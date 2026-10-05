package cli

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

func TestReconciliationPublicationRetriesExactIntentWithoutRedeploy(t *testing.T) {
	for _, scenario := range []string{"pending", "retry", "failure", "revoked", "already-published", "foreign", "non-intent"} {
		t.Run(scenario, func(t *testing.T) {
			cfg := promotion.Config{Schema: 2, Environment: "stage", Publish: true}
			var value any = promotion.Intent{ID: "approved", Environment: "stage", Status: "publication_pending"}
			if scenario == "revoked" {
				cfg.Publish = false
			}
			if scenario == "already-published" {
				value = promotion.Intent{ID: "approved", Environment: "stage", Status: "deployed"}
			}
			if scenario == "foreign" {
				value = promotion.Intent{ID: "approved", Environment: "other", Status: "publication_pending"}
			}
			if scenario == "non-intent" {
				value = promotion.Proposal{State: "resumed"}
			}
			var out bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetContext(t.Context())
			cmd.SetOut(&out)
			calls := 0
			execute := func(cmd *cobra.Command, operation string, f productionFlags) error {
				calls++
				if operation != "publish" || f.Environment != "stage" || f.IntentID != "approved" || f.Output != "" {
					t.Fatal("publication did not use exact frozen identity", operation, f)
				}
				_, _ = fmt.Fprint(cmd.OutOrStdout(), "suppressed-child")
				if scenario == "failure" {
					return fmt.Errorf("release transport unavailable")
				}
				return nil
			}
			f := productionFlags{RunID: 99, Output: "user-output.json"}
			err := completeReconciledPublicationWith(cmd, cfg, f, value, execute)
			if scenario == "retry" {
				err = completeReconciledPublicationWith(cmd, cfg, f, value, execute)
			}
			wanted := 1
			if scenario == "retry" {
				wanted = 2
			}
			if scenario == "revoked" || scenario == "already-published" || scenario == "foreign" || scenario == "non-intent" {
				wanted = 0
			}
			if calls != wanted {
				t.Fatal("unexpected publication/redeployment calls", calls, wanted)
			}
			if (err != nil) != (scenario == "failure" || scenario == "revoked" || scenario == "foreign") {
				t.Fatal(err)
			}
			_, _ = fmt.Fprint(cmd.OutOrStdout(), "restored")
			if out.String() != "restored" {
				t.Fatal("child output leaked or output not restored", out.String())
			}
		})
	}
}
