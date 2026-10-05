package cli

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/jamesonstone/mint/pkg/promotion"
	"github.com/spf13/cobra"
)

func writeDeploymentJSON(cmd *cobra.Command, value any, output string) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if output != "" {
		return os.WriteFile(output, append(data, '\n'), 0600)
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return err
}

func (o *productionOperation) environmentStatus() any {
	s := o.snapshot.State
	var desired *promotion.Candidate
	c, err := s.SelectCandidate("normal", "", o.proof)
	if err == nil {
		desired = &c
	}
	return map[string]any{"policy": o.config, "state": promotion.ProjectEnvironmentStatus(s, desired, o.config.Configuration), "paused": s.Paused, "target": s.Target}
}

// Preconditions use separately loaded verified history. A live observation is
// never substituted for approval, a baseline or successful deployment history.
func (o *productionOperation) checkPrerequisites(ctx context.Context, operation string) error {
	cfg := o.config
	if cfg.Schema != 2 {
		return nil
	}
	switch operation {
	case "propose", "propose-rollback", "rollback", "hotfix", "bootstrap", "start", "intent", "policy-request":
		if cfg.Configuration == "" || cfg.PromotionWorkflow == "" || cfg.ValidationWorkflow == "" || len(cfg.RequiredChecks) == 0 {
			return fmt.Errorf("environment is not activated: configure its runtime identity and trusted promotion adapter, then import a verified baseline")
		}
	default:
		return nil
	}
	if operation != "start" {
		return nil
	}
	intent, ok := o.snapshot.State.Intents[o.flags.IntentID]
	if !ok {
		return fmt.Errorf("unknown deployment intent")
	}
	authority, err := promotion.AuthorityDigest(cfg)
	if err != nil {
		return err
	}
	if intent.PolicyDigest != authority {
		return fmt.Errorf("environment authority changed after review; request a new deployment")
	}
	if intent.Configuration != cfg.Configuration {
		return fmt.Errorf("runtime policy changed after review; request a new deployment")
	}
	for _, name := range cfg.Requires {
		upstream, err := o.client.LoadJournal(ctx, name)
		if err != nil {
			return err
		}
		if upstream.State.Baseline == nil || upstream.State.InFlight != "" || !reflect.DeepEqual(upstream.State.Baseline.Candidate.Artifact, intent.Candidate.Artifact) {
			return fmt.Errorf("environment %s has not verified this exact artifact set", name)
		}
	}
	return nil
}

// Local observations are explicitly operator-supplied machine evidence. They
// never write a shared branch or claim an authenticated server deployment.
func runLocalEnvironment(cmd *cobra.Command, operation string, cfg promotion.Config, f productionFlags) error {
	cache, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(cache, "mint", fmt.Sprintf("%x", sha256.Sum256([]byte(cfg.Repository))))
	path := filepath.Join(dir, cfg.Environment+".json")
	state := promotion.NewState(cfg.Repository, cfg.Environment)
	state.Schema = 2
	if data, readErr := os.ReadFile(path); readErr == nil {
		state, err = promotion.DecodeState(data, cfg.Repository, cfg.Environment)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(readErr) {
		return readErr
	}
	if operation == "observe" {
		var obs promotion.Observation
		if err := readJSON(f.Input, &obs); err != nil {
			return err
		}
		if obs.Repository != cfg.Repository || obs.Environment != cfg.Environment {
			return fmt.Errorf("foreign local observation")
		}
		changed, err := promotion.ApplyObservation(&state, obs)
		if err != nil {
			return err
		}
		if changed {
			if err := os.MkdirAll(dir, 0700); err != nil {
				return err
			}
			data, err := json.MarshalIndent(state, "", "  ")
			if err != nil {
				return err
			}
			file, err := os.CreateTemp(dir, ".observation-*")
			if err != nil {
				return err
			}
			defer func() { _ = os.Remove(file.Name()) }()
			if _, err := file.Write(data); err != nil {
				_ = file.Close()
				return err
			}
			if err := file.Close(); err != nil {
				return err
			}
			if err := os.Rename(file.Name(), path); err != nil {
				return err
			}
		}
	} else if operation != "status" {
		return fmt.Errorf("local scope supports status and machine observation; local deployment remains your project's adapter")
	}
	return writeDeploymentJSON(cmd, map[string]any{"scope": "local", "policy": cfg, "state": promotion.ProjectEnvironmentStatus(state, nil, cfg.Configuration)}, f.Output)
}
