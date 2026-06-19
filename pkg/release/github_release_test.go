package release

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublishGitHubReleaseCreatesRelease(t *testing.T) {
	var postCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("Authorization header = %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Fatalf("Accept header = %q", r.Header.Get("Accept"))
		}

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/jamesonstone/mint/releases/tags/v1.2.3":
			http.NotFound(w, r)
		case r.Method == http.MethodPost && r.URL.Path == "/repos/jamesonstone/mint/releases":
			postCount++
			var payload gitHubReleaseRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if payload.TagName != "v1.2.3" {
				t.Fatalf("tag_name = %q", payload.TagName)
			}
			if payload.TargetCommitish != "abc123" {
				t.Fatalf("target_commitish = %q", payload.TargetCommitish)
			}
			if payload.Name != "Mint v1.2.3" {
				t.Fatalf("name = %q", payload.Name)
			}
			if payload.Body != "Release notes" {
				t.Fatalf("body = %q", payload.Body)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"tag_name":"v1.2.3","html_url":"https://github.com/jamesonstone/mint/releases/tag/v1.2.3"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	result, err := PublishGitHubRelease(context.Background(), GitHubReleaseOptions{
		Owner:      "jamesonstone",
		Repo:       "mint",
		Tag:        "v1.2.3",
		Target:     "abc123",
		Title:      "Mint v1.2.3",
		Notes:      "Release notes",
		Token:      "test-token",
		APIBaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("PublishGitHubRelease() error = %v", err)
	}

	if postCount != 1 {
		t.Fatalf("POST count = %d, want 1", postCount)
	}
	if !result.Created {
		t.Fatalf("Created = false, want true")
	}
	if result.TagName != "v1.2.3" {
		t.Fatalf("TagName = %q", result.TagName)
	}
	if result.URL != "https://github.com/jamesonstone/mint/releases/tag/v1.2.3" {
		t.Fatalf("URL = %q", result.URL)
	}
}

func TestPublishGitHubReleaseReturnsExistingRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/repos/jamesonstone/mint/releases/tags/v1.2.3" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"tag_name":"v1.2.3","html_url":"https://github.com/jamesonstone/mint/releases/tag/v1.2.3"}`))
	}))
	defer server.Close()

	result, err := PublishGitHubRelease(context.Background(), GitHubReleaseOptions{
		Owner:      "jamesonstone",
		Repo:       "mint",
		Tag:        "v1.2.3",
		Target:     "abc123",
		Token:      "test-token",
		APIBaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("PublishGitHubRelease() error = %v", err)
	}

	if result.Created {
		t.Fatalf("Created = true, want false")
	}
	if result.TagName != "v1.2.3" {
		t.Fatalf("TagName = %q", result.TagName)
	}
}

func TestPublishGitHubReleaseValidatesRequiredInputs(t *testing.T) {
	tests := []struct {
		name string
		opts GitHubReleaseOptions
		want string
	}{
		{name: "owner", opts: GitHubReleaseOptions{}, want: "owner"},
		{name: "tag", opts: GitHubReleaseOptions{Owner: "o", Repo: "r", Tag: "1.2.3", Target: "abc", Token: "token"}, want: "strict vX.Y.Z"},
		{name: "target", opts: GitHubReleaseOptions{Owner: "o", Repo: "r", Tag: "v1.2.3", Token: "token"}, want: "target"},
		{name: "token", opts: GitHubReleaseOptions{Owner: "o", Repo: "r", Tag: "v1.2.3", Target: "abc"}, want: "token"},
		{name: "api", opts: GitHubReleaseOptions{Owner: "o", Repo: "r", Tag: "v1.2.3", Target: "abc", Token: "token", APIBaseURL: "://bad"}, want: "api-url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PublishGitHubRelease(context.Background(), tt.opts)
			if err == nil {
				t.Fatalf("PublishGitHubRelease() error = nil, want %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("PublishGitHubRelease() error = %q, want contains %q", err.Error(), tt.want)
			}
		})
	}
}

func TestWriteGitHubReleaseOutput(t *testing.T) {
	var output bytes.Buffer
	err := WriteGitHubReleaseOutput(&output, GitHubReleaseResult{
		TagName: "v1.2.3",
		URL:     "https://github.com/jamesonstone/mint/releases/tag/v1.2.3",
		Created: true,
	})
	if err != nil {
		t.Fatalf("WriteGitHubReleaseOutput() error = %v", err)
	}

	for _, want := range []string{
		"release_tag=v1.2.3",
		"release_url=https://github.com/jamesonstone/mint/releases/tag/v1.2.3",
		"release_created=true",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("output missing %q:\n%s", want, output.String())
		}
	}
}
