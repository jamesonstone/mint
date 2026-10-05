package promotion

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestReadObservationRequiresTrustedWorkflowAndBoundedTimestamp(t *testing.T) {
	for _, scenario := range []string{"valid", "foreign-repository", "foreign-environment", "run-id", "run-url", "branch", "workflow", "failed", "before-run", "after-run", "future", "archive-digest"} {
		t.Run(scenario, func(t *testing.T) {
			o := observationFixture()
			stamp, _ := time.Parse(time.RFC3339Nano, o.ObservedAt)
			run := WorkflowRun{ID: o.RunID, Path: ".github/workflows/observe.yml@trunk", Status: "completed", Conclusion: "success", Event: "workflow_dispatch", HeadBranch: "trunk", HTMLURL: o.RunURL, CreatedAt: stamp.Add(-time.Minute).Format(time.RFC3339Nano), UpdatedAt: stamp.Add(time.Minute).Format(time.RFC3339Nano)}
			run.Repository.FullName, run.HeadRepository.FullName = o.Repository, o.Repository
			switch scenario {
			case "foreign-repository":
				o.Repository = "other/repo"
			case "foreign-environment":
				o.Environment = "stage"
			case "run-id":
				o.RunID++
			case "run-url":
				o.RunURL = "https://untrusted/evidence"
			case "branch":
				run.HeadBranch = "unreviewed"
			case "workflow":
				run.Path = ".github/workflows/untrusted.yml"
			case "failed":
				run.Conclusion = "failure"
			case "before-run":
				o.ObservedAt = stamp.Add(-2 * time.Minute).Format(time.RFC3339Nano)
			case "after-run":
				o.ObservedAt = stamp.Add(90 * time.Second).Format(time.RFC3339Nano)
			case "future":
				o.ObservedAt = stamp.Add(time.Hour).Format(time.RFC3339Nano)
			}
			var archive bytes.Buffer
			writer := zip.NewWriter(&archive)
			file, err := writer.Create("mint-observation.json")
			if err != nil {
				t.Fatal(err)
			}
			if err := json.NewEncoder(file).Encode(o); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			digest := fmt.Sprintf("sha256:%x", sha256.Sum256(archive.Bytes()))
			if scenario == "archive-digest" {
				digest = "sha256:" + strings.Repeat("f", 64)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Error("observation attempted mutation", r.Method)
				}
				switch {
				case strings.HasSuffix(r.URL.Path, "/runs/17"):
					_ = json.NewEncoder(w).Encode(run)
				case strings.HasSuffix(r.URL.Path, "/artifacts"):
					_ = json.NewEncoder(w).Encode(map[string]any{"total_count": 1, "artifacts": []map[string]any{{"id": 9, "name": "mint-observation", "digest": digest}}})
				case strings.HasSuffix(r.URL.Path, "/zip"):
					_, _ = w.Write(archive.Bytes())
				default:
					t.Error("unexpected request", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			client := Client{APIURL: server.URL, Repository: "owner/repo", Token: "test"}
			cfg := Config{Repository: "owner/repo", Environment: "production", DefaultBranch: "trunk", ObservationWorkflow: ".github/workflows/observe.yml"}
			got, err := client.ReadObservation(t.Context(), cfg, 17)
			if (err == nil) != (scenario == "valid") {
				t.Fatal(scenario, got, err)
			}
		})
	}
}
