---
name: mint-release-tooling
description: >-
  Use when a coding agent needs to operate or modify Mint, a Go CLI and GitHub
  Action for release tooling. Trigger for tasks involving SemVer release
  resolution, CHANGELOG.md generation, immutable Git tag creation, GitHub
  Release publishing, Mint GitHub Action workflows, GHCR/ECR workflow
  generation, or repository release-state boundaries.
---

# Mint Release Tooling Skill

Use this skill from the Mint repository root or from a repository that consumes
Mint as a CLI/action.

## Load

1. Read `agent-instructions/instructions.md` for the complete functionality
   manifest.
2. Read `README.md` for current examples and public documentation.
3. In the Mint repository, read `docs/agents/README.md` before editing.
4. When a decision depends on project invariants, read `docs/CONSTITUTION.md`.

## Decide

- Use `mint release resolve` when the caller needs version metadata only.
- Use `mint changelog` when the caller needs a deterministic `CHANGELOG.md`
  block from conventional commits.
- Use `mint release tag` when the caller needs an immutable annotated Git tag.
- Use `mint release github` when the Git tag exists and the caller needs a
  GitHub Release.
- Use `mint release publish` when the caller wants resolve, tag, push, and
  GitHub Release in one release-state operation.
- Use `mint release workflow` only when generating GHCR/ECR image publish YAML.
- Use the GitHub Action when the caller needs Mint behavior inside a GitHub
  Actions workflow.

## Guardrails

- Keep Mint's boundary explicit: Mint owns release state; application
  repositories own Docker builds, registry authentication, image publishing,
  infrastructure, and deployments.
- Never move an existing Git tag. Reuse same-commit tags and fail on
  conflicting tags.
- Do not reimplement Mint domain behavior in action shell or copied workflow
  shell.
- Do not add arbitrary command execution to `action.yml`.
- Do not claim behavior exists unless repository evidence proves it.
- Do not stage, commit, push, create issues, or mutate PRs without explicit
  user approval and repo-local delivery-rule loading.

## Common Workflows

### Inspect Mint

```bash
mint --help
mint release --help
```

From source:

```bash
go run ./cmd/mint --help
go run ./cmd/mint release --help
```

### Resolve Release State

```bash
mint release resolve --commitish HEAD
```

In GitHub Actions, check out full history first:

```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 0

- id: release
  uses: jamesonstone/mint@v1
  with:
    command: release-resolve
    commitish: ${{ github.sha }}
```

### Generate Changelog

```bash
mint changelog \
  --prev-tag v1.0.0 \
  --current-tag v1.1.0 \
  --current-ref HEAD \
  --owner jamesonstone \
  --repo mint \
  --output CHANGELOG.md
```

Use `--current-ref` when `--current-tag` is the release version to render but
the Git tag does not exist yet.

### Create Or Reuse A Git Tag

```bash
mint release tag \
  --tag v1.1.0 \
  --target "$(git rev-parse HEAD)" \
  --notes-file CHANGELOG.md \
  --push=true
```

Expected behavior:

- Same tag on same target commit succeeds.
- Same tag on another commit fails.
- Tags are never moved.

### Create Or Reuse A GitHub Release

```bash
GITHUB_TOKEN=... mint release github \
  --owner jamesonstone \
  --repo mint \
  --tag v1.1.0 \
  --target "$(git rev-parse HEAD)" \
  --notes-file CHANGELOG.md
```

Token lookup defaults to `MINT_GITHUB_TOKEN`, `GITHUB_TOKEN`, then `GH_TOKEN`.
Use `--token-env` to point at a different environment variable.

### Publish Release State

```bash
GITHUB_TOKEN=... mint release publish \
  --owner jamesonstone \
  --repo mint \
  --commitish HEAD
```

This command resolves the version, writes temporary release notes, creates or
reuses the Git tag, pushes it by default, and creates or reuses the GitHub
Release. It does not build images, push containers, or deploy services.

### Use Mint As A GitHub Action

```yaml
permissions:
  contents: write

steps:
  - uses: actions/checkout@v4
    with:
      fetch-depth: 0

  - id: release
    uses: jamesonstone/mint@v1
    with:
      command: release-publish
      commitish: ${{ github.sha }}
      owner: ${{ github.repository_owner }}
      repo: ${{ github.event.repository.name }}
      github-token: ${{ secrets.GITHUB_TOKEN }}
```

Use `command: none` to install the built `mint` binary into the workflow `PATH`
without running a Mint command immediately.

### Generate GHCR Or ECR Publish Workflow YAML

```bash
mint release workflow \
  --image name=api,uri=ghcr.io/jamesonstone/mint-api,dockerfile=Dockerfile.api,context=. \
  --output .github/workflows/release-publish.yml
```

Generated workflows publish images but must still delegate release resolution
and Git tag creation to the Mint action.

## Modify Mint

1. Classify the work through `docs/agents/README.md`.
2. Keep CLI adapters in `pkg/cli`.
3. Keep release and changelog behavior in `pkg/release` and `pkg/changelog`.
4. Keep `action.yml` as a CLI wrapper with an allowlisted command switch.
5. Update README and agent instructions when command behavior or workflow
   semantics change.

## Validate

For docs-only edits:

```bash
git diff --check
```

For CLI, action, release, changelog, or build-surface edits:

```bash
go test ./...
go vet ./...
make build
go run ./cmd/mint release resolve --commitish HEAD
go run ./cmd/mint release tag --help
go run ./cmd/mint release github --help
go run ./cmd/mint release publish --help
git diff --check
```
