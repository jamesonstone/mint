package promotion

import (
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
			main, err := g.git(nil, "rev-parse", "origin/main")
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
		return []Change{{SHA: candidate.SourceSHA, Version: candidate.Version, PR: candidate.SourcePR, Title: "Isolated production hotfix", PatchID: id, Hotfix: true}}, nil
	}
	expected, err := PrepareHotfix(ctx, HotfixOptions{WorkDir: g.WorkDir, Baseline: *s.Baseline, Issue: 1, Fixes: meta.Fixes, CommitterName: AutomationLogin, CommitterEmail: AutomationEmail})
	if expected.WorkDir != "" {
		defer func() { _ = os.RemoveAll(expected.WorkDir) }()
	}
	if err != nil {
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
