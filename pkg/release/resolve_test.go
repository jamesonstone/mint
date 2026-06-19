package release

import (
	"context"
	"strings"
	"testing"
)

func TestResolveFirstReleaseBumps(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
		bump Bump
	}{
		{name: "feat: first feature", want: "v0.1.0", bump: BumpMinor},
		{name: "fix: first fix", want: "v0.0.1", bump: BumpPatch},
		{name: "Update first file", want: "v0.0.1", bump: BumpPatch},
		{name: "feat!: first breaking feature", want: "v1.0.0", bump: BumpMajor},
		{name: "fix: body breaking", body: "BREAKING CHANGE: contract changed", want: "v1.0.0", bump: BumpMajor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newTestRepo(t)
			target := repo.commit(t, tt.name, tt.body, "2024-01-01T00:00:00Z")

			result, err := Resolve(context.Background(), Options{WorkDir: repo.dir})
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}

			if result.VersionTag != tt.want {
				t.Fatalf("VersionTag = %q, want %q", result.VersionTag, tt.want)
			}
			if result.VersionBump != tt.bump {
				t.Fatalf("VersionBump = %q, want %q", result.VersionBump, tt.bump)
			}
			if result.BaseTag != "" {
				t.Fatalf("BaseTag = %q, want empty", result.BaseTag)
			}
			if result.TargetSHA != target {
				t.Fatalf("TargetSHA = %q, want %q", result.TargetSHA, target)
			}
			if len(result.ShortSHA) != 12 {
				t.Fatalf("ShortSHA length = %d, want 12", len(result.ShortSHA))
			}
			if !result.NeedsGitTag {
				t.Fatalf("NeedsGitTag = false, want true")
			}
			if result.CommitCount != 1 {
				t.Fatalf("CommitCount = %d, want 1", result.CommitCount)
			}
			for _, want := range []string{tt.want, target, tt.name} {
				if !strings.Contains(result.ReleaseNotes, want) {
					t.Fatalf("ReleaseNotes missing %q:\n%s", want, result.ReleaseNotes)
				}
			}
		})
	}
}

func TestResolveReachableBaseBumps(t *testing.T) {
	tests := []struct {
		subject string
		body    string
		want    string
		bump    Bump
	}{
		{subject: "fix: repair release", want: "v1.0.1", bump: BumpPatch},
		{subject: "feat: add release command", want: "v1.1.0", bump: BumpMinor},
		{subject: "refactor!: replace release API", want: "v2.0.0", bump: BumpMajor},
		{subject: "fix: body breaking", body: "BREAKING-CHANGE: output changed", want: "v2.0.0", bump: BumpMajor},
	}

	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			repo := newTestRepo(t)
			repo.commit(t, "feat: existing release", "", "2024-01-01T00:00:00Z")
			repo.tag(t, "v1.0.0")
			repo.commit(t, tt.subject, tt.body, "2024-01-02T00:00:00Z")

			result, err := Resolve(context.Background(), Options{WorkDir: repo.dir})
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}

			if result.VersionTag != tt.want {
				t.Fatalf("VersionTag = %q, want %q", result.VersionTag, tt.want)
			}
			if result.VersionBump != tt.bump {
				t.Fatalf("VersionBump = %q, want %q", result.VersionBump, tt.bump)
			}
			if result.BaseTag != "v1.0.0" {
				t.Fatalf("BaseTag = %q, want v1.0.0", result.BaseTag)
			}
			if !result.NeedsGitTag {
				t.Fatalf("NeedsGitTag = false, want true")
			}
			if result.CommitCount != 1 {
				t.Fatalf("CommitCount = %d, want 1", result.CommitCount)
			}
		})
	}
}

func TestResolveIgnoresUnreachableHigherTag(t *testing.T) {
	repo := newTestRepo(t)
	repo.commit(t, "feat: existing release", "", "2024-01-01T00:00:00Z")
	repo.tag(t, "v1.0.0")
	main := repo.revParse(t, "HEAD")

	repo.checkoutNewBranch(t, "unrelated")
	repo.commit(t, "feat: unrelated future", "", "2024-01-02T00:00:00Z")
	repo.tag(t, "v9.0.0")

	repo.checkout(t, "main")
	repo.commit(t, "fix: target branch", "", "2024-01-03T00:00:00Z")

	result, err := Resolve(context.Background(), Options{WorkDir: repo.dir})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.BaseTag != "v1.0.0" {
		t.Fatalf("BaseTag = %q, want v1.0.0", result.BaseTag)
	}
	if result.VersionTag != "v1.0.1" {
		t.Fatalf("VersionTag = %q, want v1.0.1", result.VersionTag)
	}
	if tagCommit := repo.revParse(t, "v1.0.0"); tagCommit != main {
		t.Fatalf("test setup tag commit = %q, want %q", tagCommit, main)
	}
}

func TestResolveAlreadyTaggedTargetChoosesHighestTag(t *testing.T) {
	repo := newTestRepo(t)
	target := repo.commit(t, "feat: tagged release", "", "2024-01-01T00:00:00Z")
	repo.tag(t, "v1.0.0")
	repo.tag(t, "v1.1.0")
	repo.tag(t, "v1.1.0-alpha")

	result, err := Resolve(context.Background(), Options{WorkDir: repo.dir})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.VersionTag != "v1.1.0" {
		t.Fatalf("VersionTag = %q, want v1.1.0", result.VersionTag)
	}
	if result.VersionBump != BumpAlreadyTagged {
		t.Fatalf("VersionBump = %q, want %q", result.VersionBump, BumpAlreadyTagged)
	}
	if result.BaseTag != "v1.1.0" {
		t.Fatalf("BaseTag = %q, want v1.1.0", result.BaseTag)
	}
	if result.TargetSHA != target {
		t.Fatalf("TargetSHA = %q, want %q", result.TargetSHA, target)
	}
	if result.NeedsGitTag {
		t.Fatalf("NeedsGitTag = true, want false")
	}
	if result.CommitCount != 0 {
		t.Fatalf("CommitCount = %d, want 0", result.CommitCount)
	}
}

func TestResolveIgnoresNonStrictSemVerTags(t *testing.T) {
	repo := newTestRepo(t)
	repo.commit(t, "chore: existing pre-release", "", "2024-01-01T00:00:00Z")
	repo.tag(t, "v1.0.0-alpha")
	repo.tag(t, "1.0.0")
	repo.commit(t, "fix: target", "", "2024-01-02T00:00:00Z")

	result, err := Resolve(context.Background(), Options{WorkDir: repo.dir})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if result.BaseTag != "" {
		t.Fatalf("BaseTag = %q, want empty", result.BaseTag)
	}
	if result.VersionTag != "v0.0.1" {
		t.Fatalf("VersionTag = %q, want v0.0.1", result.VersionTag)
	}
	if result.CommitCount != 2 {
		t.Fatalf("CommitCount = %d, want 2", result.CommitCount)
	}
}

func TestResolveFailsForInvalidCommitish(t *testing.T) {
	repo := newTestRepo(t)
	repo.commit(t, "feat: existing", "", "2024-01-01T00:00:00Z")

	_, err := Resolve(context.Background(), Options{
		Commitish: "does-not-exist",
		WorkDir:   repo.dir,
	})
	if err == nil || err.Error() != "ref not found: does-not-exist" {
		t.Fatalf("Resolve() error = %v", err)
	}
}
