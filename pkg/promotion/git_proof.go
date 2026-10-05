package promotion

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GitProof independently verifies ancestry and hotfix patch provenance. The
// checkout must be the selected source when checking retained shipped patches.
type GitProof struct {
	Context       context.Context
	WorkDir       string
	TargetSHA     string
	DefaultBranch string
}

func (g GitProof) git(input []byte, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(g.Context, "git", args...)
	cmd.Dir = g.WorkDir
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git proof failed: %s", strings.TrimSpace(string(out)))
	}
	return out, nil
}

// Ancestor checks source ordering without relying on SemVer or event arrival.
func (g GitProof) Ancestor(a, b string) (bool, error) {
	if !shaPattern.MatchString(a) || !shaPattern.MatchString(b) {
		return false, fmt.Errorf("ancestry requires exact SHAs")
	}
	cmd := exec.CommandContext(g.Context, "git", "merge-base", "--is-ancestor", a, b)
	cmd.Dir = g.WorkDir
	err := cmd.Run()
	if err == nil {
		return true, nil
	}
	if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 1 {
		return false, nil
	}
	return false, err
}

// Retains uses actual reverse-apply against the selected checkout. A rewritten
// fix or ambiguous equivalence fails closed and requires reviewed reconciliation.
func (g GitProof) Retains(target string, ch Change) (bool, error) {
	ancestors, err := g.Ancestor(ch.SHA, target)
	if err != nil {
		return false, err
	}
	if ancestors && !ch.Hotfix {
		return true, nil
	}
	patch, err := g.git(nil, "diff", ch.SHA+"^", ch.SHA, "--", ".", ":(exclude).mint")
	if err != nil {
		return false, err
	}
	id, err := g.git(patch, "patch-id", "--stable")
	if err != nil {
		return false, err
	}
	fields := strings.Fields(string(id))
	if len(fields) == 0 || fields[0] != ch.PatchID {
		return false, fmt.Errorf("shipped patch provenance mismatch")
	}
	dir, cleanup, err := g.materialize(target)
	if err != nil {
		return false, err
	}
	defer cleanup()
	cmd := exec.CommandContext(g.Context, "git", "apply", "--reverse", "--check")
	cmd.Dir = dir
	cmd.Stdin = bytes.NewReader(patch)
	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}

// PatchID derives logical identity from source content, not titles or annotations.
func (g GitProof) PatchID(sha string) (string, error) {
	if !shaPattern.MatchString(sha) {
		return "", fmt.Errorf("patch identity requires exact SHA")
	}
	patch, err := g.git(nil, "show", "--format=", "--first-parent", sha, "--", ".", ":(exclude).mint")
	if err != nil {
		return "", err
	}
	id, err := g.git(patch, "patch-id", "--stable")
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(id))
	if len(fields) == 0 {
		return "empty:" + sha, nil
	}
	return fields[0], nil
}

// SourceDate uses immutable commit evidence for deterministic note headings.
func (g GitProof) SourceDate(sha string) (string, error) {
	if !shaPattern.MatchString(sha) {
		return "", fmt.Errorf("source date requires exact SHA")
	}
	out, err := g.git(nil, "show", "-s", "--format=%cI", sha)
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(out))
	if len(value) < 10 {
		return "", fmt.Errorf("invalid source date")
	}
	return value[:10], nil
}
