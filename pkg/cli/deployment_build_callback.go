package cli

import (
	"encoding/json"
	"errors"
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
	return runBuildCallbackWith(cmd, policy, f, data, runProduction)
}

// fenceOutcome is what a build callback found behind a deployment fence.
type fenceOutcome int

const (
	fenceWaits    fenceOutcome = iota // the adapter run is active, or the fence is gone
	fenceSettled                      // this callback settled an ended run; propose again
	fenceOperator                     // only an operator can release the fence
)

// settleEndedFence settles a deployment fence whose adapter run has ended.
// GitHub creates no workflow_run event for a run that the workflow token
// dispatched, so the adapter completion callback can be missing. finish reads
// the run's own verified mint-deployment manifest; an event conclusion never
// settles a fence. Any other state returns why an operator must act.
func settleEndedFence(cmd *cobra.Command, client promotion.Client, cfg promotion.Config, child productionFlags, execute func(*cobra.Command, string, productionFlags) error) (fenceOutcome, string) {
	ctx := cmd.Context()
	snapshot, err := client.LoadJournal(ctx, child.Environment)
	if err != nil {
		return fenceOperator, "deployment fence state unavailable: " + err.Error()
	}
	if snapshot.State.InFlight == "" {
		// Released since the refusal; its completion callback reconciled the queue.
		return fenceWaits, ""
	}
	intent, ok := snapshot.State.Intents[snapshot.State.InFlight]
	if !ok || intent.DeploymentRunID <= 0 {
		return fenceOperator, "the fence has no deployment run; settle it with an explicit reconcile request"
	}
	run, err := client.ObservedRun(ctx, intent.DeploymentRunID, cfg.PromotionWorkflow)
	if err != nil {
		return fenceOperator, "deployment run state unavailable: " + err.Error()
	}
	if run.Status != "completed" {
		return fenceWaits, ""
	}
	if run.HeadBranch != cfg.DefaultBranch {
		return fenceOperator, "its deployment run is not from the configured default branch; settle it with an explicit reconcile request"
	}
	settle := child
	settle.RunID, settle.IntentID, settle.Kind, settle.Pin, settle.Output = run.ID, intent.ID, "", "", ""
	original := cmd.OutOrStdout()
	cmd.SetOut(io.Discard)
	err = execute(cmd, "finish", settle)
	cmd.SetOut(original)
	if err != nil {
		return fenceOperator, "its deployment run has ended without a verified outcome (" + err.Error() + "); observe the runtime and settle it with an explicit reconcile request"
	}
	settled, err := client.LoadJournal(ctx, child.Environment)
	if err != nil {
		return fenceOperator, "settled deployment state unavailable: " + err.Error()
	}
	// Advance only after a verified deployment, as the completion callback does.
	switch status := settled.State.Intents[intent.ID].Status; {
	case settled.State.InFlight != "":
		return fenceOperator, "its deployment outcome is recorded but the fence remains; settle it with an explicit reconcile request"
	case status == "deployment_failed":
		return fenceOperator, "its deployment run failed; the unchanged runtime is recorded and no advancement is requested"
	case status != "deployed":
		return fenceOperator, "its deployment outcome is " + status + "; no advancement is requested"
	}
	return fenceSettled, ""
}

func runBuildCallbackWith(cmd *cobra.Command, policy promotion.Policy, f productionFlags, data []byte, execute func(*cobra.Command, string, productionFlags) error) error {
	var event struct {
		WorkflowRun struct{ ID int64 } `json:"workflow_run"`
	}
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}
	if event.WorkflowRun.ID <= 0 {
		return fmt.Errorf("producer callback lacks a workflow run")
	}
	client := promotion.Client{APIURL: f.APIURL, Token: os.Getenv(f.TokenEnv), Repository: policy.Repository, HumanLogin: policy.HumanLogin, Authorization: policy.Authorization, Assignees: policy.Assignees, Component: policy.ComponentScope()}
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
		producer, err := client.ObservedBuildRun(cmd.Context(), event.WorkflowRun.ID, cfg, cfg.BuildWorkflow)
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
		err = execute(cmd, "candidate", child)
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
		err = execute(cmd, "propose", child)
		cmd.SetOut(original)
		settledNote := ""
		if errors.Is(err, promotion.ErrInFlight) {
			// A fence whose run has ended is settled here from the run's
			// verified evidence, then the candidate is proposed again. While
			// the run is active, a normal candidate waits for the completion
			// callback; a hotfix has no such continuation and stays an error.
			// A fence that this callback cannot settle needs an operator, so
			// that state stays a visible failure.
			outcome, reason := settleEndedFence(cmd, client, cfg, child, execute)
			switch {
			case outcome == fenceOperator:
				results[name] = err.Error() + "; " + reason
				failures++
				continue
			case outcome == fenceSettled:
				settledNote = "ended deployment settled from its verified run evidence; "
				cmd.SetOut(io.Discard)
				err = execute(cmd, "propose", child)
				cmd.SetOut(original)
			case candidate.Kind == "normal":
				results[name] = "candidate registered; a deployment is in flight, reconciliation waits for its outcome"
				continue
			}
		}
		if err != nil {
			results[name] = settledNote + err.Error()
			failures++
			continue
		}
		results[name] = settledNote + "eligible candidate reconciled"
	}
	if err := writeDeploymentJSON(cmd, results, f.Output); err != nil {
		return err
	}
	if failures > 0 {
		return fmt.Errorf("%d environment callbacks failed; inspect per-environment results", failures)
	}
	return nil
}
