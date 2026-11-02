package promotion

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var hotfixIDPattern = regexp.MustCompile(`^(GH|[A-Z][A-Z0-9]*)-[1-9][0-9]*$`)
var linearIDPattern = regexp.MustCompile(`^[A-Z][A-Z0-9]*-[1-9][0-9]*$`)
var linearSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
var linearWorkspacePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// HotfixRequestIdentity selects one legacy GitHub issue or explicit Linear URL.
// The ticket is attribution only; authenticated Actions still authorizes requests.
func HotfixRequestIdentity(issue int, tickets ...string) (string, string, error) {
	ticket := ""
	if len(tickets) > 1 {
		return "", "", fmt.Errorf("one hotfix ticket URL required")
	}
	if len(tickets) == 1 {
		ticket = tickets[0]
	}
	if ticket == "" && issue > 0 {
		return fmt.Sprintf("GH-%d", issue), "", nil
	}
	if issue != 0 || ticket == "" {
		return "", "", fmt.Errorf("select exactly one GitHub issue or Linear ticket URL")
	}
	u, err := url.Parse(ticket)
	if err != nil || u.Scheme != "https" || u.Host != "linear.app" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.RawPath != "" {
		return "", "", fmt.Errorf("canonical https://linear.app workspace issue URL required")
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if (len(parts) != 3 && len(parts) != 4) || !linearWorkspacePattern.MatchString(parts[0]) || parts[1] != "issue" || !linearIDPattern.MatchString(parts[2]) || strings.HasPrefix(parts[2], "GH-") || (len(parts) == 4 && !linearSlugPattern.MatchString(parts[3])) {
		return "", "", fmt.Errorf("canonical Linear ticket identifier required")
	}
	return parts[2], ticket, nil
}

func (c Component) hotfixRequestBase(id string) string {
	return strings.TrimSuffix(c.HotfixBasePrefix(), "GH-") + id
}

func (c Component) hotfixRequestMarker(id, baseline string) string {
	return c.HotfixSourceMarkerPrefix() + id + ":" + baseline + " -->"
}

func (c Component) hotfixSourceRequest(body string) (string, string, bool) {
	start := strings.Index(body, c.HotfixSourceMarkerPrefix())
	if start < 0 {
		return "", "", false
	}
	rest := body[start+len(c.HotfixSourceMarkerPrefix()):]
	end := strings.Index(rest, " -->")
	if end < 0 {
		return "", "", false
	}
	id, baseline, ok := strings.Cut(rest[:end], ":")
	return id, baseline, ok && hotfixIDPattern.MatchString(id) && baseline != ""
}

// MatchesHotfixSource binds the exact request ID, component and baseline.
// This is identity validation only; callers must still verify source reviews.
func (c Component) MatchesHotfixSource(body, baseRef, baseline string) bool {
	id, markedBaseline, ok := c.hotfixSourceRequest(body)
	return ok && markedBaseline == baseline && baseRef == c.hotfixRequestBase(id)
}

// hotfixSourceBranch separates new Linear production fixes from development.
// Legacy GitHub issues keep their existing governed branch identity.
func (c Component) hotfixSourceBranch(id, ticket string) string {
	if ticket == "" {
		return id
	}
	prefix := "mint-hotfix-source/"
	if !c.IsZero() {
		prefix += c.Name + "/"
	}
	return prefix + id
}
