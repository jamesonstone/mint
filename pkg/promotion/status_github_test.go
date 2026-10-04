package promotion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func outcomeFixture() (Intent, PullRequest) {
	i := Intent{ID: "intent-1", ProposalID: "proposal-1", Kind: "normal", MergeSHA: sha(4), Candidate: candidate(1, "normal"), Status: "publication_pending", DeploymentURL: "https://github.com/owner/repo/actions/runs/7"}
	p := PullRequest{Number: 42, State: "closed", Merged: true, MergeSHA: i.MergeSHA, Body: proposalMarker(Proposal{ID: i.ProposalID, Kind: i.Kind}) + "\nHuman release summary."}
	mergedAt := "2026-10-04T10:00:00Z"
	p.MergedAt = &mergedAt
	p.User.Login = AutomationLogin
	p.Head.Repo.FullName = "owner/repo"
	p.Base.Repo.FullName = "owner/repo"
	p.Base.Ref = "main"
	return i, p
}
func commentFixture(id int64, body, login string) outcomeComment {
	c := outcomeComment{ID: id, Body: body}
	c.User.Login = login
	return c
}
func TestOutcomeUpdatesOriginalPRIdempotentlyAndPreservesHumanComments(t *testing.T) {
	i, p := outcomeFixture()
	human := "<!-- mint:outcome:intent-1 -->\nHuman observation: please investigate."
	comments := []outcomeComment{commentFixture(1, human, "jamesonstone"), commentFixture(2, "Regular discussion.", "reviewer")}
	posts, patches := 0, 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && strings.HasSuffix(r.URL.Path, "/pulls"):
			_ = json.NewEncoder(w).Encode([]PullRequest{p})
		case r.Method == "GET" && r.URL.Path == "/repos/owner/repo/issues/42/comments":
			_ = json.NewEncoder(w).Encode(comments)
		case r.Method == "POST" && r.URL.Path == "/repos/owner/repo/issues/42/comments":
			posts++
			var body struct{ Body string }
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			comments = append(comments, commentFixture(3, body.Body, AutomationLogin))
			w.WriteHeader(201)
		case r.Method == "PATCH" && r.URL.Path == "/repos/owner/repo/issues/comments/3":
			patches++
			var body struct{ Body string }
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			comments[2].Body = body.Body
		default:
			t.Errorf("unexpected API mutation or read: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(500)
		}
	}))
	defer server.Close()
	c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo", HumanLogin: "jamesonstone"}
	cfg := Config{DefaultBranch: "main"}
	for n := 0; n < 2; n++ {
		got, err := c.SyncStatus(context.Background(), cfg, i)
		if err != nil || got.Number != 42 {
			t.Fatal(got, err)
		}
	}
	if posts != 1 || patches != 0 {
		t.Fatal("duplicate callback mutated comments", posts, patches)
	}
	i.Status = "deployed"
	for n := 0; n < 2; n++ {
		if _, err := c.SyncStatus(context.Background(), cfg, i); err != nil {
			t.Fatal(err)
		}
	}
	if posts != 1 || patches != 1 || comments[0].Body != human || comments[1].Body != "Regular discussion." {
		t.Fatal("human comments changed or outcome duplicated", comments)
	}
	if !strings.Contains(comments[2].Body, "**deployed**") {
		t.Fatal("outcome not updated", comments[2])
	}
}

func TestOutcomeRefusesAmbiguousOrForeignEvidenceWithoutMutation(t *testing.T) {
	for _, scenario := range []string{"duplicate-comment", "wrong-merge", "wrong-marker", "foreign-head", "wrong-base", "duplicate-pr"} {
		t.Run(scenario, func(t *testing.T) {
			i, p := outcomeFixture()
			comments := []outcomeComment{}
			pulls := []PullRequest{p}
			switch scenario {
			case "duplicate-comment":
				comments = []outcomeComment{commentFixture(1, "<!-- mint:outcome:intent-1 -->\nold", AutomationLogin), commentFixture(2, "<!-- mint:outcome:intent-1 -->\nold", AutomationLogin)}
			case "wrong-merge":
				pulls[0].MergeSHA = sha(3)
			case "wrong-marker":
				pulls[0].Body = "human summary"
			case "foreign-head":
				pulls[0].Head.Repo.FullName = "other/repo"
			case "wrong-base":
				pulls[0].Base.Ref = "other"
			case "duplicate-pr":
				other := p
				other.Number = 43
				pulls = append(pulls, other)
			}
			writes := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					writes++
					w.WriteHeader(500)
					return
				}
				if strings.HasSuffix(r.URL.Path, "/pulls") {
					_ = json.NewEncoder(w).Encode(pulls)
				} else {
					_ = json.NewEncoder(w).Encode(comments)
				}
			}))
			defer server.Close()
			c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
			if _, err := c.SyncStatus(context.Background(), Config{DefaultBranch: "main"}, i); err == nil || writes != 0 {
				t.Fatal("ambiguous outcome accepted", err, writes)
			}
		})
	}
}
