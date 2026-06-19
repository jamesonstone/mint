package changelog

import (
	"io"
	"time"
)

const DefaultOutputFile = "CHANGELOG.md"

type Options struct {
	PrevTag       string
	CurrentTag    string
	RepoOwner     string
	RepoName      string
	OutputFile    string
	WorkDir       string
	WarningWriter io.Writer
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
	Type        string
	Scope       string
	Breaking    bool
	Subject     string
	Body        string
	Hash        string
	Date        time.Time
	IssueNumber string
}
