package promotion

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/jamesonstone/mint/pkg/release"
)

var componentNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)

// strictVersionPattern is an unprefixed vMAJOR.MINOR.PATCH release tag.
var strictVersionPattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

// anyVersionPattern also accepts a Git-ref-safe component prefix such as
// backend/. Use Component.ValidVersion where the expected component is known.
var anyVersionPattern = regexp.MustCompile(`^(?:[A-Za-z0-9][A-Za-z0-9._/-]*)?v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)

// Component scopes one independently versioned part of a repository. A policy
// without a component keeps the repository-wide version line, journals, and
// release-control files. Each component policy lives in its own file and owns
// tags with its prefix, commits touching its paths, journals named
// mint-release-state-<name>-<environment>, and control files under
// .mint/<name>/.
type Component struct {
	Name      string   `yaml:"name" json:"name,omitempty"`
	TagPrefix string   `yaml:"tag_prefix,omitempty" json:"tag_prefix,omitempty"`
	Paths     []string `yaml:"paths" json:"paths,omitempty"`
	Changelog string   `yaml:"changelog" json:"changelog,omitempty"`
}

// IsZero reports a repository-wide (unscoped) policy.
func (c Component) IsZero() bool { return c.Name == "" }

// normalized fills defaults: the tag prefix defaults to <name>/.
func (c Component) normalized() Component {
	if !c.IsZero() && c.TagPrefix == "" {
		c.TagPrefix = c.Name + "/"
	}
	c.Paths = append([]string(nil), c.Paths...)
	return c
}

func (c Component) validate() error {
	if c.IsZero() {
		return nil
	}
	if !componentNamePattern.MatchString(c.Name) {
		return fmt.Errorf("component name must match %s", componentNamePattern)
	}
	if len(c.Paths) == 0 {
		return fmt.Errorf("component %s requires at least one path", c.Name)
	}
	if err := c.ReleaseScope().Validate(); err != nil {
		return fmt.Errorf("component %s: %w", c.Name, err)
	}
	if c.TagPrefix == "" || strings.HasPrefix(c.TagPrefix, "v") {
		return fmt.Errorf("component %s tag_prefix must be nonempty and must not start with v", c.Name)
	}
	if c.Changelog == "" || strings.HasPrefix(c.Changelog, "/") || strings.Contains(c.Changelog, "..") || !strings.HasSuffix(c.Changelog, ".md") {
		return fmt.Errorf("component %s changelog must be a repository-relative .md path", c.Name)
	}
	return nil
}

// ReleaseScope returns the release scope for tags and history.
func (c Component) ReleaseScope() release.Scope {
	return release.Scope{TagPrefix: c.TagPrefix, Paths: append([]string(nil), c.Paths...)}
}

// pathspecArgs restricts git history commands to the component's paths.
func (c Component) pathspecArgs() []string {
	if len(c.Paths) == 0 {
		return nil
	}
	return append([]string{"--"}, c.Paths...)
}

// ValidVersion reports whether version is a strict release tag of this
// component: <prefix>vMAJOR.MINOR.PATCH, or vMAJOR.MINOR.PATCH when unscoped.
func (c Component) ValidVersion(version string) bool {
	return strings.HasPrefix(version, c.TagPrefix) && strictVersionPattern.MatchString(version[len(c.TagPrefix):])
}

// VersionRule describes the accepted version shape in error messages.
func (c Component) VersionRule() string { return "strict " + c.TagPrefix + "vMAJOR.MINOR.PATCH" }

func (c Component) controlFile(name string) string {
	if c.IsZero() {
		return ".mint/" + name
	}
	return ".mint/" + c.Name + "/" + name
}

// ProposalPath is the release-control declaration file.
func (c Component) ProposalPath() string { return c.controlFile("proposal.json") }

// SummaryPath is the release-control summary file.
func (c Component) SummaryPath() string { return c.controlFile("summary.md") }

// StatusPath is the optional release-control status file.
func (c Component) StatusPath() string { return c.controlFile("status.json") }

// ChangelogPath is the changelog Mint updates in release proposals.
func (c Component) ChangelogPath() string {
	if c.IsZero() {
		return "CHANGELOG.md"
	}
	return c.Changelog
}

// DefaultControlPaths are the trusted release-control files for this scope.
func (c Component) DefaultControlPaths() []string {
	return []string{c.ChangelogPath(), c.ProposalPath(), c.SummaryPath()}
}

// validateControlPaths permits exactly this scope's release-control files.
func (c Component) validateControlPaths(paths []string) error {
	allowed := map[string]bool{c.ChangelogPath(): true, c.ProposalPath(): true, c.SummaryPath(): true, c.StatusPath(): true}
	seen := map[string]bool{}
	for _, path := range paths {
		if !allowed[path] {
			return fmt.Errorf("unsafe release-control path %s", path)
		}
		if seen[path] {
			return fmt.Errorf("duplicate release-control path %s", path)
		}
		seen[path] = true
	}
	for _, required := range c.DefaultControlPaths() {
		if !seen[required] {
			return fmt.Errorf("release-control paths must include %s", required)
		}
	}
	return nil
}

// MarkerPrefix starts every proposal marker this scope owns. Component markers
// use a distinct namespace so repository-wide controllers never match them.
func (c Component) MarkerPrefix() string {
	if c.IsZero() {
		return "<!-- mint:proposal:"
	}
	return "<!-- mint:component-proposal:" + c.Name + ":"
}

// ProposalMarker identifies one proposal pull request body.
func (c Component) ProposalMarker(id, kind string) string {
	return c.MarkerPrefix() + id + ":" + kind + " -->"
}

// JournalBranch names the durable release journal. Unscoped production keeps
// mint-release-state; component journals always include the environment.
func (c Component) JournalBranch(environment string) (string, error) {
	if !ValidEnvironmentName(environment) {
		return "", fmt.Errorf("unsafe environment journal identity")
	}
	if c.IsZero() {
		if environment == "production" {
			return journalBranch, nil
		}
		return journalBranch + "-" + environment, nil
	}
	return journalBranch + "-" + c.Name + "-" + environment, nil
}

// StateKey names machine-local state for an environment.
func (c Component) StateKey(environment string) string {
	if c.IsZero() {
		return environment
	}
	return c.Name + "--" + environment
}

// workflowNameSuffix distinguishes generated workflow names per component.
func workflowNameSuffix(c Component) string {
	if c.IsZero() {
		return ""
	}
	return " (" + c.Name + ")"
}

// ControlConcurrencyGroup serializes one scope's journal writers.
func (c Component) ControlConcurrencyGroup() string {
	if c.IsZero() {
		return "mint-release-journal"
	}
	return "mint-release-journal-" + c.Name
}
