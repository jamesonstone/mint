package promotion

import "strings"

// Revert attribution uses the actual inverse patch, never a title or annotation.
func attributeReviewedRevert(g GitProof, s State, change *Change) {
	reverse, err := g.git(nil, "diff", change.SHA, change.SHA+"^", "--", ".", ":(exclude).mint")
	if err != nil {
		return
	}
	out, err := g.git(reverse, "patch-id", "--stable")
	fields := strings.Fields(string(out))
	if err != nil || len(fields) == 0 {
		return
	}
	for _, shipped := range s.Baseline.Shipped {
		if shipped.PatchID == fields[0] {
			change.Revert, change.Reverts = true, shipped.PatchID
		}
	}
}
