---
kit_metadata_version: 1
artifact: brainstorm
feature:
  id: 0001
  slug: init-project
  dir: 0001-init-project
references:
  - id: user-pasted-task
    name: Pasted init-project research request
    type: prompt
    target: /Users/jamesonstone/.codex/attachments/49a8e96c-d4ec-4809-ac61-21f1b97bbbdb/pasted-text.txt
    relation: guides
    read_policy: must
    used_for: user thesis, task boundaries, final artifact requirements, clarification workflow
    status: active
  - id: constitution
    name: Project constitution
    type: repo_doc
    target: docs/CONSTITUTION.md
    relation: constrains
    read_policy: must
    used_for: canonical project contract and target document for the follow-up specification
    status: active
  - id: project-progress-summary
    name: Project progress summary
    type: repo_doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    relation: constrains
    read_policy: must
    used_for: confirms init-project is at brainstorm phase and no higher artifact exists
    status: active
  - id: kit-map-init-project
    name: Kit map for init-project
    type: command
    target: kit map 0001-init-project
    relation: informs
    read_policy: evidence
    used_for: verified feature phase, required global docs, reference links, and absence of feature relationships
    status: active
  - id: repo-file-inventory
    name: Repository file inventory
    type: command
    target: find . -maxdepth 4 -type f -not -path './.git/*' -print
    relation: informs
    read_policy: evidence
    used_for: established that current repository contains docs and config but no source tree or dependency manifest
    status: active
  - id: readme
    name: README
    type: repo_doc
    target: README.md
    relation: informs
    read_policy: must
    used_for: only explicit product intent statement for Mint
    status: active
  - id: agent-routing-readme
    name: Agent routing entrypoint
    type: repo_doc
    target: docs/agents/README.md
    relation: guides
    read_policy: must
    used_for: task classification and just-in-time document loading
    status: active
  - id: workflows
    name: Agent workflows
    type: repo_doc
    target: docs/agents/WORKFLOWS.md
    relation: guides
    read_policy: must
    used_for: classified this as spec-driven brainstorming and confirmed BRAINSTORM.md is research context
    status: active
  - id: guardrails
    name: Agent guardrails
    type: repo_doc
    target: docs/agents/GUARDRAILS.md
    relation: constrains
    read_policy: must
    used_for: completion bar, documentation-only boundary, and placeholder removal requirements
    status: active
  - id: rlm
    name: RLM context routing
    type: repo_doc
    target: docs/agents/RLM.md
    relation: guides
    read_policy: must
    used_for: narrowed prior-work and repository discovery to relevant artifacts
    status: active
  - id: tooling
    name: Agent tooling guidance
    type: repo_doc
    target: docs/agents/TOOLING.md
    relation: guides
    read_policy: must
    used_for: confirmed skill/front-matter lookup and project-directory workflow expectations
    status: active
  - id: agents-routing
    name: AGENTS routing table
    type: repo_doc
    target: AGENTS.md
    relation: guides
    read_policy: conditional
    used_for: confirmed repo-local docs under docs/ are the system of record
    status: active
  - id: claude-routing
    name: CLAUDE routing table
    type: repo_doc
    target: CLAUDE.md
    relation: guides
    read_policy: conditional
    used_for: confirmed alternate agent entrypoint mirrors AGENTS.md
    status: active
  - id: copilot-instructions
    name: GitHub Copilot repository instructions
    type: repo_doc
    target: .github/copilot-instructions.md
    relation: guides
    read_policy: conditional
    used_for: confirmed GitHub Copilot instructions also route into docs/ and avoid monolithic guidance
    status: active
  - id: kit-config
    name: Kit project config
    type: config
    target: .kit.yaml
    relation: constrains
    read_policy: must
    used_for: confirmed goal confidence threshold, specs directory, skills directory, constitution path, and managed rulesets
    status: active
  - id: gitignore
    name: Git ignore rules
    type: config
    target: .gitignore
    relation: informs
    read_policy: evidence
    used_for: evidence of Go-oriented ignore baseline and Kit local artifact exclusions
    status: active
  - id: envrc
    name: direnv loader
    type: config
    target: .envrc
    relation: informs
    read_policy: evidence
    used_for: evidence that local environment variables are loaded from .env when present
    status: active
  - id: coderabbit
    name: CodeRabbit config
    type: config
    target: .coderabbit.yaml
    relation: informs
    read_policy: evidence
    used_for: evidence that automated review is configured to ignore docs and top-level agent routing files
    status: active
  - id: pr-template
    name: Pull request template
    type: repo_doc
    target: .github/pull_request_template.md
    relation: informs
    read_policy: conditional
    used_for: evidence of current PR description, test, and ticket headings
    status: active
  - id: references-readme
    name: References index
    type: repo_doc
    target: docs/references/README.md
    relation: guides
    read_policy: conditional
    used_for: confirmed durable references should live under docs/references and be linked from feature metadata
    status: active
  - id: testing-reference
    name: Testing reference
    type: repo_doc
    target: docs/references/testing.md
    relation: informs
    read_policy: conditional
    used_for: confirmed no project-specific testing strategy exists yet
    status: active
  - id: tooling-reference
    name: Tooling reference
    type: repo_doc
    target: docs/references/tooling.md
    relation: informs
    read_policy: conditional
    used_for: confirmed no project-specific tooling reference exists yet
    status: active
  - id: external-systems-reference
    name: External systems reference
    type: repo_doc
    target: docs/references/external-systems.md
    relation: informs
    read_policy: conditional
    used_for: confirmed no durable external systems are documented yet
    status: active
  - id: safety-guardrails-ruleset
    name: Safety guardrails ruleset
    type: ruleset
    target: docs/references/rules/safety-guardrails.md
    relation: constrains
    read_policy: conditional
    used_for: git, GitHub, identity, secret, protected-branch, and failure-handling constraints
    status: active
  - id: work-lane-gating-ruleset
    name: Work lane gating ruleset
    type: ruleset
    target: docs/references/rules/work-lane-gating.md
    relation: constrains
    read_policy: conditional
    used_for: distinguishes documentation/spec phases from implementation lanes
    status: active
  - id: github-pr-delivery-ruleset
    name: GitHub PR delivery ruleset
    type: ruleset
    target: docs/references/rules/github-pr-delivery.md
    relation: constrains
    read_policy: conditional
    used_for: on-demand issue, branch, commit, push, and PR workflow after explicit consent
    status: active
  - id: kit-capabilities-usage-ruleset
    name: Kit capabilities usage ruleset
    type: ruleset
    target: docs/references/rules/kit-capabilities-usage.md
    relation: guides
    read_policy: conditional
    used_for: command-discovery guidance when Kit command behavior is uncertain
    status: active
  - id: feature-notes
    name: Feature notes
    type: notes
    target: docs/notes/0001-init-project
    relation: informs
    read_policy: conditional
    used_for: optional pre-brainstorm research input
    status: optional
