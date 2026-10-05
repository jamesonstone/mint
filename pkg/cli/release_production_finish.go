package cli

import (
	"context"
	"fmt"
	"github.com/jamesonstone/mint/pkg/promotion"
	"reflect"
	"time"
)

func (o *productionOperation) finish(ctx context.Context) (any, bool, error) {
	s := &o.snapshot.State
	f := o.flags
	i, ok := s.Intents[f.IntentID]
	if !ok {
		return nil, false, fmt.Errorf("unknown deployment intent")
	}
	run, err := o.client.ObservedRun(ctx, f.RunID, o.config.PromotionWorkflow)
	if err != nil {
		return nil, false, err
	}
	if run.Status != "completed" || (s.Schema == 1 && run.HeadSHA != i.MergeSHA) || i.DeploymentRunID != run.ID {
		return nil, false, fmt.Errorf("deployment run does not identify the frozen intent")
	}
	if i.Status == "deployed" || i.Status == "publication_pending" || i.Status == "deployment_failed" {
		return i, false, nil
	}
	var manifest promotion.DeploymentManifest
	if err := o.client.RunManifest(ctx, f.RunID, "mint-deployment", &manifest); err != nil {
		return nil, false, fmt.Errorf("runtime outcome unknown; retain deployment fence and reconcile live runtime: %w", err)
	}
	if !manifest.Verified || manifest.IntentID != i.ID {
		return nil, false, fmt.Errorf("unverified runtime outcome; retain deployment fence")
	}
	if s.Schema == 2 && manifest.Environment != s.Environment {
		return nil, false, fmt.Errorf("runtime outcome belongs to another environment; retain deployment fence")
	}
	outcome := "failure"
	expected := i.Candidate
	configuration := i.Configuration
	if configuration == "" {
		configuration = i.Candidate.Artifact.Configuration
	}
	if run.Conclusion == "success" {
		if manifest.Outcome != "" && manifest.Outcome != "success" {
			return nil, false, fmt.Errorf("successful run has conflicting runtime outcome")
		}
		outcome = "success"
	} else {
		if manifest.Outcome != "unchanged" || s.Baseline == nil {
			return nil, false, fmt.Errorf("failed run lacks verified unchanged baseline; retain deployment fence")
		}
		expected = s.Baseline.Candidate
		configuration = s.Baseline.RuntimeConfiguration()
	}
	if manifest.SourceSHA != expected.SourceSHA || manifest.ArtifactDigest != expected.Artifact.Digest || manifest.Configuration != configuration {
		return nil, false, fmt.Errorf("runtime manifest failed exact artifact verification; retain deployment fence")
	}
	if s.Schema == 2 && (manifest.Artifact == nil || !reflect.DeepEqual(*manifest.Artifact, expected.Artifact)) {
		return nil, false, fmt.Errorf("runtime manifest must verify the complete artifact identity; retain deployment fence")
	}
	err = s.FinishDeployment(i.ID, outcome, run.HTMLURL, o.proof)
	if err == nil && outcome == "success" && s.Schema == 2 && s.Baseline != nil && s.Baseline.ID == i.ID {
		if _, timestampErr := time.Parse(time.RFC3339Nano, run.UpdatedAt); timestampErr == nil {
			s.Baseline.VerifiedAt = run.UpdatedAt
		}
	}
	return s.Intents[i.ID], err == nil, err
}
