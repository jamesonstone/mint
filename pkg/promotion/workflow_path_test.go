package promotion

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrustedRunAcceptsQualifiedWorkflowPathWithOtherFencesIntact(t *testing.T) {
	const workflow = ".github/workflows/build.yml"
	for _, mode := range []string{"plain", "qualified", "different-workflow", "empty-ref", "foreign-repository", "foreign-head", "pr-event", "failed"} {
		t.Run(mode, func(t *testing.T) {
			run := WorkflowRun{ID: 17, Path: workflow, HeadBranch: "trunk", Event: "push", Status: "completed", Conclusion: "success"}
			run.Repository.FullName, run.HeadRepository.FullName = "owner/repo", "owner/repo"
			switch mode {
			case "qualified":
				run.Path += "@trunk"
			case "different-workflow":
				run.Path = ".github/workflows/untrusted.yml@trunk"
			case "empty-ref":
				run.Path += "@"
			case "foreign-repository":
				run.Repository.FullName = "other/repo"
			case "foreign-head":
				run.HeadRepository.FullName = "fork/repo"
			case "pr-event":
				run.Event = "pull_request"
			case "failed":
				run.Conclusion = "failure"
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(run)
			}))
			defer server.Close()
			client := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
			_, err := client.TrustedRun(t.Context(), 17, workflow)
			if (err == nil) != (mode == "plain" || mode == "qualified") {
				t.Fatalf("trust decision %s: %v", mode, err)
			}
		})
	}
}

func TestScanDiscoversQualifiedPathOnCustomDefaultBranch(t *testing.T) {
	run := WorkflowRun{ID: 17, Path: ".github/workflows/build.yml@trunk", HeadBranch: "trunk", Event: "push", Status: "completed", Conclusion: "success"}
	run.Repository.FullName, run.HeadRepository.FullName = "owner/repo", "owner/repo"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("branch") != "trunk" {
			t.Error("lost configured default branch")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"workflow_runs": []WorkflowRun{run}})
	}))
	defer server.Close()
	client := Client{APIURL: server.URL, Token: "test", Repository: "owner/repo"}
	runs, err := client.SuccessfulBuilds(t.Context(), ".github/workflows/build.yml", "trunk")
	if err != nil || len(runs) != 1 || runs[0].ID != 17 {
		t.Fatalf("qualified workflow discovery: %v %v", runs, err)
	}
}