relationships: []
---
# BRAINSTORM

## SUMMARY

Mint is currently a newly initialized, Kit-managed repository with a terse product intent in `README.md` but no source tree, module manifest, runtime architecture, or project-specific constitution content yet. The likely next step is a focused `SPEC.md` for replacing the placeholder `docs/CONSTITUTION.md` with durable rules that capture the present documentation-first workflow, the inferred Go/release-tool direction, and explicit boundaries for future implementation decisions.

## USER THESIS

Please update /Users/jamesonstone/go/src/github.com/jamesonstone/mint/docs/CONSTITUTION.md with all patterns, strategy, implementation details, process, and long-term vision for this project.
This document will drive the "rules for development" going forward.

Analyze the codebase at /Users/jamesonstone/go/src/github.com/jamesonstone/mint to extract:

- Architectural patterns and conventions
- Code style and naming conventions
- Dependencies and their purposes
- Non-negotiable constraints
- Project goals and non-goals

Rules:

- PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature at all times

## RELATIONSHIPS

No prior feature relationships. `kit map 0001-init-project` reports no incoming relationships, no outgoing relationships, and only the optional `docs/notes/0001-init-project` reference.

## CODEBASE FINDINGS

1. `README.md` is the only product-facing intent statement. It names the project `mint` and describes the desired outcome as computing the next version, writing the changelog, and minting the release. That supports a release-automation direction, but it does not yet define CLI commands, package boundaries, supported versioning semantics, changelog format, or release targets.
2. There is no implemented runtime architecture yet. The repository inventory contains documentation and configuration files only; no `go.mod`, `package.json`, `pyproject.toml`, `Cargo.toml`, `Makefile`, `Dockerfile`, source directory, tests, or CI workflow was found. Any constitution language about concrete implementation patterns must therefore be framed as intended direction or future constraints, not as extracted code behavior.
3. `.gitignore` is Go-oriented and ignores compiled binaries, Go test binaries, coverage/profile artifacts, optional vendor directories, and Go workspace files. This is the strongest local signal that Mint is intended to become a Go project, but the absence of `go.mod` means the language choice is not yet committed by implementation.
4. `.gitignore` also ignores Kit local runtime artifacts under `.kit/runs/`, `.kit/loops/`, `.kit/state.json`, `.kit/cache/`, `.kit/tmp/`, `.kit/temp/`, `.kit/*.tmp`, and `.kit/*.lock`. That reinforces the project preference for committing durable docs and config while keeping generated local execution state out of version control.
5. `.envrc` uses `dotenv_if_exists`; `.env` exists locally but is empty and ignored. The future constitution should treat local environment as optional and private, and should prohibit committing secrets or machine-local state.
6. `.kit.yaml` sets `goal_percentage: 95`, `specs_dir: docs/specs`, `skills_dir: .agents/skills`, and `constitution_path: docs/CONSTITUTION.md`. It also disables out-of-order workflow progression. The constitution should preserve the formal workflow expectation that ambiguous work is driven to high confidence before implementation.
7. `docs/CONSTITUTION.md` currently contains placeholders for principles, constraints, non-goals, and definitions, plus Kit-managed baseline rules. It already states that `docs/CONSTITUTION.md` is the canonical project contract and that `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` must stay aligned with repo-local docs.
8. `docs/PROJECT_PROGRESS_SUMMARY.md` correctly shows `0001-init-project` in `brainstorm` phase. No `SPEC.md`, `PLAN.md`, or `TASKS.md` exists yet, so the highest completed artifact is currently `BRAINSTORM.md`.
9. `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` are intentionally short routing tables. All three direct agents to start with `docs/agents/README.md`, treat `docs/` as the source of truth, and avoid creating an always-loaded monolithic instruction file.
10. `docs/agents/WORKFLOWS.md` classifies this work as spec-driven because it updates project-wide rules and already has feature docs. It also defines authority order: safety and permissions, current user request, `docs/CONSTITUTION.md`, formal feature artifacts, and repo conventions.
11. `docs/agents/RLM.md` requires progressive disclosure for broad repository analysis. This feature followed that by using `kit map 0001-init-project`, `docs/PROJECT_PROGRESS_SUMMARY.md`, and targeted file inventory before reading broader references.
12. `docs/agents/GUARDRAILS.md` requires placeholder sections in feature docs to be populated, prevents claims about tests that did not run, and keeps canonical front matter references current when feature docs are touched.
13. `docs/agents/TOOLING.md` defines `.agents/skills/*/SKILL.md` as the canonical repo-local skills location, but this repository currently has no `.agents/skills` files and the active brainstorm front matter has no `skills` entries.
14. `docs/references/testing.md`, `docs/references/tooling.md`, and `docs/references/external-systems.md` are starter placeholders with no project-specific content. Future constitution work should not invent detailed testing, tooling, or integration rules without either implementation evidence or explicit product decisions.
15. `docs/references/rules/safety-guardrails.md` and `docs/references/rules/work-lane-gating.md` materially constrain process. Documentation and Kit phases must not be forced into PR workflow by agent initiative, while git/GitHub mutations require recon, identity checks, and failure handling.
16. `docs/references/rules/github-pr-delivery.md` defines an explicit opt-in GitHub delivery process with issue-number branches in `GH-123` form, explicit staging, required PR template use, and no agent attribution. Constitution updates should point to this ruleset for delivery detail instead of duplicating it.
17. `.coderabbit.yaml` excludes `docs/**`, `AGENTS.md`, and `CLAUDE.md` from automated review. Constitution/spec work should assume automated review may not cover docs changes and should rely on direct human/agent review for documentation quality.
18. `.github/pull_request_template.md` contains basic `Description`, `How to Test`, and `Ticket` sections. PR delivery rules add stricter process around that template when a PR is explicitly requested.

