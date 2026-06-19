---
kit_metadata_version: 1
artifact: tasks
feature:
  id: 0007
  slug: release-state-ownership
  dir: 0007-release-state-ownership
source_artifacts:
  - docs/specs/0007-release-state-ownership/SPEC.md
  - docs/specs/0007-release-state-ownership/PLAN.md
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Add release tag package behavior | done | agent | |
| T002 | Add release publish package composition | done | agent | T001 |
| T003 | Add `mint release tag` and `mint release publish` CLI adapters | done | agent | T002 |
| T004 | Extend action command allowlist and outputs | done | agent | T003 |
| T005 | Update self-release and generated workflow boundaries | done | agent | T004 |
| T006 | Add tests for tag, publish, CLI, action, and workflow behavior | done | agent | T005 |
| T007 | Update README and durable docs | done | agent | T006 |
| T008 | Run final validation and readiness checks | done | agent | T007 |

## TASK LIST

- [x] T001: Add release tag package behavior.
- [x] T002: Add release publish package composition.
- [x] T003: Add `mint release tag` and `mint release publish` CLI adapters.
- [x] T004: Extend action command allowlist and outputs.
- [x] T005: Update self-release and generated workflow boundaries.
- [x] T006: Add tests for tag, publish, CLI, action, and workflow behavior.
- [x] T007: Update README and durable docs.
- [x] T008: Run final validation and readiness checks.

## VERIFY

- `go test ./...`
- `go run ./cmd/mint release resolve --commitish HEAD`
- `go run ./cmd/mint release tag --help`
- `go run ./cmd/mint release github --help`
- `go run ./cmd/mint release publish --help`
- `go vet ./...`
- `make build`
- `git diff --check`
