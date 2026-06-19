---
kit_metadata_version: 1
artifact: tasks
feature:
  id: 0003
  slug: github-action
  dir: 0003-github-action
summary: Task plan for exposing the Mint CLI as a public GitHub composite action.
relationships:
  - type: depends_on
    feature: 0002-cli-patterns
    reason: The action builds and runs the existing Mint CLI.
references:
  - id: spec
    name: GitHub Action specification
    type: feature_doc
    target: docs/specs/0003-github-action/SPEC.md
    relation: constrains
    read_policy: must
    used_for: requirements and acceptance criteria
    status: active
  - id: plan
    name: GitHub Action plan
    type: feature_doc
    target: docs/specs/0003-github-action/PLAN.md
    relation: constrains
    read_policy: must
    used_for: implementation sequence and verification strategy
    status: active
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Add root GitHub Action metadata | done | agent | |
| T002 | Document GitHub Action quick start | done | agent | T001 |
| T003 | Update durable project docs | done | agent | T001,T002 |
| T004 | Run validation suite | done | agent | T001,T002,T003 |

## TASK LIST

- [x] T001: Add root GitHub Action metadata [PLAN-APPROACH](./PLAN.md#approach)
- [x] T002: Document GitHub Action quick start [PLAN-COMPONENTS](./PLAN.md#components)
- [x] T003: Update durable project docs [PLAN-COMPONENTS](./PLAN.md#components)
- [x] T004: Run validation suite [PLAN-TESTING](./PLAN.md#testing)

## TASK DETAILS

### T001
- **GOAL**: Add public GitHub Action metadata that builds and exposes the Mint CLI.
- **VERIFY**:
  - `ruby -ryaml -e "YAML.load_file('action.yml')"`
  - `go build -ldflags "-X github.com/jamesonstone/mint/pkg/cli.Version=dev" -o bin/mint-action ./cmd/mint`
- **EXPECTED FILES**:
  - `action.yml`

### T002
- **GOAL**: Show users how to call Mint from a GitHub Actions workflow.
- **VERIFY**:
- `rg -n -e "GitHub Action Quick Start" -e "uses: jamesonstone/mint" README.md`
- **EXPECTED FILES**:
  - `README.md`

### T003
- **GOAL**: Keep durable docs aligned with the new action surface.
- **VERIFY**:
  - `rg -n -e "action.yml" -e "GitHub Action" docs/CONSTITUTION.md docs/references/tooling.md docs/references/testing.md docs/PROJECT_PROGRESS_SUMMARY.md`
- **EXPECTED FILES**:
  - `docs/CONSTITUTION.md`
  - `docs/references/tooling.md`
  - `docs/references/testing.md`
  - `docs/PROJECT_PROGRESS_SUMMARY.md`

### T004
- **GOAL**: Validate the complete GitHub Action wrapper change.
- **VERIFY**:
  - `ruby -ryaml -e "YAML.load_file('action.yml')"`
  - `go build -ldflags "-X github.com/jamesonstone/mint/pkg/cli.Version=dev" -o bin/mint-action ./cmd/mint`
  - `./bin/mint-action version`
  - `./bin/mint-action --help`
  - `go test ./...`
  - `go vet ./...`
  - `make build`
  - `git diff --check -- action.yml README.md docs`
  - `kit map 0003-github-action`
- **EXPECTED FILES**:
  - all changed files

## NOTES

1. The GitHub Action wraps the CLI and does not implement release behavior.
2. The supported action commands are `version`, `help`, and `none`.
3. Reflection review confirmed the action remains a CLI wrapper and does not add release automation.

<!-- REFLECTION_COMPLETE -->
