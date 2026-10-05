package promotion

import (
	"strings"
	"testing"
)

const environmentPolicyYAML = `schema_version: 2
mode: deployment
repository: owner/repo
default_branch: main
human_login: jamesonstone
build_workflow: .github/workflows/build.yaml
validation_workflow: .github/workflows/checks.yaml
required_checks: [native]
artifact:
  repository: ghcr.io/owner/repo
environments:
  dev:
    scope: shared
    deploy: automatic
    follow: latest
    image_tag: dev
  production:
    scope: shared
    deploy: reviewed
    target: v1.2.3
    image_tag: stable
    requires: [dev]
    promotion_workflow: .github/workflows/deploy.yaml
    observation_workflow: .github/workflows/observe.yaml
    configuration_sha256: sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
    publish: true
`

func TestEnvironmentPolicyResolvesAndPreservesExplicitIntent(t *testing.T) {
	p, err := ParsePolicy([]byte(environmentPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := p.ForEnvironment("production")
	if err != nil || cfg.Schema != 2 || cfg.Mode != "deployment" || cfg.Environment != "production" || cfg.Target != "v1.2.3" || cfg.Follow != "" || cfg.Deploy != "reviewed" || cfg.Scope != "shared" || !cfg.Publish || cfg.ArtifactRepository != "ghcr.io/owner/repo" || cfg.Operation != "release" || cfg.Configuration != "sha256:"+strings.Repeat("a", 64) || len(cfg.Requires) != 1 {
		t.Fatal(cfg, err)
	}
	if cfg.ValidationWorkflow != ".github/workflows/checks.yaml" || cfg.BaselineBuildWorkflow != cfg.BuildWorkflow {
		t.Fatal("shared defaults not inherited", cfg)
	}
	cfg.RequiredChecks[0] = "changed"
	cfg.ControlPaths[0] = "changed"
	cfg.Requires[0] = "changed"
	if p.RequiredChecks[0] != "native" || p.ControlPaths[0] != "CHANGELOG.md" || p.Environments["production"].Requires[0] != "dev" {
		t.Fatal("effective policy aliases mutable policy slices")
	}
	dev, err := p.ForEnvironment("dev")
	if err != nil || dev.Configuration != "" || dev.PromotionWorkflow != "" || dev.Follow != "latest" {
		t.Fatal("unactivated environment rejected", dev, err)
	}
	if _, err := p.ForEnvironment(""); err == nil {
		t.Fatal("ambiguous omitted environment accepted")
	}
	if _, err := p.ForEnvironment("missing"); err == nil {
		t.Fatal("unknown environment accepted")
	}
}
func TestNonRuntimePoliciesNeedNoRuntimeOrMigrationWorkflow(t *testing.T) {
	for _, mode := range []string{"artifact", "package"} {
		t.Run(mode, func(t *testing.T) {
			data := `schema_version: 2
mode: ` + mode + `
repository: owner/repo
default_branch: trunk
human_login: jamesonstone
build_workflow: .github/workflows/release.yaml
environments:
  distribution:
    scope: shared
    deploy: manual
    follow: latest
    publish: true
`
			p, err := ParsePolicy([]byte(data))
			if err != nil {
				t.Fatal(err)
			}
			cfg, err := p.ForEnvironment("")
			if err != nil || cfg.Mode != mode || cfg.Environment != "distribution" || cfg.PromotionWorkflow != "" || cfg.ObservationWorkflow != "" {
				t.Fatal(cfg, err)
			}
		})
	}
}
func TestEnvironmentPolicyRejectsAmbiguityAndUnsafePolicy(t *testing.T) {
	cases := map[string]string{
		"unknown":               environmentPolicyYAML + "surprise: true\n",
		"duplicate":             environmentPolicyYAML + "mode: deployment\n",
		"duplicate-environment": strings.Replace(environmentPolicyYAML, "  production:", "  dev:", 1),
		"unknown-env":           strings.Replace(environmentPolicyYAML, "    image_tag: dev", "    surprise: true", 1),
		"alias":                 strings.Replace(environmentPolicyYAML, "    requires: [dev]", "    requires: &deps [dev]", 1),
		"merge":                 environmentPolicyYAML + "<<: {}\n",
		"multidoc":              environmentPolicyYAML + "---\nschema_version: 2\n",
		"scalar":                "hello",
		"bad-schema":            strings.Replace(environmentPolicyYAML, "schema_version: 2", "schema_version: 3", 1),
		"mode":                  strings.Replace(environmentPolicyYAML, "mode: deployment", "mode: invented", 1),
		"unsafe-environment":    strings.Replace(environmentPolicyYAML, "  dev:", "  ../dev:", 1),
		"missing-scope":         strings.Replace(environmentPolicyYAML, "    scope: shared\n", "", 1),
		"unknown-deploy":        strings.Replace(environmentPolicyYAML, "deploy: automatic", "deploy: yes", 1),
		"xor":                   strings.Replace(environmentPolicyYAML, "    follow: latest", "    follow: latest\n    target: v1.2.3", 1),
		"missing-selection":     strings.Replace(environmentPolicyYAML, "    follow: latest\n", "", 1),
		"mutable-target":        strings.Replace(environmentPolicyYAML, "target: v1.2.3", "target: main", 1),
		"digest":                strings.Replace(environmentPolicyYAML, "sha256:"+strings.Repeat("a", 64), "sha256:unknown", 1),
		"tag-collision":         strings.Replace(environmentPolicyYAML, "image_tag: stable", "image_tag: dev", 1),
		"immutable-alias":       strings.Replace(environmentPolicyYAML, "image_tag: stable", "image_tag: v1.2.3", 1),
		"missing-requirement":   strings.Replace(environmentPolicyYAML, "requires: [dev]", "requires: [missing]", 1),
		"duplicate-requirement": strings.Replace(environmentPolicyYAML, "requires: [dev]", "requires: [dev, dev]", 1),
		"cycle":                 strings.Replace(environmentPolicyYAML, "    image_tag: dev", "    image_tag: dev\n    requires: [production]", 1),
		"local-upstream":        strings.Replace(environmentPolicyYAML, "scope: shared", "scope: local", 1),
		"workflow":              strings.Replace(environmentPolicyYAML, ".github/workflows/deploy.yaml", ".github/workflows/nested/deploy.yaml", 1),
		"checks":                strings.Replace(environmentPolicyYAML, "required_checks: [native]", "required_checks: [native, native]", 1),
		"empty-controlpaths":    environmentPolicyYAML + "control_paths: []\n",
		"unsafe-controlpaths":   environmentPolicyYAML + "control_paths: [CHANGELOG.md, .mint/proposal.json, .mint/summary.md, app.go]\n",
		"recovery-reason":       strings.Replace(environmentPolicyYAML, "    target: v1.2.3", "    target: v1.2.3\n    operation: rollback", 1),
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := ParsePolicy([]byte(data)); err == nil {
				t.Fatal("invalid policy accepted")
			}
		})
	}
}
func TestEnvironmentRecoveryOperationAndOverrides(t *testing.T) {
	data := strings.Replace(environmentPolicyYAML, "    target: v1.2.3", "    target: v1.2.3\n    operation: rollback\n    reason: Restore working version\n    validation_workflow: .github/workflows/recovery-checks.yaml\n    required_checks: [recovery]", 1)
	p, err := ParsePolicy([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := p.ForEnvironment("production")
	if err != nil || cfg.Operation != "rollback" || cfg.Reason != "Restore working version" || cfg.ValidationWorkflow != ".github/workflows/recovery-checks.yaml" || cfg.RequiredChecks[0] != "recovery" {
		t.Fatal(cfg, err)
	}
}
func TestLegacyPolicyKeepsProductionBehavior(t *testing.T) {
	data := `schema_version: 1
repository: owner/repo
environment: production
default_branch: main
human_login: jamesonstone
build_workflow: .github/workflows/build.yaml
validation_workflow: .github/workflows/checks.yaml
promotion_workflow: .github/workflows/deploy.yaml
required_checks: [native]
`
	p, err := ParsePolicy([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := p.ForEnvironment("")
	if err != nil || cfg.Schema != 1 || cfg.Environment != "production" || cfg.Mode != "deployment" || cfg.Deploy != "reviewed" || cfg.Follow != "latest" || !cfg.Publish {
		t.Fatal(cfg, err)
	}
	if _, err := ParsePolicy([]byte(data + "unexpected: true\n")); err == nil {
		t.Fatal("legacy unknown field accepted")
	}
}

func TestExplicitDefaultEnvironmentDoesNotInferProduction(t *testing.T) {
	p, err := ParsePolicy([]byte(environmentPolicyYAML + "default_environment: dev\n"))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := p.ForEnvironment("")
	if err != nil || cfg.Environment != "dev" {
		t.Fatal(cfg, err)
	}
	if _, err := ParsePolicy([]byte(environmentPolicyYAML + "default_environment: missing\n")); err == nil {
		t.Fatal("unknown default environment accepted")
	}
}
