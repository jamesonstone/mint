package promotion

import (
	"context"
	"fmt"
	"net/url"
)

// PullRequest describes exact review identity and merged release declaration.
type PullRequest struct {
	Number   int     `json:"number"`
	State    string  `json:"state"`
	Body     string  `json:"body"`
	Merged   bool    `json:"merged"`
	MergeSHA string  `json:"merge_commit_sha"`
	MergedAt *string `json:"merged_at"`
	User     struct {
		Login string `json:"login"`
	} `json:"user"`
	Head struct {
		SHA, Ref string
		Repo     struct {
			FullName string `json:"full_name"`
		} `json:"repo"`
	} `json:"head"`
	Base struct {
		Ref  string
		Repo struct {
			FullName string `json:"full_name"`
		} `json:"repo"`
	} `json:"base"`
}

// Pull reads the durable PR number; title parsing is never identity.
func (c Client) Pull(ctx context.Context, number int) (PullRequest, error) {
	var pr PullRequest
	status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("pulls/%d", number)), nil, &pr)
	if err != nil {
		return pr, err
	}
	if status != 200 || pr.Number != number || pr.Head.Repo.FullName != c.Repository || pr.Base.Repo.FullName != c.Repository {
		return pr, fmt.Errorf("foreign or unavailable proposal PR")
	}
	return pr, nil
}

// CheckHead requires every configured check on the current head and an
// independent approval for that head. Missing, pending and skipped never pass.
func (c Client) CheckHead(ctx context.Context, pr PullRequest, required []string) error {
	if !shaPattern.MatchString(pr.Head.SHA) {
		return fmt.Errorf("invalid proposal head")
	}
	observed := map[string]bool{}
	for page := 1; ; page++ {
		var checks struct {
			Total int                                         `json:"total_count"`
			Runs  []struct{ Name, Status, Conclusion string } `json:"check_runs"`
		}
		status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("commits/%s/check-runs?per_page=100&page=%d", pr.Head.SHA, page)), nil, &checks)
		if err != nil {
			return err
		}
		if status != 200 {
			return fmt.Errorf("head checks unavailable")
		}
		for _, check := range checks.Runs {
			passed := check.Status == "completed" && check.Conclusion == "success"
			if prior, ok := observed[check.Name]; ok {
				observed[check.Name] = prior && passed
			} else {
				observed[check.Name] = passed
			}
		}
		if page*100 >= checks.Total {
			break
		}
	}
	for _, name := range required {
		if !observed[name] {
			return fmt.Errorf("required current-head check %s has not passed", name)
		}
	}
	latest := map[string]struct{ State, Commit string }{}
	for page := 1; ; page++ {
		var reviews []struct {
			State  string
			Commit string                 `json:"commit_id"`
			User   struct{ Login string } `json:"user"`
		}
		status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("pulls/%d/reviews?per_page=100&page=%d", pr.Number, page)), nil, &reviews)
		if err != nil {
			return err
		}
		if status != 200 {
			return fmt.Errorf("reviews unavailable")
		}
		for _, review := range reviews {
			if review.State != "COMMENTED" {
				latest[review.User.Login] = struct{ State, Commit string }{review.State, review.Commit}
			}
		}
		if len(reviews) < 100 {
			break
		}
	}
	approved := false
	for actor, review := range latest {
		if review.State == "CHANGES_REQUESTED" {
			return fmt.Errorf("review changes remain requested")
		}
		if actor != pr.User.Login && review.State == "APPROVED" && review.Commit == pr.Head.SHA {
			approved = true
		}
	}
	if !approved {
		return fmt.Errorf("independent current-head approval is required")
	}
	return nil
}

// MergedPRs returns all associated merged PRs with pagination, excluding closed
// unmerged work. The adapter selects the authoritative merge for version mapping.
func (c Client) MergedPRs(ctx context.Context, sha string) ([]PullRequest, error) {
	if !shaPattern.MatchString(sha) {
		return nil, fmt.Errorf("PR mapping requires exact commit")
	}
	result := []PullRequest{}
	for page := 1; ; page++ {
		var pulls []PullRequest
		status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("commits/%s/pulls?per_page=100&page=%d", url.PathEscape(sha), page)), nil, &pulls)
		if err != nil {
			return nil, err
		}
		if status != 200 {
			return nil, fmt.Errorf("PR mapping unavailable")
		}
		for _, pr := range pulls {
			if pr.MergedAt != nil && pr.Base.Repo.FullName == c.Repository {
				result = append(result, pr)
			}
		}
		if len(pulls) < 100 {
			break
		}
	}
	return result, nil
}
