package cli

import (
	"context"
	"fmt"
	"github.com/jamesonstone/mint/pkg/promotion"
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
	if run.Status != "completed" || run.HeadSHA != i.MergeSHA || i.DeploymentRunID != run.ID {
		return nil, false, fmt.Errorf("deployment run does not identify the frozen intent")
	}
	var manifest promotion.DeploymentManifest
	if err := o.client.RunManifest(ctx, f.RunID, "mint-deployment", &manifest); err != nil {
		return nil, false, fmt.Errorf("runtime outcome unknown; retain deployment fence and reconcile live runtime: %w", err)
	}
	if !manifest.Verified || manifest.IntentID != i.ID {
		return nil, false, fmt.Errorf("unverified runtime outcome; retain deployment fence")
	}
	outcome := "failure"
	expected := i.Candidate
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
	}
	if manifest.SourceSHA != expected.SourceSHA || manifest.ArtifactDigest != expected.Artifact.Digest || manifest.Configuration != expected.Artifact.Configuration {
		return nil, false, fmt.Errorf("runtime manifest failed exact artifact verification; retain deployment fence")
	}
	err = s.FinishDeployment(i.ID, outcome, run.HTMLURL, o.proof)
	return s.Intents[i.ID], err == nil, err
}
