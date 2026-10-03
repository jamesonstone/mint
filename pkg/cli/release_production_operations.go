package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jamesonstone/mint/pkg/promotion"
	"strings"
)

func (o *productionOperation) propose(ctx context.Context) (any, bool, error) {
	s := &o.snapshot.State
	f := o.flags
	if s.Baseline == nil {
		return nil, false, fmt.Errorf("verified production baseline is required")
	}
	prior, exists := s.Proposals[f.Kind]
	if exists && prior.PR > 0 && prior.State != "deployed" {
		pr, err := o.client.Pull(ctx, prior.PR)
		if err != nil {
			return nil, false, err
		}
		if pr.State == "open" && f.Event == "close" {
			f.Event = "candidate"
		}
		if prior.State == "paused" && pr.State == "open" {
			f.Event = "reopen"
		}
		if pr.Merged {
			return nil, false, fmt.Errorf("merged proposal outcome must be reconciled before updating")
		}
		if prior.State == "paused" && pr.State == "closed" && f.Event != "reopen" {
			return prior, false, nil
		}
		if pr.State == "closed" && f.Event != "reopen" {
			f.Event = "close"
		}
		if f.Event == "reopen" && pr.State != "open" {
			return nil, false, fmt.Errorf("reopen the paused PR in GitHub before reconciling")
		}
		summary, err := o.client.ReadFile(ctx, ".mint/summary.md", pr.Head.SHA)
		if err != nil {
			return nil, false, err
		}
		if f.Summary == "" {
			f.Summary = strings.TrimSpace(string(summary))
		}
		data, err := o.client.ReadFile(ctx, ".mint/proposal.json", pr.Head.SHA)
		if err != nil {
			return nil, false, err
		}
		var d promotion.Declaration
		if err := json.Unmarshal(data, &d); err != nil {
			return nil, false, err
		}
		if d.ProposalID != prior.ID || d.Repository != s.Repository {
			return nil, false, fmt.Errorf("manual declaration changed proposal identity")
		}
		if d.Selection == "pinned" {
			f.Pin = d.Candidate.SourceSHA
		}
	}
	generation := fmt.Sprintf("%s-%s", f.Kind, s.Baseline.ID)
	p, err := s.Reconcile(f.Kind, f.Event, f.Pin, generation, f.Summary, o.proof)
	if errors.Is(err, promotion.ErrNoEligible) || errors.Is(err, promotion.ErrAlreadyDeployed) {
		return map[string]string{"state": "idle"}, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if p.State == "paused" {
		return p, true, nil
	}
	p, err = o.client.SyncProposal(ctx, o.config, s, p)
	return p, true, err
}
func (o *productionOperation) intent(ctx context.Context, freeze bool) (any, bool, error) {
	pr, err := o.client.Pull(ctx, o.flags.PR)
	if err != nil {
		return nil, false, err
	}
	s := &o.snapshot.State
	if freeze && (!pr.Merged || pr.MergeSHA != o.flags.MergeSHA) {
		return nil, false, fmt.Errorf("intent must read the exact merged proposal SHA")
	}
	ref := pr.Head.SHA
	if freeze {
		ref = pr.MergeSHA
	}
	data, err := o.client.ReadFile(ctx, ".mint/proposal.json", ref)
	if err != nil {
		return nil, false, err
	}
	var d promotion.Declaration
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, false, err
	}
	if freeze {
		for _, frozen := range s.Intents {
			if frozen.MergeSHA == pr.MergeSHA && frozen.ProposalID == d.ProposalID {
				left, _ := json.Marshal(frozen.Candidate)
				right, _ := json.Marshal(d.Candidate)
				if string(left) != string(right) || frozen.Notes != d.Notes {
					return nil, false, fmt.Errorf("merged declaration conflicts with immutable intent")
				}
				return frozen, false, nil
			}
		}
	}
	p, err := s.ValidateDeclaration(d)
	if err != nil {
		return nil, false, err
	}
	if p.PR != pr.Number || p.HeadSHA != pr.Head.SHA || p.Branch != pr.Head.Ref || pr.Base.Ref != o.config.DefaultBranch || !o.client.IsReleaseAuthor(pr.User.Login) || !strings.Contains(pr.Body, "<!-- mint:proposal:"+p.ID+":"+p.Kind+" -->") {
		return nil, false, fmt.Errorf("proposal PR/head/author identity mismatch")
	}
	if freeze {
		if err := o.client.CheckHead(ctx, pr, o.config.RequiredChecks); err != nil {
			return nil, false, err
		}
	}
	if err := o.client.CheckProposalFiles(ctx, pr.Number, o.config.ControlPaths); err != nil {
		return nil, false, err
	}
	if err := o.proof.VerifyTag(d.Candidate.Version, d.Candidate.SourceSHA); err != nil {
		return nil, false, err
	}
	if !freeze {
		return d, false, nil
	}
	i, err := s.FreezeIntent(p, d.Candidate, pr.MergeSHA)
	return i, true, err
}
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
	outcome := "failure"
	if run.Conclusion == "success" {
		var manifest promotion.DeploymentManifest
		if err := o.client.RunManifest(ctx, f.RunID, "mint-deployment", &manifest); err != nil {
			return nil, false, err
		}
		if !manifest.Verified || manifest.IntentID != i.ID || manifest.SourceSHA != i.Candidate.SourceSHA || manifest.ArtifactDigest != i.Candidate.Artifact.Digest || manifest.Configuration != i.Candidate.Artifact.Configuration {
			return nil, false, fmt.Errorf("deployment manifest failed exact artifact verification")
		}
		outcome = "success"
	}
	err = s.FinishDeployment(i.ID, outcome, run.HTMLURL, o.proof)
	return s.Intents[i.ID], true, err
}
func (o *productionOperation) bootstrap(ctx context.Context) (any, bool, error) {
	s := &o.snapshot.State
	if s.Baseline != nil {
		return nil, false, fmt.Errorf("production baseline already exists; do not re-import")
	}
	var baseline promotion.Baseline
	if err := o.client.RunManifest(ctx, o.flags.RunID, "mint-baseline", &baseline); err != nil {
		return nil, false, err
	}
	run, err := o.client.TrustedRun(ctx, o.flags.RunID, o.config.BaselineWorkflow)
	if err != nil {
		return nil, false, err
	}
	if run.HeadBranch != o.config.DefaultBranch || baseline.ID == "" || baseline.Candidate.Repository != s.Repository || baseline.Candidate.Environment != s.Environment || baseline.DeploymentURL != run.HTMLURL {
		return nil, false, fmt.Errorf("baseline is not bound to verified deployment evidence")
	}
	build, err := o.client.TrustedRun(ctx, baseline.Candidate.RunID, o.config.BaselineBuildWorkflow)
	if err != nil {
		return nil, false, err
	}
	if build.HeadBranch != o.config.DefaultBranch || build.HeadSHA != baseline.Candidate.SourceSHA || baseline.Candidate.RunURL != build.HTMLURL || baseline.Candidate.Kind != "normal" || baseline.PublicationPending || len(baseline.Shipped) != 0 {
		return nil, false, fmt.Errorf("bootstrap lacks exact successful source-build and verified runtime evidence")
	}
	if err := promotion.ValidateBaselineCandidate(baseline.Candidate); err != nil {
		return nil, false, err
	}
	if err := o.proof.VerifyTag(baseline.Candidate.Version, baseline.Candidate.SourceSHA); err != nil {
		return nil, false, err
	}
	baseline.MainAnchor = baseline.Candidate.SourceSHA
	s.History[baseline.ID] = baseline
	s.Candidates[baseline.Candidate.SourceSHA] = baseline.Candidate
	s.Baseline = &baseline
	s.MainAnchor = baseline.Candidate.SourceSHA
	return baseline, true, nil
}
func (o *productionOperation) hotfix(ctx context.Context) (any, bool, error) {
	s := &o.snapshot.State
	if s.Baseline == nil || s.InFlight != "" {
		return nil, false, fmt.Errorf("hotfix preparation requires a verified idle production baseline")
	}
	var request struct {
		Issue      int
		BaselineID string
		Fixes      []string
	}
	if err := readJSON(o.flags.Input, &request); err != nil {
		return nil, false, err
	}
	if request.BaselineID != s.Baseline.ID {
		return nil, false, fmt.Errorf("hotfix request baseline is stale")
	}
	if err := o.client.VerifyIssue(ctx, request.Issue, o.config.HumanLogin); err != nil {
		return nil, false, err
	}
	recovered, err := o.client.RecoverHotfixSource(ctx, *s.Baseline, request.Issue, request.Fixes)
	if err != nil {
		return nil, false, err
	}
	var result promotion.HotfixSource
	if recovered != nil {
		result = *recovered
	} else {
		result, err = promotion.PrepareHotfix(ctx, promotion.HotfixOptions{WorkDir: o.proof.WorkDir, Baseline: *s.Baseline, Issue: request.Issue, Fixes: request.Fixes, CommitterName: promotion.AutomationLogin, CommitterEmail: promotion.AutomationEmail})
		if err != nil {
			return result, false, err
		}
	}
	pull, err := o.client.PublishHotfixSource(ctx, result, *s.Baseline, request.Issue)
	if err != nil {
		return result, false, err
	}
	if err := o.client.DispatchChecks(ctx, o.config.ValidationWorkflow, result.Branch, pull.Number); err != nil {
		return pull, false, err
	}
	return pull, false, nil
}

func (o *productionOperation) start(ctx context.Context) (any, bool, error) {
	s := &o.snapshot.State
	f := o.flags
	i, ok := s.Intents[f.IntentID]
	if !ok {
		return nil, false, fmt.Errorf("unknown intent")
	}
	if i.Status == "deployed" || i.Status == "publication_pending" {
		return i, false, nil
	}
	run, err := o.client.ObservedRun(ctx, f.RunID, o.config.PromotionWorkflow)
	if err != nil {
		return nil, false, err
	}
	if run.HeadSHA != i.MergeSHA || run.Status != "in_progress" {
		return nil, false, fmt.Errorf("active promotion run must identify the exact merged intent")
	}
	if i.DeploymentRunID != 0 && i.Status == "deploying" && i.DeploymentRunID != run.ID {
		return nil, false, fmt.Errorf("prior deployment outcome must be reconciled before retry")
	}
	out, err := s.StartDeployment(i.ID, o.proof)
	if err != nil {
		return nil, false, err
	}
	out.DeploymentRunID = run.ID
	s.Intents[out.ID] = out
	return out, true, nil
}
