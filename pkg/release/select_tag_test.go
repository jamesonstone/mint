package release

import (
	"context"
	"strings"
	"testing"
)

func TestSelectTagUsesRequestedTag(t *testing.T) {
	repo := newTestRepo(t)
	target := repo.commit(t, "feat: deployable release", "", "2024-01-01T00:00:00Z")

	result, err := SelectTag(context.Background(), SelectTagOptions{
		Commitish:    "HEAD",
		RequestedTag: "v1.2.3",
		WorkDir:      repo.dir,
	})
	if err != nil {
		t.Fatalf("SelectTag() error = %v", err)
	}

	if result.VersionTag != "v1.2.3" {
		t.Fatalf("VersionTag = %q, want v1.2.3", result.VersionTag)
	}
	if result.TagSource != SelectTagSourceRequested {
		t.Fatalf("TagSource = %q, want %q", result.TagSource, SelectTagSourceRequested)
	}
	if result.TargetSHA != target {
		t.Fatalf("TargetSHA = %q, want %q", result.TargetSHA, target)
	}
	if len(result.ShortSHA) != 12 {
		t.Fatalf("ShortSHA length = %d, want 12", len(result.ShortSHA))
	}
}

func TestSelectTagRejectsInvalidRequestedTag(t *testing.T) {
	repo := newTestRepo(t)
	repo.commit(t, "feat: deployable release", "", "2024-01-01T00:00:00Z")

	_, err := SelectTag(context.Background(), SelectTagOptions{
		RequestedTag: "v1.2.3-alpha",
		WorkDir:      repo.dir,
	})
	if err == nil || err.Error() != "requested-tag: must be a strict vX.Y.Z SemVer tag" {
		t.Fatalf("SelectTag() error = %v", err)
	}
}

func TestSelectTagChoosesHighestSemVerTagOnTarget(t *testing.T) {
	repo := newTestRepo(t)
	target := repo.commit(t, "feat: tagged release", "", "2024-01-01T00:00:00Z")
	repo.tag(t, "v1.0.0")
	repo.tag(t, "v1.2.0")
	repo.tag(t, "v1.2.0-alpha")

	result, err := SelectTag(context.Background(), SelectTagOptions{WorkDir: repo.dir})
	if err != nil {
		t.Fatalf("SelectTag() error = %v", err)
	}

	if result.VersionTag != "v1.2.0" {
		t.Fatalf("VersionTag = %q, want v1.2.0", result.VersionTag)
	}
	if result.TagSource != SelectTagSourceCommitTag {
		t.Fatalf("TagSource = %q, want %q", result.TagSource, SelectTagSourceCommitTag)
	}
	if result.TargetSHA != target {
		t.Fatalf("TargetSHA = %q, want %q", result.TargetSHA, target)
	}
}

func TestSelectTagFailsWhenTargetHasNoSemVerTag(t *testing.T) {
	repo := newTestRepo(t)
	repo.commit(t, "feat: untagged release", "", "2024-01-01T00:00:00Z")
	repo.tag(t, "v1.0.0-alpha")

	_, err := SelectTag(context.Background(), SelectTagOptions{WorkDir: repo.dir})
	if err == nil || !strings.Contains(err.Error(), "no strict SemVer tag points at HEAD") {
		t.Fatalf("SelectTag() error = %v", err)
	}
}

func TestSelectTagFailsForInvalidCommitish(t *testing.T) {
	repo := newTestRepo(t)
	repo.commit(t, "feat: deployable release", "", "2024-01-01T00:00:00Z")

	_, err := SelectTag(context.Background(), SelectTagOptions{
		Commitish: "does-not-exist",
		WorkDir:   repo.dir,
	})
	if err == nil || err.Error() != "ref not found: does-not-exist" {
		t.Fatalf("SelectTag() error = %v", err)
	}
}
