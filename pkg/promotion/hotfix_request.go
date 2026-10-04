package promotion

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// HotfixFixes resolves a reviewed merged source PR into only its fix commits.
// The checkout must contain full default history and refs/pull/<number>/head.
// Conflicting merge resolutions and ambiguous histories require a new reviewed
// isolated fix PR rather than guessing which queued changes belong to the fix.
func (c Client) HotfixFixes(ctx context.Context, cfg Config, number int, proof GitProof) (PullRequest, []string, error) {
	pr, err := c.Pull(ctx, number)
	if err != nil {
		return pr, nil, err
	}
	if number <= 0 || cfg.Repository != c.Repository || !pr.Merged || pr.MergedAt == nil || pr.Base.Ref != cfg.DefaultBranch || !shaPattern.MatchString(pr.MergeSHA) {
		return pr, nil, fmt.Errorf("hotfix requires a merged same-repository PR targeting %s", cfg.DefaultBranch)
	}
	if err := c.CheckHead(ctx, pr, cfg.RequiredChecks); err != nil {
		return pr, nil, err
	}
	var ref struct {
		Object struct{ SHA string } `json:"object"`
	}
	status, err := c.request(ctx, "GET", c.repoPath("git/ref/heads/"+url.PathEscape(cfg.DefaultBranch)), nil, &ref)
	if err != nil || status != 200 || !shaPattern.MatchString(ref.Object.SHA) {
		return pr, nil, fmt.Errorf("default branch evidence unavailable: %v", err)
	}
	proof.Context = ctx
	history, err := proof.git(nil, "rev-list", "--first-parent", ref.Object.SHA)
	if err != nil || !containsSHA(strings.Fields(string(history)), pr.MergeSHA) {
		return pr, nil, fmt.Errorf("merged fix is absent from verified default-branch history; fetch full main and PR head history")
	}
	commits, err := c.hotfixPRCommits(ctx, number)
	if err != nil {
		return pr, nil, err
	}
	fixes, err := resolveHotfixCommits(proof, pr, commits)
	return pr, fixes, err
}

func (c Client) hotfixPRCommits(ctx context.Context, number int) ([]string, error) {
	var commits []string
	for page := 1; ; page++ {
		var batch []struct{ SHA string }
		status, err := c.request(ctx, "GET", c.repoPath(fmt.Sprintf("pulls/%d/commits?per_page=100&page=%d", number, page)), nil, &batch)
		if err != nil || status != 200 {
			return nil, fmt.Errorf("reviewed PR commits unavailable: %v", err)
		}
		for _, commit := range batch {
			if !shaPattern.MatchString(commit.SHA) || containsSHA(commits, commit.SHA) {
				return nil, fmt.Errorf("invalid or duplicate reviewed PR commit")
			}
			commits = append(commits, commit.SHA)
		}
		// GitHub limits this endpoint to 250 commits. Refuse the limit because
		// a truncated source range cannot establish a complete reviewed patch.
		if len(commits) >= 250 {
			return nil, fmt.Errorf("hotfix PR has too many commits; prepare a smaller reviewed isolated fix")
		}
		if len(batch) < 100 {
			break
		}
	}
	return commits, nil
}

func resolveHotfixCommits(proof GitProof, pr PullRequest, commits []string) ([]string, error) {
	fail := func() ([]string, error) {
		return nil, fmt.Errorf("cannot isolate PR #%d safely; fetch its full head history or prepare a reviewed isolated fix PR (merge resolutions and non-linear histories are unsupported)", pr.Number)
	}
	if len(commits) == 0 || commits[len(commits)-1] != pr.Head.SHA {
		return fail()
	}
	sourceBase := ""
	for i, commit := range commits {
		parents, err := hotfixParents(proof, commit)
		if err != nil || len(parents) != 1 || (i > 0 && parents[0] != commits[i-1]) {
			return fail()
		}
		if i == 0 {
			sourceBase = parents[0]
		}
	}
	reviewed, err := hotfixDiff(proof, sourceBase, pr.Head.SHA)
	if err != nil || reviewed == "" {
		return fail()
	}
	parents, err := hotfixParents(proof, pr.MergeSHA)
	if err != nil || len(parents) == 0 || len(parents) > 2 {
		return fail()
	}
	merged, err := hotfixDiff(proof, parents[0], pr.MergeSHA)
	if len(parents) == 2 {
		if err != nil || parents[1] != pr.Head.SHA || merged != reviewed {
			return fail()
		}
		return commits, nil
	}
	// A squash (including a one-commit rebase) is exactly the reviewed net
	// patch on a single-parent main commit. It excludes all earlier main work.
	if err == nil && merged == reviewed {
		return []string{pr.MergeSHA}, nil
	}
	// For rebases, match the complete contiguous main range to each original
	// reviewed patch, rather than trusting commit messages or SHA counts alone.
	fixes := make([]string, len(commits))
	current := pr.MergeSHA
	for i := len(commits) - 1; i >= 0; i-- {
		parents, err := hotfixParents(proof, current)
		if err != nil || len(parents) != 1 {
			return fail()
		}
		actual, err := hotfixDiff(proof, parents[0], current)
		originalParents, parentErr := hotfixParents(proof, commits[i])
		if parentErr != nil || len(originalParents) != 1 {
			return fail()
		}
		original, originalErr := hotfixDiff(proof, originalParents[0], commits[i])
		if err != nil || originalErr != nil || actual == "" || actual != original {
			return fail()
		}
		fixes[i], current = current, parents[0]
	}
	combined, err := hotfixDiff(proof, current, pr.MergeSHA)
	if err != nil || combined != reviewed {
		return fail()
	}
	return fixes, nil
}

func hotfixParents(proof GitProof, sha string) ([]string, error) {
	if !shaPattern.MatchString(sha) {
		return nil, fmt.Errorf("invalid fix SHA")
	}
	out, err := proof.git(nil, "rev-list", "--parents", "-n", "1", sha)
	fields := strings.Fields(string(out))
	if err != nil || len(fields) == 0 || fields[0] != sha {
		return nil, fmt.Errorf("fix commit history unavailable")
	}
	return fields[1:], nil
}

// hotfixDiff retains exact patch content, modes and hunk positions. Blob index
// hashes can differ when main has unrelated changes; context is omitted, but
// whitespace or shifted hunks remain different and require isolated review.
func hotfixDiff(proof GitProof, base, head string) (string, error) {
	patch, err := proof.git(nil, "diff", "--binary", "--unified=0", "--no-renames", "--no-ext-diff", "--no-textconv", "--color=never", "--src-prefix=a/", "--dst-prefix=b/", base, head, "--")
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(patch), "\n")
	retained := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.HasPrefix(line, "index ") {
			retained = append(retained, line)
		}
	}
	return strings.TrimSuffix(strings.Join(retained, "\n"), "\n"), nil
}

func containsSHA(commits []string, sha string) bool {
	for _, commit := range commits {
		if commit == sha {
			return true
		}
	}
	return false
}
