package cli

import (
	"context"
	"fmt"
	"github.com/jamesonstone/mint/pkg/promotion"
	"strings"
)

// review validates automation-authored control PRs without requiring a review
// before CI runs. Independent current-head approval is enforced when freezing.
func (o *productionOperation) review(ctx context.Context) (any, bool, error) {
	pr, err := o.client.Pull(ctx, o.flags.PR)
	if err != nil {
		return nil, false, err
	}
	if value, handled, err := o.reviewPolicyChange(ctx, pr); handled || err != nil {
		return value, false, err
	}
	s := o.snapshot.State
	for _, p := range s.Proposals {
		marker := "<!-- mint:proposal:" + p.ID + ":" + p.Kind + " -->"
		if strings.Contains(pr.Body, marker) {
			return o.intent(ctx, false)
		}
	}
	if strings.Contains(pr.Body, "<!-- mint:hotfix-source:") {
		if s.Baseline == nil || !o.client.IsReleaseAuthor(pr.User.Login) || !strings.HasPrefix(pr.Base.Ref, "mint-hotfix-base/GH-") {
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
