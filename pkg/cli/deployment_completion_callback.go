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

// Adapter callbacks reconcile authenticated runtime evidence. Event conclusions
// never unlock deployments, and an observer alone never advances release intent.
func runAdapterCallback(cmd *cobra.Command, policy promotion.Policy, f productionFlags, data []byte) (bool, error) {
	return runAdapterCallbackWith(cmd, policy, f, data, runProduction)
}

func runAdapterCallbackWith(cmd *cobra.Command, policy promotion.Policy, f productionFlags, data []byte, execute func(*cobra.Command, string, productionFlags) error) (bool, error) {
	var event struct {
		WorkflowRun struct{ ID int64 } `json:"workflow_run"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return false, err
	}
	if event.WorkflowRun.ID <= 0 {
		return false, fmt.Errorf("adapter callback lacks a workflow run")
	}
	client := promotion.Client{APIURL: f.APIURL, Token: os.Getenv(f.TokenEnv), Repository: policy.Repository, HumanLogin: policy.HumanLogin}
	source, err := client.WorkflowRunIdentity(cmd.Context(), event.WorkflowRun.ID)
	if err != nil {
		return false, err
	}
	names := make([]string, 0, len(policy.Environments))
	for name := range policy.Environments {
		names = append(names, name)
	}
	sort.Strings(names)
	handled, failures := false, 0
	observerConfigured, observerMatched := false, false
	results := map[string]string{}
	original := cmd.OutOrStdout()
	defer cmd.SetOut(original)
	for _, name := range names {
		cfg, err := policy.ForEnvironment(name)
		if err != nil {
			return handled, err
		}
		observer := cfg.ObservationWorkflow != "" && promotion.WorkflowPathMatches(source.Path, cfg.ObservationWorkflow)
		deployer := cfg.PromotionWorkflow != "" && promotion.WorkflowPathMatches(source.Path, cfg.PromotionWorkflow)
		if !observer && !deployer {
			continue
		}
		handled = true
		if cfg.Mode != "deployment" || cfg.Scope != "shared" || cfg.Configuration == "" {
			results[name] = "not activated"
			continue
		}
		if err := client.AuthorizeAmbientRun(cmd.Context(), cfg, f.RunID, f.Event); err != nil {
			return true, err
		}
		if source.Status != "completed" || source.HeadBranch != cfg.DefaultBranch || (source.Event != "push" && source.Event != "workflow_dispatch") {
			results[name] = "adapter completion is not from the configured default branch"
			failures++
			continue
		}
		child := f
		child.Environment, child.RunID, child.Output = name, source.ID, ""
		var snapshot promotion.JournalSnapshot
		loaded := false
		if observer && deployer {
			snapshot, err = client.LoadJournal(cmd.Context(), name)
			if err != nil {
				results[name] = err.Error()
				failures++
				continue
			}
			loaded = true
			for _, intent := range snapshot.State.Intents {
				if intent.DeploymentRunID == source.ID {
					observer = false
					break
				}
			}
		}
		if observer {
			observerConfigured = true
			if source.Conclusion != "success" {
				results[name] = "observation unsuccessful; live outcome remains unknown"
				failures++
				continue
			}
			var observation promotion.Observation
			if err := client.RunManifest(cmd.Context(), source.ID, "mint-observation", &observation); err != nil {
				results[name] = err.Error()
				failures++
				continue
			}
			if observation.Environment != name {
				continue
			}
			observerMatched = true
			cmd.SetOut(io.Discard)
			err := execute(cmd, "observe", child)
			cmd.SetOut(original)
			if err != nil {
				results[name] = err.Error()
				failures++
			} else {
				results[name] = "authenticated observation recorded"
			}
			continue
		}
		if !loaded {
			snapshot, err = client.LoadJournal(cmd.Context(), name)
			if err != nil {
				results[name] = err.Error()
				failures++
				continue
			}
		}
		intentID := ""
		for id, intent := range snapshot.State.Intents {
			if intent.DeploymentRunID == source.ID {
				if intentID != "" {
					intentID = ""
					err = fmt.Errorf("adapter run identifies multiple intents")
					break
				}
				intentID = id
			}
		}
		if err != nil {
			results[name] = err.Error()
			failures++
			continue
		}
		if intentID == "" {
			results[name] = "no deployment intent for this adapter run"
			continue
		}
		child.IntentID = intentID
		prior := snapshot.State.Intents[intentID]
		if prior.Status == "deployment_failed" {
			results[name] = "verified unchanged failure already recorded; no advancement requested"
			continue
		}
		if prior.Status != "deployed" && prior.Status != "publication_pending" {
			cmd.SetOut(io.Discard)
			err = execute(cmd, "finish", child)
			cmd.SetOut(original)
			if err != nil {
				results[name] = "runtime outcome unknown; deployment fence retained: " + err.Error()
				failures++
				continue
			}
		}
		completed, err := client.LoadJournal(cmd.Context(), name)
		if err != nil {
			results[name] = err.Error()
			failures++
			continue
		}
		intent, ok := completed.State.Intents[intentID]
		if !ok {
			results[name] = "completed intent unavailable; no advancement requested"
			failures++
			continue
		}
		if intent.Status != "deployed" && intent.Status != "publication_pending" {
			results[name] = "verified deployment outcome recorded; no advancement requested"
			continue
		}
		results[name] = "verified deployment recorded"
		if intent.Status == "publication_pending" {
			if !cfg.Publish {
				results[name] += "; publication authority revoked; release remains pending"
				failures++
				continue
			}
			cmd.SetOut(io.Discard)
			err = execute(cmd, "publish", child)
			cmd.SetOut(original)
			if err != nil {
				results[name] += "; publication failed: " + err.Error()
				failures++
				continue
			}
			published, err := client.LoadJournal(cmd.Context(), name)
			if err != nil {
				results[name] += "; publication state unavailable: " + err.Error()
				failures++
				continue
			}
			if published.State.Intents[intentID].Status != "deployed" {
				results[name] += "; publication remains pending"
				failures++
				continue
			}
		}
		if completed.State.Baseline != nil && completed.State.Baseline.ID != intentID {
			results[name] += "; historical completion replay ignored"
			continue
		}
		if completed.State.Paused {
			results[name] += "; environment paused; explicit resume required"
			continue
		}
		if cfg.Deploy == "manual" {
			continue
		}
		child.RunID, child.IntentID, child.Kind, child.Pin, child.Summary = f.RunID, "", "normal", "", ""
		cmd.SetOut(io.Discard)
		err = execute(cmd, "propose", child)
		cmd.SetOut(original)
		if err != nil {
			results[name] += "; queued reconciliation failed: " + err.Error()
			failures++
			continue
		}
		results[name] += "; queued candidates reconciled"
	}
	if observerConfigured && !observerMatched && failures == 0 {
		results["observer"] = "observation manifest does not identify an activated environment"
		failures++
	}
	if !handled {
		return false, nil
	}
	cmd.SetOut(original)
	if err := writeDeploymentJSON(cmd, results, f.Output); err != nil {
		return true, err
	}
	if failures > 0 {
		return true, fmt.Errorf("%d adapter callbacks failed; inspect per-environment results", failures)
	}
	return true, nil
}
