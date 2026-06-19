---
kit_metadata_version: 1
artifact: tasks
feature:
  id: 0006
  slug: github-release-publishing
  dir: 0006-github-release-publishing
source_artifacts:
  - docs/specs/0006-github-release-publishing/SPEC.md
  - docs/specs/0006-github-release-publishing/PLAN.md
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Implement GitHub Release API client and output writer | done | agent | |
| T002 | Add package tests for create, existing, validation, and output behavior | done | agent | T001 |
| T003 | Add `mint release github` CLI adapter | done | agent | T001 |
| T004 | Extend composite action inputs, outputs, and command allowlist | done | agent | T003 |
| T005 | Configure Mint self-release workflow | done | agent | T004 |
| T006 | Update README and durable docs | done | agent | T005 |
| T007 | Run final validation and readiness checks | done | agent | T006 |

## TASK LIST

- [x] T001: Implement GitHub Release API client and output writer.
- [x] T002: Add package tests for create, existing, validation, and output behavior.
- [x] T003: Add `mint release github` CLI adapter.
- [x] T004: Extend composite action inputs, outputs, and command allowlist.
- [x] T005: Configure Mint self-release workflow.
- [x] T006: Update README and durable docs.
- [x] T007: Run final validation and readiness checks.

## TASK DETAILS

### T001

- **GOAL**: Add `pkg/release` support for GitHub Release publishing.
- **VERIFY**: `go test ./pkg/release -run GitHubRelease`

### T002

- **GOAL**: Cover API create, existing-release idempotency, validation, and
  action output behavior without live GitHub calls.
- **VERIFY**: `go test ./pkg/release -run GitHubRelease`

### T003

- **GOAL**: Expose the publisher as `mint release github`.
- **VERIFY**: `go test ./pkg/cli -run ReleaseGitHub`

### T004

- **GOAL**: Expose the CLI behavior through the composite action command
  allowlist.
- **VERIFY**: `go test ./pkg/release -run Action`

### T005

- **GOAL**: Add the repository workflow that uses Mint to publish Mint.
- **VERIFY**: `go test ./pkg/release -run SelfReleaseWorkflow`

### T006

- **GOAL**: Update user-facing and durable project docs.
- **VERIFY**: Review `README.md`, `docs/CONSTITUTION.md`,
  `docs/references/tooling.md`, `docs/references/testing.md`, and
  `docs/PROJECT_PROGRESS_SUMMARY.md`.

### T007

- **GOAL**: Complete repository validation and final diff review.
- **VERIFY**: `go fmt ./...`, `go test ./...`, `go vet ./...`, `make build`,
  and `git diff --check`.
