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

// WriteReleaseTagOutputFile appends release Git tag fields to a GitHub output
// file.
func WriteReleaseTagOutputFile(path string, result TagResult) error {
	file, err := openOutputFile(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return WriteReleaseTagOutput(file, result)
}

// WriteSelectTagOutputFile appends selected release tag fields to a GitHub
// output file.
func WriteSelectTagOutputFile(path string, result SelectTagResult) error {
	file, err := openOutputFile(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return WriteSelectTagOutput(file, result)
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

// WriteReleaseTagOutput writes release Git tag fields in GitHub Actions output
// format.
func WriteReleaseTagOutput(writer io.Writer, result TagResult) error {
	fields := []struct {
		name  string
		value string
	}{
		{name: "tag_name", value: result.TagName},
		{name: "tag_target_sha", value: result.TargetSHA},
		{name: "tag_created", value: strconv.FormatBool(result.Created)},
		{name: "tag_reused", value: strconv.FormatBool(result.Reused)},
		{name: "tag_pushed", value: strconv.FormatBool(result.Pushed)},
	}

	for _, field := range fields {
		if _, err := fmt.Fprintf(writer, "%s=%s\n", field.name, field.value); err != nil {
			return err
		}
	}
	return nil
}

// WriteSelectTagOutput writes selected release tag fields in GitHub Actions
// output format.
func WriteSelectTagOutput(writer io.Writer, result SelectTagResult) error {
	fields := []struct {
		name  string
		value string
	}{
		{name: "version_tag", value: result.VersionTag},
		{name: "tag_source", value: string(result.TagSource)},
		{name: "target_sha", value: result.TargetSHA},
		{name: "short_sha", value: result.ShortSHA},
	}

	for _, field := range fields {
		if _, err := fmt.Fprintf(writer, "%s=%s\n", field.name, field.value); err != nil {
			return err
		}
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
