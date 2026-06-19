package release

import (
	"context"
)

// Resolve computes release metadata from reachable Git history without
// mutating repository state.
func Resolve(ctx context.Context, opts Options) (Result, error) {
	commitish := opts.Commitish
	if commitish == "" {
		commitish = DefaultCommitish
	}

	target, err := resolveCommit(ctx, opts.WorkDir, commitish)
	if err != nil {
		return Result{}, err
	}

	short, err := shortCommit(ctx, opts.WorkDir, target)
	if err != nil {
		return Result{}, err
	}

	targetTags, err := semVerTagsPointingAt(ctx, opts.WorkDir, target)
	if err != nil {
		return Result{}, err
	}
	if tag, ok := highestSemVer(targetTags); ok {
		result := Result{
			VersionTag:  tag.Name,
			VersionBump: BumpAlreadyTagged,
			BaseTag:     tag.Name,
			TargetSHA:   target,
			ShortSHA:    short,
			NeedsGitTag: false,
			CommitCount: 0,
		}
		result.ReleaseNotes = renderReleaseNotes(result, nil)
		return result, nil
	}

	base, hasBase, err := resolveBaseTag(ctx, opts.WorkDir, target)
	if err != nil {
		return Result{}, err
	}

	baseTag := ""
	if hasBase {
		baseTag = base.Name
	}

	commits, err := loadCommits(ctx, opts.WorkDir, baseTag, target)
	if err != nil {
		return Result{}, err
	}

	if len(commits) == 0 {
		versionTag := "v0.0.0"
		if hasBase {
			versionTag = base.Name
		}
		result := Result{
			VersionTag:  versionTag,
			VersionBump: BumpAlreadyTagged,
			BaseTag:     baseTag,
			TargetSHA:   target,
			ShortSHA:    short,
			NeedsGitTag: false,
			CommitCount: 0,
		}
		result.ReleaseNotes = renderReleaseNotes(result, nil)
		return result, nil
	}

	evaluations, bump, rank := evaluateCommits(commits)
	versionTag := nextVersion(base, hasBase, rank)
	result := Result{
		VersionTag:   versionTag,
		VersionBump:  bump,
		BaseTag:      baseTag,
		TargetSHA:    target,
		ShortSHA:     short,
		NeedsGitTag:  true,
		CommitCount:  len(commits),
		ReleaseNotes: "",
	}
	result.ReleaseNotes = renderReleaseNotes(result, evaluations)
	return result, nil
}

func resolveBaseTag(ctx context.Context, workDir string, target string) (semverTag, bool, error) {
	tags, err := reachableSemVerTags(ctx, workDir, target)
	if err != nil {
		return semverTag{}, false, err
	}
	tag, ok := highestSemVer(tags)
	return tag, ok, nil
}

func nextVersion(base semverTag, hasBase bool, rank int) string {
	major := 0
	minor := 0
	patch := 0
	if hasBase {
		major = base.Major
		minor = base.Minor
		patch = base.Patch
	}

	switch rank {
	case bumpRankMajor:
		if !hasBase {
			return "v1.0.0"
		}
		return formatSemVer(major+1, 0, 0)
	case bumpRankMinor:
		return formatSemVer(major, minor+1, 0)
	case bumpRankPatch:
		return formatSemVer(major, minor, patch+1)
	default:
		if hasBase {
			return base.Name
		}
		return "v0.0.0"
	}
}
