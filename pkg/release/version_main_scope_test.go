package release

import (
	"context"
	"testing"
)

func TestVersionMainScopedVersionsOnlyComponentCommits(t *testing.T) {
	r := newTestRepo(t)
	backendFirst := r.commitAt(t, "backend/main.go", "feat: api", "2026-01-01T00:00:00Z")
	r.commitAt(t, "frontend/app.ts", "feat!: redesign", "2026-01-02T00:00:00Z")
	backendFix := r.commitAt(t, "backend/fix.go", "fix: nil check", "2026-01-03T00:00:00Z")
	frontendLast := r.commitAt(t, "frontend/style.css", "fix: color", "2026-01-04T00:00:00Z")

	values, err := VersionMain(context.Background(), MainVersionOptions{WorkDir: r.dir, MainRef: "main", Scope: backendScope})
	if err != nil || len(values) != 2 {
		t.Fatalf("VersionMain() = %v, %v", values, err)
	}
	if values[0].SHA != backendFirst || values[0].Version != "backend/v0.1.0" || values[1].SHA != backendFix || values[1].Version != "backend/v0.1.1" {
		t.Fatalf("VersionMain() = %+v", values)
	}
	if r.revParse(t, "backend/v0.1.1^{commit}") != backendFix {
		t.Fatal("backend tag not on the backend commit")
	}

	replay, err := VersionMain(context.Background(), MainVersionOptions{WorkDir: r.dir, MainRef: "main", Scope: backendScope})
	if err != nil || len(replay) != 1 || replay[0].Version != "backend/v0.1.1" || replay[0].SHA != backendFix {
		t.Fatalf("replay must report the newest backend version at its own commit: %+v, %v", replay, err)
	}

	frontend, err := VersionMain(context.Background(), MainVersionOptions{WorkDir: r.dir, MainRef: "main", Scope: frontendScope})
	if err != nil || len(frontend) != 2 || frontend[0].Version != "frontend/v1.0.0" || frontend[1].Version != "frontend/v1.0.1" || frontend[1].SHA != frontendLast {
		t.Fatalf("frontend VersionMain() = %+v, %v", frontend, err)
	}

	whole, err := VersionMain(context.Background(), MainVersionOptions{WorkDir: r.dir, MainRef: "main"})
	if err != nil || len(whole) != 4 || whole[3].Version != "v1.0.2" {
		t.Fatalf("component tags must not change repository-wide versioning: %+v, %v", whole, err)
	}
}

func TestVersionMainScopedWithoutComponentCommits(t *testing.T) {
	r := newTestRepo(t)
	r.commitAt(t, "frontend/app.ts", "feat: ui", "2026-01-01T00:00:00Z")
	values, err := VersionMain(context.Background(), MainVersionOptions{WorkDir: r.dir, MainRef: "main", Scope: backendScope})
	if err != nil || len(values) != 0 {
		t.Fatalf("VersionMain() = %v, %v; want no versions", values, err)
	}
}

func TestVersionIdentityAndHotfixScoped(t *testing.T) {
	r := newTestRepo(t)
	sha := r.commitAt(t, "backend/main.go", "feat: api", "2026-01-01T00:00:00Z")
	r.tag(t, "backend/v1.0.0")
	r.tag(t, "v9.0.0")

	identity, err := VersionIdentity(context.Background(), r.dir, sha, nil, backendScope)
	if err != nil || identity.Version != "backend/v1.0.0" {
		t.Fatalf("VersionIdentity() = %+v, %v", identity, err)
	}
	hotfix, err := AllocateHotfixVersion("backend/v1.0.0", []string{"backend/v1.0.1", "v1.0.2"}, backendScope)
	if err != nil || hotfix != "backend/v1.0.2" {
		t.Fatalf("AllocateHotfixVersion() = %q, %v", hotfix, err)
	}
	if _, err := AllocateHotfixVersion("v1.0.0", nil, backendScope); err == nil {
		t.Fatal("unprefixed production version accepted for a component")
	}
}
