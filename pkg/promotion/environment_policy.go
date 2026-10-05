package promotion

import (
	"fmt"
	"os"
)

// Policy is repository-owned build, publication and environment intent.
type Policy struct {
	DefaultEnvironment    string                       `yaml:"default_environment,omitempty"`
	Schema                int                          `yaml:"schema_version"`
	Mode                  string                       `yaml:"mode"`
	Repository            string                       `yaml:"repository"`
	DefaultBranch         string                       `yaml:"default_branch"`
	HumanLogin            string                       `yaml:"human_login,omitempty"`
	Authorization         string                       `yaml:"authorization,omitempty" json:"Authorization,omitempty"`
	Assignees             []string                     `yaml:"assignees,omitempty" json:"Assignees,omitempty"`
	BuildWorkflow         string                       `yaml:"build_workflow"`
	ValidationWorkflow    string                       `yaml:"validation_workflow,omitempty"`
	ControlWorkflow       string                       `yaml:"control_workflow,omitempty"`
	BaselineWorkflow      string                       `yaml:"baseline_workflow,omitempty"`
	BaselineBuildWorkflow string                       `yaml:"baseline_build_workflow,omitempty"`
	RequiredChecks        []string                     `yaml:"required_checks,omitempty"`
	ControlPaths          []string                     `yaml:"control_paths,omitempty"`
	Artifact              ArtifactPolicy               `yaml:"artifact,omitempty"`
	Environments          map[string]EnvironmentPolicy `yaml:"environments"`
}

// ArtifactPolicy identifies the optional project-owned artifact repository.
type ArtifactPolicy struct {
	Repository string `yaml:"repository"`
}

// EnvironmentPolicy is desired policy; it never asserts deployed runtime state.
type EnvironmentPolicy struct {
	Operation             string   `yaml:"operation,omitempty"`
	Reason                string   `yaml:"reason,omitempty"`
	Scope                 string   `yaml:"scope"`
	Deploy                string   `yaml:"deploy"`
	Follow                string   `yaml:"follow,omitempty"`
	Target                string   `yaml:"target,omitempty"`
	ImageTag              string   `yaml:"image_tag,omitempty"`
	Configuration         string   `yaml:"configuration_sha256"`
	PromotionWorkflow     string   `yaml:"promotion_workflow,omitempty"`
	ObservationWorkflow   string   `yaml:"observation_workflow,omitempty"`
	ValidationWorkflow    string   `yaml:"validation_workflow,omitempty"`
	BaselineWorkflow      string   `yaml:"baseline_workflow,omitempty"`
	BaselineBuildWorkflow string   `yaml:"baseline_build_workflow,omitempty"`
	RequiredChecks        []string `yaml:"required_checks,omitempty"`
	Requires              []string `yaml:"requires,omitempty"`
	Publish               bool     `yaml:"publish,omitempty"`
}

func LoadPolicy(path string) (Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, err
	}
	return ParsePolicy(data)
}

// ForEnvironment resolves a named policy without changing authenticated builds.
// An omitted name is unambiguous only when the policy has one environment.
func (p Policy) ForEnvironment(name string) (Config, error) {
	if name == "" && p.DefaultEnvironment != "" {
		name = p.DefaultEnvironment
	}
	if name == "" && len(p.Environments) == 1 {
		for environment := range p.Environments {
			name = environment
		}
	}
	env, ok := p.Environments[name]
	if !ok {
		return Config{}, fmt.Errorf("select a configured environment")
	}
	validation := env.ValidationWorkflow
	if validation == "" {
		validation = p.ValidationWorkflow
	}
	checks := env.RequiredChecks
	if len(checks) == 0 {
		checks = p.RequiredChecks
	}
	baseline := env.BaselineWorkflow
	if baseline == "" {
		baseline = p.BaselineWorkflow
	}
	baselineBuild := env.BaselineBuildWorkflow
	if baselineBuild == "" {
		baselineBuild = p.BaselineBuildWorkflow
	}
	if baselineBuild == "" {
		baselineBuild = p.BuildWorkflow
	}
	return Config{Operation: env.Operation, Reason: env.Reason, Schema: p.Schema, Mode: p.Mode, Repository: p.Repository, Environment: name, DefaultBranch: p.DefaultBranch, HumanLogin: p.HumanLogin, Authorization: p.Authorization, Assignees: cloneAssignees(p.Assignees), BuildWorkflow: p.BuildWorkflow, ValidationWorkflow: validation, ControlWorkflow: p.ControlWorkflow, BaselineWorkflow: baseline, BaselineBuildWorkflow: baselineBuild, PromotionWorkflow: env.PromotionWorkflow, ObservationWorkflow: env.ObservationWorkflow, RequiredChecks: append([]string(nil), checks...), ControlPaths: append([]string(nil), p.ControlPaths...), Scope: env.Scope, Deploy: env.Deploy, Follow: env.Follow, Target: env.Target, ImageTag: env.ImageTag, Configuration: env.Configuration, Requires: append([]string(nil), env.Requires...), Publish: env.Publish, ArtifactRepository: p.Artifact.Repository}, nil
}

func legacyPolicy(cfg Config) Policy {
	return Policy{Schema: 1, Mode: "deployment", Repository: cfg.Repository, DefaultBranch: cfg.DefaultBranch, HumanLogin: cfg.HumanLogin, Authorization: cfg.Authorization, Assignees: cloneAssignees(cfg.Assignees), BuildWorkflow: cfg.BuildWorkflow, ValidationWorkflow: cfg.ValidationWorkflow, ControlWorkflow: cfg.ControlWorkflow, BaselineWorkflow: cfg.BaselineWorkflow, BaselineBuildWorkflow: cfg.BaselineBuildWorkflow, RequiredChecks: cfg.RequiredChecks, ControlPaths: cfg.ControlPaths, Environments: map[string]EnvironmentPolicy{"production": {Scope: "shared", Deploy: "reviewed", Follow: "latest", PromotionWorkflow: cfg.PromotionWorkflow, Publish: true}}}
}
