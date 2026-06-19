---
kit_metadata_version: 1
artifact: spec
feature:
  id: 0006
  slug: github-release-publishing
  dir: 0006-github-release-publishing
summary: Add GitHub Release publishing to Mint and configure Mint to publish Mint's own GitHub Releases through the Mint action.
parallelization_mode: rlm
skills:
  - name: rlm
    source: repo-local guide
    path: docs/agents/RLM.md
    trigger: "new release behavior; workflow configuration; GitHub Actions integration"
    required: true
relationships:
  - type: depends_on
    target: 0005-pull-forward-v1-features
    reason: GitHub Release publishing depends on existing release resolution and action output fields.
references:
  - id: constitution
    name: Mint constitution
    type: repo_doc
    target: docs/CONSTITUTION.md
    relation: constrains
    read_policy: must
    used_for: CLI-first architecture, action boundary, and non-goal boundaries
    status: active
  - id: action-yml
    name: Mint composite action
    type: action
    target: action.yml
    relation: updates
    read_policy: evidence
    used_for: allowlisted command, input, and output integration
    status: active
  - id: release-package
    name: Release package
    type: source
    target: pkg/release
    relation: updates
    read_policy: evidence
    used_for: GitHub Release publishing implementation and tests
    status: active
  - id: github-release-api
    name: GitHub REST Releases API
    type: url
    target: https://docs.github.com/en/rest/releases/releases
    relation: guides
    read_policy: evidence
    used_for: create release, get release by tag, token headers, and response status behavior
    status: active
  - id: github-token-docs
    name: GitHub Actions GITHUB_TOKEN authentication
    type: url
    target: https://docs.github.com/actions/reference/authentication-in-a-workflow
    relation: guides
    read_policy: evidence
    used_for: workflow token passing and minimum permissions
    status: active
---
# SPEC

## SUMMARY

Add GitHub Release publishing to Mint and configure this repository so Mint uses
the Mint action to publish Mint's own GitHub Releases. The feature is GitHub
Release only: no container image, package-manager artifact, or service deploy
behavior is in scope.

## PROBLEM

Mint can resolve the next strict SemVer tag and expose the result through its
GitHub Action, but it cannot publish the GitHub Release artifact. The Mint repo
also has no workflow that uses Mint itself to publish Mint releases.

## GOALS

- Add `mint release github` to create or reuse a GitHub Release for a strict
  `vX.Y.Z` tag.
- Keep GitHub Release publishing in the Go CLI, with the action as a wrapper.
- Configure `.github/workflows/release.yaml` so Mint uses the local Mint action
  to resolve and publish Mint's GitHub Release.
- Expose action inputs and outputs needed by release workflows.
- Document the CLI and action quick start.

## NON-GOALS

- Do not publish container images from the Mint self-release workflow.
- Do not add release asset upload.
- Do not add package-manager publishing.
- Do not add ECS or other environment deployment.
- Do not require the GitHub CLI.

## CONTRACT

- Required CLI inputs: `--owner`, `--repo`, `--tag`, `--target`, and a GitHub
  token from the environment.
- Optional CLI inputs: `--title`, `--notes-file`, `--token-env`, `--api-url`,
  and `--github-output`.
- The tag must match strict `vX.Y.Z` SemVer.
- The command first checks for an existing release by tag.
- If the release exists, return success and mark `release_created=false`.
- If the release does not exist, create it through the GitHub REST API with
  `tag_name`, `target_commitish`, `name`, `body`, `draft=false`, and
  `prerelease=false`.
- Action command `github-release` must call `mint release github`.
- Action outputs are `release_tag`, `release_url`, and `release_created`.

## ACCEPTANCE CRITERIA

- `mint release github` creates a GitHub Release through the GitHub REST API in
  local HTTP tests.
- Existing releases are treated as idempotent success.
- Missing owner, repo, tag, target, or token fails closed.
- Invalid non-`vX.Y.Z` tags fail closed.
- The composite action allowlist includes `github-release` and does not execute
  arbitrary shell.
- The self-release workflow uses `uses: ./` Mint action steps for both
  `release-resolve` and `github-release`.
- The self-release workflow has `contents: write` and no container publish
  steps.
- README, constitution, tooling, testing, and progress docs reflect the new
  implemented behavior.
