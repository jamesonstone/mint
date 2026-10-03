package promotion

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// PublishHotfixSource exposes the isolated fix as a human-reviewed source PR against
// an immutable production base branch. It does not merge, build or deploy it.
func (c Client) PublishHotfixSource(ctx context.Context, source HotfixSource, baseline Baseline, issue int) (PullRequest, error) {
	var none PullRequest
	if source.Branch != fmt.Sprintf("GH-%d", issue) || source.BaselineID != baseline.ID {
		return none, fmt.Errorf("invalid governed hotfix source")
	}
	baseBranch := fmt.Sprintf("mint-hotfix-base/GH-%d", issue)
	var ref gitObject
	status, err := c.request(ctx, "GET", c.repoPath("git/ref/heads/"+baseBranch), nil, &ref)
	if err != nil {
		return none, err
	}
	if status == 404 {
		status, err = c.request(ctx, "POST", c.repoPath("git/refs"), map[string]any{"ref": "refs/heads/" + baseBranch, "sha": baseline.Candidate.SourceSHA}, nil)
		if err != nil {
			return none, err
		}
		if status != 201 {
			return none, fmt.Errorf("hotfix base creation failed")
		}
	} else if status != 200 || ref.Object.SHA != baseline.Candidate.SourceSHA {
		return none, fmt.Errorf("hotfix base is not the reviewed production source")
	}
	if source.WorkDir != "" {
		var repo struct {
			CloneURL string `json:"clone_url"`
		}
		status, err = c.request(ctx, "GET", "repos/"+c.Repository, nil, &repo)
		if err != nil {
			return none, err
		}
		if status != 200 || !strings.HasPrefix(repo.CloneURL, "https://") {
			return none, fmt.Errorf("authenticated clone URL unavailable")
		}
		auth := base64.StdEncoding.EncodeToString([]byte("x-access-token:" + c.Token))
		cmd := exec.CommandContext(ctx, "git", "push", repo.CloneURL, "HEAD:refs/heads/"+source.Branch)
		cmd.Dir = source.WorkDir
		cmd.Env = append(os.Environ(), "GIT_CONFIG_COUNT=1", "GIT_CONFIG_KEY_0=http.extraheader", "GIT_CONFIG_VALUE_0=AUTHORIZATION: basic "+auth, "GIT_TERMINAL_PROMPT=0")
		if err := cmd.Run(); err != nil {
			return none, fmt.Errorf("hotfix push failed; source remains in %s", source.WorkDir)
		}
	} else {
		var remote gitObject
		status, err = c.request(ctx, "GET", c.repoPath("git/ref/heads/"+source.Branch), nil, &remote)
		if err != nil {
			return none, err
		}
		if status != 200 || remote.Object.SHA != source.SourceSHA {
			return none, fmt.Errorf("recovered hotfix source changed")
		}
	}
	marker := fmt.Sprintf("<!-- mint:hotfix-source:GH-%d:%s -->", issue, baseline.ID)
	body := marker + "\n\nIsolated production hotfix against `" + baseline.Candidate.Version + "` (`" + baseline.Candidate.SourceSHA + "`). Ordinary queued main changes are excluded.\n\n"
	for _, fix := range source.Fixes {
		body += "- Original fix: https://github.com/" + c.Repository + "/commit/" + fix + "\n"
	}
	if len(source.Fixes) == 0 {
		body += "Author only the intended fix on this branch. The baseline metadata alone is not an eligible application artifact.\n"
	}
	recovered, err := c.discoverHotfixPull(ctx, source.Branch, baseBranch, marker)
	if err != nil {
		return none, err
	}
	var pull PullRequest
	if recovered != nil {
		pull = *recovered
	} else {
		status, err = c.request(ctx, "POST", c.repoPath("pulls"), map[string]any{"title": fmt.Sprintf("fix(GH-%d): :bug: isolated production hotfix", issue), "head": source.Branch, "base": baseBranch, "body": body, "draft": false}, &pull)
		if err != nil {
			return none, err
		}
		if status != 201 {
			return none, fmt.Errorf("hotfix source PR creation failed")
		}
	}
	status, err = c.request(ctx, "POST", c.repoPath(fmt.Sprintf("issues/%d/assignees", pull.Number)), map[string]any{"assignees": []string{c.HumanLogin}}, nil)
	if err != nil {
		return none, err
	}
	if status != 201 {
		return none, fmt.Errorf("hotfix source PR assignment failed")
	}
	return pull, nil
}

// VerifyHotfixSource authenticates the source review independently of its label.
func (c Client) VerifyHotfixSource(ctx context.Context, candidate Candidate, required []string) error {
	pr, err := c.Pull(ctx, candidate.SourcePR)
	if err != nil {
		return err
	}
	if !pr.Merged || pr.MergeSHA != candidate.SourceSHA || !c.IsReleaseAuthor(pr.User.Login) || !strings.HasPrefix(pr.Base.Ref, "mint-hotfix-base/GH-") {
		return fmt.Errorf("hotfix has no independently reviewed production-base source PR")
	}
	return c.CheckHead(ctx, pr, required)
}
