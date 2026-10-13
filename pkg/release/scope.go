package release

import (
	"regexp"
	"strings"
)

// tagPrefixPattern allows Git-ref-safe component prefixes such as "backend/"
// or "web-". It rejects empty path segments and parent references.
var tagPrefixPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)

// Scope narrows release state to one component of a repository. The zero
// value is the whole repository with unprefixed vX.Y.Z tags.
type Scope struct {
	// TagPrefix precedes the SemVer tag, for example "backend/" gives
	// backend/v1.2.0. Tags without the prefix are ignored.
	TagPrefix string
	// Paths limits commit history to commits that touch these Git pathspecs.
	Paths []string
}

// Validate reports whether the scope can form safe tag names and pathspecs.
func (s Scope) Validate() error {
	if s.TagPrefix != "" {
		if !tagPrefixPattern.MatchString(s.TagPrefix) ||
			strings.Contains(s.TagPrefix, "..") ||
			strings.Contains(s.TagPrefix, "//") ||
			strings.HasSuffix(s.TagPrefix, ".") ||
			strings.HasSuffix(s.TagPrefix, ".lock") {
			return validationError("tag-prefix", "must be a Git-ref-safe prefix such as backend/ (letters, digits, . _ / -)")
		}
	}
	for _, path := range s.Paths {
		trimmed := strings.TrimSpace(path)
		if trimmed == "" || trimmed != path || strings.HasPrefix(path, "-") {
			return validationError("path", "paths must be non-empty Git pathspecs without surrounding spaces or a leading dash")
		}
	}
	return nil
}

// FormatTag returns the scoped tag for a version, for example backend/v1.2.0.
func (s Scope) FormatTag(major int, minor int, patch int) string {
	return s.TagPrefix + formatSemVer(major, minor, patch)
}

// parseTag parses a scoped SemVer tag; tags outside the scope do not match.
func (s Scope) parseTag(name string) (semverTag, bool) {
	if !strings.HasPrefix(name, s.TagPrefix) {
		return semverTag{}, false
	}
	tag, ok := parseSemVerTag(strings.TrimPrefix(name, s.TagPrefix))
	if !ok {
		return semverTag{}, false
	}
	tag.Name = name
	return tag, true
}

// tagRule describes the accepted tag shape in validation messages.
func (s Scope) tagRule() string {
	return "must be a strict " + s.TagPrefix + "vX.Y.Z SemVer tag"
}

// pathspecArgs returns the trailing git arguments that restrict history to the
// scope's paths, or nothing for the whole repository.
func (s Scope) pathspecArgs() []string {
	if len(s.Paths) == 0 {
		return nil
	}
	return append([]string{"--"}, s.Paths...)
}
