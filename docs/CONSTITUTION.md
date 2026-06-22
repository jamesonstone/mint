# CONSTITUTION

## PURPOSE

This document is the canonical project contract for Mint. It defines the
development rules, architecture, implementation boundaries, process, and
long-term direction that future work must preserve unless a later user request
and feature specification explicitly change them.

Mint is a Go CLI and GitHub Action for release tooling. Its product intent is
to compute the next version, write the changelog, and mint the release while
keeping release behavior explicit, testable, and traceable through Kit-managed
documents.

## PRINCIPLES

### Correctness Before Motion

- Prefer accurate, evidence-backed changes over quick changes.
- Do not guess repository behavior, file contents, command behavior,
  dependencies, implemented architecture, or workflow semantics.
- When current repository evidence conflicts with prior notes or intent, treat
  current evidence as stronger and update canonical docs before implementation.
- Fail closed for invalid release, changelog, Git, workflow, or file state.

### Minimal, Durable Changes

- Make the smallest production-ready change that satisfies the active request
  and feature artifact.
- Prefer explicit, idiomatic Go over cleverness or premature generalization.
- Add abstractions only when they remove real duplication, clarify ownership,
  or match an established local pattern.
- Keep public API surface small. If a symbol is only needed inside one package,
  keep it unexported.

### CLI-First Product Shape

- The `mint` CLI is the core product surface.
- GitHub Action behavior wraps the CLI and must not become an independent
  implementation of Mint behavior.
- Workflow-generation features may emit shell or YAML, but the validation and
  decision logic must remain in Go where it can be tested.
- User-facing commands should remain script-friendly: deterministic stdout,
  explicit flags, clear errors, and no hidden interactive prompts unless a
  future spec requires them.

### Document-First Traceability

- Durable decisions belong in canonical markdown artifacts, not only in chat.
- Feature implementation must trace back to `SPEC.md`, `PLAN.md`, and
  `TASKS.md` before product files change.
- Documentation must be updated when implementation reality changes behavior,
  command surface, release semantics, dependencies, or operator expectations.
- Keep implemented facts separate from future intent.

## CURRENT IMPLEMENTED SURFACE

The current repository implements:

- A Go module at `github.com/jamesonstone/mint`.
- A thin binary entrypoint at `cmd/mint/main.go` that delegates to
  `pkg/cli.Execute()`.
- A Cobra CLI under `pkg/cli` with:
  - root help and Kit-style command grouping;
  - `mint version` and `mint --version`;
  - `mint changelog`;
  - root-level changelog flags for script compatibility;
  - `mint release resolve`;
  - `mint release tag`;
  - `mint release github`;
  - `mint release publish`;
  - `mint release workflow`.
- A `pkg/changelog` package that generates `CHANGELOG.md` release blocks from
  conventional commits and Git refs.
- A `pkg/release` package that resolves release metadata, creates immutable Git
  tags, publishes GitHub Releases, and renders GHCR/ECR publish workflows.
- A root `action.yml` composite GitHub Action that builds `cmd/mint`, adds the
  binary to `PATH`, and runs an allowlisted Mint command.
- A repository release workflow that uses the Mint action itself to resolve and
  publish Mint GitHub Releases.
- A Kit-style `Makefile`, Go tests, and repository-managed pre-commit build
  hook.
- Kit-managed docs under `docs/agents`, `docs/specs`, `docs/references`, and
  this constitution.

Mint currently does not build Docker images directly, authenticate to registries
directly, deploy services, upload release assets, publish package-manager
artifacts, support registries beyond GHCR/ECR, or make the CLI resolver push
tags or images directly.

## ARCHITECTURE

### Package Ownership

- `cmd/mint` contains only the executable entrypoint and should remain a thin
  delegation layer.
- `pkg/cli` owns Cobra command registration, flag binding, command help,
  stdout/stderr behavior, version output, and adapter code from CLI flags to
  domain packages.
- `pkg/changelog` owns changelog generation, including Git ref validation,
  commit collection, conventional commit parsing, grouping, rendering, and
  atomic file writes.