## AFFECTED FILES

1. `docs/CONSTITUTION.md` - target of the next specification and eventual documentation change; currently mostly placeholder content plus Kit-managed baseline rules.
2. `docs/specs/0001-init-project/BRAINSTORM.md` - current research artifact and source of truth for this phase.
3. `docs/PROJECT_PROGRESS_SUMMARY.md` - must continue to reflect the highest completed artifact for `0001-init-project`; currently correct at `brainstorm`.
4. `README.md` - only concise product vision signal for Mint as a release/version/changelog tool.
5. `.kit.yaml` - defines Kit workflow paths, 95 percent confidence target, no out-of-order progression, and installed managed rulesets.
6. `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md` - short agent routing tables that should remain aligned with `docs/agents/*` and should not absorb constitution detail.
7. `docs/agents/README.md`, `docs/agents/WORKFLOWS.md`, `docs/agents/GUARDRAILS.md`, `docs/agents/RLM.md`, `docs/agents/TOOLING.md` - runtime guidance for future agents and source-of-truth ordering.
8. `docs/references/README.md`, `docs/references/testing.md`, `docs/references/tooling.md`, `docs/references/external-systems.md` - durable reference locations that can later hold detailed stable guidance once implementation exists.
9. `docs/references/rules/safety-guardrails.md`, `docs/references/rules/work-lane-gating.md`, `docs/references/rules/github-pr-delivery.md`, `docs/references/rules/kit-capabilities-usage.md` - process rules that should remain pointer-loaded rather than copied into the constitution.
10. `.gitignore` - current evidence of Go-oriented and Kit-local artifact boundaries.
11. `.envrc` - current evidence of direnv/dotenv local environment loading.
12. `.coderabbit.yaml` - current evidence that automated review intentionally excludes docs and agent routing files.
13. `.github/pull_request_template.md` - current PR structure used by the delivery rules when PR workflow is requested.
14. `docs/notes/0001-init-project/.gitkeep` - only file in the feature notes directory; no durable note content exists to promote.

