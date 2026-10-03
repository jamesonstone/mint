package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jamesonstone/mint/pkg/promotion"
	"reflect"
	"strings"
)

// review validates automation-authored control PRs without requiring a review
// before CI runs. Independent current-head approval is enforced when freezing.
func (o *productionOperation) review(ctx context.Context) (any, bool, error) {
	pr, err := o.client.Pull(ctx, o.flags.PR)
	if err != nil {
		return nil, false, err
	}
	s := o.snapshot.State
	for _, p := range s.Proposals {
		marker := "<!-- mint:proposal:" + p.ID + ":" + p.Kind + " -->"
		if strings.Contains(pr.Body, marker) {
			return o.intent(ctx, false)
		}
	}
	for _, i := range s.Intents {
		marker := "<!-- mint:proposal:" + i.ID + ":" + i.Status + ":status -->"
		if !strings.Contains(pr.Body, marker) {
			continue
		}
		if pr.User.Login != o.config.HumanLogin || pr.Base.Ref != o.config.DefaultBranch {
			return nil, false, fmt.Errorf("foreign deployment-status PR")
		}
		data, err := o.client.ReadFile(ctx, ".mint/status.json", pr.Head.SHA)
		if err != nil {
			return nil, false, err
		}
		var recorded promotion.Intent
		if err := json.Unmarshal(data, &recorded); err != nil {
			return nil, false, err
		}
		if !reflect.DeepEqual(recorded, i) {
			return nil, false, fmt.Errorf("status PR differs from observed production outcome")
		}
		if err := o.client.CheckProposalFiles(ctx, pr.Number, o.config.ControlPaths); err != nil {
			return nil, false, err
		}
		return recorded, false, nil
	}
	if strings.Contains(pr.Body, "<!-- mint:hotfix-source:") {
		if s.Baseline == nil || pr.User.Login != o.config.HumanLogin || !strings.HasPrefix(pr.Base.Ref, "mint-hotfix-base/GH-") {
			return nil, false, fmt.Errorf("hotfix source identity mismatch")
		}
		c := promotion.Candidate{Kind: "hotfix", SourceSHA: pr.Head.SHA, SourcePR: pr.Number, BaselineID: s.Baseline.ID, Version: s.Baseline.Candidate.Version}
		changes, err := o.client.CollectChanges(ctx, o.proof, s, c, o.config.ControlPaths)
		return changes, false, err
	}
	if strings.Contains(pr.Body, "<!-- mint:") {
		return nil, false, fmt.Errorf("unknown Mint control identity")
	}
	return map[string]string{"status": "ordinary source PR; native checks apply"}, false, nil
}
