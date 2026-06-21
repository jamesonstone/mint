package release

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	// DefaultTagRemote is the Git remote used by release tag publishing.
	DefaultTagRemote = "origin"
)

const tagConflictRecovery = "inspect and correct the conflicting tag, or push a dummy commit after correction to trigger a clean release calculation"

// TagOptions configures release Git tag creation.
type TagOptions struct {
	Tag       string
	Target    string
	NotesFile string
	Remote    string
	Push      bool
	WorkDir   string
}

// TagResult describes release Git tag creation or reuse.
type TagResult struct {
	TagName   string
	TargetSHA string
	Created   bool
	Reused    bool
	Pushed    bool
}

// CreateReleaseTag creates an annotated release tag, reuses an existing tag on
// the same target commit, and never moves an existing tag.
func CreateReleaseTag(ctx context.Context, opts TagOptions) (TagResult, error) {
	normalized, err := normalizeTagOptions(opts)
	if err != nil {
		return TagResult{}, err
	}

	target, err := resolveCommit(ctx, normalized.WorkDir, normalized.Target)
	if err != nil {
		return TagResult{}, err
	}

	existing, exists, err := localTagCommit(ctx, normalized.WorkDir, normalized.Tag)
	if err != nil {
		return TagResult{}, err
	}
	if exists {
		if existing == target {
			result := TagResult{
				TagName:   normalized.Tag,
				TargetSHA: target,
				Reused:    true,
			}
			if normalized.Push {
				if err := runGitNoOutput(ctx, normalized.WorkDir, "push", normalized.Remote, "refs/tags/"+normalized.Tag); err != nil {
					return TagResult{}, err
				}
				result.Pushed = true
			}
			return result, nil
		}
		return TagResult{}, fmt.Errorf("tag %s already exists on %s, expected %s; recovery: %s", normalized.Tag, existing, target, tagConflictRecovery)
	}

	if err := createAnnotatedTag(ctx, normalized.WorkDir, normalized.Tag, target, normalized.NotesFile); err != nil {
		return TagResult{}, err
	}

	result := TagResult{
		TagName:   normalized.Tag,
		TargetSHA: target,
		Created:   true,
	}
	if normalized.Push {
		if err := runGitNoOutput(ctx, normalized.WorkDir, "push", normalized.Remote, "refs/tags/"+normalized.Tag); err != nil {
			return TagResult{}, err
		}
		result.Pushed = true
	}
	return result, nil
}

func normalizeTagOptions(opts TagOptions) (TagOptions, error) {
	normalized := opts
	normalized.Tag = strings.TrimSpace(normalized.Tag)
	normalized.Target = strings.TrimSpace(normalized.Target)
	normalized.NotesFile = strings.TrimSpace(normalized.NotesFile)
	normalized.Remote = strings.TrimSpace(normalized.Remote)

	if _, ok := parseSemVerTag(normalized.Tag); !ok {
		return TagOptions{}, validationError("tag", "must be a strict vX.Y.Z SemVer tag")
	}
	if normalized.Target == "" {
		return TagOptions{}, validationError("target", "target commitish is required")
	}
	if normalized.NotesFile == "" {
		return TagOptions{}, validationError("notes-file", "release notes file is required")
	}
	if info, err := os.Stat(normalized.NotesFile); err != nil {
		return TagOptions{}, validationError("notes-file", "cannot read release notes file: %v", err)
	} else if info.IsDir() {
		return TagOptions{}, validationError("notes-file", "must be a file")
	}
	if normalized.Remote == "" {
		normalized.Remote = DefaultTagRemote
	}
	if normalized.Push && normalized.Remote == "" {
		return TagOptions{}, validationError("remote", "remote is required when push is enabled")
	}
	return normalized, nil
}

func localTagCommit(ctx context.Context, workDir string, tag string) (string, bool, error) {
	out, err := runGit(ctx, workDir, "rev-parse", "--verify", "--quiet", "refs/tags/"+tag+"^{commit}")
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return "", false, nil
		}
		return "", false, err
	}
	return strings.TrimSpace(string(out)), true, nil
}

func createAnnotatedTag(ctx context.Context, workDir string, tag string, target string, notesFile string) error {
	env := gitTaggerEnv()
	return runGitNoOutputEnv(ctx, workDir, env, "tag", "-a", tag, target, "-F", notesFile)
}

func gitTaggerEnv() []string {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return nil
	}
	return []string{
		"GIT_COMMITTER_NAME=github-actions[bot]",
		"GIT_COMMITTER_EMAIL=41898282+github-actions[bot]@users.noreply.github.com",
	}
}
