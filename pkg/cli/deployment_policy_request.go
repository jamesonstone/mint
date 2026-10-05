package cli

import (
	"context"
	"fmt"
	"github.com/jamesonstone/mint/pkg/promotion"
	"strings"
)

// The reviewed YAML PR is itself the release approval artifact. It never opens
// another PR, and uses checks and runtime policy from before the change.
func (o *productionOperation) policyRequest(ctx context.Context, freeze bool) (any, bool, error) {
	pr, err := o.client.Pull(ctx, o.flags.PR)
	if err != nil {
		return nil, false, err
	}
	if pr.Base.Ref != o.config.DefaultBranch {
		return nil, false, fmt.Errorf("request must target the configured default branch")
	}
	beforeSHA, afterSHA := pr.Base.SHA, pr.Head.SHA
	if freeze {
		if !pr.Merged || pr.MergeSHA != o.flags.MergeSHA {
			return nil, false, fmt.Errorf("request must identify its exact reviewed merge")
		}
		beforeSHA, err = o.client.PolicyParent(ctx, pr.MergeSHA)
		if err != nil {
			return nil, false, err
		}
		afterSHA = pr.MergeSHA
	}
	oldData, err := o.client.ReadFile(ctx, o.flags.Config, beforeSHA)
	if err != nil {
		return nil, false, err
	}
	newData, err := o.client.ReadFile(ctx, o.flags.Config, afterSHA)
	if err != nil {
		return nil, false, err
	}
	before, err := promotion.ParsePolicy(oldData)
	if err != nil {
		return nil, false, err
	}
	after, err := promotion.ParsePolicy(newData)
	if err != nil {
		return nil, false, err
	}
	name, err := promotion.PolicyRequestEnvironment(before, after)
	if err != nil {
		return nil, false, err
	}
	if err := o.client.CheckPolicyRequestSource(ctx, pr, o.flags.Config, before, name); err != nil {
		return nil, false, err
	}
	if name != o.config.Environment {
		return nil, false, fmt.Errorf("request belongs to environment %s", name)
	}
	trusted, err := before.ForEnvironment(name)
	if err != nil {
		return nil, false, err
	}
	requested, err := after.ForEnvironment(name)
	if err != nil {
		return nil, false, err
	}
	if trusted.Mode != "deployment" || trusted.Scope != "shared" {
		return nil, false, fmt.Errorf("YAML request requires a shared deployment environment")
	}
	allowed := append([]string{o.flags.Config}, trusted.ControlPaths...)
	if err := o.client.CheckProposalFiles(ctx, pr.Number, allowed); err != nil {
		return nil, false, err
	}
	if freeze {
		if err := o.client.CheckHead(ctx, pr, trusted.RequiredChecks); err != nil {
			return nil, false, err
		}
	}
	s := &o.snapshot.State
	for _, i := range s.Intents {
		if i.MergeSHA == pr.MergeSHA && freeze {
			return i, false, nil
		}
	}
	if s.Baseline == nil || s.InFlight != "" {
		return nil, false, fmt.Errorf("request requires a verified idle environment baseline")
	}
	if requested.Operation == "resume" {
		if !freeze {
			return map[string]string{"state": "resume requested"}, false, nil
		}
		return s.Resume("resume-policy-" + pr.MergeSHA)
	}
	if trusted.Configuration == "" {
		return nil, false, fmt.Errorf("trusted environment runtime configuration must be activated before requesting deployment")
	}
	s.Schema, s.Configuration, s.Publish, s.Target = 2, trusted.Configuration, trusted.Publish, requested.Target
	o.config = trusted
	trusted.ControlPaths = append(trusted.ControlPaths, o.flags.Config)
	trusted.PolicyPath = o.flags.Config
	s.PolicyDigest, err = promotion.AuthorityDigest(trusted)
	if err != nil {
		return nil, false, err
	}
	if requested.Target == "" {
		if requested.Operation == "release" && strings.Contains(pr.Body, "<!-- mint:proposal:") {
			return o.canonicalPolicyRequest(ctx, pr, afterSHA, freeze)
		}
		return nil, false, fmt.Errorf("YAML deployment approval requires an exact target; use operation: resume with follow: latest to restart the updating release PR")
	}
	kind := "normal"
	if requested.Operation == "hotfix" {
		kind = "hotfix"
	}
	for _, p := range s.Proposals {
		if p.State == "open" && p.PR != 0 && p.PR != pr.Number {
			return nil, false, fmt.Errorf("edit the active release PR #%d or close it before a separate YAML request", p.PR)
		}
	}
	id := fmt.Sprintf("%s:policy-pr:%d", name, pr.Number)
	var p promotion.Proposal
	if requested.Operation == "rollback" {
		target, resolveErr := s.ResolveRollbackTarget(requested.Target)
		if resolveErr != nil {
			return nil, false, resolveErr
		}
		p, err = s.ReconcileRollback(target.Candidate.SourceSHA, id, requested.Reason)
	} else {
		pin := ""
		if requested.Target != "" {
			for sha, c := range s.Candidates {
				if c.Version == requested.Target && c.Kind == kind {
					if pin != "" && pin != sha {
						return nil, false, fmt.Errorf("target version has ambiguous builds")
					}
					pin = sha
				}
			}
			if pin == "" {
				return nil, false, fmt.Errorf("target version has no authenticated candidate")
			}
		}
		p, err = s.Reconcile(kind, "candidate", pin, id, requested.Reason, o.proof)
	}
	if err != nil {
		return nil, false, err
	}
	p.PR, p.Branch, p.HeadSHA = pr.Number, pr.Head.Ref, pr.Head.SHA
	s.Proposals[p.Kind] = p
	candidate := s.Candidates[p.CandidateSHA]
	if err := o.proof.VerifyTag(candidate.Version, candidate.SourceSHA); err != nil {
		return nil, false, err
	}
	if !freeze {
		return map[string]any{"environment": name, "candidate": candidate, "configuration": trusted.Configuration, "kind": p.Kind}, false, nil
	}
	intent, err := s.FreezeIntent(p, candidate, pr.MergeSHA)
	return intent, err == nil, err
}
