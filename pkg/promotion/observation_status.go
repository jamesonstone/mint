package promotion

import (
	"reflect"
	"time"
)

// EnvironmentStatus keeps editable intent, deployment truth and timestamped
// live evidence separate. It is a projection and performs no lifecycle mutation.
type EnvironmentStatus struct {
	Repository   string       `json:"repository"`
	Environment  string       `json:"environment"`
	Desired      *Candidate   `json:"desired,omitempty"`
	LastVerified *Baseline    `json:"last_verified,omitempty"`
	Live         *Observation `json:"live,omitempty"`
	Status       string       `json:"status"`
	AsOf         string       `json:"as_of,omitempty"`
	Evidence     string       `json:"evidence,omitempty"`
	InFlight     string       `json:"in_flight,omitempty"`
}

// ProjectEnvironmentStatus reports drift without accepting it as deployment
// truth. A live observation never overrides an unresolved deployment fence.
func ProjectEnvironmentStatus(s State, desired *Candidate, desiredConfiguration string) EnvironmentStatus {
	result := EnvironmentStatus{Repository: s.Repository, Environment: s.Environment, Desired: desired, LastVerified: s.Baseline, Live: s.Observation, Status: "unknown", InFlight: s.InFlight}
	if s.Observation == nil {
		return result
	}
	live := s.Observation
	result.AsOf, result.Evidence = live.ObservedAt, live.RunURL
	if _, err := validateObservation(*live, time.Now().UTC()); err != nil || live.Repository != s.Repository || live.Environment != s.Environment {
		return result
	}
	if live.Status == "unknown" || observationPredatesBaseline(s) {
		return result
	}
	if live.Status == "deploying" || s.InFlight != "" {
		result.Status = "deploying"
		return result
	}
	expected := desired
	configuration := desiredConfiguration
	if expected == nil && s.Baseline != nil {
		expected = &s.Baseline.Candidate
		configuration = s.Baseline.Configuration
	}
	if expected == nil {
		return result
	}
	if configuration == "" {
		configuration = expected.Artifact.Configuration // Legacy schema 1 binding.
	}
	result.Status = "drift"
	if live.SourceSHA == expected.SourceSHA && live.Version == expected.Version && live.Configuration == configuration && reflect.DeepEqual(live.Artifact, expected.Artifact) {
		result.Status = "live"
	}
	return result
}

// Server-verified completion supersedes an older snapshot, without deleting it.
// Missing timestamps retain conservative legacy behavior.
func observationPredatesBaseline(s State) bool {
	if s.Baseline == nil || s.Observation == nil {
		return false
	}
	verified, err := time.Parse(time.RFC3339Nano, s.Baseline.VerifiedAt)
	observed, observedErr := time.Parse(time.RFC3339Nano, s.Observation.ObservedAt)
	return err == nil && observedErr == nil && observed.Before(verified)
}
