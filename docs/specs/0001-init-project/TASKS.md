---
kit_metadata_version: 1
artifact: tasks
feature:
  id: 0001
  slug: init-project
  dir: 0001-init-project
summary: Task plan for implementing the Mint project constitution as a documentation-only change with explicit validation evidence.
skills:
  - name: rlm
    source: repo-local guide
    path: docs/agents/RLM.md
    trigger: "broad/noisy repository discovery; analyze codebase; scan all files; large repository analysis; scan repository; recursive language model"
    required: true
relationships: []
references:
  - id: user-pasted-tasks-task
    name: Pasted task-plan request
    type: prompt
    target: /Users/jamesonstone/.codex/attachments/6cf2f260-452e-4e95-a789-42894b6531c7/pasted-text.txt
    relation: guides
    read_policy: must
    used_for: task format, required sections, done-condition requirements, and final response contract
    status: active
  - id: fixed-spec
    name: Fixed specification
    type: feature_doc
    target: docs/specs/0001-init-project/SPEC.md
    selector_type: artifact
    selector: SPEC.md
    relation: constrains
    read_policy: must
    used_for: binding requirements, non-goals, acceptance criteria, and edge cases
    status: active
  - id: implementation-plan
    name: Implementation plan
    type: feature_doc
    target: docs/specs/0001-init-project/PLAN.md
    selector_type: artifact
    selector: PLAN.md
    relation: constrains
    read_policy: must
    used_for: implementation sequence, component boundaries, validation strategy, and risks
    status: active
  - id: upstream-brainstorm
    name: Upstream brainstorm
    type: feature_doc
    target: docs/specs/0001-init-project/BRAINSTORM.md
    selector_type: artifact
    selector: BRAINSTORM.md
    relation: informs
    read_policy: conditional
    used_for: current repository facts, evidence inputs, and tradeoff rationale
    status: active
  - id: constitution
    name: Project constitution
    type: repo_doc
    target: docs/CONSTITUTION.md
    relation: implements
    read_policy: must
    used_for: primary implementation artifact and Kit-managed baseline preservation
    status: active
  - id: project-progress-summary
    name: Project progress summary
    type: repo_doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    relation: implements
    read_policy: must
    used_for: supporting artifact for highest completed feature phase tracking
    status: active
  - id: kit-map-init-project
    name: Kit map for init-project
    type: command
    target: kit map 0001-init-project
    selector_type: command
    selector: kit map 0001-init-project
    relation: verifies
    read_policy: evidence
    used_for: feature phase, relationship absence, and reference resolution
    status: active
  - id: agent-workflows
    name: Agent workflows
    type: repo_doc
    target: docs/agents/WORKFLOWS.md
    relation: constrains
    read_policy: must
    used_for: source-of-truth order and spec-driven workflow expectations
    status: active
  - id: agent-guardrails
    name: Agent guardrails
    type: repo_doc
    target: docs/agents/GUARDRAILS.md
    relation: constrains
    read_policy: must
    used_for: completion bar, placeholder cleanup, and documentation-only boundary
    status: active
  - id: agent-rlm
    name: RLM context routing
    type: skill
    target: docs/agents/RLM.md
    relation: guides
    read_policy: must
    used_for: selected task execution skill and discovery-first routing
    status: active
  - id: kit-config
    name: Kit project config
    type: config
    target: .kit.yaml
    relation: constrains
    read_policy: must
    used_for: feature paths, constitution path, and no out-of-order progression
    status: active
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Confirm implementation scope and baseline evidence | done | agent | |
| T002 | Rewrite the constitution contract | done | agent | T001 |
| T003 | Review constitution against SPEC and PLAN | done | agent | T002 |
| T004 | Update progress summary for implementation readiness | done | agent | T003 |
| T005 | Run required validation commands and remediate docs | done | agent | T004 |
| T006 | Perform final scope and evidence review | done | agent | T005 |

## TASK LIST

