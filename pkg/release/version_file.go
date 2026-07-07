package release

import (
	"errors"
	"os"
	"strings"
)

// WriteVersionFile writes a strict SemVer value without the Git tag prefix.
func WriteVersionFile(path string, versionTag string) error {
	if path == "" {
		return nil
	}
	version, err := VersionFromTag(versionTag)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(version+"\n"), 0o644)
}

// VersionFromTag converts a strict vX.Y.Z Git tag to X.Y.Z for .version.
func VersionFromTag(versionTag string) (string, error) {
	tag, ok := parseSemVerTag(strings.TrimSpace(versionTag))
	if !ok {
		return "", errors.New("version tag must be strict vX.Y.Z SemVer")
	}
	return strings.TrimPrefix(tag.Name, "v"), nil
}
