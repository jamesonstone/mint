package promotion

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidEnvironmentName permits one safe, case-sensitive ref suffix.
func ValidEnvironmentName(name string) bool {
	return regexp.MustCompile(`^[a-z][a-z0-9_-]{0,62}$`).MatchString(name)
}

func validateEnvironment(env EnvironmentPolicy) error {
	if env.Scope != "local" && env.Scope != "shared" {
		return fmt.Errorf("scope must be local or shared")
	}
	if env.Deploy != "manual" && env.Deploy != "automatic" && env.Deploy != "reviewed" {
		return fmt.Errorf("deploy must be manual, automatic or reviewed")
	}
	if (env.Follow != "") == (env.Target != "") || (env.Follow != "" && env.Follow != "latest") || (env.Target != "" && !versionPattern.MatchString(env.Target)) {
		return fmt.Errorf("select follow: latest or one strict target version")
	}
	if env.Configuration != "" && !digestPattern.MatchString(env.Configuration) {
		return fmt.Errorf("configuration_sha256 must be an exact SHA-256 digest")
	}
	if env.ImageTag != "" && (!regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`).MatchString(env.ImageTag) || versionPattern.MatchString(env.ImageTag)) {
		return fmt.Errorf("image_tag must be a safe alias distinct from immutable version tags")
	}
	for _, workflow := range []string{env.PromotionWorkflow, env.ObservationWorkflow, env.ValidationWorkflow, env.BaselineWorkflow, env.BaselineBuildWorkflow} {
		if workflow != "" && !ValidWorkflowPath(workflow) {
			return fmt.Errorf("environment workflow must be directly under .github/workflows")
		}
	}
	switch env.Operation {
	case "", "release", "resume":
	case "hotfix", "rollback":
		if strings.TrimSpace(env.Reason) == "" {
			return fmt.Errorf("recovery operation requires a reason")
		}
	default:
		return fmt.Errorf("operation must be release, hotfix, rollback or resume")
	}
	if err := validateCheckNames(env.RequiredChecks); err != nil {
		return err
	}
	return nil
}
func validateCheckNames(checks []string) error {
	seen := map[string]bool{}
	for _, check := range checks {
		if strings.TrimSpace(check) == "" || seen[check] {
			return fmt.Errorf("required checks must be nonempty and unique")
		}
		seen[check] = true
	}
	return nil
}
func validateControlPaths(paths []string) error {
	seen := map[string]bool{}
	for _, path := range paths {
		if path != "CHANGELOG.md" && path != ".mint/proposal.json" && path != ".mint/summary.md" && path != ".mint/status.json" {
			return fmt.Errorf("unsafe release-control path %s", path)
		}
		if seen[path] {
			return fmt.Errorf("duplicate release-control path %s", path)
		}
		seen[path] = true
	}
	for _, required := range []string{"CHANGELOG.md", ".mint/proposal.json", ".mint/summary.md"} {
		if !seen[required] {
			return fmt.Errorf("release-control paths must include %s", required)
		}
	}
	return nil
}
func (p Policy) validateDependencies() error {
	visiting, visited := map[string]bool{}, map[string]bool{}
	var visit func(string) error
	visit = func(name string) error {
		if visiting[name] {
			return fmt.Errorf("environment dependency cycle at %s", name)
		}
		if visited[name] {
			return nil
		}
		visiting[name] = true
		seen := map[string]bool{}
		for _, required := range p.Environments[name].Requires {
			upstream, ok := p.Environments[required]
			if !ok || seen[required] {
				return fmt.Errorf("environment %s has missing or duplicate prerequisite %s", name, required)
			}
			seen[required] = true
			if p.Environments[name].Scope == "shared" && upstream.Scope == "local" {
				return fmt.Errorf("shared environment cannot depend on machine-local evidence")
			}
			if err := visit(required); err != nil {
				return err
			}
		}
		visiting[name] = false
		visited[name] = true
		return nil
	}
	for name := range p.Environments {
		if err := visit(name); err != nil {
			return err
		}
	}
	return nil
}
