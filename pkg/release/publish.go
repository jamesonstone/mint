package release

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// PublishOptions configures full release-state publishing.
type PublishOptions struct {
	Commitish   string
	WorkDir     string
	Owner       string
	Repo        string
	Title       string
	Token       string
	APIBaseURL  string
	Remote      string
	PushTag     bool
	VersionFile string
	HTTPClient  *http.Client
}

// PublishResult contains release resolution, Git tag, and GitHub Release state.
type PublishResult struct {
	Release       Result
	Tag           TagResult
	GitHubRelease GitHubReleaseResult
}

// PublishRelease resolves release metadata, creates or reuses the Git tag, and
// creates or reuses the GitHub Release.
func PublishRelease(ctx context.Context, opts PublishOptions) (PublishResult, error) {
	resolved, err := Resolve(ctx, Options{
		Commitish: opts.Commitish,
		WorkDir:   opts.WorkDir,
	})
	if err != nil {
		return PublishResult{}, err
	}

	versionFile := opts.VersionFile
	if versionFile == "" {
		versionFile = DefaultVersionFile
	}
	if opts.WorkDir != "" && !filepath.IsAbs(versionFile) {
		versionFile = filepath.Join(opts.WorkDir, versionFile)
	}
	if err := WriteVersionFile(versionFile, resolved.VersionTag); err != nil {
		return PublishResult{}, err
	}

	notesFile, err := writeTempReleaseNotes(resolved.ReleaseNotes)
	if err != nil {
		return PublishResult{}, err
	}
	defer os.Remove(notesFile)

	tag, err := CreateReleaseTag(ctx, TagOptions{
		Tag:       resolved.VersionTag,
		Target:    resolved.TargetSHA,
		NotesFile: notesFile,
		Remote:    opts.Remote,
		Push:      opts.PushTag,
		WorkDir:   opts.WorkDir,
	})
	if err != nil {
		return PublishResult{}, err
	}

	title := opts.Title
	if title == "" {
		title = resolved.VersionTag
	}
	github, err := PublishGitHubRelease(ctx, GitHubReleaseOptions{
		Owner:      opts.Owner,
		Repo:       opts.Repo,
		Tag:        resolved.VersionTag,
		Target:     resolved.TargetSHA,
		Title:      title,
		Notes:      resolved.ReleaseNotes,
		Token:      opts.Token,
		APIBaseURL: opts.APIBaseURL,
		HTTPClient: opts.HTTPClient,
	})
	if err != nil {
		return PublishResult{}, err
	}

	return PublishResult{
		Release:       resolved,
		Tag:           tag,
		GitHubRelease: github,
	}, nil
}

// WritePublishOutputFile appends full publish result fields to a GitHub Actions
// output file.
func WritePublishOutputFile(path string, result PublishResult) error {
	file, err := openOutputFile(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return WritePublishOutput(file, result)
}

// WritePublishOutput writes full publish result fields in GitHub Actions output
// format.
func WritePublishOutput(writer io.Writer, result PublishResult) error {
	if err := WriteGitHubOutput(writer, result.Release); err != nil {
		return err
	}
	if err := WriteReleaseTagOutput(writer, result.Tag); err != nil {
		return err
	}
	return WriteGitHubReleaseOutput(writer, result.GitHubRelease)
}

func writeTempReleaseNotes(notes string) (string, error) {
	file, err := os.CreateTemp("", "mint-release-notes-*.md")
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := file.WriteString(notes); err != nil {
		return "", err
	}
	if _, err := file.WriteString("\n"); err != nil {
		return "", err
	}
	return file.Name(), nil
}
