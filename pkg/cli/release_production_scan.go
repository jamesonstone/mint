package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/jamesonstone/mint/pkg/promotion"
)

// scan recovers successful producer runs missed by webhook delivery or Actions'
// single-pending concurrency slot. Durable candidate identity makes replay safe.
func (o *productionOperation) scan(ctx context.Context) (any, bool, error) {
	runs, err := o.client.SuccessfulBuilds(ctx, o.config.BuildWorkflow, o.config.DefaultBranch)
	if err != nil {
		return nil, false, err
	}
	registered := 0
	for _, run := range runs {
		knownRun := false
		for _, c := range o.snapshot.State.Candidates {
			if c.RunID == run.ID {
				knownRun = true
				break
			}
		}
		if knownRun {
			continue
		}
		var candidate promotion.Candidate
		if err := o.client.RunManifest(ctx, run.ID, "mint-candidate", &candidate); errors.Is(err, promotion.ErrManifestUnavailable) {
			continue
		} else if err != nil {
			return nil, false, err
		}
		if (candidate.Kind == "normal" && candidate.SourceSHA != run.HeadSHA) || candidate.RunID != run.ID {
			return nil, false, fmt.Errorf("producer manifest/run identity conflict")
		}
		if _, known := o.snapshot.State.Candidates[candidate.SourceSHA]; known {
			continue
		}
		o.flags.RunID = run.ID
		if _, _, err := o.execute(ctx, "candidate"); err != nil {
			return nil, false, err
		}
		registered++
	}
	return map[string]int{"registered": registered}, registered > 0, nil
}
