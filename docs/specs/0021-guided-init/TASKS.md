---
kit_metadata_version: 1
artifact: tasks
feature:
  id: "0021"
  slug: guided-init
  dir: 0021-guided-init
source_artifacts:
  - docs/specs/0021-guided-init/SPEC.md
  - docs/specs/0021-guided-init/PLAN.md
---
# Guided init tasks

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| --- | --- | --- | --- | --- |
| T001 | Add guided per-environment config | done | agent | |
| T002 | Generate the workflows | done | agent | T001 |
| T003 | Document and validate generated setup | done | agent | T002 |
| T004 | Complete independent review | done | agent | T003 |
| T005 | Verify hosted single-project promotion | in progress | agent | T004 |

## TASK LIST

- [x] T001: Add guided per-environment config.
- [x] T002: Generate the workflows.
- [x] T003: Document and validate generated setup.
- [x] T004: Complete independent review.
- [ ] T005: Verify hosted single-project promotion.

## TASK DETAILS

### T001

Validate legacy and per-environment settings. Confirm guided choices before creating files and preserve existing files.

### T002

Generate push and dispatch entry points that share the promotion job. Use the exact retained artifact on dispatch without a build path.

### T003

Explain setup, project hooks and compatibility. Run native tests, Actionlint, a local promotion fixture and the documentation build.

### T004

Review the complete change in a fresh context and verify substantive corrections.

### T005

Run the generated workflows in the private single-project testbed. Verify both runtime receipts against the retained build. Confirm the manual build job was skipped.

## DEPENDENCIES

The config type supplies validated environment modes to the workflow renderer. Complete both before testing the compiled command.

## NOTES

Provider smoke tests and Kit integration are separate.
