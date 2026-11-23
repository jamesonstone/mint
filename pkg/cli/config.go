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
	Artifact     string   `yaml:"artifact" json:"artifact"`
	Deployment   string   `yaml:"deployment" json:"deployment"`
	Environments []string `yaml:"environments" json:"environments"`
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
	switch cfg.Artifact {
	case "binary", "image":
	default:
		return cfg, fmt.Errorf("artifact must be binary or image")
	}
	switch cfg.Deployment {
	case "continuous", "manual":
	default:
		return cfg, fmt.Errorf("deployment must be continuous or manual")
	}
	if len(cfg.Environments) == 0 {
		return cfg, fmt.Errorf("environments must name at least one stage")
	}
	seen := make(map[string]bool, len(cfg.Environments))
	for _, name := range cfg.Environments {
		if !environmentName.MatchString(name) {
			return cfg, fmt.Errorf("invalid environment name %q", name)
		}
		if seen[name] {
			return cfg, fmt.Errorf("duplicate environment %q", name)
		}
		seen[name] = true
	}
	return cfg, nil
}
