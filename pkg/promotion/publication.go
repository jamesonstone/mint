package promotion

import (
	"context"
	"fmt"
)

// PublishIntent is production-only publication. It validates the immutable tag
// and existing Release target/content; retries do not deploy or retag source.
func (c Client) PublishIntent(ctx context.Context, i Intent) error {
	if i.Kind == "rollback" {
		return fmt.Errorf("rollback is a deployment history event; no Release publication")
	}
	if i.Status != "publication_pending" && i.Status != "deployed" {
		return fmt.Errorf("GitHub Release requires a verified production deployment")
	}
	var ref struct{ Object struct{ SHA, Type string } }
	status, err := c.request(ctx, "GET", c.repoPath("git/ref/tags/"+i.Candidate.Version), nil, &ref)
	if err != nil {
		return err
	}
	if status != 200 {
		return fmt.Errorf("existing source tag required before publication")
	}
	for depth := 0; ref.Object.Type == "tag"; depth++ {
		if depth > 4 {
			return fmt.Errorf("invalid recursive tag object")
		}
		status, err = c.request(ctx, "GET", c.repoPath("git/tags/"+ref.Object.SHA), nil, &ref)
		if err != nil {
			return err
		}
		if status != 200 {
			return fmt.Errorf("tag target evidence unavailable")
		}
	}
	if ref.Object.Type != "commit" || ref.Object.SHA != i.Candidate.SourceSHA {
		return fmt.Errorf("existing Release tag target conflicts with deployed source")
	}
	body := i.Notes + "\nDeployment evidence: " + i.DeploymentURL + "\n"
	var existing struct {
		ID                int64
		TagName           string `json:"tag_name"`
		Body              string
		Draft, Prerelease bool
	}
	status, err = c.request(ctx, "GET", c.repoPath("releases/tags/"+i.Candidate.Version), nil, &existing)
	if err != nil {
		return err
	}
	if status == 200 {
		if existing.TagName != i.Candidate.Version || existing.Body != body || existing.Draft || existing.Prerelease {
			return fmt.Errorf("existing Release content conflicts with verified production intent; preserve historical release")
		}
		return c.attachReleaseManifest(ctx, existing.ID, i, true)
	}
	if status != 404 {
		return fmt.Errorf("release lookup unavailable")
	}
	status, err = c.request(ctx, "POST", c.repoPath("releases"), map[string]any{"tag_name": i.Candidate.Version, "target_commitish": i.Candidate.SourceSHA, "name": i.Candidate.Version, "body": body, "draft": false, "prerelease": false, "make_latest": "true"}, &existing)
	if err != nil {
		return err
	}
	if status != 201 {
		return fmt.Errorf("production Release publication failed")
	}
	return c.attachReleaseManifest(ctx, existing.ID, i, true)
}

// VerifyPublication rejects local acknowledgements without matching server evidence.
func (c Client) VerifyPublication(ctx context.Context, i Intent) error {
	var r struct {
		ID                int64
		TagName           string `json:"tag_name"`
		Body              string
		Draft, Prerelease bool
	}
	status, err := c.request(ctx, "GET", c.repoPath("releases/tags/"+i.Candidate.Version), nil, &r)
	if err != nil {
		return err
	}
	if status != 200 || r.TagName != i.Candidate.Version || r.Draft || r.Prerelease || r.Body != i.Notes+"\nDeployment evidence: "+i.DeploymentURL+"\n" {
		return fmt.Errorf("production Release is not verified")
	}
	return c.attachReleaseManifest(ctx, r.ID, i, false)
}
