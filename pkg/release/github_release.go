package release

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	// DefaultGitHubAPIBaseURL is the public GitHub REST API base URL.
	DefaultGitHubAPIBaseURL = "https://api.github.com"
	gitHubAPIVersion        = "2026-03-10"
)

// GitHubReleaseOptions configures GitHub Release publishing.
type GitHubReleaseOptions struct {
	Owner      string
	Repo       string
	Tag        string
	Target     string
	Title      string
	Notes      string
	Token      string
	APIBaseURL string
	HTTPClient *http.Client
}

// GitHubReleaseResult contains the published or existing GitHub Release.
type GitHubReleaseResult struct {
	TagName string
	URL     string
	Created bool
}

type gitHubReleaseRequest struct {
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish"`
	Name            string `json:"name"`
	Body            string `json:"body"`
	Draft           bool   `json:"draft"`
	Prerelease      bool   `json:"prerelease"`
}

type gitHubReleaseResponse struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// PublishGitHubRelease creates a GitHub Release for a strict SemVer tag or
// returns the existing release for that tag.
func PublishGitHubRelease(ctx context.Context, opts GitHubReleaseOptions) (GitHubReleaseResult, error) {
	normalized, err := normalizeGitHubReleaseOptions(opts)
	if err != nil {
		return GitHubReleaseResult{}, err
	}

	baseURL, err := url.Parse(normalized.APIBaseURL)
	if err != nil {
		return GitHubReleaseResult{}, validationError("api-url", "invalid GitHub API URL: %v", err)
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return GitHubReleaseResult{}, validationError("api-url", "must be an absolute URL")
	}

	client := normalized.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	existing, found, err := getGitHubReleaseByTag(ctx, client, baseURL, normalized)
	if err != nil {
		return GitHubReleaseResult{}, err
	}
	if found {
		return existing, nil
	}

	created, err := createGitHubRelease(ctx, client, baseURL, normalized)
	if err != nil {
		return GitHubReleaseResult{}, err
	}
	return created, nil
}

