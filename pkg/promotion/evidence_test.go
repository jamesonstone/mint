package promotion

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunManifestBindsUniqueArchiveDigestAndStrictSchema(t *testing.T) {
	for _, mode := range []string{"valid", "digest", "duplicate", "expired", "schema"} {
		t.Run(mode, func(t *testing.T) {
			var archive bytes.Buffer
			zw := zip.NewWriter(&archive)
			file, _ := zw.Create("mint-candidate.json")
			payload := `{"source_sha":"` + sha(1) + `"}`
			if mode == "schema" {
				payload = `{"injected":true}`
			}
			_, _ = file.Write([]byte(payload))
			_ = zw.Close()
			sum := sha256.Sum256(archive.Bytes())
			digest := fmt.Sprintf("sha256:%x", sum)
			if mode == "digest" {
				digest = "sha256:" + strings.Repeat("0", 64)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer test" {
					t.Error("missing API authentication")
				}
				if strings.HasSuffix(r.URL.Path, "/artifacts") {
					a := map[string]any{"id": 7, "name": "mint-candidate", "expired": mode == "expired", "digest": digest}
					list := []any{a}
					if mode == "duplicate" {
						list = append(list, a)
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"total_count": len(list), "artifacts": list})
				} else {
					_, _ = w.Write(archive.Bytes())
				}
			}))
			defer server.Close()
			c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
			var candidate Candidate
			err := c.RunManifest(context.Background(), 1, "mint-candidate", &candidate)
			if (err == nil) != (mode == "valid") {
				t.Fatal(mode, err)
			}
		})
	}
}
func TestCurrentHeadChecksPaginateAndRejectStaleReviews(t *testing.T) {
	for _, mode := range []string{"pass", "stale", "pending", "changes"} {
		t.Run(mode, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.Contains(r.URL.Path, "check-runs") {
					runs := []map[string]string{}
					if r.URL.Query().Get("page") == "1" {
						for n := 0; n < 100; n++ {
							runs = append(runs, map[string]string{"name": fmt.Sprint(n), "status": "completed", "conclusion": "success"})
						}
					} else {
						status := "completed"
						if mode == "pending" {
							status = "in_progress"
						}
						runs = append(runs, map[string]string{"name": "required", "status": status, "conclusion": "success"})
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"total_count": 101, "check_runs": runs})
				} else {
					commit := sha(1)
					if mode == "stale" {
						commit = sha(2)
					}
					state := "APPROVED"
					if mode == "changes" {
						state = "CHANGES_REQUESTED"
					}
					_ = json.NewEncoder(w).Encode([]any{map[string]any{"state": state, "commit_id": commit, "user": map[string]string{"login": "reviewer"}}})
				}
			}))
			defer server.Close()
			c := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
			var pr PullRequest
			pr.Number = 1
			pr.Head.SHA = sha(1)
			pr.User.Login = "author"
			err := c.CheckHead(context.Background(), pr, []string{"required"})
			if (err == nil) != (mode == "pass") {
				t.Fatal(mode, err)
			}
		})
	}
}
func TestCandidateRerunPreservesFirstAttestation(t *testing.T) {
	s := NewState("owner/repo", "production")
	c := candidate(1, "normal")
	if err := s.RegisterCandidate(c, evidence(c)); err != nil {
		t.Fatal(err)
	}
	c.RunID = 100
	c.RunURL = "https://build/100"
	if err := s.RegisterCandidate(c, evidence(c)); err != nil {
		t.Fatal(err)
	}
	if s.Candidates[c.SourceSHA].RunID == 100 {
		t.Fatal("first attestation replaced")
	}
	c.Artifact.Configuration = "sha256:" + strings.Repeat("c", 64)
	if s.RegisterCandidate(c, evidence(c)) == nil {
		t.Fatal("configuration conflict accepted")
	}
}
