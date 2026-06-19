---
kit_metadata_version: 1
artifact: tasks
feature:
  id: 0004
  slug: changelog-generation
  dir: 0004-changelog-generation
summary: Task plan for conventional-commit CHANGELOG.md generation.
relationships:
  - type: depends_on
    feature: 0002-cli-patterns
    reason: The new behavior is exposed through the existing Mint CLI structure.
references:
  - id: spec
    name: Changelog generation specification
    type: feature_doc
    target: docs/specs/0004-changelog-generation/SPEC.md
    relation: constrains
    read_policy: must
    used_for: requirements and acceptance criteria
    status: active
  - id: plan
    name: Changelog generation plan
    type: feature_doc
    target: docs/specs/0004-changelog-generation/PLAN.md
    relation: constrains
    read_policy: must
    used_for: implementation sequence and validation strategy
    status: active
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Add changelog generation core package | done | agent | |
| T002 | Add CLI command and root flag routing | done | agent | T001 |
| T003 | Add fixture-based tests | done | agent | T001,T002 |
| T004 | Update README, action, and durable docs | done | agent | T001,T002 |
| T005 | Run validation suite | done | agent | T001,T002,T003,T004 |

## TASK LIST

- [x] T001: Add changelog generation core package [PLAN-APPROACH](./PLAN.md#approach)
- [x] T002: Add CLI command and root flag routing [PLAN-APPROACH](./PLAN.md#approach)
- [x] T003: Add fixture-based tests [PLAN-TESTING](./PLAN.md#testing)
- [x] T004: Update README, action, and durable docs [PLAN-COMPONENTS](./PLAN.md#components)
- [x] T005: Run validation suite [PLAN-TESTING](./PLAN.md#testing)

## TASK DETAILS

### T001
- **GOAL**: Implement the generator API, git range collection, parsing, grouping, rendering, and atomic write behavior.
- **VERIFY**:
  - `go test ./pkg/changelog`
- **EXPECTED FILES**:
  - `pkg/changelog/generator.go`
  - `pkg/changelog/generator_test.go`

### T002
- **GOAL**: Expose generation through `mint changelog` and the requested root-level flags.
- **VERIFY**:
  - `go test ./pkg/cli`
  - `go build -o bin/mint ./cmd/mint`
  - `./bin/mint changelog --help`
- **EXPECTED FILES**:
  - `pkg/cli/changelog.go`
  - `pkg/cli/root.go`
  - `pkg/cli/root_help.go`

### T003
- **GOAL**: Cover feat, fix, breaking, excluded types, issue extraction, duplicate version, and tag failures with tests.
- **VERIFY**:
  - `go test ./...`
- **EXPECTED FILES**:
  - `pkg/changelog/generator_test.go`

### T004
- **GOAL**: Document the new command and action integration.
- **VERIFY**:
  - `rg -n -e "changelog" -e "CHANGELOG" README.md action.yml docs/CONSTITUTION.md docs/references/tooling.md docs/references/testing.md docs/PROJECT_PROGRESS_SUMMARY.md`
- **EXPECTED FILES**:
  - `README.md`
  - `action.yml`
  - `docs/CONSTITUTION.md`
  - `docs/references/tooling.md`
  - `docs/references/testing.md`
  - `docs/PROJECT_PROGRESS_SUMMARY.md`

### T005
- **GOAL**: Validate the complete feature.
- **VERIFY**:
  - `go test ./...`
  - `go vet ./...`
  - `make build`
  - `./bin/mint --help`
  - `./bin/mint changelog --help`
  - `git diff --check -- README.md action.yml docs pkg`
  - `kit map 0004-changelog-generation`
- **EXPECTED FILES**:
  - all changed files

## NOTES

1. The implementation must fail closed for invalid tags, empty ranges, duplicate versions, and malformed existing release headers.
2. Reflection review confirmed the generator remains changelog-only and does not add version calculation, tagging, publishing, or GitHub release behavior.

<!-- REFLECTION_COMPLETE -->
