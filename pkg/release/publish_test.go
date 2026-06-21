package release

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublishReleaseResolvesTagsAndCreatesGitHubRelease(t *testing.T) {
	repo := newTestRepo(t)
	target := repo.commit(t, "feat: release publishing", "", "2024-01-01T00:00:00Z")

	var posted gitHubReleaseRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/repos/jamesonstone/mint/releases/tags/v0.1.0":
			http.NotFound(w, r)
		case r.Method == http.MethodPost && r.URL.Path == "/repos/jamesonstone/mint/releases":
			if err := json.NewDecoder(r.Body).Decode(&posted); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"tag_name":"v0.1.0","html_url":"https://github.com/jamesonstone/mint/releases/tag/v0.1.0"}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	result, err := PublishRelease(context.Background(), PublishOptions{
		WorkDir:    repo.dir,
		Commitish:  "HEAD",
		Owner:      "jamesonstone",
		Repo:       "mint",
		Token:      "test-token",
		APIBaseURL: server.URL,
		PushTag:    false,
	})
	if err != nil {
		t.Fatalf("PublishRelease() error = %v", err)
	}

	if result.Release.VersionTag != "v0.1.0" {
		t.Fatalf("VersionTag = %q, want v0.1.0", result.Release.VersionTag)
	}
	if !result.Tag.Created || result.Tag.Reused || result.Tag.Pushed {
		t.Fatalf("tag created/reused/pushed = %t/%t/%t", result.Tag.Created, result.Tag.Reused, result.Tag.Pushed)
	}
	if result.Tag.TargetSHA != target {
		t.Fatalf("tag target = %q, want %q", result.Tag.TargetSHA, target)
	}
	if !result.GitHubRelease.Created {
		t.Fatalf("GitHubRelease.Created = false, want true")
	}
	if posted.TagName != "v0.1.0" || posted.TargetCommitish != target {
		t.Fatalf("posted tag/target = %q/%q, want v0.1.0/%q", posted.TagName, posted.TargetCommitish, target)
	}
	if !strings.Contains(posted.Body, "feat: release publishing") {
		t.Fatalf("posted body missing release notes:\n%s", posted.Body)
	}
	if got := repo.revParse(t, "v0.1.0^{commit}"); got != target {
		t.Fatalf("created tag target = %q, want %q", got, target)
	}
}
