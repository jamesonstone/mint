package cli

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesonstone/mint/pkg/promotion"
)

func candidateFixture(t *testing.T) (productionOperation, promotion.Candidate, promotion.WorkflowRun) {
	t.Helper()
	dir := t.TempDir()
	git := func(args ...string) string {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git fixture: %v %s", err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git("init", "-b", "main")
	git("config", "user.name", "Human")
	git("config", "user.email", "human@example.com")
	if err := os.WriteFile(filepath.Join(dir, "source.txt"), []byte("baseline"), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", "source.txt")
	git("commit", "-m", "feat: baseline")
	base := git("rev-parse", "HEAD")
	git("tag", "v1.0.0")
	if err := os.WriteFile(filepath.Join(dir, "source.txt"), []byte("fixed"), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", "source.txt")
	git("commit", "-m", "fix: application correction")
	source := git("rev-parse", "HEAD")
	git("tag", "v1.0.1")
	c := promotion.Candidate{Repository: "owner/repo", Environment: "production", SourceSHA: source, Version: "v1.0.1", Kind: "normal", RunID: 7, RunURL: "untrusted payload URL", ControlOnly: true, SourceDate: "payload date", Artifact: promotion.Artifact{Reference: "immutable/object", Digest: "sha256:" + strings.Repeat("a", 64), Configuration: "sha256:" + strings.Repeat("b", 64)}}
	run := promotion.WorkflowRun{ID: 7, HeadSHA: source, HeadBranch: "main", Event: "push", Path: ".github/workflows/build.yaml", Status: "completed", Conclusion: "success", HTMLURL: "https://github.com/owner/repo/actions/runs/7"}
	run.Repository.FullName = "owner/repo"
	run.HeadRepository.FullName = "owner/repo"
	state := promotion.NewState("owner/repo", "production")
	state.MainAnchor = base
	state.Baseline = &promotion.Baseline{ID: "production-0", Candidate: promotion.Candidate{SourceSHA: base}, Shipped: []promotion.Change{}}
	return productionOperation{config: promotion.Config{Repository: "owner/repo", DefaultBranch: "main", BuildWorkflow: run.Path}, snapshot: promotion.JournalSnapshot{State: state}, flags: productionFlags{RunID: 7}, proof: promotion.GitProof{Context: context.Background(), WorkDir: dir}}, c, run
}
func manifestArchive(t *testing.T, c promotion.Candidate) ([]byte, string) {
	t.Helper()
	var archive bytes.Buffer
	zw := zip.NewWriter(&archive)
	f, err := zw.Create("mint-candidate.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(f).Encode(c); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(archive.Bytes())
	return archive.Bytes(), fmt.Sprintf("sha256:%x", sum)
}
func TestCandidateCallbacksAndScanDownloadOnceAndRebuildProvenance(t *testing.T) {
	for _, mode := range []string{"candidate", "scan"} {
		t.Run(mode, func(t *testing.T) {
			op, c, run := candidateFixture(t)
			run.Path += "@refs/heads/main"
			archive, digest := manifestArchive(t, c)
			downloads, runReads := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == "/repos/owner/repo/actions/workflows/build.yaml/runs":
					// Duplicate delivery in one scan must not duplicate the archive fetch.
					_ = json.NewEncoder(w).Encode(map[string]any{"workflow_runs": []promotion.WorkflowRun{run, run}})
				case r.URL.Path == "/repos/owner/repo/actions/runs/7":
					runReads++
					_ = json.NewEncoder(w).Encode(run)
				case r.URL.Path == "/repos/owner/repo/actions/runs/7/artifacts":
					_ = json.NewEncoder(w).Encode(map[string]any{"total_count": 1, "artifacts": []any{map[string]any{"id": 9, "name": "mint-candidate", "digest": digest}}})
				case r.URL.Path == "/repos/owner/repo/actions/artifacts/9/zip":
					downloads++
					_, _ = w.Write(archive)
				case strings.HasSuffix(r.URL.Path, "/pulls"):
					_ = json.NewEncoder(w).Encode([]any{})
				default:
					t.Error("unexpected request", r.URL.Path)
					w.WriteHeader(500)
				}
			}))
			defer server.Close()
			op.client = promotion.Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
			value, changed, err := op.execute(context.Background(), mode)
			if err != nil || !changed || downloads != 1 {
				t.Fatal(value, changed, err, downloads)
			}
			expectedRunReads := 0
			if mode == "candidate" {
				expectedRunReads = 1
			}
			if runReads != expectedRunReads {
				t.Fatal("redundant or absent run authentication", runReads)
			}
			registered := op.snapshot.State.Candidates[c.SourceSHA]
			if registered.ControlOnly || registered.RunURL != run.HTMLURL || registered.SourceDate == "payload date" || len(registered.Changes) != 1 || registered.Changes[0].Title != "fix: application correction" {
				t.Fatal("producer payload substituted for source provenance", registered)
			}
			if mode == "scan" {
				_, again, err := op.scan(context.Background())
				if err != nil || again || downloads != 1 {
					t.Fatal("known runs fetched again", again, err, downloads)
				}
			}
		})
	}
}
func TestRegistrationRejectsUntrustedRunAndIdentityConflicts(t *testing.T) {
	op, c, run := candidateFixture(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/pulls/0") {
			w.WriteHeader(404)
			return
		}
		if !strings.HasSuffix(r.URL.Path, "/pulls") {
			t.Error("unexpected provenance request", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode([]any{})
	}))
	defer server.Close()
	op.client = promotion.Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
	for _, scenario := range []string{"failed", "active", "workflow", "event", "repository", "head-repository", "branch", "run", "source", "tag", "artifact", "configuration", "candidate-repository", "candidate-environment", "hotfix-review"} {
		t.Run(scenario, func(t *testing.T) {
			candidate, observed := c, run
			switch scenario {
			case "failed":
				observed.Conclusion = "failure"
			case "active":
				observed.Status = "in_progress"
			case "workflow":
				observed.Path = ".github/workflows/other.yaml"
			case "event":
				observed.Event = "pull_request"
			case "repository":
				observed.Repository.FullName = "other/repo"
			case "head-repository":
				observed.HeadRepository.FullName = "other/repo"
			case "branch":
				observed.HeadBranch = "feature"
			case "run":
				candidate.RunID = 8
			case "source":
				observed.HeadSHA = strings.Repeat("a", 40)
			case "tag":
				candidate.Version = "v1.0.0"
			case "artifact":
				candidate.Artifact.Digest = "invalid"
			case "configuration":
				candidate.Artifact.Configuration = "invalid"
			case "candidate-repository":
				candidate.Repository = "other/repo"
			case "candidate-environment":
				candidate.Environment = "staging"
			case "hotfix-review":
				candidate.Kind = "hotfix"
				candidate.SourcePR = 0
			}
			_, _, err := op.registerCandidate(context.Background(), candidate, observed)
			if err == nil || len(op.snapshot.State.Candidates) != 0 {
				t.Fatal("untrusted candidate registered", scenario, err)
			}
		})
	}
}
func TestScanRejectsConflictingReplayOfKnownSource(t *testing.T) {
	op, c, run := candidateFixture(t)
	prior := c
	prior.RunID = 6
	op.snapshot.State.Candidates[c.SourceSHA] = prior
	c.Artifact.Reference = "conflicting/object"
	archive, digest := manifestArchive(t, c)
	downloads := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/owner/repo/actions/workflows/build.yaml/runs":
			_ = json.NewEncoder(w).Encode(map[string]any{"workflow_runs": []promotion.WorkflowRun{run}})
		case "/repos/owner/repo/actions/runs/7/artifacts":
			_ = json.NewEncoder(w).Encode(map[string]any{"total_count": 1, "artifacts": []any{map[string]any{"id": 9, "name": "mint-candidate", "digest": digest}}})
		case "/repos/owner/repo/actions/artifacts/9/zip":
			downloads++
			_, _ = w.Write(archive)
		default:
			t.Error("unexpected request", r.URL.Path)
			w.WriteHeader(500)
		}
	}))
	defer server.Close()
	op.client = promotion.Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
	_, changed, err := op.scan(context.Background())
	if err == nil || changed || downloads != 1 || op.snapshot.State.Candidates[c.SourceSHA].Artifact.Reference != prior.Artifact.Reference {
		t.Fatal("conflicting source replay skipped or replaced", changed, err, downloads)
	}
}
