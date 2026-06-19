package release

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type testRepo struct {
	dir     string
	counter int
}

func newTestRepo(t *testing.T) *testRepo {
	t.Helper()

	repo := &testRepo{dir: t.TempDir()}
	repo.git(t, nil, "init", "--initial-branch", "main")
	repo.git(t, nil, "config", "user.name", "Mint Test")
	repo.git(t, nil, "config", "user.email", "mint@example.com")
	return repo
}

func (repo *testRepo) commit(t *testing.T, subject string, body string, date string) string {
	t.Helper()

	repo.counter++
	path := filepath.Join(repo.dir, fmt.Sprintf("file-%02d.txt", repo.counter))
	if err := os.WriteFile(path, []byte(subject+"\n"), 0o644); err != nil {
		t.Fatalf("write fixture file: %v", err)
	}
	repo.git(t, nil, "add", ".")

	args := []string{"commit", "-m", subject}
	if body != "" {
		args = append(args, "-m", body)
	}
	env := []string{
		"GIT_AUTHOR_DATE=" + date,
		"GIT_COMMITTER_DATE=" + date,
	}
	repo.git(t, env, args...)
	return repo.revParse(t, "HEAD")
}

func (repo *testRepo) tag(t *testing.T, tag string) {
	t.Helper()
	repo.git(t, nil, "tag", tag)
}

func (repo *testRepo) checkoutNewBranch(t *testing.T, branch string) {
	t.Helper()
	repo.git(t, nil, "checkout", "-b", branch)
}

func (repo *testRepo) checkout(t *testing.T, ref string) {
	t.Helper()
	repo.git(t, nil, "checkout", ref)
}

func (repo *testRepo) revParse(t *testing.T, ref string) string {
	t.Helper()
	return strings.TrimSpace(repo.gitOutput(t, nil, "rev-parse", ref))
}

func (repo *testRepo) git(t *testing.T, env []string, args ...string) {
	t.Helper()
	_ = repo.gitOutput(t, env, args...)
}

func (repo *testRepo) gitOutput(t *testing.T, env []string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = repo.dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
	return string(out)
}
