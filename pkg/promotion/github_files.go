package promotion

import (
	"context"
	"fmt"
	"regexp"
)

// CheckProposalFiles prevents an application/control mixed PR from deploying as
// a release declaration. Trusted paths come only from repository policy.
func (c Client) CheckProposalFiles(ctx context.Context, number int, allowed []string) error {
	count := 0
	for page := 1; ; page++ {
		var files []struct{ Filename, Status string }
		status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("pulls/%d/files?per_page=100&page=%d", number, page)), nil, &files)
		if err != nil {
			return err
		}
		if status != 200 {
			return fmt.Errorf("proposal file evidence unavailable")
		}
		for _, file := range files {
			count++
			valid := false
			for _, path := range allowed {
				if file.Filename == path {
					valid = true
				}
			}
			if !valid || file.Status == "removed" || file.Status == "renamed" {
				return fmt.Errorf("release proposal changes unapproved path %s", file.Filename)
			}
		}
		if len(files) < 100 {
			break
		}
	}
	if count == 0 {
		return fmt.Errorf("release proposal has no reviewed control diff")
	}
	return nil
}

// VerifyIssue requires a human-owned request or an authenticated generated
// hotfix tracking issue. Both must be open and assigned to the configured human.
func (c Client) VerifyIssue(ctx context.Context, number int, human string) error {
	var issue struct {
		Number      int
		Body        string
		PullRequest any `json:"pull_request"`
		State       string
		User        struct{ Login string }
		Assignees   []struct{ Login string }
	}
	status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("issues/%d", number)), nil, &issue)
	if err != nil {
		return err
	}
	generated := issue.User.Login == AutomationLogin && regexp.MustCompile(`^<!-- mint:hotfix-request:[1-9][0-9]*:[a-f0-9]{40} -->\n<!-- mint:hotfix-baseline:[^\n<>]+ -->\n`).MatchString(issue.Body)
	if status != 200 || issue.Number != number || issue.State != "open" || issue.PullRequest != nil || (issue.User.Login != human && !generated) {
		return fmt.Errorf("hotfix issue is not an open human-owned request")
	}
	for _, assignee := range issue.Assignees {
		if assignee.Login == human {
			return nil
		}
	}
	return fmt.Errorf("hotfix request must be assigned to the human")
}
