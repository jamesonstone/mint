package changelog

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestLinearFooterLinks(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		want       int
	}{
		{"canonical", "Refs: https://linear.app/acme/issue/ENG-123/add-invitations", 1},
		{"bare", "https://linear.app/acme/issue/ENG-123", 1},
		{"closing", "Closes https://linear.app/acme/issue/ENG-123", 1},
		{"duplicate", "Refs: https://linear.app/acme/issue/ENG-123\nRefs: https://linear.app/acme/issue/ENG-123", 1},
		{"two", "Issue: https://linear.app/acme/issue/ENG-123\nFixes: https://linear.app/acme/issue/OPS-9", 2},
		{"text", "example https://linear.app/acme/issue/ENG-123 in prose", 0},
		{"lookalike host", "Refs: https://linear.app.evil/acme/issue/ENG-123", 0},
		{"credentials", "Refs: https://secret@linear.app/acme/issue/ENG-123", 0},
		{"query", "Refs: https://linear.app/acme/issue/ENG-123?token=secret", 0},
		{"markdown injection", "Refs: https://linear.app/acme/issue/ENG-123/x)[bad](evil", 0},
		{"http", "Refs: http://linear.app/acme/issue/ENG-123", 0},
		{"identifier only", "Refs: ENG-123", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := linearIssueLinks(tc.body)
			if len(got) != tc.want {
				t.Fatalf("got %#v", got)
			}
		})
	}
}

func TestMixedProviderChangelogOffline(t *testing.T) {
	// There is deliberately no provider client or token prerequisite.
	t.Setenv("LINEAR_API_KEY", "")
	var warnings bytes.Buffer
	parsed := parseCommits([]rawCommit{
		{Subject: "feat(ENG-123): :sparkles: add invitations (#91)", Body: "Closes #42\n\nRefs: https://linear.app/acme/issue/ENG-123/add-invitations", Hash: "1234567890"},
		{Subject: "fix(GH-10): old issue (#10)", Hash: "abcdef1234"},
		{Subject: "fix(OPS-9): unavailable ticket", Body: "Refs: https://linear.app/acme/issue/OPS-9", Hash: "9876543210"},
		{Subject: "feat(ENG-99): scope without URL", Hash: "7654321000"},
	}, &warnings)
	if warnings.Len() != 0 || len(parsed) != 4 {
		t.Fatalf("%#v %s", parsed, warnings.String())
	}
	rendered := renderReleaseBlock("1.0.0", "v1.0.0", time.Unix(0, 0), "owner", "repo", parsed)
	for _, want := range []string{
		"[ENG-123](https://linear.app/acme/issue/ENG-123/add-invitations)",
		"[#42](https://github.com/owner/repo/issues/42)",
		"[#10](https://github.com/owner/repo/issues/10)",
		"[OPS-9](https://linear.app/acme/issue/OPS-9)",
		"[1234567](https://github.com/owner/repo/commit/1234567)",
		"**ENG-99:** scope without URL",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("missing %s in %s", want, rendered)
		}
	}
	for _, bad := range []string{"issues/123", "issues/9", "[ENG-99](", "issues/91"} {
		if strings.Contains(rendered, bad) {
			t.Fatalf("unexpected %s in %s", bad, rendered)
		}
	}
}
