package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type pipelineConfig struct {
	Artifact     string          `yaml:"artifact" json:"artifact"`
	Deployment   string          `yaml:"deployment,omitempty" json:"deployment,omitempty"`
	Environments []pipelineStage `yaml:"environments" json:"environments"`
}

type pipelineStage struct {
	Name       string `yaml:"name" json:"name"`
	Deployment string `yaml:"deployment" json:"deployment"`
}

func (stage *pipelineStage) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		return node.Decode(&stage.Name)
	}
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("each environment must be a name or a name and deployment mapping")
	}
	for index := 0; index < len(node.Content); index += 2 {
		key := node.Content[index].Value
		if key != "name" && key != "deployment" {
			return fmt.Errorf("unknown environment field %q", key)
		}
	}
	type stageFields pipelineStage
	if err := node.Decode((*stageFields)(stage)); err != nil {
		return err
	}
	if !validDeployment(stage.Deployment) {
		return fmt.Errorf("environment %q deployment must be continuous or manual", stage.Name)
	}
	return nil
}

func (stage pipelineStage) MarshalJSON() ([]byte, error) {
	if stage.Deployment == "" {
		return json.Marshal(stage.Name)
	}
	type stageFields pipelineStage
	return json.Marshal(stageFields(stage))
}

var environmentName = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

func init() {
	rootCmd.AddCommand(newConfigCommand())
}

func newConfigCommand() *cobra.Command {
	var file string
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Validate and print the release pipeline configuration as JSON",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			input, err := os.Open(file)
			if err != nil {
				return err
			}
			defer input.Close()
			cfg, err := readPipelineConfig(input)
			if err != nil {
				return fmt.Errorf("%s: %w", file, err)
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(cfg)
		},
	}
	cmd.Flags().StringVar(&file, "file", ".mint.yaml", "pipeline configuration file")
	return cmd
}

func readPipelineConfig(input io.Reader) (pipelineConfig, error) {
	var cfg pipelineConfig
	decoder := yaml.NewDecoder(input)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("read pipeline configuration: %w", err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return cfg, fmt.Errorf("read pipeline configuration: %w", err)
		}
		return cfg, fmt.Errorf("expected one YAML document")
	}
	if !validArtifact(cfg.Artifact) {
		return pipelineConfig{}, fmt.Errorf("artifact must be binary or image")
	}
	if cfg.Deployment != "" && !validDeployment(cfg.Deployment) {
		return pipelineConfig{}, fmt.Errorf("deployment must be continuous or manual")
	}
	if len(cfg.Environments) == 0 {
		return pipelineConfig{}, fmt.Errorf("environments must name at least one stage")
	}
	names := make([]string, 0, len(cfg.Environments))
	for _, stage := range cfg.Environments {
		if stage.Deployment == "" && !validDeployment(cfg.Deployment) {
			return pipelineConfig{}, fmt.Errorf("deployment must be continuous or manual for a list of environment names")
		}
		if stage.Deployment != "" && cfg.Deployment != "" {
			return pipelineConfig{}, fmt.Errorf("use either a top-level deployment mode or a mode per environment")
		}
		names = append(names, stage.Name)
	}
	if err := validateEnvironmentNames(names); err != nil {
		return pipelineConfig{}, err
	}
	return cfg, nil
}

func validArtifact(artifact string) bool {
	return artifact == "binary" || artifact == "image"
}

func validDeployment(deployment string) bool {
	return deployment == "continuous" || deployment == "manual"
}

func validateEnvironmentNames(names []string) error {
	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if !environmentName.MatchString(name) {
			return fmt.Errorf("invalid environment name %q", name)
		}
		if seen[name] {
			return fmt.Errorf("duplicate environment %q", name)
		}
		seen[name] = true
	}
	return nil
}