- `pkg/release` owns release metadata resolution, SemVer tag handling, bump
  classification, GitHub output writing, Git tag creation/reuse, GitHub Release
  publishing, image spec validation, and publish workflow rendering.
- `action.yml` is the public GitHub Action metadata. It builds and invokes the
  CLI through fixed command branches.
- `docs/specs/<feature>` owns feature-scoped requirements, plans, tasks, and
  reflection records.
- `docs/references` owns durable cross-feature references and rulesets.

### Adapter And Domain Boundary

- CLI command functions should be thin adapters. They parse flags, call domain
  packages, and print user-facing output.
- Domain packages should not depend on Cobra.
- Domain packages should expose focused `Options` and `Result` structs for
  command-level behavior.
- Git-backed behavior should accept `context.Context` and optional `WorkDir`
  values so tests can run in deterministic temporary repositories.
- Domain logic should return data to callers instead of printing directly,
  except where warning writers are explicitly part of the contract.

### Git And File Boundaries

- Git operations should be explicit subprocess calls with fixed argument lists.
- Validate refs before using them in release or changelog calculations.
- Parse Git output using stable delimiters or structured formats, not ad hoc
  human-oriented text.
- Changelog file writes must be atomic: read, validate, write a temporary file,
  preserve permissions, and rename.
- Generated container workflows may push images when users run them, but Git
  tag creation must be delegated to Mint release-state commands rather than
  copied shell.
- The local CLI resolver must remain read-only.
- GitHub Release publishing may call the GitHub API to create a Release, but
  local Git tag creation and pushing belongs to `mint release tag` or
  `mint release publish`.

### GitHub Action Boundary

- The action must build the Go CLI from source on the runner.
- The action must use an allowlist for `command`; it must not run arbitrary
  shell, `eval`, or user-provided command strings.
- The action should expose typed outputs for command behavior that workflows
  need to consume.
- The action should keep `mint-path` and captured `output` stable for general
  usage.
- The action may expose release-state behavior only by invoking the Mint CLI,
  not by duplicating Git tag or release API logic in shell.

## CODE STYLE AND NAMING

- Use idiomatic Go with explicit error returns.
- Keep package names short and domain-specific: `cli`, `changelog`, `release`.
- Use `Options` for input structs and `Result` for returned command/domain
  outcomes.
- Prefer unexported helper types and functions unless a CLI adapter, test, or
  external package boundary needs the symbol.
- Public exported symbols must have useful comments.
- Prefer package-level compiled regular expressions for stable parsers.
- Use `strings.Builder` for deterministic multi-line rendering.
- Keep validation errors specific to the invalid field or state.
- Use table-driven tests for matrix behavior such as SemVer bumps, image
  validation, invalid refs, and generated workflow variants.
- Test Git behavior with temporary repositories that configure local test
  authors and deterministic dates.
- Keep source files near 300 lines when splitting improves clarity; this does
  not apply to docs or Kit generated/local state.

## DEPENDENCIES

- `github.com/spf13/cobra`: CLI command tree, flags, help, and version command
  behavior.
- `github.com/spf13/pflag`: Cobra's flag implementation, indirect.
- `github.com/inconshreveable/mousetrap`: Cobra's Windows console support,
  indirect.
- `golang.org/x/term`: terminal detection for human-friendly help styling.
- `golang.org/x/sys`: terminal/platform support through `x/term`, indirect.
- `gopkg.in/yaml.v3`: YAML parsing in tests for `action.yml` and generated
  workflow validation.

Do not add dependencies for convenience. Each new dependency must have a clear
runtime or test purpose, fit the CLI-first architecture, and be recorded in the
relevant feature docs and durable references.

## RELEASE AND CHANGELOG CONTRACTS

### Changelog Generation

- `current_tag`, `repo_owner`, and `repo_name` are required.
- `prev_tag` may be empty for a first release.
- `current_ref` may be provided when `current_tag` is a newly resolved version
  that does not exist yet; in that case `current_ref` is the commit range end
  and `current_tag` is still the rendered release tag.
