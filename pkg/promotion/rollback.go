package promotion

import "fmt"

// ReconcileRollback selects only an already verified production artifact. The
// rollback is separately reviewed against the current baseline and never mints
// another version or publishes another GitHub Release.
func (s *State) ReconcileRollback(sourceSHA, id, summary string) (Proposal, error) {
	if s.Baseline == nil || s.InFlight != "" {
		return Proposal{}, fmt.Errorf("rollback requires verified idle production")
	}
	var target *Baseline
	for _, baseline := range s.History {
		if baseline.Candidate.SourceSHA == sourceSHA && !baseline.PublicationPending {
			copy := baseline
			target = &copy
			break
		}
	}
	if target == nil || target.Candidate.SourceSHA == s.Baseline.Candidate.SourceSHA {
		return Proposal{}, fmt.Errorf("rollback target must be a previously verified different production artifact")
	}
	prior, exists := s.Proposals["rollback"]
	if exists && prior.State != "deployed" {
		return Proposal{}, fmt.Errorf("existing rollback must finish or be explicitly resumed before replacement")
	}
	p := Proposal{ID: id, Kind: "rollback", State: "open", Selection: "pinned", CandidateSHA: sourceSHA, BaselineID: s.Baseline.ID, Summary: summary, Notes: "## Rollback to " + target.Candidate.Version + "\n\n" + summary + "\n\nExisting production artifact: `" + target.Candidate.Artifact.Digest + "`.\n"}
	s.Candidates[sourceSHA] = target.Candidate
	s.Proposals["rollback"] = p
	return p, nil
}
func (s *State) proposalKind(i Intent) string {
	if i.Kind != "" {
		return i.Kind
	}
	return i.Candidate.Kind
}
func (s *State) finishRollback(i *Intent, evidenceURL string) error {
	var target *Baseline
	for _, baseline := range s.History {
		if baseline.Candidate.SourceSHA == i.Candidate.SourceSHA && baseline.Candidate.Artifact == i.Candidate.Artifact && !baseline.PublicationPending {
			copy := baseline
			target = &copy
			break
		}
	}
	if target == nil {
		return fmt.Errorf("verified rollback target disappeared")
	}
	target.ID = i.ID
	target.DeploymentURL = evidenceURL
	target.PublicationPending = false
	s.Baseline = target
	s.MainAnchor = target.MainAnchor
	s.History[i.ID] = *target
	i.Status = "deployed"
	return nil
}
