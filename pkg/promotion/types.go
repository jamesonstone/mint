// Package promotion owns release eligibility, proposals, immutable intents and
// production history. Application adapters own artifact building and deployment.
package promotion

import (
	"encoding/json"
	"fmt"
	"regexp"
)

var shaPattern = regexp.MustCompile(`^[a-f0-9]{40}$`)
var digestPattern = regexp.MustCompile(`^sha256:[a-f0-9]{64}$`)
var versionPattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

// Artifact identifies an immutable object and its exact build configuration.
type ArtifactObject struct {
	Reference     string `json:"reference"`
	Digest        string `json:"digest"`
	Configuration string `json:"configuration_sha256"`
}

type Artifact struct {
	Members       map[string]ArtifactObject `json:"members,omitempty"`
	Reference     string                    `json:"reference"`
	Digest        string                    `json:"digest"`
	Configuration string                    `json:"configuration_sha256"`
}

// Change records reviewed logical provenance. PatchID must be independently
// recomputed from Git history; titles and issue text never prove equivalence.
type Change struct {
	Hotfix          bool   `json:"hotfix,omitempty"`
	OriginalVersion string `json:"original_version,omitempty"`
	Reverts         string `json:"reverts_patch_id,omitempty"`
	SHA             string `json:"sha"`
	Version         string `json:"version"`
	PR              int    `json:"pr,omitempty"`
	Title           string `json:"title"`
	PatchID         string `json:"patch_id"`
	Revert          bool   `json:"revert,omitempty"`
}

// Candidate is published only after the trusted source build succeeds.
type Candidate struct {
	SourceDate  string   `json:"source_date"`
	SourcePR    int      `json:"source_pr,omitempty"`
	Repository  string   `json:"repository"`
	Environment string   `json:"environment"`
	SourceSHA   string   `json:"source_sha"`
	Version     string   `json:"version"`
	Kind        string   `json:"kind"`
	Artifact    Artifact `json:"artifact"`
	RunID       int64    `json:"run_id"`
	RunURL      string   `json:"run_url"`
	Changes     []Change `json:"changes"`
	BaselineID  string   `json:"baseline_id,omitempty"`
	ControlOnly bool     `json:"control_only"`
}

// Baseline represents verified production, not a tag or a build.
type Baseline struct {
	VerifiedAt         string    `json:"verified_at,omitempty"`
	Configuration      string    `json:"runtime_configuration_sha256,omitempty"`
	PreviousID         string    `json:"previous_id,omitempty"`
	MainAnchor         string    `json:"main_anchor,omitempty"`
	ID                 string    `json:"id"`
	Candidate          Candidate `json:"candidate"`
	DeploymentURL      string    `json:"deployment_url"`
	Shipped            []Change  `json:"shipped"`
	PublicationPending bool      `json:"publication_pending"`
}

// Proposal has a stable journal identity independent of title or branch loss.
type Proposal struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	PR           int    `json:"pr"`
	Branch       string `json:"branch"`
	State        string `json:"state"`
	Selection    string `json:"selection"`
	CandidateSHA string `json:"candidate_sha"`
	BaselineID   string `json:"baseline_id"`
	Summary      string `json:"summary"`
	HeadSHA      string `json:"head_sha"`
	Notes        string `json:"notes"`
}

// Intent freezes exactly the reviewed merge declaration and canonical notes.
type Intent struct {
	PolicyDigest    string    `json:"policy_digest,omitempty"`
	Publish         bool      `json:"publish,omitempty"`
	Environment     string    `json:"environment,omitempty"`
	Configuration   string    `json:"runtime_configuration_sha256,omitempty"`
	Kind            string    `json:"kind"`
	DeploymentRunID int64     `json:"deployment_run_id,omitempty"`
	ID              string    `json:"id"`
	ProposalID      string    `json:"proposal_id"`
	MergeSHA        string    `json:"merge_sha"`
	BaselineID      string    `json:"baseline_id"`
	Candidate       Candidate `json:"candidate"`
	Notes           string    `json:"notes"`
	Status          string    `json:"status"`
	DeploymentURL   string    `json:"deployment_url,omitempty"`
}

// State is persisted atomically using the journal revision as a CAS fence.
type State struct {
	PolicyDigest    string               `json:"policy_digest,omitempty"`
	Observation     *Observation         `json:"observation,omitempty"`
	Target          string               `json:"target,omitempty"`
	Configuration   string               `json:"runtime_configuration_sha256,omitempty"`
	Publish         bool                 `json:"publish,omitempty"`
	Paused          bool                 `json:"paused,omitempty"`
	ControlRequests map[string]Proposal  `json:"control_requests,omitempty"`
	History         map[string]Baseline  `json:"history"`
	MainAnchor      string               `json:"main_anchor"`
	Schema          int                  `json:"schema_version"`
	Repository      string               `json:"repository"`
	Environment     string               `json:"environment"`
	Baseline        *Baseline            `json:"baseline"`
	Candidates      map[string]Candidate `json:"candidates"`
	Proposals       map[string]Proposal  `json:"proposals"`
	Intents         map[string]Intent    `json:"intents"`
	InFlight        string               `json:"in_flight,omitempty"`
}

// NewState starts unactivated, without fabricating a production baseline.
func NewState(repository, environment string) State {
	return State{Schema: 1, Repository: repository, Environment: environment, Candidates: map[string]Candidate{}, Proposals: map[string]Proposal{}, Intents: map[string]Intent{}, History: map[string]Baseline{}}
}

// DecodeState validates journal identity before using remote JSON.
func DecodeState(data []byte, repository, environment string) (State, error) {
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return s, err
	}
	if (s.Schema != 1 && s.Schema != 2) || s.Repository != repository || s.Environment != environment || s.Candidates == nil || s.Proposals == nil || s.Intents == nil {
		return s, fmt.Errorf("invalid or foreign release journal")
	}
	if s.History == nil {
		s.History = map[string]Baseline{}
	}
	if s.Baseline != nil {
		s.History[s.Baseline.ID] = cloneBaseline(*s.Baseline)
	}
	return s, nil
}
func validateCandidate(c Candidate) error {
	if err := ValidateArtifact(c.Artifact); err != nil {
		return err
	}
	if !shaPattern.MatchString(c.SourceSHA) || !versionPattern.MatchString(c.Version) || !digestPattern.MatchString(c.Artifact.Digest) || !digestPattern.MatchString(c.Artifact.Configuration) || c.Artifact.Reference == "" || c.RunID <= 0 || c.RunURL == "" {
		return fmt.Errorf("candidate lacks exact source, version, artifact, configuration or build identity")
	}
	if c.Kind != "normal" && c.Kind != "hotfix" {
		return fmt.Errorf("unknown candidate kind")
	}
	for _, ch := range c.Changes {
		if !shaPattern.MatchString(ch.SHA) || !versionPattern.MatchString(ch.Version) || ch.PatchID == "" {
			return fmt.Errorf("change lacks verified provenance")
		}
	}
	return nil
}

// MainVersionSelection binds the version used by a source build, including hotfix provenance.
type MainVersionSelection struct {
	SourceSHA  string `json:"source_sha"`
	Version    string `json:"version"`
	SourcePR   int    `json:"source_pr,omitempty"`
	BaselineID string `json:"baseline_id,omitempty"`
}

// ValidateBaselineCandidate checks the complete immutable bootstrap identity.
func ValidateBaselineCandidate(c Candidate) error { return validateCandidate(c) }
