package promotion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const componentPolicyYAML = `schema_version: 2
mode: deployment
repository: owner/repo
default_branch: main
human_login: jamesonstone
build_workflow: .github/workflows/backend-build.yaml
control_workflow: .github/workflows/backend-mint.yaml
component:
  name: backend
  paths: [backend, shared/go.mod]
  changelog: backend/CHANGELOG.md
environments:
  production:
    scope: shared
    deploy: reviewed
    target: backend/v1.2.3
    promotion_workflow: .github/workflows/backend-deploy.yaml
    publish: true
`

var backend = Component{Name: "backend", TagPrefix: "backend/", Paths: []string{"backend"}, Changelog: "backend/CHANGELOG.md"}

func TestComponentPolicyParsesAndScopesConfig(t *testing.T) {
	p, err := ParsePolicy([]byte(componentPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}
	component := p.ComponentScope()
	if component.Name != "backend" || component.TagPrefix != "backend/" || len(component.Paths) != 2 {
		t.Fatalf("component = %+v", component)
	}
	want := []string{"backend/CHANGELOG.md", ".mint/backend/proposal.json", ".mint/backend/summary.md"}
	if strings.Join(p.ControlPaths, ",") != strings.Join(want, ",") {
		t.Fatalf("control paths = %v, want %v", p.ControlPaths, want)
	}
	cfg, err := p.ForEnvironment("production")
	if err != nil || cfg.Component.Name != "backend" || cfg.Target != "backend/v1.2.3" {
		t.Fatalf("ForEnvironment() = %+v, %v", cfg.Component, err)
	}
	unscoped, err := ParsePolicy([]byte(environmentPolicyYAML))
	if err != nil || !unscoped.ComponentScope().IsZero() {
		t.Fatalf("unscoped policy gained a component: %+v %v", unscoped.ComponentScope(), err)
	}
}

func TestComponentPolicyValidation(t *testing.T) {
	cases := map[string]string{
		"no control workflow":    strings.Replace(componentPolicyYAML, "control_workflow: .github/workflows/backend-mint.yaml\n", "", 1),
		"bad name":               strings.Replace(componentPolicyYAML, "name: backend", "name: Backend", 1),
		"no paths":               strings.Replace(componentPolicyYAML, "paths: [backend, shared/go.mod]", "paths: []", 1),
		"no changelog":           strings.Replace(componentPolicyYAML, "  changelog: backend/CHANGELOG.md\n", "", 1),
		"unprefixed target":      strings.Replace(componentPolicyYAML, "target: backend/v1.2.3", "target: v1.2.3", 1),
		"foreign target":         strings.Replace(componentPolicyYAML, "target: backend/v1.2.3", "target: frontend/v1.2.3", 1),
		"repo-wide control":      strings.Replace(componentPolicyYAML, "component:", "control_paths: [CHANGELOG.md, .mint/proposal.json, .mint/summary.md]\ncomponent:", 1),
		"prefix starting with v": strings.Replace(componentPolicyYAML, "  name: backend\n", "  name: backend\n  tag_prefix: v\n", 1),
	}
	for name, policy := range cases {
		if _, err := ParsePolicy([]byte(policy)); err == nil {
			t.Errorf("%s: ParsePolicy() error = nil", name)
		}
	}
	pkg := strings.Replace(componentPolicyYAML, "mode: deployment", "mode: package", 1)
	pkg = strings.Replace(pkg, "control_workflow: .github/workflows/backend-mint.yaml\n", "", 1)
	pkg = strings.Replace(pkg, "    promotion_workflow: .github/workflows/backend-deploy.yaml\n", "", 1)
	if _, err := ParsePolicy([]byte(pkg)); err != nil {
		t.Fatalf("package component without a controller rejected: %v", err)
	}
	custom := strings.Replace(componentPolicyYAML, "  name: backend\n", "  name: backend\n  tag_prefix: api-\n", 1)
	custom = strings.Replace(custom, "target: backend/v1.2.3", "target: api-v1.2.3", 1)
	p, err := ParsePolicy([]byte(custom))
	if err != nil || p.ComponentScope().TagPrefix != "api-" {
		t.Fatalf("custom prefix: %+v %v", p.ComponentScope(), err)
	}
}

func TestComponentNamingIsolatesScopes(t *testing.T) {
	unscoped := Component{}
	if unscoped.ProposalPath() != ".mint/proposal.json" || unscoped.ChangelogPath() != "CHANGELOG.md" || unscoped.MarkerPrefix() != "<!-- mint:proposal:" || unscoped.ControlConcurrencyGroup() != "mint-release-journal" || unscoped.StateKey("dev") != "dev" {
		t.Fatal("unscoped naming changed")
	}
	if b, _ := unscoped.JournalBranch("production"); b != "mint-release-state" {
		t.Fatal(b)
	}
	if backend.ProposalPath() != ".mint/backend/proposal.json" || backend.SummaryPath() != ".mint/backend/summary.md" || backend.ChangelogPath() != "backend/CHANGELOG.md" {
		t.Fatal("component control paths")
	}
	if b, _ := backend.JournalBranch("production"); b != "mint-release-state-backend-production" {
		t.Fatal(b)
	}
	marker := backend.ProposalMarker("p1", "normal")
	if marker != "<!-- mint:component-proposal:backend:p1:normal -->" || strings.Contains(marker, unscoped.MarkerPrefix()) {
		t.Fatalf("component marker %q must not match the unscoped prefix", marker)
	}
	if strings.Contains(marker, Component{Name: "back"}.MarkerPrefix()) {
		t.Fatal("component marker prefix matches a shorter component name")
	}
	if !backend.ValidVersion("backend/v1.2.3") || backend.ValidVersion("v1.2.3") || backend.ValidVersion("frontend/v1.2.3") || unscoped.ValidVersion("backend/v1.2.3") || !unscoped.ValidVersion("v1.2.3") {
		t.Fatal("version scoping")
	}
}

func TestComponentStateRejectsForeignJournalsAndVersions(t *testing.T) {
	state := NewComponentState("owner/repo", "production", backend)
	data, _ := json.Marshal(state)
	if _, err := DecodeComponentState(data, "owner/repo", "production", backend); err != nil {
		t.Fatalf("own journal rejected: %v", err)
	}
	if _, err := DecodeState(data, "owner/repo", "production"); err == nil {
		t.Fatal("unscoped decode accepted a component journal")
	}
	if _, err := DecodeComponentState(data, "owner/repo", "production", Component{Name: "frontend", TagPrefix: "frontend/"}); err == nil {
		t.Fatal("another component's journal accepted")
	}

	c := candidate(1, "normal")
	c.Version, c.Changes[0].Version = "backend/v0.1.1", "backend/v0.1.1"
	if err := state.RegisterCandidate(c, evidence(c)); err != nil {
		t.Fatalf("component candidate rejected: %v", err)
	}
	unscoped := NewState("owner/repo", "production")
	if err := unscoped.RegisterCandidate(c, evidence(c)); err == nil {
		t.Fatal("unscoped journal accepted a component version")
	}
	plain := candidate(2, "normal")
	if err := state.RegisterCandidate(plain, evidence(plain)); err == nil {
		t.Fatal("component journal accepted an unprefixed version")
	}
}

func TestComponentJournalUsesComponentBranch(t *testing.T) {
	var requested string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = r.URL.Path
		w.WriteHeader(404)
	}))
	defer server.Close()
	c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo", Component: backend}
	snapshot, err := c.LoadJournal(context.Background(), "production")
	if err != nil {
		t.Fatal(err)
	}
	if requested != "/repos/owner/repo/git/ref/heads/mint-release-state-backend-production" {
		t.Fatalf("requested %s", requested)
	}
	if snapshot.State.ComponentName != "backend" || snapshot.State.TagPrefix != "backend/" {
		t.Fatalf("new journal lacks component identity: %+v", snapshot.State)
	}
	foreign := JournalSnapshot{State: NewState("owner/repo", "production")}
	if _, err := c.SaveJournal(context.Background(), foreign, "x"); err == nil {
		t.Fatal("component client saved an unscoped journal")
	}
}

