package promotion

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

const legacyProductionPolicyYAML = `schema_version: 1
repository: owner/repo
environment: production
default_branch: main
human_login: jamesonstone
build_workflow: .github/workflows/build.yaml
validation_workflow: .github/workflows/checks.yaml
promotion_workflow: .github/workflows/deploy.yaml
required_checks: [native]
`

const chainedBuildEventsYAML = "build_events: [push, workflow_dispatch, workflow_run]\n"

func TestBuildEventsKeepStrictDefaultAndFollowPolicyOptIn(t *testing.T) {
	strict := []string{"push", "workflow_dispatch"}
	cases := []struct {
		name, field      string
		allowed, refused []string
	}{
		{"absent", "", strict, []string{"workflow_run", "pull_request", "pull_request_target", "schedule", ""}},
		{"null", "build_events:\n", strict, []string{"workflow_run", "pull_request"}},
		{"explicit-default", "build_events: [push, workflow_dispatch]\n", strict, []string{"workflow_run", "pull_request"}},
		{"chained-only", "build_events: [workflow_run]\n", []string{"workflow_run"}, []string{"push", "workflow_dispatch", "pull_request"}},
		{"all", chainedBuildEventsYAML, []string{"push", "workflow_dispatch", "workflow_run"}, []string{"pull_request", "pull_request_target", "Workflow_Run"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			policy, err := ParsePolicy([]byte(environmentPolicyYAML + tc.field))
			if err != nil {
				t.Fatal(err)
			}
			for _, name := range []string{"dev", "production"} {
				cfg, err := policy.ForEnvironment(name)
				if err != nil {
					t.Fatal(err)
				}
				for _, event := range tc.allowed {
					if !cfg.BuildEventAllowed(event) {
						t.Fatal("configured producer event refused", name, event)
					}
				}
				for _, event := range tc.refused {
					if cfg.BuildEventAllowed(event) {
						t.Fatal("unconfigured producer event trusted", name, event)
					}
				}
			}
		})
	}
	policy, err := ParsePolicy([]byte(environmentPolicyYAML + chainedBuildEventsYAML))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := policy.ForEnvironment("dev")
	if err != nil {
		t.Fatal(err)
	}
	cfg.BuildEvents[0] = "pull_request"
	if policy.BuildEvents[0] != "push" {
		t.Fatal("effective policy aliases the mutable producer event list")
	}
	// The effective list is sorted, and the default set is nil: order and an
	// explicit default are not authority changes.
	for field, want := range map[string][]string{
		"build_events: [workflow_dispatch, push]\n":               nil,
		"build_events: [workflow_run, workflow_dispatch, push]\n": {"push", "workflow_dispatch", "workflow_run"},
		"build_events: [workflow_run, workflow_dispatch]\n":       {"workflow_dispatch", "workflow_run"},
	} {
		policy, err := ParsePolicy([]byte(environmentPolicyYAML + field))
		if err != nil {
			t.Fatal(err)
		}
		cfg, err := policy.ForEnvironment("dev")
		if err != nil || !reflect.DeepEqual(cfg.BuildEvents, want) {
			t.Fatal("producer events not normalized", field, cfg.BuildEvents, err)
		}
	}
}

func TestBuildEventsRejectEmptyRepeatedUnknownAndLegacyPolicy(t *testing.T) {
	cases := map[string]string{
		"empty":        "build_events: []\n",
		"repeated":     "build_events: [push, push]\n",
		"unknown":      "build_events: [push, pull_request]\n",
		"fork-capable": "build_events: [pull_request_target]\n",
		"blank":        "build_events: ['']\n",
		"case":         "build_events: [Workflow_Run]\n",
		"scalar":       "build_events: workflow_run\n",
		"mapping":      "build_events: {workflow_run: true}\n",
	}
	for name, field := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParsePolicy([]byte(environmentPolicyYAML + field)); err == nil {
				t.Fatal("invalid producer event list accepted")
			}
		})
	}
	if _, err := ParsePolicy([]byte(legacyProductionPolicyYAML)); err != nil {
		t.Fatal("legacy fixture should be valid without the new fields", err)
	}
	for _, field := range []string{chainedBuildEventsYAML, "activation_variable: MINT_OTHER_ENABLED\n"} {
		if _, err := ParsePolicy([]byte(legacyProductionPolicyYAML + field)); err == nil {
			t.Fatal("schema 1 policy accepted a schema 2 field", field)
		}
	}
}

