package changelog

import "regexp"

// Only explicit whole-line references are release-note inputs. Do not crawl
// arbitrary description text, query Linear, or infer workspace from an ID.
var linearFooterPattern = regexp.MustCompile(`(?m)^(?:(?:Refs|Issue|Closes|Fixes|Resolves):? +)?(https://linear\.app/[A-Za-z0-9][A-Za-z0-9_-]*/issue/([A-Z][A-Z0-9]*-[1-9][0-9]*)(?:/[A-Za-z0-9_-]+)?)[ \t]*$`)

type issueLink struct {
	Identifier string
	URL        string
}

func linearIssueLinks(body string) []issueLink {
	var result []issueLink
	seen := map[string]bool{}
	for _, match := range linearFooterPattern.FindAllStringSubmatch(body, -1) {
		if seen[match[1]] {
			continue
		}
		seen[match[1]] = true
		result = append(result, issueLink{Identifier: match[2], URL: match[1]})
	}
	return result
}
