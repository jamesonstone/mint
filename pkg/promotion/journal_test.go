package promotion

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJournalExactReadHumanAndCAS(t *testing.T) {
	state, _ := fixture()
	data, _ := json.Marshal(state)
	calls := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.RequestURI())
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing auth")
		}
		switch r.URL.Path {
		case "/user":
			_ = json.NewEncoder(w).Encode(map[string]string{"login": "human", "type": "User"})
		case "/repos/owner/repo/collaborators/trusted/permission":
			_ = json.NewEncoder(w).Encode(map[string]string{"permission": "write"})
		case "/repos/owner/repo/git/ref/heads/mint-release-state":
			_ = json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": sha(1)}})
		case "/repos/owner/repo/git/commits/" + sha(1):
			_ = json.NewEncoder(w).Encode(map[string]any{"tree": map[string]string{"sha": sha(2)}})
		case "/repos/owner/repo/git/trees/" + sha(2):
			_ = json.NewEncoder(w).Encode(map[string]any{"tree": []any{map[string]string{"path": journalPath, "type": "blob", "sha": sha(3)}}})
		case "/repos/owner/repo/git/blobs/" + sha(3):
			_ = json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "content": base64.StdEncoding.EncodeToString(data)})
		case "/repos/owner/repo/git/trees":
			w.WriteHeader(201)
			_ = json.NewEncoder(w).Encode(map[string]string{"sha": sha(2)})
		case "/repos/owner/repo/git/commits":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["author"].(map[string]any)["name"] != "Human" {
				t.Error("bot author")
			}
			w.WriteHeader(201)
			_ = json.NewEncoder(w).Encode(map[string]string{"sha": sha(3)})
		case "/repos/owner/repo/git/refs/heads/mint-release-state":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["force"] != false {
				t.Error("forcing journal")
			}
			w.WriteHeader(422)
		default:
			t.Error("unexpected path", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	client := Client{APIURL: server.URL, Token: "test-token", Repository: "owner/repo", HumanLogin: "human", HumanName: "Human", HumanEmail: "human@example.com"}
	if err := client.VerifyHuman(context.Background(), "trusted"); err != nil {
		t.Fatal(err)
	}
	snapshot, err := client.LoadJournal(context.Background(), "production")
	if err != nil || snapshot.Revision != sha(1) {
		t.Fatal(snapshot, err)
	}
	if _, err := client.SaveJournal(context.Background(), snapshot, "record candidate"); err == nil {
		t.Fatal("concurrent journal overwrite accepted")
	}
	count := 0
	for _, call := range calls {
		if strings.HasPrefix(call, "PATCH ") {
			count++
		}
	}
	if count != 1 {
		t.Fatal("mutation retried", calls)
	}
}
func TestForeignJournalAndBotRejected(t *testing.T) {
	s, _ := fixture()
	data, _ := json.Marshal(s)
	if _, err := DecodeState(data, "foreign/repo", "production"); err == nil {
		t.Fatal("foreign journal accepted")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"login": "human[bot]", "type": "Bot"})
	}))
	defer server.Close()
	c := Client{APIURL: server.URL, Token: "test", HumanLogin: "human", HumanName: "Human", HumanEmail: "human@example.com"}
	if c.VerifyHuman(context.Background(), "trusted") == nil {
		t.Fatal("bot credential accepted")
	}
}
func TestTrustedRunRejectsForeignUnsuccessfulAndWrongWorkflow(t *testing.T) {
	for _, mutate := range []func(*WorkflowRun){func(r *WorkflowRun) { r.Conclusion = "failure" }, func(r *WorkflowRun) { r.HeadRepository.FullName = "foreign/repo" }, func(r *WorkflowRun) { r.Path = ".github/workflows/untrusted.yaml" }, func(r *WorkflowRun) { r.Event = "pull_request" }} {
		run := WorkflowRun{ID: 1, HeadSHA: sha(1), Event: "push", Path: ".github/workflows/build.yaml", Status: "completed", Conclusion: "success"}
		run.Repository.FullName = "owner/repo"
		run.HeadRepository.FullName = "owner/repo"
		mutate(&run)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _ = json.NewEncoder(w).Encode(run) }))
		c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
		if _, err := c.TrustedRun(context.Background(), 1, ".github/workflows/build.yaml"); err == nil {
			t.Fatal("untrusted run accepted")
		}
		server.Close()
	}
}
