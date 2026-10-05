package promotion

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func hotfixProvenanceServer(t *testing.T, dir string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/repos/owner/repo":
			_ = json.NewEncoder(w).Encode(map[string]string{"full_name": "owner/repo", "default_branch": "main"})
		case strings.HasPrefix(r.URL.Path, "/repos/owner/repo/contents/"):
			data := gitTest(t, dir, "show", r.URL.Query().Get("ref")+":.mint/hotfix.json")
			_ = json.NewEncoder(w).Encode(map[string]string{"encoding": "base64", "content": base64.StdEncoding.EncodeToString([]byte(data))})
		case strings.HasSuffix(r.URL.Path, "/pulls"):
			_, _ = w.Write([]byte("[]"))
		default:
			t.Error("unexpected provenance request", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestConflictHotfixResolvedPatchHasActualProvenance(t *testing.T) {
	for _, scenario := range []string{"resolved", "merged", "metadata-only", "queued-ancestry", "changed-metadata", "cleared-fixes"} {
		t.Run(scenario, func(t *testing.T) {
			dir, baseline, _ := hotfixRepo(t)
			writeTest(t, dir, "production.txt", "conflicting original fix\n")
			gitTest(t, dir, "add", "production.txt")
			gitTest(t, dir, "commit", "-m", "fix: original conflicting request")
			fix := gitTest(t, dir, "rev-parse", "HEAD")
			source, err := PrepareHotfix(context.Background(), HotfixOptions{WorkDir: dir, Baseline: baseline, Issue: 789, Fixes: []string{fix}, CommitterName: "Human", CommitterEmail: "human@example.com"})
			if source.WorkDir != "" {
				defer func() { _ = os.RemoveAll(source.WorkDir) }()
			}
			if err == nil || !source.Conflict {
				t.Fatal("fixture did not conflict", err)
			}
			gitTest(t, source.WorkDir, "reset", "--hard", source.SourceSHA)
			if scenario != "metadata-only" {
				writeTest(t, source.WorkDir, "production.txt", "reviewed resolved correction\n")
				gitTest(t, source.WorkDir, "add", "production.txt")
				if scenario == "changed-metadata" || scenario == "cleared-fixes" {
					metadata := gitTest(t, source.WorkDir, "show", "HEAD:.mint/hotfix.json")
					if scenario == "cleared-fixes" {
						var request map[string]any
						if err := json.Unmarshal([]byte(metadata), &request); err != nil {
							t.Fatal(err)
						}
						request["fixes"], request["patch_ids"] = []string{}, []string{}
						changed, err := json.MarshalIndent(request, "", "  ")
						if err != nil {
							t.Fatal(err)
						}
						metadata = string(changed)
					} else {
						metadata = strings.ReplaceAll(metadata, `"issue": 789`, `"issue": 999`)
					}
					writeTest(t, source.WorkDir, ".mint/hotfix.json", metadata)
					gitTest(t, source.WorkDir, "add", ".mint/hotfix.json")
				}
				gitTest(t, source.WorkDir, "commit", "-m", "fix: independently reviewed resolution")
			}
			if scenario == "queued-ancestry" {
				gitTest(t, source.WorkDir, "merge", "--no-ff", "-s", "ours", "origin/main", "-m", "merge queued ancestry")
			}
			if scenario == "merged" {
				gitTest(t, source.WorkDir, "checkout", "-b", "immutable-base", baseline.Candidate.SourceSHA)
				gitTest(t, source.WorkDir, "merge", "--no-ff", source.Branch, "-m", "merge reviewed resolution")
			}
			candidate := Candidate{Kind: "hotfix", Version: "v1.0.1", BaselineID: baseline.ID, SourceSHA: gitTest(t, source.WorkDir, "rev-parse", "HEAD"), SourcePR: 42}
			state := NewState("owner/repo", "production")
			state.Baseline = &baseline
			server := hotfixProvenanceServer(t, source.WorkDir)
			defer server.Close()
			client := Client{Repository: "owner/repo", APIURL: server.URL, Token: "test"}
			proof := GitProof{Context: context.Background(), WorkDir: source.WorkDir}
			changes, err := client.collectHotfixChanges(context.Background(), proof, state, candidate)
			if scenario != "resolved" && scenario != "merged" {
				if err == nil {
					t.Fatalf("unsafe resolution accepted: %s %#v", scenario, changes)
				}
				return
			}
			actual, patchErr := proof.PatchID(candidate.SourceSHA)
			original, originalErr := proof.PatchID(fix)
			if err != nil || patchErr != nil || originalErr != nil || len(changes) != 1 || changes[0].SHA != candidate.SourceSHA || changes[0].PatchID != actual || changes[0].PatchID == original {
				t.Fatal("resolved patch was credited as original", changes, err)
			}
			retained, err := proof.Retains(candidate.SourceSHA, changes[0])
			if err != nil || !retained {
				t.Fatal("actual resolution cannot be retained", err)
			}
		})
	}
}

func TestRevertHotfixCanRemovePreviouslyShippedHotfix(t *testing.T) {
	dir, baseline, shippedSHA := hotfixRepo(t)
	proof := GitProof{Context: context.Background(), WorkDir: dir}
	shippedID, err := proof.PatchID(shippedSHA)
	if err != nil {
		t.Fatal(err)
	}
	baseline.Candidate.SourceSHA = shippedSHA
	baseline.Shipped = []Change{{SHA: shippedSHA, Version: "v1.0.1", PatchID: shippedID, Hotfix: true}}
	gitTest(t, dir, "revert", "--no-edit", shippedSHA)
	revert := gitTest(t, dir, "rev-parse", "HEAD")
	source, err := PrepareHotfix(context.Background(), HotfixOptions{WorkDir: dir, Baseline: baseline, Issue: 789, Fixes: []string{revert}, CommitterName: "Human", CommitterEmail: "human@example.com"})
	if source.WorkDir != "" {
		defer func() { _ = os.RemoveAll(source.WorkDir) }()
	}
	if err != nil {
		t.Fatal(err)
	}
	server := hotfixProvenanceServer(t, source.WorkDir)
	defer server.Close()
	client := Client{Repository: "owner/repo", APIURL: server.URL, Token: "test"}
	state := NewState("owner/repo", "production")
	state.Baseline = &baseline
	proof.WorkDir = source.WorkDir
	candidate := Candidate{Kind: "hotfix", SourceSHA: source.SourceSHA, Version: "v1.0.2", BaselineID: baseline.ID}
	changes, err := client.collectHotfixChanges(context.Background(), proof, state, candidate)
	if err != nil || len(changes) != 1 || !changes[0].Revert || changes[0].Reverts != shippedID {
		t.Fatal("actual inverse patch not attributed", changes, err)
	}
	candidate.Changes = changes
	if _, err := state.NewChanges(candidate, proof); err != nil {
		t.Fatal("reviewed revert hotfix blocked", err)
	}
}
