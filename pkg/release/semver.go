package release

import (
	"fmt"
	"regexp"
	"strconv"
)

var strictSemVerTagPattern = regexp.MustCompile(`^v([0-9]+)\.([0-9]+)\.([0-9]+)$`)

func parseSemVerTag(name string) (semverTag, bool) {
	matches := strictSemVerTagPattern.FindStringSubmatch(name)
	if matches == nil {
		return semverTag{}, false
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return semverTag{}, false
	}
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return semverTag{}, false
	}
	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return semverTag{}, false
	}

	return semverTag{
		Name:  name,
		Major: major,
		Minor: minor,
		Patch: patch,
	}, true
}

func compareSemVer(a semverTag, b semverTag) int {
	if a.Major != b.Major {
		return compareInt(a.Major, b.Major)
	}
	if a.Minor != b.Minor {
		return compareInt(a.Minor, b.Minor)
	}
	return compareInt(a.Patch, b.Patch)
}

func compareInt(a int, b int) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}

func highestSemVer(tags []semverTag) (semverTag, bool) {
	if len(tags) == 0 {
		return semverTag{}, false
	}

	highest := tags[0]
	for _, tag := range tags[1:] {
		if compareSemVer(tag, highest) > 0 {
			highest = tag
		}
	}
	return highest, true
}

func formatSemVer(major int, minor int, patch int) string {
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}
