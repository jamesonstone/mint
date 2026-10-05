package promotion

import (
	"context"
	"fmt"
	"net/url"
	"path"
)

// SuccessfulBuilds paginates successful authoritative source builds. Candidate
// replay still authenticates each run and its uploaded manifest independently.
func (c Client) SuccessfulBuilds(ctx context.Context, workflow, branch string) ([]WorkflowRun, error) {
	var result []WorkflowRun
	for page := 1; ; page++ {
		var list struct {
			Runs []WorkflowRun `json:"workflow_runs"`
		}
		status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("actions/workflows/%s/runs?branch=%s&status=success&per_page=100&page=%d", url.PathEscape(path.Base(workflow)), url.QueryEscape(branch), page)), nil, &list)
		if err != nil {
			return nil, err
		}
		if status != 200 {
			return nil, fmt.Errorf("successful build discovery unavailable")
		}
		for _, run := range list.Runs {
			if run.HeadBranch == branch && WorkflowPathMatches(run.Path, workflow) && run.Repository.FullName == c.Repository && run.HeadRepository.FullName == c.Repository && run.Status == "completed" && run.Conclusion == "success" && (run.Event == "push" || run.Event == "workflow_dispatch") {
				result = append(result, run)
			}
		}
		if len(list.Runs) < 100 {
			break
		}
	}
	return result, nil
}