func controllerGate(t *testing.T, rendered string) string {
	t.Helper()
	var workflow map[string]any
	if err := yaml.Unmarshal([]byte(rendered), &workflow); err != nil {
		t.Fatal(err)
	}
	return workflow["jobs"].(map[string]any)["control"].(map[string]any)["if"].(string)
}

func TestActivationVariableSelectsTheGeneratedControllerGate(t *testing.T) {
	policy, err := ParsePolicy([]byte(environmentPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := RenderEnvironmentWorkflow(policy, sha(1), ".mint.yaml", "Build reusable artifact")
	if err != nil {
		t.Fatal(err)
	}
	// Output of v1.0.3 (a628b39) for this policy: a policy without the field
	// must keep its controller byte for byte.
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(rendered))); got != "bb463927162edd5b03fb2d1f51c3fe638019eaddb6715ea3fb0a59d57f038698" {
		t.Fatal("generated controller changed for a policy without activation_variable", got)
	}
	if gate := controllerGate(t, rendered); gate != "vars.MINT_RELEASE_ENABLED == 'true'" {
		t.Fatal(gate)
	}
	custom, err := ParsePolicy([]byte(environmentPolicyYAML + "activation_variable: MINT_SANDBOX_ENABLED\n"))
	if err != nil {
		t.Fatal(err)
	}
	if custom.ControllerActivationVariable() != "MINT_SANDBOX_ENABLED" || policy.ControllerActivationVariable() != DefaultActivationVariable {
		t.Fatal("activation variable not resolved")
	}
	rendered, err = RenderEnvironmentWorkflow(custom, sha(1), ".mint.yaml", "Build reusable artifact")
	if err != nil {
		t.Fatal(err)
	}
	if gate := controllerGate(t, rendered); gate != "vars.MINT_SANDBOX_ENABLED == 'true'" {
		t.Fatal(gate)
	}
	if strings.Contains(rendered, "MINT_RELEASE_ENABLED") || strings.Count(rendered, "vars.") != 1 {
		t.Fatal("controller keeps a second activation gate")
	}
}

