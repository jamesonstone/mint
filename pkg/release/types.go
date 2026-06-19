package release

import "time"

const (
	// DefaultCommitish is the Git ref resolved when callers do not provide one.
	DefaultCommitish = "HEAD"
	// DefaultMintRef is the Mint action ref used by generated workflows.
	DefaultMintRef = "v1"
)

// Options configures read-only release resolution.
type Options struct {
	Commitish string
	WorkDir   string
}

// Result contains release metadata returned by Resolve.
type Result struct {
	VersionTag   string
	VersionBump  Bump
	BaseTag      string
	TargetSHA    string
	ShortSHA     string
	NeedsGitTag  bool
	CommitCount  int
	ReleaseNotes string
}

// Bump identifies the selected release bump.
type Bump string

const (
	// BumpAlreadyTagged means no new release tag is needed.
	BumpAlreadyTagged Bump = "already-tagged"
	// BumpPatch increments patch.
	BumpPatch Bump = "patch"
	// BumpMinor increments minor and resets patch.
	BumpMinor Bump = "minor"
	// BumpMajor increments major and resets minor and patch.
	BumpMajor Bump = "major"
)

type rawCommit struct {
	SHA     string
	Short   string
	Date    time.Time
	Subject string
	Body    string
}

type commitEvaluation struct {
	SHA      string
	Short    string
	Subject  string
	Body     string
	Type     string
	Breaking bool
	Reason   string
	Bump     Bump
	Rank     int
}

type semverTag struct {
	Name   string
	Major  int
	Minor  int
	Patch  int
	Commit string
}

// RegistryKind identifies the supported registry branch for generated workflows.
type RegistryKind string

const (
	// RegistryGHCR represents GitHub Container Registry.
	RegistryGHCR RegistryKind = "ghcr"
	// RegistryECR represents AWS Elastic Container Registry.
	RegistryECR RegistryKind = "ecr"
)

// ImageSpec describes one image repository and Docker build input.
type ImageSpec struct {
	Name         string
	URI          string
	Dockerfile   string
	Context      string
	RegistryHost string
	RegistryKind RegistryKind
}

// WorkflowOptions configures GitHub Actions publish workflow generation.
type WorkflowOptions struct {
	Images  []ImageSpec
	MintRef string
}
