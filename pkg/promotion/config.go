package promotion

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"regexp"
)

// Config is repository-owned release policy; issue content cannot override it.
type Config struct {
	Schema             int      `yaml:"schema_version"`
	Repository         string   `yaml:"repository"`
	Environment        string   `yaml:"environment"`
	DefaultBranch      string   `yaml:"default_branch"`
	HumanLogin         string   `yaml:"human_login"`
	HumanName          string   `yaml:"human_name"`
	HumanEmail         string   `yaml:"human_email"`
	BuildWorkflow      string   `yaml:"build_workflow"`
	ValidationWorkflow string   `yaml:"validation_workflow"`
	PromotionWorkflow  string   `yaml:"promotion_workflow"`
	RequiredChecks     []string `yaml:"required_checks"`
	ControlPaths       []string `yaml:"control_paths"`
}

// LoadConfig rejects unknown fields and unscoped release policies.
func LoadConfig(path string) (Config, error) {
	var cfg Config
	file, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer func() { _ = file.Close() }()
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, err
	}
	if cfg.Schema != 1 || !regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`).MatchString(cfg.Repository) || cfg.Environment != "production" || cfg.DefaultBranch == "" || cfg.HumanLogin == "" || cfg.HumanName == "" || cfg.HumanEmail == "" || cfg.BuildWorkflow == "" || cfg.ValidationWorkflow == "" || cfg.PromotionWorkflow == "" || len(cfg.RequiredChecks) == 0 {
		return cfg, fmt.Errorf("incomplete production release policy")
	}
	for _, p := range cfg.ControlPaths {
		if p != "CHANGELOG.md" && p != ".mint/proposal.json" && p != ".mint/summary.md" && p != ".mint/status.json" {
			return cfg, fmt.Errorf("unsafe release-control path %s", p)
		}
	}
	return cfg, nil
}
