---
kit_metadata_version: 1
artifact: tasks
feature:
  id: 0002
  slug: cli-patterns
  dir: 0002-cli-patterns
summary: Task plan for adopting Kit-style CLI, README, Makefile, and build patterns.
skills:
  - name: rlm
    source: repo-local guide
    path: docs/agents/RLM.md
    trigger: "compare another repository; adopt external repo patterns; broad repository context"
    required: true
relationships:
  - type: depends_on
    feature: 0001-init-project
    reason: The completed constitution controls source-of-truth and implementation evidence wording.
references:
  - id: spec
    name: CLI pattern specification
    type: feature_doc
    target: docs/specs/0002-cli-patterns/SPEC.md
    relation: constrains
    read_policy: must
    used_for: requirements and acceptance criteria
    status: active
  - id: plan
    name: CLI pattern plan
    type: feature_doc
    target: docs/specs/0002-cli-patterns/PLAN.md
    relation: constrains
    read_policy: must
    used_for: implementation sequence and validation strategy
    status: active
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Add Kit-style Go CLI scaffold | done | agent | |
| T002 | Add Makefile, hooks, and ignore rules | done | agent | T001 |
| T003 | Rewrite README with Kit-style structure | done | agent | T001 |
| T004 | Update durable project docs | done | agent | T001,T002,T003 |
| T005 | Run validation suite | done | agent | T001,T002,T003,T004 |

## TASK LIST

- [x] T001: Add Kit-style Go CLI scaffold [PLAN-APPROACH](./PLAN.md#approach)
- [x] T002: Add Makefile, hooks, and ignore rules [PLAN-COMPONENTS](./PLAN.md#components)
- [x] T003: Rewrite README with Kit-style structure [PLAN-COMPONENTS](./PLAN.md#components)
- [x] T004: Update durable project docs [PLAN-COMPONENTS](./PLAN.md#components)
- [x] T005: Run validation suite [PLAN-TESTING](./PLAN.md#testing)

## TASK DETAILS

### T001
- **GOAL**: Add a minimal Kit-style Go/Cobra command surface.
- **VERIFY**:
  - `go test ./...`
  - `go vet ./...`
- **EXPECTED FILES**:
  - `go.mod`
  - `go.sum`
  - `cmd/mint/main.go`
  - `pkg/cli/root.go`
  - `pkg/cli/root_help.go`
  - `pkg/cli/human_output.go`
  - `pkg/cli/version.go`
  - `pkg/cli/version_test.go`
  - `pkg/cli/root_help_test.go`

### T002
- **GOAL**: Add Kit-style build and hook patterns.
- **VERIFY**:
  - `make build`
  - `./bin/mint version`
- **EXPECTED FILES**:
  - `Makefile`
  - `.githooks/pre-commit`
  - `.gitignore`

### T003
- **GOAL**: Rewrite README in Kit's visual/operational style.
- **VERIFY**:
  - `sed -n '1,220p' README.md`
- **EXPECTED FILES**:
  - `README.md`

### T004
- **GOAL**: Keep durable docs aligned with the new implemented CLI/build facts.
- **VERIFY**:
  - `rg -n -e "Go/Cobra" -e "Makefile" docs/CONSTITUTION.md docs/references/tooling.md docs/references/testing.md docs/PROJECT_PROGRESS_SUMMARY.md`
- **EXPECTED FILES**:
  - `docs/CONSTITUTION.md`
  - `docs/references/tooling.md`
  - `docs/references/testing.md`
  - `docs/PROJECT_PROGRESS_SUMMARY.md`

### T005
- **GOAL**: Validate the complete change set.
- **VERIFY**:
  - `go test ./...`
  - `go vet ./...`
  - `make build`
  - `./bin/mint version`
  - `./bin/mint --help`
  - `git diff --check -- README.md Makefile go.mod go.sum cmd pkg docs .gitignore .githooks`
  - `kit map 0002-cli-patterns`
- **EXPECTED FILES**:
  - all changed files

## NOTES

1. This feature intentionally does not implement release computation or changelog generation.
2. The completed CLI surface is root help, `--version`, and `version`.
3. Reflection review confirmed the change follows the Kit-style CLI/build scaffold without adding release semantics.

<!-- REFLECTION_COMPLETE -->
