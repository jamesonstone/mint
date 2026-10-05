package cli

import (
	"os"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

func newAdapterPolicyCommand() *cobra.Command {
	var config, role, event, apiURL, tokenEnv string
	var runID int64
	cmd := &cobra.Command{Use: "adapter-policy", Hidden: true, Args: cobra.NoArgs,
		Short: "Authenticate a compatibility Actions adapter and return its policy",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := promotion.LoadConfig(config)
			if err != nil {
				return err
			}
			client := promotion.Client{APIURL: apiURL, Token: os.Getenv(tokenEnv), Repository: cfg.Repository}
			if err := client.VerifyRepository(cmd.Context()); err != nil {
				return err
			}
			if err := client.AuthorizeAdapterRun(cmd.Context(), cfg, runID, role, event); err != nil {
				return err
			}
			return writeDeploymentJSON(cmd, cfg, "")
		},
	}
	f := cmd.Flags()
	f.StringVar(&config, "config", ".mint.yaml", "repository-owned release policy")
	f.StringVar(&role, "role", "", "source, promote, or control adapter")
	f.StringVar(&event, "event", "", "Actions event")
	f.Int64Var(&runID, "run-id", 0, "active server-verified Actions run")
	f.StringVar(&apiURL, "api-url", "https://api.github.com", "GitHub API URL")
	f.StringVar(&tokenEnv, "token-env", "GH_TOKEN", "job-scoped credential variable")
	return cmd
}
