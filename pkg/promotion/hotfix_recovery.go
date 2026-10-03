package promotion

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strings"
)

// RecoverHotfixSource reuses an interrupted request without manufacturing a new
// timestamped cherry-pick commit or rewriting a human's reviewed source branch.
func (c Client) RecoverHotfixSource(ctx context.Context, baseline Baseline, issue int, fixes []string) (*HotfixSource, error) {
	branch := fmt.Sprintf("GH-%d", issue)
	var ref gitObject
	status, err := c.request(ctx, "GET", c.repoPath("git/ref/heads/"+branch), nil, &ref)
	if err != nil {
		return nil, err
	}
	if status == 404 {
		return nil, nil
	}
	if status != 200 {
		return nil, fmt.Errorf("hotfix recovery ref unavailable")
	}
	data, err := c.ReadFile(ctx, ".mint/hotfix.json", ref.Object.SHA)
	if err != nil {
		return nil, fmt.Errorf("existing issue branch is not this hotfix request")
	}
	var provenance struct {
		BaselineID  string   `json:"baseline_id"`
		BaselineSHA string   `json:"baseline_sha"`
		Fixes       []string `json:"fixes"`
		PatchIDs    []string `json:"patch_ids"`
	}
	if err := json.Unmarshal(data, &provenance); err != nil {
		return nil, err
	}
	if provenance.BaselineID != baseline.ID || provenance.BaselineSHA != baseline.Candidate.SourceSHA || !slices.Equal(provenance.Fixes, fixes) {
		return nil, fmt.Errorf("existing hotfix branch has different baseline or fixes; preserve it")
	}
	return &HotfixSource{Branch: branch, SourceSHA: ref.Object.SHA, BaselineID: baseline.ID, Fixes: fixes, PatchIDs: provenance.PatchIDs}, nil
}
func (c Client) discoverHotfixPull(ctx context.Context, branch, base, marker string) (*PullRequest, error) {
	var found *PullRequest
	for page := 1; ; page++ {
		var pulls []PullRequest
		status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("pulls?state=all&head=%s&base=%s&per_page=100&page=%d", url.QueryEscape(strings.Split(c.Repository, "/")[0]+":"+branch), url.QueryEscape(base), page)), nil, &pulls)
		if err != nil {
			return nil, err
		}
		if status != 200 {
			return nil, fmt.Errorf("hotfix PR recovery unavailable")
		}
		for _, pr := range pulls {
			if pr.Head.Ref != branch || pr.Base.Ref != base {
				continue
			}
			if pr.User.Login != c.HumanLogin || pr.Head.Repo.FullName != c.Repository || pr.Body == "" || !strings.Contains(pr.Body, marker) {
				return nil, fmt.Errorf("existing hotfix PR identity differs; preserve it")
			}
			if found != nil {
				return nil, fmt.Errorf("ambiguous hotfix PR identity")
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
