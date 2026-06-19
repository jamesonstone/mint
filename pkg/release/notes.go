package release

import (
	"fmt"
	"strings"
)

func renderReleaseNotes(result Result, evaluations []commitEvaluation) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "Release: %s\n", result.VersionTag)
	if result.BaseTag == "" {
		builder.WriteString("Base: none\n")
	} else {
		fmt.Fprintf(&builder, "Base: %s\n", result.BaseTag)
	}
	fmt.Fprintf(&builder, "Target: %s\n", result.TargetSHA)
	fmt.Fprintf(&builder, "Bump: %s\n", result.VersionBump)

	if len(evaluations) == 0 {
		builder.WriteString("\nNo commits after base tag; reusing existing release version.\n")
		return builder.String()
	}

	builder.WriteString("\nCommits:\n")
	for _, evaluation := range evaluations {
		fmt.Fprintf(&builder, "- %s %s (%s, %s)\n",
			evaluation.Short,
			evaluation.Subject,
			evaluation.Reason,
			evaluation.Bump,
		)
	}
	return strings.TrimRight(builder.String(), "\n")
}
