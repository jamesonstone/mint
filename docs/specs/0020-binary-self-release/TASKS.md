---
kit_metadata_version: 1
artifact: tasks
feature:
  id: "0020"
  slug: binary-self-release
  dir: 0020-binary-self-release
source_artifacts:
  - docs/specs/0020-binary-self-release/SPEC.md
  - docs/specs/0020-binary-self-release/PLAN.md
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| --- | --- | --- | --- | --- |
| T001 | Implement configuration reader | done | agent | |
| T002 | Implement binary promotion and publication | done | agent | T001 |
| T003 | Validate and document the pipeline | in progress | agent | T002 |

## TASK LIST

- [x] T001: Implement configuration reader.
- [x] T002: Implement binary promotion and publication.
- [ ] T003: Validate and document the pipeline.

## TASK DETAILS

### T001

Validate artifact type, deployment mode, and ordered environments. Print machine-readable JSON.

### T002

Retain the binary once and verify the same artifact through staging and production. Publish only with the expected GitHub asset digest.

### T003

Run native, workflow, and documentation tests. Observe hosted read-only promotion checks.

## DEPENDENCIES

Publication runs on the trusted default branch after a merge.

## NOTES

The compressed Actions artifact and the released executable have distinct digests. Verify the executable bytes.
