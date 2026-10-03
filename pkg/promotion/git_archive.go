package promotion

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// materialize isolates source-tree proof from the runner's moving checkout.
func (g GitProof) materialize(target string) (string, func(), error) {
	if !shaPattern.MatchString(target) {
		return "", func() {}, fmt.Errorf("tree proof requires exact source SHA")
	}
	archive, err := g.git(nil, "archive", "--format=tar", target)
	if err != nil {
		return "", func() {}, err
	}
	dir, err := os.MkdirTemp("", "mint-proof-*")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	reader := tar.NewReader(bytes.NewReader(archive))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			cleanup()
			return "", func() {}, err
		}
		path := filepath.Clean(header.Name)
		if filepath.IsAbs(path) || path == ".." || strings.HasPrefix(path, ".."+string(filepath.Separator)) {
			cleanup()
			return "", func() {}, fmt.Errorf("unsafe source-tree archive path")
		}
		destination := filepath.Join(dir, path)
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(destination, 0755)
		case tar.TypeReg:
			if header.Size > 64<<20 {
				cleanup()
				return "", func() {}, fmt.Errorf("tree proof file too large")
			}
			err = os.MkdirAll(filepath.Dir(destination), 0755)
			if err == nil {
				var data []byte
				data, err = io.ReadAll(reader)
				if err == nil {
					err = os.WriteFile(destination, data, os.FileMode(header.Mode)&0777)
				}
			}
		default: // Refuse ambiguous symlink/submodule proof instead of following paths.
			cleanup()
			return "", func() {}, fmt.Errorf("source tree includes unsupported archive entry %s; explicit retained-patch proof required", path)
		}
		if err != nil {
			cleanup()
			return "", func() {}, err
		}
	}
	return dir, cleanup, nil
}
