package release

import (
	"context"
	"strings"
)

// SelectTagOptions configures read-only release tag selection.
type SelectTagOptions struct {
	Commitish    string
	RequestedTag string
	WorkDir      string
}

// SelectTag chooses an existing strict SemVer tag for a commit, or validates a
// caller-provided tag. It never computes a new release version.
func SelectTag(ctx context.Context, opts SelectTagOptions) (SelectTagResult, error) {
	commitish := opts.Commitish
	if commitish == "" {
		commitish = DefaultCommitish
	}

	target, err := resolveCommit(ctx, opts.WorkDir, commitish)
	if err != nil {
		return SelectTagResult{}, err
	}

	short, err := shortCommit(ctx, opts.WorkDir, target)
	if err != nil {
		return SelectTagResult{}, err
	}

	requested := strings.TrimSpace(opts.RequestedTag)
	if requested != "" {
		if _, ok := parseSemVerTag(requested); !ok {
			return SelectTagResult{}, validationError("requested-tag", "must be a strict vX.Y.Z SemVer tag")
		}
		return SelectTagResult{
			VersionTag: requested,
			TagSource:  SelectTagSourceRequested,
			TargetSHA:  target,
			ShortSHA:   short,
		}, nil
	}

	tags, err := semVerTagsPointingAt(ctx, opts.WorkDir, target)
	if err != nil {
		return SelectTagResult{}, err
	}
	tag, ok := highestSemVer(tags)
	if !ok {
		return SelectTagResult{}, validationError("release tag", "no strict SemVer tag points at %s; run release publishing before deployment or pass --requested-tag", commitish)
	}

	return SelectTagResult{
		VersionTag: tag.Name,
		TagSource:  SelectTagSourceCommitTag,
		TargetSHA:  target,
		ShortSHA:   short,
	}, nil
}
