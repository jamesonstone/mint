---
kit_metadata_version: 1
artifact: plan
feature:
  id: 0005
  slug: pull-forward-v1-features
  dir: 0005-pull-forward-v1-features
summary: Implementation plan for adding release resolution and GHCR/ECR publish workflow generation to Mint.
parallelization_mode: rlm
skills:
  - name: rlm
    source: repo-local guide
    path: docs/agents/RLM.md
    trigger: "analyze codebase; scan all files; large repository analysis; scan repository; recursive language model; broad release workflow discovery"
    required: true
relationships:
  - type: depends_on
    target: 0002-cli-patterns
    reason: Release commands must preserve the established Go/Cobra CLI and build pattern.
  - type: depends_on
    target: 0003-github-action
    reason: Action support extends the existing composite action allowlist and output model.
  - type: builds_on
    target: 0004-changelog-generation
    reason: Release resolution reuses the repository's Git-backed package and temporary Git test approach while keeping changelog rendering separate.
references:
  - id: spec
    name: Pull-forward v1 features specification
    type: feature_doc
    target: docs/specs/0005-pull-forward-v1-features/SPEC.md
    relation: constrains
    read_policy: must
    used_for: binding scope, command names, output fields, edge cases, and acceptance criteria
    status: active
  - id: brainstorm
    name: Pull-forward brainstorm
    type: feature_doc
    target: docs/specs/0005-pull-forward-v1-features/BRAINSTORM.md
    relation: informs
    read_policy: conditional
    used_for: source workflow research, tradeoffs, and rejected alternatives
    status: active
  - id: constitution
    name: Mint constitution
    type: repo_doc
    target: docs/CONSTITUTION.md
    relation: constrains
    read_policy: must
    used_for: document-first workflow, current implementation facts, and non-goal boundaries
    status: active
  - id: rlm-guide
    name: RLM guide
    type: repo_doc
    target: docs/agents/RLM.md
    relation: guides
    read_policy: must
    used_for: broad-context implementation sequencing and downstream planning mode
    status: active
  - id: progress-summary
    name: Project progress summary
    type: repo_doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    relation: updates
    read_policy: must
    used_for: feature phase and plan completion status
    status: active
  - id: cli-patterns-plan
    name: CLI patterns plan
    type: feature_doc
    target: docs/specs/0002-cli-patterns/PLAN.md
    relation: informs
    read_policy: conditional
    used_for: existing CLI/build implementation sequencing
    status: active
  - id: github-action-plan
    name: GitHub action plan
    type: feature_doc
    target: docs/specs/0003-github-action/PLAN.md
    relation: informs
    read_policy: conditional
    used_for: existing composite action implementation sequencing
    status: active
  - id: changelog-plan
    name: Changelog generation plan
    type: feature_doc
    target: docs/specs/0004-changelog-generation/PLAN.md
    relation: informs
    read_policy: conditional
    used_for: Git-backed package and temporary repository test sequencing
    status: active
  - id: cli-root
    name: Mint root command
    type: source
    target: pkg/cli/root.go
    selector_type: symbol
    selector: rootCmd
    relation: updates
    read_policy: evidence
    used_for: root wording, command registration, and release command group integration
    status: active
  - id: cli-changelog
    name: Mint changelog command
    type: source
    target: pkg/cli/changelog.go
    selector_type: symbol
    selector: changelogCmd
    relation: informs
    read_policy: evidence
    used_for: Cobra command, flag binding, and stdout/stderr handling pattern
    status: active
  - id: changelog-package
    name: Changelog package
    type: source
    target: pkg/changelog
    relation: informs
    read_policy: evidence
    used_for: Git command helper, parser, renderer, file boundary, and temp Git test pattern
    status: active
  - id: action-yml
    name: Mint composite action
    type: action
    target: action.yml
    relation: updates
    read_policy: evidence
    used_for: command allowlist, build step, environment wiring, and action output model
    status: active
  - id: readme
    name: Mint README
    type: repo_doc
    target: README.md
    relation: updates
    read_policy: evidence
    used_for: release CLI/action documentation and future-scoped wording cleanup
    status: active
  - id: r2-release-script
    name: r2 release resolver
    type: source
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/r2/scripts/resolve-release-version.sh
    relation: informs
    read_policy: evidence
    used_for: source behavior for release resolution, output fields, patch fallback, and release notes
    status: active
  - id: r2-main-workflow
    name: r2 publish workflow
    type: workflow
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/r2/.github/workflows/main.yaml
    relation: informs
    read_policy: evidence
    used_for: single-image ECR workflow parity
    status: active
  - id: flowcore-main-workflow
    name: Flowcore publish workflow
    type: workflow
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/flowcore/.github/workflows/main.yaml
    relation: informs
    read_policy: evidence
    used_for: multi-image ECR workflow parity
    status: active
  - id: event-sink-main-workflow
    name: event-sink publish workflow
    type: workflow
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/event-sink/.github/workflows/main.yaml
    relation: informs
    read_policy: evidence
    used_for: single-image ECR workflow confirmation
    status: active
  - id: flowcore-ecs-deploy
    name: Flowcore ECS deploy workflow
    type: workflow
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/flowcore/.github/workflows/deploy-ecs.yaml
    relation: constrains
    read_policy: evidence
    used_for: negative scope and no-ECS acceptance checks
    status: active
  - id: conventional-commits
    name: Conventional Commits 1.0.0
    type: url
    target: https://www.conventionalcommits.org/en/v1.0.0/
    relation: guides
    read_policy: evidence
    used_for: commit parsing and bump ranking semantics
    status: active
  - id: semver
    name: Semantic Versioning 2.0.0
    type: url
    target: https://semver.org/
    relation: guides
    read_policy: evidence
    used_for: version increment semantics and immutable-release rationale
    status: active
  - id: github-workflow-syntax
    name: GitHub Actions workflow syntax
    type: url
    target: https://docs.github.com/actions/using-workflows/workflow-syntax-for-github-actions
    relation: guides
    read_policy: evidence
    used_for: generated workflow permissions, jobs, steps, and outputs
    status: active
  - id: github-concurrency
    name: GitHub Actions concurrency
    type: url
    target: https://docs.github.com/actions/writing-workflows/choosing-what-your-workflow-does/control-the-concurrency-of-workflows-and-jobs
    relation: guides
    read_policy: evidence
    used_for: generated release-publish concurrency block
    status: active
  - id: github-container-registry
    name: GitHub Container Registry
    type: url
    target: https://docs.github.com/packages/working-with-a-github-packages-registry/working-with-the-container-registry
    relation: guides
    read_policy: evidence
    used_for: GHCR auth and package permission behavior
    status: active
  - id: docker-login-action
    name: Docker login action
    type: url
    target: https://github.com/docker/login-action
    relation: guides
    read_policy: evidence
    used_for: GHCR login step shape
    status: active
  - id: aws-configure-credentials
    name: AWS configure credentials action
    type: url
    target: https://github.com/aws-actions/configure-aws-credentials
    relation: guides
    read_policy: evidence
    used_for: ECR OIDC credential step shape
    status: active
  - id: aws-ecr-login
    name: AWS Amazon ECR login action
    type: url
    target: https://github.com/aws-actions/amazon-ecr-login
    relation: guides
    read_policy: evidence
    used_for: ECR Docker login step shape
    status: active
