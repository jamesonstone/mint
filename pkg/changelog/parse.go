package changelog

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

var conventionalCommitPattern = regexp.MustCompile(`^(feat|fix|perf|refactor|docs|test|chore|build|ci)(\(.+?\))?(!)?:\s(.+)$`)
var bodyIssuePattern = regexp.MustCompile(`(?i)\b(?:closes|fixes|resolves)\s+#(\d+)\b`)
var subjectIssuePattern = regexp.MustCompile(`\s+\(#(\d+)\)$`)
var semverPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)

func parseCommits(rawCommits []rawCommit, warnings io.Writer) []commit {
	commits := make([]commit, 0, len(rawCommits))
	for _, raw := range rawCommits {
		match := conventionalCommitPattern.FindStringSubmatch(raw.Subject)
		if match == nil {
			if warnings != nil {
				_, _ = fmt.Fprintf(warnings, "warning: skipping non-conventional commit %s: %s\n", shortHash(raw.Hash), raw.Subject)
			}
			continue
		}

		scope := strings.TrimPrefix(strings.TrimSuffix(match[2], ")"), "(")
		issue, subject := issueNumberAndSubject(match[4], raw.Body)
		parsed := commit{
			Type:        match[1],
			Scope:       scope,
			Breaking:    match[3] == "!" || strings.Contains(raw.Body, "BREAKING CHANGE:"),
			Subject:     subject,
			Body:        raw.Body,
			Hash:        shortHash(raw.Hash),
			Date:        raw.Date,
			IssueNumber: issue,
		}
		if shouldInclude(parsed) {
			commits = append(commits, parsed)
		}
	}
	return commits
}

func shouldInclude(parsed commit) bool {
	switch parsed.Type {
	case "docs", "test", "chore":
		return false
	default:
		return true
	}
}

func issueNumberAndSubject(subject string, body string) (string, string) {
	if match := bodyIssuePattern.FindStringSubmatch(body); match != nil {
		return match[1], subject
	}
	if match := subjectIssuePattern.FindStringSubmatch(subject); match != nil {
		return match[1], strings.TrimSpace(subjectIssuePattern.ReplaceAllString(subject, ""))
	}
	return "", subject
}

func shortHash(hash string) string {
	if len(hash) <= 7 {
		return hash
	}
	return hash[:7]
}

func versionFromTag(tag string) (string, error) {
	version := strings.TrimPrefix(tag, "v")
	if !semverPattern.MatchString(version) {
		return "", fmt.Errorf("version %s is not semver", version)
	}
	return version, nil
}
