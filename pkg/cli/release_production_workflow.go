package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func writeControlWorkflow(cmd *cobra.Command, cfg promotion.Config, f productionFlags) error {
	text, err := promotion.RenderControlWorkflow(cfg, f.MintRef, f.Config)
	if cfg.Schema == 2 {
		policy, loadErr := promotion.LoadPolicy(f.Config)
		if loadErr != nil {
			return loadErr
		}
		paths := []string{cfg.BuildWorkflow}
		for _, env := range policy.Environments {
			for _, path := range []string{env.PromotionWorkflow, env.ObservationWorkflow} {
				if path != "" {
					paths = append(paths, path)
				}
			}
		}
		names := []string{}
		seen := map[string]bool{}
		for _, path := range paths {
			if seen[path] {
				continue
			}
			seen[path] = true
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return fmt.Errorf("read callback workflow name: %w", readErr)
			}
			var workflow struct {
				Name string `yaml:"name"`
			}
			if err := yaml.Unmarshal(data, &workflow); err != nil {
				return err
			}
			if workflow.Name == "" {
				workflow.Name = path
			}
			names = append(names, workflow.Name)
		}
		text, err = promotion.RenderEnvironmentWorkflow(policy, f.MintRef, f.Config, names...)
	}
	if err != nil {
		return err
	}
	if f.Output != "" {
		if !promotion.SafeRepositoryPath(f.Output) || f.Output != cfg.ControlWorkflowPath() {
			return fmt.Errorf("workflow output must match configured control_workflow %s", cfg.ControlWorkflowPath())
		}
		if err := os.MkdirAll(filepath.Dir(f.Output), 0755); err != nil {
			return err
		}
		return os.WriteFile(f.Output, []byte(text), 0600)
	}
	_, err = fmt.Fprint(cmd.OutOrStdout(), text)
	return err
}