---
# PLAN

## SUMMARY

Implement the feature as a small release domain package with thin CLI and action adapters, plus deterministic workflow rendering and fixture-backed tests. Keep mutation boundaries explicit: `mint release resolve` only computes release metadata, while generated workflows own Git tag creation and image publishing after validating tags, registry kind, and image specs.

## APPROACH

1. Build the release domain in `pkg/release` before changing CLI/action/docs.
   - Keep Git commands behind package-level helpers, matching the `pkg/changelog` pattern.
   - Keep SemVer parsing strict and local to `vX.Y.Z`; do not add a SemVer dependency unless implementation proves it materially reduces complexity.
   - Separate release resolution from workflow generation so the resolver remains testable without YAML or registry concerns.
2. Add CLI commands after the package contracts are stable.
   - Add a `mint release` command group in `pkg/cli/release.go`.
   - Keep `mint release resolve` as a pure command that prints the resolved tag and optionally writes GitHub output fields.
   - Keep `mint release workflow` as a renderer that writes to stdout or `--output`.
   - Follow `pkg/cli/changelog.go` for flag structs, `RunE`, stdout/stderr boundaries, and command registration.
3. Extend `action.yml` as a safe adapter, not a second implementation.
   - Preserve the current build step and command allowlist pattern.
   - Add `release-resolve` to the allowlist and map release outputs from a GitHub output file into first-class action outputs.
   - Do not add shell passthrough or generated workflow publishing behavior inside the composite action.
