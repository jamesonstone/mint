package cli

import (
	"context"
	"errors"
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
	knownRuns := make(map[int64]bool, len(o.snapshot.State.Candidates))
	for _, candidate := range o.snapshot.State.Candidates {
		knownRuns[candidate.RunID] = true
	}
	for _, run := range runs {
		if knownRuns[run.ID] {
			continue
		}
		var candidate promotion.Candidate
		if err := o.client.RunManifest(ctx, run.ID, "mint-candidate", &candidate); errors.Is(err, promotion.ErrManifestUnavailable) {
			continue
		} else if err != nil {
			return nil, false, err
		}
		if _, changed, err := o.registerCandidate(ctx, candidate, run); err != nil {
			return nil, false, err
		} else if changed {
			registered++
		}
		knownRuns[run.ID] = true
	}
	return map[string]int{"registered": registered}, registered > 0, nil
}
