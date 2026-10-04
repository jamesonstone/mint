package promotion

import (
	"strings"
	"testing"
)

func rollbackHistory() State {
	s, _ := fixture()
	first := *s.Baseline
	s.History[first.ID] = first
	second := Baseline{ID: "production-1", PreviousID: first.ID, Candidate: candidate(1, "normal"), DeploymentURL: "https://deployment/1", PublicationPending: true}
	s.History[second.ID] = second
	s.Baseline = &second
	return s
}

func TestRollbackTargetFollowsDeploymentAncestry(t *testing.T) {
	s := rollbackHistory()
	unrelated := Baseline{ID: "unrelated", Candidate: candidate(2, "normal"), DeploymentURL: "https://deployment/2"}
	s.History[unrelated.ID] = unrelated
	repeated := *s.Baseline
	repeated.ID = "repeat"
	repeated.PreviousID = s.Baseline.ID
	s.History[repeated.ID] = repeated
	s.Baseline = &repeated
	target, err := s.ResolveRollbackTarget("")
	if err != nil || target.ID != "production-0" {
		t.Fatalf("selected unrelated or repeated deployment: %s %v", target.ID, err)
	}
}

func TestRollbackTargetLegacyRequiresExplicitVersion(t *testing.T) {
	s := rollbackHistory()
	s.Baseline.PreviousID = ""
	if _, err := s.ResolveRollbackTarget(""); err == nil {
		t.Fatal("inferred previous deployment from unordered legacy history")
	}
	target, err := s.ResolveRollbackTarget("v0.1.0")
	if err != nil || target.ID != "production-0" {
		t.Fatal("explicit verified legacy version unavailable", err)
	}
}

func TestRollbackTargetRejectsMissingCyclicAndUnverifiedHistory(t *testing.T) {
	for _, mutate := range []func(*State){
		func(s *State) { delete(s.History, "production-0") },
		func(s *State) { s.Baseline.PreviousID = s.Baseline.ID },
		func(s *State) { b := s.History["production-0"]; b.DeploymentURL = ""; s.History[b.ID] = b },
	} {
		s := rollbackHistory()
		mutate(&s)
		if _, err := s.ResolveRollbackTarget(""); err == nil {
			t.Fatal("invalid history accepted")
		}
	}
}

func TestRollbackVersionRequiresUniqueVerifiedArtifact(t *testing.T) {
	s := rollbackHistory()
	for _, version := range []string{"0.1.0", "v01.1.0", "v0.1.9", "v0.1.1"} {
		if _, err := s.ResolveRollbackTarget(version); err == nil {
			t.Fatalf("invalid, unknown or current version accepted: %s", version)
		}
	}
	b := s.History["production-0"]
	b.ID = "equivalent"
	b.PublicationPending = true
	s.History[b.ID] = b
	if _, err := s.ResolveRollbackTarget("v0.1.0"); err != nil {
		t.Fatal("equivalent verified deployment awaiting publication rejected", err)
	}
	b.ID = "ambiguous"
	b.Candidate.Artifact.Digest = "sha256:" + strings.Repeat("c", 64)
	s.History[b.ID] = b
	if _, err := s.ResolveRollbackTarget("v0.1.0"); err == nil {
		t.Fatal("ambiguous artifact accepted")
	}
}

func TestRollbackRequestReplayKeepsReviewedProposal(t *testing.T) {
	s := rollbackHistory()
	p, err := s.ReconcileRollback(sha(0), "request-1", "restore login")
	if err != nil {
		t.Fatal(err)
	}
	p.PR, p.State = 123, "merged"
	s.Proposals["rollback"] = p
	s.InFlight = "deployment-1"
	replay, err := s.ReconcileRollback(sha(0), "request-1", "restore login")
	if err != nil || replay != p {
		t.Fatal("identical retry replaced reviewed proposal", err)
	}
	if _, err := s.ReconcileRollback(sha(0), "request-1", "changed reason"); err == nil {
		t.Fatal("changed request accepted")
	}
}
