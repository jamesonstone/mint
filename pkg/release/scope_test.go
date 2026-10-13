package release

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// commitAt writes one file at relPath and commits it.
func (repo *testRepo) commitAt(t *testing.T, relPath string, subject string, date string) string {
	t.Helper()
	repo.counter++
	full := filepath.Join(repo.dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(full, []byte(subject+"\n"), 0o644); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}
	repo.git(t, nil, "add", ".")
	repo.git(t, []string{"GIT_AUTHOR_DATE=" + date, "GIT_COMMITTER_DATE=" + date}, "commit", "-m", subject)
	return repo.revParse(t, "HEAD")
}

var (
	backendScope  = Scope{TagPrefix: "backend/", Paths: []string{"backend"}}
	frontendScope = Scope{TagPrefix: "frontend/", Paths: []string{"frontend"}}
)

func TestScopeValidate(t *testing.T) {
	valid := []Scope{{}, backendScope, {TagPrefix: "web-"}, {TagPrefix: "apps/api/"}, {Paths: []string{"backend", ":(glob)docs/**"}}}
	for _, scope := range valid {
		if err := scope.Validate(); err != nil {
			t.Errorf("Validate(%+v) error = %v", scope, err)
		}
	}
	invalid := []Scope{
		{TagPrefix: "/backend/"}, {TagPrefix: "back end/"}, {TagPrefix: "a..b/"}, {TagPrefix: "a//b"},
		{TagPrefix: "x."}, {TagPrefix: "-x"}, {Paths: []string{""}}, {Paths: []string{" backend"}}, {Paths: []string{"--all"}},
	}
	for _, scope := range invalid {
		if err := scope.Validate(); err == nil {
			t.Errorf("Validate(%+v) error = nil, want error", scope)
		}
	}
}

func TestResolveScopedComponentsAreIndependent(t *testing.T) {
	repo := newTestRepo(t)
	repo.commitAt(t, "README.md", "docs: root readme", "2024-01-01T00:00:00Z")
	repo.tag(t, "v1.0.0")
	backendCommit := repo.commitAt(t, "backend/main.go", "feat: add api", "2024-01-02T00:00:00Z")
	repo.commitAt(t, "frontend/app.ts", "fix: render home", "2024-01-03T00:00:00Z")

	backend, err := Resolve(context.Background(), Options{WorkDir: repo.dir, Scope: backendScope})
	if err != nil {
		t.Fatalf("Resolve(backend) error = %v", err)
	}
	if backend.VersionTag != "backend/v0.1.0" || backend.VersionBump != BumpMinor || backend.CommitCount != 1 || backend.BaseTag != "" {
		t.Fatalf("Resolve(backend) = %+v", backend)
	}
	if !strings.Contains(backend.ReleaseNotes, "add api") || strings.Contains(backend.ReleaseNotes, "render home") {
		t.Fatalf("backend notes include other components:\n%s", backend.ReleaseNotes)
	}

	frontend, err := Resolve(context.Background(), Options{WorkDir: repo.dir, Scope: frontendScope})
	if err != nil {
		t.Fatalf("Resolve(frontend) error = %v", err)
	}
	if frontend.VersionTag != "frontend/v0.0.1" || frontend.CommitCount != 1 {
		t.Fatalf("Resolve(frontend) = %+v", frontend)
	}

	repo.git(t, nil, "tag", "backend/v0.1.0", backendCommit)
	whole, err := Resolve(context.Background(), Options{WorkDir: repo.dir})
	if err != nil {
		t.Fatalf("Resolve(whole) error = %v", err)
	}
	if whole.BaseTag != "v1.0.0" || whole.VersionTag != "v1.1.0" || whole.CommitCount != 2 {
		t.Fatalf("Resolve(whole) = %+v; component tags must not affect unscoped releases", whole)
	}
}

func TestResolveScopedBumpsFromPrefixedBase(t *testing.T) {
	repo := newTestRepo(t)
	repo.commitAt(t, "backend/main.go", "feat: add api", "2024-01-01T00:00:00Z")
	repo.tag(t, "backend/v0.1.0")
	repo.tag(t, "v9.0.0")
	repo.commitAt(t, "frontend/app.ts", "feat!: redesign", "2024-01-02T00:00:00Z")

	unchanged, err := Resolve(context.Background(), Options{WorkDir: repo.dir, Scope: backendScope})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if unchanged.VersionTag != "backend/v0.1.0" || unchanged.NeedsGitTag || unchanged.CommitCount != 0 {
		t.Fatalf("frontend-only commit changed backend: %+v", unchanged)
	}

	repo.commitAt(t, "backend/fix.go", "fix: handle nil", "2024-01-03T00:00:00Z")
	next, err := Resolve(context.Background(), Options{WorkDir: repo.dir, Scope: backendScope})
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if next.BaseTag != "backend/v0.1.0" || next.VersionTag != "backend/v0.1.1" || next.CommitCount != 1 {
		t.Fatalf("Resolve() = %+v", next)
	}
}

