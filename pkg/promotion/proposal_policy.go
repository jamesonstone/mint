package promotion

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v3"
)

// proposalPolicy records an explicit selection in the same approval PR. The
// repository file remains canonical; journal selection is not a hidden override.
func (c Client) proposalPolicy(ctx context.Context, cfg Config, s State, p Proposal, base string) (string, bool, error) {
	if s.Schema != 2 || cfg.PolicyPath == "" {
		return "", false, nil
	}
	data, err := c.ReadFile(ctx, cfg.PolicyPath, base)
	if err != nil {
		return "", false, err
	}
	policy, err := ParsePolicy(data)
	if err != nil {
		return "", false, err
	}
	env, ok := policy.Environments[cfg.Environment]
	if !ok {
		return "", false, fmt.Errorf("proposal environment missing from policy")
	}
	candidate := s.Candidates[p.CandidateSHA]
	original := env
	env.Operation, env.Reason = "release", p.Summary
	if p.Selection == "pinned" {
		env.Target, env.Follow = candidate.Version, ""
	} else {
		env.Target, env.Follow = "", "latest"
	}
	if p.Kind == "hotfix" || p.Kind == "rollback" {
		env.Operation = p.Kind
		if env.Reason == "" {
			env.Reason = "Promote the reviewed " + p.Kind + " " + candidate.Version
		}
	}
	if original.Target == env.Target && original.Follow == env.Follow && original.Operation == env.Operation && original.Reason == env.Reason {
		return "", false, nil
	}
	policy.Environments[cfg.Environment] = env
	serialized, err := yaml.Marshal(policy)
	return string(serialized), err == nil, err
}
