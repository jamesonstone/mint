package release

import "testing"

func TestHotfixConventionalCommitClassification(t *testing.T) {
	tests := []struct {
		subject string
		body    string
		bump    Bump
		rank    int
		reason  string
	}{
		{subject: "hotfix: repair login", bump: BumpPatch, rank: bumpRankPatch, reason: "fix"},
		{subject: "hotfix(GH-123): :firetruck: repair login", bump: BumpPatch, rank: bumpRankPatch, reason: "fix"},
		{subject: "hotfix(GH-123): 🚒 repair login", bump: BumpPatch, rank: bumpRankPatch, reason: "fix"},
		{subject: "hotfix(api)!: repair contract", bump: BumpMajor, rank: bumpRankMajor, reason: "breaking change"},
		{subject: "hotfix: repair contract", body: "BREAKING CHANGE: API changed", bump: BumpMajor, rank: bumpRankMajor, reason: "breaking change"},
	}
	for _, tt := range tests {
		t.Run(tt.subject+tt.body, func(t *testing.T) {
			got := evaluateCommit(rawCommit{Subject: tt.subject, Body: tt.body})
			if got.Type != "hotfix" || got.Bump != tt.bump || got.Rank != tt.rank || got.Reason != tt.reason {
				t.Fatalf("classification = %#v", got)
			}
		})
	}
}
