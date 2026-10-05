package cli

import (
	"context"
	"fmt"
	"github.com/jamesonstone/mint/pkg/promotion"
	"strings"
)

func (o *productionOperation) execute(ctx context.Context, operation string) (any, bool, error) {
	s := &o.snapshot.State
	f := o.flags
	switch operation {
	case "dispatch":
		i, ok := s.Intents[f.IntentID]
		if !ok {
			return nil, false, fmt.Errorf("unknown intent")
		}
		return i, false, o.client.DispatchIntent(ctx, o.config, i)
	case "recovery-propose":
		if err := o.client.AuthorizeControlRun(ctx, o.config, f.RunID, "workflow_dispatch"); err != nil {
			return nil, false, err
		}
		if s.Schema != 2 || s.Baseline == nil || s.Baseline.ID != f.IntentID || s.Intents[f.IntentID].Status != "deployed" || s.Paused || s.InFlight != "" {
			return map[string]string{"state": "idle"}, false, nil
		}
		return o.requestPromotion(ctx, o.config.Target, "")
	case "hotfix":
		return o.requestHotfix(ctx)
	case "rollback":
		return o.requestRollback(ctx)
	case "control":
		return o.control(ctx)
	case "scan":
		return o.scan(ctx)
	case "validate-review":
		return o.review(ctx)
	case "status":
		if s.Schema == 2 {
			return o.environmentStatus(), false, nil
		}
		return s, false, nil
	case "observe":
		obs, err := o.client.ReadObservation(ctx, o.config, f.RunID)
		if err != nil {
			return nil, false, err
		}
		changed, err := promotion.ApplyObservation(s, obs)
		return obs, changed, err
	case "resume":
		if err := o.client.AuthorizeControlRun(ctx, o.config, f.RunID, "workflow_dispatch"); err != nil {
			return nil, false, err
		}
		return s.Resume(fmt.Sprintf("resume-control-%d", f.RunID))
	case "policy-request":
		return o.policyRequest(ctx, true)
	case "validate-policy-request":
		return o.policyRequest(ctx, false)
	case "candidate":
		return o.candidate(ctx)
	case "propose-rollback":
		if s.Baseline == nil {
			return nil, false, fmt.Errorf("verified baseline required")
		}
		p, err := s.ReconcileRollback(f.Pin, "rollback-"+s.Baseline.ID, f.Summary)
		if err != nil {
			return nil, false, err
		}
		if p.State != "open" {
			return p, false, nil
		}
		p, err = o.client.SyncProposal(ctx, o.config, s, p)
		return p, true, err
	case "status-pr", "report":
		i, ok := s.Intents[f.IntentID]
		if !ok {
			return nil, false, fmt.Errorf("unknown intent")
		}
		p, err := o.client.SyncStatus(ctx, o.config, i)
		return p, false, err
	case "propose":
		return o.propose(ctx)
	case "validate", "intent":
		return o.intent(ctx, operation == "intent")
	case "start":
		return o.start(ctx)
	case "finish":
		return o.finish(ctx)
	case "published":
		if err := o.client.VerifyPublication(ctx, s.Intents[f.IntentID]); err != nil {
			return nil, false, err
		}
		return s, true, s.MarkPublished(f.IntentID)
	case "publish":
		i, ok := s.Intents[f.IntentID]
		if !ok {
			return nil, false, fmt.Errorf("unknown intent")
		}
		if err := o.client.PublishIntent(ctx, i); err != nil {
			return nil, false, err
		}
		return i, true, s.MarkPublished(i.ID)
	case "bootstrap":
		return o.bootstrap(ctx)
	case "prepare-hotfix":
		return o.hotfix(ctx)
	case "version-hotfix":
		return o.versionHotfix(ctx)
	}
	return nil, false, fmt.Errorf("unsupported production operation %s", strings.TrimSpace(operation))
}
