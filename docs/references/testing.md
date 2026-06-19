# Testing Reference

## Purpose

- Record durable repo-wide testing guidance that is broader than one feature
- Keep feature-specific testing details in the current feature's `PLAN.md` or `TASKS.md`

## Current State

- Go package tests are the current project-specific test strategy.
- Use `go test ./...` for the script-friendly full test suite.
- Use `make test` when following the repository Makefile pattern.
- Use `go vet ./...` or `make vet` before considering CLI/build changes complete.
- Use `make build` as the local binary smoke test for `cmd/mint`.
- Use `go test ./pkg/changelog` for changelog parser, renderer, Git fixture, and file-handling coverage.
- For GitHub Action metadata changes, parse `action.yml` as YAML and run a local action-equivalent build of `./cmd/mint`.
- For GitHub Action wrapper changes, smoke-test the locally built binary with `version` and `--help`.
- For changelog CLI changes, smoke-test `./bin/mint changelog --help` and run the full Go test suite.
- Use `go test ./pkg/release` for release resolver, image validation, GitHub output, action metadata, and workflow rendering coverage.
- For release workflow changes, assert generated YAML parses and keeps tag creation before image publishing.
- For release resolver changes, use temporary Git repositories with deterministic commits, dates, and tags.
