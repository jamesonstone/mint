package cli

import (
	"bytes"
	"runtime/debug"
	"testing"
)

func TestCurrentVersionPrefersLinkerInjectedVersion(t *testing.T) {
	originalVersion := Version
	originalBuildInfoReader := buildInfoReader
	Version = "v1.2.3"
	defer func() {
		Version = originalVersion
		buildInfoReader = originalBuildInfoReader
	}()

	buildInfoReader = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}}, true
	}

	if got, want := currentVersion(), "v1.2.3"; got != want {
		t.Fatalf("currentVersion() = %q, want %q", got, want)
	}
}

func TestCurrentVersionFallsBackToBuildInfo(t *testing.T) {
	originalVersion := Version
	originalBuildInfoReader := buildInfoReader
	Version = "dev"
	defer func() {
		Version = originalVersion
		buildInfoReader = originalBuildInfoReader
	}()

	buildInfoReader = func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}}, true
	}

	if got, want := currentVersion(), "v1.2.3"; got != want {
		t.Fatalf("currentVersion() = %q, want %q", got, want)
	}
}

func TestRunVersionPrintsCurrentVersion(t *testing.T) {
	originalVersion := Version
	originalBuildInfoReader := buildInfoReader
	Version = "v1.2.3"
	defer func() {
		Version = originalVersion
		buildInfoReader = originalBuildInfoReader
	}()

	buildInfoReader = func() (*debug.BuildInfo, bool) {
		return nil, false
	}

	output := &bytes.Buffer{}
	versionCmd.SetOut(output)
	defer versionCmd.SetOut(nil)

	if err := runVersion(versionCmd, nil); err != nil {
		t.Fatalf("runVersion() error = %v", err)
	}

	if got, want := output.String(), "v1.2.3\n"; got != want {
		t.Fatalf("runVersion() output = %q, want %q", got, want)
	}
}
