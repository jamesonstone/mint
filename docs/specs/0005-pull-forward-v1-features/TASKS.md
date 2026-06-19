---
kit_metadata_version: 1
artifact: tasks
feature:
  id: 0005
  slug: pull-forward-v1-features
  dir: 0005-pull-forward-v1-features
source_artifacts:
  - docs/specs/0005-pull-forward-v1-features/BRAINSTORM.md
  - docs/specs/0005-pull-forward-v1-features/SPEC.md
  - docs/specs/0005-pull-forward-v1-features/PLAN.md
---
# TASKS

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Define release package contracts and test fixtures | done | agent | |
| T002 | Implement Git and SemVer discovery | done | agent | T001 |
| T003 | Implement release bump evaluation and notes | done | agent | T002 |
| T004 | Cover release resolution with temp Git tests | done | agent | T003 |
| T005 | Implement image spec and registry validation | done | agent | T004 |
| T006 | Implement publish workflow rendering | done | agent | T005 |
| T007 | Cover workflow rendering and registry behavior | done | agent | T006 |
| T008 | Add release CLI commands | done | agent | T007 |
| T009 | Extend composite action release-resolve support | done | agent | T008 |
| T010 | Update release documentation and project facts | done | agent | T009 |
| T011 | Run final validation and readiness checks | done | agent | T010 |

## TASK LIST

- [x] T001: Define release package contracts and test fixtures.
- [x] T002: Implement Git and SemVer discovery.
- [x] T003: Implement release bump evaluation and notes.
- [x] T004: Cover release resolution with temp Git tests.
- [x] T005: Implement image spec and registry validation.
- [x] T006: Implement publish workflow rendering.
- [x] T007: Cover workflow rendering and registry behavior.
- [x] T008: Add release CLI commands.
- [x] T009: Extend composite action release-resolve support.
- [x] T010: Update release documentation and project facts.
- [x] T011: Run final validation and readiness checks.

## TASK DETAILS

### T001

- **GOAL**: Establish the release package API and fixture helpers used by the resolver and workflow renderer.
- **SCOPE**:
  - Add `pkg/release` types for release resolution options, release results, bump kind, commit metadata, image specs, registry kind, and workflow options.
  - Add package-level error values or typed validation helpers for fail-closed user errors.
  - Add test helpers for temporary Git repositories, dated commits, tags, and command execution.
  - Keep the package behavior-free except for simple constructors or validation primitives needed by later tasks.
