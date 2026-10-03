package promotion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublicationRequiresVerifiedProductionAndMatchingTag(t *testing.T) {
	for _, target := range []string{sha(1), sha(2)} {
		writes := 0
		i := Intent{Candidate: candidate(1, "normal"), Notes: "## v0.1.1\n\nFix.", Status: "publication_pending", DeploymentURL: "https://deployment/1"}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/repos/owner/repo/git/ref/tags/v0.1.1":
				_ = json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"type": "commit", "sha": target}})
			case "/repos/owner/repo/releases/tags/v0.1.1":
				w.WriteHeader(404)
			case "/repos/owner/repo/releases":
				writes++
				var body map[string]any
				_ = json.NewDecoder(r.Body).Decode(&body)
				if body["tag_name"] != "v0.1.1" || body["make_latest"] != "true" {
					t.Error("wrong production release")
				}
				w.WriteHeader(201)
				_ = json.NewEncoder(w).Encode(map[string]any{"id": 42})
			case "/repos/owner/repo/releases/42/assets":
				if r.Method == "GET" {
					_ = json.NewEncoder(w).Encode([]any{})
				} else {
					writes++
					w.WriteHeader(201)
				}
			default:
				t.Error("unexpected request", r.URL.Path)
			}
		}))
		c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
		err := c.PublishIntent(context.Background(), i)
		if target == sha(1) && (err != nil || writes != 2) {
			t.Fatal(err, writes)
		}
		if target != sha(1) && (err == nil || writes != 0) {
			t.Fatal("conflicting tag released")
		}
		i.Status = "deployment_failed"
		if c.PublishIntent(context.Background(), i) == nil {
			t.Fatal("failed deployment published")
		}
		server.Close()
	}
}
func TestPublicationRetryPreservesConflictingExistingRelease(t *testing.T) {
	i := Intent{Candidate: candidate(1, "normal"), Notes: "canonical notes", Status: "publication_pending", DeploymentURL: "https://deployment/1"}
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writes++
			t.Error("historical Release overwritten")
		}
		if r.URL.Path == "/repos/owner/repo/git/ref/tags/v0.1.1" {
			_ = json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"type": "commit", "sha": sha(1)}})
		} else {
			_ = json.NewEncoder(w).Encode(map[string]any{"tag_name": "v0.1.1", "body": "legacy release before production", "draft": false, "prerelease": false})
		}
	}))
	defer server.Close()
	c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
	if c.PublishIntent(context.Background(), i) == nil || writes != 0 {
		t.Fatal("existing conflicting release accepted")
	}
}
