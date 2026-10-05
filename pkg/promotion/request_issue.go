package promotion

import (
	"context"
	"fmt"
	"strings"
)

// EnsureHotfixRequestIssue allocates an isolated lane, rather than overwriting
// the original fix's GH issue branch. PR/merge identity deduplicates CI callbacks.
func (c Client) EnsureHotfixRequestIssue(ctx context.Context, pr PullRequest, baseline Baseline, reason string) (int, error) {
	if pr.Number <= 0 || !shaPattern.MatchString(pr.MergeSHA) || baseline.ID == "" {
		return 0, fmt.Errorf("hotfix request requires exact reviewed PR and production identity")
	}
	marker := fmt.Sprintf("<!-- mint:hotfix-request:%d:%s -->", pr.Number, pr.MergeSHA)
	baseMarker := "<!-- mint:hotfix-baseline:" + baseline.ID + " -->"
	found := 0
	for page := 1; ; page++ {
		var issues []struct {
			Number      int
			Body        string
			User        struct{ Login string }
			PullRequest any `json:"pull_request"`
		}
		status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("issues?state=all&per_page=100&page=%d", page)), nil, &issues)
		if err != nil {
			return 0, err
		}
		if status != 200 {
			return 0, fmt.Errorf("hotfix request discovery unavailable")
		}
		for _, issue := range issues {
			if issue.PullRequest != nil || !strings.HasPrefix(issue.Body, marker+"\n") {
				continue
			}
			if !c.IsReleaseAuthor(issue.User.Login) || !strings.Contains(issue.Body, baseMarker) {
				return 0, fmt.Errorf("registered hotfix request has foreign or stale production identity; inspect its existing issue")
			}
			if found != 0 && found != issue.Number {
				return 0, fmt.Errorf("ambiguous hotfix request issues")
			}
			found = issue.Number
		}
		if len(issues) < 100 {
			break
		}
	}
	if found != 0 {
		return found, nil
	}
	var issue struct{ Number int }
	body := marker + "\n" + baseMarker + "\n\nIsolate reviewed fix [#" + fmt.Sprint(pr.Number) + "](https://github.com/" + c.Repository + "/pull/" + fmt.Sprint(pr.Number) + ") against production `" + baseline.Candidate.Version + "`.\n\nReason: " + escapeMarkdown(reason) + "\n\nPreparation does not approve, merge, or deploy."
	status, err := c.request(ctx, "POST", c.repoPath("issues"), map[string]any{"title": fmt.Sprintf("Production hotfix for PR #%d", pr.Number), "body": body, "assignees": []string{c.HumanLogin}}, &issue)
	if err != nil {
		return 0, err
	}
	if status != 201 || issue.Number <= 0 {
		return 0, fmt.Errorf("hotfix request issue creation failed")
	}
	return issue.Number, nil
}
