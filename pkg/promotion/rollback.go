package promotion

import "fmt"

// ReconcileRollback selects only an already verified production artifact. The
// rollback is separately reviewed against the current baseline and never mints
// another version or publishes another GitHub Release.
func (s *State) ReconcileRollback(sourceSHA, id, summary string) (Proposal, error) {
	prior, exists := s.Proposals["rollback"]
	if exists && prior.ID == id {
		if prior.CandidateSHA != sourceSHA || prior.Summary != summary {
			return Proposal{}, fmt.Errorf("registered rollback request differs; use a new request identity")
		}
		return prior, nil
	}
	if s.Baseline == nil || s.InFlight != "" {
		return Proposal{}, fmt.Errorf("rollback requires verified idle production")
	}
	target, err := s.rollbackTargetMatching(func(b Baseline) bool { return b.Candidate.SourceSHA == sourceSHA })
	if err != nil {
		return Proposal{}, err
	}
	if exists && prior.State != "deployed" {
		return Proposal{}, fmt.Errorf("existing rollback must finish or be explicitly resumed before replacement")
	}
	p := Proposal{ID: id, Kind: "rollback", State: "open", Selection: "pinned", CandidateSHA: sourceSHA, BaselineID: s.Baseline.ID, Summary: summary, Notes: "## Rollback from " + s.Baseline.Candidate.Version + " to " + target.Candidate.Version + "\n\n" + summary + "\n\nExisting production artifact: `" + target.Candidate.Artifact.Digest + "`.\n\nPrefer a roll-forward hotfix when practical. Rollback restores application artifacts only; it does not reverse database or data changes. Verify compatibility before approval.\n"}
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
	target, err := s.rollbackTargetMatching(func(b Baseline) bool {
		return b.Candidate.SourceSHA == i.Candidate.SourceSHA && b.Candidate.Artifact == i.Candidate.Artifact
	})
	if err != nil {
		return fmt.Errorf("verified rollback target unavailable: %w", err)
	}
	s.History[s.Baseline.ID] = *s.Baseline
	target.PreviousID = s.Baseline.ID
	target.ID = i.ID
	target.DeploymentURL = evidenceURL
	target.PublicationPending = false
	s.Baseline = &target
	s.MainAnchor = target.MainAnchor
	s.History[i.ID] = target
	i.Status = "deployed"
	return nil
}
