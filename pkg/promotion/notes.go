package promotion

import (
	"fmt"
	"strings"
)

func escapeMarkdown(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	replacer := strings.NewReplacer("\\", "\\\\", "[", "\\[", "]", "\\]", "*", "\\*", "_", "\\_", "<", "&lt;", ">", "&gt;", "`", "\\`")
	return replacer.Replace(s)
}

// RenderNotes produces the shared canonical CHANGELOG and GitHub Release entry.
// The date is supplied by immutable source evidence, never a rerun clock.
func RenderNotes(repository, version, summary string, changes []Change, dates ...string) string {
	if strings.TrimSpace(summary) == "" {
		summary = fmt.Sprintf("Includes %d reviewed change(s) since the last successful production release.", len(changes))
	}
	var b strings.Builder
	heading := version
	if len(dates) > 0 && dates[0] != "" {
		heading += " — " + dates[0]
	}
	fmt.Fprintf(&b, "## %s\n\n%s\n\n", heading, strings.TrimSpace(summary))
	seenPR := map[int]bool{}
	for _, ch := range changes {
		if ch.PR > 0 && seenPR[ch.PR] {
			continue
		}
		prefix := fmt.Sprintf("- [%s](https://github.com/%s/tree/%s) — ", ch.Version, repository, ch.Version)
		if ch.OriginalVersion != "" {
			prefix += fmt.Sprintf("(from [%s](https://github.com/%s/tree/%s)) — ", ch.OriginalVersion, repository, ch.OriginalVersion)
		}
		if ch.PR > 0 {
			fmt.Fprintf(&b, "%s[#%d: %s](https://github.com/%s/pull/%d)\n", prefix, ch.PR, escapeMarkdown(ch.Title), repository, ch.PR)
			seenPR[ch.PR] = true
		} else {
			fmt.Fprintf(&b, "%s[%s: %s](https://github.com/%s/commit/%s)\n", prefix, ch.SHA[:12], escapeMarkdown(ch.Title), repository, ch.SHA)
		}
	}
	return b.String()
}
