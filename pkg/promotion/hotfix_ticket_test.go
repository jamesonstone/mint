package promotion

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const linearTestTicket = "https://linear.app/acme/issue/ENG-42/restore-login"

func TestLinearHotfixIdentityRejectsMixedOrNonCanonicalRequests(t *testing.T) {
	for _, ticket := range []string{"", "http://linear.app/acme/issue/ENG-42", "https://linear.app.evil/acme/issue/ENG-42", "https://x@linear.app/acme/issue/ENG-42", "https://linear.app/acme/issue/ENG-42?token=x", "https://linear.app/acme/issue/GH-42", "https://linear.app/acme/issue/eng-42", "https://linear.app/acme/issue/ENG-0", "https://linear.app/acme/issue/ENG-42/../", "https://linear.app/acme/issue/ENG-42/bad)"} {
		if _, _, err := HotfixRequestIdentity(0, ticket); err == nil {
			t.Errorf("accepted %q", ticket)
		}
	}
	if _, _, err := HotfixRequestIdentity(42, linearTestTicket); err == nil {
		t.Fatal("mixed tracker accepted")
	}
	id, ticket, err := HotfixRequestIdentity(0, linearTestTicket)
	if err != nil || id != "ENG-42" || ticket != linearTestTicket {
		t.Fatal(id, ticket, err)
	}
}

func TestLinearHotfixPreservesProductionIsolationAndTicketProvenance(t *testing.T) {
	dir, baseline, fix := hotfixRepo(t)
	source, err := PrepareHotfix(context.Background(), HotfixOptions{WorkDir: dir, Baseline: baseline, TicketURL: linearTestTicket, Fixes: []string{fix}, CommitterName: "Human", CommitterEmail: "human@example.com"})
	if source.WorkDir != "" {
		defer os.RemoveAll(source.WorkDir)
	}
	if err != nil {
		t.Fatal(err)
	}
	if source.Branch != "ENG-42" || gitTest(t, source.WorkDir, "rev-parse", "HEAD^") != baseline.Candidate.SourceSHA {
		t.Fatal("wrong identity/base", source)
	}
	if _, err := os.Stat(filepath.Join(source.WorkDir, "queued-b.txt")); !os.IsNotExist(err) {
		t.Fatal("queued source entered hotfix")
	}
	data, err := os.ReadFile(filepath.Join(source.WorkDir, ".mint/hotfix.json"))
	if err != nil {
		t.Fatal(err)
	}
	var metadata map[string]any
	if err := json.Unmarshal(data, &metadata); err != nil {
		t.Fatal(err)
	}
	if metadata["ticket_url"] != linearTestTicket || metadata["issue"] != float64(0) {
		t.Fatal(metadata)
	}
	if len(source.PatchIDs) != 1 {
		t.Fatal("lost fix provenance")
	}
}

func TestLinearHotfixRecoveryRejectsSameIDFromOtherWorkspace(t *testing.T) {
	baseline := Baseline{ID: "base", Candidate: Candidate{SourceSHA: sha(1)}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "git/ref") {
			json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": sha(2)}})
			return
		}
		data, _ := json.Marshal(map[string]any{"baseline_id": "base", "baseline_sha": sha(1), "fixes": []string{sha(3)}, "ticket_url": linearTestTicket})
		json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "content": base64.StdEncoding.EncodeToString(data)})
	}))
	defer server.Close()
	client := Client{APIURL: server.URL, Repository: "owner/repo", Token: "test"}
	if _, err := client.RecoverHotfixSource(context.Background(), baseline, 0, []string{sha(3)}, "https://linear.app/other/issue/ENG-42/restore-login"); err == nil {
		t.Fatal("foreign workspace recovered")
	}
}

func TestLinearHotfixSourceUsesExactScopeAndRequestBase(t *testing.T) {
	component := Component{Name: "backend"}
	body := component.hotfixRequestMarker("ENG-42", "base")
	id, base, ok := component.hotfixSourceRequest(body)
	if !ok || id != "ENG-42" || base != "base" || component.hotfixRequestBase(id) != "mint-hotfix-base/backend/ENG-42" {
		t.Fatal(id, base, ok)
	}
	if _, _, ok := (Component{Name: "frontend"}).hotfixSourceRequest(body); ok {
		t.Fatal("cross-component marker accepted")
	}
}

func TestLinearHotfixPublishesSourcePRWithoutCreatingIssue(t *testing.T) {
	source := HotfixSource{Branch: "ENG-42", BaselineID: "base", SourceSHA: sha(2)}
	baseline := Baseline{ID: "base", Candidate: Candidate{SourceSHA: sha(1), Version: "v1.2.3"}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && strings.Contains(r.URL.Path, "git/ref/heads/"):
			value := sha(2)
			if strings.Contains(r.URL.Path, "mint-hotfix-base/") {
				value = sha(1)
			}
			json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": value}})
		case r.Method == "GET" && r.URL.Path == "/repos/owner/repo/pulls":
			w.Write([]byte("[]"))
		case r.Method == "POST" && r.URL.Path == "/repos/owner/repo/pulls":
			var request map[string]any
			json.NewDecoder(r.Body).Decode(&request)
			if request["head"] != "ENG-42" || request["base"] != "mint-hotfix-base/ENG-42" || !strings.Contains(request["body"].(string), linearTestTicket) {
				t.Error("lost ticket or source identity", request)
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(PullRequest{Number: 9})
		case r.Method == "POST" && r.URL.Path == "/repos/owner/repo/issues/9/assignees":
			w.WriteHeader(http.StatusCreated)
		default:
			t.Error("unexpected request, including issue creation", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	client := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo", HumanLogin: "human"}
	if _, err := client.PublishHotfixSource(context.Background(), source, baseline, 0, linearTestTicket); err != nil {
		t.Fatal(err)
	}
}

func TestTicketAwareSourceBaseRejectsCrossRequestScopeAndBaseline(t *testing.T) {
	c := Component{Name: "backend"}
	body := c.hotfixRequestMarker("ENG-42", "verified")
	for _, bad := range []string{"mint-hotfix-base/backend/ENG-43", "mint-hotfix-base/frontend/ENG-42", "mint-hotfix-base/ENG-42", "mint-hotfix-base/backend/ENG-42/extra"} {
		if c.MatchesHotfixSource(body, bad, "verified") {
			t.Error("accepted wrong base", bad)
		}
	}
	if c.MatchesHotfixSource(body, "mint-hotfix-base/backend/ENG-42", "other") {
		t.Fatal("accepted stale baseline")
	}
	if !c.MatchesHotfixSource(body, "mint-hotfix-base/backend/ENG-42", "verified") {
		t.Fatal("exact source rejected")
	}
}
