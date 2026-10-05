package promotion

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func observationFixture() Observation {
	return Observation{Repository: "owner/repo", Environment: "production", SourceSHA: sha(1), Version: "v1.2.3", Artifact: candidate(1, "normal").Artifact,
		Configuration: "sha256:" + strings.Repeat("d", 64), Status: "live", ObservedAt: time.Now().UTC().Add(-2 * time.Minute).Format(time.RFC3339Nano), RunID: 17, RunURL: "https://github.com/owner/repo/actions/runs/17"}
}

func TestApplyObservationIsolatedIdempotentAndMonotonic(t *testing.T) {
	o := observationFixture()
	s := NewState(o.Repository, o.Environment)
	s.Baseline = &Baseline{ID: "verified", Candidate: candidate(1, "normal")}
	s.InFlight = "unresolved-deployment"
	s.Intents[s.InFlight] = Intent{ID: s.InFlight, Status: "deploying"}
	before, _ := json.Marshal(s)
	changed, err := ApplyObservation(&s, o)
	if err != nil || !changed {
		t.Fatal(changed, err)
	}
	if changed, err := ApplyObservation(&s, o); err != nil || changed {
		t.Fatal("exact observation replay not idempotent", changed, err)
	}
	observed := s.Observation
	s.Observation = nil
	after, _ := json.Marshal(s)
	if string(before) != string(after) {
		t.Fatal("observation mutated deployment baseline, history or fence")
	}
	s.Observation = observed
	for _, scenario := range []string{"stale", "timestamp-conflict", "run-conflict", "foreign-repository", "foreign-environment", "future", "unsupported-status", "missing-live-artifact", "malformed-artifact"} {
		t.Run(scenario, func(t *testing.T) {
			bad := o
			switch scenario {
			case "stale":
				bad.ObservedAt = time.Now().UTC().Add(-time.Hour).Format(time.RFC3339Nano)
			case "timestamp-conflict":
				bad.RunID++
				bad.SourceSHA = sha(2)
			case "run-conflict":
				bad.ObservedAt = time.Now().UTC().Add(-time.Minute).Format(time.RFC3339Nano)
			case "foreign-repository":
				bad.Repository = "other/repo"
			case "foreign-environment":
				bad.Environment = "stage"
			case "future":
				bad.ObservedAt = time.Now().UTC().Add(time.Hour).Format(time.RFC3339Nano)
			case "unsupported-status":
				bad.Status = "approved"
			case "missing-live-artifact":
				bad.Artifact = Artifact{}
			case "malformed-artifact":
				bad.Artifact.Digest = "latest"
			}
			if changed, err := ApplyObservation(&s, bad); err == nil || changed {
				t.Fatal("hostile observation accepted", changed, err)
			}
			if s.Observation != observed {
				t.Fatal("rejected evidence replaced prior observation")
			}
		})
	}
}

func TestObservationStatusSeparatesLiveDesiredAndVerified(t *testing.T) {
	o := observationFixture()
	desired := Candidate{SourceSHA: o.SourceSHA, Version: o.Version, Artifact: o.Artifact}
	s := NewState(o.Repository, o.Environment)
	s.Baseline = &Baseline{ID: "older-verified", Candidate: candidate(2, "normal")}
	if got := ProjectEnvironmentStatus(s, &desired, o.Configuration); got.Status != "unknown" || got.Live != nil {
		t.Fatal(got)
	}
	if _, err := ApplyObservation(&s, o); err != nil {
		t.Fatal(err)
	}
	got := ProjectEnvironmentStatus(s, &desired, o.Configuration)
	if got.Status != "live" || got.LastVerified.ID != "older-verified" || got.AsOf != o.ObservedAt || got.Evidence != o.RunURL {
		t.Fatal("live observation fabricated new verified baseline", got)
	}
	if got := ProjectEnvironmentStatus(s, &desired, "sha256:"+strings.Repeat("e", 64)); got.Status != "drift" {
		t.Fatal("runtime configuration drift hidden", got)
	}
	s.InFlight = "pending-verification"
	if got := ProjectEnvironmentStatus(s, &desired, o.Configuration); got.Status != "deploying" || got.InFlight == "" {
		t.Fatal("live evidence hid deployment fence", got)
	}
	s.Observation.Status = "unknown"
	if got := ProjectEnvironmentStatus(s, &desired, o.Configuration); got.Status != "unknown" || s.InFlight == "" {
		t.Fatal("missing proof turned into deployment success", got)
	}
}

func TestMixedAndUnknownObservationDoNotInventIdentity(t *testing.T) {
	for _, status := range []string{"deploying", "unknown"} {
		o := observationFixture()
		o.Status, o.SourceSHA, o.Version, o.Configuration, o.Artifact = status, "", "", "", Artifact{}
		s := NewState(o.Repository, o.Environment)
		if changed, err := ApplyObservation(&s, o); err != nil || !changed {
			t.Fatal(status, changed, err)
		}
		if got := ProjectEnvironmentStatus(s, nil, ""); got.Status != status || got.Live.SourceSHA != "" {
			t.Fatal("mixed runtime was assigned fabricated source identity", got)
		}
	}
}

func TestObservationBundleBindsAllMembersAndDetachesInput(t *testing.T) {
	o := observationFixture()
	o.Artifact.Members = map[string]ArtifactObject{
		"api": {Reference: "registry/api@sha256:immutable", Digest: "sha256:" + strings.Repeat("a", 64), Configuration: o.Artifact.Configuration},
		"ui":  {Reference: "registry/ui@sha256:immutable", Digest: "sha256:" + strings.Repeat("b", 64), Configuration: o.Artifact.Configuration},
	}
	digest, err := BundleDigest(o.Artifact.Members)
	if err != nil {
		t.Fatal(err)
	}
	o.Artifact.Digest = digest
	s := NewState(o.Repository, o.Environment)
	if changed, err := ApplyObservation(&s, o); err != nil || !changed {
		t.Fatal(changed, err)
	}
	member := o.Artifact.Members["ui"]
	member.Digest = "sha256:" + strings.Repeat("c", 64)
	o.Artifact.Members["ui"] = member
	if s.Observation.Artifact.Members["ui"].Digest == member.Digest {
		t.Fatal("caller mutated stored bundle observation")
	}
	o.RunID++
	o.ObservedAt = time.Now().UTC().Add(-time.Minute).Format(time.RFC3339Nano)
	if changed, err := ApplyObservation(&s, o); err == nil || changed {
		t.Fatal("unbound mixed bundle accepted", changed, err)
	}
	o.Artifact.Digest, err = BundleDigest(o.Artifact.Members)
	if err != nil {
		t.Fatal(err)
	}
	o.Status, o.SourceSHA, o.Version, o.Configuration = "deploying", "", "", ""
	if changed, err := ApplyObservation(&s, o); err != nil || !changed {
		t.Fatal("attested mixed rollout rejected", changed, err)
	}
	if got := ProjectEnvironmentStatus(s, nil, ""); got.Status != "deploying" {
		t.Fatal("mixed rollout projected stable live", got)
	}
}