// WriteGitHubReleaseOutputFile appends GitHub Release result fields to a
// GitHub Actions output file.
func WriteGitHubReleaseOutputFile(path string, result GitHubReleaseResult) error {
	file, err := openOutputFile(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return WriteGitHubReleaseOutput(file, result)
}

// WriteGitHubReleaseOutput writes GitHub Release fields in GitHub Actions
// output format.
func WriteGitHubReleaseOutput(writer io.Writer, result GitHubReleaseResult) error {
	fields := []struct {
		name  string
		value string
	}{
		{name: "release_tag", value: result.TagName},
		{name: "release_url", value: result.URL},
		{name: "release_created", value: strconv.FormatBool(result.Created)},
	}

	for _, field := range fields {
		if _, err := fmt.Fprintf(writer, "%s=%s\n", field.name, field.value); err != nil {
			return err
		}
	}
	return nil
}

func normalizeGitHubReleaseOptions(opts GitHubReleaseOptions) (GitHubReleaseOptions, error) {
	normalized := opts
	normalized.Owner = strings.TrimSpace(normalized.Owner)
	normalized.Repo = strings.TrimSpace(normalized.Repo)
	normalized.Tag = strings.TrimSpace(normalized.Tag)
	normalized.Target = strings.TrimSpace(normalized.Target)
	normalized.Title = strings.TrimSpace(normalized.Title)
	normalized.Token = strings.TrimSpace(normalized.Token)
	normalized.APIBaseURL = strings.TrimSpace(normalized.APIBaseURL)

	if normalized.Owner == "" {
		return GitHubReleaseOptions{}, validationError("owner", "GitHub repository owner is required")
	}
	if normalized.Repo == "" {
		return GitHubReleaseOptions{}, validationError("repo", "GitHub repository name is required")
	}
	if _, ok := parseSemVerTag(normalized.Tag); !ok {
		return GitHubReleaseOptions{}, validationError("tag", "must be a strict vX.Y.Z SemVer tag")
	}
	if normalized.Target == "" {
		return GitHubReleaseOptions{}, validationError("target", "target commitish is required")
	}
	if normalized.Token == "" {
		return GitHubReleaseOptions{}, validationError("token", "GitHub token is required")
	}
	if normalized.Title == "" {
		normalized.Title = normalized.Tag
	}
	if normalized.APIBaseURL == "" {
		normalized.APIBaseURL = DefaultGitHubAPIBaseURL
	}
	return normalized, nil
}

func getGitHubReleaseByTag(ctx context.Context, client *http.Client, baseURL *url.URL, opts GitHubReleaseOptions) (GitHubReleaseResult, bool, error) {
	endpoint := gitHubAPIURL(baseURL, "repos", opts.Owner, opts.Repo, "releases", "tags", opts.Tag)
	status, response, err := doGitHubReleaseRequest(ctx, client, http.MethodGet, endpoint, opts.Token, nil)
	if err != nil {
		return GitHubReleaseResult{}, false, err
	}
	switch status {
	case http.StatusOK:
		result, err := parseGitHubReleaseResponse(response, false)
		if err != nil {
			return GitHubReleaseResult{}, false, err
		}
		if result.TagName != opts.Tag {
			return GitHubReleaseResult{}, false, validationError("tag", "existing release tag %q does not match requested tag %q", result.TagName, opts.Tag)
		}
		return result, true, nil
	case http.StatusNotFound:
		return GitHubReleaseResult{}, false, nil
	default:
		return GitHubReleaseResult{}, false, gitHubResponseError("get release by tag", status, response)
	}
}

func createGitHubRelease(ctx context.Context, client *http.Client, baseURL *url.URL, opts GitHubReleaseOptions) (GitHubReleaseResult, error) {
	endpoint := gitHubAPIURL(baseURL, "repos", opts.Owner, opts.Repo, "releases")
	payload := gitHubReleaseRequest{
		TagName:         opts.Tag,
		TargetCommitish: opts.Target,
		Name:            opts.Title,
		Body:            opts.Notes,
		Draft:           false,
		Prerelease:      false,
	}

	status, response, err := doGitHubReleaseRequest(ctx, client, http.MethodPost, endpoint, opts.Token, payload)
	if err != nil {
		return GitHubReleaseResult{}, err
	}
	if status != http.StatusCreated {
		return GitHubReleaseResult{}, gitHubResponseError("create release", status, response)
	}
	return parseGitHubReleaseResponse(response, true)
}

func doGitHubReleaseRequest(ctx context.Context, client *http.Client, method string, endpoint string, token string, body any) (int, []byte, error) {
	var reader io.Reader
	if body != nil {
		var encoded bytes.Buffer
		if err := json.NewEncoder(&encoded).Encode(body); err != nil {
			return 0, nil, err
		}
		reader = &encoded
	}

	request, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return 0, nil, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("X-GitHub-Api-Version", gitHubAPIVersion)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := client.Do(request)
	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()

	data, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return response.StatusCode, nil, err
	}
	return response.StatusCode, data, nil
}

func parseGitHubReleaseResponse(data []byte, created bool) (GitHubReleaseResult, error) {
	var decoded gitHubReleaseResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		return GitHubReleaseResult{}, err
	}
	if decoded.TagName == "" {
		return GitHubReleaseResult{}, validationError("github-release", "response did not include tag_name")
	}
	if decoded.HTMLURL == "" {
		return GitHubReleaseResult{}, validationError("github-release", "response did not include html_url")
	}
	return GitHubReleaseResult{
		TagName: decoded.TagName,
		URL:     decoded.HTMLURL,
		Created: created,
	}, nil
}

func gitHubAPIURL(baseURL *url.URL, parts ...string) string {
	clone := *baseURL
	escaped := make([]string, 0, len(parts))
	for _, part := range parts {
		escaped = append(escaped, url.PathEscape(part))
	}

	prefix := strings.TrimRight(clone.Path, "/")
	clone.Path = prefix + "/" + strings.Join(escaped, "/")
	return clone.String()
}

func gitHubResponseError(operation string, status int, body []byte) error {
	message := strings.TrimSpace(string(body))
	if message == "" {
		return fmt.Errorf("%s failed: GitHub API returned status %d", operation, status)
	}
	return fmt.Errorf("%s failed: GitHub API returned status %d: %s", operation, status, message)
}
