package release

import (
	"errors"
	"os"
	"strings"
)

// WriteVersionFile writes a strict SemVer value without the Git tag prefix.
func WriteVersionFile(path string, versionTag string) error {
	return WriteScopedVersionFile(path, versionTag, Scope{})
}

// WriteScopedVersionFile writes the bare X.Y.Z version of a scoped tag.
func WriteScopedVersionFile(path string, versionTag string, scope Scope) error {
	if path == "" {
		return nil
	}
	version, err := ScopedVersionFromTag(versionTag, scope)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(version+"\n"), 0o644)
}

// VersionFromTag converts a strict vX.Y.Z Git tag to X.Y.Z for .version.
func VersionFromTag(versionTag string) (string, error) {
	return ScopedVersionFromTag(versionTag, Scope{})
}

// ScopedVersionFromTag returns X.Y.Z from a tag such as backend/vX.Y.Z.
func ScopedVersionFromTag(versionTag string, scope Scope) (string, error) {
	tag, ok := scope.parseTag(strings.TrimSpace(versionTag))
	if !ok {
		if scope.TagPrefix == "" {
			return "", errors.New("version tag must be strict vX.Y.Z SemVer")
		}
		return "", errors.New("version tag must be strict " + scope.TagPrefix + "vX.Y.Z SemVer")
	}
	return strings.TrimPrefix(strings.TrimPrefix(tag.Name, scope.TagPrefix), "v"), nil
}
