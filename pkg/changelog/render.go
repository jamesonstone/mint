package changelog

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type groupDefinition struct {
	title string
	match func(commit) bool
}

var groupDefinitions = []groupDefinition{
	{
		title: "breaking changes",
		match: func(item commit) bool {
			return item.Breaking
		},
	},
	{
		title: "features",
		match: func(item commit) bool {
			return !item.Breaking && item.Type == "feat"
		},
	},
	{
		title: "fixes",
		match: func(item commit) bool {
			return !item.Breaking && item.Type == "fix"
		},
	},
	{
		title: "perf",
		match: func(item commit) bool {
			return !item.Breaking && item.Type == "perf"
		},
	},
	{
		title: "other",
		match: func(item commit) bool {
			return !item.Breaking && (item.Type == "refactor" || item.Type == "build" || item.Type == "ci")
		},
	},
}

func renderReleaseBlock(version string, tag string, releaseDate time.Time, repoOwner string, repoName string, commits []commit) string {
	var builder strings.Builder
	tagURL := fmt.Sprintf("https://github.com/%s/%s/releases/tag/%s", repoOwner, repoName, tag)

	fmt.Fprintf(&builder, "## [%s](%s) - %s\n", version, tagURL, releaseDate.Format("2006-01-02"))

	for _, group := range groupDefinitions {
		items := commitsForGroup(commits, group)
		if len(items) == 0 {
			continue
		}

		fmt.Fprintf(&builder, "\n### %s\n", group.title)
		fmt.Fprintln(&builder)
		for _, item := range items {
			fmt.Fprintf(&builder, "- %s%s\n", scopePrefix(item.Scope), item.Subject)
			if item.IssueNumber != "" {
				issueURL := fmt.Sprintf("https://github.com/%s/%s/issues/%s", repoOwner, repoName, item.IssueNumber)
				fmt.Fprintf(&builder, "  ([#%s](%s))\n", item.IssueNumber, issueURL)
			}
			commitURL := fmt.Sprintf("https://github.com/%s/%s/commit/%s", repoOwner, repoName, item.Hash)
			fmt.Fprintf(&builder, "  ([%s](%s))\n", item.Hash, commitURL)
		}
	}

	return strings.TrimRight(builder.String(), "\n") + "\n"
}

func commitsForGroup(commits []commit, definition groupDefinition) []commit {
	items := make([]commit, 0)
	for _, item := range commits {
		if definition.match(item) {
			items = append(items, item)
		}
	}
	sort.SliceStable(items, func(i int, j int) bool {
		if items[i].Scope != items[j].Scope {
			return items[i].Scope < items[j].Scope
		}
		if !items[i].Date.Equal(items[j].Date) {
			return items[i].Date.Before(items[j].Date)
		}
		return items[i].Hash < items[j].Hash
	})
	return items
}

func scopePrefix(scope string) string {
	if scope == "" {
		return ""
	}
	return fmt.Sprintf("**%s:** ", scope)
}

func breakingCount(commits []commit) int {
	count := 0
	for _, item := range commits {
		if item.Breaking {
			count++
		}
	}
	return count
}