4. Generate workflow YAML from typed workflow data.
   - Model registry kind, image specs, permissions, and publish steps before rendering text.
   - Use direct `docker buildx build --push` shell steps for source-workflow parity.
   - Keep GHCR and ECR generation as two renderer branches selected by validated image URI host.
   - Fail before rendering if image specs are incomplete, tagged, duplicate, unsupported, or mixed-registry.
5. Update documentation after behavior is implemented.
   - Rewrite README release/action sections to document new commands and still-explicit non-goals.
   - Update `docs/CONSTITUTION.md` only after implementation evidence exists.
   - Update `docs/PROJECT_PROGRESS_SUMMARY.md` as supporting project state.
6. Preserve RLM for downstream implementation planning.
   - Initial implementation can proceed serially through package -> CLI -> action/workflow -> docs -> validation.
   - If execution is split later, use file-scoped lanes: release package/tests, CLI/action integration, workflow fixtures, documentation.

Tradeoff decisions:

1. Choose a new `pkg/release` package over expanding `pkg/changelog` so release notes, version resolution, registry detection, and workflow rendering do not blur the CHANGELOG.md boundary.
2. Choose Git CLI subprocesses over a Go Git library to stay consistent with `pkg/changelog` and avoid new dependencies.
3. Choose typed workflow data plus string rendering over ad hoc concatenation so validation can happen before YAML rendering.
4. Choose direct Buildx shell steps over `docker/build-push-action` because the source workflows already use direct Buildx and acceptance requires parity with that behavior.
5. Choose generated workflow files over action-only orchestration because caller workflows must own permissions, checkout depth, concurrency, tag creation, and publish ordering.

## COMPONENTS

1. `pkg/release`
   - Owns public release APIs, release-resolution data, registry validation, workflow model data, and rendering.
   - Keeps internal helpers focused by concern: Git command execution, SemVer tag parsing/comparison, commit classification, release notes, image spec parsing, registry detection, and workflow rendering.
   - Provides test helpers or package-internal test fixtures for temporary Git repositories and workflow rendering assertions.
2. `pkg/cli/release.go`
   - Owns `mint release`, `mint release resolve`, and `mint release workflow` command definitions.
   - Converts flags into `pkg/release` options.
   - Handles stdout, `--output`, and `--github-output` without implementing release logic.
3. `pkg/cli/root.go`
   - Updates release-surface wording only after release behavior exists.
   - Leaves version/help behavior unchanged.
4. `action.yml`
   - Builds the Mint binary exactly as it does today.
   - Adds release-related inputs and outputs.
   - Routes only allowlisted commands to the binary.
5. `README.md`
   - Documents release resolution, workflow generation, GitHub Action usage, GHCR/ECR examples, and remaining non-goals.
6. `docs/CONSTITUTION.md`
   - Updates implementation facts after the release package, CLI, action, and workflow generator exist.
7. `docs/PROJECT_PROGRESS_SUMMARY.md`
   - Tracks phase and completion status for this feature.

## DATA

