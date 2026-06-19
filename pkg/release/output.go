package release

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// WriteGitHubOutputFile appends release result fields to a GitHub output file.
func WriteGitHubOutputFile(path string, result Result) error {
	file, err := openOutputFile(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return WriteGitHubOutput(file, result)
}

func openOutputFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
}

// WriteGitHubOutput writes release result fields in GitHub Actions output format.
func WriteGitHubOutput(writer io.Writer, result Result) error {
	fields := []struct {
		name  string
		value string
	}{
		{name: "version_tag", value: result.VersionTag},
		{name: "version_bump", value: string(result.VersionBump)},
		{name: "base_tag", value: result.BaseTag},
		{name: "target_sha", value: result.TargetSHA},
		{name: "short_sha", value: result.ShortSHA},
		{name: "needs_git_tag", value: strconv.FormatBool(result.NeedsGitTag)},
		{name: "commit_count", value: strconv.Itoa(result.CommitCount)},
	}

	for _, field := range fields {
		if _, err := fmt.Fprintf(writer, "%s=%s\n", field.name, field.value); err != nil {
			return err
		}
	}

	delimiter := outputDelimiter("MINT_RELEASE_NOTES", result.ReleaseNotes)
	if _, err := fmt.Fprintf(writer, "release_notes<<%s\n%s\n%s\n", delimiter, result.ReleaseNotes, delimiter); err != nil {
		return err
	}
	return nil
}

func outputDelimiter(base string, value string) string {
	delimiter := base
	for i := 1; strings.Contains(value, delimiter); i++ {
		delimiter = fmt.Sprintf("%s_%d", base, i)
	}
	return delimiter
}
