package release

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteGitHubOutputIncludesAllReleaseFields(t *testing.T) {
	var output bytes.Buffer
	result := Result{
		VersionTag:   "v1.2.3",
		VersionBump:  BumpMinor,
		BaseTag:      "v1.1.0",
		TargetSHA:    strings.Repeat("a", 40),
		ShortSHA:     strings.Repeat("a", 12),
		NeedsGitTag:  true,
		CommitCount:  2,
		ReleaseNotes: "Release notes\nMINT_RELEASE_NOTES\nmore notes",
	}

	if err := WriteGitHubOutput(&output, result); err != nil {
		t.Fatalf("WriteGitHubOutput() error = %v", err)
	}

	value := output.String()
	for _, want := range []string{
		"version_tag=v1.2.3\n",
		"version_bump=minor\n",
		"base_tag=v1.1.0\n",
		"target_sha=" + result.TargetSHA + "\n",
		"short_sha=" + result.ShortSHA + "\n",
		"needs_git_tag=true\n",
		"commit_count=2\n",
		"release_notes<<MINT_RELEASE_NOTES_1\n",
		"Release notes\nMINT_RELEASE_NOTES\nmore notes\n",
		"MINT_RELEASE_NOTES_1\n",
	} {
		if !strings.Contains(value, want) {
			t.Fatalf("output missing %q:\n%s", want, value)
		}
	}
}