1. Release options:
   - `Commitish`: Git ref to resolve, default `HEAD`.
   - `WorkDir`: optional repository directory for tests and future embedding.
   - `GitHubOutput`: optional file path for GitHub output writing.
2. Release result:
   - `VersionTag`
   - `VersionBump`
   - `BaseTag`
   - `TargetSHA`
   - `ShortSHA`
   - `NeedsGitTag`
   - `CommitCount`
   - `ReleaseNotes`
3. Bump enum:
   - `already-tagged`
   - `patch`
   - `minor`
   - `major`
4. Internal commit evaluation record:
   - full SHA
   - 12-character short SHA
   - subject
   - body
   - normalized type
   - breaking flag
   - reason
   - bump
   - rank
5. SemVer tag value:
   - original tag string
   - major
   - minor
   - patch
   - target commit SHA when needed for head-tag checks
6. Image spec:
   - `Name`
   - `URI`
   - `Dockerfile`
   - `Context`, default `.`
   - derived registry host
   - derived registry kind
7. Registry kind enum:
   - `ghcr`
   - `ecr`
8. Workflow options:
   - image specs
   - output path
   - Mint action ref, default `v1`
9. Workflow model:
   - workflow name
   - permissions
   - environment variables or direct URI values
   - release-resolution step
   - tag-creation step
   - registry-login steps
   - per-image Buildx publish steps
10. No persistent storage, database schema, local state file, or generated cache is introduced.

## INTERFACES

1. `mint release resolve`
   - Inputs: `--commitish`, `--github-output`.
   - Output: resolved `version_tag` on stdout.
   - Side effects: writes GitHub output fields only when `--github-output` is set.
   - No side effects: does not create tags, push tags, write changelog files, or publish images.
2. `mint release workflow`
   - Inputs: repeatable `--image`, optional `--output`, optional `--mint-ref`.
   - Output: generated workflow YAML to stdout or the requested output file.
   - Side effects: writes only the requested `--output` file.
   - No side effects: does not call GitHub, Docker, AWS, or create workflow directories unless the provided output path requires a normal file write.
3. `action.yml`
   - Input: `command: release-resolve`.
   - Release inputs: `commitish`, plus existing `go-version` and command input.
   - Outputs: release result fields plus existing `mint-path` and `output`.
   - Side effects: builds Mint and runs the allowlisted command.
4. Generated GHCR workflow
   - Permissions include tag creation and package publishing.
   - Authenticates to `ghcr.io` with `${{ github.actor }}` and `${{ secrets.GITHUB_TOKEN }}`.
   - Publishes every image with `:${{ steps.release.outputs.version_tag }}` and `:latest`.
5. Generated ECR workflow
   - Permissions include tag creation and OIDC.
   - Requires `AWS_PUBLISH_ROLE_TO_ASSUME`.
   - Configures AWS credentials, logs in to ECR, and publishes every image with the release tag and `latest`.
6. File/artifact touch points during implementation:
   - add `pkg/release/*`
   - add or update `pkg/cli/release.go`
   - update `pkg/cli/root.go`
   - update `action.yml`
   - update `README.md`
   - update `docs/CONSTITUTION.md`
   - update `docs/PROJECT_PROGRESS_SUMMARY.md`
   - add tests under `pkg/release` and relevant CLI/action fixture tests

## DEPENDENCIES

References are tracked in front matter.

Implementation dependencies:

1. Git CLI must be available for release-resolution tests and runtime behavior.
2. Existing Go module and Cobra dependency remain the CLI foundation.
3. Existing `pkg/changelog` tests provide the temporary Git repository pattern to reuse, not a shared release implementation.
4. Generated workflow validation needs a YAML parser available through the current Go or local toolchain. If no YAML parser is already available, implementation may use a test-only YAML dependency or a lightweight syntax check chosen during tasks.
5. GitHub Actions, GHCR, Docker login, AWS credentials, and ECR login docs constrain generated workflow shape.
6. No Figma, MCP design resource, dataset, hosted service, secret store, or external runtime system is required for this plan.

