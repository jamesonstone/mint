package release

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestVersionMainBatchReplayAndTagCollision(t *testing.T) {
	r := newTestRepo(t)
	r.commit(t, "fix: initial", "", "2026-01-01T00:00:00Z")
	r.tag(t, "v0.1.0")
	first := r.commit(t, "docs: ordinary patch fallback", "", "2026-01-02T00:00:00Z")
	r.tag(t, "v0.1.1")
	r.git(t, nil, "checkout", "-b", "other")
	other := r.commit(t, "fix: other branch", "", "2026-01-03T00:00:00Z")
	r.tag(t, "v0.2.0")
	r.git(t, nil, "checkout", "main")
	second := r.commit(t, "feat: queued feature", "", "2026-01-03T00:00:00Z")
	third := r.commit(t, "fix: small correction", "", "2026-01-04T00:00:00Z")
	values, err := VersionMain(context.Background(), MainVersionOptions{WorkDir: r.dir, MainRef: "main"})
	if err != nil || len(values) != 2 {
		t.Fatal(values, err)
	}
	if values[0].SHA != second || values[0].Version != "v0.2.1" || values[1].SHA != third || values[1].Version != "v0.2.2" {
		t.Fatal(values)
	}
	replay, err := VersionMain(context.Background(), MainVersionOptions{WorkDir: r.dir, MainRef: "main"})
	if err != nil || len(replay) != 1 || replay[0].Version != "v0.2.2" {
		t.Fatal(replay, err)
	}
	if r.revParse(t, "v0.1.1") != first || r.revParse(t, "v0.2.0^{commit}") != other {
		t.Fatal("tag moved")
	}
}
func TestVersionMainControlClassificationAndMixedChange(t *testing.T) {
	r := newTestRepo(t)
	r.commit(t, "fix: initial", "", "2026-01-01T00:00:00Z")
	r.tag(t, "v0.1.0")
	p := filepath.Join(r.dir, "CHANGELOG.md")
	if err := os.WriteFile(p, []byte("release pending"), 0600); err != nil {
		t.Fatal(err)
	}
	r.git(t, nil, "add", "CHANGELOG.md")
	r.git(t, nil, "commit", "-m", "chore: release documentation")
	v, err := VersionMain(context.Background(), MainVersionOptions{WorkDir: r.dir, MainRef: "main", ControlPaths: []string{"CHANGELOG.md"}})
	if err != nil || !v[0].ControlOnly {
		t.Fatal(v, err)
	}
	r.commit(t, "chore: mixed application changes", "", "2026-01-03T00:00:00Z")
	v, err = VersionMain(context.Background(), MainVersionOptions{WorkDir: r.dir, MainRef: "main", ControlPaths: []string{"CHANGELOG.md"}})
	if err != nil || v[0].ControlOnly {
		t.Fatal(v, err)
	}
}

func TestVersionMainMergedChangelogIsControlOnly(t *testing.T) {
	r := newTestRepo(t)
	r.commit(t, "fix: initial", "", "2026-01-01T00:00:00Z")
	r.tag(t, "v0.1.0")
	r.git(t, nil, "checkout", "-b", "release-notes")
	if err := os.WriteFile(filepath.Join(r.dir, "CHANGELOG.md"), []byte("pending"), 0600); err != nil {
		t.Fatal(err)
	}
	r.git(t, nil, "add", "CHANGELOG.md")
	r.git(t, nil, "commit", "-m", "chore: release notes")
	r.git(t, nil, "checkout", "main")
	r.git(t, nil, "merge", "--no-ff", "release-notes", "-m", "chore: merge release notes")
	versions, err := VersionMain(context.Background(), MainVersionOptions{WorkDir: r.dir, MainRef: "main", ControlPaths: []string{"CHANGELOG.md"}})
	if err != nil || len(versions) != 1 || !versions[0].ControlOnly {
		t.Fatal("merged release notes were classified as an application change", versions, err)
	}
}