func TestComponentCollectChangesOnlyWalksComponentPaths(t *testing.T) {
	dir := t.TempDir()
	gitTest(t, dir, "init", "-b", "main")
	gitTest(t, dir, "config", "user.name", "Human")
	gitTest(t, dir, "config", "user.email", "human@example.com")
	writeTest(t, dir, "README.md", "root\n")
	gitTest(t, dir, "add", ".")
	gitTest(t, dir, "commit", "-m", "chore: root")
	anchor := gitTest(t, dir, "rev-parse", "HEAD")
	gitTest(t, dir, "tag", "v9.9.9")
	gitTest(t, dir, "tag", "backend/v0.1.0")
	writeTest(t, dir, "frontend.txt", "ui\n")
	gitTest(t, dir, "add", ".")
	gitTest(t, dir, "commit", "-m", "feat: frontend only")
	gitTest(t, dir, "tag", "frontend/v0.1.0")
	if err := gitMkdirWrite(dir, "backend/api.go", "package api\n"); err != nil {
		t.Fatal(err)
	}
	gitTest(t, dir, "add", ".")
	gitTest(t, dir, "commit", "-m", "feat: backend api")
	source := gitTest(t, dir, "rev-parse", "HEAD")
	gitTest(t, dir, "tag", "backend/v0.2.0")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()
	client := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo", Component: backend}
	state := NewComponentState("owner/repo", "production", backend)
	state.MainAnchor = anchor
	base := candidate(0, "normal")
	base.SourceSHA, base.Version, base.Changes = anchor, "backend/v0.1.0", nil
	state.Baseline = &Baseline{ID: "production-0", Candidate: base}
	proof := GitProof{Context: t.Context(), WorkDir: dir, DefaultBranch: "main"}
	changes, err := client.CollectChanges(t.Context(), proof, state, Candidate{Kind: "normal", SourceSHA: source, Version: "backend/v0.2.0"}, nil)
	if err != nil {
		t.Fatalf("CollectChanges() error = %v (frontend commit must be skipped)", err)
	}
	if len(changes) != 1 || changes[0].Version != "backend/v0.2.0" || changes[0].SHA != source {
		t.Fatalf("changes = %+v", changes)
	}
}

