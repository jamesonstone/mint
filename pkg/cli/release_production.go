package cli

import (
	"encoding/json"
	"fmt"
	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
)

type productionFlags struct {
	Config, Input, Output, APIURL, TokenEnv, Kind, Event, IntentID, MergeSHA, Summary, Pin string
	Format, Environment, Version, Reason, MintRef                                          string
	Issue, FixPR                                                                           int
	RunID                                                                                  int64
	PR                                                                                     int
}
type productionOperation struct {
	config   promotion.Config
	client   promotion.Client
	snapshot promotion.JournalSnapshot
	flags    productionFlags
	proof    promotion.GitProof
}

func init() {
	releaseCmd.AddCommand(newProductionCommand())
	generic := newProductionCommand()
	generic.Use = "deployment"
	generic.Short = "Inspect environments and request immutable deployments"
	generic.Long = "Use repository Actions or a target-only .mint.yaml PR to request a deployment. Inspect desired, verified and observed state here. Project adapters own deployment operations."
	rootCmd.AddCommand(generic)
}
func readJSON(path string, value any) error {
	if path == "" {
		return fmt.Errorf("typed input file required")
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return fmt.Errorf("trailing input content")
	}
	return nil
}
func runProduction(cmd *cobra.Command, operation string, f productionFlags) error {
	if operation == "status" && f.Format != "" && f.Format != "json" && f.Format != "markdown" {
		return fmt.Errorf("status format must be json or markdown")
	}
	policy, err := promotion.LoadPolicy(f.Config)
	if err != nil {
		return err
	}
	if operation == "status" && policy.Schema == 2 && (f.Environment == "" || f.Format == "markdown") {
		return runEnvironmentOverview(cmd, policy, f)
	}
	if operation == "control" && policy.Schema == 2 && !strings.HasPrefix(f.Event, "authorized-") {
		return runPolicyControl(cmd, policy, f)
	}
	f.Event = strings.TrimPrefix(f.Event, "authorized-")
	cfg, err := policy.ForEnvironment(f.Environment)
	if err != nil {
		return err
	}
	if operation == "bootstrap" {
		if err := cfg.ValidateBootstrap(); err != nil {
			return err
		}
	}
	if operation == "policy" {
		return writeDeploymentJSON(cmd, cfg, f.Output)
	}
	if cfg.Mode != "deployment" && cfg.Schema == 2 && operation != "status" {
		return fmt.Errorf("%s policy uses publishing commands; runtime deployment operations are unavailable", cfg.Mode)
	}
	if cfg.Schema == 2 {
		if !promotion.SafeRepositoryPath(f.Config) {
			return fmt.Errorf("schema 2 policy must use a safe repository-relative path")
		}
		cfg.PolicyPath = f.Config
		cfg.ControlPaths = append(cfg.ControlPaths, f.Config)
	}
	if cfg.Schema == 2 && !cfg.Publish && (operation == "publish" || operation == "published") {
		return fmt.Errorf("publication is disabled for this environment")
	}
	client := promotion.Client{APIURL: f.APIURL, Token: os.Getenv(f.TokenEnv), Repository: cfg.Repository, HumanLogin: cfg.HumanLogin, Authorization: cfg.Authorization, Assignees: cfg.Assignees}
	if operation == "workflow" {
		return writeControlWorkflow(cmd, cfg, f)
	}
	if operation != "status" && operation != "validate" && operation != "validate-review" {
		if err := client.VerifyRepository(cmd.Context()); err != nil {
			return err
		}
	}
	if cfg.Scope == "local" {
		return runLocalEnvironment(cmd, operation, cfg, f)
	}
	snapshot, err := client.LoadJournal(cmd.Context(), cfg.Environment)
	if err != nil {
		return err
	}
	if cfg.Schema == 2 {
		snapshot.State.Schema = 2
		snapshot.State.Target = cfg.Target
		snapshot.State.Configuration = cfg.Configuration
		snapshot.State.Publish = cfg.Publish
		snapshot.State.PolicyDigest, err = promotion.AuthorityDigest(cfg)
		if err != nil {
			return err
		}
	}
	workDir, err := os.Getwd()
	if err != nil {
		return err
	}
	op := productionOperation{config: cfg, client: client, snapshot: snapshot, flags: f, proof: promotion.GitProof{Context: cmd.Context(), WorkDir: workDir, DefaultBranch: cfg.DefaultBranch}}
	if err := op.checkPrerequisites(cmd.Context(), operation); err != nil {
		return err
	}
	value, mutated, err := op.execute(cmd.Context(), operation)
	if err != nil {
		return err
	}
	if mutated {
		if _, err := client.SaveJournal(cmd.Context(), op.snapshot, cfg.Environment+" "+operation); err != nil {
			return err
		}
	}
	if operation == "control" && cfg.Schema == 2 {
		if i, ok := value.(promotion.Intent); ok && i.Status == "publication_pending" {
			if err := completeReconciledPublication(cmd, cfg, f, value); err != nil {
				return err
			}
			updated, err := client.LoadJournal(cmd.Context(), cfg.Environment)
			if err != nil {
				return err
			}
			value = updated.State.Intents[i.ID]
		}
	}
	if operation == "control" && cfg.Schema == 2 {
		if err := advanceReconciledQueue(cmd, cfg, f, value); err != nil {
			return err
		}
	}
	if mutated && cfg.Schema == 2 && operation == "finish" && cfg.ObservationWorkflow != "" {
		if err := client.DispatchObservation(cmd.Context(), cfg); err != nil {
			return fmt.Errorf("deployment outcome recorded; follow-up observation request failed: %w", err)
		}
	}
	if (operation == "recovery-propose" || operation == "propose" || operation == "propose-rollback" || operation == "rollback" || (operation == "control" && oIsProposal(value))) && mutated {
		p, ok := value.(promotion.Proposal)
		if ok && p.State == "open" {
			if err := client.DispatchChecks(cmd.Context(), cfg.ValidationWorkflow, p.Branch, p.PR); err != nil {
				return err
			}
		}
	}
	if mutated && cfg.Schema == 2 {
		if i, ok := value.(promotion.Intent); ok && (operation == "policy-request" || (operation == "propose" && cfg.Deploy == "automatic") || (operation == "intent" && f.Event == "pull_request_target")) {
			if err := client.DispatchIntent(cmd.Context(), cfg, i); err != nil {
				return fmt.Errorf("frozen intent recorded; adapter dispatch failed: %w", err)
			}
		}
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if f.Output != "" {
		if err := os.WriteFile(f.Output, append(data, '\n'), 0600); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return err
}
func productionHelp(operation string) string {
	switch operation {
	case "hotfix":
		return "Prepare an isolated fix by reviewed PR or new-fix issue"
	case "rollback":
		return "Propose restoring a verified deployed version; prefer a roll-forward hotfix"
	case "control":
		return "Handle authenticated repository Actions recovery requests"
	case "workflow":
		return "Generate the repository Actions hotfix/rollback entry point"
	case "report", "status-pr":
		return "Record deployment outcome on the original release PR"
	case "status":
		return "Inspect authoritative production state without mutation"
	default:
		return "Run the " + operation + " lifecycle adapter"
	}
}
