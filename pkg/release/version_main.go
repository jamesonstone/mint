package release

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// MainVersionOptions versions first-parent commits after the latest tagged main ancestor.
// Callers must serialize this operation per repository and fetch immutable tags first.
type MainVersionOptions struct {
	WorkDir, MainRef, Remote string
	Push                     bool
	ControlPaths             []string
}

// MainVersion records immutable source identity independently of artifact eligibility.
type MainVersion struct {
	SHA         string `json:"source_sha"`
	Version     string `json:"version"`
	ControlOnly bool   `json:"control_only"`
}

// VersionMain allocates one immutable SemVer per first-parent main commit. It does
// not create GitHub Releases. Out-of-order event handlers must use current main.
func VersionMain(ctx context.Context, opts MainVersionOptions) ([]MainVersion, error) {
	target, err := resolveCommit(ctx, opts.WorkDir, opts.MainRef)
	if err != nil {
		return nil, err
	}
	history, err := runGit(ctx, opts.WorkDir, "rev-list", "--first-parent", target)
	if err != nil {
		return nil, err
	}
	pending := []string{}
	var base semverTag
	hasBase := false
	for _, sha := range splitLines(history) {
		tags, err := semVerTagsPointingAt(ctx, opts.WorkDir, sha)
		if err != nil {
			return nil, err
		}
		if tag, found := highestSemVer(tags); found {
			base, hasBase = tag, true
			break
		}
		pending = append(pending, sha)
	}
	// Every occupied tag, including divergent hotfixes, fences new allocation.
	names, err := runGit(ctx, opts.WorkDir, "tag", "--list")
	if err != nil {
		return nil, err
	}
	occupied := map[string]bool{}
	for _, name := range splitLines(names) {
		occupied[name] = true
	}
	results := []MainVersion{}
	for i := len(pending) - 1; i >= 0; i-- {
		sha := pending[i]
		commits, err := loadCommits(ctx, opts.WorkDir, sha+"^", sha)
		if err != nil {
			// Only a true root commit has no parent; do not swallow other Git failures.
			parents, parentErr := runGit(ctx, opts.WorkDir, "rev-list", "--parents", "-n", "1", sha)
			if parentErr != nil || len(strings.Fields(string(parents))) != 1 {
				return nil, err
			}
			commits, err = loadCommits(ctx, opts.WorkDir, "", sha)
		}
		if err != nil {
			return nil, err
		}
		_, _, rank := evaluateCommits(commits)
		if rank < bumpRankPatch {
			rank = bumpRankPatch
		}
		version := nextVersion(base, hasBase, rank)
		next, _ := parseSemVerTag(version)
		for occupied[version] {
			next.Patch++
			version = formatSemVer(next.Major, next.Minor, next.Patch)
		}
		notes, err := os.CreateTemp("", "mint-version-*.md")
		if err != nil {
			return nil, err
		}
		_, writeErr := fmt.Fprintf(notes, "Source version %s for %s. Artifact and production status are independent.\n", version, sha)
		closeErr := notes.Close()
		if writeErr != nil || closeErr != nil {
			_ = os.Remove(notes.Name())
			return nil, fmt.Errorf("write version notes: %v %v", writeErr, closeErr)
		}
		_, err = CreateReleaseTag(ctx, TagOptions{Tag: version, Target: sha, NotesFile: notes.Name(), Remote: opts.Remote, Push: opts.Push, WorkDir: opts.WorkDir})
		_ = os.Remove(notes.Name())
		if err != nil {
			return nil, err
		}
		control, err := controlOnlyCommit(ctx, opts.WorkDir, sha, opts.ControlPaths)
		if err != nil {
			return nil, err
		}
		results = append(results, MainVersion{SHA: sha, Version: version, ControlOnly: control})
		base, hasBase, occupied[version] = next, true, true
	}
	if len(results) == 0 {
		control, err := controlOnlyCommit(ctx, opts.WorkDir, target, opts.ControlPaths)
		if err != nil {
			return nil, err
		}
		results = append(results, MainVersion{SHA: target, Version: base.Name, ControlOnly: control})
	}
	return results, nil
}

func controlOnlyCommit(ctx context.Context, dir, sha string, allowed []string) (bool, error) {
	out, err := runGit(ctx, dir, "diff-tree", "-m", "--root", "--no-commit-id", "--name-only", "-r", "--first-parent", sha)
	if err != nil {
		return false, err
	}
	paths := splitLines(out)
	if len(paths) == 0 || len(allowed) == 0 {
		return false, nil
	}
	for _, path := range paths {
		found := false
		for _, entry := range allowed {
			if path == entry {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}
	return true, nil
}

// AllocateHotfixVersion chooses the next unused patch in production's series;
// queued minor versions do not change that compatibility series.
func AllocateHotfixVersion(production string, occupied []string) (string, error) {
	base, ok := parseSemVerTag(production)
	if !ok {
		return "", fmt.Errorf("production version must be strict SemVer")
	}
	used := map[string]bool{}
	for _, tag := range occupied {
		used[tag] = true
	}
	for {
		base.Patch++
		next := formatSemVer(base.Major, base.Minor, base.Patch)
		if !used[next] {
			return next, nil
		}
	}
}

// VersionIdentity reads an already minted source version and its control classification.
func VersionIdentity(ctx context.Context, dir, commitish string, paths []string) (MainVersion, error) {
	selected, err := SelectTag(ctx, SelectTagOptions{Commitish: commitish, WorkDir: dir})
	if err != nil {
		return MainVersion{}, err
	}
	control, err := controlOnlyCommit(ctx, dir, selected.TargetSHA, paths)
	if err != nil {
		return MainVersion{}, err
	}
	return MainVersion{SHA: selected.TargetSHA, Version: selected.VersionTag, ControlOnly: control}, nil
}