func TestActivationVariableRejectsUnsafeNames(t *testing.T) {
	policy, err := ParsePolicy([]byte(environmentPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}
	// A switch outside the MINT_ prefix would enter the compatibility adapters'
	// runtime configuration identity and change it on every flip.
	for _, name := range []string{"CD_ENABLED", "SANDBOX_ENABLED", "A", "MINT_", "MINT", "MINTX_ENABLED", "XMINT_ENABLED", "mint_sandbox_enabled", "mint_ENABLED", "Mint_ENABLED", "Mint_Enabled", "MINT_enabled", "1ENABLED", "_ENABLED", "GITHUB_TOKEN", "GITHUB_ENABLED", "MINT_SANDBOX-ENABLED", "MINT_SANDBOX ENABLED", "MINT_SANDBOX.ENABLED", "MINT_X == 'true' || true", "MINT_${{ secrets.TOKEN }}", "MINT_X\nY", "MINT_" + strings.Repeat("A", 251)} {
		t.Run(name, func(t *testing.T) {
			scalar, err := yaml.Marshal(name)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ParsePolicy([]byte(environmentPolicyYAML + "activation_variable: " + string(scalar))); err == nil {
				t.Fatal("unsafe activation variable accepted by policy")
			}
			unsafe := policy
			unsafe.ActivationVariable = name
			if _, err := RenderEnvironmentWorkflow(unsafe, sha(1), ".mint.yaml"); err == nil {
				t.Fatal("unsafe activation variable rendered into the controller")
			}
		})
	}
	for _, name := range []string{"MINT_RELEASE_ENABLED", "MINT_SANDBOX_ENABLED", "MINT_CD_2", "MINT_" + strings.Repeat("A", 250)} {
		if _, err := ParsePolicy([]byte(environmentPolicyYAML + "activation_variable: " + name + "\n")); err != nil {
			t.Fatal("valid activation variable rejected", name, err)
		}
	}
}

func TestAuthorityDigestIsStableWithoutNewFieldsAndBindsThem(t *testing.T) {
	policy, err := ParsePolicy([]byte(environmentPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}
	// Digests of v1.0.3 (a628b39) for this policy. Journals store them: a
	// change would invalidate every pending approval of an existing consumer.
	golden := map[string]string{
		"dev":        "sha256:8498e4c2d367891faa6b43ffabbe7a46726a8f6a8f8781f076c58085001f0497",
		"production": "sha256:427356f64495234f22a8ea5a21cba9ac462d1e6b6966ba3f9d334d6c802a9417",
	}
	for name, want := range golden {
		cfg, err := policy.ForEnvironment(name)
		if err != nil {
			t.Fatal(err)
		}
		if got, err := AuthorityDigest(cfg); err != nil || got != want {
			t.Fatal("authority digest changed for a policy without the new fields", name, got, err)
		}
	}
	cfg, err := policy.ForEnvironment("production")
	if err != nil {
		t.Fatal(err)
	}
	mutations := map[string]func(*Config){
		"chained-producer": func(c *Config) { c.BuildEvents = []string{"push", "workflow_dispatch", "workflow_run"} },
		"chained-only":     func(c *Config) { c.BuildEvents = []string{"workflow_run"} },
		"activation":       func(c *Config) { c.ActivationVariable = "MINT_SANDBOX_ENABLED" },
	}
	seen := map[string]string{golden["production"]: "unset"}
	for name, mutate := range mutations {
		altered := cfg
		mutate(&altered)
		digest, err := AuthorityDigest(altered)
		if err != nil {
			t.Fatal(err)
		}
		if prior, ok := seen[digest]; ok {
			t.Fatal("authority change omitted from frozen digest", name, prior)
		}
		seen[digest] = name
	}
	for _, field := range []string{chainedBuildEventsYAML, "activation_variable: MINT_SANDBOX_ENABLED\n"} {
		changed, err := ParsePolicy([]byte(environmentPolicyYAML + field))
		if err != nil {
			t.Fatal(err)
		}
		cfg, err := changed.ForEnvironment("production")
		if err != nil {
			t.Fatal(err)
		}
		if digest, err := AuthorityDigest(cfg); err != nil || digest == golden["production"] {
			t.Fatal("parsed policy field omitted from frozen digest", field, err)
		}
	}
	digest := func(field string) string {
		t.Helper()
		parsed, err := ParsePolicy([]byte(environmentPolicyYAML + field))
		if err != nil {
			t.Fatal(err)
		}
		cfg, err := parsed.ForEnvironment("production")
		if err != nil {
			t.Fatal(err)
		}
		value, err := AuthorityDigest(cfg)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	// The same trust set has one digest: writing the default, or another order,
	// must not invalidate pending approvals.
	for _, field := range []string{"build_events: [push, workflow_dispatch]\n", "build_events: [workflow_dispatch, push]\n", "build_events:\n"} {
		if digest(field) != golden["production"] {
			t.Fatal("default producer events changed the authority digest", field)
		}
	}
	if digest("build_events: [workflow_run, push]\n") != digest("build_events: [push, workflow_run]\n") {
		t.Fatal("producer event order changed the authority digest")
	}
}

func TestTargetOnlyRequestCannotChangeBuildEventsOrActivationVariable(t *testing.T) {
	requested := strings.Replace(environmentPolicyYAML, "target: v1.2.3", "target: v1.2.4", 1)
	parse := func(data string) Policy {
		t.Helper()
		policy, err := ParsePolicy([]byte(data))
		if err != nil {
			t.Fatal(err)
		}
		return policy
	}
	for _, field := range []string{chainedBuildEventsYAML, "build_events: [workflow_run]\n", "activation_variable: MINT_OTHER_ENABLED\n"} {
		t.Run(strings.TrimSpace(field), func(t *testing.T) {
			if _, err := PolicyRequestEnvironment(parse(environmentPolicyYAML), parse(requested+field)); err == nil {
				t.Fatal("deployment request added policy authority")
			}
			if _, err := PolicyRequestEnvironment(parse(environmentPolicyYAML+field), parse(requested)); err == nil {
				t.Fatal("deployment request removed policy authority")
			}
			if name, err := PolicyRequestEnvironment(parse(environmentPolicyYAML+field), parse(requested+field)); err != nil || name != "production" {
				t.Fatal("unchanged authority refused a target-only request", name, err)
			}
		})
	}
}

func TestProducerRunAuthenticationFollowsBuildEvents(t *testing.T) {
	chained := Config{BuildEvents: []string{"push", "workflow_dispatch", "workflow_run"}}
	chainedOnly := Config{BuildEvents: []string{"workflow_run"}}
	path := ".github/workflows/build.yaml"
	// adapter: promotion, baseline and observation runs keep the fixed set.
	// strict, optIn, chainOnly: a successful producer under each policy.
	// observed: the opt-in policy before completion is known.
	cases := []struct {
		name                                        string
		mutate                                      func(*WorkflowRun)
		adapter, strict, optIn, chainOnly, observed bool
	}{
		{"push", func(r *WorkflowRun) { r.Event = "push" }, true, true, true, false, true},
		{"dispatch", func(r *WorkflowRun) { r.Event = "workflow_dispatch" }, true, true, true, false, true},
		{"chained", func(r *WorkflowRun) { r.Event = "workflow_run" }, false, false, true, true, true},
		{"pull-request", func(r *WorkflowRun) { r.Event = "pull_request" }, false, false, false, false, false},
		{"fork-capable", func(r *WorkflowRun) { r.Event = "pull_request_target" }, false, false, false, false, false},
		{"chained-foreign-repository", func(r *WorkflowRun) { r.Event, r.Repository.FullName = "workflow_run", "other/repo" }, false, false, false, false, false},
		{"chained-foreign-head", func(r *WorkflowRun) { r.Event, r.HeadRepository.FullName = "workflow_run", "fork/repo" }, false, false, false, false, false},
		{"chained-other-workflow", func(r *WorkflowRun) { r.Event, r.Path = "workflow_run", ".github/workflows/other.yaml" }, false, false, false, false, false},
		{"chained-failed", func(r *WorkflowRun) { r.Event, r.Conclusion = "workflow_run", "failure" }, false, false, false, false, true},
		{"chained-active", func(r *WorkflowRun) { r.Event, r.Status, r.Conclusion = "workflow_run", "in_progress", "" }, false, false, false, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			run := WorkflowRun{ID: 17, HeadSHA: sha(1), HeadBranch: "main", Path: path, Status: "completed", Conclusion: "success"}
			run.Repository.FullName, run.HeadRepository.FullName = "owner/repo", "owner/repo"
			tc.mutate(&run)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(run)
			}))
			defer server.Close()
			client := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
			check := func(label string, want bool, err error) {
				t.Helper()
				if (err == nil) != want {
					t.Fatalf("%s: accepted=%v, want %v (%v)", label, err == nil, want, err)
				}
			}
			_, err := client.TrustedRun(t.Context(), 17, path)
			check("adapter run", tc.adapter, err)
			_, err = client.TrustedBuildRun(t.Context(), 17, Config{}, path)
			check("producer under the strict default", tc.strict, err)
			_, err = client.TrustedBuildRun(t.Context(), 17, chained, path)
			check("producer under the opt-in", tc.optIn, err)
			_, err = client.TrustedBuildRun(t.Context(), 17, chainedOnly, path)
			check("producer under a chained-only policy", tc.chainOnly, err)
			// Observation ignores completion; it must still refuse foreign runs.
			_, err = client.ObservedBuildRun(t.Context(), 17, chained, path)
			check("observed producer under the opt-in", tc.observed, err)
		})
	}
}

