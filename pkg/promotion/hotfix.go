package promotion

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// HotfixOptions identifies reviewed fixes and an exact verified production base.
// Preparation never checks out main, rewrites tags, deploys, or modifies the queue.
type HotfixOptions struct {
	WorkDir               string
	Baseline              Baseline
	Issue                 int
	Fixes                 []string
	HumanName, HumanEmail string
}

// HotfixSource records the isolated result and recoverable conflict checkout.
type HotfixSource struct {
	Branch     string   `json:"branch"`
	SourceSHA  string   `json:"source_sha"`
	BaselineID string   `json:"baseline_id"`
	Fixes      []string `json:"fixes"`
	PatchIDs   []string `json:"patch_ids"`
	WorkDir    string   `json:"-"`
}

// PrepareHotfix creates a task-owned clone from production and cherry-picks only
// explicit reviewed commits. On conflicts it preserves that clone for resolution.
func PrepareHotfix(ctx context.Context, o HotfixOptions) (HotfixSource, error) {
	result := HotfixSource{Branch: "GH-" + strconv.Itoa(o.Issue), BaselineID: o.Baseline.ID, Fixes: append([]string{}, o.Fixes...)}
	if o.Issue <= 0 || o.Baseline.ID == "" || !shaPattern.MatchString(o.Baseline.Candidate.SourceSHA) || o.HumanName == "" || o.HumanEmail == "" {
		return result, fmt.Errorf("hotfix requires a verified base, governed issue, explicit fixes and human identity")
	}
	seen := map[string]bool{}
	for _, fix := range o.Fixes {
		if !shaPattern.MatchString(fix) || seen[fix] {
			return result, fmt.Errorf("fixes must be unique exact SHAs")
		}
		seen[fix] = true
	}
	dir, err := os.MkdirTemp("", "mint-hotfix-*")
	if err != nil {
		return result, err
	}
	result.WorkDir = dir
	run := func(args ...string) (string, error) {
		cmd := exec.CommandContext(ctx, "git", args...)
		cmd.Dir = dir
		data, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("hotfix preparation failed; resolve in %s: %s", dir, strings.TrimSpace(string(data)))
		}
		return strings.TrimSpace(string(data)), nil
	}
	if _, err := run("clone", "--shared", "--no-checkout", o.WorkDir, dir); err != nil {
		return result, err
	}
	if _, err := run("config", "user.name", o.HumanName); err != nil {
		return result, err
	}
	if _, err := run("config", "user.email", o.HumanEmail); err != nil {
		return result, err
	}
	if _, err := run("checkout", "-b", result.Branch, o.Baseline.Candidate.SourceSHA); err != nil {
		return result, err
	}
	for _, fix := range o.Fixes {
		parents, err := run("rev-list", "--parents", "-n", "1", fix)
		if err != nil {
			return result, err
		}
		if len(strings.Fields(parents)) != 2 {
			return result, fmt.Errorf("selected fix %s must be a single-parent commit; isolate a reviewed patch first", fix)
		}
		proof := GitProof{Context: ctx, WorkDir: dir}
		id, err := proof.PatchID(fix)
		if err != nil {
			return result, err
		}
		result.PatchIDs = append(result.PatchIDs, id)
		if _, err := run("cherry-pick", "--no-commit", fix); err != nil {
			return result, err
		}
	}
	metadata := map[string]any{"schema_version": 1, "baseline_id": result.BaselineID, "baseline_sha": o.Baseline.Candidate.SourceSHA, "fixes": result.Fixes, "patch_ids": result.PatchIDs, "issue": o.Issue}
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return result, err
	}
	if err := os.MkdirAll(filepath.Join(dir, ".mint"), 0755); err != nil {
		return result, err
	}
	if err := os.WriteFile(filepath.Join(dir, ".mint", "hotfix.json"), data, 0644); err != nil {
		return result, err
	}
	if _, err := run("add", ".mint/hotfix.json"); err != nil {
		return result, err
	}
	if _, err := run("commit", "-m", fmt.Sprintf("fix(GH-%d): :bug: isolated production hotfix", o.Issue)); err != nil {
		return result, err
	}
	result.SourceSHA, err = run("rev-parse", "HEAD")
	if err != nil {
		return result, err
	}
	parent, err := run("rev-parse", "HEAD^")
	if err != nil || parent != o.Baseline.Candidate.SourceSHA {
		return result, fmt.Errorf("hotfix ancestry is not exactly the verified production base")
	}
	return result, nil
}
