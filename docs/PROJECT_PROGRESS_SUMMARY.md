# PROJECT PROGRESS SUMMARY

## FEATURE PROGRESS TABLE

| ID | FEATURE | PATH | PHASE | PAUSED | CREATED | SUMMARY |
| -- | ------- | ---- | ----- | ------ | ------- | ------- |
| 0001 | init-project | `docs/specs/0001-init-project` | complete | no | 2026-06-19 | Defines the project constitution for Mint as a Kit-managed, documentation-first release tooling project before runtime code exists. The constitution must lock durable development rules, product intent, constraints, non-goals, and terms without inventing implementation details or duplicating pointer-loaded workflow rules. |
| 0002 | cli-patterns | `docs/specs/0002-cli-patterns` | complete | no | 2026-06-19 | Adopt Kit-style CLI, README, Makefile, and build patterns for Mint without implementing release algorithms yet. |
| 0003 | github-action | `docs/specs/0003-github-action` | complete | no | 2026-06-19 | Expose the Mint CLI through a public GitHub composite action without adding release behavior. |
| 0004 | changelog-generation | `docs/specs/0004-changelog-generation` | complete | no | 2026-06-19 | Generate CHANGELOG.md release entries from conventional commits between Git refs. |
| 0005 | pull-forward-v1-features | `docs/specs/0005-pull-forward-v1-features` | complete | no | 2026-06-19 | Adds release resolution and GHCR/ECR publish workflow generation to Mint using the proven Git-tag-first container release pattern; ECS deployment and GitHub Release creation stay out of scope. |
| 0006 | github-release-publishing | `docs/specs/0006-github-release-publishing` | complete | no | 2026-06-19 | Add GitHub Release publishing to Mint and configure Mint to publish Mint's own GitHub Releases through the Mint action. |
| 0007 | release-state-ownership | `docs/specs/0007-release-state-ownership` | complete | no | 2026-06-19 | Make Mint own release-state operations through CLI/action commands while application repositories keep Docker publishing and deployment workflows. |
| 0008 | release-tag-selection | `docs/specs/0008-release-tag-selection` | complete | no | 2026-07-02 | Add a read-only release tag selector so deploy workflows can recover an already-published SemVer image tag without copied shell. |

## PROJECT INTENT

Mint is a document-first release tooling CLI and GitHub Action. It turns repository-local Git history and explicit workflow inputs into deterministic release metadata, changelog entries, and publish workflow artifacts while keeping implementation decisions traceable through Kit-managed markdown artifacts.

## GLOBAL CONSTRAINTS

See `docs/CONSTITUTION.md` for project-wide constraints and principles.

## FEATURE SUMMARIES

### init-project

- **STATUS**: complete
- **PAUSED**: no
- **INTENT**: Mint currently has a short `README.md` product signal and Kit scaffold documentation, but `docs/CONSTITUTION.md` is mostly placeholder text. Without a completed constitution, future agents and humans lack a canonical project contract for source-of-truth order, workflow progression, scope boundaries, implementation posture, and product direction.
- **APPROACH**: 1. Treat `SPEC.md` as binding and `BRAINSTORM.md` as supporting context only. 2. Use discovery-first routing for the implementation plan and downstream execution metadata: `parallelization_mode: "rlm"`. 3. Edit `docs/CONSTITUTION.md` as the primary artifact by replacing placeholder prose, not by expanding top-level agent routing files. 4. Preserve the Kit-managed baseline block byte-for-byte unless Kit regenerates it. 5. Structure the constitution as a high-level contract: - principles for correctness, clarity, minimalism, evidence, and document-first traceability; - constraints for source-of-truth order, ordered Kit progression, progress summary maintenance, secret/local state safety, generated artifact handling, and fact-vs-intent wording; - existing change classification retained and made concrete enough to guide future work; - non-goals that block premature release-algorithm, CI, hosted-service, package-manager, and external-integration scope; - definitions for the terms future agents need before adding product code. 6. Point detailed procedures to `docs/agents/*` and `docs/references/rules/*` instead of copying those rules into the constitution. 7. Update `docs/PROJECT_PROGRESS_SUMMARY.md` only as supporting documentation to keep the feature phase, summary, approach, and open items aligned with the completed plan. 8. Stop before product implementation; the next phase is task generation.
- **OPEN ITEMS**: none
- **POINTERS**: `docs/specs/0001-init-project/BRAINSTORM.md`, `docs/specs/0001-init-project/SPEC.md`, `docs/specs/0001-init-project/PLAN.md`, `docs/specs/0001-init-project/TASKS.md`

