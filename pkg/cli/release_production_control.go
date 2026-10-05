package cli

import (
	"context"
	"fmt"
	"github.com/jamesonstone/mint/pkg/promotion"
	"os"
)

func (o *productionOperation) control(ctx context.Context) (any, bool, error) {
	if err := o.client.AuthorizeControlRun(ctx, o.config, o.flags.RunID, o.flags.Event); err != nil {
		return nil, false, err
	}
	data, err := os.ReadFile(o.flags.Input)
	if err != nil {
		return nil, false, err
	}
	r, err := o.client.ControlEvent(ctx, o.config, o.flags.Event, data)
	if err != nil {
		return nil, false, err
	}
	o.flags.FixPR, o.flags.Issue, o.flags.Version, o.flags.Reason = r.FixPR, r.Issue, r.Version, r.Reason
	switch r.Operation {
	case "hotfix":
		return o.requestHotfix(ctx)
	case "rollback":
		key := fmt.Sprintf("rollback-control-%d", o.flags.RunID)
		s := &o.snapshot.State
		if prior, ok := s.ControlRequests[key]; ok {
			if current, exists := s.Proposals[prior.Kind]; exists && current.ID == prior.ID {
				prior = current
			}
			for _, i := range s.Intents {
				if i.ProposalID == prior.ID {
					prior.State = i.Status
				}
			}
			return prior, false, nil
		}
		value, _, err := o.requestRollback(ctx)
		if err != nil {
			return nil, false, err
		}
		if s.ControlRequests == nil {
			s.ControlRequests = map[string]promotion.Proposal{}
		}
		s.ControlRequests[key] = value.(promotion.Proposal)
		return value, true, nil
	default:
		return map[string]string{"state": "idle", "next_action": "No hotfix request on this event."}, false, nil
	}
}
