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
	if b, _ := backend.JournalBranch("production"); b != "mint-release-state.backend.production" {
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
	if requested != "/repos/owner/repo/git/ref/heads/mint-release-state.backend.production" {
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

// Unscoped authority digests must match releases before components existed, or
// recorded intents would be refused after upgrading Mint.
func TestUnscopedAuthorityDigestAndJSONUnchanged(t *testing.T) {
	p, err := ParsePolicy([]byte(environmentPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := p.ForEnvironment("production")
	if err != nil {
		t.Fatal(err)
	}
	digest, err := AuthorityDigest(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if digest != "sha256:427356f64495234f22a8ea5a21cba9ac462d1e6b6966ba3f9d334d6c802a9417" {
		t.Fatalf("unscoped AuthorityDigest changed: %s", digest)
	}
	data, _ := json.Marshal(cfg)
	if strings.Contains(string(data), "Component") {
		t.Fatalf("unscoped Config JSON gained a Component key: %s", data)
	}
	scoped, err := ParsePolicy([]byte(componentPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}
	scopedCfg, _ := scoped.ForEnvironment("production")
	data, _ = json.Marshal(scopedCfg)
	if !strings.Contains(string(data), `"Component":{"name":"backend"`) {
		t.Fatalf("component Config JSON lacks the component: %s", data)
	}
}

func TestJournalAndStateNamesNeverCollide(t *testing.T) {
	seen := map[string]string{}
	add := func(label, branch string, err error) {
		t.Helper()
		if err != nil {
			t.Fatal(label, err)
		}
		if prior, ok := seen[branch]; ok {
			t.Fatalf("%s and %s share journal %s", prior, label, branch)
		}
		seen[branch] = label
	}
	for _, env := range []string{"production", "backend-production", "b-c", "c", "a--b"} {
		b, err := Component{}.JournalBranch(env)
		add("unscoped/"+env, b, err)
	}
	for _, pair := range [][2]string{{"backend", "production"}, {"a", "b-c"}, {"a-b", "c"}} {
		b, err := Component{Name: pair[0]}.JournalBranch(pair[1])
		add(pair[0]+"/"+pair[1], b, err)
	}
	keys := map[string]bool{}
	for _, key := range []string{Component{}.StateKey("a--b"), Component{Name: "a"}.StateKey("b-c"), Component{Name: "a-b"}.StateKey("c"), Component{}.StateKey("a-b-c")} {
		if keys[key] {
			t.Fatalf("state key collision %s", key)
		}
		keys[key] = true
	}
}

func TestComponentHotfixIdentityIsScoped(t *testing.T) {
	unscoped := Component{}
	if unscoped.HotfixBaseBranch(7) != "mint-hotfix-base/GH-7" || unscoped.HotfixSourceMarker(7, "production-a") != "<!-- mint:hotfix-source:GH-7:production-a -->" || unscoped.HotfixRequestMarker(3, sha(1)) != "<!-- mint:hotfix-request:3:"+sha(1)+" -->" {
		t.Fatal("unscoped hotfix identity changed")
	}
	if backend.HotfixBaseBranch(7) != "mint-hotfix-base/backend/GH-7" {
		t.Fatal(backend.HotfixBaseBranch(7))
	}
	body := "intro\n" + backend.HotfixSourceMarker(7, "production-b") + "\nmore"
	if id, ok := backend.HotfixSourceBaseline(body); !ok || id != "production-b" {
		t.Fatalf("HotfixSourceBaseline() = %q, %v", id, ok)
	}
	if _, ok := unscoped.HotfixSourceBaseline(body); ok {
		t.Fatal("repository-wide scope matched a component hotfix")
	}
	if _, ok := (Component{Name: "frontend"}).HotfixSourceBaseline(body); ok {
		t.Fatal("another component matched a backend hotfix")
	}
	request := backend.HotfixRequestMarker(3, sha(1)) + "\n<!-- mint:hotfix-baseline:production-b -->\n"
	if !backend.hotfixRequestPattern().MatchString(request) || unscoped.hotfixRequestPattern().MatchString(request) {
		t.Fatal("hotfix request pattern not scoped")
	}
}

func TestVerifyHotfixSourceRequiresOwnScopeAndBaseline(t *testing.T) {
	pull := func(body, base string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"number": 9, "merged": true, "merge_commit_sha": sha(5), "body": body, "user": map[string]string{"login": "jamesonstone"},
				"head": map[string]any{"sha": sha(5), "ref": "GH-7", "repo": map[string]string{"full_name": "owner/repo"}},
				"base": map[string]any{"ref": base, "repo": map[string]string{"full_name": "owner/repo"}},
			})
		}))
	}
	candidate := Candidate{SourcePR: 9, SourceSHA: sha(5), BaselineID: "production-b"}
	for name, server := range map[string]*httptest.Server{
		"other baseline":  pull(backend.HotfixSourceMarker(7, "production-a"), backend.HotfixBaseBranch(7)),
		"other component": pull(Component{Name: "frontend"}.HotfixSourceMarker(7, "production-b"), Component{Name: "frontend"}.HotfixBaseBranch(7)),
		"repo-wide base":  pull(backend.HotfixSourceMarker(7, "production-b"), Component{}.HotfixBaseBranch(7)),
	} {
		c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo", HumanLogin: "jamesonstone", Component: backend}
		if err := c.VerifyHotfixSource(context.Background(), candidate, nil); err == nil || !strings.Contains(err.Error(), "production-base source PR") {
			t.Errorf("%s: VerifyHotfixSource() error = %v", name, err)
		}
		server.Close()
	}
}

func TestImageTagAndChangelogValidation(t *testing.T) {
	withTag := func(policy, tag string) string {
		return strings.Replace(policy, "    deploy: reviewed\n", "    deploy: reviewed\n    image_tag: "+tag+"\n", 1)
	}
	devTag := func(tag string) string {
		return strings.Replace(environmentPolicyYAML, "    image_tag: dev\n", "    image_tag: "+tag+"\n", 1)
	}
	if _, err := ParsePolicy([]byte(devTag("stable-v1.2.3"))); err != nil {
		t.Fatalf("unscoped alias stable-v1.2.3 rejected: %v", err)
	}
	if _, err := ParsePolicy([]byte(devTag("v1.2.3"))); err == nil {
		t.Fatal("image_tag equal to a version tag accepted")
	}
	if _, err := ParsePolicy([]byte(withTag(componentPolicyYAML, "backend-v1.0.0"))); err != nil {
		t.Fatalf("component alias rejected: %v", err)
	}
	for _, changelog := range []string{"./x.md", "a//b.md", ".mint/summary.md", ".mint/frontend/summary.md", ".github/CHANGELOG.md", "/abs.md", "../up.md", "notes.txt"} {
		policy := strings.Replace(componentPolicyYAML, "changelog: backend/CHANGELOG.md", "changelog: "+changelog, 1)
		if _, err := ParsePolicy([]byte(policy)); err == nil {
			t.Errorf("unsafe changelog %q accepted", changelog)
		}
	}
}

func TestObservationVersionMatchesJournalScope(t *testing.T) {
	observation := func(version string) Observation {
		return Observation{Repository: "owner/repo", Environment: "production", Version: version, Status: "unknown",
			ObservedAt: "2026-01-01T00:00:00Z", RunID: 1, RunURL: "https://github.com/owner/repo/actions/runs/1"}
	}
	component := NewComponentState("owner/repo", "production", backend)
	if _, err := ApplyObservation(&component, observation("v1.2.3")); err == nil {
		t.Fatal("component journal accepted an unprefixed observation")
	}
	if _, err := ApplyObservation(&component, observation("backend/v1.2.3")); err != nil {
		t.Fatalf("component observation rejected: %v", err)
	}
	unscoped := NewState("owner/repo", "production")
	if _, err := ApplyObservation(&unscoped, observation("backend/v1.2.3")); err == nil {
		t.Fatal("repository-wide journal accepted a component observation")
	}
}

func TestForeignControlPRDetection(t *testing.T) {
	frontend := Component{Name: "frontend", TagPrefix: "frontend/"}
	cases := []struct {
		scope Component
		body  string
		want  bool
	}{
		{Component{}, backend.ProposalMarker("p", "normal"), true},
		{Component{}, backend.HotfixSourceMarker(1, "b"), true},
		{Component{}, Component{}.ProposalMarker("p", "normal"), false},
		{Component{}, "<!-- mint:entry:x -->", false},
		{backend, Component{}.ProposalMarker("p", "normal"), true},
		{backend, frontend.ProposalMarker("p", "normal"), true},
		{backend, Component{}.HotfixSourceMarker(1, "b"), true},
		{backend, backend.ProposalMarker("p", "normal"), false},
		{backend, backend.HotfixSourceMarker(1, "b"), false},
	}
	for _, tt := range cases {
		if got := tt.scope.ForeignControlPR(tt.body); got != tt.want {
			t.Errorf("%q.ForeignControlPR(%q) = %v, want %v", tt.scope.Name, tt.body, got, tt.want)
		}
	}
}

func TestComponentRejectsLockEnvironment(t *testing.T) {
	policy := strings.Replace(componentPolicyYAML, "  production:\n", "  lock:\n", 1)
	if _, err := ParsePolicy([]byte(policy)); err == nil {
		t.Fatal("component environment named lock accepted; its journal ref would end in .lock")
	}
	unscoped := strings.Replace(environmentPolicyYAML, "  dev:\n", "  lock:\n", 1)
	unscoped = strings.Replace(unscoped, "requires: [dev]", "requires: [lock]", 1)
	if _, err := ParsePolicy([]byte(unscoped)); err != nil {
		t.Fatalf("unscoped lock environment rejected: %v", err)
	}
}
