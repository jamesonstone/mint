package release

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func resolveCommit(ctx context.Context, workDir string, commitish string) (string, error) {
	if commitish == "" {
		commitish = DefaultCommitish
	}

	out, err := runGit(ctx, workDir, "rev-parse", "--verify", commitish+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("ref not found: %s", commitish)
	}
	return strings.TrimSpace(string(out)), nil
}

func shortCommit(ctx context.Context, workDir string, sha string) (string, error) {
	out, err := runGit(ctx, workDir, "rev-parse", "--short=12", sha)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func semVerTagsPointingAt(ctx context.Context, workDir string, target string) ([]semverTag, error) {
	out, err := runGit(ctx, workDir, "tag", "--points-at", target)
	if err != nil {
		return nil, err
	}

	var tags []semverTag
	for _, name := range splitLines(out) {
		tag, ok := parseSemVerTag(name)
		if !ok {
			continue
		}
		tag.Commit = target
		tags = append(tags, tag)
	}
	return tags, nil
}

func reachableSemVerTags(ctx context.Context, workDir string, target string) ([]semverTag, error) {
	out, err := runGit(ctx, workDir, "tag", "--merged", target)
	if err != nil {
		return nil, err
	}

	var tags []semverTag
	for _, name := range splitLines(out) {
		tag, ok := parseSemVerTag(name)
		if !ok {
			continue
		}
		commit, err := tagCommit(ctx, workDir, name)
		if err != nil {
			return nil, err
		}
		tag.Commit = commit
		tags = append(tags, tag)
	}
	return tags, nil
}

func tagCommit(ctx context.Context, workDir string, tag string) (string, error) {
	out, err := runGit(ctx, workDir, "rev-list", "-n", "1", tag)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func loadCommits(ctx context.Context, workDir string, baseTag string, target string) ([]rawCommit, error) {
	rangeSpec := target
	if baseTag != "" {
		rangeSpec = baseTag + ".." + target
	}

	out, err := runGit(ctx, workDir, "log", "--reverse", "--format=%x1e%H%x1f%aI%x1f%s%x1f%b", rangeSpec)
	if err != nil {
		return nil, err
	}

	records := strings.Split(string(out), "\x1e")
	commits := make([]rawCommit, 0, len(records))
	for _, record := range records {
		record = strings.Trim(record, "\n")
		if record == "" {
			continue
		}
		fields := strings.SplitN(record, "\x1f", 4)
		if len(fields) != 4 {
			return nil, fmt.Errorf("cannot parse git log output")
		}
		date, err := time.Parse(time.RFC3339, strings.TrimSpace(fields[1]))
		if err != nil {
			return nil, fmt.Errorf("cannot parse commit date: %w", err)
		}

		sha := strings.TrimSpace(fields[0])
		short := sha
		if len(short) > 12 {
			short = short[:12]
		}
		commits = append(commits, rawCommit{
			SHA:     sha,
			Short:   short,
			Date:    date,
			Subject: strings.TrimSpace(fields[2]),
			Body:    strings.Trim(fields[3], "\n"),
		})
	}

	return commits, nil
}

func runGit(ctx context.Context, workDir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if workDir != "" {
		cmd.Dir = workDir
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			return nil, err
		}
		return nil, fmt.Errorf("%s", message)
	}
	return out, nil
}

func splitLines(out []byte) []string {
	lines := strings.Split(string(out), "\n")
	values := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			values = append(values, line)
		}
	}
	return values
}
