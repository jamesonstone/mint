---
kit_metadata_version: 1
artifact: tasks
feature:
  id: "0019"
  slug: direct-promotion-v2
  dir: 0019-direct-promotion-v2
source_artifacts:
  - docs/specs/0019-direct-promotion-v2/SPEC.md
  - docs/specs/0019-direct-promotion-v2/PLAN.md
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| --- | --- | --- | --- | --- |
| T001 | Prove staged and manual promotion | done | agent | |
| T002 | Verify refusal, failure, and interruption | done | agent | T001 |
| T003 | Remove the controller and preserve release primitives | done | agent | |
| T004 | Update public and agent guides | done | agent | T003 |
| T005 | Record validation for review | done | agent | T001, T002, T003, T004 |

## TASK LIST

- [x] T001: Validate continuous staging and manual production in both private testbeds.
- [x] T002: Validate staging evidence refusal and failed or interrupted runtime verification.
- [x] T003: Remove the Mint controller and maintain remaining release primitives.
- [x] T004: Update Mint's public and agent guides and run project checks.
- [x] T005: Record evidence for the Mint pull request.

## TASK DETAILS

### T001

Compare hosted staging and production receipts with each retained artifact. Inspect the shared promotion job and confirm zero production build jobs.

### T002

Dispatch the missing, failed, cancelled, wrong-project, and wrong-digest cases. Confirm that failed and cancelled runtime verification emits no successful receipt or publication marker.

### T003

Remove controller-only Go and Python code, CLI and GitHub Action surfaces, and CI. Validate the remaining release primitives on the `/v2` module.

### T004

Describe the direct deployment path and migration from v1 in the site and agent guides. Validate both LLM exports.

### T005

Record native validation, hosted proof, and release dependencies in the spec, program, and review description.

## DEPENDENCIES

Provider smoke and downstream Kit migration are separate. GitHub-hosted fixture runs are available to repository members.

## NOTES

The specification records the results. Work Ledger WL-59 keeps private run links.
