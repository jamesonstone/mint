package promotion

import (
	"fmt"
	"reflect"
	"sort"
)

// ResolveRollbackTarget selects verified deployment history, never build or tag
// order. An omitted version follows durable deployment ancestry to the previous
// distinct artifact. Legacy history without ancestry requires an explicit version.
func (s *State) ResolveRollbackTarget(version string) (Baseline, error) {
	if s.Baseline == nil {
		return Baseline{}, fmt.Errorf("rollback requires verified production")
	}
	if version != "" {
		if !versionPattern.MatchString(version) {
			return Baseline{}, fmt.Errorf("rollback version must be strict vMAJOR.MINOR.PATCH")
		}
		return s.rollbackTargetMatching(func(b Baseline) bool { return b.Candidate.Version == version })
	}
	previous := s.Baseline.PreviousID
	visited := map[string]bool{s.Baseline.ID: true}
	for previous != "" {
		if visited[previous] {
			return Baseline{}, fmt.Errorf("deployment history contains a cycle; reconcile verified history")
		}
		visited[previous] = true
		b, ok := s.History[previous]
		if !ok || b.ID != previous || !s.verifiedRollbackBaseline(b) {
			return Baseline{}, fmt.Errorf("previous verified deployment history is missing; specify a verified version")
		}
		if !sameDeployedArtifact(b, *s.Baseline) {
			return cloneBaseline(b), nil
		}
		previous = b.PreviousID
	}
	return Baseline{}, fmt.Errorf("previous distinct verified deployment is unknown; specify a verified version")
}

func sameDeployedArtifact(a, b Baseline) bool {
	return a.Candidate.SourceSHA == b.Candidate.SourceSHA && reflect.DeepEqual(a.Candidate.Artifact, b.Candidate.Artifact)
}

func (s *State) verifiedRollbackBaseline(b Baseline) bool {
	return b.ID != "" && b.DeploymentURL != "" && b.Candidate.Repository == s.Repository && s.AcceptsEnvironment(b.Candidate.Environment) && validateCandidate(b.Candidate) == nil
}

// Repeated deployments of one immutable version are equivalent targets. Sort
// IDs solely to make equivalent record selection stable, never to infer recency.
func (s *State) rollbackTargetMatching(matches func(Baseline) bool) (Baseline, error) {
	ids := make([]string, 0, len(s.History))
	for id := range s.History {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	var target *Baseline
	for _, id := range ids {
		b := s.History[id]
		if !matches(b) || !s.verifiedRollbackBaseline(b) || b.ID != id {
			continue
		}
		if target != nil && !sameDeployedArtifact(*target, b) {
			return Baseline{}, fmt.Errorf("rollback target has ambiguous verified artifacts; reconcile deployment history")
		}
		if target == nil {
			target = &b
		}
	}
	if target == nil {
		return Baseline{}, fmt.Errorf("rollback target is not a known verified deployment")
	}
	if sameDeployedArtifact(*target, *s.Baseline) {
		return Baseline{}, fmt.Errorf("rollback target is already deployed")
	}
	return cloneBaseline(*target), nil
}
