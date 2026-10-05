package promotion

import (
	"strings"
	"testing"
)

func TestPublishingEnvironmentRecordsConfigurationRedeployWithoutRepublishingVersion(t *testing.T) {
	s, p := genericFixture("live")
	s.Publish = true
	c := neutralCandidate(1)
	register(t, &s, c)
	proposal, e := s.Reconcile("normal", "candidate", "", "new-release", "", p)
	if e != nil {
		t.Fatal(e)
	}
	i, e := s.FreezeIntent(proposal, c, sha(4))
	if e != nil || !i.Publish {
		t.Fatal(i, e)
	}
	if _, e = s.StartDeployment(i.ID, p); e != nil {
		t.Fatal(e)
	}
	if e = s.FinishDeployment(i.ID, "success", "https://run/1", p); e != nil {
		t.Fatal(e)
	}
	if e = s.MarkPublished(i.ID); e != nil {
		t.Fatal(e)
	}
	original := s.Intents[i.ID]
	s.Configuration = "sha256:" + strings.Repeat("e", 64)
	proposal, e = s.Reconcile("normal", "candidate", "", "configuration-redeploy", "", p)
	if e != nil {
		t.Fatal(e)
	}
	again, e := s.FreezeIntent(proposal, c, sha(5))
	if e != nil || again.Publish {
		t.Fatal("same version republishes conflicting canonical Release", again, e)
	}
	if _, e = s.StartDeployment(again.ID, p); e != nil {
		t.Fatal(e)
	}
	if e = s.FinishDeployment(again.ID, "success", "https://run/2", p); e != nil {
		t.Fatal(e)
	}
	if s.Intents[again.ID].Status != "deployed" || s.Baseline.PublicationPending || s.Intents[original.ID].DeploymentURL != original.DeploymentURL {
		t.Fatal("original publication changed or redeploy blocked", s.Baseline)
	}
	if s.shouldPublish(c) {
		t.Fatal("historically published version should not republish")
	}
}
