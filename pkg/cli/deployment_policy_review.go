package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jamesonstone/mint/pkg/promotion"
)

func (o *productionOperation) reviewPolicyChange(ctx context.Context, pr promotion.PullRequest) (any, bool, error) {
	if o.config.Schema != 2 {
		return nil, false, nil
	}
	oldData, err := o.client.ReadFile(ctx, o.flags.Config, pr.Base.SHA)
	if err != nil {
		return nil, false, err
	}
	newData, err := o.client.ReadFile(ctx, o.flags.Config, pr.Head.SHA)
	if err != nil {
		return nil, false, err
	}
	if string(oldData) == string(newData) {
		return nil, false, nil
	}
	before, err := promotion.ParsePolicy(oldData)
	if err != nil {
		return nil, false, err
	}
	after, err := promotion.ParsePolicy(newData)
	if err != nil {
		return nil, false, err
	}
	_, err = promotion.PolicyRequestEnvironment(before, after)
	if err == nil {
		value, _, err := o.policyRequest(ctx, false)
		return value, true, err
	}
	if strings.Contains(pr.Body, "<!-- mint:proposal:") {
		for name, env := range before.Environments {
			next, ok := after.Environments[name]
			if !ok || env.Target != next.Target || env.Follow != next.Follow || env.Operation != next.Operation || env.Reason != next.Reason {
				return nil, true, err
			}
		}
		// Control proposals must not carry authority changes alongside their release.
		left, _ := json.Marshal(before)
		right, _ := json.Marshal(after)
		if string(left) != string(right) {
			return nil, true, err
		}
	}
	return nil, false, nil
}

// Generated latest-follow PRs contain canonical artifact evidence on the exact
// reviewed head. A plain YAML request needs an exact target instead.
func (o *productionOperation) canonicalPolicyRequest(ctx context.Context, pr promotion.PullRequest, ref string, freeze bool) (any, bool, error) {
	data, err := o.client.ReadFile(ctx, ".mint/proposal.json", ref)
	if err != nil {
		return nil, false, err
	}
	var declaration promotion.Declaration
	if err := json.Unmarshal(data, &declaration); err != nil {
		return nil, false, err
	}
	s := &o.snapshot.State
	p, err := s.ValidateDeclaration(declaration)
	if err != nil {
		return nil, false, err
	}
	marker := fmt.Sprintf("<!-- mint:proposal:%s:%s -->", p.ID, p.Kind)
	if declaration.Kind != "normal" || declaration.Selection != "latest" || p.PR != pr.Number || p.HeadSHA != pr.Head.SHA || p.Branch != pr.Head.Ref || !strings.Contains(pr.Body, marker) || !o.client.IsReleaseAuthor(pr.User.Login) {
		return nil, false, fmt.Errorf("latest request lacks exact reviewed canonical proposal identity")
	}
	if err := o.proof.VerifyTag(declaration.Candidate.Version, declaration.Candidate.SourceSHA); err != nil {
		return nil, false, err
	}
	if !freeze {
		return declaration, false, nil
	}
	intent, err := s.FreezeIntent(p, declaration.Candidate, pr.MergeSHA)
	return intent, err == nil, err
}
