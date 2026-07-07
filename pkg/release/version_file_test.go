package release

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteVersionFileWritesSemanticVersionWithoutTagPrefix(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".version")
	if err := WriteVersionFile(path, "v1.2.3"); err != nil {
		t.Fatalf("WriteVersionFile() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "1.2.3\n" {
		t.Fatalf(".version = %q, want 1.2.3 newline", string(data))
	}
}

func TestWriteVersionFileRejectsNonStrictSemVerTag(t *testing.T) {
	if err := WriteVersionFile(filepath.Join(t.TempDir(), ".version"), "1.2.3"); err == nil {
		t.Fatal("WriteVersionFile() error = nil, want invalid tag error")
	}
}
