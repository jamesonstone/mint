---
kit_metadata_version: 1
artifact: spec
feature:
  id: 0007
  slug: release-state-ownership
  dir: 0007-release-state-ownership
summary: Make Mint own release-state operations through CLI/action commands while application repositories keep Docker publishing and deployment workflows.
parallelization_mode: rlm
skills:
  - name: rlm
    source: repo-local guide
    path: docs/agents/RLM.md
    trigger: "release ownership boundary; CLI/action behavior; workflow boundary"
    required: true
relationships:
  - type: builds_on
    target: 0005-pull-forward-v1-features
    reason: Release-state ownership builds on existing SemVer resolution and generated workflow behavior.
  - type: builds_on
    target: 0006-github-release-publishing
    reason: Release-state ownership composes GitHub Release creation with first-class Git tag creation.
references:
  - id: constitution
    name: Mint constitution
    type: repo_doc
    target: docs/CONSTITUTION.md
    relation: constrains
    read_policy: must
    used_for: CLI-first behavior, action boundary, and Docker/deploy non-goals
    status: active
  - id: release-package
    name: Release package
    type: source
    target: pkg/release
    relation: updates
    read_policy: evidence
    used_for: release tag and publish state implementation
    status: active
  - id: cli-release
    name: Release CLI
    type: source
    target: pkg/cli
    relation: updates
    read_policy: evidence
    used_for: release tag and publish command adapters
    status: active
  - id: action-yml
    name: Mint composite action
    type: action
    target: action.yml
    relation: updates
    read_policy: evidence
    used_for: allowlisted release-state commands and outputs
    status: active
---
# SPEC

## SUMMARY

Mint must own release-state operations directly: version resolution,
release-note/changelog support, immutable Git tag creation, and GitHub Release
creation. Application repositories keep Docker image building, registry
authentication, registry publishing, deployment, and infrastructure workflows.

## GOALS

- Add `mint release tag` for annotated SemVer Git tag creation/reuse.
- Add `mint release publish` to run resolve, release-note temp file creation,
  tag creation/reuse, tag push, and GitHub Release creation/reuse.
- Extend the GitHub Action with fixed `release-tag` and `release-publish`
  commands and typed tag outputs.
- Update generated Docker publish workflows to call Mint for tag creation
  instead of embedding tag shell.
- Update README and durable docs with the release-state/application-boundary
  contract.

## NON-GOALS

- Do not make Mint build Docker images.
- Do not make Mint authenticate to ECR/GHCR directly.
- Do not make Mint own ECS/service deployment.
- Do not move or rewrite existing Git tags.
- Do not add arbitrary shell execution to `action.yml`.

## ACCEPTANCE CRITERIA

- Git tag creation exists as package and CLI behavior, not only generated shell.
- Same-commit tags are reused successfully.
- Conflicting tags fail before mutation and include the documented recovery
  path.
- Missing target/tag validation fails closed.
- Tags are never moved.
- `release publish` composes resolver, tag, and GitHub Release behavior.
- Action metadata supports `release-tag` and `release-publish` with tag
  created/reused/pushed outputs.
- Docker and ECS behavior remain application-owned.
