package promotion

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// FreezeIntent checks the exact merged declaration against the journal. Callers
// must authenticate the PR marker, approved file set and head checks independently.
func (s *State) FreezeIntent(p Proposal, c Candidate, mergeSHA string) (Intent, error) {
	if s.Schema == 2 && !digestPattern.MatchString(s.Configuration) {
		return Intent{}, fmt.Errorf("environment intent requires exact runtime configuration")
	}
	id := p.ID + ":" + mergeSHA
	if prior, exists := s.Intents[id]; exists {
		if prior.ProposalID != p.ID || prior.MergeSHA != mergeSHA || !reflect.DeepEqual(prior.Candidate, c) || prior.Notes != p.Notes || prior.Configuration != s.Configuration || prior.PolicyDigest != s.PolicyDigest {
			return Intent{}, fmt.Errorf("frozen intent conflict")
		}
		return cloneIntent(prior), nil
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
	intent := Intent{PolicyDigest: s.PolicyDigest, Publish: s.shouldPublish(c), Environment: s.Environment, Configuration: s.Configuration, Kind: p.Kind, ID: id, ProposalID: p.ID, MergeSHA: mergeSHA, BaselineID: p.BaselineID, Candidate: c, Notes: p.Notes, Status: "pending"}
	s.Intents[id] = cloneIntent(intent)
	p.State = "merged"
	s.Proposals[p.Kind] = p
	return cloneIntent(intent), nil
}

// StartDeployment is the shared normal/hotfix production fence. Replaying the
// same intent is safe; a successfully deployed intent is publication-only.
func (s *State) StartDeployment(id string, proof Proof) (Intent, error) {
	i, ok := s.Intents[id]
	if !ok {
		return Intent{}, fmt.Errorf("unknown intent")
	}
	if i.Status == "deployed" || i.Status == "publication_pending" {
		return cloneIntent(i), nil
	}
	if s.Baseline == nil || i.BaselineID != s.Baseline.ID {
		return Intent{}, fmt.Errorf("stale reviewed production baseline; reconcile and review a new proposal")
	}
	if s.InFlight != "" && s.InFlight != id {
		return Intent{}, fmt.Errorf("another production deployment is in flight")
	}
	if s.Schema == 2 && s.Observation != nil && !observationPredatesBaseline(*s) {
		live := s.Observation
		if live.Status != "live" || live.SourceSHA != s.Baseline.Candidate.SourceSHA || live.Version != s.Baseline.Candidate.Version || live.Configuration != s.Baseline.RuntimeConfiguration() || !reflect.DeepEqual(live.Artifact, s.Baseline.Candidate.Artifact) {
			return Intent{}, fmt.Errorf("live environment differs from its verified baseline; reconcile before deployment")
		}
	}
	if s.proposalKind(i) != "rollback" {
		if _, err := s.NewChanges(i.Candidate, proof); err != nil {
			return Intent{}, err
		}
	}
	i.Status = "deploying"
	s.InFlight = id
	s.Intents[id] = cloneIntent(i)
	return cloneIntent(i), nil
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
		shipped := append([]Change{}, s.Baseline.Shipped...)
		// Apply logical events in order so a change introduced and reverted in
		// one release is not incorrectly retained as an active shipped fix.
		for _, change := range added {
			active := []Change{}
			for _, prior := range shipped {
				if prior.PatchID != change.Reverts {
					active = append(active, prior)
				}
			}
			shipped = append(active, change)
		}
		if i.Candidate.Kind == "normal" {
			s.MainAnchor = i.Candidate.SourceSHA
		}
		s.History[s.Baseline.ID] = cloneBaseline(*s.Baseline)
		s.Baseline = &Baseline{PreviousID: s.Baseline.ID, MainAnchor: s.MainAnchor, ID: id, Candidate: cloneCandidate(i.Candidate), DeploymentURL: evidenceURL, Shipped: shipped, Configuration: i.Configuration, PublicationPending: s.Schema == 1 || i.Publish}
		i.Status = "publication_pending"
		if s.Schema == 2 && !i.Publish {
			i.Status = "deployed"
		}
	} else {
		i.Status = "deployment_failed"
	}
	i.DeploymentURL = evidenceURL
	s.InFlight = ""
	s.Intents[id] = cloneIntent(i)
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
	s.Intents[id] = cloneIntent(i)
	if s.Baseline != nil && s.Baseline.ID == id {
		s.Baseline.PublicationPending = false
		s.History[id] = cloneBaseline(*s.Baseline)
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

// Redeploying an existing artifact is deployment history, not a new version
// publication. Preserve the original canonical Release and its evidence.
func (s State) shouldPublish(c Candidate) bool {
	if s.Schema == 1 {
		return true
	}
	if !s.Publish {
		return false
	}
	same := func(prior Candidate) bool {
		return prior.SourceSHA == c.SourceSHA && prior.Version == c.Version && reflect.DeepEqual(prior.Artifact, c.Artifact)
	}
	if s.Baseline != nil && same(s.Baseline.Candidate) {
		return false
	}
	for _, prior := range s.Intents {
		if prior.Publish && prior.Status == "deployed" && same(prior.Candidate) {
			return false
		}
	}
	return true
}
