# CONSTITUTION

## PRINCIPLES

### Correctness Before Motion

- Prefer accurate, evidence-backed changes over quick changes.
- Do not guess repository behavior, file contents, command behavior, dependencies, or implemented architecture.
- When current repository evidence conflicts with prior notes or intent, treat current evidence as stronger and update the docs before implementation.

### Minimal, Durable Changes

- Make the smallest production-ready change that satisfies the active feature artifact.
- Prefer explicit, idiomatic implementation over cleverness or premature generalization.
- Add abstractions only when they remove real duplication, clarify ownership, or match an established local pattern.

### Document-First Traceability

- Use Kit feature artifacts to move from research to specification, planning, tasks, implementation, reflection, and completion.
- Keep durable decisions in canonical markdown artifacts, not only in chat.
- Make implementation work trace back to `SPEC.md`, `PLAN.md`, and `TASKS.md` before editing product files.

### Fact Versus Intent

- Clearly separate implemented facts from intended direction.
- Mint's current product intent is to compute the next version, write the changelog, and mint the release.
- The current repository contains a minimal Go/Cobra CLI scaffold, Kit-style README, Makefile build targets, version reporting, conventional-commit CHANGELOG.md generation, and a GitHub Action wrapper that builds and exposes the CLI in workflows.
- Mint does not yet contain a release algorithm, tagging behavior, publishing behavior, external integrations, or package-manager-specific release behavior.

## CONSTRAINTS

### Source Of Truth

- Authority order is: safety and permission constraints, current user request, `docs/CONSTITUTION.md`, `SPEC.md`, `PLAN.md`, `TASKS.md`, `BRAINSTORM.md`, then repo conventions.
- `docs/CONSTITUTION.md` is the canonical project contract.
- `docs/specs/<feature>/` is the source of truth for feature-scoped requirements, plans, and tasks.
- `BRAINSTORM.md` is research context, not binding implementation authority.

### Kit Workflow

- Spec-driven work must proceed through Kit artifacts in order: optional `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, `TASKS.md`, implementation, reflection, and completion.
- Do not move out of order unless the user explicitly overrides the project workflow.
- `docs/PROJECT_PROGRESS_SUMMARY.md` must reflect the highest completed artifact or active phase for each feature whenever feature docs advance.
- Use RLM-style just-in-time context loading for broad or noisy repository analysis.

### Documentation Boundaries

- Keep `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` as short routing tables.
- Put durable workflow detail in `docs/agents/*` and durable rulesets in `docs/references/rules/*`.
- Do not turn top-level agent entrypoints into always-loaded monolithic manuals.
- Do not duplicate detailed git, GitHub, lane-gating, PR delivery, or Kit command-discovery procedures in this constitution; link to the pointer-loaded rule files instead.

### Implementation Evidence

- Do not claim that Mint has implemented release behavior, CI, external integrations, package-manager behavior, or tests beyond the artifacts that exist in the repository.
- Future product behavior, release algorithms, external integrations, and test strategy must be defined in feature specs before implementation.
- The Go module, `cmd/mint`, `pkg/cli`, `pkg/changelog`, Makefile, and `action.yml` are implementation evidence for the current CLI scaffold, changelog generation, and GitHub Action wrapper only.

### Local And Generated State

- Never commit secrets, credentials, `.env` values, private keys, tokens, or machine-local configuration.
- Treat `.env`, `.envrc`-loaded values, `.kit/runs/`, `.kit/loops/`, `.kit/state.json`, `.kit/cache/`, `.kit/tmp/`, `.kit/temp/`, `.kit/*.tmp`, and `.kit/*.lock` as local or generated state unless a future spec says otherwise.
- Keep generated state, caches, and scratch artifacts out of durable project docs except as documented boundaries.

### Kit-Managed Baseline Rules

<!-- BEGIN KIT-MANAGED BASELINE RULES -->
- Treat `docs/CONSTITUTION.md` as the canonical project contract.
- Keep `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` aligned with the repo-local docs tree.
- Prefer implementation/source code files around 300 lines or less when splitting improves clarity and ownership.
- Do not apply the code-file size guideline to documentation files, all `docs/**`, all `.kit/**`, or `.kit.yaml`.
- Do not split or rewrite docs, generated state, or Kit config artifacts solely because they exceed 300 lines.
<!-- END KIT-MANAGED BASELINE RULES -->

## CHANGE CLASSIFICATION

All work must be classified before editing.

### Spec-Driven (Formal)

- Use for new features, substantial architectural or behavioral changes, cross-component changes, and Kit pipeline phases.
- Workflow: optional `BRAINSTORM.md` -> `SPEC.md` -> `PLAN.md` -> `TASKS.md` -> implement -> reflect.
- Ask clarifying questions until confidence is high and unresolved assumptions are closed before implementation.

### Ad Hoc (Lightweight)

- Use for contained bug fixes, security reviews, refactors, dependency updates, config changes, and small refinements.
- Workflow: understand -> implement -> verify.
- Update only practical docs affected by the change.
- Do not create `SPEC.md`, `PLAN.md`, or `TASKS.md` for ad hoc work unless the scope grows into spec-driven work.

### Ad Hoc with Existing Specs

- If a change touches behavior covered by existing feature docs, default to updating those docs.
- Skip spec updates only for purely mechanical changes such as formatting, typo fixes, or dependency bumps that do not alter behavior.

## NON-GOALS

- Do not define concrete release algorithms before a feature spec covers them.
- Do not add package-manager-specific release behavior before the supported package ecosystems are specified.
- Do not build a hosted release service, web application, or external orchestration platform without a dedicated spec.
- Do not introduce broad CI orchestration, publishing automation, or GitHub release behavior before the release contract is specified.
- Do not invent external-system integrations when `docs/references/external-systems.md` has no durable project-specific integration guidance.
- Do not treat starter reference files as complete project policy until stable project-specific facts are added.
- Do not use the constitution to replace detailed workflow, delivery, or command-discovery rules that belong in pointer-loaded references.

## DEFINITIONS

- **Mint**: The project in this repository. It is a Go CLI whose current implemented surface is root help, version reporting, conventional-commit CHANGELOG.md generation, and a GitHub Action wrapper that builds and exposes the CLI in workflows; its product intent is to compute the next version, write the changelog, and mint the release.
- **Constitution**: `docs/CONSTITUTION.md`, the canonical project contract and highest repo-local project rule after safety constraints and the current user request.
- **Kit-managed project**: A repository using Kit artifacts such as `.kit.yaml`, `docs/CONSTITUTION.md`, `docs/agents/*`, and `docs/specs/<feature>/`.
- **Feature artifact**: A canonical markdown document under `docs/specs/<feature>/`, such as `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, or `TASKS.md`.
- **Release**: A future Mint operation that prepares a version, changelog, and release output. The exact release behavior is not implemented yet.
- **Version**: A future release identifier computed by Mint. Versioning rules must be specified before implementation.
- **Changelog**: A human-readable release note artifact produced or updated by Mint from conventional commits and Git refs.
- **Local generated state**: Machine-local files, caches, dotenv inputs, Kit runtime state, and scratch artifacts that should not be committed as durable project state.
- **Implementation evidence**: Repository artifacts such as source files, manifests, tests, commands, docs, or config that prove behavior exists.
