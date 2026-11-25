package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	initConfigPath    = ".mint.yaml"
	initDeployPath    = ".github/workflows/mint-deploy.yml"
	initPromotionPath = ".github/workflows/mint-promote.yml"
)

func init() {
	rootCmd.AddCommand(newInitCommand())
}

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Guide setup of environment deployment workflows",
		Long:  "Run from the repository root. Choose an artifact type and deployment mode for each environment, then review the configuration and workflow files Mint will create.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runGuidedInit(cmd.InOrStdin(), cmd.OutOrStdout())
		},
	}
}

func runGuidedInit(input io.Reader, output io.Writer) error {
	files := []initFile{
		{path: initConfigPath},
		{path: initDeployPath},
		{path: initPromotionPath},
	}
	if err := checkInitPaths(files); err != nil {
		return err
	}
	prompt := initPrompt{input: bufio.NewScanner(input), output: output}
	config, err := prompt.pipeline()
	if err != nil {
		return err
	}
	confirmed, err := prompt.confirm(config, files)
	if err != nil {
		return err
	}
	if !confirmed {
		_, err := fmt.Fprintln(output, "Setup cancelled.")
		return err
	}

	files[0].content, err = yaml.Marshal(config)
	if err != nil {
		return err
	}
	files[1].content, files[2].content, err = renderInitWorkflows(config.Environments)
	if err != nil {
		return err
	}
	if err := installInitFiles(files); err != nil {
		return err
	}
	_, err = fmt.Fprintln(output, "Created the configuration and workflows. Implement the project hooks in .github/mint and review GitHub environment protections before enabling deployments.")
	return err
}

type initPrompt struct {
	input  *bufio.Scanner
	output io.Writer
}

func (prompt initPrompt) line(label, fallback string, validate func(string) error) (string, error) {
	for {
		if _, err := fmt.Fprintf(prompt.output, "%s [%s]: ", label, fallback); err != nil {
			return "", err
		}
		if !prompt.input.Scan() {
			if err := prompt.input.Err(); err != nil {
				return "", fmt.Errorf("read setup input: %w", err)
			}
			return "", fmt.Errorf("setup input ended before confirmation")
		}
		value := strings.TrimSpace(prompt.input.Text())
		if value == "" {
			value = fallback
		}
		if err := validate(value); err != nil {
			if _, writeErr := fmt.Fprintln(prompt.output, err); writeErr != nil {
				return "", writeErr
			}
			continue
		}
		return value, nil
	}
}

func (prompt initPrompt) pipeline() (pipelineConfig, error) {
	artifact, err := prompt.line("Artifact type (binary/image)", "binary", func(value string) error {
		if !validArtifact(value) {
			return fmt.Errorf("enter binary or image")
		}
		return nil
	})
	if err != nil {
		return pipelineConfig{}, err
	}
	rawNames, err := prompt.line("Environments in promotion order (comma-separated)", "dev,stg,prd", func(value string) error {
		return validateEnvironmentNames(splitEnvironmentNames(value))
	})
	if err != nil {
		return pipelineConfig{}, err
	}
	config := pipelineConfig{Artifact: artifact}
	for _, name := range splitEnvironmentNames(rawNames) {
		fallback := "continuous"
		switch name {
		case "prd", "prod", "production":
			fallback = "manual"
		}
		mode, err := prompt.line(fmt.Sprintf("Deployment for %s (continuous/manual)", name), fallback, func(value string) error {
			if !validDeployment(value) {
				return fmt.Errorf("enter continuous or manual")
			}
			return nil
		})
		if err != nil {
			return pipelineConfig{}, err
		}
		config.Environments = append(config.Environments, pipelineStage{Name: name, Deployment: mode})
	}
	return config, nil
}

func (prompt initPrompt) confirm(config pipelineConfig, files []initFile) (bool, error) {
	var summary strings.Builder
	fmt.Fprintf(&summary, "\nArtifact: %s\n", config.Artifact)
	for _, stage := range config.Environments {
		fmt.Fprintf(&summary, "  %s: %s\n", stage.Name, stage.Deployment)
	}
	fmt.Fprintln(&summary, "Files:")
	for _, file := range files {
		fmt.Fprintf(&summary, "  %s\n", file.path)
	}
	if _, err := io.WriteString(prompt.output, summary.String()); err != nil {
		return false, err
	}
	confirm, err := prompt.line("Write these files? (y/n)", "n", func(value string) error {
		switch strings.ToLower(value) {
		case "y", "yes", "n", "no":
			return nil
		default:
			return fmt.Errorf("enter y or n")
		}
	})
	return strings.EqualFold(confirm, "y") || strings.EqualFold(confirm, "yes"), err
}

func splitEnvironmentNames(value string) []string {
	names := strings.Split(value, ",")
	for index, name := range names {
		names[index] = strings.TrimSpace(name)
	}
	return names
}
