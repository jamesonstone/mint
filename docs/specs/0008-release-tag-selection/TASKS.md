---
kit_metadata_version: 1
artifact: tasks
feature:
  id: 0008
  slug: release-tag-selection
  dir: 0008-release-tag-selection
source_artifacts:
  - docs/specs/0008-release-tag-selection/SPEC.md
  - docs/specs/0008-release-tag-selection/PLAN.md
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Add read-only selector package behavior | done | agent | |
| T002 | Add selector GitHub output fields | done | agent | T001 |
| T003 | Add `mint release select-tag` CLI adapter | done | agent | T002 |
| T004 | Extend action command allowlist and outputs | done | agent | T003 |
| T005 | Update README, references, constitution, and agent manifests | done | agent | T004 |
| T006 | Add package, CLI, and action tests | done | agent | T005 |
| T007 | Run final validation and readiness checks | done | agent | T006 |

## TASK LIST

- [x] T001: Add read-only selector package behavior.
- [x] T002: Add selector GitHub output fields.
- [x] T003: Add `mint release select-tag` CLI adapter.
- [x] T004: Extend action command allowlist and outputs.
- [x] T005: Update README, references, constitution, and agent manifests.
- [x] T006: Add package, CLI, and action tests.
- [x] T007: Run final validation and readiness checks.

## VERIFY

- `go test ./...`
- `go vet ./...`
- `make build`
- `go run ./cmd/mint release select-tag --help`
- `git diff --check`