## DEPENDENCIES

1. Runtime dependencies: none declared. No language module file, lockfile, package manifest, or source import graph exists yet.
2. Intended language/toolchain signal: Go is suggested by `.gitignore` and the repository path, but not yet confirmed by a `go.mod` or source files.
3. Workflow dependency: Kit is the dominant project workflow dependency. It provides `kit map`, staged feature artifacts under `docs/specs`, managed rulesets, progress summary, agent routing docs, and the no-out-of-order workflow.
4. Local environment dependency: direnv is implied by `.envrc`; dotenv values are optional and private.
5. Review dependency: CodeRabbit is configured, but documentation and agent routing files are excluded from review.
6. GitHub dependency: GitHub PR delivery is documented but opt-in; no issue, branch, commit, push, or PR work is part of this brainstorm phase.
7. External systems: none documented. `docs/references/external-systems.md` is a placeholder.

## QUESTIONS

1. Should the constitution treat Mint as a Go CLI/release automation tool before `go.mod` exists?
   Recommended default: yes, but phrase it as an intended direction supported by `README.md` and `.gitignore`, not as implemented architecture.
2. Should the constitution contain detailed GitHub delivery, work-lane, and Kit command rules?
   Recommended default: no. It should state the durable invariants and point to `docs/references/rules/*` and `docs/agents/*` for procedural detail.
