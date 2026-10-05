package promotion

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func proposalMarker(p Proposal) string { return "<!-- mint:proposal:" + p.ID + ":" + p.Kind + " -->" }

// DiscoverProposal recovers interrupted API updates using stable identity across
// open/closed PRs; merged historical proposals never become a paused proposal.
func (c Client) DiscoverProposal(ctx context.Context, p Proposal) (*PullRequest, error) {
	var found *PullRequest
	for page := 1; ; page++ {
		var pulls []PullRequest
		status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("pulls?state=all&per_page=100&page=%d", page)), nil, &pulls)
		if err != nil {
			return nil, err
		}
		if status != 200 {
			return nil, fmt.Errorf("proposal discovery unavailable")
		}
		for _, pr := range pulls {
			if !strings.Contains(pr.Body, proposalMarker(p)) || pr.MergedAt != nil || pr.Merged {
				continue
			}
			if !c.IsReleaseAuthor(pr.User.Login) || pr.Head.Repo.FullName != c.Repository || pr.Base.Repo.FullName != c.Repository {
				return nil, fmt.Errorf("proposal marker is attached to a foreign identity")
			}
			if found != nil && found.Number != pr.Number {
				return nil, fmt.Errorf("multiple proposals have the same generation marker")
			}
			copy := pr
			found = &copy
		}
		if len(pulls) < 100 {
			break
		}
	}
	return found, nil
}

// EnsureProposalIssue obtains the exact governed GH issue branch idempotently.
func (c Client) EnsureProposalIssue(ctx context.Context, p Proposal) (int, error) {
	marker := proposalMarker(p)
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
			return 0, fmt.Errorf("proposal issue discovery unavailable")
		}
		for _, issue := range issues {
			if issue.PullRequest == nil && strings.Contains(issue.Body, marker) {
				if !c.IsReleaseAuthor(issue.User.Login) {
					return 0, fmt.Errorf("release issue has foreign authorship")
				}
				return issue.Number, nil
			}
		}
		if len(issues) < 100 {
			break
		}
	}
	var issue struct{ Number int }
	status, err := c.request(ctx, "POST", c.repoPath("issues"), c.withAssignments(map[string]any{"title": "Production release proposal " + p.Kind, "body": marker + "\nTracks a reviewed exact production selection. Closing the associated release PR pauses it.", "assignees": c.assignmentLogins()}), &issue)
	if err != nil {
		return 0, err
	}
	if status != 201 || issue.Number <= 0 {
		return 0, fmt.Errorf("release issue creation failed")
	}
	return issue.Number, nil
}

