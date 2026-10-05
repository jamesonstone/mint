package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

// Each successful producer callback is authenticated independently of its event.
// Manual environments collect candidates; reviewed/automatic environments then
// reconcile using the same frozen-intent engine. No environment rewrites builds.
func runBuildCallback(cmd *cobra.Command, policy promotion.Policy, f productionFlags, data []byte) error {
	var event struct {
		WorkflowRun struct{ ID int64 } `json:"workflow_run"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}
	if event.WorkflowRun.ID <= 0 {
		return fmt.Errorf("producer callback lacks a workflow run")
	}
	client := promotion.Client{APIURL: f.APIURL, Token: os.Getenv(f.TokenEnv), Repository: policy.Repository, HumanLogin: policy.HumanLogin, Authorization: policy.Authorization, Assignees: policy.Assignees}
	names := make([]string, 0, len(policy.Environments))
	for name := range policy.Environments {
		names = append(names, name)
	}
	sort.Strings(names)
	results := map[string]string{}
	failures := 0
	original := cmd.OutOrStdout()
	for _, name := range names {
		cfg, err := policy.ForEnvironment(name)
		if err != nil {
			return err
		}
		if err := client.AuthorizeAmbientRun(cmd.Context(), cfg, f.RunID, f.Event); err != nil {
			return err
		}
		producer, err := client.ObservedRun(cmd.Context(), event.WorkflowRun.ID, cfg.BuildWorkflow)
		if err != nil {
			return err
		}
		if producer.HeadBranch != cfg.DefaultBranch {
			return fmt.Errorf("producer callback is not from the configured default branch")
		}
		if producer.Status != "completed" || producer.Conclusion != "success" {
			results[name] = "producer did not succeed; no candidate registered"
			continue
		}
		if cfg.Mode != "deployment" || cfg.Scope != "shared" || cfg.Configuration == "" {
			results[name] = "not activated"
			continue
		}
		snapshot, err := client.LoadJournal(cmd.Context(), name)
		if err != nil {
			return err
		}
		if snapshot.State.Baseline == nil {
			results[name] = "verified bootstrap required"
			continue
		}
		var candidate promotion.Candidate
		if err := client.RunManifest(cmd.Context(), producer.ID, "mint-candidate", &candidate); err != nil {
			return err
		}
		if candidate.Kind == "hotfix" && candidate.BaselineID != snapshot.State.Baseline.ID {
			results[name] = "hotfix belongs to another environment baseline"
			continue
		}
		child := f
		child.Environment, child.RunID = name, producer.ID
		cmd.SetOut(io.Discard)
		err = runProduction(cmd, "candidate", child)
		cmd.SetOut(original)
		if err != nil {
			results[name] = err.Error()
			failures++
			continue
		}
		if cfg.Deploy == "manual" {
			results[name] = "candidate registered; explicit promotion required"
			continue
		}
		child.RunID = f.RunID
		if candidate.Kind == "hotfix" {
			child.Kind = "hotfix"
			child.Pin = candidate.SourceSHA
		}
		cmd.SetOut(io.Discard)
		err = runProduction(cmd, "propose", child)
		cmd.SetOut(original)
		if err != nil {
			results[name] = err.Error()
			failures++
			continue
		}
		results[name] = "eligible candidate reconciled"
	}
	if err := writeDeploymentJSON(cmd, results, f.Output); err != nil {
		return err
	}
	if failures > 0 {
		return fmt.Errorf("%d environment callbacks failed; inspect per-environment results", failures)
	}
	return nil
}
