package changelog

import (
	"io"
	"time"
)

const DefaultOutputFile = "CHANGELOG.md"

type Options struct {
	PrevTag       string
	CurrentTag    string
	CurrentRef    string
	RepoOwner     string
	RepoName      string
	OutputFile    string
	WorkDir       string
	WarningWriter io.Writer
	// TagPrefix scopes tags to one component, for example "backend/" for
	// backend/v1.2.0. Headers keep the bare X.Y.Z version.
	TagPrefix string
	// Paths limits entries to commits touching these Git pathspecs.
	Paths []string
}

type Result struct {
	Path          string
	Tag           string
	Version       string
	CommitCount   int
	BreakingCount int
}

type rawCommit struct {
	Hash    string
	Date    time.Time
	Subject string
	Body    string
}

type commit struct {
	Type         string
	Scope        string
	Breaking     bool
	Subject      string
	Body         string
	Hash         string
	Date         time.Time
	IssueNumber  string
	LinearIssues []issueLink
}
