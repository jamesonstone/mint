package promotion

import (
	"fmt"
	"reflect"
)

// Proof is supplied by the Git/GitHub adapter, not an event payload or issue text.
type Proof interface {
	Ancestor(older, newer string) (bool, error)
	Retains(candidateSHA string, shipped Change) (bool, error)
}

// BuildEvidence binds successful trusted workflow evidence to source and artifact.
type BuildEvidence struct {
	Repository, SourceSHA, TagSHA, ArtifactDigest, Configuration string
	RunID                                                        int64
	Success, Trusted                                             bool
}

// RegisterCandidate refuses failed/foreign builds and immutable identity conflicts.
func (s *State) RegisterCandidate(c Candidate, e BuildEvidence) error {
	if err := validateCandidate(c); err != nil {
		return err
	}
	if c.Repository != s.Repository || c.Environment != s.Environment || !e.Trusted || !e.Success || e.Repository != s.Repository || e.SourceSHA != c.SourceSHA || e.TagSHA != c.SourceSHA || e.ArtifactDigest != c.Artifact.Digest || e.Configuration != c.Artifact.Configuration || e.RunID != c.RunID {
		return fmt.Errorf("candidate build/tag/artifact evidence does not match")
	}
	if c.Kind == "hotfix" && (s.Baseline == nil || c.BaselineID != s.Baseline.ID) {
		return fmt.Errorf("hotfix built against stale or absent production")
	}
	if prior, ok := s.Candidates[c.SourceSHA]; ok {
		replay := c
		replay.RunID = prior.RunID
		replay.RunURL = prior.RunURL
		if !reflect.DeepEqual(prior, replay) {
			return fmt.Errorf("candidate identity is immutable")
		}
		return nil // Keep the first successful attestation of this exact artifact.
	}
	s.Candidates[c.SourceSHA] = c
	return nil
}

// SelectCandidate recomputes latest by source ancestry; event delivery order and
// numeric version order never select production. A pinned selection is explicit.
func (s *State) SelectCandidate(kind, pin string, proof Proof) (Candidate, error) {
	var selected *Candidate
	if pin != "" {
		c, ok := s.Candidates[pin]
		if !ok || c.Kind != kind {
			return Candidate{}, fmt.Errorf("pinned candidate unavailable")
		}
		selected = &c
	}
	for _, c := range s.Candidates {
		if pin != "" {
			break
		}
		if c.Kind != kind || c.ControlOnly {
			continue
		}
		if kind == "hotfix" && (s.Baseline == nil || c.BaselineID != s.Baseline.ID) {
			continue
		}
		if selected == nil {
			copy := c
			selected = &copy
			continue
		}
		older, err := proof.Ancestor(selected.SourceSHA, c.SourceSHA)
		if err != nil {
			return Candidate{}, err
		}
		if older {
			copy := c
			selected = &copy
			continue
		}
		newer, err := proof.Ancestor(c.SourceSHA, selected.SourceSHA)
		if err != nil {
			return Candidate{}, err
		}
		if !newer {
			return Candidate{}, fmt.Errorf("divergent candidates require an explicit pin")
		}
	}
	if selected == nil {
		return Candidate{}, fmt.Errorf("no eligible %s candidate", kind)
	}
	if selected.ControlOnly {
		return Candidate{}, fmt.Errorf("control-only candidate cannot generate a release")
	}
	if s.Baseline == nil {
		return Candidate{}, fmt.Errorf("verified production baseline must be imported before proposing")
	}
	if selected.SourceSHA == s.Baseline.Candidate.SourceSHA {
		return Candidate{}, fmt.Errorf("candidate is already deployed")
	}
	if _, err := s.NewChanges(*selected, proof); err != nil {
		return Candidate{}, err
	}
	return *selected, nil
}

// NewChanges excludes shipped logical equivalents only when the candidate tree
// independently retains them. Explicit reverts remain visible.
func (s *State) NewChanges(c Candidate, proof Proof) ([]Change, error) {
	if s.Baseline == nil {
		return nil, fmt.Errorf("production baseline unavailable")
	}
	shipped := map[string]Change{}
	for _, ch := range s.Baseline.Shipped {
		shipped[ch.PatchID] = ch
	}
	for _, ch := range s.Baseline.Shipped {
		retained, err := proof.Retains(c.SourceSHA, ch)
		if err != nil {
			return nil, err
		}
		if !retained {
			explicitRevert := false
			for _, candidateChange := range c.Changes {
				if candidateChange.Revert && candidateChange.Reverts == ch.PatchID {
					explicitRevert = true
				}
			}
			if !explicitRevert {
				return nil, fmt.Errorf("candidate may undo shipped change %s; reviewed forward integration required", ch.SHA)
			}
		}
	}
	result := []Change{}
	seen := map[string]bool{}
	for _, ch := range c.Changes {
		identity := ch.PatchID
		if ch.Revert {
			identity = "revert:" + ch.SHA
		}
		if seen[identity] {
			continue
		}
		seen[identity] = true
		if _, ok := shipped[ch.PatchID]; ok && !ch.Revert {
			continue
		}
		result = append(result, ch)
	}
	return result, nil
}

// Reconcile updates one ordinary proposal or one explicitly scoped hotfix.
// Paused proposals survive subsequent candidates; only reopen resumes them.
func (s *State) Reconcile(kind, event, pin, id, summary string, proof Proof) (Proposal, error) {
	if kind != "normal" && kind != "hotfix" {
		return Proposal{}, fmt.Errorf("invalid proposal kind")
	}
	prior, exists := s.Proposals[kind]
	if event == "close" {
		if !exists || prior.State != "open" {
			return Proposal{}, fmt.Errorf("no open proposal to pause")
		}
		prior.State = "paused"
		s.Proposals[kind] = prior
		return prior, nil
	}
	if exists && prior.State == "paused" && event != "reopen" {
		return prior, nil
	}
	if s.InFlight != "" {
		return Proposal{}, fmt.Errorf("production intent %s is in flight; collect candidates and reconcile its outcome first", s.InFlight)
	}
	if event == "reopen" {
		pin = ""
		prior.Selection = "latest"
	}
	if pin == "" && prior.Selection == "pinned" && event != "reopen" {
		pin = prior.CandidateSHA
	}
	c, err := s.SelectCandidate(kind, pin, proof)
	if err != nil {
		return Proposal{}, err
	}
	if !exists || prior.State == "deployed" {
		if id == "" {
			return Proposal{}, fmt.Errorf("new proposal requires durable generation identity")
		}
		prior = Proposal{ID: id, Kind: kind, State: "open", Selection: "latest"}
	}
	if prior.State == "merged" || prior.State == "deployment_failed" || prior.State == "publication_pending" {
		return Proposal{}, fmt.Errorf("merged intent must be retried or explicitly superseded, not silently replaced")
	}
	if summary != "" {
		prior.Summary = summary
	}
	changes, err := s.NewChanges(c, proof)
	if err != nil {
		return Proposal{}, err
	}
	prior.CandidateSHA = c.SourceSHA
	prior.BaselineID = s.Baseline.ID
	prior.State = "open"
	if pin != "" {
		prior.Selection = "pinned"
	}
	prior.Notes = RenderNotes(s.Repository, c.Version, prior.Summary, changes, c.SourceDate)
	s.Proposals[kind] = prior
	return prior, nil
}
