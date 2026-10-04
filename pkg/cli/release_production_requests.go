package cli

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/jamesonstone/mint/pkg/promotion"
)

func (o *productionOperation) requestHotfix(ctx context.Context) (any, bool, error) {
	s := &o.snapshot.State
	f := o.flags
	if s.Baseline == nil || s.InFlight != "" {
		return nil, false, fmt.Errorf("hotfix requires verified idle production; inspect the current deployment first")
	}
	if (f.FixPR > 0) == (f.Issue > 0) || f.FixPR < 0 || f.Issue < 0 {
		return nil, false, fmt.Errorf("select exactly one fix PR or issue for a newly authored fix")
	}
	if strings.TrimSpace(f.Reason) == "" {
		return nil, false, fmt.Errorf("hotfix reason is required")
	}
	issue := f.Issue
	var fixes []string
	if f.FixPR > 0 {
		if _, err := o.client.Pull(ctx, f.FixPR); err != nil {
			return nil, false, err
		}
		fetch := exec.CommandContext(ctx, "git", "fetch", "--no-tags", "origin", fmt.Sprintf("refs/pull/%d/head", f.FixPR))
		fetch.Dir = o.proof.WorkDir
		if err := fetch.Run(); err != nil {
			return nil, false, fmt.Errorf("fix PR source unavailable; fetch its reviewed head and full default-branch history")
		}
		pr, resolved, err := o.client.HotfixFixes(ctx, o.config, f.FixPR, o.proof)
		if err != nil {
			return nil, false, err
		}
		fixes = resolved
		issue, err = o.client.EnsureHotfixRequestIssue(ctx, pr, *s.Baseline, f.Reason)
		if err != nil {
			return nil, false, err
		}
	}
	return o.prepareHotfix(ctx, issue, s.Baseline.ID, fixes)
}

func (o *productionOperation) requestRollback(ctx context.Context) (any, bool, error) {
	if strings.TrimSpace(o.flags.Reason) == "" {
		return nil, false, fmt.Errorf("rollback reason is required; prefer a corrective or revert hotfix when practical")
	}
	target, err := o.snapshot.State.ResolveRollbackTarget(o.flags.Version)
	if err != nil {
		return nil, false, err
	}
	o.flags.Pin = target.Candidate.SourceSHA
	o.flags.Summary = o.flags.Reason
	return o.execute(ctx, "propose-rollback")
}

func oIsProposal(value any) bool { _, ok := value.(promotion.Proposal); return ok }