- **PLAN LINKS**: [COMPONENTS](PLAN.md#components), [DATA](PLAN.md#data), [TESTING](PLAN.md#testing)
- **ACCEPTANCE**:
  - Release result fields cover `version_tag`, `version_bump`, `base_tag`, `target_sha`, `short_sha`, `needs_git_tag`, `commit_count`, and `release_notes`.
  - Workflow option fields can represent repeatable image specs with `name`, `uri`, `dockerfile`, and `context`.
  - Test helpers can create deterministic temporary Git repositories without relying on global Git user configuration.
  - No CLI, action, README, or workflow generation behavior is added in this task.
- **VERIFY**:
  - `go test ./pkg/release`
- **EXPECTED FILES**:
  - `pkg/release/types.go`
  - `pkg/release/errors.go`
  - `pkg/release/test_helpers_test.go`
- **RISK**: Low; this is an additive package foundation with limited behavior.
- **ROLLBACK**: Remove the new `pkg/release` files if later package design changes substantially.
- **NOTES**: Keep names aligned with `action.yml` output names from the specification.

### T002

- **GOAL**: Implement Git ref resolution, strict SemVer tag discovery, and commit range loading.
- **SCOPE**:
  - Resolve `--commitish` to a commit SHA and fail if it is missing or not a commit.
  - Discover SemVer tags matching exactly `^v[0-9]+\.[0-9]+\.[0-9]+$`.
  - Detect SemVer tags already pointing at the target commit and choose the highest version on that commit.
  - Find the highest reachable SemVer tag from the target commit while ignoring tags from unrelated branches.
  - Load commits oldest-first from the selected base through the target, excluding the base tag commit.
- **PLAN LINKS**: [APPROACH](PLAN.md#approach), [COMPONENTS](PLAN.md#components), [DATA](PLAN.md#data)
- **ACCEPTANCE**:
  - Missing or non-commit refs fail closed with a clear error.
  - Nonmatching tags, pre-release tags, and build-metadata tags are ignored.
  - If no reachable SemVer tag exists, the base tag value is the implicit `v0.0.0`.
  - Commit loading returns author dates and full commit messages needed by bump and notes logic.
  - Already-tagged targets return enough data for T003 to produce `needs_git_tag=false`.
- **VERIFY**:
  - `go test ./pkg/release`
- **EXPECTED FILES**:
  - `pkg/release/git.go`
  - `pkg/release/semver.go`
  - `pkg/release/git_test.go`
- **RISK**: Medium; Git graph behavior is central to release correctness.
- **ROLLBACK**: Revert `pkg/release/git.go`, `pkg/release/semver.go`, and related tests.
- **NOTES**: Prefer explicit `git` invocations, following the existing `pkg/changelog` package style.

### T003

- **GOAL**: Convert loaded commits into deterministic release version, bump, and release-note results.
- **SCOPE**:
  - Parse conventional commit subjects for `feat`, `fix`, and breaking `!` markers.
  - Detect breaking changes from body or footer lines containing `BREAKING CHANGE:` or `BREAKING-CHANGE:`.
  - Rank bumps as breaking major, `feat` minor, and `fix`, other conventional, or nonconventional patch.
  - Apply first-release rules from the specification for `v0.1.0`, `v0.0.1`, and `v1.0.0`.
  - Generate lightweight annotation release notes from the commits included in the resolved range.
  - Return `short_sha` as exactly 12 characters and preserve the full target SHA for consumers.
- **PLAN LINKS**: [APPROACH](PLAN.md#approach), [DATA](PLAN.md#data), [INTERFACES](PLAN.md#interfaces)
- **ACCEPTANCE**:
  - Already-tagged targets resolve to the highest SemVer tag on the target, `version_bump=already-tagged`, `needs_git_tag=false`, and `commit_count=0`.
  - Untagged targets set `needs_git_tag=true` and compute the next strict `vX.Y.Z` tag.
  - Breaking commits dominate feature, fix, other conventional, and nonconventional commits.
  - Commit evaluation is deterministic and oldest-first.
  - The resolver performs no tag creation, tag push, changelog writes, image publishing, or workflow writes.
- **VERIFY**:
  - `go test ./pkg/release`
- **EXPECTED FILES**:
  - `pkg/release/resolve.go`
  - `pkg/release/commits.go`
  - `pkg/release/notes.go`
  - `pkg/release/resolve_test.go`
- **RISK**: Medium; this task defines the release algorithm.
- **ROLLBACK**: Revert resolver files and tests, keeping T001/T002 contracts if still useful.
- **NOTES**: Treat release notes as tag annotation text only; do not modify `CHANGELOG.md`.

### T004

- **GOAL**: Prove the release resolver against fixture Git histories that match the specification.
- **SCOPE**:
  - Add temp-repository tests for patch, minor, major, first-release, already-tagged, and unrelated-branch-tag cases.
  - Add tests for nonconventional commits, strict tag filtering, invalid commitish handling, and empty ranges.
  - Add tests for `short_sha`, commit count, release notes, base tag, and target SHA output fields.
  - Ensure tests are deterministic on local machines and GitHub runners.
- **PLAN LINKS**: [TESTING](PLAN.md#testing), [RISKS](PLAN.md#risks)
- **ACCEPTANCE**:
  - Acceptance criteria AC-03 through AC-13 are directly covered by tests or explicit assertions.
  - The resolver fails closed for invalid target refs and non-commit refs.
  - A target already tagged with multiple SemVer tags chooses the highest tag on that commit.
  - Tags from unrelated branches do not influence the base tag.
  - Tests do not depend on network access, repository global config, or wall-clock dates.
- **VERIFY**:
  - `go test ./pkg/release -run 'Resolve|SemVer|Git'`
  - `go test ./...`
- **EXPECTED FILES**:
  - `pkg/release/resolve_test.go`
  - `pkg/release/git_test.go`
  - `pkg/release/test_helpers_test.go`
- **RISK**: Medium; tests may expose gaps in earlier resolver assumptions.
- **ROLLBACK**: Revert test additions, then reassess resolver behavior before implementation continues.
- **NOTES**: Keep fixture histories small and named by behavior.

### T005

- **GOAL**: Validate image specifications and registry compatibility before workflow rendering.
- **SCOPE**:
  - Parse repeatable image arguments in the form `name=<name>,uri=<image-uri>,dockerfile=<path>,context=<path>`.
  - Require `name`, `uri`, and `dockerfile`; default `context` to `.`.
  - Reject duplicate names, tagged image URIs, unsupported registries, and mixed registry kinds.
  - Detect GHCR from `ghcr.io/...` image URIs.
  - Detect ECR from standard AWS ECR hostnames.
- **PLAN LINKS**: [APPROACH](PLAN.md#approach), [DATA](PLAN.md#data), [RISKS](PLAN.md#risks)
- **ACCEPTANCE**:
  - Invalid image specs fail before workflow YAML is rendered.
  - Image URI validation rejects values with tags, digests, missing path segments, or unsupported hosts.
  - GHCR and ECR registry detection returns a single registry kind for all images.
  - Error messages identify the invalid image spec field without exposing secrets.
- **VERIFY**:
  - `go test ./pkg/release -run 'Image|Registry'`
  - `go test ./pkg/release`
- **EXPECTED FILES**:
  - `pkg/release/image.go`
  - `pkg/release/image_test.go`
- **RISK**: Medium; registry validation controls whether generated workflows are safe to publish.
- **ROLLBACK**: Revert `image.go` and `image_test.go`.
- **NOTES**: Keep parsing strict; do not infer registry credentials or account values beyond the image URI host.

### T006

- **GOAL**: Render deterministic GitHub Actions publish workflows for validated GHCR and ECR image sets.
- **SCOPE**:
  - Build a typed workflow model before rendering YAML text.
  - Generate an `on: push` workflow limited to the repository default branch.
  - Add `concurrency` group `release-publish` with `cancel-in-progress: false`.
  - Add full checkout, tag refresh, release resolution, tag creation or same-commit tag guard, and direct `docker buildx build --push` steps.
  - Generate GHCR auth using `${{ github.actor }}` and `${{ secrets.GITHUB_TOKEN }}`.
  - Generate ECR auth using `aws-actions/configure-aws-credentials`, `aws-actions/amazon-ecr-login`, and `AWS_PUBLISH_ROLE_TO_ASSUME`.
  - Exclude ECS deploy jobs, `workflow_dispatch` deploy environments, service/container names, task-definition mutation, and GitHub Release creation.
- **PLAN LINKS**: [APPROACH](PLAN.md#approach), [COMPONENTS](PLAN.md#components), [INTERFACES](PLAN.md#interfaces)
- **ACCEPTANCE**:
  - Workflow YAML contains release resolver output names from T001/T003.
  - Tag creation happens before any image build or push when `needs_git_tag=true`.
  - Existing same-commit tags allow publish to continue; tags pointing to another commit fail.
  - Every configured image emits exactly two publish tags: resolved SemVer tag and `latest`.
  - GHCR workflow permissions allow content tag writes and package writes.
  - ECR workflow contains AWS role assumption and ECR login but no GHCR login.
  - Generated YAML is stable across repeated runs with the same inputs.
- **VERIFY**:
  - `go test ./pkg/release -run 'Workflow|Render'`
  - `go test ./pkg/release`
- **EXPECTED FILES**:
  - `pkg/release/workflow.go`
  - `pkg/release/workflow_test.go`
- **RISK**: High; generated workflows perform repository tag and image-publish operations when used by consumers.
- **ROLLBACK**: Revert workflow rendering files and keep resolver functionality isolated.
- **NOTES**: Do not use arbitrary workflow shell templates when typed data can make validation explicit.

### T007

- **GOAL**: Prove generated workflows match the source patterns and exclude deploy behavior.
- **SCOPE**:
  - Add YAML structure tests for branch filters, permissions, concurrency, checkout depth, tag fetch, resolver invocation, tag guards, and buildx publish commands.
  - Add GHCR-specific tests for auth and package permissions.
  - Add ECR-specific tests for AWS credential and ECR login steps.
  - Add negative assertions for ECS deploy strings, task-definition mutation, GitHub Release creation, `workflow_dispatch` deploy blocks, and unsupported mixed registries.
- **PLAN LINKS**: [TESTING](PLAN.md#testing), [DEPENDENCIES](PLAN.md#dependencies), [RISKS](PLAN.md#risks)
- **ACCEPTANCE**:
  - Acceptance criteria AC-17 through AC-27 are covered by tests or fixture assertions.
  - Generated YAML parses with the repository's chosen YAML parser or validation approach.
  - Test fixtures make accidental deployment-scope additions visible in diffs.
  - Renderer tests fail when duplicate image names, tagged URIs, unsupported registries, or mixed registries are accepted.
- **VERIFY**:
  - `go test ./pkg/release -run 'Workflow|Registry|Image'`
  - `go test ./...`
- **EXPECTED FILES**:
  - `pkg/release/workflow_test.go`
  - `pkg/release/testdata/workflows/ghcr.yml`
  - `pkg/release/testdata/workflows/ecr.yml`
- **RISK**: Medium; fixture maintenance can hide behavior if assertions are too shallow.
- **ROLLBACK**: Revert workflow tests and fixtures, then preserve renderer code only if covered elsewhere.
- **NOTES**: If fixtures are added, keep them intentionally small and deterministic.

### T008

- **GOAL**: Expose release resolution and workflow generation through idiomatic Mint CLI commands.
- **SCOPE**:
  - Add `mint release` as a command group in `pkg/cli`.
  - Add `mint release resolve --commitish HEAD --github-output <path>`.
  - Add `mint release workflow --image ... --output <path> --mint-ref <ref>`.
  - Register the release command from the root command without changing existing `mint changelog` behavior.
  - Print the resolved `version_tag` for `release resolve` and write all required fields to a GitHub output file when requested.
  - Write workflow YAML to stdout or `--output`, matching existing CLI output boundaries.
- **PLAN LINKS**: [APPROACH](PLAN.md#approach), [COMPONENTS](PLAN.md#components), [INTERFACES](PLAN.md#interfaces)
- **ACCEPTANCE**:
  - `mint help` lists `release` and preserves existing root, version, and changelog help behavior.
  - `mint release --help`, `mint release resolve --help`, and `mint release workflow --help` describe flags and examples.
  - `mint release resolve` performs no Git mutation and no file writes except the optional GitHub output file.
  - `mint release workflow` writes only the requested output file or stdout.
  - Acceptance criteria AC-01, AC-02, AC-14, and AC-15 are covered by tests or CLI assertions.
- **VERIFY**:
  - `go test ./pkg/cli`
  - `go test ./...`
  - `go run ./cmd/mint release --help`
  - `go run ./cmd/mint release resolve --help`
  - `go run ./cmd/mint release workflow --help`
- **EXPECTED FILES**:
  - `pkg/cli/release.go`
  - `pkg/cli/root.go`
  - `pkg/cli/release_test.go`
- **RISK**: Medium; CLI flag and output contracts become user-facing.
- **ROLLBACK**: Revert CLI release files and root registration, leaving `pkg/release` intact for rework.
- **NOTES**: Follow `pkg/cli/changelog.go` conventions for structs, `RunE`, and error handling.

### T009

- **GOAL**: Extend the public composite action so workflows can call Mint release resolution safely.
- **SCOPE**:
  - Add `release-resolve` to the action command allowlist.
  - Add action inputs needed by release resolution, including commitish.
  - Invoke the built Mint binary with `mint release resolve` and a temporary GitHub output file.
  - Map resolver output fields into first-class action outputs.
  - Preserve existing `version`, `help`, and `none` behavior.
  - Avoid shell passthrough, workflow generation, tag creation, or image publishing inside the composite action.
- **PLAN LINKS**: [APPROACH](PLAN.md#approach), [COMPONENTS](PLAN.md#components), [INTERFACES](PLAN.md#interfaces)
- **ACCEPTANCE**:
  - `action.yml` exposes outputs for `version_tag`, `version_bump`, `base_tag`, `target_sha`, `short_sha`, `needs_git_tag`, `commit_count`, and `release_notes`.
  - The action still builds the local CLI before running allowed commands.
  - Unsupported action commands fail closed.
  - Acceptance criteria AC-14, AC-15, and AC-16 are covered by tests or YAML assertions.
- **VERIFY**:
  - `go test ./...`
  - `grep -n "release-resolve" action.yml`
- **EXPECTED FILES**:
  - `action.yml`
  - `pkg/cli/release_test.go`
  - `pkg/release/action_test.go`
- **RISK**: Medium; action output names are consumed by external workflows.
- **ROLLBACK**: Revert `action.yml` and action-specific tests while keeping the CLI command available.
- **NOTES**: Treat the action as an adapter around the CLI, not a parallel implementation.

### T010

- **GOAL**: Update documentation and durable project facts after the release behavior exists.
- **SCOPE**:
  - Update `README.md` with release command quick starts, action usage, GHCR workflow generation, ECR workflow generation, outputs, and explicit non-goals.
  - Update `docs/CONSTITUTION.md` to replace future-scoped release statements with current implementation facts where implementation evidence exists.
  - Update `docs/PROJECT_PROGRESS_SUMMARY.md` to reflect implementation status and pointers.
  - Keep changelog generation documented as separate from release resolution.
- **PLAN LINKS**: [APPROACH](PLAN.md#approach), [COMPONENTS](PLAN.md#components), [RISKS](PLAN.md#risks)
- **ACCEPTANCE**:
  - README includes a quick start for `mint release resolve`.
  - README includes a quick start for using `jamesonstone/mint` as a GitHub Action in a workflow.
  - README includes examples for generating GHCR and ECR publish workflows.
  - Documentation states that ECS deployment and GitHub Release creation are out of scope.
  - Acceptance criteria AC-28 and AC-29 are covered.
- **VERIFY**:
  - `rg -n "release resolve|release workflow|release-resolve|GHCR|ECR|ECS|GitHub Release" README.md docs/CONSTITUTION.md docs/PROJECT_PROGRESS_SUMMARY.md`
- **EXPECTED FILES**:
  - `README.md`
  - `docs/CONSTITUTION.md`
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
- **RISK**: Low; this task updates docs after code behavior is verifiable.
- **ROLLBACK**: Revert documentation edits if implementation behavior changes.
- **NOTES**: Avoid claiming generated workflows deploy services; they publish tags and images only.

### T011

- **GOAL**: Run the full repository validation set and leave the feature ready for review.
- **SCOPE**:
  - Run repository tests and build checks.
  - Run formatting and static validation commands already used by the project.
  - Run markdown and YAML sanity checks where tooling is available.
  - Confirm generated workflow fixtures contain no ECS deploy behavior.
  - Confirm project docs and Kit phase metadata are consistent.
  - Inspect `git diff` for unrelated churn, secrets, local paths, and generated artifacts that should not be committed.
- **PLAN LINKS**: [TESTING](PLAN.md#testing), [RISKS](PLAN.md#risks)
- **ACCEPTANCE**:
  - Acceptance criteria AC-30 through AC-35 are satisfied or explicitly documented if a local tool is unavailable.
  - `go test ./...` passes.
  - `go vet ./...` passes.
  - `make build` passes.
  - `gofmt` produces no pending changes.
  - `git diff --check` passes.
  - `kit map 0005-pull-forward-v1-features` shows artifacts aligned with the current phase.
- **VERIFY**:
  - `gofmt -w cmd pkg`
  - `go test ./...`
  - `go vet ./...`
  - `make build`
  - `git diff --check`
  - `kit map 0005-pull-forward-v1-features`
- **EXPECTED FILES**:
  - `docs/PROJECT_PROGRESS_SUMMARY.md`
- **RISK**: Medium; final validation may reveal earlier implementation gaps.
- **ROLLBACK**: Revert only the failing task's owned files and rerun this validation task.
- **NOTES**: Do not stage, commit, push, or open a pull request unless that is requested separately after validation.

## DEPENDENCIES

Tasks are intentionally ordered so implementation can proceed without guessing at later contracts:

1. T001 creates the package contracts and deterministic test helpers.
2. T002 and T003 implement release resolution behavior before any adapter uses it.
3. T004 proves the resolver against Git fixture histories before workflow work begins.
4. T005, T006, and T007 add and prove workflow generation after release outputs are stable.
5. T008 exposes the package through the CLI.
6. T009 adapts the CLI through the public composite action.
7. T010 updates documentation only after behavior is implemented.
8. T011 runs repository-wide validation and updates final project state.

## NOTES

- `SPEC.md` and `PLAN.md` are fixed inputs for this task list.
- Keep `mint changelog` separate from `mint release`; release notes here are lightweight Git tag annotation text, not `CHANGELOG.md` generation.
- Keep the resolver pure. Tag creation, tag push, and image publish behavior belong only in generated workflows.
- Keep ECS deployment, GitHub Release creation, service names, container names, and task-definition mutation out of scope.
- Reflection review fixed release-note handling in generated workflows so tag annotation text is passed through an environment variable instead of a fixed heredoc delimiter.
- Reflection review reduced unnecessary public error surface in `pkg/release` and added public API comments for agent readability.
- Project refresh advisory: no project refresh needed.

<!-- REFLECTION_COMPLETE -->