func TestRecoveryScanFollowsBuildEvents(t *testing.T) {
	build := func(id int64, event string, mutate func(*WorkflowRun)) WorkflowRun {
		run := WorkflowRun{ID: id, Path: ".github/workflows/build.yaml", HeadBranch: "main", Event: event, Status: "completed", Conclusion: "success"}
		run.Repository.FullName, run.HeadRepository.FullName = "owner/repo", "owner/repo"
		if mutate != nil {
			mutate(&run)
		}
		return run
	}
	runs := []WorkflowRun{
		build(1, "push", nil),
		build(2, "workflow_dispatch", nil),
		build(3, "workflow_run", nil),
		build(4, "pull_request", nil),
		build(5, "workflow_run", func(r *WorkflowRun) { r.HeadRepository.FullName = "fork/repo" }),
		build(6, "workflow_run", func(r *WorkflowRun) { r.Conclusion = "failure" }),
		build(7, "workflow_run", func(r *WorkflowRun) { r.HeadBranch = "feature" }),
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"workflow_runs": runs})
	}))
	defer server.Close()
	client := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
	ids := func(found []WorkflowRun) []int64 {
		out := []int64{}
		for _, run := range found {
			out = append(out, run.ID)
		}
		return out
	}
	cfg := Config{BuildWorkflow: ".github/workflows/build.yaml", DefaultBranch: "main"}
	strict, err := client.SuccessfulBuilds(t.Context(), cfg)
	if err != nil || !reflect.DeepEqual(ids(strict), []int64{1, 2}) {
		t.Fatal("default policy scan trusted a chained or foreign run", ids(strict), err)
	}
	cfg.BuildEvents = []string{"push", "workflow_dispatch", "workflow_run"}
	chained, err := client.SuccessfulBuilds(t.Context(), cfg)
	if err != nil || !reflect.DeepEqual(ids(chained), []int64{1, 2, 3}) {
		t.Fatal("opt-in scan lost a chained producer or trusted a foreign run", ids(chained), err)
	}
}

