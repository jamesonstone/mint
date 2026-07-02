---
kit_metadata_version: 1
artifact: plan
feature:
  id: 0008
  slug: release-tag-selection
  dir: 0008-release-tag-selection
summary: Implementation plan for read-only release tag selection.
parallelization_mode: rlm
skills:
  - name: rlm
    source: repo-local guide
    path: docs/agents/RLM.md
    trigger: "release command surface; action behavior; workflow boundary"
    required: true
relationships:
  - type: depends_on
    target: docs/specs/0008-release-tag-selection/SPEC.md
    reason: This plan implements the release tag selection contract.
references:
  - id: spec
    name: Release tag selection spec
    type: feature_doc
    target: docs/specs/0008-release-tag-selection/SPEC.md
    relation: constrains
    read_policy: must
    used_for: scope, non-goals, and acceptance criteria
    status: active
  - id: release-package
    name: Release package
    type: source
    target: pkg/release
    relation: updates
    read_policy: evidence
    used_for: selector behavior and output writing
    status: active
  - id: action-yml
    name: Mint composite action
    type: action
    target: action.yml
    relation: updates
    read_policy: evidence
    used_for: command allowlist and output wiring
    status: active
---
# PLAN

## APPROACH

1. Add a small `pkg/release` selector that resolves the target commit, validates
   an optional requested tag, or chooses the highest strict SemVer tag pointing
   at the target.
2. Add GitHub Actions output writing for `version_tag`, `tag_source`,
   `target_sha`, and `short_sha`.
3. Add `mint release select-tag` as a thin Cobra adapter.
4. Extend `action.yml` with `release-select-tag`, `requested-tag`, and typed
   outputs.
5. Update README, durable references, the constitution, and agent manifests.
6. Validate with focused Go tests, action YAML parsing, command help smoke
   tests, and repository checks.

## TESTING

- Temp Git repositories cover requested tag validation, highest tag selection,
  missing target tag failure, and invalid refs.
- CLI tests cover stdout and GitHub output behavior.
- Action tests parse `action.yml` and assert fixed allowlist wiring.
