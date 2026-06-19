---
kit_metadata_version: 1
artifact: spec
feature:
  id: 0003
  slug: github-action
  dir: 0003-github-action
summary: Expose the Mint CLI through a public GitHub composite action without adding release behavior.
relationships:
  - type: depends_on
    feature: 0002-cli-patterns
    reason: The GitHub Action wraps the Go CLI and Makefile-era command surface created by the CLI pattern feature.
references:
  - id: constitution
    name: Project constitution
    type: repo_doc
    target: docs/CONSTITUTION.md
    relation: constrains
    read_policy: must
    used_for: implementation evidence, source-of-truth order, and release-behavior non-goals
    status: active
  - id: cli-patterns
    name: CLI pattern specification
    type: feature_doc
    target: docs/specs/0002-cli-patterns/SPEC.md
    relation: depends_on
    read_policy: must
    used_for: existing CLI command and build surface
    status: active
  - id: tooling
    name: Tooling reference
    type: repo_doc
    target: docs/references/tooling.md
    relation: updates
    read_policy: must
    used_for: durable action usage notes
    status: active
---
# SPEC

## SUMMARY

Mint should be usable as a public GitHub Action while remaining a Go CLI first. The action should live at the repository root as `action.yml`, build `cmd/mint` from the checked-out action source, add the resulting binary to the runner `PATH`, and optionally run a small supported command.

## PROBLEM

The repository has a CLI scaffold but no GitHub Actions entrypoint. Consumers should be able to reference `jamesonstone/mint` from a workflow and run the same `mint` binary without copying build commands into every workflow.

## GOALS

1. Add a root GitHub Action metadata file that makes Mint publicly consumable with `uses: jamesonstone/mint@<ref>`.
2. Keep the core executable as `cmd/mint`; the action must build and run the CLI rather than reimplementing behavior in shell.
3. Add safe action inputs for the currently implemented command surface: `version`, `help`, and `none`.
4. Expose the built binary path and captured command output as action outputs.
5. Update the README with a GitHub Action quick start and workflow examples.
6. Update durable docs and feature progress to reflect the new public integration surface.

## NON-GOALS

1. Do not implement release computation, changelog generation, tagging, publishing, or GitHub release creation.
2. Do not add broad CI orchestration or repository release workflows.
3. Do not pass arbitrary user-provided command strings to a shell.
4. Do not vendor a prebuilt binary into the repository.

## REQUIREMENTS

1. [SPEC-01] The action metadata must live at repository root in `action.yml`.
2. [SPEC-02] The action must use the composite action format and build the Go CLI from `cmd/mint`.
3. [SPEC-03] The action must set up Go with a configurable `go-version` input that defaults to the repository Go version.
4. [SPEC-04] The action must add the built CLI directory to `GITHUB_PATH`.
5. [SPEC-05] The action must support `command: version`, `command: help`, and `command: none`.
6. [SPEC-06] Unsupported command values must fail with a clear GitHub Actions error.
7. [SPEC-07] The action must expose `mint-path` and `output` outputs.
8. [SPEC-08] README quick start documentation must show how to use Mint in a GitHub Actions workflow.

## ACCEPTANCE

1. [AC-01] `action.yml` parses as YAML.
2. [AC-02] A local build equivalent to the action build step succeeds.
3. [AC-03] The locally built action binary runs `version` and `--help`.
4. [AC-04] `go test ./...` passes.
5. [AC-05] `go vet ./...` passes.
6. [AC-06] `make build` succeeds.
7. [AC-07] `git diff --check` reports no whitespace errors for touched files.
8. [AC-08] `kit map 0003-github-action` resolves the feature.

## OPEN QUESTIONS

none
