package promotion

import (
	"strings"
	"testing"
)

func TestResumeRequestReplayCannotClearLaterRollbackPause(t *testing.T) {
	s, _ := genericFixture("stage")
	s.Paused = true
	request, changed, err := s.Resume("resume-reviewed-1")
	if err != nil || !changed || s.Paused || request.BaselineID != s.Baseline.ID {
		t.Fatal(request, changed, err)
	}
	// A later verified rollback establishes another pause on a new baseline.
	s.Baseline.ID = "rollback-after-first-resume"
	s.Paused = true
	prior, changed, err := s.Resume("resume-reviewed-1")
	if err != nil || changed || !s.Paused || prior.BaselineID == s.Baseline.ID {
		t.Fatal("old resume request cleared later pause", prior, changed, err)
	}
	if _, changed, err := s.Resume("resume-reviewed-2"); err != nil || !changed || s.Paused {
		t.Fatal("new reviewed resume did not clear pause", changed, err)
	}
	if _, _, err := s.Resume(""); err == nil {
		t.Fatal("unidentified resume accepted")
	}
}
func TestSchemaTwoPinnedRequestTracksOnlyAfterExplicitResume(t *testing.T) {
	s, p := genericFixture("stage")
	register(t, &s, neutralCandidate(1))
	register(t, &s, neutralCandidate(2))
	pinned, err := s.Reconcile("normal", "candidate", sha(1), "stage-exact-request", "exact version requested", p)
	if err != nil || pinned.CandidateSHA != sha(1) || pinned.Selection != "pinned" {
		t.Fatal(pinned, err)
	}
	register(t, &s, neutralCandidate(3))
	still, err := s.Reconcile("normal", "candidate", "", "", "", p)
	if err != nil || still.CandidateSHA != sha(1) || still.Selection != "pinned" || still.ID != pinned.ID {
		t.Fatal("subsequent build changed exact reviewed request", still, err)
	}
	if _, _, err := s.Resume("resume-latest"); err != nil {
		t.Fatal(err)
	}
	latest, err := s.Reconcile("normal", "candidate", "", "", "", p)
	if err != nil || latest.CandidateSHA != sha(3) || latest.Selection != "latest" {
		t.Fatal("resume failed to restore latest tracking", latest, err)
	}
}
func TestResumePreservesExplicitEnvironmentTarget(t *testing.T) {
	s, p := genericFixture("stage")
	register(t, &s, neutralCandidate(1))
	register(t, &s, neutralCandidate(2))
	s.Target = "v0.1.1"
	requested, err := s.Reconcile("normal", "candidate", sha(1), "stage-fixed-request", "fixed environment", p)
	if err != nil {
		t.Fatal(err)
	}
	s.Paused = true
	if _, _, err := s.Resume("resume-fixed-target"); err != nil {
		t.Fatal(err)
	}
	selected, err := s.Reconcile("normal", "candidate", "", "", "", p)
	if err != nil || selected.CandidateSHA != requested.CandidateSHA || selected.Selection != "pinned" || s.Target != "v0.1.1" {
		t.Fatal("resume erased explicit target", selected, err)
	}
}
func TestPublicationPolicyIsFrozenWithIntent(t *testing.T) {
	for _, publish := range []bool{false, true} {
		t.Run(map[bool]string{false: "no-publication", true: "publication"}[publish], func(t *testing.T) {
			s, p := genericFixture("stage")
			s.Publish = publish
			s.PolicyDigest = "sha256:" + strings.Repeat("f", 64)
			c := neutralCandidate(1)
			register(t, &s, c)
			proposal := propose(t, &s, p, "normal")
			i, err := s.FreezeIntent(proposal, c, sha(4))
			if err != nil || i.Publish != publish || i.PolicyDigest != s.PolicyDigest {
				t.Fatal(i, err)
			}
			if _, err := s.StartDeployment(i.ID, p); err != nil {
				t.Fatal(err)
			}
			s.Publish = !publish
			if err := s.FinishDeployment(i.ID, "success", "https://verified/result", p); err != nil {
				t.Fatal(err)
			}
			expected := "deployed"
			if publish {
				expected = "publication_pending"
			}
			if s.Intents[i.ID].Status != expected || s.Baseline.PublicationPending != publish {
				t.Fatal("mutable state changed frozen publication decision", s.Intents[i.ID], s.Baseline)
			}
		})
	}
}
func TestAuthorityDigestExcludesRequestsButBindsTrustedPolicy(t *testing.T) {
	policy, err := ParsePolicy([]byte(environmentPolicyYAML))
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := policy.ForEnvironment("production")
	if err != nil {
		t.Fatal(err)
	}
	original, err := AuthorityDigest(cfg)
	if err != nil {
		t.Fatal(err)
	}
	requested := cfg
	requested.Target = "v9.8.7"
	requested.Follow = "latest"
	requested.Operation = "rollback"
	requested.Reason = "Recovery requested"
	same, err := AuthorityDigest(requested)
	if err != nil || same != original {
		t.Fatal("request fields changed authority digest", same, err)
	}
	mutations := map[string]func(*Config){
		"adapter":       func(c *Config) { c.PromotionWorkflow = ".github/workflows/other.yaml" },
		"observer":      func(c *Config) { c.ObservationWorkflow = ".github/workflows/other-observer.yaml" },
		"validation":    func(c *Config) { c.ValidationWorkflow = ".github/workflows/other-checks.yaml" },
		"configuration": func(c *Config) { c.Configuration = "sha256:" + strings.Repeat("b", 64) },
		"checks":        func(c *Config) { c.RequiredChecks = []string{"different"} },
		"operator":      func(c *Config) { c.HumanLogin = "different" },
		"publication":   func(c *Config) { c.Publish = !c.Publish },
		"prerequisite":  func(c *Config) { c.Requires = []string{"different"} },
		"scope":         func(c *Config) { c.Scope = "local" },
		"approval":      func(c *Config) { c.Deploy = "automatic" },
		"environment":   func(c *Config) { c.Environment = "different" },
		"build":         func(c *Config) { c.BuildWorkflow = ".github/workflows/other-build.yaml" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			altered := cfg
			mutate(&altered)
			digest, err := AuthorityDigest(altered)
			if err != nil || digest == original {
				t.Fatal("authority change omitted from frozen digest", name, digest, err)
			}
		})
	}
}
