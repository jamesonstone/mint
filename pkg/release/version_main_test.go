package release

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestVersionMainBatchReplayAndHotfixCollision(t *testing.T) {
	r := newTestRepo(t)
	r.commit(t, "fix: initial", "", "2026-01-01T00:00:00Z")
	r.tag(t, "v0.1.0")
	first := r.commit(t, "docs: ordinary patch fallback", "", "2026-01-02T00:00:00Z")
	r.tag(t, "v0.1.1")
	second := r.commit(t, "feat: queued feature", "", "2026-01-03T00:00:00Z")
	third := r.commit(t, "fix: small correction", "", "2026-01-04T00:00:00Z")
	values, err := VersionMain(context.Background(), MainVersionOptions{WorkDir: r.dir, MainRef: "main"})
	if err != nil || len(values) != 2 {
		t.Fatal(values, err)
	}
	if values[0].SHA != second || values[0].Version != "v0.2.0" || values[1].SHA != third || values[1].Version != "v0.2.1" {
		t.Fatal(values)
	}
	replay, err := VersionMain(context.Background(), MainVersionOptions{WorkDir: r.dir, MainRef: "main"})
	if err != nil || len(replay) != 1 || replay[0].Version != "v0.2.1" {
		t.Fatal(replay, err)
	}
	if r.revParse(t, "v0.1.1") != first {
		t.Fatal("tag moved")
	}
	hotfix, err := AllocateHotfixVersion("v0.1.0", []string{"v0.1.1", "v0.1.2", "v0.2.0"})
	if err != nil || hotfix != "v0.1.3" {
		t.Fatal(hotfix, err)
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

func TestVersionMainMergedProposalDoesNotBecomeAnApplicationCandidate(t *testing.T) {
	r := newTestRepo(t)
	r.commit(t, "fix: initial", "", "2026-01-01T00:00:00Z")
	r.tag(t, "v0.1.0")
	r.git(t, nil, "checkout", "-b", "proposal")
	if err := os.WriteFile(filepath.Join(r.dir, "CHANGELOG.md"), []byte("pending"), 0600); err != nil {
		t.Fatal(err)
	}
	r.git(t, nil, "add", "CHANGELOG.md")
	r.git(t, nil, "commit", "-m", "chore: proposal")
	r.git(t, nil, "checkout", "main")
	r.git(t, nil, "merge", "--no-ff", "proposal", "-m", "chore: merge release proposal")
	versions, err := VersionMain(context.Background(), MainVersionOptions{WorkDir: r.dir, MainRef: "main", ControlPaths: []string{"CHANGELOG.md"}})
	if err != nil || len(versions) != 1 || !versions[0].ControlOnly {
		t.Fatal("merged control-only PR became another release candidate", versions, err)
	}
}
