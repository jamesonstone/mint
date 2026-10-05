package cli

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

type environmentOverview struct {
	Environment string `json:"environment"`
	Mode        string `json:"mode"`
	Scope       string `json:"scope"`
	Desired     string `json:"desired"`
	Verified    string `json:"last_verified"`
	Live        string `json:"live"`
	Status      string `json:"status"`
	AsOf        string `json:"as_of"`
	Evidence    string `json:"evidence"`
	Paused      bool   `json:"paused"`
}

func runEnvironmentOverview(cmd *cobra.Command, policy promotion.Policy, f productionFlags) error {
	client := promotion.Client{APIURL: f.APIURL, Token: os.Getenv(f.TokenEnv), Repository: policy.Repository, HumanLogin: policy.HumanLogin, Authorization: policy.Authorization, Assignees: policy.Assignees}
	names := make([]string, 0, len(policy.Environments))
	for name := range policy.Environments {
		names = append(names, name)
	}
	sort.Strings(names)
	if f.Environment != "" {
		if _, ok := policy.Environments[f.Environment]; !ok {
			return fmt.Errorf("unknown environment")
		}
		names = []string{f.Environment}
	}
	rows := make([]environmentOverview, 0, len(names))
	for _, name := range names {
		cfg, err := policy.ForEnvironment(name)
		if err != nil {
			return err
		}
		state := promotion.NewState(cfg.Repository, name)
		state.Schema = 2
		if cfg.Mode == "deployment" && cfg.Scope == "shared" {
			journal, err := client.LoadJournal(cmd.Context(), name)
			if err != nil {
				return err
			}
			state = journal.State
		} else if cfg.Scope == "local" {
			cache, err := os.UserCacheDir()
			if err != nil {
				return err
			}
			path := filepath.Join(cache, "mint", fmt.Sprintf("%x", sha256.Sum256([]byte(cfg.Repository))), name+".json")
			data, err := os.ReadFile(path)
			if err == nil {
				state, err = promotion.DecodeState(data, cfg.Repository, name)
				if err != nil {
					return err
				}
			} else if !os.IsNotExist(err) {
				return err
			}
		}
		row := environmentOverview{Environment: name, Mode: cfg.Mode, Scope: cfg.Scope, Desired: cfg.Target, Verified: "—", Live: "—", Status: "unknown", AsOf: "—", Paused: state.Paused}
		if cfg.Follow == "latest" {
			row.Desired = "latest"
		}
		if state.Baseline != nil {
			row.Verified = state.Baseline.Candidate.Version
		}
		state.Schema = 2
		state.Target = cfg.Target
		state.Configuration = cfg.Configuration
		var desired *promotion.Candidate
		if cfg.Target != "" {
			for _, c := range state.Candidates {
				if c.Version == cfg.Target {
					copy := c
					desired = &copy
				}
			}
		}
		if cfg.Mode == "deployment" && cfg.Scope == "shared" {
			workDir, err := os.Getwd()
			if err != nil {
				return err
			}
			pin := ""
			if p, ok := state.Proposals["normal"]; ok && p.Selection == "pinned" && p.State == "open" {
				pin = p.CandidateSHA
			}
			candidate, err := state.SelectCandidate("normal", pin, promotion.GitProof{Context: cmd.Context(), WorkDir: workDir, DefaultBranch: cfg.DefaultBranch})
			if err == nil {
				desired = &candidate
				row.Desired = candidate.Version
			}
		}
		status := promotion.ProjectEnvironmentStatus(state, desired, cfg.Configuration)
		row.Status = status.Status
		if status.Live != nil {
			row.Live = status.Live.Version
			row.AsOf = status.AsOf
			row.Evidence = status.Evidence
		}
		if state.Paused {
			row.Status = "paused (" + row.Status + ")"
		}
		rows = append(rows, row)
	}
	if f.Format == "" || f.Format == "json" {
		return writeDeploymentJSON(cmd, rows, f.Output)
	}
	if f.Format != "markdown" {
		return fmt.Errorf("status format must be json or markdown")
	}
	var out strings.Builder
	out.WriteString("| Environment | Mode / scope | Desired | Last verified | Live | Status | Observed at | Evidence |\n| --- | --- | --- | --- | --- | --- | --- | --- |\n")
	for _, row := range rows {
		fmt.Fprintf(&out, "| %s | %s / %s | %s | %s | %s | %s | %s | %s |\n", markdownCell(row.Environment), row.Mode, row.Scope, markdownCell(row.Desired), markdownCell(row.Verified), markdownCell(row.Live), markdownCell(row.Status), markdownCell(row.AsOf), markdownCell(row.Evidence))
	}
	if f.Output != "" {
		return os.WriteFile(f.Output, []byte(out.String()), 0600)
	}
	_, err := fmt.Fprint(cmd.OutOrStdout(), out.String())
	return err
}
func markdownCell(value string) string {
	return strings.NewReplacer("|", "\\|", "\n", " ", "\r", " ").Replace(value)
}