// SyncProposal updates only declaration, manual summary and changelog; all app
// source comes from current default branch. It never force-pushes a proposal.
func (c Client) SyncProposal(ctx context.Context, cfg Config, s *State, p Proposal) (Proposal, error) {
	discovered, err := c.DiscoverProposal(ctx, p)
	if err != nil {
		return p, err
	}
	if discovered != nil {
		p.PR = discovered.Number
		p.Branch = discovered.Head.Ref
		p.HeadSHA = discovered.Head.SHA
		if discovered.State == "closed" {
			p.State = "paused"
			s.Proposals[p.Kind] = p
			return p, nil
		}
	}
	if p.Branch == "" {
		issue, err := c.EnsureProposalIssue(ctx, p)
		if err != nil {
			return p, err
		}
		p.Branch = fmt.Sprintf("GH-%d", issue)
	}
	var base gitObject
	status, err := c.request(ctx, "GET", c.repoPath("git/ref/heads/"+cfg.DefaultBranch), nil, &base)
	if err != nil {
		return p, err
	}
	if status != 200 {
		return p, fmt.Errorf("default branch unavailable")
	}
	var baseCommit struct {
		SHA  string
		Tree gitObject
	}
	status, err = c.request(ctx, "GET", c.repoPath("git/commits/"+base.Object.SHA), nil, &baseCommit)
	if err != nil {
		return p, err
	}
	if status != 200 {
		return p, fmt.Errorf("default tree unavailable")
	}
	s.Proposals[p.Kind] = p
	declaration, err := s.Declare(p)
	if err != nil {
		return p, err
	}
	data, err := json.MarshalIndent(declaration, "", "  ")
	if err != nil {
		return p, err
	}
	changelog, err := c.ReadFile(ctx, "CHANGELOG.md", base.Object.SHA)
	if err != nil && err != ErrFileNotFound {
		return p, err
	}
	marker := "<!-- mint:entry:" + p.ID + " -->"
	block := marker + "\n" + p.Notes + "\nReviewed release proposal; deployment outcome is recorded on its release PR.\n<!-- mint:entry:end -->\n\n"
	text := string(changelog)
	if start := strings.Index(text, marker); start >= 0 {
		end := strings.Index(text[start:], "<!-- mint:entry:end -->")
		if end < 0 {
			return p, fmt.Errorf("malformed release changelog block")
		}
		text = text[:start] + text[start+end+len("<!-- mint:entry:end -->"):]
	}
	files := map[string]string{".mint/proposal.json": string(data) + "\n", ".mint/summary.md": p.Summary + "\n", "CHANGELOG.md": block + text}
	policyText, policyChanged, err := c.proposalPolicy(ctx, cfg, *s, p, base.Object.SHA)
	if err != nil {
		return p, err
	}
	if policyChanged {
		files[cfg.PolicyPath] = policyText
	}
	var ref gitObject
	status, err = c.request(ctx, "GET", c.repoPath("git/ref/heads/"+p.Branch), nil, &ref)
	if err != nil {
		return p, err
	}
	parent := base.Object.SHA
	extra := []string{}
	exists := status == 200
	if exists {
		parent = ref.Object.SHA
		extra = append(extra, base.Object.SHA)
		unchanged := true
		for path, want := range files {
			got, readErr := c.ReadFile(ctx, path, parent)
			if readErr != nil || string(got) != want {
				unchanged = false
				break
			}
		}
		if unchanged {
			p.HeadSHA = parent
			s.Proposals[p.Kind] = p
			return p, nil
		}
	}
	head, err := c.createCommit(ctx, baseCommit.Tree.SHA, files, parent, fmt.Sprintf("chore(%s): :wrench: reconcile production release proposal", p.Branch), extra...)
	if err != nil {
		return p, err
	}
	if exists {
		status, err = c.request(ctx, "PATCH", c.repoPath("git/refs/heads/"+p.Branch), map[string]any{"sha": head, "force": false}, nil)
	} else {
		status, err = c.request(ctx, "POST", c.repoPath("git/refs"), map[string]any{"ref": "refs/heads/" + p.Branch, "sha": head}, nil)
	}
	if err != nil {
		return p, err
	}
	if status != 200 && status != 201 {
		return p, fmt.Errorf("proposal ref update failed")
	}
	body := proposalMarker(p) + "\n\n" + p.Notes + "\nProduction deployment pending.\n\nBaseline: `" + p.BaselineID + "`\n\nArtifact: `" + declaration.Candidate.Artifact.Reference + "` (`" + declaration.Candidate.Artifact.Digest + "`)\n\nValidation: " + declaration.Candidate.RunURL
	payload := map[string]any{"title": "Release to " + cfg.Environment + ": " + declaration.Candidate.Version, "body": body, "head": p.Branch, "base": cfg.DefaultBranch, "draft": false}
	var pr PullRequest
	if p.PR == 0 {
		status, err = c.request(ctx, "POST", c.repoPath("pulls"), payload, &pr)
	} else {
		status, err = c.request(ctx, "PATCH", c.repoPath(fmt.Sprintf("pulls/%d", p.PR)), payload, &pr)
	}
	if err != nil {
		return p, err
	}
	if status != 200 && status != 201 {
		return p, fmt.Errorf("proposal PR update failed")
	}
	p.PR = pr.Number
	p.HeadSHA = head
	s.Proposals[p.Kind] = p
	if err := c.assign(ctx, p.PR); err != nil {
		return p, err
	}

	return p, nil
}