- `output_file` defaults to `CHANGELOG.md`.
- Tags must resolve to commits when used as range refs. `current_ref` must
  resolve to a commit when provided.
- Commit parsing follows the conventional commit subject contract implemented
  by `pkg/changelog`.
- Non-conventional commits are warnings and are skipped.
- `docs`, `test`, and `chore` commits are excluded from rendered release
  entries.
- Rendered groups are ordered as breaking changes, features, fixes, perf, then
  other.
- Issue links come from `closes`, `fixes`, or `resolves` body text before a
  subject `(#123)` suffix.
- Existing changelog release headers must parse. Unparseable changelogs fail.
- Duplicate versions fail and must not overwrite existing content.

### Release Resolution

- Release tags are strict `vX.Y.Z` SemVer Git tags.
- Pre-release tags, build-metadata tags, and non-`v` tags are ignored by the
  resolver.
- The target commit defaults to `HEAD` and must resolve to a commit.
- If the target commit already has strict SemVer tags, choose the highest tag,
  set `version_bump=already-tagged`, `needs_git_tag=false`, and
  `commit_count=0`.
- Otherwise, choose the highest reachable strict SemVer base tag.
- Breaking commits produce a major bump. `feat` commits produce a minor bump.
  `fix`, other conventional, and non-conventional commits produce a patch bump.
- First release defaults are `v1.0.0` for breaking changes, `v0.1.0` for
  features, and `v0.0.1` otherwise.
- The resolver returns `version_tag`, `version_bump`, `base_tag`,
  `target_sha`, `short_sha`, `needs_git_tag`, `commit_count`, and
  `release_notes`.

### Git Tag Creation

- `mint release tag` creates or reuses annotated strict `vX.Y.Z` Git tags.
- Required inputs are release tag, target commitish, and release notes file.
- The target commitish must resolve to a commit before any tag mutation.
- If the tag already exists on the same target commit, the command succeeds and
  reports tag reuse.
- If the tag already exists on another commit, the command fails and must never
  move the tag.
- Conflicting tag errors must include the recovery path: inspect and correct
  the conflicting tag, or push a dummy commit after correction to trigger a
  clean release calculation.
- A newly created tag is annotated from the provided notes file.
- Tag pushing defaults to enabled for CLI/action CI usage and can be disabled
  with `--push=false`.

### GitHub Release Publishing

- `mint release github` creates or reuses a GitHub Release for a strict
  `vX.Y.Z` SemVer tag.
- Required inputs are repository owner, repository name, release tag, target
  commitish, and a GitHub token.
- The token must come from an environment variable, not from a required command
  argument. Default lookup order is `MINT_GITHUB_TOKEN`, `GITHUB_TOKEN`, then
  `GH_TOKEN`, unless `--token-env` points at another variable.
- The command uses the GitHub REST API directly and must not require the GitHub
  CLI to be installed.
- Existing releases for the same tag are treated as idempotent success.
- The command writes `release_tag`, `release_url`, and `release_created` when
  asked to append GitHub Actions output.
- `mint release github` must not create local tags, push tags, push images,
  upload release assets, or deploy services.
- `mint release publish` composes release-state operations: resolve, write
  release notes to a temporary file, create or reuse the Git tag, push the tag
  when enabled, and create or reuse the GitHub Release.
- `mint release publish` must not build Docker images, authenticate to
  registries, push containers, upload release assets, or deploy services.
- The self-release workflow for this repository must use the Mint action to
  publish release state, with `contents: write` permissions and without
  container-image publishing.

### Publish Workflow Generation

- Image specs use `name=<name>,uri=<image-uri>,dockerfile=<path>,context=<path>`.
- `name`, `uri`, and `dockerfile` are required; `context` defaults to `.`.
- Image URIs must be repository URIs without tags or digests.
- Supported registries are GHCR (`ghcr.io`) and AWS ECR
  (`<account>.dkr.ecr.<region>.amazonaws.com`).
- A generated workflow must not mix registry kinds.
- Generated workflows run on `push` and guard the publish job to the repository
  default branch.
