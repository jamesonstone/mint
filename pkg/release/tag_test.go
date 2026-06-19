package release

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateReleaseTagCreatesAnnotatedTagWithoutPush(t *testing.T) {
	repo := newTestRepo(t)
	target := repo.commit(t, "feat: release", "", "2024-01-01T00:00:00Z")
	notesFile := writeReleaseNotesFixture(t, repo.dir)

	result, err := CreateReleaseTag(context.Background(), TagOptions{
		Tag:       "v1.2.3",
		Target:    "HEAD",
		NotesFile: notesFile,
		Push:      false,
		WorkDir:   repo.dir,
	})
	if err != nil {
		t.Fatalf("CreateReleaseTag() error = %v", err)
	}

	if !result.Created || result.Reused || result.Pushed {
		t.Fatalf("result created/reused/pushed = %t/%t/%t", result.Created, result.Reused, result.Pushed)
	}
	if result.TagName != "v1.2.3" || result.TargetSHA != target {
		t.Fatalf("result tag/target = %q/%q, want v1.2.3/%q", result.TagName, result.TargetSHA, target)
	}
	if got := strings.TrimSpace(repo.gitOutput(t, nil, "cat-file", "-t", "v1.2.3")); got != "tag" {
		t.Fatalf("tag object type = %q, want tag", got)
	}
	if got := repo.revParse(t, "v1.2.3^{commit}"); got != target {
		t.Fatalf("tag target = %q, want %q", got, target)
	}
}

func TestCreateReleaseTagReusesSameCommitTag(t *testing.T) {
	repo := newTestRepo(t)
	target := repo.commit(t, "feat: release", "", "2024-01-01T00:00:00Z")
	notesFile := writeReleaseNotesFixture(t, repo.dir)
	repo.git(t, nil, "tag", "-a", "v1.2.3", target, "-F", notesFile)

	result, err := CreateReleaseTag(context.Background(), TagOptions{
		Tag:       "v1.2.3",
		Target:    target,
		NotesFile: notesFile,
		Push:      false,
		WorkDir:   repo.dir,
	})
	if err != nil {
		t.Fatalf("CreateReleaseTag() error = %v", err)
	}

	if result.Created || !result.Reused || result.Pushed {
		t.Fatalf("result created/reused/pushed = %t/%t/%t", result.Created, result.Reused, result.Pushed)
	}
	if result.TargetSHA != target {
		t.Fatalf("TargetSHA = %q, want %q", result.TargetSHA, target)
	}
}

func TestCreateReleaseTagFailsForConflictingTagAndDoesNotMoveIt(t *testing.T) {
	repo := newTestRepo(t)
	original := repo.commit(t, "feat: first", "", "2024-01-01T00:00:00Z")
	notesFile := writeReleaseNotesFixture(t, repo.dir)
	repo.git(t, nil, "tag", "-a", "v1.2.3", original, "-F", notesFile)
	target := repo.commit(t, "feat: second", "", "2024-01-02T00:00:00Z")

	_, err := CreateReleaseTag(context.Background(), TagOptions{
		Tag:       "v1.2.3",
		Target:    target,
		NotesFile: notesFile,
		Push:      false,
		WorkDir:   repo.dir,
	})
	if err == nil {
		t.Fatalf("CreateReleaseTag() error = nil, want conflict")
	}
	for _, want := range []string{
		"tag v1.2.3 already exists",
		"recovery: inspect and correct the conflicting tag",
		"push a dummy commit",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error missing %q:\n%s", want, err.Error())
		}
	}
	if got := repo.revParse(t, "v1.2.3^{commit}"); got != original {
		t.Fatalf("tag moved to %q, want original %q", got, original)
	}
}

func TestCreateReleaseTagValidatesInputs(t *testing.T) {
	repo := newTestRepo(t)
	repo.commit(t, "feat: release", "", "2024-01-01T00:00:00Z")
	notesFile := writeReleaseNotesFixture(t, repo.dir)

	tests := []struct {
		name string
		opts TagOptions
		want string
	}{
		{name: "missing tag", opts: TagOptions{Target: "HEAD", NotesFile: notesFile, WorkDir: repo.dir}, want: "strict vX.Y.Z"},
		{name: "invalid tag", opts: TagOptions{Tag: "1.2.3", Target: "HEAD", NotesFile: notesFile, WorkDir: repo.dir}, want: "strict vX.Y.Z"},
		{name: "missing target", opts: TagOptions{Tag: "v1.2.3", NotesFile: notesFile, WorkDir: repo.dir}, want: "target"},
		{name: "missing notes", opts: TagOptions{Tag: "v1.2.3", Target: "HEAD", WorkDir: repo.dir}, want: "notes-file"},
		{name: "unknown target", opts: TagOptions{Tag: "v1.2.3", Target: "missing-ref", NotesFile: notesFile, WorkDir: repo.dir}, want: "ref not found: missing-ref"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateReleaseTag(context.Background(), tt.opts)
			if err == nil {
				t.Fatalf("CreateReleaseTag() error = nil, want %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("CreateReleaseTag() error = %q, want contains %q", err.Error(), tt.want)
			}
		})
	}
}

func writeReleaseNotesFixture(t *testing.T, dir string) string {
	t.Helper()

	path := filepath.Join(dir, "notes.md")
	if err := os.WriteFile(path, []byte("Release notes\n"), 0o644); err != nil {
		t.Fatalf("write notes fixture: %v", err)
	}
	return path
}
