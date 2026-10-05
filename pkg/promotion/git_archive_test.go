package promotion

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitArchiveAcceptsPAXMetadataAndRejectsLinks(t *testing.T) {
	for _, link := range []bool{false, true} {
		t.Run(map[bool]string{false: "regular", true: "symlink"}[link], func(t *testing.T) {
			dir := t.TempDir()
			git := func(args ...string) string {
				cmd := exec.Command("git", args...)
				cmd.Dir = dir
				out, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("git: %v %s", err, out)
				}
				return strings.TrimSpace(string(out))
			}
			git("init", "-b", "main")
			git("config", "user.name", "Human")
			git("config", "user.email", "human@example.com")
			if err := os.WriteFile(filepath.Join(dir, "source.txt"), []byte("content"), 0600); err != nil {
				t.Fatal(err)
			}
			if link {
				if err := os.Symlink("source.txt", filepath.Join(dir, "link.txt")); err != nil {
					t.Fatal(err)
				}
			}
			git("add", "source.txt")
			if link {
				git("add", "link.txt")
			}
			git("commit", "-m", "feat: fixture")
			proof := GitProof{Context: t.Context(), WorkDir: dir}
			tree, cleanup, err := proof.materialize(git("rev-parse", "HEAD"))
			defer cleanup()
			if link {
				if err == nil {
					t.Fatal("symlink accepted")
				}
				return
			}
			if err != nil {
				t.Fatal("ordinary Git archive rejected", err)
			}
			data, err := os.ReadFile(filepath.Join(tree, "source.txt"))
			if err != nil || string(data) != "content" {
				t.Fatal(string(data), err)
			}
		})
	}
}
