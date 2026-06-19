package changelog

import (
	"context"
	"fmt"
)

func Generate(ctx context.Context, options Options) (Result, error) {
	normalized := normalizeOptions(options)
	if normalized.CurrentTag == "" {
		return Result{}, fmt.Errorf("current tag is required")
	}
	if normalized.RepoOwner == "" {
		return Result{}, fmt.Errorf("owner is required")
	}
	if normalized.RepoName == "" {
		return Result{}, fmt.Errorf("repo is required")
	}

	rangeEnd := normalized.CurrentTag
	if normalized.CurrentRef != "" {
		rangeEnd = normalized.CurrentRef
		if err := validateCommitish(ctx, normalized.WorkDir, normalized.CurrentRef); err != nil {
			return Result{}, err
		}
	} else {
		if err := validateTag(ctx, normalized.WorkDir, normalized.CurrentTag); err != nil {
			return Result{}, err
		}
	}
	if normalized.PrevTag != "" {
		if err := validateTag(ctx, normalized.WorkDir, normalized.PrevTag); err != nil {
			return Result{}, err
		}
	}

	rawCommits, err := loadCommits(ctx, normalized.WorkDir, normalized.PrevTag, rangeEnd)
	if err != nil {
		return Result{}, err
	}
	if len(rawCommits) == 0 {
		return Result{}, fmt.Errorf("no commits in range %s..%s", normalized.PrevTag, rangeEnd)
	}

	version, err := versionFromTag(normalized.CurrentTag)
	if err != nil {
		return Result{}, err
	}
	existing, mode, err := readExistingChangelog(normalized.OutputFile, version)
	if err != nil {
		return Result{}, err
	}

	releaseDate, err := commitishAuthorDate(ctx, normalized.WorkDir, rangeEnd)
	if err != nil {
		return Result{}, err
	}

	commits := parseCommits(rawCommits, normalized.WarningWriter)
	block := renderReleaseBlock(version, normalized.CurrentTag, releaseDate, normalized.RepoOwner, normalized.RepoName, commits)
	if err := prependRelease(normalized.OutputFile, block, existing, mode); err != nil {
		return Result{}, err
	}

	return Result{
		Path:          normalized.OutputFile,
		Tag:           normalized.CurrentTag,
		Version:       version,
		CommitCount:   len(commits),
		BreakingCount: breakingCount(commits),
	}, nil
}

func normalizeOptions(options Options) Options {
	if options.OutputFile == "" {
		options.OutputFile = DefaultOutputFile
	}
	return options
}

func validateTag(ctx context.Context, workDir string, tag string) error {
	exists, err := tagExists(ctx, workDir, tag)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("tag not found: %s", tag)
	}
	return nil
}

func validateCommitish(ctx context.Context, workDir string, ref string) error {
	exists, err := commitishExists(ctx, workDir, ref)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("ref not found: %s", ref)
	}
	return nil
}