- [x] T001: Confirm implementation scope and baseline evidence [PLAN-APPROACH](./PLAN.md#approach), [PLAN-INTERFACES](./PLAN.md#interfaces)
- [x] T002: Rewrite the constitution contract [PLAN-COMPONENTS](./PLAN.md#components), [PLAN-DATA](./PLAN.md#data)
- [x] T003: Review constitution against SPEC and PLAN [PLAN-RISKS](./PLAN.md#risks), [PLAN-TESTING](./PLAN.md#testing)
- [x] T004: Update progress summary for implementation readiness [PLAN-COMPONENTS](./PLAN.md#components), [PLAN-INTERFACES](./PLAN.md#interfaces)
- [x] T005: Run required validation commands and remediate docs [PLAN-TESTING](./PLAN.md#testing)
- [x] T006: Perform final scope and evidence review [PLAN-RISKS](./PLAN.md#risks), [PLAN-TESTING](./PLAN.md#testing)

## TASK DETAILS

### T001
- **GOAL**: Establish the exact documentation-only implementation boundary before editing.
- **SCOPE**:
  - Re-read `docs/specs/0001-init-project/SPEC.md` and `docs/specs/0001-init-project/PLAN.md`.
  - Inspect current `docs/CONSTITUTION.md` and identify the Kit-managed baseline block.
  - Confirm no product code, tests, runtime config, generated artifacts, or machine-local state are in scope.
- **ACCEPTANCE**:
  - The implementation scope is limited to `docs/CONSTITUTION.md` plus `docs/PROJECT_PROGRESS_SUMMARY.md` if progress metadata needs alignment.
  - The managed block markers in `docs/CONSTITUTION.md` are identified before editing.
  - Evidence artifacts are the inspected file list and initial `git status --short --branch` output.
- **VERIFY**:
  - `git status --short --branch`
  - `sed -n '1,220p' docs/CONSTITUTION.md`
- **EXPECTED FILES**:
  - no files changed by this task
- **RISK**: Low; discovery-only step.
- **ROLLBACK**: not required.
- **NOTES**: Do not stage, commit, branch, push, or create GitHub delivery artifacts.

### T002
- **GOAL**: Replace placeholder constitution content with the project contract required by the fixed SPEC.
- **SCOPE**:
  - Populate `PRINCIPLES` with durable decision principles.
  - Populate `CONSTRAINTS` while preserving the Kit-managed baseline block unchanged.
  - Make `CHANGE CLASSIFICATION` concrete for spec-driven, ad hoc, and ad hoc-with-existing-spec work.
  - Populate `NON-GOALS` and `DEFINITIONS`.
  - Include current product intent and current implementation absence without inventing runtime behavior.
- **ACCEPTANCE**:
  - `docs/CONSTITUTION.md` contains no placeholder-only required section.
  - The Kit-managed baseline block remains present with the same markers and rules.
  - The constitution points detailed procedures to `docs/agents/*` and `docs/references/rules/*` instead of copying full procedures.
  - Evidence artifact is the `docs/CONSTITUTION.md` diff.
- **VERIFY**:
  - `git diff -- docs/CONSTITUTION.md`
  - `rg --files-without-match -e TODO -e TBD docs/CONSTITUTION.md`
- **EXPECTED FILES**:
  - `docs/CONSTITUTION.md`
- **RISK**: Medium; the main risk is overstating unimplemented architecture.
- **ROLLBACK**: Revert only this file's working-tree changes if review finds scope drift.
- **NOTES**: Use evidence-qualified wording for Go/tooling intent because no `go.mod` or source tree exists.

### T003
- **GOAL**: Prove the constitution content satisfies the fixed SPEC and does not add new scope.
- **SCOPE**:
  - Compare `docs/CONSTITUTION.md` against `SPEC.md` requirements `SPEC-01` through `SPEC-20`.
  - Confirm acceptance criteria `AC-01` through `AC-06` are addressed in the document text.
  - Confirm no product code, tests, runtime config, release algorithm, CI, external integration, or package-manager behavior was introduced.
- **ACCEPTANCE**:
  - Each required constitution section maps to the fixed SPEC.
  - Current facts and future intended direction are visibly separated.
  - Detailed procedures remain pointer-loaded.
  - Evidence artifact is a reviewed diff with no required follow-up edits outstanding.
- **VERIFY**:
  - `git diff -- docs/CONSTITUTION.md`
  - `rg -n -e docs/agents -e docs/references/rules -e go.mod -e "source tree" -e "not yet" docs/CONSTITUTION.md`
- **EXPECTED FILES**:
  - `docs/CONSTITUTION.md`
- **RISK**: Medium; content may be complete but too broad.
- **ROLLBACK**: Narrow the edited constitution text before moving on.
- **NOTES**: This is a review-and-remediate task; leave the task incomplete until any required wording fixes are made.

### T004
- **GOAL**: Keep the feature progress summary aligned with implementation readiness.
- **SCOPE**:
  - Update `docs/PROJECT_PROGRESS_SUMMARY.md` only if status, summary, approach, open items, or timestamp is stale after tasks/implementation readiness advances.
  - Keep open items as `none` unless a real blocker is discovered during T001-T003.
  - Avoid adding implementation claims beyond the constitution/doc-only scope.
- **ACCEPTANCE**:
  - `docs/PROJECT_PROGRESS_SUMMARY.md` reflects the highest completed artifact/active phase for `0001-init-project`.
  - Feature summary still points to the feature artifacts.
  - Evidence artifact is the `docs/PROJECT_PROGRESS_SUMMARY.md` diff or an explicit note that no change was needed.
- **VERIFY**:
  - `sed -n '1,220p' docs/PROJECT_PROGRESS_SUMMARY.md`
  - `git diff -- docs/PROJECT_PROGRESS_SUMMARY.md`
- **EXPECTED FILES**:
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
- **RISK**: Low; supporting metadata update.
- **ROLLBACK**: Restore the prior summary text if it no longer matches Kit phase output.
- **NOTES**: Do not edit generated Kit state files.

### T005
- **GOAL**: Run the explicit validation suite and fix documentation-only failures.
- **SCOPE**:
  - Run feature-map validation.
  - Run whitespace validation for touched docs.
  - Run placeholder validation for required docs.
  - Remediate only documentation issues in `docs/CONSTITUTION.md` or `docs/PROJECT_PROGRESS_SUMMARY.md`.
- **ACCEPTANCE**:
  - `kit map 0001-init-project` resolves the feature.
  - `git diff --check -- docs/CONSTITUTION.md docs/PROJECT_PROGRESS_SUMMARY.md` reports no errors.
  - Placeholder scan has no unresolved placeholder hits in `docs/CONSTITUTION.md`.
  - Evidence artifacts are the command outputs.
- **VERIFY**:
  - `kit map 0001-init-project`
  - `git diff --check -- docs/CONSTITUTION.md docs/PROJECT_PROGRESS_SUMMARY.md`
  - `rg --files-without-match -e TODO -e TBD docs/CONSTITUTION.md`
  - `rg --files-without-match -e TODO -e TBD docs/PROJECT_PROGRESS_SUMMARY.md`
- **EXPECTED FILES**:
  - `docs/CONSTITUTION.md`
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
- **RISK**: Low; validation may expose wording gaps.
- **ROLLBACK**: not required unless remediation creates scope drift; then revert the offending doc edit.
- **NOTES**: Do not claim tests passed; this feature uses documentation validation commands, not product tests.

### T006
- **GOAL**: Complete final implementation review with evidence that scope and safety constraints held.
- **SCOPE**:
  - Review final diff for `docs/CONSTITUTION.md` and `docs/PROJECT_PROGRESS_SUMMARY.md`.
  - Confirm no product files, tests, runtime config, generated artifacts, secrets, `.env`, or machine-local state were modified.
  - Confirm the Kit-managed baseline block remains present.
  - Record any skipped validation or residual risk in the final implementation response.
- **ACCEPTANCE**:
  - Final diff is documentation-only and limited to expected files.
  - `git status --short --branch` shows no unexpected implementation artifacts created by this work.
  - There are no blocked items.
  - Evidence artifacts are final diff and final status output.
- **VERIFY**:
  - `git diff -- docs/CONSTITUTION.md docs/PROJECT_PROGRESS_SUMMARY.md`
  - `git status --short --branch`
  - `rg -n -e "BEGIN KIT-MANAGED BASELINE RULES" -e "END KIT-MANAGED BASELINE RULES" docs/CONSTITUTION.md`
- **EXPECTED FILES**:
  - `docs/CONSTITUTION.md`
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
- **RISK**: Low; final review only.
- **ROLLBACK**: Revert only unexpected documentation edits if found.
- **NOTES**: Leave unrelated dirty or untracked files untouched.

## DEPENDENCIES

1. Task order is linear: T001 -> T002 -> T003 -> T004 -> T005 -> T006.
2. No external blockers or missing decisions are known.
3. GitHub delivery is out of scope unless the user explicitly requests it.

## NOTES

1. This task plan is documentation-only.
2. `parallelization_mode: "rlm"` remains the execution metadata from `PLAN.md`, but these tasks should execute serially because they touch the same constitution/progress artifacts.
3. `docs/CONSTITUTION.md` is the only primary implementation artifact; `docs/PROJECT_PROGRESS_SUMMARY.md` is supporting metadata.
4. Reflection review fixed the verifier command contract for placeholder checks; no product code, source manifests, or runtime artifacts were introduced.
5. Project refresh advisory: no project refresh needed.

<!-- REFLECTION_COMPLETE -->
