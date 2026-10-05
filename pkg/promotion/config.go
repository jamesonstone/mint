package promotion

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"regexp"
)

// Config is repository-owned release policy; issue content cannot override it.
type Config struct {
	PolicyPath            string   `yaml:"-" json:"-"`
	Operation             string   `yaml:"-"`
	Reason                string   `yaml:"-"`
	Mode                  string   `yaml:"-"`
	Scope                 string   `yaml:"-"`
	Deploy                string   `yaml:"-"`
	Follow                string   `yaml:"-"`
	Target                string   `yaml:"-"`
	Configuration         string   `yaml:"-"`
	ObservationWorkflow   string   `yaml:"-"`
	Requires              []string `yaml:"-"`
	Publish               bool     `yaml:"-"`
	ArtifactRepository    string   `yaml:"-"`
	ImageTag              string   `yaml:"-"`
	ControlWorkflow       string   `yaml:"control_workflow"`
	Schema                int      `yaml:"schema_version"`
	Repository            string   `yaml:"repository"`
	Environment           string   `yaml:"environment"`
	DefaultBranch         string   `yaml:"default_branch"`
	HumanLogin            string   `yaml:"human_login"`
	BuildWorkflow         string   `yaml:"build_workflow"`
	ValidationWorkflow    string   `yaml:"validation_workflow"`
	BaselineWorkflow      string   `yaml:"baseline_workflow"`
	BaselineBuildWorkflow string   `yaml:"baseline_build_workflow"`
	PromotionWorkflow     string   `yaml:"promotion_workflow"`
	RequiredChecks        []string `yaml:"required_checks"`
	ControlPaths          []string `yaml:"control_paths"`
}

// LoadConfig rejects unknown fields and unscoped release policies.
func LoadConfig(path string) (Config, error) {
	policy, err := LoadPolicy(path)
	if err != nil {
		return Config{}, err
	}
	return policy.ForEnvironment("")
}

func parseLegacyConfig(data []byte) (Config, error) {
	var cfg Config

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return cfg, fmt.Errorf("release policy must contain exactly one YAML document")
	}
	if cfg.Schema != 1 || !regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`).MatchString(cfg.Repository) || cfg.Environment != "production" || cfg.DefaultBranch == "" || cfg.HumanLogin == "" || cfg.BuildWorkflow == "" || cfg.ValidationWorkflow == "" || cfg.PromotionWorkflow == "" || len(cfg.RequiredChecks) == 0 {
		return cfg, fmt.Errorf("incomplete production release policy")
	}
	if !ValidWorkflowPath(cfg.ControlWorkflowPath()) {
		return cfg, fmt.Errorf("control_workflow must be a YAML file directly under .github/workflows")
	}
	if cfg.BaselineBuildWorkflow == "" {
		cfg.BaselineBuildWorkflow = cfg.BuildWorkflow
	}
	var document yaml.Node
	if err := yaml.Unmarshal(data, &document); err != nil {
		return cfg, err
	}
	explicitPaths := false
	for i := 0; i+1 < len(document.Content[0].Content); i += 2 {
		if document.Content[0].Content[i].Value == "<<" {
			return cfg, fmt.Errorf("release policy must use explicit fields rather than YAML merge keys")
		}
		if document.Content[0].Content[i].Value == "control_paths" {
			explicitPaths = true
		}
	}
	if !explicitPaths {
		cfg.ControlPaths = []string{"CHANGELOG.md", ".mint/proposal.json", ".mint/summary.md"}
	}
	seen := map[string]bool{}
	for _, p := range cfg.ControlPaths {
		if p != "CHANGELOG.md" && p != ".mint/proposal.json" && p != ".mint/summary.md" && p != ".mint/status.json" {
			return cfg, fmt.Errorf("unsafe release-control path %s", p)
		}
		if seen[p] {
			return cfg, fmt.Errorf("duplicate release-control path %s", p)
		}
		seen[p] = true
	}
	for _, required := range []string{"CHANGELOG.md", ".mint/proposal.json", ".mint/summary.md"} {
		if !seen[required] {
			return cfg, fmt.Errorf("release-control paths must include %s", required)
		}
	}
	return cfg, nil
}

// ValidateBootstrap requires the trusted baseline verifier only when importing
// an initial deployment. Ordinary commands need no legacy migration workflow.
func (cfg Config) ValidateBootstrap() error {
	if cfg.BaselineWorkflow == "" || cfg.BaselineBuildWorkflow == "" {
		return fmt.Errorf("bootstrap requires a trusted baseline workflow and producer workflow")
	}
	return nil
}