func TestResolveRejectsInvalidScope(t *testing.T) {
	repo := newTestRepo(t)
	repo.commitAt(t, "a.txt", "feat: a", "2024-01-01T00:00:00Z")
	if _, err := Resolve(context.Background(), Options{WorkDir: repo.dir, Scope: Scope{TagPrefix: "bad prefix/"}}); err == nil {
		t.Fatal("Resolve() error = nil, want tag-prefix error")
	}
}

func TestSelectTagScoped(t *testing.T) {
	repo := newTestRepo(t)
	repo.commitAt(t, "backend/main.go", "feat: add api", "2024-01-01T00:00:00Z")
	repo.tag(t, "v1.0.0")
	repo.tag(t, "backend/v0.2.0")
	repo.tag(t, "frontend/v3.0.0")

	scoped, err := SelectTag(context.Background(), SelectTagOptions{WorkDir: repo.dir, Scope: backendScope})
	if err != nil || scoped.VersionTag != "backend/v0.2.0" {
		t.Fatalf("SelectTag(backend) = %+v, %v", scoped, err)
	}
	whole, err := SelectTag(context.Background(), SelectTagOptions{WorkDir: repo.dir})
	if err != nil || whole.VersionTag != "v1.0.0" {
		t.Fatalf("SelectTag(whole) = %+v, %v", whole, err)
	}
	if _, err := SelectTag(context.Background(), SelectTagOptions{WorkDir: repo.dir, Scope: backendScope, RequestedTag: "v1.0.0"}); err == nil ||
		err.Error() != "requested-tag: must be a strict backend/vX.Y.Z SemVer tag" {
		t.Fatalf("SelectTag(requested unscoped) error = %v", err)
	}

	empty := newTestRepo(t)
	empty.commitAt(t, "backend/main.go", "feat: add api", "2024-01-01T00:00:00Z")
	_, err = SelectTag(context.Background(), SelectTagOptions{WorkDir: empty.dir, Scope: backendScope})
	if err == nil || !strings.Contains(err.Error(), "no strict backend/vX.Y.Z SemVer tag points at HEAD") {
		t.Fatalf("SelectTag(no tag) error = %v", err)
	}
}

func TestCreateReleaseTagScoped(t *testing.T) {
	repo := newTestRepo(t)
	target := repo.commitAt(t, "backend/main.go", "feat: add api", "2024-01-01T00:00:00Z")
	notes := filepath.Join(t.TempDir(), "notes.md")
	if err := os.WriteFile(notes, []byte("notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := CreateReleaseTag(context.Background(), TagOptions{Tag: "backend/v0.1.0", Target: target, NotesFile: notes, WorkDir: repo.dir, Scope: backendScope})
	if err != nil || !result.Created || result.TagName != "backend/v0.1.0" {
		t.Fatalf("CreateReleaseTag(scoped) = %+v, %v", result, err)
	}
	if _, err := CreateReleaseTag(context.Background(), TagOptions{Tag: "v0.1.0", Target: target, NotesFile: notes, WorkDir: repo.dir, Scope: backendScope}); err == nil {
		t.Fatal("unprefixed tag accepted for a scoped release")
	}
	if _, err := CreateReleaseTag(context.Background(), TagOptions{Tag: "backend/v0.2.0", Target: target, NotesFile: notes, WorkDir: repo.dir}); err == nil {
		t.Fatal("prefixed tag accepted without a scope")
	}
}

func TestScopedVersionFromTag(t *testing.T) {
	if got, err := ScopedVersionFromTag("backend/v1.2.3", backendScope); err != nil || got != "1.2.3" {
		t.Fatalf("ScopedVersionFromTag() = %q, %v", got, err)
	}
	if _, err := ScopedVersionFromTag("v1.2.3", backendScope); err == nil || err.Error() != "version tag must be strict backend/vX.Y.Z SemVer" {
		t.Fatalf("ScopedVersionFromTag(unprefixed) error = %v", err)
	}
	if _, err := VersionFromTag("backend/v1.2.3"); err == nil || err.Error() != "version tag must be strict vX.Y.Z SemVer" {
		t.Fatalf("VersionFromTag(prefixed) error = %v", err)
	}
}

func TestPublishGitHubReleaseScopedTag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.EscapedPath() == "/repos/acme/app/releases/tags/backend%2Fv1.2.3":
			http.NotFound(w, r)
		case r.Method == http.MethodPost && r.URL.Path == "/repos/acme/app/releases":
			var payload gitHubReleaseRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if payload.TagName != "backend/v1.2.3" {
				t.Fatalf("tag_name = %q", payload.TagName)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"tag_name":"backend/v1.2.3","html_url":"https://github.com/acme/app/releases/tag/backend%2Fv1.2.3"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.EscapedPath())
		}
	}))
	defer server.Close()

	result, err := PublishGitHubRelease(context.Background(), GitHubReleaseOptions{
		Owner: "acme", Repo: "app", Tag: "backend/v1.2.3", Target: "abc123", Notes: "notes",
		Token: "t", APIBaseURL: server.URL, HTTPClient: server.Client(), Scope: backendScope,
	})
	if err != nil || result.TagName != "backend/v1.2.3" {
		t.Fatalf("PublishGitHubRelease() = %+v, %v", result, err)
	}
}
