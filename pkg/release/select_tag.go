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
	// Scope limits selection to one component's tags.
	Scope Scope
}

// SelectTag chooses an existing strict SemVer tag for a commit, or validates a
// caller-provided tag. It never computes a new release version.
func SelectTag(ctx context.Context, opts SelectTagOptions) (SelectTagResult, error) {
	if err := opts.Scope.Validate(); err != nil {
		return SelectTagResult{}, err
	}
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
		if _, ok := opts.Scope.parseTag(requested); !ok {
			return SelectTagResult{}, validationError("requested-tag", "%s", opts.Scope.tagRule())
		}
		return SelectTagResult{
			VersionTag: requested,
			TagSource:  SelectTagSourceRequested,
			TargetSHA:  target,
			ShortSHA:   short,
		}, nil
	}

	tags, err := semVerTagsPointingAt(ctx, opts.WorkDir, target, opts.Scope)
	if err != nil {
		return SelectTagResult{}, err
	}
	tag, ok := highestSemVer(tags)
	if !ok {
		kind := "SemVer"
		if opts.Scope.TagPrefix != "" {
			kind = opts.Scope.TagPrefix + "vX.Y.Z SemVer"
		}
		return SelectTagResult{}, validationError("release tag", "no strict %s tag points at %s; run release publishing before deployment or pass --requested-tag", kind, commitish)
	}

	return SelectTagResult{
		VersionTag: tag.Name,
		TagSource:  SelectTagSourceCommitTag,
		TargetSHA:  target,
		ShortSHA:   short,
	}, nil
}
