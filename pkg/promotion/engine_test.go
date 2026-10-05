package promotion

import (
	"reflect"
	"strings"
	"testing"
)

type proof struct {
	order   map[string]int
	missing map[string]bool
}

func (p proof) Ancestor(a, b string) (bool, error)          { return p.order[a] <= p.order[b], nil }
func (p proof) Retains(sha string, ch Change) (bool, error) { return !p.missing[sha+ch.PatchID], nil }
func sha(n int) string                                      { return strings.Repeat(string(rune('a'+n)), 40) }
func candidate(n int, kind string) Candidate {
	return Candidate{Repository: "owner/repo", Environment: "production", SourceSHA: sha(n), Version: "v0.1." + string(rune('0'+n)), Kind: kind, Artifact: Artifact{Reference: "immutable/object", Digest: "sha256:" + strings.Repeat("a", 64), Configuration: "sha256:" + strings.Repeat("b", 64)}, RunID: int64(n + 1), RunURL: "https://github.com/owner/repo/actions/runs/1", Changes: []Change{{SHA: sha(n), Version: "v0.1." + string(rune('0'+n)), PR: n + 1, Title: "Fix [something]", PatchID: "patch-" + string(rune('0'+n))}}}
}
func evidence(c Candidate) BuildEvidence {
	return BuildEvidence{Repository: c.Repository, SourceSHA: c.SourceSHA, TagSHA: c.SourceSHA, ArtifactDigest: c.Artifact.Digest, Configuration: c.Artifact.Configuration, RunID: c.RunID, Success: true, Trusted: true}
}
func fixture() (State, proof) {
	s := NewState("owner/repo", "production")
	base := candidate(0, "normal")
	s.Baseline = &Baseline{ID: "production-0", Candidate: base, Shipped: base.Changes, DeploymentURL: "https://deployment/0"}
	return s, proof{order: map[string]int{sha(0): 0, sha(1): 1, sha(2): 2, sha(3): 3, sha(4): 4}, missing: map[string]bool{}}
}
func register(t *testing.T, s *State, c Candidate) {
	t.Helper()
	if err := s.RegisterCandidate(c, evidence(c)); err != nil {
		t.Fatal(err)
	}
}
func propose(t *testing.T, s *State, p proof, kind string) Proposal {
	t.Helper()
	v, err := s.Reconcile(kind, "candidate", "", "generation-1", "manual summary", p)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func start(t *testing.T, s *State, p proof, kind string) Intent {
	t.Helper()
	proposal := propose(t, s, p, kind)
	i, err := s.FreezeIntent(proposal, s.Candidates[proposal.CandidateSHA], sha(4))
	if err != nil {
		t.Fatal(err)
	}
	i, err = s.StartDeployment(i.ID, p)
	if err != nil {
		t.Fatal(err)
	}
	return i
}

func TestRegisterRejectsUntrustedOrFailedEvidenceAndConflicts(t *testing.T) {
	for _, mutate := range []func(*BuildEvidence){func(e *BuildEvidence) { e.Success = false }, func(e *BuildEvidence) { e.Trusted = false }, func(e *BuildEvidence) { e.TagSHA = sha(3) }, func(e *BuildEvidence) { e.Repository = "foreign/repo" }, func(e *BuildEvidence) { e.ArtifactDigest = "sha256:" + strings.Repeat("c", 64) }} {
		s, _ := fixture()
		c := candidate(1, "normal")
		e := evidence(c)
		mutate(&e)
		if s.RegisterCandidate(c, e) == nil {
			t.Fatal("invalid build accepted")
		}
		if len(s.Candidates) != 0 {
			t.Fatal("failed registration mutated state")
		}
	}
	s, _ := fixture()
	c := candidate(1, "normal")
	register(t, &s, c)
	register(t, &s, c)
	c.Artifact.Reference = "another/object"
	if s.RegisterCandidate(c, evidence(c)) == nil {
		t.Fatal("artifact identity overwritten")
	}
}
func TestLatestOutOfOrderPauseReopenAndSummary(t *testing.T) {
	s, p := fixture()
	register(t, &s, candidate(2, "normal"))
	register(t, &s, candidate(1, "normal"))
	initial := propose(t, &s, p, "normal")
	if initial.CandidateSHA != sha(2) {
		t.Fatal("event order regressed latest")
	}
	paused, err := s.Reconcile("normal", "close", "", "", "", p)
	if err != nil || paused.State != "paused" {
		t.Fatal(paused, err)
	}
	register(t, &s, candidate(3, "normal"))
	still, err := s.Reconcile("normal", "candidate", "", "", "", p)
	if err != nil || !reflect.DeepEqual(still, paused) {
		t.Fatal("paused proposal changed")
	}
	reopened, err := s.Reconcile("normal", "reopen", "", "", "", p)
	if err != nil || reopened.CandidateSHA != sha(3) || reopened.Summary != "manual summary" {
		t.Fatal(reopened, err)
	}
}
func TestNormalPinRejectedAndLegacyPinTracksLatest(t *testing.T) {
	s, p := fixture()
	register(t, &s, candidate(1, "normal"))
	register(t, &s, candidate(2, "normal"))
	if _, err := s.Reconcile("normal", "candidate", sha(1), "generation", "custom", p); err == nil || !strings.Contains(err.Error(), "pause") {
		t.Fatal("normal pin accepted or missing pause guidance", err)
	}
	if _, err := s.SelectCandidate("normal", sha(1), p); err == nil {
		t.Fatal("direct normal pin accepted")
	}
	s.Proposals["normal"] = Proposal{ID: "generation", Kind: "normal", State: "open", Selection: "pinned", CandidateSHA: sha(1), Summary: "custom"}
	v, err := s.Reconcile("normal", "candidate", "", "", "", p)
	if err != nil || v.CandidateSHA != sha(2) || v.Selection != "latest" || v.ID != "generation" || v.Summary != "custom" {
		t.Fatal(v, err)
	}
	register(t, &s, candidate(3, "normal"))
	v, err = s.Reconcile("normal", "candidate", "", "", "", p)
	if err != nil || v.CandidateSHA != sha(3) || v.ID != "generation" {
		t.Fatal(v, err)
	}
}
func TestPausedLegacyNormalPinMigratesWithoutResuming(t *testing.T) {
	s, p := fixture()
	register(t, &s, candidate(2, "normal"))
	s.Proposals["normal"] = Proposal{ID: "generation", Kind: "normal", State: "paused", Selection: "pinned", CandidateSHA: sha(1)}
	v, err := s.Reconcile("normal", "candidate", "", "", "", p)
	if err != nil || v.State != "paused" || v.CandidateSHA != sha(1) || v.Selection != "latest" {
		t.Fatal(v, err)
	}
	v, err = s.Reconcile("normal", "reopen", "", "", "", p)
	if err != nil || v.CandidateSHA != sha(2) || v.State != "open" {
		t.Fatal(v, err)
	}
}
func TestHotfixPinStillScopesExplicitCandidate(t *testing.T) {
	s, p := fixture()
	for _, n := range []int{1, 2} {
		c := candidate(n, "hotfix")
		c.BaselineID = s.Baseline.ID
		register(t, &s, c)
	}
	v, err := s.Reconcile("hotfix", "candidate", sha(1), "hotfix-generation", "urgent fix", p)
	if err != nil || v.CandidateSHA != sha(1) || v.Selection != "pinned" {
		t.Fatal(v, err)
	}
	v, err = s.Reconcile("hotfix", "candidate", "", "", "", p)
	if err != nil || v.CandidateSHA != sha(1) {
		t.Fatal(v, err)
	}
}
func TestControlOnlyDoesNotCreateRecursiveProposal(t *testing.T) {
	s, p := fixture()
	c := candidate(1, "normal")
	c.ControlOnly = true
	register(t, &s, c)
	if _, err := s.SelectCandidate("normal", "", p); err == nil {
		t.Fatal("control-only proposal created")
	}
}
func TestFailureBaselineAndExactRetry(t *testing.T) {
	s, p := fixture()
	register(t, &s, candidate(1, "normal"))
	i := start(t, &s, p, "normal")
	baseline := s.Baseline.ID
	if err := s.FinishDeployment(i.ID, "unknown", "https://attempt/1", p); err == nil || s.InFlight != i.ID {
		t.Fatal("unknown outcome cleared fence")
	}
	if err := s.FinishDeployment(i.ID, "failure", "https://attempt/1", p); err != nil || s.Baseline.ID != baseline {
		t.Fatal(err)
	}
	again, err := s.StartDeployment(i.ID, p)
	if err != nil || again.Candidate.SourceSHA != i.Candidate.SourceSHA {
		t.Fatal("retry replaced intent", err)
	}
}
func TestSuccessfulDeploymentPublicationRetryAndNextGeneration(t *testing.T) {
	s, p := fixture()
	register(t, &s, candidate(1, "normal"))
	i := start(t, &s, p, "normal")
	if err := s.FinishDeployment(i.ID, "success", "https://attempt/1", p); err != nil {
		t.Fatal(err)
	}
	if s.Baseline.ID != i.ID || !s.Baseline.PublicationPending {
		t.Fatal("production not persisted before publication")
	}
	retry, err := s.StartDeployment(i.ID, p)
	if err != nil || retry.Status != "publication_pending" || s.InFlight != "" {
		t.Fatal("publication retry redeploys")
	}
	register(t, &s, candidate(2, "normal"))
	if _, err := s.Reconcile("normal", "candidate", "", "next", "", p); err == nil {
		t.Fatal("unpublished outcome silently replaced")
	}
	if err := s.MarkPublished(i.ID); err != nil {
		t.Fatal(err)
	}
	next, err := s.Reconcile("normal", "candidate", "", "next", "", p)
	if err != nil || next.ID == i.ProposalID || next.BaselineID != i.ID {
		t.Fatal(next, err)
	}
}
func TestHotfixPreservesQueueAndFencesStaleNormal(t *testing.T) {
	s, p := fixture()
	register(t, &s, candidate(1, "normal"))
	normal := propose(t, &s, p, "normal")
	normalIntent, err := s.FreezeIntent(normal, s.Candidates[normal.CandidateSHA], sha(3))
	if err != nil {
		t.Fatal(err)
	}
	hotfix := candidate(2, "hotfix")
	hotfix.BaselineID = s.Baseline.ID
	register(t, &s, hotfix)
	i := start(t, &s, p, "hotfix")
	if _, err := s.StartDeployment(normalIntent.ID, p); err == nil {
		t.Fatal("concurrent production promotion accepted")
	}
	if s.Proposals["normal"].ID != normal.ID {
		t.Fatal("normal queue overwritten")
	}
	if err := s.FinishDeployment(i.ID, "success", "https://hotfix/1", p); err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartDeployment(normalIntent.ID, p); err == nil {
		t.Fatal("stale baseline deployed")
	}
}
func TestHotfixRequiresForwardIntegrationAndDeduplicatesProvenance(t *testing.T) {
	s, p := fixture()
	hotfix := candidate(1, "hotfix")
	hotfix.BaselineID = s.Baseline.ID
	register(t, &s, hotfix)
	i := start(t, &s, p, "hotfix")
	if err := s.FinishDeployment(i.ID, "success", "https://hotfix/1", p); err != nil {
		t.Fatal(err)
	}
	c := candidate(2, "normal")
	equivalent := hotfix.Changes[0]
	equivalent.SHA = sha(3)
	c.Changes = append(c.Changes, equivalent)
	register(t, &s, c)
	p.missing[c.SourceSHA+equivalent.PatchID] = true
	if _, err := s.NewChanges(c, p); err == nil {
		t.Fatal("missing forward integration accepted")
	}
	delete(p.missing, c.SourceSHA+equivalent.PatchID)
	changes, err := s.NewChanges(c, p)
	if err != nil || len(changes) != 1 || changes[0].SHA != c.SourceSHA {
		t.Fatal(changes, err)
	}
}
func TestMergedIntentRejectsChangedArtifactAndBaseline(t *testing.T) {
	s, p := fixture()
	c := candidate(1, "normal")
	register(t, &s, c)
	v := propose(t, &s, p, "normal")
	c.Artifact.Reference = "different"
	if _, err := s.FreezeIntent(v, c, sha(3)); err == nil {
		t.Fatal("artifact switched after review")
	}
	c = s.Candidates[v.CandidateSHA]
	s.Baseline.ID = "changed"
	if _, err := s.FreezeIntent(v, c, sha(3)); err == nil {
		t.Fatal("stale review accepted")
	}
}
func TestNotesCumulativeLinksAndDirectCommit(t *testing.T) {
	s, p := fixture()
	c := candidate(3, "normal")
	c.Changes = append([]Change{}, s.Baseline.Shipped...)
	for n := 1; n <= 3; n++ {
		c.Changes = append(c.Changes, candidate(n, "normal").Changes...)
	}
	c.Changes = append(c.Changes, Change{SHA: sha(4), Version: "v0.1.4", Title: "direct change", PatchID: "direct"})
	register(t, &s, c)
	v := propose(t, &s, p, "normal")
	if strings.Contains(v.Notes, "/pull/1)") {
		t.Fatal("already shipped PR repeated")
	}
	for _, want := range []string{"## v0.1.3", "manual summary", "/tree/v0.1.1", "/pull/2", "/pull/3", "/pull/4", "/commit/" + sha(4), "Fix \\[something\\]"} {
		if !strings.Contains(v.Notes, want) {
			t.Fatalf("notes missing %s: %s", want, v.Notes)
		}
	}
}
