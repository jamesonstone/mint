package promotion

import (
	"context"
	"fmt"
	"strings"
)

// SyncStatus projects journal outcomes onto the original merged release PR.
// Only its machine-owned comment changes; no extra PR or source commit is made.
func (c Client) SyncStatus(ctx context.Context, cfg Config, i Intent) (PullRequest, error) {
	var none PullRequest
	if i.Status != "deployed" && i.Status != "publication_pending" && i.Status != "deployment_failed" {
		return none, fmt.Errorf("deployment outcome is not observed")
	}
	if i.ID == "" || i.ProposalID == "" || i.DeploymentURL == "" {
		return none, fmt.Errorf("deployment outcome identity or evidence is missing")
	}
	pulls, err := c.MergedPRs(ctx, i.MergeSHA)
	if err != nil {
		return none, err
	}
	kind := i.Kind
	if kind == "" {
		kind = i.Candidate.Kind
	}
	marker := proposalMarker(Proposal{ID: i.ProposalID, Kind: kind})
	var pull *PullRequest
	for _, pr := range pulls {
		if pr.MergeSHA != i.MergeSHA || !strings.Contains(pr.Body, marker) {
			continue
		}
		if !c.IsReleaseAuthor(pr.User.Login) || pr.Head.Repo.FullName != c.Repository || pr.Base.Ref != cfg.DefaultBranch {
			return none, fmt.Errorf("release outcome PR has a foreign identity")
		}
		if pull != nil && pull.Number != pr.Number {
			return none, fmt.Errorf("multiple merged PRs match release outcome")
		}
		copy := pr
		pull = &copy
	}
	if pull == nil {
		return none, fmt.Errorf("original merged release PR unavailable")
	}
	outcomeMarker := "<!-- mint:outcome:" + i.ID + " -->"
	body := outcomeMarker + "\n\nDeployment status: **" + i.Status + "**. [Workflow evidence](" + i.DeploymentURL + ").\n\nThe release journal is authoritative. This comment records the outcome of this release; it does not initiate deployment."
	var existing *outcomeComment
	for page := 1; ; page++ {
		var comments []outcomeComment
		status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("issues/%d/comments?per_page=100&page=%d", pull.Number, page)), nil, &comments)
		if err != nil {
			return none, err
		}
		if status != 200 {
			return none, fmt.Errorf("release outcome comments unavailable")
		}
		for _, comment := range comments {
			if comment.User.Login != AutomationLogin || !strings.HasPrefix(comment.Body, outcomeMarker+"\n") {
				continue
			}
			if existing != nil && existing.ID != comment.ID {
				return none, fmt.Errorf("multiple machine-owned release outcome comments")
			}
			copy := comment
			existing = &copy
		}
		if len(comments) < 100 {
			break
		}
	}
	if existing != nil && existing.Body == body {
		return *pull, nil
	}
	method, path := "POST", fmt.Sprintf("issues/%d/comments", pull.Number)
	if existing != nil {
		method, path = "PATCH", fmt.Sprintf("issues/comments/%d", existing.ID)
	}
	status, err := c.request(ctx, method, c.repoPath(path), map[string]string{"body": body}, nil)
	if err != nil {
		return none, err
	}
	if status != 200 && status != 201 {
		return none, fmt.Errorf("release outcome comment update failed")
	}
	return *pull, nil
}

type outcomeComment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
	User struct {
		Login string `json:"login"`
	} `json:"user"`
}
