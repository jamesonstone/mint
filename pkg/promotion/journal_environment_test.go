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

func TestEnvironmentJournalReadsAndWritesIsolatedRefs(t *testing.T) {
	for _, environment := range []string{"production", "stage", "dev-local"} {
		t.Run(environment, func(t *testing.T) {
			branch, err := JournalBranch(environment)
			if err != nil {
				t.Fatal(err)
			}
			if (environment == "production" && branch != "mint-release-state") || (environment != "production" && branch != "mint-release-state-"+environment) {
				t.Fatal(branch)
			}
			data, _ := json.Marshal(NewState("owner/repo", environment))
			calls := map[string]int{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls[r.Method+" "+r.URL.Path]++
				switch {
				case r.Method == "GET" && r.URL.Path == "/repos/owner/repo/git/ref/heads/"+branch:
					_ = json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": sha(1)}})
				case r.Method == "GET" && strings.Contains(r.URL.Path, "/git/commits/"):
					_ = json.NewEncoder(w).Encode(map[string]any{"tree": map[string]string{"sha": sha(2)}})
				case r.Method == "GET" && strings.Contains(r.URL.Path, "/git/trees/"):
					_ = json.NewEncoder(w).Encode(map[string]any{"tree": []any{map[string]string{"path": "release-state.json", "sha": sha(3), "type": "blob"}}})
				case r.Method == "GET" && strings.Contains(r.URL.Path, "/git/blobs/"):
					_ = json.NewEncoder(w).Encode(map[string]string{"content": base64.StdEncoding.EncodeToString(data), "encoding": "base64"})
				case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/git/trees"):
					w.WriteHeader(201)
					_ = json.NewEncoder(w).Encode(map[string]string{"sha": sha(2)})
				case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/git/commits"):
					w.WriteHeader(201)
					_ = json.NewEncoder(w).Encode(map[string]string{"sha": sha(4)})
				case r.Method == "PATCH" && r.URL.Path == "/repos/owner/repo/git/refs/heads/"+branch:
					var body struct {
						Force bool
						SHA   string
					}
					_ = json.NewDecoder(r.Body).Decode(&body)
					if body.Force || body.SHA != sha(4) {
						t.Error(body)
					}
				default:
					t.Error("unexpected journal request", r.Method, r.URL.Path)
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
			snapshot, err := c.LoadJournal(context.Background(), environment)
			if err != nil || snapshot.State.Environment != environment {
				t.Fatal(snapshot, err)
			}
			if _, err := c.SaveJournal(context.Background(), snapshot, "record"); err != nil {
				t.Fatal(err)
			}
			if calls["PATCH /repos/owner/repo/git/refs/heads/"+branch] != 1 {
				t.Fatal(calls)
			}
		})
	}
}
func TestJournalRejectsUnsafeOrForeignIdentityBeforeMutation(t *testing.T) {
	for _, environment := range []string{"", "../production", "stage/production", "production.lock", "production@{1}", "stage space"} {
		if _, err := JournalBranch(environment); err == nil {
			t.Fatal("unsafe ref accepted", environment)
		}
		if _, err := (Client{}).LoadJournal(context.Background(), environment); err == nil {
			t.Fatal("unsafe journal read accepted")
		}
	}
	c := Client{Repository: "owner/repo"}
	if _, err := c.SaveJournal(context.Background(), JournalSnapshot{State: NewState("other/repo", "stage")}, "write"); err == nil {
		t.Fatal("foreign journal write accepted")
	}
}
func TestEnvironmentJournalRefusesCrossEnvironmentContent(t *testing.T) {
	data, _ := json.Marshal(NewState("owner/repo", "production"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/git/ref/"):
			_ = json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": sha(1)}})
		case strings.Contains(r.URL.Path, "/git/commits/"):
			_ = json.NewEncoder(w).Encode(map[string]any{"tree": map[string]string{"sha": sha(2)}})
		case strings.Contains(r.URL.Path, "/git/trees/"):
			_ = json.NewEncoder(w).Encode(map[string]any{"tree": []any{map[string]string{"path": "release-state.json", "sha": sha(3), "type": "blob"}}})
		default:
			_ = json.NewEncoder(w).Encode(map[string]string{"content": base64.StdEncoding.EncodeToString(data), "encoding": "base64"})
		}
	}))
	defer server.Close()
	c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
	if _, err := c.LoadJournal(context.Background(), "stage"); err == nil {
		t.Fatal("foreign environment state accepted")
	}
}
