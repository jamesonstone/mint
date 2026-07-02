---
kit_metadata_version: 1
artifact: spec
feature:
  id: 0008
  slug: release-tag-selection
  dir: 0008-release-tag-selection
summary: Add a read-only release tag selector so deploy workflows can recover an already-published SemVer image tag without copied shell.
parallelization_mode: rlm
skills:
  - name: rlm
    source: repo-local guide
    path: docs/agents/RLM.md
    trigger: "release command surface; action behavior; workflow boundary"
    required: true
relationships:
  - type: builds_on
    target: 0007-release-state-ownership
    reason: Release tag selection extends Mint's release-state boundary without adding Docker or deploy ownership.
references:
  - id: constitution
    name: Mint constitution
    type: repo_doc
    target: docs/CONSTITUTION.md
    relation: constrains
    read_policy: must
    used_for: CLI-first behavior, action boundary, and deployment non-goals
    status: active
  - id: r2-main-workflow
    name: r2 main workflow
    type: external_source
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/r2/.github/workflows/main.yaml
    relation: informs
    read_policy: evidence
    used_for: release-publish output and image tag creation behavior
    status: active
  - id: r2-deploy-workflow
    name: r2 deploy workflow
    type: external_source
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/r2/.github/workflows/deploy.yaml
    relation: informs
    read_policy: evidence
    used_for: copied SemVer image tag lookup to replace with Mint helper
    status: active
---
# SPEC

## SUMMARY

Mint must expose a read-only release tag selector for deployment handoff
workflows. The selector validates a manually requested strict SemVer tag or
selects the highest strict SemVer tag already pointing at a target commit. It
must never compute a new release version, create or move tags, publish images,
or deploy services.

## GOALS

- Add `mint release select-tag` for existing-tag selection.
- Add `command: release-select-tag` to the GitHub Action allowlist.
- Expose `version_tag`, `tag_source`, `target_sha`, and `short_sha` outputs.
- Keep manual deployment overrides simple through an optional requested tag.
- Document how application deploy workflows can consume the helper.

## NON-GOALS

- Do not make Mint query ECR, GHCR, ECS, or any deployment target.
- Do not make Mint build, push, retag, or verify container images.
- Do not make Mint compute the next release version in this command.
- Do not change generated publish workflow ownership.
- Do not update downstream application repositories in this Mint PR.

## ACCEPTANCE CRITERIA

- Requested tags must be strict `vX.Y.Z` SemVer tags.
- Without a requested tag, Mint selects the highest strict SemVer tag pointing
  at the target commit.
- If no strict SemVer tag points at the target commit, Mint fails closed with a
  clear deployment-handoff error.
- The CLI prints the selected tag to stdout.
- The GitHub Action exposes typed outputs that deploy workflows can consume.
- Tests cover requested tag validation, commit-tag selection, no-tag failure,
  CLI output, and action wiring.