### cli-patterns

- **STATUS**: complete
- **PAUSED**: no
- **INTENT**: Mint now has a completed constitution and Kit scaffold, but no source code, Go module, Makefile, README structure, or build/test entrypoints. Future release-tooling work needs the same lightweight operator ergonomics Kit uses before release behavior is specified.
- **APPROACH**: 1. Add the Go module and dependencies needed for the same CLI style as Kit: Cobra plus terminal-aware styling. 2. Create `cmd/mint/main.go` as a thin binary entrypoint that delegates to `pkg/cli.Execute()`. 3. Create a small `pkg/cli` surface: - `root.go` for root command, banner, version variable, and error handling. - `root_help.go` for sectioned Kit-style help rendering. - `human_output.go` for terminal-aware labels and help templates. - `version.go` for `mint version` and build-info fallback. 4. Add focused tests for version resolution and help rendering. 5. Add Kit-style Makefile and pre-commit hook. 6. Rewrite README in Kit's style while deferring release behavior to later feature work. 7. Update constitution and repo-wide references now that Go CLI/build tooling exists.
- **OPEN ITEMS**: none
- **POINTERS**: `docs/specs/0002-cli-patterns/SPEC.md`, `docs/specs/0002-cli-patterns/PLAN.md`, `docs/specs/0002-cli-patterns/TASKS.md`

### github-action

- **STATUS**: complete
- **PAUSED**: no
- **INTENT**: The repository has a CLI scaffold but no GitHub Actions entrypoint. Consumers should be able to reference `jamesonstone/mint` from a workflow and run the same `mint` binary without copying build commands into every workflow.
- **APPROACH**: 1. Add `action.yml` as a composite action because the current integration is a reusable sequence of runner steps around a CLI. 2. Use `actions/setup-go` with a configurable `go-version` input so the action can build the Go module on hosted runners. 3. Build `./cmd/mint` into `$RUNNER_TEMP/mint-action/mint`, inject the action ref as `pkg/cli.Version` when the ref is safe, and append the binary directory to `GITHUB_PATH`. 4. Run only a fixed allowlist of current commands after building: `version`, `help`, or `none`. 5. Capture command stdout in the `output` action output and expose the built path as `mint-path`. 6. Update README and durable repo docs so the action is documented as a CLI wrapper, not release automation.
- **OPEN ITEMS**: none
- **POINTERS**: `docs/specs/0003-github-action/SPEC.md`, `docs/specs/0003-github-action/PLAN.md`, `docs/specs/0003-github-action/TASKS.md`

### changelog-generation

- **STATUS**: complete
- **PAUSED**: no
- **INTENT**: Generate CHANGELOG.md release entries from conventional commits between Git refs.
- **APPROACH**: 1. Add a small `pkg/changelog` package for testable generation behavior: git collection, conventional commit parsing, issue extraction, grouping, rendering, and atomic file prepending. 2. Add `mint changelog` with the required flags, and route root-level `mint --prev-tag ... --current-tag ... --owner ... --repo ... --output ...` to the same implementation to satisfy the requested invocation. 3. Keep raw Git work behind `git` command invocations and make tag/range errors explicit. 4. Add tests using temporary Git repositories with fixture commits and tags. 5. Update help, README, durable docs, and feature progress now that changelog generation exists.
- **OPEN ITEMS**: none
- **POINTERS**: `docs/specs/0004-changelog-generation/SPEC.md`, `docs/specs/0004-changelog-generation/PLAN.md`, `docs/specs/0004-changelog-generation/TASKS.md`

### pull-forward-v1-features

