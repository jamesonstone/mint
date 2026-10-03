package promotion

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInstallationCredentialDoesNotRequirePersonalUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/user" {
			t.Error("installation token queried personal user")
			w.WriteHeader(403)
			return
		}
		if r.URL.Path != "/repos/owner/repo" || r.Header.Get("Authorization") != "Bearer installation-token" {
			t.Error("wrong scoped credential target")
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"full_name": "owner/repo"})
	}))
	defer server.Close()
	c := Client{APIURL: server.URL, Token: "installation-token", Repository: "owner/repo", HumanLogin: "human"}
	if err := c.VerifyRepository(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, login := range []string{"human", AutomationLogin} {
		if !c.IsReleaseAuthor(login) {
			t.Fatal("expected generated author", login)
		}
	}
	for _, login := range []string{"", "other", "other[bot]"} {
		if c.IsReleaseAuthor(login) {
			t.Fatal("foreign author allowed", login)
		}
	}
}

func TestGeneratedProposalValidationDispatchPinsBranchAndPR(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/repos/owner/repo/actions/workflows/ci.yaml/dispatches" {
			t.Error("unexpected dispatch", r.URL.Path)
		}
		var payload struct {
			Ref    string
			Inputs map[string]string
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		if payload.Ref != "GH-123" || payload.Inputs["mint_pr"] != "456" {
			t.Error("unscoped validation", payload)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()
	c := Client{APIURL: server.URL, Token: "installation-token", Repository: "owner/repo"}
	if err := c.DispatchChecks(context.Background(), "ci.yaml", "GH-123", 456); err != nil {
		t.Fatal(err)
	}
}
