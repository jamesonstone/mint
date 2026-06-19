package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "mint",
	Short: "🪙 Mint computes versions, changelogs, and releases",
	Long: banner() + `
Mint is a release tooling CLI for computing the next version, writing the
changelog, and minting the release.

The current command surface provides version reporting, CHANGELOG.md generation
from conventional commits, release tag resolution, GitHub Release publishing,
and GHCR/ECR publish workflow generation. Mint resolves release metadata
directly, publishes GitHub Releases through the GitHub API, and leaves generated
container workflows to own Git tag creation and image publishing.

` + releaseFlow(),
	Version: Version,
}

func init() {
	rootCmd.SetVersionTemplate("mint version {{.Version}}\n")
	configureRootHelp()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		var exitErr *cliExitError
		if errors.As(err, &exitErr) {
			if !exitErr.silent {
				fmt.Fprintln(os.Stderr, exitErr.Error())
			}
			os.Exit(exitErr.code)
		}
		var silentErr *silentCLIError
		if !errors.As(err, &silentErr) {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func banner() string {
	return `███╗   ███╗██╗███╗   ██╗████████╗
████╗ ████║██║████╗  ██║╚══██╔══╝
██╔████╔██║██║██╔██╗ ██║   ██║
██║╚██╔╝██║██║██║╚██╗██║   ██║
██║ ╚═╝ ██║██║██║ ╚████║   ██║
╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝   ╚═╝
`
}

func releaseFlow() string {
	return `🪙 Release Intent
  ┌──────────────┐    ┌───────────┐    ┌─────────┐
  │ Next Version │ -> │ Changelog │ -> │ Release │
  └──────────────┘    └───────────┘    └─────────┘`
}

type silentCLIError struct {
	err error
}

func (e *silentCLIError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *silentCLIError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}

type cliExitError struct {
	err    error
	code   int
	silent bool
}

func (e *cliExitError) Error() string {
	if e == nil || e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *cliExitError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}