## RISKS

1. Risk: Release resolution could accidentally use tags from unrelated history.
   Mitigation: Resolve base tags only with reachability from the target commit and test with an unrelated higher tag.
2. Risk: The resolver could mutate Git state.
   Mitigation: Keep resolver helpers read-only and put tag creation only in generated workflow text.
3. Risk: Action output wiring could silently drop multiline `release_notes`.
   Mitigation: Use a temporary GitHub output file in the action step and parse/map each field into explicit action outputs; test multiline output behavior.
4. Risk: Workflow rendering could produce syntactically valid YAML with unsafe or wrong step ordering.
   Mitigation: Test both YAML parsing and semantic string/structure assertions for checkout, resolve, tag-create, auth, and publish ordering.
5. Risk: Registry detection could accept ambiguous or tagged image URIs.
   Mitigation: Validate URI host, reject image tags, require one registry kind per workflow, and test unsupported/mixed registries.
6. Risk: Documentation could overclaim release implementation scope.
   Mitigation: Update README and constitution after implementation evidence exists and keep GitHub Releases, ECS deployment, package-manager publishing, and unsupported registries listed as non-goals.
7. Risk: The package grows too large if resolver and workflow rendering live in one file.
   Mitigation: Split implementation files by responsibility within `pkg/release` while keeping one public package.
8. Risk: Generated ECR defaults could leak source repo account values.
   Mitigation: Use placeholders or caller-owned variables/secrets, never hard-code sibling repository account IDs from evidence workflows.

## TESTING

1. Release-resolution unit and integration tests:
   - temporary Git repo with first `feat:` release -> `v0.1.0` for [AC-03]
   - temporary Git repo with first `fix:` release -> `v0.0.1` for [AC-04]
   - temporary Git repo with first non-conventional release -> `v0.0.1` for [AC-05]
   - temporary Git repo with first breaking release -> `v1.0.0` for [AC-06]
   - reachable base tag plus `fix:` -> `v1.0.1` for [AC-07]
   - reachable base tag plus `feat:` -> `v1.1.0` for [AC-08]
   - reachable base tag plus breaking commit -> `v2.0.0` for [AC-09]
   - unrelated higher tag ignored -> [AC-10]
   - target already SemVer-tagged -> [AC-11]
   - result fields and 12-character short SHA -> [AC-12], [AC-13]
2. CLI tests:
   - command help includes `mint release resolve` and `mint release workflow` for [AC-01], [AC-02]
   - `mint release resolve` prints only the version tag on success unless GitHub output is requested
   - `mint release workflow` writes stdout and `--output` paths predictably
3. Action tests:
   - `action.yml` parses as YAML
   - allowlist includes `release-resolve` for [AC-14]
   - first-class outputs are present for [AC-15]
   - unsupported commands still fail without shell passthrough for [AC-16]
4. Workflow-rendering tests:
   - GHCR image renders successfully for [AC-17]
   - ECR image renders successfully for [AC-18]
   - two-image workflow renders successfully for [AC-19]
   - unsupported registry fails for [AC-20]
   - mixed registry kinds fail for [AC-21]
   - concurrency block appears for [AC-22]
   - tag-creation step appears before Docker publish steps for [AC-23]
   - tag collision guard is rendered and never moves tags for [AC-24]
   - every image gets release tag and `latest` for [AC-25]
   - rendered YAML parses for [AC-26]
   - rendered YAML omits ECS deployment content for [AC-27]
5. Documentation checks:
   - README release CLI/action sections satisfy [AC-28]
   - constitution and progress summary distinguish implemented release scope after implementation for [AC-29]
6. Project validation:
   - `go test ./...` for [AC-30]
   - `go vet ./...` for [AC-31]
   - `make build` for [AC-32]
   - `git diff --check` for [AC-33]
   - secret/local-state scan over touched files for [AC-34]
   - `kit map 0005-pull-forward-v1-features` for [AC-35]
