package release

import (
	"context"
	"errors"
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
	// Scope versions one component: only first-parent commits touching its
	// paths get <prefix>vX.Y.Z tags. The zero value versions every commit.
	Scope Scope
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
	if err := opts.Scope.Validate(); err != nil {
		return nil, err
	}
	target, err := resolveCommit(ctx, opts.WorkDir, opts.MainRef)
	if err != nil {
		return nil, err
	}
	history, err := runGit(ctx, opts.WorkDir, append([]string{"rev-list", "--first-parent", target}, opts.Scope.pathspecArgs()...)...)
	if err != nil {
		return nil, err
	}
	pending := []string{}
	var base semverTag
	hasBase := false
	for _, sha := range splitLines(history) {
		tags, err := semVerTagsPointingAt(ctx, opts.WorkDir, sha, opts.Scope)
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
		commits, err := loadCommits(ctx, opts.WorkDir, sha+"^", sha, opts.Scope)
		if err != nil {
			// Only a true root commit has no parent; do not swallow other Git failures.
			parents, parentErr := runGit(ctx, opts.WorkDir, "rev-list", "--parents", "-n", "1", sha)
			if parentErr != nil || len(strings.Fields(string(parents))) != 1 {
				return nil, err
			}
			commits, err = loadCommits(ctx, opts.WorkDir, "", sha, opts.Scope)
		}
		if err != nil {
			return nil, err
		}
		_, _, rank := evaluateCommits(commits)
		if rank < bumpRankPatch {
			rank = bumpRankPatch
		}
		version := nextVersion(base, hasBase, rank, opts.Scope)
		next, _ := opts.Scope.parseTag(version)
		for occupied[version] {
			next.Patch++
			version = opts.Scope.FormatTag(next.Major, next.Minor, next.Patch)
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
		_, err = CreateReleaseTag(ctx, TagOptions{Tag: version, Target: sha, NotesFile: notes.Name(), Remote: opts.Remote, Push: opts.Push, WorkDir: opts.WorkDir, Scope: opts.Scope})
		_ = os.Remove(notes.Name())
		if err != nil {
			return nil, err
		}
		control, err := controlOnlyCommit(ctx, opts.WorkDir, sha, opts.ControlPaths)
		if err != nil {
			return nil, err
		}
		results = append(results, MainVersion{SHA: sha, Version: version, ControlOnly: control})
		next.Name, next.Commit = version, sha
		base, hasBase, occupied[version] = next, true, true
	}
	if len(results) == 0 {
		// Nothing new to version: report the newest versioned source. For a
		// component that is the last commit touching its paths, which can be
		// older than the target; with no versioned source there is nothing.
		if !hasBase {
			return results, nil
		}
		control, err := controlOnlyCommit(ctx, opts.WorkDir, base.Commit, opts.ControlPaths)
		if err != nil {
			return nil, err
		}
		results = append(results, MainVersion{SHA: base.Commit, Version: base.Name, ControlOnly: control})
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
func AllocateHotfixVersion(production string, occupied []string, scope Scope) (string, error) {
	base, ok := scope.parseTag(production)
	if !ok {
		return "", fmt.Errorf("production version must be strict %sSemVer", scope.TagPrefix)
	}
	used := map[string]bool{}
	for _, tag := range occupied {
		used[tag] = true
	}
	for {
		base.Patch++
		next := scope.FormatTag(base.Major, base.Minor, base.Patch)
		if !used[next] {
			return next, nil
		}
	}
}

// VersionIdentity reads an already minted source version and its control classification.
func VersionIdentity(ctx context.Context, dir, commitish string, paths []string, scope Scope) (MainVersion, error) {
	selected, err := SelectTag(ctx, SelectTagOptions{Commitish: commitish, WorkDir: dir, Scope: scope})
	if err != nil {
		return MainVersion{}, err
	}
	control, err := controlOnlyCommit(ctx, dir, selected.TargetSHA, paths)
	if err != nil {
		return MainVersion{}, err
	}
	return MainVersion{SHA: selected.TargetSHA, Version: selected.VersionTag, ControlOnly: control}, nil
}

// ComponentVersionIdentity reports a component's version at commitish. When no
// component tag points at commitish (the commit did not touch the component's
// paths), changed is false and the result has an empty Version, so build
// workflows can skip instead of failing.
func ComponentVersionIdentity(ctx context.Context, dir, commitish string, controlPaths []string, scope Scope) (MainVersion, bool, error) {
	identity, err := VersionIdentity(ctx, dir, commitish, controlPaths, scope)
	if err == nil {
		return identity, true, nil
	}
	var issue *validationIssue
	if !errors.As(err, &issue) || issue.Field != "release tag" {
		return MainVersion{}, false, err
	}
	target, err := resolveCommit(ctx, dir, commitish)
	if err != nil {
		return MainVersion{}, false, err
	}
	control, err := controlOnlyCommit(ctx, dir, target, controlPaths)
	if err != nil {
		return MainVersion{}, false, err
	}
	return MainVersion{SHA: target, ControlOnly: control}, false, nil
}
