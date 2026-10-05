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

func TestBootstrapRequiresArchivedRuntimeAndIndependentSuccessfulSourceBuild(t *testing.T) {
	dir := t.TempDir()
	git := func(args ...string) string {
		command := exec.Command("git", args...)
		command.Dir = dir
		data, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("git fixture: %v %s", err, data)
		}
		return strings.TrimSpace(string(data))
	}
	git("init", "-b", "main")
	git("config", "user.name", "Human")
	git("config", "user.email", "human@example.com")
	if err := os.WriteFile(filepath.Join(dir, "source.txt"), []byte("source"), 0600); err != nil {
		t.Fatal(err)
	}
	git("add", "source.txt")
	git("commit", "-m", "feat: source")
	source := git("rev-parse", "HEAD")
	git("tag", "v1.0.0")
	for _, mode := range []string{"valid", "source", "workflow", "runtime"} {
		t.Run(mode, func(t *testing.T) {
			candidate := promotion.Candidate{Repository: "owner/repo", Environment: "production", SourceSHA: source, Version: "v1.0.0", Kind: "normal", Artifact: promotion.Artifact{Reference: "immutable/object", Digest: "sha256:" + strings.Repeat("a", 64), Configuration: "sha256:" + strings.Repeat("b", 64)}, RunID: 7, RunURL: "https://github.com/owner/repo/actions/runs/7"}
			baseline := promotion.Baseline{ID: "verified-runtime", Candidate: candidate, DeploymentURL: "https://github.com/owner/repo/actions/runs/8", Shipped: []promotion.Change{}}
			if mode == "runtime" {
				baseline.DeploymentURL = "https://arbitrary/success"
			}
			var archive bytes.Buffer
			zw := zip.NewWriter(&archive)
			file, err := zw.Create("mint-baseline.json")
			if err != nil {
				t.Fatal(err)
			}
			if err := json.NewEncoder(file).Encode(baseline); err != nil {
				t.Fatal(err)
			}
			if err := zw.Close(); err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(archive.Bytes())
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/repos/owner/repo/actions/runs/8/artifacts":
					_ = json.NewEncoder(w).Encode(map[string]any{"total_count": 1, "artifacts": []any{map[string]any{"id": 9, "name": "mint-baseline", "digest": fmt.Sprintf("sha256:%x", sum)}}})
				case "/repos/owner/repo/actions/artifacts/9/zip":
					_, _ = w.Write(archive.Bytes())
				default:
					id := int64(8)
					sha := strings.Repeat("c", 40)
					path := ".github/workflows/baseline.yaml"
					if strings.HasSuffix(r.URL.Path, "/7") {
						id = 7
						sha = source
						path = ".github/workflows/legacy-build.yaml"
						if mode == "source" {
							sha = strings.Repeat("d", 40)
						}
					} else if mode == "workflow" {
						path = ".github/workflows/foreign.yaml"
					}
					run := promotion.WorkflowRun{ID: id, HeadSHA: sha, HeadBranch: "main", Event: "workflow_dispatch", Path: path, Status: "completed", Conclusion: "success", HTMLURL: fmt.Sprintf("https://github.com/owner/repo/actions/runs/%d", id)}
					run.Repository.FullName = "owner/repo"
					run.HeadRepository.FullName = "owner/repo"
					_ = json.NewEncoder(w).Encode(run)
				}
			}))
			defer server.Close()
			operation := productionOperation{config: promotion.Config{DefaultBranch: "main", BaselineWorkflow: ".github/workflows/baseline.yaml", BaselineBuildWorkflow: ".github/workflows/legacy-build.yaml"}, client: promotion.Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}, snapshot: promotion.JournalSnapshot{State: promotion.NewState("owner/repo", "production")}, flags: productionFlags{RunID: 8, Input: "never-trust-local.json"}, proof: promotion.GitProof{Context: context.Background(), WorkDir: dir}}
			_, changed, err := operation.bootstrap(context.Background())
			if mode == "valid" {
				if err != nil || !changed || operation.snapshot.State.Baseline == nil {
					t.Fatal(err)
				}
			} else if err == nil || operation.snapshot.State.Baseline != nil {
				t.Fatal("unverified bootstrap advanced production", mode)
			}
		})
	}
}