- Generated workflows must fetch full history and tags, resolve the release
  through the Mint action, create or validate the Git tag through the Mint
  action before publishing, set up Docker Buildx, authenticate to the selected
  registry, and publish every image with the resolved SemVer tag and `latest`.
- Generated workflows must not include ECS deployment, task-definition
  mutation, `workflow_dispatch` deployment gates, GitHub Release publishing, or
  package-manager publishing.

## DEVELOPMENT PROCESS

### Source Of Truth

Authority order:

1. safety and permission constraints
2. current user request
3. `docs/CONSTITUTION.md`
4. `SPEC.md`
5. `PLAN.md`
6. `TASKS.md`
7. `BRAINSTORM.md`
8. repo conventions

`docs/CONSTITUTION.md` is the canonical project contract.
`docs/specs/<feature>` is the source of truth for feature-scoped requirements,
plans, tasks, and reflection. `BRAINSTORM.md` is research context, not binding
implementation authority.

### Work Classification

- Classify all work before editing.
- Use spec-driven work for new features, substantial behavioral changes,
  cross-component changes, and Kit pipeline phases.
- Use ad hoc work for contained bug fixes, reviews, dependency updates, config
  changes, and small refinements.
- Do not create feature docs for ad hoc work unless the scope grows into a new
  feature.
- If a change touches behavior covered by existing feature docs, update those
  docs unless the change is purely mechanical and behavior-neutral.

### Kit Workflow