- **STATUS**: complete
- **PAUSED**: no
- **INTENT**: Mint now extends its CLI scaffold, changelog generation, and composite GitHub Action wrapper with release computation, Git tag creation rules for generated workflows, and container publish workflow generation. The release-publish pattern already exists in `r2`, Flowcore, and event-sink as copied shell/YAML logic, which leaves maintainers repeating bespoke scripts for SemVer resolution, tag-first publishing, and image tagging.
- **APPROACH**: 1. Added `pkg/release` contracts and deterministic Git test helpers. 2. Implemented Git/SemVer discovery, commit loading, bump classification, and release-note generation before exposing adapters. 3. Proved resolver behavior with temporary Git repositories. 4. Added image parsing, registry validation, and deterministic GHCR/ECR workflow rendering with negative checks for deploy-scope behavior. 5. Added `mint release` CLI commands, then extended the composite action as a safe CLI adapter. 6. Updated README, constitution facts, and project summary after implementation evidence exists. 7. Completed full repository validation and readiness checks.
- **OPEN ITEMS**: none
- **POINTERS**: `docs/specs/0005-pull-forward-v1-features/BRAINSTORM.md`, `docs/specs/0005-pull-forward-v1-features/SPEC.md`, `docs/specs/0005-pull-forward-v1-features/PLAN.md`, `docs/specs/0005-pull-forward-v1-features/TASKS.md`

### github-release-publishing

- **STATUS**: complete
- **PAUSED**: no
- **INTENT**: Mint can resolve the next strict SemVer tag and expose that result through its GitHub Action, but it cannot publish a GitHub Release artifact. The Mint repository also needs to use Mint itself to publish Mint's own GitHub Releases without container-image publishing.
- **APPROACH**: 1. Add a small standard-library GitHub REST API client in `pkg/release` for get-by-tag and create-release behavior. 2. Treat existing releases by tag as idempotent success. 3. Add `mint release github` as a thin Cobra adapter that reads tokens from environment variables. 4. Extend `action.yml` with a fixed `github-release` allowlist command and typed outputs. 5. Add `.github/workflows/release.yaml` so the repository uses the local Mint action to resolve and publish its GitHub Release. 6. Update README, constitution, tooling, testing, and progress docs after implementation evidence exists. 7. Complete repository validation and readiness checks.
- **OPEN ITEMS**: none
- **POINTERS**: `docs/specs/0006-github-release-publishing/SPEC.md`, `docs/specs/0006-github-release-publishing/PLAN.md`, `docs/specs/0006-github-release-publishing/TASKS.md`

### release-state-ownership

- **STATUS**: complete
- **PAUSED**: no
- **INTENT**: Mint should own release-state operations directly: version resolution, release-note/changelog support, immutable Git tag creation, and GitHub Release creation. Application repositories should keep Docker image building, registry authentication, registry publishing, deployment, and infrastructure workflows.
- **APPROACH**: 1. Add first-class release tag package behavior with strict SemVer validation, target validation, same-commit reuse, conflict failure, annotated tags, and optional push. 2. Add release publish composition over resolver, tag creation, and GitHub Release creation. 3. Add `mint release tag` and `mint release publish` CLI adapters. 4. Extend the action allowlist and outputs. 5. Update the self-release workflow and generated Docker workflow renderer to delegate release-state behavior to Mint. 6. Update README, constitution, tooling, testing, and feature docs. 7. Complete final validation and readiness checks.
- **OPEN ITEMS**: none
- **POINTERS**: `docs/specs/0007-release-state-ownership/SPEC.md`, `docs/specs/0007-release-state-ownership/PLAN.md`, `docs/specs/0007-release-state-ownership/TASKS.md`

### release-tag-selection

- **STATUS**: complete
- **PAUSED**: no
- **INTENT**: Downstream deploy workflows need a small Mint helper that recovers the strict SemVer image tag already created by release publishing, or validates a manually requested tag, without copying Git tag lookup shell into each repository.
- **APPROACH**: 1. Add read-only package behavior that resolves a target commit, validates optional requested tags, and selects the highest strict SemVer tag pointing at the target. 2. Add `mint release select-tag` as a thin CLI adapter. 3. Extend the composite action with `release-select-tag`, `requested-tag`, and typed outputs. 4. Update README, constitution, tooling references, agent manifests, and feature docs. 5. Validate package, CLI, and action behavior with focused tests.
- **OPEN ITEMS**: none
- **POINTERS**: `docs/specs/0008-release-tag-selection/SPEC.md`, `docs/specs/0008-release-tag-selection/PLAN.md`, `docs/specs/0008-release-tag-selection/TASKS.md`

## LAST UPDATED

2026-07-02 17:25:34 EDT