func TestWholePolicyRequestSourceBindsBuildEventsAndActivationVariable(t *testing.T) {
	target := func(policy string) string { return strings.Replace(policy, "target: v1.2.3", "target: v1.2.4", 1) }
	for _, field := range []string{chainedBuildEventsYAML, "activation_variable: MINT_OTHER_ENABLED\n"} {
		withField := environmentPolicyYAML + field
		cases := []struct {
			name, base, head, before string
			accepted                 bool
		}{
			{"unchanged-authority", withField, target(withField), withField, true},
			{"added-inside-request", environmentPolicyYAML, target(withField), environmentPolicyYAML, false},
			{"removed-inside-request", withField, target(environmentPolicyYAML), withField, false},
			// The reviewed PR itself is target-only, but the trusted policy at
			// its merge parent has other authority than its base.
			{"added-outside-request", environmentPolicyYAML, target(environmentPolicyYAML), withField, false},
			{"removed-outside-request", withField, target(withField), environmentPolicyYAML, false},
		}
		for _, tc := range cases {
			t.Run(strings.SplitN(field, ":", 2)[0]+"/"+tc.name, func(t *testing.T) {
				before, err := ParsePolicy([]byte(tc.before))
				if err != nil {
					t.Fatal(err)
				}
				pr := PullRequest{Number: 42}
				pr.Head.SHA = sha(3)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					switch r.URL.Path {
					case "/repos/owner/repo/pulls/42/commits":
						_ = json.NewEncoder(w).Encode([]any{map[string]string{"sha": sha(1)}, map[string]string{"sha": sha(3)}})
					case "/repos/owner/repo/git/commits/" + sha(1):
						_ = json.NewEncoder(w).Encode(map[string]any{"sha": sha(1), "parents": []any{map[string]string{"sha": sha(0)}}})
					case "/repos/owner/repo/contents/.mint.yaml":
						content := tc.base
						if r.URL.Query().Get("ref") == sha(3) {
							content = tc.head
						}
						_ = json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "content": base64.StdEncoding.EncodeToString([]byte(content))})
					default:
						t.Error("unexpected proof request", r.URL.Path)
						w.WriteHeader(500)
					}
				}))
				defer server.Close()
				client := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
				err = client.CheckPolicyRequestSource(context.Background(), pr, ".mint.yaml", before, "production")
				if (err == nil) != tc.accepted {
					t.Fatalf("accepted=%v, want %v (%v)", err == nil, tc.accepted, err)
				}
			})
		}
	}
}

