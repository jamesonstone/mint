package promotion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func (c Client) collectHotfixChanges(ctx context.Context, g GitProof, s State, candidate Candidate) ([]Change, error) {
	if s.Baseline == nil || candidate.BaselineID != s.Baseline.ID {
		return nil, fmt.Errorf("hotfix production baseline changed")
	}
	data, err := c.ReadFile(ctx, ".mint/hotfix.json", candidate.SourceSHA)
	if err != nil {
		return nil, err
	}
	var meta struct {
		BaselineID  string   `json:"baseline_id"`
		BaselineSHA string   `json:"baseline_sha"`
		Fixes       []string `json:"fixes"`
		PatchIDs    []string `json:"patch_ids"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	if meta.BaselineID != s.Baseline.ID || meta.BaselineSHA != s.Baseline.Candidate.SourceSHA || len(meta.Fixes) != len(meta.PatchIDs) {
		return nil, fmt.Errorf("hotfix metadata does not identify reviewed production base")
	}
	if err := verifyHotfixRequestMetadata(g, meta.BaselineSHA, candidate.SourceSHA, data); err != nil {
		return nil, err
	}
	changes := []Change{}
	if len(meta.Fixes) == 0 {
		// Newly authored fixes must remain a production-base branch. Ancestry
		// from queued main commits is not an isolation proof.
		commits, err := g.git(nil, "rev-list", meta.BaselineSHA+".."+candidate.SourceSHA, "--first-parent")
		if err != nil {
			return nil, err
		}
		for _, commit := range strings.Fields(string(commits)) {
			parents, err := g.git(nil, "rev-list", "--parents", "-n", "1", commit)
			if err != nil {
				return nil, err
			}
			fields := strings.Fields(string(parents))
			if len(fields) > 2 && (commit != candidate.SourceSHA || fields[1] != meta.BaselineSHA) {
				return nil, fmt.Errorf("authored hotfix merged queued or unrelated ancestry")
			}
			if g.DefaultBranch == "" {
				return nil, fmt.Errorf("authored hotfix requires the configured default branch for isolation proof")
			}
			main, err := g.git(nil, "rev-parse", "--verify", "refs/remotes/origin/"+g.DefaultBranch)
			if err != nil {
				return nil, err
			}
			shared, err := g.Ancestor(commit, strings.TrimSpace(string(main)))
			if err != nil {
				return nil, err
			}
			if shared {
				return nil, fmt.Errorf("authored hotfix includes queued main ancestry; select explicit fixes")
			}
		}
		id, err := g.PatchID(candidate.SourceSHA)
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(id, "empty:") {
			return nil, fmt.Errorf("authored hotfix contains no application change")
		}
		ch := Change{SHA: candidate.SourceSHA, Version: candidate.Version, PR: candidate.SourcePR, Title: "Isolated production hotfix", PatchID: id, Hotfix: true}
		attributeReviewedRevert(g, s, &ch)
		return []Change{ch}, nil
	}
	for n, fix := range meta.Fixes {
		id, err := g.PatchID(fix)
		if err != nil || id != meta.PatchIDs[n] {
			return nil, fmt.Errorf("original fix patch identity changed")
		}
	}
	expected, err := PrepareHotfix(ctx, HotfixOptions{WorkDir: g.WorkDir, Baseline: *s.Baseline, Issue: 1, Fixes: meta.Fixes, CommitterName: AutomationLogin, CommitterEmail: AutomationEmail})
	if expected.WorkDir != "" {
		defer func() { _ = os.RemoveAll(expected.WorkDir) }()
	}
	if err != nil {
		if expected.Conflict {
			return c.collectResolvedHotfix(ctx, g, s, candidate, data)
		}
		return nil, err
	}
	expectedProof := GitProof{Context: ctx, WorkDir: expected.WorkDir}
	expectedPatch, err := expectedProof.git(nil, "diff", meta.BaselineSHA, expected.SourceSHA, "--", ".", ":(exclude).mint")
	if err != nil {
		return nil, err
	}
	actualPatch, err := g.git(nil, "diff", meta.BaselineSHA, candidate.SourceSHA, "--", ".", ":(exclude).mint")
	if err != nil {
		return nil, err
	}
	expectedID, err := g.git(expectedPatch, "patch-id", "--stable")
	if err != nil {
		return nil, err
	}
	actualID, err := g.git(actualPatch, "patch-id", "--stable")
	if err != nil {
		return nil, err
	}
	if string(expectedID) != string(actualID) {
		return nil, fmt.Errorf("hotfix source diff includes changes outside the explicit fixes")
	}
	for n, fix := range meta.Fixes {
		id, err := g.PatchID(fix)
		if err != nil {
			return nil, err
		}
		if id != meta.PatchIDs[n] {
			return nil, fmt.Errorf("original fix patch identity changed")
		}
		title, err := g.git(nil, "show", "-s", "--format=%s", fix)
		if err != nil {
			return nil, err
		}
		ch := Change{SHA: fix, Version: candidate.Version, Title: strings.TrimSpace(string(title)), PatchID: id, Hotfix: true}
		attributeReviewedRevert(g, s, &ch)
		retained, err := g.Retains(candidate.SourceSHA, ch)
		if err != nil {
			return nil, err
		}
		if !retained {
			return nil, fmt.Errorf("hotfix tree does not retain exactly requested fix %s", fix)
		}
		tags, err := g.git(nil, "tag", "--points-at", fix)
		if err != nil {
			return nil, err
		}
		for _, tag := range strings.Fields(string(tags)) {
			if versionPattern.MatchString(tag) {
				if ch.OriginalVersion != "" {
					return nil, fmt.Errorf("original fix has ambiguous version")
				}
				ch.OriginalVersion = tag
			}
		}
		pulls, err := c.MergedPRs(ctx, fix)
		if err != nil {
			return nil, err
		}
		for _, pr := range pulls {
			if pr.MergeSHA == fix {
				if ch.PR != 0 {
					return nil, fmt.Errorf("original fix PR mapping is ambiguous")
				}
				ch.PR = pr.Number
			}
		}
		changes = append(changes, ch)
	}
	// A reviewed source PR may not smuggle queued features into an explicit-pick hotfix.
	parents, err := g.git(nil, "rev-list", "--first-parent", meta.BaselineSHA+".."+candidate.SourceSHA)
	if err != nil {
		return nil, err
	}
	// Preparation emits one isolated commit, and GitHub may add one merge commit.
	if count := len(strings.Fields(string(parents))); count > 2 {
		return nil, fmt.Errorf("hotfix includes unexpected source ancestry; isolate and review its patches")
	}
	return changes, nil
}

// Bind classification to the original baseline-child preparation, not metadata
// edited on a later reviewed head. Inspect both sides of a final merge.
func verifyHotfixRequestMetadata(g GitProof, baseline, source string, metadata []byte) error {
	ancestor, err := g.Ancestor(baseline, source)
	if err != nil || !ancestor {
		return fmt.Errorf("hotfix request must descend from verified production")
	}
	commits, err := g.git(nil, "rev-list", "--reverse", "--topo-order", baseline+".."+source)
	if err != nil {
		return err
	}
	preparations := 0
	for _, commit := range strings.Fields(string(commits)) {
		parents, err := hotfixParents(g, commit)
		if err != nil || len(parents) == 0 || len(parents) > 2 {
			return fmt.Errorf("hotfix request ancestry is unavailable")
		}
		if len(parents) == 2 {
			if commit != source || parents[0] != baseline {
				return fmt.Errorf("hotfix includes queued or unrelated merge ancestry")
			}
			continue
		}
		if parents[0] != baseline {
			continue
		}
		original, err := g.git(nil, "show", commit+":.mint/hotfix.json")
		if err != nil || !bytes.Equal(original, metadata) {
			return fmt.Errorf("hotfix request metadata differs from original preparation")
		}
		preparations++
	}
	if preparations != 1 {
		return fmt.Errorf("hotfix must have one authenticated baseline-child preparation")
	}
	return nil
}

// A conflict resolution is newly reviewed authored code. Its actual patch,
// rather than the conflicting original patch, becomes shipped provenance.
func (c Client) collectResolvedHotfix(ctx context.Context, g GitProof, s State, candidate Candidate, metadata []byte) ([]Change, error) {
	base := s.Baseline.Candidate.SourceSHA
	ancestor, err := g.Ancestor(base, candidate.SourceSHA)
	if err != nil || !ancestor {
		return nil, fmt.Errorf("resolved hotfix must descend from verified production")
	}
	var repo struct {
		FullName      string `json:"full_name"`
		DefaultBranch string `json:"default_branch"`
	}
	status, err := c.request(ctx, "GET", "repos/"+c.Repository, nil, &repo)
	if err != nil || status != 200 || repo.FullName != c.Repository || repo.DefaultBranch == "" {
		return nil, fmt.Errorf("resolved hotfix requires authoritative default-branch identity")
	}
	main, err := g.git(nil, "rev-parse", "--verify", "refs/remotes/origin/"+repo.DefaultBranch)
	if err != nil {
		return nil, err
	}
	commits, err := g.git(nil, "rev-list", "--reverse", "--topo-order", base+".."+candidate.SourceSHA)
	if err != nil {
		return nil, err
	}
	application, preparations := 0, 0
	for _, commit := range strings.Fields(string(commits)) {
		shared, err := g.Ancestor(commit, strings.TrimSpace(string(main)))
		if err != nil || shared {
			return nil, fmt.Errorf("resolved hotfix includes queued default-branch ancestry")
		}
		parents, err := hotfixParents(g, commit)
		if err != nil || len(parents) == 0 || len(parents) > 2 {
			return nil, fmt.Errorf("resolved hotfix source ancestry is unavailable")
		}
		if len(parents) == 2 {
			if commit != candidate.SourceSHA || parents[0] != base {
				return nil, fmt.Errorf("resolved hotfix includes unrelated merge ancestry")
			}
			mergedTree, err := g.git(nil, "rev-parse", commit+"^{tree}")
			sourceTree, sourceErr := g.git(nil, "rev-parse", parents[1]+"^{tree}")
			if err != nil || sourceErr != nil || !bytes.Equal(mergedTree, sourceTree) {
				return nil, fmt.Errorf("resolved hotfix merge differs from reviewed source tree")
			}
			continue
		}
		paths, err := g.git(nil, "diff", "--name-only", parents[0], commit, "--", ".", ":(exclude).mint")
		if err != nil {
			return nil, err
		}
		if len(strings.Fields(string(paths))) > 0 {
			controls, err := g.git(nil, "diff", "--name-only", parents[0], commit, "--", ".mint")
			if err != nil || len(strings.Fields(string(controls))) != 0 {
				return nil, fmt.Errorf("resolved application commit must preserve request metadata")
			}
			application++
			continue
		}
		controls, err := g.git(nil, "diff", "--name-only", parents[0], commit)
		if err != nil || parents[0] != base || strings.TrimSpace(string(controls)) != ".mint/hotfix.json" {
			return nil, fmt.Errorf("resolved hotfix preparation must contain only immutable request metadata")
		}
		original, err := g.git(nil, "show", commit+":.mint/hotfix.json")
		if err != nil || !bytes.Equal(original, metadata) {
			return nil, fmt.Errorf("resolved hotfix request metadata changed")
		}
		preparations++
	}
	if application != 1 || preparations != 1 {
		return nil, fmt.Errorf("commit the reviewed resolved application patch once on the metadata-only recovery branch")
	}
	id, err := g.PatchID(candidate.SourceSHA)
	if err != nil || strings.HasPrefix(id, "empty:") {
		return nil, fmt.Errorf("resolved hotfix contains no application patch")
	}
	ch := Change{SHA: candidate.SourceSHA, Version: candidate.Version, PR: candidate.SourcePR, Title: "Reviewed production hotfix conflict resolution", PatchID: id, Hotfix: true}
	attributeReviewedRevert(g, s, &ch)
	return []Change{ch}, nil
}