func TestComponentPublicationEscapesTagAndSkipsLatest(t *testing.T) {
	var refPath, releasePath string
	var latest any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && strings.Contains(r.URL.Path, "/git/ref/tags/"):
			refPath = r.URL.EscapedPath()
			_ = json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": sha(1), "type": "commit"}})
		case r.Method == "GET" && strings.Contains(r.URL.EscapedPath(), "/releases/tags/"):
			releasePath = r.URL.EscapedPath()
			w.WriteHeader(404)
		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/releases"):
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			latest = body["make_latest"]
			w.WriteHeader(201)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 7, "tag_name": body["tag_name"]})
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo", Component: backend}
	intent := Intent{Status: "deployed", Candidate: Candidate{Version: "backend/v1.2.3", SourceSHA: sha(1)}}
	_ = c.PublishIntent(context.Background(), intent)
	if refPath != "/repos/owner/repo/git/ref/tags/backend/v1.2.3" {
		t.Errorf("ref path = %s", refPath)
	}
	if releasePath != "/repos/owner/repo/releases/tags/backend%2Fv1.2.3" {
		t.Errorf("release path = %s", releasePath)
	}
	if latest != "false" {
		t.Errorf("component release make_latest = %v, want false", latest)
	}
	if (Client{}).makeLatest() != "true" {
		t.Error("repository-wide releases must stay latest")
	}
}

func TestComponentWorkflowsAreNamedAndSerializedPerComponent(t *testing.T) {
	cfg := controlConfig()
	cfg.Component = backend
	cfg.ControlWorkflow = ".github/workflows/backend-mint.yaml"
	out, err := RenderControlWorkflow(cfg, strings.Repeat("a", 40), ".mint/backend.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "name: Production release control (backend)") || !strings.Contains(out, "group: mint-release-journal-backend") {
		t.Fatalf("control workflow not scoped:\n%s", out)
	}
	p, err := ParsePolicy([]byte(componentPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}
	env, err := RenderEnvironmentWorkflow(p, strings.Repeat("a", 40), ".mint/backend.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(env, "name: Mint environment control (backend)\n") {
		t.Fatalf("environment workflow not named per component:\n%s", env[:80])
	}
}

func gitMkdirWrite(dir, rel, value string) error {
	path := dir + "/" + rel
	if err := os.MkdirAll(path[:strings.LastIndex(path, "/")], 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(value), 0o644)
}