func TestSchemaOneSourceAdapterKeepsFixedProducerEvents(t *testing.T) {
	cfg := Config{Schema: 1, Repository: "owner/repo", DefaultBranch: "main", BuildWorkflow: ".github/workflows/build.yaml", BuildEvents: []string{"push", "workflow_dispatch", "workflow_run"}}
	for _, event := range []string{"push", "workflow_run"} {
		t.Run(event, func(t *testing.T) {
			run := WorkflowRun{ID: 7, Path: cfg.BuildWorkflow, Event: event, Status: "in_progress", HeadBranch: "main"}
			run.Repository.FullName, run.HeadRepository.FullName = cfg.Repository, cfg.Repository
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(run)
			}))
			defer server.Close()
			err := (Client{APIURL: server.URL, Token: "test", Repository: cfg.Repository}).AuthorizeAdapterRun(context.Background(), cfg, 7, "source", event)
			if (err == nil) != (event == "push") {
				t.Fatal("schema 1 source adapter event set changed", event, err)
			}
		})
	}
}

func TestReconcileMarksDeferralBehindDeploymentInFlight(t *testing.T) {
	s, p := genericFixture("stage")
	register(t, &s, neutralCandidate(1))
	s.InFlight = "frozen-intent"
	_, err := s.Reconcile("normal", "candidate", "", "stage-generation", "", p)
	if !errors.Is(err, ErrInFlight) || !strings.Contains(err.Error(), "frozen-intent") {
		t.Fatal("in-flight deferral is not identifiable", err)
	}
	s.InFlight = ""
	if _, err := s.Reconcile("normal", "candidate", "", "stage-generation", "", p); err != nil || errors.Is(err, ErrInFlight) {
		t.Fatal("idle environment deferred", err)
	}
	// Only a deployment in flight is a deferral. A rollback pause and an intent
	// that is frozen but not settled need an operator and stay plain failures.
	paused, proof := genericFixture("stage")
	register(t, &paused, neutralCandidate(1))
	paused.Paused = true
	if _, err := paused.Reconcile("normal", "candidate", "", "stage-generation", "", proof); err == nil || errors.Is(err, ErrInFlight) {
		t.Fatal("rollback pause reported as a deferral", err)
	}
	for _, state := range []string{"merged", "deployment_failed", "publication_pending"} {
		frozen, proof := genericFixture("stage")
		register(t, &frozen, neutralCandidate(1))
		frozen.Proposals["normal"] = Proposal{ID: "stage-generation", Kind: "normal", State: state, Selection: "latest"}
		if _, err := frozen.Reconcile("normal", "candidate", "", "", "", proof); err == nil || errors.Is(err, ErrInFlight) {
			t.Fatal("unsettled intent reported as a deferral", state, err)
		}
	}
}
