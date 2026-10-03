package promotion

import "testing"

func TestReviewedRollbackUsesExistingArtifactWithoutPublication(t *testing.T) {
	s, p := fixture()
	s.Baseline.MainAnchor = s.Baseline.Candidate.SourceSHA
	s.History[s.Baseline.ID] = *s.Baseline
	register(t, &s, candidate(1, "normal"))
	i := start(t, &s, p, "normal")
	if err := s.FinishDeployment(i.ID, "success", "https://deploy/1", p); err != nil {
		t.Fatal(err)
	}
	if err := s.MarkPublished(i.ID); err != nil {
		t.Fatal(err)
	}
	rollback, err := s.ReconcileRollback(sha(0), "rollback-1", "Restore previous verified version.")
	if err != nil {
		t.Fatal(err)
	}
	intent, err := s.FreezeIntent(rollback, s.Candidates[sha(0)], sha(3))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartDeployment(intent.ID, p); err != nil {
		t.Fatal(err)
	}
	if err := s.FinishDeployment(intent.ID, "success", "https://deploy/rollback", p); err != nil {
		t.Fatal(err)
	}
	if s.Baseline.Candidate.SourceSHA != sha(0) || s.Baseline.PublicationPending || s.Intents[intent.ID].Status != "deployed" || s.MainAnchor != sha(0) {
		t.Fatal("rollback was incorrectly published or did not restore verified history")
	}
	if err := (Client{}).PublishIntent(t.Context(), s.Intents[intent.ID]); err == nil {
		t.Fatal("rollback created a Release")
	}
	if _, err := s.ReconcileRollback(sha(2), "unknown", "unknown"); err == nil {
		t.Fatal("unverified rollback target accepted")
	}
}
func TestFrozenIntentReplayDoesNotRequireAnotherDeployment(t *testing.T) {
	s, p := fixture()
	register(t, &s, candidate(1, "normal"))
	proposal := propose(t, &s, p, "normal")
	original, err := s.FreezeIntent(proposal, s.Candidates[sha(1)], sha(4))
	if err != nil {
		t.Fatal(err)
	}
	replay, err := s.FreezeIntent(proposal, s.Candidates[sha(1)], sha(4))
	if err != nil || replay.ID != original.ID {
		t.Fatal("exact frozen merge replay rejected", err)
	}
}
