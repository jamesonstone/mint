---
kit_metadata_version: 1
artifact: plan
feature:
  id: 0006
  slug: github-release-publishing
  dir: 0006-github-release-publishing
summary: Implementation plan for GitHub Release publishing and Mint self-release workflow configuration.
parallelization_mode: rlm
skills:
  - name: rlm
    source: repo-local guide
    path: docs/agents/RLM.md
    trigger: "new release behavior; workflow configuration; GitHub Actions integration"
    required: true
relationships:
  - type: depends_on
    target: 0006-github-release-publishing/SPEC.md
    reason: The plan implements the accepted GitHub Release publishing contract.
references:
  - id: spec
    name: GitHub Release publishing spec
    type: feature_doc
    target: docs/specs/0006-github-release-publishing/SPEC.md
    relation: constrains
    read_policy: must
    used_for: binding scope, command contract, and acceptance criteria
    status: active
  - id: release-package
    name: Release package
    type: source
    target: pkg/release
    relation: updates
    read_policy: evidence
    used_for: API client, output writer, and package tests
    status: active
  - id: cli-release
    name: Release CLI commands
    type: source
    target: pkg/cli/release.go
    relation: updates
    read_policy: evidence
    used_for: Cobra command, flags, token lookup, and stdout behavior
    status: active
  - id: action-yml
    name: Mint composite action
    type: action
    target: action.yml
    relation: updates
    read_policy: evidence
    used_for: command allowlist, inputs, outputs, and token wiring
    status: active
  - id: self-release-workflow
    name: Mint self-release workflow
    type: workflow
    target: .github/workflows/release.yaml
    relation: creates
    read_policy: evidence
    used_for: using Mint to publish Mint GitHub Releases
    status: active
---
# PLAN

## APPROACH

1. Add a small GitHub REST API client in `pkg/release` for `get release by tag`
   and `create release`.
2. Validate required GitHub Release inputs before any network call.
3. Add GitHub Actions output writing for release publish results.
4. Add `mint release github` as a thin Cobra adapter over `pkg/release`.
5. Extend `action.yml` with the `github-release` command, inputs, and outputs.
6. Add `.github/workflows/release.yaml` that runs Mint's local action twice:
   resolve the release, then publish the GitHub Release.
7. Update README and durable docs after implementation evidence exists.

## DESIGN NOTES

- Use the Go standard library HTTP client instead of adding a dependency.
- Treat existing releases by tag as success so reruns are idempotent.
- Keep tokens in environment variables or action inputs, never positional
  arguments.
- Use local HTTP test servers instead of live GitHub API tests.
- Keep container workflow generation unchanged; GitHub Release publishing is a
  separate CLI/action command.

## VALIDATION

- `go fmt ./...`
- `go test ./...`
- `go vet ./...`
- `make build`
- `git diff --check`
- YAML parsing tests for `action.yml` and `.github/workflows/release.yaml`
