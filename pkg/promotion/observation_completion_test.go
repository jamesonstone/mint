package promotion

import (
	"testing"
	"time"
)

func TestVerifiedCompletionSupersedesOnlyOlderObservations(t *testing.T) {
	for _, ordering := range []string{"older", "newer", "unknown-completion"} {
		t.Run(ordering, func(t *testing.T) {
			s, p := genericFixture("stage")
			c := neutralCandidate(1)
			register(t, &s, c)
			proposal, err := s.Reconcile("normal", "candidate", "", "stage-release", "", p)
			if err != nil {
				t.Fatal(err)
			}
			intent, err := s.FreezeIntent(proposal, c, sha(4))
			if err != nil {
				t.Fatal(err)
			}
			now := time.Now().UTC()
			s.Baseline.VerifiedAt = now.Add(-time.Minute).Format(time.RFC3339Nano)
			observed := now.Add(-2 * time.Minute)
			if ordering == "newer" {
				observed = now.Add(-30 * time.Second)
			}
			if ordering == "unknown-completion" {
				s.Baseline.VerifiedAt = ""
			}
			s.Observation = &Observation{Repository: s.Repository, Environment: s.Environment, SourceSHA: sha(8), Version: "v9.9.9", Artifact: c.Artifact, Configuration: s.Configuration, Status: "live", ObservedAt: observed.Format(time.RFC3339Nano), RunID: 7, RunURL: "https://run/7"}
			_, err = s.StartDeployment(intent.ID, p)
			if ordering == "older" {
				if err != nil {
					t.Fatal("superseded observation blocked next deployment", err)
				}
				s.InFlight = ""
				if status := ProjectEnvironmentStatus(s, nil, s.Configuration); status.Status != "unknown" || status.Live == nil {
					t.Fatal("superseded evidence was deleted or reported current", status)
				}
			} else if err == nil {
				t.Fatal("newer or unsequenced drift did not retain fence")
			}
		})
	}
}
