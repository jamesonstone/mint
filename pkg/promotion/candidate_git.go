package promotion

import (
	"context"
	"fmt"
	"strings"
)

// CollectChanges rebuilds cumulative first-parent source provenance from Git and
// paginated GitHub PR associations. Hotfixes use their pinned production base.
func (c Client) CollectChanges(ctx context.Context, g GitProof, s State, candidate Candidate, controlPaths []string) ([]Change, error) {
	if candidate.Kind == "hotfix" {
		return c.collectHotfixChanges(ctx, g, s, candidate)
	}
	base := s.MainAnchor
	if candidate.Kind == "hotfix" && s.Baseline != nil {
		base = s.Baseline.Candidate.SourceSHA
	}
	if !shaPattern.MatchString(base) {
		return nil, fmt.Errorf("verified source anchor unavailable")
	}
	out, err := g.git(nil, "rev-list", "--first-parent", "--reverse", base+".."+candidate.SourceSHA)
	if err != nil {
		return nil, err
	}
	changes := []Change{}
	for _, sha := range strings.Fields(string(out)) {
		paths, err := g.git(nil, "diff-tree", "--root", "--no-commit-id", "--name-only", "-r", "--first-parent", sha)
		if err != nil {
			return nil, err
		}
		control := len(strings.Fields(string(paths))) > 0
		for _, path := range strings.Split(strings.TrimSpace(string(paths)), "\n") {
			allowed := false
			for _, entry := range controlPaths {
				if path == entry {
					allowed = true
				}
			}
			if !allowed {
				control = false
			}
		}
		if control {
			continue
		}
		tags, err := g.git(nil, "tag", "--points-at", sha)
		if err != nil {
			return nil, err
		}
		version := ""
		for _, tag := range strings.Fields(string(tags)) {
			if versionPattern.MatchString(tag) {
				if version != "" {
					return nil, fmt.Errorf("multiple source versions require explicit mapping for %s", sha)
				}
				version = tag
			}
		}
		// Hotfix preparations can contain several reviewed commits under one final tag.
		if version == "" && candidate.Kind == "hotfix" {
			version = candidate.Version
		}
		if version == "" {
			return nil, fmt.Errorf("source commit %s has no immutable version", sha)
		}
		title, err := g.git(nil, "show", "-s", "--format=%s", sha)
		if err != nil {
			return nil, err
		}
		patch, err := g.PatchID(sha)
		if err != nil {
			return nil, err
		}
		change := Change{SHA: sha, Version: version, Title: strings.TrimSpace(string(title)), PatchID: patch}
		pulls, err := c.MergedPRs(ctx, sha)
		if err != nil {
			return nil, err
		}
		for _, pr := range pulls {
			if pr.MergeSHA == sha {
				if change.PR != 0 {
					return nil, fmt.Errorf("ambiguous merged PR mapping")
				}
				change.PR = pr.Number
			}
		}
		reverse, reverseErr := g.git(nil, "diff", sha, sha+"^", "--", ".", ":(exclude).mint")
		if reverseErr == nil {
			ids, idErr := g.git(reverse, "patch-id", "--stable")
			if idErr == nil {
				parts := strings.Fields(string(ids))
				if len(parts) > 0 {
					for _, shipped := range s.Baseline.Shipped {
						if shipped.PatchID == parts[0] {
							change.Revert = true
							change.Reverts = shipped.PatchID
						}
					}
				}
			}
		}
		changes = append(changes, change)
	}
	return changes, nil
}

// VerifyTag rejects foreign/tag/artifact correspondence before journal writes.
func (g GitProof) VerifyTag(version, source string) error {
	if !versionPattern.MatchString(version) || !shaPattern.MatchString(source) {
		return fmt.Errorf("invalid version/source identity")
	}
	sha, err := g.git(nil, "rev-parse", "--verify", version+"^{commit}")
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(sha)) != source {
		return fmt.Errorf("version tag does not identify the selected source")
	}
	return nil
}