3. Should the constitution define specific release algorithms now, such as SemVer bump detection, changelog grouping, tagging, and publishing?
   Recommended default: no. It should name these as product goals and require future specs to define concrete behavior before implementation.
4. Should `PROJECT_PROGRESS_SUMMARY.md` be edited during this brainstorm update?
   Recommended default: no. It already reports `0001-init-project` at `brainstorm`, which is the highest completed artifact.
5. Should docs-only research create an issue, branch, commit, or PR?
   Recommended default: no. `work-lane-gating` explicitly excludes Kit pipeline phases and standalone documentation edits from agent-initiated PR workflow.
6. Current understanding: 96 percent. The remaining uncertainty is product intent depth, not research-phase scope; the next `SPEC.md` can carry the recommended defaults above without blocking.

## OPTIONS

1. Recommended: write a compact, durable constitution that defines principles, non-negotiable constraints, workflow boundaries, product intent, implementation posture, and non-goals, while linking to `docs/agents/*` and `docs/references/rules/*` for detailed process.
   Tradeoff: avoids a monolithic constitution and matches the repo's routing-table pattern, but requires future agents to follow references for operational detail.
2. Alternative: write a comprehensive constitution that copies detailed Kit, git, GitHub, testing, CLI, release, and dependency rules into one document.
   Tradeoff: easier to read in one file, but conflicts with the repo's explicit guidance to avoid always-loaded monolithic instruction files and risks stale duplicated rules.
3. Alternative: defer constitution content until actual source code exists.
   Tradeoff: avoids over-specifying unimplemented architecture, but leaves the canonical project contract mostly empty while implementation decisions begin.
4. Alternative: immediately update `docs/references/testing.md`, `docs/references/tooling.md`, and `docs/references/external-systems.md` with guessed project policies.
   Tradeoff: could make the docs feel more complete, but would invent durable guidance without implementation evidence.

## RECOMMENDED STRATEGY

1. Treat `docs/CONSTITUTION.md` as a high-level contract, not a full manual.
2. Preserve the Kit-managed baseline section exactly unless Kit regenerates it.
3. Replace placeholder sections with concrete, durable rules:
   - principles for correctness, clarity, minimalism, and document-first traceability;
   - constraints around source-of-truth order, no out-of-order feature progression, docs-first behavior changes, no committed secrets, no generated local state, and no invented implementation behavior;
   - current product intent for Mint as a release/version/changelog tool, clearly marked as not yet implemented;
   - non-goals that prevent early scope creep, such as building package-manager-specific release behavior, hosted release services, or broad CI orchestration before specs define them;
   - definitions for Mint, constitution, Kit-managed project, feature artifact, release, version, changelog, and local generated state.
4. Keep detailed operational procedures in existing references:
   - `docs/agents/*` for agent workflow and source-of-truth routing;
   - `docs/references/rules/*` for git, GitHub, work-lane, safety, and Kit command-discovery rules;
   - `docs/references/testing.md`, `docs/references/tooling.md`, and `docs/references/external-systems.md` only after stable project-specific facts exist.
5. In the follow-up spec, require the constitution update to avoid claiming implemented runtime architecture, dependencies, tests, external integrations, or release algorithms that the repository does not yet contain.
6. Keep `docs/PROJECT_PROGRESS_SUMMARY.md` unchanged during brainstorm because the highest artifact remains `BRAINSTORM.md`; update it only when Kit creates or advances later artifacts.

## NEXT STEP

Run `kit spec init-project` to turn this research into a concrete specification for updating `docs/CONSTITUTION.md`.
