# Tooling Reference

## Purpose

- Record durable repo-wide tooling notes, command references, and local development expectations
- Keep short-lived implementation notes in feature docs instead of here

## Current State

- Mint uses a Go module at `github.com/jamesonstone/mint`.
- The CLI binary entrypoint is `cmd/mint`.
- Reusable command code belongs under `pkg/cli`.
- Changelog generation code belongs under `pkg/changelog`.
- Release resolution, GitHub Release publishing, and publish workflow
  generation code belongs under `pkg/release`.
- `Makefile` is the durable local build surface.
- `action.yml` is the public GitHub Action metadata and wraps the CLI instead of reimplementing Mint behavior in workflow shell.

## Commands

- `make build` builds `bin/mint` with linker-injected `pkg/cli.Version`.
- `make build-windows` builds `bin/mint.exe` for Windows amd64.
- `make install` installs `cmd/mint` with the same linker-injected version path.
- `make install-git-hooks` configures this clone to use `.githooks/`.
- `make fmt`, `make vet`, `make test`, `make lint`, `make tidy`, and `make all` mirror the Kit repository's local development targets.
- `mint changelog --prev-tag <tag> --current-tag <tag> --owner <owner> --repo <repo> --output CHANGELOG.md` prepends a conventional-commit release block.
- The root command accepts the same changelog flags directly for script-friendly usage.
- `mint release resolve --commitish <ref>` resolves strict SemVer release metadata from reachable Git history.
- `mint release resolve --commitish <ref> --github-output "$GITHUB_OUTPUT"` writes release metadata fields for GitHub Actions.
- `mint release github --owner <owner> --repo <repo> --tag <tag> --target <sha> --notes-file <path>` creates or reuses a GitHub Release.
- `mint release workflow --image name=<name>,uri=<repo-uri>,dockerfile=<path>,context=<path> --output <path>` renders a GHCR or ECR publish workflow.

## GitHub Action

- Consumers can use Mint from a workflow with `uses: jamesonstone/mint@<ref>`.
- The action is a composite action that sets up Go, builds `./cmd/mint`, adds the built binary directory to `GITHUB_PATH`, and optionally runs a supported command.
- Supported `command` input values are `version`, `help`, `changelog`, `release-resolve`, `github-release`, and `none`.
- The `go-version` input defaults to `1.25.5`, matching `go.mod`.
- `command: changelog` uses `prev-tag`, `current-tag`, `owner`, `repo`, and `output` inputs.
- `command: release-resolve` uses the `commitish` input and exposes typed release outputs.
- `command: github-release` uses `owner`, `repo`, `release-tag`, `target-sha`, `release-title`, `release-notes-file`, `github-token`, and `github-api-url` inputs.
- The action outputs `mint-path`, captured `output`, and release fields for `release-resolve`.
- GitHub Release publishing outputs include `release_tag`, `release_url`, and `release_created`.
- Additional release behavior should be added to the CLI through feature specs first, then exposed through the action.
