package promotion

import (
	"context"
	"fmt"
	"time"
)

// ReadObservation authenticates a unique digest-attested observation manifest
// from the configured successful default-branch observer. Observing a runtime
// version does not imply the observer workflow checked out that runtime source.
func (c Client) ReadObservation(ctx context.Context, cfg Config, runID int64) (Observation, error) {
	var observation Observation
	if cfg.Mode != "" && cfg.Mode != "deployment" {
		return observation, fmt.Errorf("artifact and package policies do not have running deployment observations")
	}
	if cfg.ObservationWorkflow == "" || cfg.Repository != c.Repository {
		return observation, fmt.Errorf("observation requires a repository-owned trusted observer workflow")
	}
	run, err := c.TrustedRun(ctx, runID, cfg.ObservationWorkflow)
	if err != nil {
		return observation, err
	}
	if run.HeadBranch != cfg.DefaultBranch {
		return observation, fmt.Errorf("observation workflow must run from the trusted default branch")
	}
	if err := c.RunManifest(ctx, runID, "mint-observation", &observation); err != nil {
		return observation, err
	}
	if observation.Repository != cfg.Repository || observation.Environment != cfg.Environment || observation.RunID != run.ID || observation.RunURL != run.HTMLURL {
		return observation, fmt.Errorf("observation manifest differs from its repository, environment or authoritative workflow identity")
	}
	observed, err := validateObservation(observation, time.Now().UTC())
	if err != nil {
		return observation, err
	}
	created, createdErr := time.Parse(time.RFC3339Nano, run.CreatedAt)
	updated, updatedErr := time.Parse(time.RFC3339Nano, run.UpdatedAt)
	if createdErr != nil || updatedErr != nil || updated.Before(created) || observed.Before(created) || observed.After(updated) {
		return observation, fmt.Errorf("observation timestamp is outside its authoritative workflow interval")
	}
	return observation, nil
}
