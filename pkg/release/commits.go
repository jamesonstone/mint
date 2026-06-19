package release

import (
	"regexp"
	"strings"
)

const (
	bumpRankNone = iota
	bumpRankPatch
	bumpRankMinor
	bumpRankMajor
)

var conventionalSubjectPattern = regexp.MustCompile(`^([A-Za-z]+)(?:\([^)\n]+\))?(!)?:\s(.+)$`)
var breakingBodyPattern = regexp.MustCompile(`BREAKING[ -]CHANGE:`)

func evaluateCommits(commits []rawCommit) ([]commitEvaluation, Bump, int) {
	evaluations := make([]commitEvaluation, 0, len(commits))
	selectedBump := BumpAlreadyTagged
	selectedRank := bumpRankNone

	for _, commit := range commits {
		evaluation := evaluateCommit(commit)
		evaluations = append(evaluations, evaluation)
		if evaluation.Rank > selectedRank {
			selectedRank = evaluation.Rank
			selectedBump = evaluation.Bump
		}
	}

	return evaluations, selectedBump, selectedRank
}

func evaluateCommit(commit rawCommit) commitEvaluation {
	evaluation := commitEvaluation{
		SHA:     commit.SHA,
		Short:   commit.Short,
		Subject: commit.Subject,
		Body:    commit.Body,
		Bump:    BumpPatch,
		Rank:    bumpRankPatch,
	}
	bodyBreaking := breakingBodyPattern.MatchString(commit.Body)

	matches := conventionalSubjectPattern.FindStringSubmatch(commit.Subject)
	if matches == nil {
		if bodyBreaking {
			evaluation.Breaking = true
			evaluation.Reason = "breaking change"
			evaluation.Bump = BumpMajor
			evaluation.Rank = bumpRankMajor
			return evaluation
		}
		evaluation.Reason = "non-conventional commit"
		return evaluation
	}

	evaluation.Type = strings.ToLower(matches[1])
	evaluation.Breaking = matches[2] == "!" || bodyBreaking
	if evaluation.Breaking {
		evaluation.Reason = "breaking change"
		evaluation.Bump = BumpMajor
		evaluation.Rank = bumpRankMajor
		return evaluation
	}

	switch evaluation.Type {
	case "feat":
		evaluation.Reason = "feature"
		evaluation.Bump = BumpMinor
		evaluation.Rank = bumpRankMinor
	case "fix":
		evaluation.Reason = "fix"
	default:
		evaluation.Reason = "conventional " + evaluation.Type
	}

	return evaluation
}
