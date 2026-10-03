package promotion

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// FreezeIntent checks the exact merged declaration against the journal. Callers
// must authenticate the PR marker, approved file set and head checks independently.
func (s *State) FreezeIntent(p Proposal, c Candidate, mergeSHA string) (Intent, error) {
	id := p.ID + ":" + mergeSHA
	if prior, exists := s.Intents[id]; exists {
		if prior.ProposalID != p.ID || prior.MergeSHA != mergeSHA || !reflect.DeepEqual(prior.Candidate, c) || prior.Notes != p.Notes {
			return Intent{}, fmt.Errorf("frozen intent conflict")
		}
		return prior, nil
	}
	current, ok := s.Proposals[p.Kind]
	if !ok || !reflect.DeepEqual(current, p) || p.State != "open" || !shaPattern.MatchString(mergeSHA) {
		return Intent{}, fmt.Errorf("merged proposal identity or reviewed head is stale")
	}
	registered, ok := s.Candidates[c.SourceSHA]
	if !ok || !reflect.DeepEqual(registered, c) || p.CandidateSHA != c.SourceSHA || s.Baseline == nil || p.BaselineID != s.Baseline.ID {
		return Intent{}, fmt.Errorf("declaration source/artifact/baseline differs from reviewed proposal")
	}
	id = p.ID + ":" + mergeSHA
	intent := Intent{Kind: p.Kind, ID: id, ProposalID: p.ID, MergeSHA: mergeSHA, BaselineID: p.BaselineID, Candidate: c, Notes: p.Notes, Status: "pending"}
	s.Intents[id] = intent
	p.State = "merged"
	s.Proposals[p.Kind] = p
	return intent, nil
}

// StartDeployment is the shared normal/hotfix production fence. Replaying the
// same intent is safe; a successfully deployed intent is publication-only.
func (s *State) StartDeployment(id string, proof Proof) (Intent, error) {
	i, ok := s.Intents[id]
	if !ok {
		return Intent{}, fmt.Errorf("unknown intent")
	}
	if i.Status == "deployed" || i.Status == "publication_pending" {
		return i, nil
	}
	if s.Baseline == nil || i.BaselineID != s.Baseline.ID {
		return Intent{}, fmt.Errorf("stale reviewed production baseline; reconcile and review a new proposal")
	}
	if s.InFlight != "" && s.InFlight != id {
		return Intent{}, fmt.Errorf("another production deployment is in flight")
	}
	if s.proposalKind(i) != "rollback" {
		if _, err := s.NewChanges(i.Candidate, proof); err != nil {
			return Intent{}, err
		}
	}
	i.Status = "deploying"
	s.InFlight = id
	s.Intents[id] = i
	return i, nil
}

// FinishDeployment advances baseline only on verified success, before Release
// publication. Unknown/failed observations never become successful production.
func (s *State) FinishDeployment(id, outcome, evidenceURL string, proof Proof) error {
	i, ok := s.Intents[id]
	if !ok {
		return fmt.Errorf("unknown intent")
	}
	if i.Status == "publication_pending" || i.Status == "deployed" {
		return nil
	}
	if s.InFlight != id || i.Status != "deploying" {
		return fmt.Errorf("intent is not the active deployment")
	}
	if outcome != "success" && outcome != "failure" {
		return fmt.Errorf("unknown verification outcome; retain in-flight fence")
	}
	if evidenceURL == "" {
		return fmt.Errorf("deployment evidence required")
	}
	if outcome == "success" && s.proposalKind(i) == "rollback" {
		if err := s.finishRollback(&i, evidenceURL); err != nil {
			return err
		}
	} else if outcome == "success" {
		added, err := s.NewChanges(i.Candidate, proof)
		if err != nil {
			return err
		}
		shipped := []Change{}
		for _, prior := range s.Baseline.Shipped {
			removed := false
			for _, change := range added {
				if change.Reverts == prior.PatchID {
					removed = true
				}
			}
			if !removed {
				shipped = append(shipped, prior)
			}
		}
		shipped = append(shipped, added...)
		if i.Candidate.Kind == "normal" {
			s.MainAnchor = i.Candidate.SourceSHA
		}
		s.History[s.Baseline.ID] = *s.Baseline
		s.Baseline = &Baseline{MainAnchor: s.MainAnchor, ID: id, Candidate: i.Candidate, DeploymentURL: evidenceURL, Shipped: shipped, PublicationPending: true}
		i.Status = "publication_pending"
	} else {
		i.Status = "deployment_failed"
	}
	i.DeploymentURL = evidenceURL
	s.InFlight = ""
	s.Intents[id] = i
	p := s.Proposals[s.proposalKind(i)]
	p.State = i.Status
	s.Proposals[p.Kind] = p
	return nil
}

// MarkPublished acknowledges publication without redeploying or retagging.
func (s *State) MarkPublished(id string) error {
	i, ok := s.Intents[id]
	if !ok || (i.Status != "publication_pending" && i.Status != "deployed") {
		return fmt.Errorf("intent has no verified deployment")
	}
	i.Status = "deployed"
	s.Intents[id] = i
	if s.Baseline != nil && s.Baseline.ID == id {
		s.Baseline.PublicationPending = false
		s.History[id] = *s.Baseline
	}
	p := s.Proposals[s.proposalKind(i)]
	if p.ID == i.ProposalID {
		p.State = "deployed"
		s.Proposals[p.Kind] = p
	}
	return nil
}

// CanonicalIntent returns a stable immutable manifest suitable for hashing.
func CanonicalIntent(i Intent) ([]byte, error) { return json.Marshal(i) }
