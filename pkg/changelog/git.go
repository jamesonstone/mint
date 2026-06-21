package changelog

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func tagExists(ctx context.Context, workDir string, tag string) (bool, error) {
	return commitishExists(ctx, workDir, tag)
}

func commitishExists(ctx context.Context, workDir string, ref string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	if workDir != "" {
		cmd.Dir = workDir
	}
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func tagAuthorDate(ctx context.Context, workDir string, tag string) (time.Time, error) {
	return commitishAuthorDate(ctx, workDir, tag)
}

func commitishAuthorDate(ctx context.Context, workDir string, ref string) (time.Time, error) {
	out, err := runGit(ctx, workDir, "show", "-s", "--format=%aI", ref+"^{commit}")
	if err != nil {
		return time.Time{}, err
	}
	date, err := time.Parse(time.RFC3339, strings.TrimSpace(string(out)))
	if err != nil {
		return time.Time{}, fmt.Errorf("cannot parse commit date for %s: %w", ref, err)
	}
	return date, nil
}

func loadCommits(ctx context.Context, workDir string, prevTag string, currentTag string) ([]rawCommit, error) {
	rangeSpec := currentTag
	if prevTag != "" {
		rangeSpec = prevTag + ".." + currentTag
	}

	out, err := runGit(ctx, workDir, "log", "--format=%x1e%H%x1f%aI%x1f%s%x1f%b", rangeSpec)
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
		commits = append(commits, rawCommit{
			Hash:    strings.TrimSpace(fields[0]),
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
