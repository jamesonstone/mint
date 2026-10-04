package promotion

import "strings"

// WorkflowPathMatches accepts GitHub's optional @ref suffix on a server-attested
// run path. Repository, event and source identity must still be checked separately.
// The configured path remains exact: a basename or arbitrary suffix is not enough.
func WorkflowPathMatches(observed, configured string) bool {
	if configured == "" || strings.Contains(configured, "@") {
		return false
	}
	path, ref, qualified := strings.Cut(observed, "@")
	return path == configured && (!qualified || ref != "")
}
