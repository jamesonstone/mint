package promotion

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func genericFixture(environment string) (State, proof) {
	s, p := fixture()
	s.Schema = 2
	s.Environment = environment
	s.Configuration = "sha256:" + strings.Repeat("c", 64)
	s.Baseline.Candidate.Environment = "build"
	s.Baseline.Configuration = s.Configuration
	return s, p
}
func neutralCandidate(n int) Candidate {
	c := candidate(n, "normal")
	c.Environment = "build"
	return c
}
func bundledCandidate(t *testing.T) Candidate {
	t.Helper()
	c := neutralCandidate(1)
	c.Artifact.Reference = "immutable/bundle"
	c.Artifact.Members = map[string]ArtifactObject{
		"api": {Reference: "registry/api@sha256:a", Digest: "sha256:" + strings.Repeat("a", 64), Configuration: c.Artifact.Configuration},
		"web": {Reference: "registry/web@sha256:d", Digest: "sha256:" + strings.Repeat("d", 64), Configuration: c.Artifact.Configuration},
	}
	digest, err := BundleDigest(c.Artifact.Members)
	if err != nil {
		t.Fatal(err)
	}
	c.Artifact.Digest = digest
	return c
}
func TestNeutralBuildIsReusedWithoutRewritingAttestedCandidate(t *testing.T) {
	c := neutralCandidate(1)
	original, _ := json.Marshal(c)
	for _, environment := range []string{"stage", "production", "developer-local"} {
		s, _ := genericFixture(environment)
		register(t, &s, c)
		recorded, _ := json.Marshal(s.Candidates[c.SourceSHA])
		if string(recorded) != string(original) {
			t.Fatal("target environment rewrote build attestation", environment)
		}
	}
	legacy, _ := fixture()
	if legacy.RegisterCandidate(c, evidence(c)) == nil {
		t.Fatal("legacy journal accepted neutral build scope")
	}
	s, _ := genericFixture("stage")
	foreign := c
	foreign.Environment = "production"
	if s.RegisterCandidate(foreign, evidence(foreign)) == nil {
		t.Fatal("foreign environment candidate accepted as reusable build")
	}
}
func TestExactEnvironmentTargetCannotBeBypassedOrGuessed(t *testing.T) {
	s, p := genericFixture("stage")
	register(t, &s, neutralCandidate(1))
	register(t, &s, neutralCandidate(2))
	s.Target = "v0.1.1"
	selected, err := s.SelectCandidate("normal", "", p)
	if err != nil || selected.SourceSHA != sha(1) {
		t.Fatal(selected, err)
	}
	if _, err := s.SelectCandidate("normal", sha(2), p); err == nil {
		t.Fatal("candidate pin bypassed explicit target")
	}
	s.Target = "v9.9.9"
	if _, err := s.SelectCandidate("normal", sha(2), p); err == nil {
		t.Fatal("unbuilt target substituted with caller candidate")
	}
	ambiguous := neutralCandidate(3)
	ambiguous.Version = "v0.1.1"
	register(t, &s, ambiguous)
	s.Target = "v0.1.1"
	if _, err := s.SelectCandidate("normal", "", p); err == nil {
		t.Fatal("ambiguous version selected by map order")
	}
	empty, _ := genericFixture("stage")
	if _, err := empty.SelectCandidate("normal", "", p); !errors.Is(err, ErrNoEligible) {
		t.Fatal("missing latest build not reported", err)
	}
}
func TestConfigurationOnlyChangeCreatesExactRedeployment(t *testing.T) {
	s, p := genericFixture("stage")
	c := neutralCandidate(0)
	s.Baseline.Candidate = c
	register(t, &s, c)
	if _, err := s.SelectCandidate("normal", "", p); !errors.Is(err, ErrAlreadyDeployed) {
		t.Fatal(err)
	}
	s.Configuration = "sha256:" + strings.Repeat("d", 64)
	proposal, err := s.Reconcile("normal", "candidate", "", "stage-configuration-request", "runtime settings", p)
	if err != nil {
		t.Fatal(err)
	}
	i, err := s.FreezeIntent(proposal, c, sha(4))
	if err != nil || i.Configuration != s.Configuration {
		t.Fatal(i, err)
	}
	if _, err := s.StartDeployment(i.ID, p); err != nil {
		t.Fatal(err)
	}
	s.Configuration = "sha256:" + strings.Repeat("e", 64)
	if err := s.FinishDeployment(i.ID, "success", "https://verified/deploy", p); err != nil {
		t.Fatal(err)
	}
	if s.Baseline.RuntimeConfiguration() != i.Configuration || !reflect.DeepEqual(s.Baseline.Candidate, c) || s.Baseline.PublicationPending || s.Intents[i.ID].Status != "deployed" {
		t.Fatal("runtime/configuration drift substituted frozen intent", s.Baseline, s.Intents[i.ID])
	}
}
func TestRollbackFreezesRuntimeConfigurationAndPausesFollowing(t *testing.T) {
	s, p := genericFixture("stage")
	old := cloneBaseline(*s.Baseline)
	old.Configuration = "sha256:" + strings.Repeat("d", 64)
	s.History[old.ID] = old
	current := neutralCandidate(2)
	s.Baseline = &Baseline{ID: "production-2", Candidate: current, Configuration: s.Configuration, DeploymentURL: "https://verified/current", PreviousID: old.ID}
	proposal, err := s.ReconcileRollback(old.Candidate.SourceSHA, "stage-rollback", "Restore working build")
	if err != nil {
		t.Fatal(err)
	}
	i, err := s.FreezeIntent(proposal, s.Candidates[proposal.CandidateSHA], sha(4))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartDeployment(i.ID, p); err != nil {
		t.Fatal(err)
	}
	s.Configuration = "sha256:" + strings.Repeat("e", 64)
	if err := s.FinishDeployment(i.ID, "success", "https://verified/rollback", p); err != nil {
		t.Fatal(err)
	}
	if !s.Paused || s.Baseline.Configuration != i.Configuration || s.Baseline.Configuration == old.Configuration || s.Baseline.PublicationPending || s.InFlight != "" {
		t.Fatal("rollback lost frozen configuration or resume fence", s.Baseline, s.Paused)
	}
	register(t, &s, neutralCandidate(3))
	if _, err := s.SelectCandidate("normal", "", p); err == nil {
		t.Fatal("latest advanced immediately after rollback")
	}
}
func TestBundleIdentityBindsAllMembersAndBuildConfigurations(t *testing.T) {
	original := bundledCandidate(t)
	if err := ValidateArtifact(original.Artifact); err != nil {
		t.Fatal(err)
	}
	reversed := map[string]ArtifactObject{"web": original.Artifact.Members["web"], "api": original.Artifact.Members["api"]}
	digest, err := BundleDigest(reversed)
	if err != nil || digest != original.Artifact.Digest {
		t.Fatal("bundle identity depends on map order", digest, err)
	}
	for _, change := range []string{"missing", "digest", "configuration", "reference"} {
		t.Run(change, func(t *testing.T) {
			altered := cloneCandidate(original)
			member := altered.Artifact.Members["web"]
			switch change {
			case "missing":
				delete(altered.Artifact.Members, "web")
			case "digest":
				member.Digest = "sha256:" + strings.Repeat("e", 64)
			case "configuration":
				member.Configuration = "sha256:" + strings.Repeat("f", 64)
			case "reference":
				member.Reference = "different/object"
			}
			if change != "missing" {
				altered.Artifact.Members["web"] = member
			}
			if ValidateArtifact(altered.Artifact) == nil {
				t.Fatal("partial bundle accepted under original fingerprint")
			}
		})
	}
	if _, err := BundleDigest(nil); err == nil {
		t.Fatal("empty bundle accepted")
	}
	s, _ := genericFixture("stage")
	register(t, &s, original)
	conflicting := cloneCandidate(original)
	member := conflicting.Artifact.Members["web"]
	member.Digest = "sha256:" + strings.Repeat("e", 64)
	conflicting.Artifact.Members["web"] = member
	conflicting.Artifact.Digest, _ = BundleDigest(conflicting.Artifact.Members)
	if s.RegisterCandidate(conflicting, evidence(conflicting)) == nil {
		t.Fatal("valid different bundle replaced immutable source candidate")
	}
}
func TestMutableBundleAndProvenanceCannotChangeFrozenAuthority(t *testing.T) {
	s, p := genericFixture("stage")
	caller := bundledCandidate(t)
	original := cloneCandidate(caller)
	register(t, &s, caller)
	caller.Artifact.Members["web"] = ArtifactObject{Reference: "tampered"}
	caller.Changes[0].Title = "tampered"
	if !reflect.DeepEqual(s.Candidates[original.SourceSHA], original) {
		t.Fatal("caller modified journal candidate")
	}
	proposal := propose(t, &s, p, "normal")
	declaration, err := s.Declare(proposal)
	if err != nil {
		t.Fatal(err)
	}
	declaration.Candidate.Artifact.Members["web"] = ArtifactObject{Reference: "tampered declaration"}
	declaration.Candidate.Changes[0].Title = "tampered declaration"
	if !reflect.DeepEqual(s.Candidates[original.SourceSHA], original) {
		t.Fatal("returned declaration modified journal authority")
	}
	selected, err := s.SelectCandidate("normal", "", p)
	if err != nil {
		t.Fatal(err)
	}
	selected.Artifact.Members["web"] = ArtifactObject{Reference: "tampered selection"}
	i, err := s.FreezeIntent(proposal, s.Candidates[original.SourceSHA], sha(4))
	if err != nil {
		t.Fatal(err)
	}
	frozen := cloneIntent(i)
	i.Candidate.Artifact.Members["web"] = ArtifactObject{Reference: "tampered intent"}
	i.Candidate.Changes[0].Title = "tampered intent"
	if !reflect.DeepEqual(s.Intents[frozen.ID], frozen) || !reflect.DeepEqual(s.Candidates[original.SourceSHA], original) {
		t.Fatal("returned values modified frozen journal authority")
	}
	running, err := s.StartDeployment(frozen.ID, p)
	if err != nil {
		t.Fatal(err)
	}
	running.Candidate.Artifact.Members["web"] = ArtifactObject{Reference: "tampered running intent"}
	if err := s.FinishDeployment(frozen.ID, "success", "https://verified/deploy", p); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(s.Baseline.Candidate, original) || !reflect.DeepEqual(s.Intents[frozen.ID].Candidate, original) {
		t.Fatal("returned running intent modified baseline")
	}
}
