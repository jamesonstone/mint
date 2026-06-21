---
kit_metadata_version: 1
artifact: plan
feature:
  id: 0007
  slug: release-state-ownership
  dir: 0007-release-state-ownership
summary: Implementation plan for making Mint own release-state operations.
parallelization_mode: rlm
skills:
  - name: rlm
    source: repo-local guide
    path: docs/agents/RLM.md
    trigger: "release ownership boundary; CLI/action behavior; workflow boundary"
    required: true
relationships:
  - type: depends_on
    target: docs/specs/0007-release-state-ownership/SPEC.md
    reason: This plan implements the release-state ownership contract.
references:
  - id: spec
    name: Release-state ownership spec
    type: feature_doc
    target: docs/specs/0007-release-state-ownership/SPEC.md
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
    used_for: tag creation, publish composition, and tests
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

1. Add `pkg/release` tag creation with strict SemVer validation, target
   validation, same-commit reuse, conflict failure, annotated tags, and optional
   push.
2. Add `pkg/release` publish composition over existing resolver, tag creation,
   and GitHub Release publishing.
3. Add `mint release tag` and `mint release publish` as thin CLI adapters.
4. Extend `action.yml` with `release-tag`, `release-publish`, and tag outputs.
5. Update the generated Docker workflow renderer to call Mint for tag creation
   while leaving Docker publishing in the generated workflow.
6. Update self-release workflow and documentation to prefer release-state
   commands.
7. Validate with Go tests, command help smoke tests, and diff checks.

## TESTING

- Temp Git repositories cover tag creation, reuse, conflicts, and no tag
  movement.
- Local HTTP servers cover GitHub Release behavior without live mutations.
- Action/workflow tests parse YAML and assert the fixed command allowlist.
