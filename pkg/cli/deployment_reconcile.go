package cli

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/jamesonstone/mint/pkg/promotion"
)

// Reconciliation is an explicit human recovery request. Fresh authenticated
// runtime evidence can settle a completed adapter whose outcome was unknown;
// ambient observations alone never advance the verified deployment journal.
func (o *productionOperation) reconcileDeployment(ctx context.Context, intentID string, observationRunID int64) (any, bool, error) {
	cfg, s := o.config, &o.snapshot.State
	if cfg.Schema != 2 || cfg.Scope != "shared" || cfg.Mode != "deployment" {
		return nil, false, fmt.Errorf("reconciliation requires a shared deployment environment")
	}
	if err := o.client.AuthorizeControlRun(ctx, cfg, o.flags.RunID, "workflow_dispatch"); err != nil {
		return nil, false, err
	}
	key := fmt.Sprintf("reconcile-control-%d", o.flags.RunID)
	if prior, ok := s.ControlRequests[key]; ok {
		if intentID != "" && prior.ID != intentID {
			return nil, false, fmt.Errorf("reconciliation request identity changed")
		}
		return s.Intents[prior.ID], false, nil
	}
	if intentID == "" {
		intentID = s.InFlight
	}
	if observationRunID == 0 && s.Observation != nil {
		observationRunID = s.Observation.RunID
	}
	intent, ok := s.Intents[intentID]
	if !ok || intent.Status != "deploying" || s.InFlight != intentID || s.Baseline == nil || intent.BaselineID != s.Baseline.ID || intent.DeploymentRunID <= 0 {
		return nil, false, fmt.Errorf("reconciliation requires the active deployment intent and its original baseline")
	}
	authority, err := promotion.AuthorityDigest(cfg)
	if err != nil {
		return nil, false, err
	}
	if intent.PolicyDigest != authority || intent.Configuration != cfg.Configuration {
		return nil, false, fmt.Errorf("reviewed environment authority changed; deployment fence retained")
	}
	run, err := o.client.ObservedRun(ctx, intent.DeploymentRunID, cfg.PromotionWorkflow)
	if err != nil {
		return nil, false, err
	}
	if run.Status != "completed" || run.HeadBranch != cfg.DefaultBranch {
		return nil, false, fmt.Errorf("original deployment is not completed on the trusted default branch; deployment fence retained")
	}
	observation, err := o.client.ReadObservation(ctx, cfg, observationRunID)
	if err != nil {
		return nil, false, err
	}
	outcome, err := reconciledOutcome(*s, intent, run, observation)
	if err != nil {
		return nil, false, err
	}
	// Validate observation ordering before any lifecycle mutation.
	observedState := *s
	if _, err := promotion.ApplyObservation(&observedState, observation); err != nil {
		return nil, false, err
	}
	evidence := run.HTMLURL + "; reconciled runtime: " + observation.RunURL
	if err := s.FinishDeployment(intentID, outcome, evidence, o.proof); err != nil {
		return nil, false, err
	}
	if outcome == "success" {
		s.Baseline.VerifiedAt = observation.ObservedAt
	}
	s.Observation = observedState.Observation
	if s.ControlRequests == nil {
		s.ControlRequests = map[string]promotion.Proposal{}
	}
	s.ControlRequests[key] = promotion.Proposal{ID: intentID, Kind: "reconcile", State: outcome, BaselineID: intent.BaselineID}
	return s.Intents[intentID], true, nil
}

func reconciledOutcome(s promotion.State, intent promotion.Intent, run promotion.WorkflowRun, observation promotion.Observation) (string, error) {
	completed, err := time.Parse(time.RFC3339Nano, run.UpdatedAt)
	observed, observedErr := time.Parse(time.RFC3339Nano, observation.ObservedAt)
	if err != nil || observedErr != nil || !observed.After(completed) {
		return "", fmt.Errorf("runtime observation must be taken after the original deployment completed; deployment fence retained")
	}
	if observation.Status != "live" || observation.Repository != s.Repository || observation.Environment != s.Environment {
		return "", fmt.Errorf("runtime outcome unknown or mixed; deployment fence retained")
	}
	if observation.SourceSHA == intent.Candidate.SourceSHA && observation.Version == intent.Candidate.Version && observation.Configuration == intent.Configuration && reflect.DeepEqual(observation.Artifact, intent.Candidate.Artifact) {
		return "success", nil
	}
	if s.Baseline != nil && observation.SourceSHA == s.Baseline.Candidate.SourceSHA && observation.Version == s.Baseline.Candidate.Version && observation.Configuration == s.Baseline.RuntimeConfiguration() && reflect.DeepEqual(observation.Artifact, s.Baseline.Candidate.Artifact) {
		return "failure", nil
	}
	return "", fmt.Errorf("live runtime matches neither approved intent nor prior baseline; deployment fence retained")
}
