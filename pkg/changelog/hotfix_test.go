package changelog

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestHotfixChangelogPreservesScopeAndEmojiInFixes(t *testing.T) {
	var warnings bytes.Buffer
	commits := parseCommits([]rawCommit{{
		Hash:    "123456789abcdef",
		Subject: "hotfix(GH-123): :firetruck: repair production login (#123)",
		Date:    time.Date(2026, 10, 4, 0, 0, 0, 0, time.UTC),
	}}, &warnings)
	if len(commits) != 1 || warnings.Len() != 0 {
		t.Fatalf("commits = %#v, warnings = %q", commits, warnings.String())
	}
	got := commits[0]
	if got.Type != "hotfix" || got.Scope != "GH-123" || got.IssueNumber != "123" {
		t.Fatalf("parsed hotfix = %#v", got)
	}
	content := renderReleaseBlock("1.2.5", "v1.2.5", got.Date, "owner", "repo", commits)
	for _, want := range []string{"### fixes", "- **GH-123:** :firetruck: repair production login", "[#123](https://github.com/owner/repo/issues/123)"} {
		if !strings.Contains(content, want) {
			t.Fatalf("changelog missing %q:\n%s", want, content)
		}
	}
	if strings.Contains(content, "### other") || strings.Contains(content, "### breaking changes") {
		t.Fatalf("hotfix rendered in wrong group:\n%s", content)
	}
}

func TestBreakingHotfixAppearsOnlyInBreakingChanges(t *testing.T) {
	commits := parseCommits([]rawCommit{{Subject: "hotfix(api)!: repair incompatible contract"}}, nil)
	content := renderReleaseBlock("2.0.0", "v2.0.0", time.Time{}, "owner", "repo", commits)
	if !strings.Contains(content, "### breaking changes") || strings.Contains(content, "### fixes") {
		t.Fatalf("breaking hotfix rendered in wrong group:\n%s", content)
	}
}
