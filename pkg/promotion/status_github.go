package promotion

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// SyncStatus opens a control-only review PR for verified deployment history.
// It never writes directly to main or conflates a merged proposal with shipping.
func (c Client) SyncStatus(ctx context.Context, cfg Config, i Intent) (PullRequest, error) {
	var none PullRequest
	if i.Status != "deployed" && i.Status != "publication_pending" && i.Status != "deployment_failed" {
		return none, fmt.Errorf("deployment outcome is not observed")
	}
	p := Proposal{ID: i.ID + ":" + i.Status, Kind: "status"}
	existing, err := c.DiscoverProposal(ctx, p)
	if err != nil {
		return none, err
	}
	if existing != nil && existing.State == "closed" {
		return *existing, nil
	}
	var base gitObject
	status, err := c.request(ctx, "GET", c.repoPath("git/ref/heads/"+cfg.DefaultBranch), nil, &base)
	if err != nil {
		return none, err
	}
	if status != 200 {
		return none, fmt.Errorf("status base unavailable")
	}
	changelog, err := c.ReadFile(ctx, "CHANGELOG.md", base.Object.SHA)
	if err != nil {
		return none, err
	}
	marker := "<!-- mint:entry:" + i.ProposalID + " -->"
	start := strings.Index(string(changelog), marker)
	if start < 0 {
		return none, fmt.Errorf("merged changelog entry missing")
	}
	end := strings.Index(string(changelog)[start:], "<!-- mint:entry:end -->")
	if end < 0 {
		return none, fmt.Errorf("merged changelog entry malformed")
	}
	end += start
	block := string(changelog)[start:end]
	pending := "<!-- mint:deployment:"
	label := "Deployment status: **" + i.Status + "**. [Verified workflow evidence](" + i.DeploymentURL + ").\n"
	if strings.Contains(block, label) {
		return none, nil
	}
	markerStart := strings.Index(block, pending)
	if markerStart < 0 {
		return none, fmt.Errorf("changelog deployment marker missing")
	}
	text := string(changelog)[:start] + block[:markerStart] + "<!-- mint:deployment:observed -->\n" + label + string(changelog)[end:]
	data, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return none, err
	}
	issue, err := c.EnsureProposalIssue(ctx, p)
	if err != nil {
		return none, err
	}
	branch := fmt.Sprintf("GH-%d", issue)
	var baseCommit struct{ Tree gitObject }
	status, err = c.request(ctx, "GET", c.repoPath("git/commits/"+base.Object.SHA), nil, &baseCommit)
	if err != nil {
		return none, err
	}
	if status != 200 {
		return none, fmt.Errorf("status base tree unavailable")
	}
	var ref gitObject
	status, err = c.request(ctx, "GET", c.repoPath("git/ref/heads/"+branch), nil, &ref)
	if err != nil {
		return none, err
	}
	exists := status == 200
	parent := base.Object.SHA
	extra := []string{}
	if exists {
		parent = ref.Object.SHA
		extra = append(extra, base.Object.SHA)
	}
	head, err := c.createCommit(ctx, baseCommit.Tree.SHA, map[string]string{"CHANGELOG.md": text, ".mint/status.json": string(data) + "\n"}, parent, fmt.Sprintf("chore(%s): :wrench: record verified production deployment status", branch), extra...)
	if err != nil {
		return none, err
	}
	if exists {
		status, err = c.request(ctx, "PATCH", c.repoPath("git/refs/heads/"+branch), map[string]any{"sha": head, "force": false}, nil)
	} else {
		status, err = c.request(ctx, "POST", c.repoPath("git/refs"), map[string]any{"ref": "refs/heads/" + branch, "sha": head}, nil)
	}
	if err != nil {
		return none, err
	}
	if status != 200 && status != 201 {
		return none, fmt.Errorf("status ref update failed")
	}
	body := proposalMarker(p) + "\n\n" + label + "\nThis records an observed outcome; it does not deploy or publish."
	var pull PullRequest
	payload := map[string]any{"title": fmt.Sprintf("chore(GH-%d): :wrench: record %s deployment %s", issue, i.Candidate.Version, i.Status), "body": body, "head": branch, "base": cfg.DefaultBranch, "draft": false}
	if existing == nil {
		status, err = c.request(ctx, "POST", c.repoPath("pulls"), payload, &pull)
	} else {
		status, err = c.request(ctx, "PATCH", c.repoPath(fmt.Sprintf("pulls/%d", existing.Number)), payload, &pull)
	}
	if err != nil {
		return none, err
	}
	if status != 200 && status != 201 {
		return none, fmt.Errorf("status PR update failed")
	}
	_, err = c.request(ctx, "POST", c.repoPath(fmt.Sprintf("issues/%d/assignees", pull.Number)), map[string]any{"assignees": []string{c.HumanLogin}}, nil)
	if err != nil {
		return pull, err
	}
	return pull, c.DispatchChecks(ctx, cfg.ValidationWorkflow, branch, pull.Number)
}