- Spec-driven work proceeds through Kit artifacts in order: optional
  `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, `TASKS.md`, implementation,
  reflection, and completion.
- Do not move out of order unless the user explicitly overrides the workflow.
- `docs/PROJECT_PROGRESS_SUMMARY.md` must reflect the highest completed
  artifact or active phase for each feature at all times.
- Use RLM-style just-in-time context loading for broad or noisy analysis.
- Use `kit dispatch` or subagents only after discovery narrows the work into
  distinct low-overlap lanes. Keep the main agent responsible for synthesis,
  integration, validation, and communication.

### Documentation Boundaries

- Keep `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` as short
  routing tables.
- Put durable workflow detail in `docs/agents/*`.
- Put durable rulesets in `docs/references/rules/*`.
- Put durable cross-feature technical references in `docs/references/*`.
- Do not turn top-level instruction files into always-loaded manuals.
- Do not copy full PR delivery, branch, issue, or Kit command-discovery
  procedures into this constitution; load the pointer rules when those
  decisions are active.

### Validation

- Run the smallest relevant verification for ad hoc docs work.
- Run package tests for changed Go domain packages.
- Run `go test ./...`, `go vet ./...`, `make build`, and `git diff --check`
  for CLI, action, release, changelog, or build-surface changes.
- Parse `action.yml` or generated workflow YAML when action/workflow metadata
  changes.
- Never claim validation passed unless it actually ran.
- If validation cannot run, state why and report the residual risk.

### Git And Delivery

- Do not stage, commit, push, create issues, or mutate PRs without explicit
  user approval.
- In Kit-managed projects, GitHub delivery must follow repo-local delivery
  rules under `docs/agents/GUARDRAILS.md` and
  `docs/references/rules/github-pr-delivery.md`.
- Branch, issue, commit, push, and PR defaults from global tools are not
  authoritative in this repository.
- Never commit secrets, `.env` values, private keys, local tokens, or
  machine-local config.

## LOCAL AND GENERATED STATE

- Treat `.env`, `.envrc`-loaded values, `.kit/runs/`, `.kit/loops/`,
  `.kit/state.json`, `.kit/cache/`, `.kit/tmp/`, `.kit/temp/`, `.kit/*.tmp`,
  `.kit/*.lock`, and `bin/` as local or generated state unless a future spec
  says otherwise.
- Do not cite local generated state as durable project truth except when
  reporting validation artifact locations.
- Keep generated state, caches, and scratch artifacts out of durable docs.

### Kit-Managed Baseline Rules

<!-- BEGIN KIT-MANAGED BASELINE RULES -->
- Treat `docs/CONSTITUTION.md` as the canonical project contract.
- Keep `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` aligned with the repo-local docs tree.
- Prefer implementation/source code files around 300 lines or less when splitting improves clarity and ownership.
- Do not apply the code-file size guideline to documentation files, all `docs/**`, all `.kit/**`, or `.kit.yaml`.
- Do not split or rewrite docs, generated state, or Kit config artifacts solely because they exceed 300 lines.
<!-- END KIT-MANAGED BASELINE RULES -->
## NON-NEGOTIABLE CONSTRAINTS

- Do not claim implemented behavior that is not backed by repository evidence.
- Do not add package-manager-specific release behavior before the supported
  package ecosystems are specified.
- Do not add GitHub Release asset upload, package publication, or deployment
  side effects before a dedicated spec defines ownership, content,
  idempotency, and failure behavior.
- Do not add ECS/service deployment or environment-specific deployment behavior
  to release workflow generation before a dedicated spec covers it.
- Do not add unsupported registry publishing before a dedicated spec defines the
  registry contract.
- Do not build a hosted release service, web application, or external
  orchestration platform without a dedicated spec.
- Do not invent external-system integrations when
  `docs/references/external-systems.md` has no durable project-specific
  integration guidance.
- Do not allow the GitHub Action to execute arbitrary command strings.
- Do not reimplement CLI domain logic in `action.yml` shell.
- Do not use this constitution to replace pointer-loaded workflow and delivery
  rules.

## LONG-TERM VISION

Mint should grow into a focused release tooling layer that turns repository
history and explicit release configuration into repeatable release artifacts.
The durable direction is:

- Keep the CLI as the source of product behavior and the GitHub Action as a
  transport wrapper.
- Keep release decisions deterministic, inspectable, and script-friendly.
- Prefer typed Go domain packages over shell scripts for parsing, validation,
  and rendering.
- Expand release capabilities incrementally through specs: richer release
  plans, changelog integration, release assets, additional registries,
  package-manager publishing, deployment handoff, and policy checks should each
  have explicit contracts before implementation.
- Make generated workflows boring and auditable: tag first, publish with
  immutable version tags plus `latest` only where specified, and fail on
  conflicting remote state.
- Preserve a clear boundary between release artifact generation and
  environment-specific deployment.
- Keep docs, tests, and command output aligned so future maintainers can trust
  the repository without reconstructing decisions from chat history.

## DEFINITIONS

- **Mint**: The project in this repository. It is a Go CLI and GitHub Action
  whose current implemented surface is root help, version reporting,
  conventional-commit `CHANGELOG.md` generation, release tag resolution,
  GitHub Release publishing, GHCR/ECR publish workflow generation, and
  action-based CLI execution.
- **Constitution**: `docs/CONSTITUTION.md`, the canonical project contract and
  highest repo-local project rule after safety constraints and the current user
  request.
- **Kit-managed project**: A repository using Kit artifacts such as
  `.kit.yaml`, `docs/CONSTITUTION.md`, `docs/agents/*`, and
  `docs/specs/<feature>`.
- **Feature artifact**: A canonical markdown document under
  `docs/specs/<feature>`, such as `BRAINSTORM.md`, `SPEC.md`, `PLAN.md`, or
  `TASKS.md`.
- **Release**: A Mint operation that prepares version metadata, changelog
  output, or release workflow output while keeping deploy and package
  publishing scope explicit.
- **Version**: A strict `vX.Y.Z` SemVer Git tag computed by Mint release
  resolution from reachable Git history.
- **Changelog**: A human-readable release note artifact produced or updated by
  Mint from conventional commits and Git refs.
- **Generated workflow**: YAML rendered by Mint for users to commit into their
  own repositories when they want tag-first GHCR or ECR publishing.
- **Local generated state**: Machine-local files, caches, dotenv inputs, Kit
  runtime state, build output, and scratch artifacts that should not be
  committed as durable project state.
- **Implementation evidence**: Repository artifacts such as source files,
  manifests, tests, commands, docs, or config that prove behavior exists.


## CONSTRAINTS

These are runtime and behavioral invariants Mint must always guarantee. They
complement the scope-limiting rules in `## NON-NEGOTIABLE CONSTRAINTS`.

- Release resolution and the local CLI resolver are read-only; they never
  create, move, push, or delete Git tags or container images.
- Mint never moves an existing tag. A strict SemVer tag that already exists on a
  different commit is a hard failure with a stated recovery path, never a silent
  retag.
- Release-state operations are idempotent: re-running `mint release tag`,
  `mint release github`, or `mint release publish` for the same tag and target
  commit succeeds without duplicating Git tags or GitHub Releases.
- Invalid release, changelog, Git ref, workflow, or file state fails closed.
  Mint never emits partial or ambiguous release artifacts.
- Refs are validated to resolve to commits before any range calculation or tag
  mutation.
- Changelog writes are atomic and never overwrite an existing version block;
  duplicate versions fail.
- The GitHub Action runs only allowlisted commands and never executes arbitrary
  shell, `eval`, or user-supplied command strings.
- GitHub tokens are read from environment variables, never required as command
  arguments, and never logged.
- Generated workflows are tag-first, never mix registry kinds, and never deploy
  services.
- Secrets, `.env` values, tokens, and machine-local state are never committed or
  cited as durable project truth.

### Kit-Managed Baseline Rules

<!-- BEGIN KIT-MANAGED BASELINE RULES -->
- Treat `docs/CONSTITUTION.md` as the canonical project contract.
- Keep `AGENTS.md`, `CLAUDE.md`, and `.github/copilot-instructions.md` aligned with the repo-local docs tree.
- Prefer implementation/source code files around 300 lines or less when splitting improves clarity and ownership.
- Do not apply the code-file size guideline to documentation files, all `docs/**`, all `.kit/**`, or `.kit.yaml`.
- Do not split or rewrite docs, generated state, or Kit config artifacts solely because they exceed 300 lines.
<!-- END KIT-MANAGED BASELINE RULES -->

## CHANGE CLASSIFICATION

All work falls into one of three tracks. Classify before acting. This mirrors
the `DEVELOPMENT PROCESS → Work Classification` section and makes each track
concrete.

### Spec-Driven (Formal)

- Use for new features, substantial architectural or behavioral changes,
  cross-component changes, and Kit pipeline phases.
- Workflow: optional `BRAINSTORM.md`, then `SPEC.md`, `PLAN.md`, `TASKS.md`,
  implementation, reflection, and completion, in order.
- Create feature artifacts under `docs/specs/<feature>/` and keep
  `docs/PROJECT_PROGRESS_SUMMARY.md` aligned with the highest completed
  artifact or active phase.

### Ad Hoc (Lightweight)

- Use for contained bug fixes, security reviews, refactors, dependency updates,
  config changes, and small refinements.
- Workflow: understand, implement, verify.
- Update only practical docs such as README, inline docs, and API docs.
- Do not create `SPEC.md`, `PLAN.md`, or `TASKS.md` for ad hoc work.

### Ad Hoc with Existing Specs

- If the change touches behavior covered by existing feature docs, default to
  updating those docs.
- Skip spec updates only for purely mechanical, behavior-neutral changes such as
  formatting, typo fixes, or dependency bumps.

## NON-GOALS

Mint deliberately excludes the following until a dedicated spec defines the
contract, ownership, idempotency, and failure behavior:

- Mint does not build Docker images, authenticate to registries, or push
  containers itself; generated workflows perform those steps in the user's
  repository.
- Mint does not upload GitHub Release assets or publish package-manager
  artifacts (npm, Homebrew, apt, PyPI, Go proxy, and similar).
- Mint does not deploy services, mutate ECS task definitions, or perform
  environment-specific deployment.
- Mint does not support container registries beyond GHCR and AWS ECR.
- Mint is not a hosted release service, web application, or external
  orchestration platform.
- The GitHub Action is not an independent reimplementation of Mint behavior and
  never executes arbitrary command strings.
- Mint does not invent external-system integrations absent durable guidance in
  `docs/references/external-systems.md`.
- The local CLI resolver does not push tags or images.
