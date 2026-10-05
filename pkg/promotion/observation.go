package promotion

import (
	"fmt"
	"reflect"
	"time"
)

// Observation is read-only live evidence. Runtime configuration is separate
// from the artifact's immutable build configuration. Mixed or unknown runtime
// states need not invent a common source, version or complete artifact identity.
type Observation struct {
	Repository    string   `json:"repository"`
	Environment   string   `json:"environment"`
	SourceSHA     string   `json:"source_sha,omitempty"`
	Version       string   `json:"version,omitempty"`
	Artifact      Artifact `json:"artifact"`
	Configuration string   `json:"configuration_sha256,omitempty"`
	Status        string   `json:"status"`
	ObservedAt    string   `json:"observed_at"`
	RunID         int64    `json:"run_id"`
	RunURL        string   `json:"run_url"`
}

func validateObservation(o Observation, now time.Time) (time.Time, error) {
	observed, err := time.Parse(time.RFC3339Nano, o.ObservedAt)
	if err != nil || observed.After(now) || o.Repository == "" || o.Environment == "" || o.RunID <= 0 || o.RunURL == "" {
		return observed, fmt.Errorf("observation lacks scoped current timestamp or run evidence")
	}
	if o.Status != "live" && o.Status != "deploying" && o.Status != "unknown" {
		return observed, fmt.Errorf("observation status must be live, deploying or unknown")
	}
	if (o.SourceSHA != "" && !shaPattern.MatchString(o.SourceSHA)) || (o.Version != "" && !versionPattern.MatchString(o.Version)) || (o.Configuration != "" && !digestPattern.MatchString(o.Configuration)) {
		return observed, fmt.Errorf("observation contains invalid runtime identity")
	}
	artifactPresent := !reflect.DeepEqual(o.Artifact, Artifact{})
	if artifactPresent {
		if err := ValidateArtifact(o.Artifact); err != nil {
			return observed, err
		}
	}
	if o.Status == "live" && (o.SourceSHA == "" || o.Version == "" || !artifactPresent || o.Configuration == "") {
		return observed, fmt.Errorf("live observation requires complete source, artifact and runtime configuration evidence")
	}
	return observed, nil
}

// ApplyObservation advances only the observation record. It never advances
// verified deployment history or clears an active deployment fence.
func ApplyObservation(s *State, observation Observation) (bool, error) {
	if s == nil || observation.Repository != s.Repository || observation.Environment != s.Environment {
		return false, fmt.Errorf("observation belongs to a foreign repository or environment")
	}
	observed, err := validateObservation(observation, time.Now().UTC())
	if err != nil {
		return false, err
	}
	if s.Observation != nil {
		previous, err := time.Parse(time.RFC3339Nano, s.Observation.ObservedAt)
		if err != nil {
			return false, fmt.Errorf("prior observation timestamp is invalid; reconcile its evidence")
		}
		if observed.Before(previous) {
			return false, fmt.Errorf("observation is older than the recorded live evidence")
		}
		if reflect.DeepEqual(*s.Observation, observation) {
			return false, nil
		}
		if observed.Equal(previous) || s.Observation.RunID == observation.RunID {
			return false, fmt.Errorf("observation conflicts with an existing timestamp or immutable run identity")
		}
	}
	// Detach the stored bundle from caller-owned maps; observation input must not
	// later mutate journal truth without a fresh authenticated evidence record.
	copy := observation
	if observation.Artifact.Members != nil {
		copy.Artifact.Members = make(map[string]ArtifactObject, len(observation.Artifact.Members))
		for name, member := range observation.Artifact.Members {
			copy.Artifact.Members[name] = member
		}
	}
	s.Observation = &copy
	return true, nil
}
