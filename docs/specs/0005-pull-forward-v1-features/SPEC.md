---
kit_metadata_version: 1
artifact: spec
feature:
  id: 0005
  slug: pull-forward-v1-features
  dir: 0005-pull-forward-v1-features
summary: Adds release resolution and GHCR/ECR publish workflow generation to Mint using the proven Git-tag-first container release pattern; ECS deployment and GitHub Release creation stay out of scope.
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
    reason: Release commands must follow the established Go module, Cobra command, Makefile, README, and package layout.
  - type: depends_on
    target: 0003-github-action
    reason: Release behavior must extend the public composite action without arbitrary command execution.
  - type: builds_on
    target: 0004-changelog-generation
    reason: Release resolution shares Git ref and Conventional Commit concepts with changelog generation, while keeping CHANGELOG.md rendering separate.
references:
  - id: constitution
    name: Mint constitution
    type: repo_doc
    target: docs/CONSTITUTION.md
    relation: constrains
    read_policy: must
    used_for: source-of-truth order, release-contract boundary, and fact-versus-intent wording
    status: active
  - id: rlm-guide
    name: RLM guide
    type: repo_doc
    target: docs/agents/RLM.md
    relation: guides
    read_policy: must
    used_for: broad-context discovery and downstream planning mode
    status: active
  - id: progress-summary
    name: Project progress summary
    type: repo_doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    relation: informs
    read_policy: must
    used_for: feature phase, prior work index, and spec completion status
    status: active
  - id: brainstorm
    name: Pull-forward brainstorm
    type: feature_doc
    target: docs/specs/0005-pull-forward-v1-features/BRAINSTORM.md
    relation: informs
    read_policy: must
    used_for: upstream research, source map, tradeoffs, and defaults resolved into this spec
    status: active
  - id: cli-patterns-spec
    name: CLI patterns spec
    type: feature_doc
    target: docs/specs/0002-cli-patterns/SPEC.md
    relation: constrains
    read_policy: must
    used_for: existing CLI/build/README pattern
    status: active
  - id: github-action-spec
    name: GitHub action spec
    type: feature_doc
    target: docs/specs/0003-github-action/SPEC.md
    relation: constrains
    read_policy: must
    used_for: composite action wrapper and fixed command allowlist contract
    status: active
  - id: changelog-spec
    name: Changelog generation spec
    type: feature_doc
    target: docs/specs/0004-changelog-generation/SPEC.md
    relation: informs
    read_policy: must
    used_for: adjacent Git ref and Conventional Commit behavior
    status: active
  - id: readme
    name: Mint README
    type: repo_doc
    target: README.md
    relation: updates
    read_policy: evidence
    used_for: current user-facing release claims and post-feature quick start updates
    status: active
  - id: action-yml
    name: Mint composite action
    type: action
    target: action.yml
    relation: updates
    read_policy: evidence
    used_for: action inputs, outputs, and supported command allowlist
    status: active
  - id: cli-root
    name: Mint root command
    type: source
    target: pkg/cli/root.go
    selector_type: symbol
    selector: rootCmd
    relation: updates
    read_policy: evidence
    used_for: root help, release surface wording, and version command pattern
    status: active
  - id: cli-changelog
    name: Mint changelog command
    type: source
    target: pkg/cli/changelog.go
    selector_type: symbol
    selector: changelogCmd
    relation: informs
    read_policy: evidence
    used_for: Cobra command style and flag binding pattern
    status: active
  - id: changelog-package
    name: Changelog package
    type: source
    target: pkg/changelog
    relation: informs
    read_policy: evidence
    used_for: Git-backed package and temporary repository test pattern
    status: active
  - id: r2-release-script
    name: r2 release resolver
    type: source
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/r2/scripts/resolve-release-version.sh
    relation: informs
    read_policy: evidence
    used_for: proven release-resolution rules and GitHub output fields
    status: active
  - id: r2-main-workflow
    name: r2 publish workflow
    type: workflow
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/r2/.github/workflows/main.yaml
    relation: informs
    read_policy: evidence
    used_for: single-image ECR publish workflow behavior
    status: active
  - id: flowcore-main-workflow
    name: Flowcore publish workflow
    type: workflow
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/flowcore/.github/workflows/main.yaml
    relation: informs
    read_policy: evidence
    used_for: multi-image ECR publish workflow behavior
    status: active
  - id: event-sink-main-workflow
    name: event-sink publish workflow
    type: workflow
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/event-sink/.github/workflows/main.yaml
    relation: informs
    read_policy: evidence
    used_for: single-image ECR publish workflow confirmation
    status: active
  - id: flowcore-ecs-deploy
    name: Flowcore ECS deploy workflow
    type: workflow
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/flowcore/.github/workflows/deploy-ecs.yaml
    relation: constrains
    read_policy: evidence
    used_for: explicit negative scope for environment-specific ECS deployment
    status: active
  - id: conventional-commits
    name: Conventional Commits 1.0.0
    type: url
    target: https://www.conventionalcommits.org/en/v1.0.0/
    relation: guides
    read_policy: evidence
    used_for: commit structure and feat/fix/breaking SemVer mapping
    status: active
  - id: semver
    name: Semantic Versioning 2.0.0
    type: url
    target: https://semver.org/
    relation: guides
    read_policy: evidence
    used_for: major/minor/patch semantics and immutable release expectation
    status: active
  - id: github-workflow-syntax
    name: GitHub Actions workflow syntax
    type: url
    target: https://docs.github.com/actions/using-workflows/workflow-syntax-for-github-actions
    relation: guides
    read_policy: evidence
    used_for: workflow permissions, jobs, and output wiring
    status: active
  - id: github-concurrency
    name: GitHub Actions concurrency
    type: url
    target: https://docs.github.com/actions/writing-workflows/choosing-what-your-workflow-does/control-the-concurrency-of-workflows-and-jobs
    relation: guides
    read_policy: evidence
    used_for: release-publish concurrency behavior
    status: active
  - id: github-container-registry
    name: GitHub Container Registry
    type: url
    target: https://docs.github.com/packages/working-with-a-github-packages-registry/working-with-the-container-registry
    relation: guides
    read_policy: evidence
    used_for: GHCR authentication and package permission requirements
    status: active
  - id: docker-login-action
    name: Docker login action
    type: url
    target: https://github.com/docker/login-action
    relation: guides
    read_policy: evidence
    used_for: GHCR registry login step
    status: active
  - id: aws-configure-credentials
    name: AWS configure credentials action
    type: url
    target: https://github.com/aws-actions/configure-aws-credentials
    relation: guides
    read_policy: evidence
    used_for: ECR OIDC credential setup
    status: active
  - id: aws-ecr-login
    name: AWS Amazon ECR login action
    type: url
    target: https://github.com/aws-actions/amazon-ecr-login
    relation: guides
    read_policy: evidence
    used_for: ECR Docker login step
    status: active
---
# SPEC

## SUMMARY

Adds release resolution and publish workflow generation to Mint so maintainers can compute the next `vX.Y.Z` tag from Git history and generate Git-tag-first GHCR/ECR container publish workflows. The feature must keep `mint` as the core CLI, expose safe GitHub Action outputs, and exclude ECS deployment, GitHub Release creation, and package-manager-specific publishing.

## PROBLEM

Mint currently has a CLI scaffold, changelog generation, and a composite GitHub Action wrapper, but release computation, Git tag creation rules, and container publish workflow generation are still future-scoped. The release-publish pattern already exists in `r2`, Flowcore, and event-sink as copied shell/YAML logic, which leaves maintainers repeating bespoke scripts for SemVer resolution, tag-first publishing, and image tagging.

## GOALS

1. Provide a Mint release-resolution command that computes a deterministic SemVer Git tag from a target commit and reachable Git history.
2. Preserve the proven release semantics from `r2`, Flowcore, and event-sink: reachable tag lookup, Conventional Commit bump ranking, patch fallback, and already-tagged target detection.
3. Expose release-resolution outputs through the CLI and GitHub Action using stable field names.
4. Provide a Mint workflow-generation command that emits a default-branch publish workflow for GHCR or ECR images based on full image URI detection.
5. Support one or more container image specs so both single-image services and Flowcore-style multi-image services are covered.
6. Generate or document publish workflows that create the Git SemVer tag before any image push, never move tags, and publish both `:<version>` and `:latest`.
7. Update README, action documentation, and project docs so implemented release behavior is no longer described as absent after implementation.

## NON-GOALS

1. Do not generate ECS deployment workflows, ECS variables, ECS service updates, task definition mutation, or deployment commands.
2. Do not create GitHub Releases through the GitHub API.
3. Do not implement package-manager-specific publishing for Go modules, npm, Homebrew, container attestations, SBOMs, or language registries.
4. Do not support registries beyond GHCR and AWS ECR in this feature.
5. Do not accept arbitrary shell commands through `action.yml`.
6. Do not make `mint release resolve` create or push Git tags directly.
7. Do not merge CHANGELOG.md rendering into release notes; CHANGELOG.md generation remains owned by `mint changelog`.
8. Do not add secrets, `.env` values, private keys, local machine configuration, or generated runtime state.

## USERS

1. Maintainers of Go services who want a shared release/versioning contract instead of copied release scripts.
2. Maintainers publishing one or more container images from GitHub Actions to GHCR or AWS ECR.
3. Coding agents implementing and validating Mint release behavior from a binding spec.
4. Repository operators reviewing generated workflows before committing them to downstream projects.

## SKILLS

Skills are tracked in front matter.

Selected skill set:

1. `rlm` from `docs/agents/RLM.md` is required for this feature because implementation planning may otherwise load broad repo, workflow, sibling-repo, and external-action context. Use it for trigger phrases including `analyze codebase`, `scan all files`, `large repository analysis`, `scan repository`, and `recursive language model`.
2. No repo-local `.agents/skills/*/SKILL.md` files exist.
3. No additional global skill applies to this release-spec feature.

## RELATIONSHIPS

Relationships are tracked in front matter.

1. `depends_on: 0002-cli-patterns`
2. `depends_on: 0003-github-action`
3. `builds_on: 0004-changelog-generation`

## DEPENDENCIES

References are tracked in front matter.

Runtime and workflow dependencies for implementation:

1. The Git CLI is required for commit resolution, tag lookup, and commit traversal.
2. Existing Go/Cobra CLI dependencies are sufficient for command exposure.
3. GHCR publish workflows require GitHub token/package permissions and Docker registry login.
4. ECR publish workflows require AWS OIDC credentials and Amazon ECR login.
5. Generated workflows require Docker Buildx and a full checkout with tags.
6. No external Go SemVer library is required by this spec; a strict `vX.Y.Z` parser is sufficient unless implementation proves otherwise.

## COMMANDS

1. Mint must add a `mint release` command group.
2. Mint must add `mint release resolve`.
3. Mint must add `mint release workflow`.
4. `mint changelog` remains the command for CHANGELOG.md generation.

### `mint release resolve`

Inputs:

1. `--commitish`: Git ref to resolve, default `HEAD`.
2. `--github-output`: optional path to a GitHub Actions output file. When provided, write all release output fields in GitHub Actions output format.

Outputs:

1. Print the resolved `version_tag` to stdout on success.
2. Return or expose these fields from the release resolver:
   - `version_tag`
   - `version_bump`
   - `base_tag`
   - `target_sha`
   - `short_sha`
   - `needs_git_tag`
   - `commit_count`
   - `release_notes`
3. `short_sha` must use 12 Git hash characters.
4. `needs_git_tag` must be `true` only when a new SemVer Git tag should be created for the target commit.

### `mint release workflow`

Inputs:

1. `--image`: repeatable image spec in the form `name=<name>,uri=<image-uri>,dockerfile=<path>,context=<path>`.
2. `--output`: optional workflow file path. If omitted, render the workflow YAML to stdout.
3. `--mint-ref`: action ref used by generated workflows, default `v1`.

Image spec rules:

1. `name`, `uri`, and `dockerfile` are required.
2. `context` is optional and defaults to `.`.
3. `uri` must be an image repository URI without a tag.
4. Every image in one generated workflow must use the same supported registry kind.
5. Duplicate image names must fail validation.

Generated workflow outputs:

1. The workflow must publish each image with `:<version_tag>` and `:latest`.
2. The workflow must use the Mint action or CLI to resolve the release before tag creation and publishing.
3. The workflow must be valid GitHub Actions YAML.

## RELEASE RESOLUTION

1. SemVer Git tags are tags whose names match `^v[0-9]+\.[0-9]+\.[0-9]+$`.
2. Pre-release and build metadata tags are not release base tags in this feature.
3. The target commit is resolved from `--commitish` as a Git commit.
4. If `--commitish` cannot be resolved to a commit, resolution must fail closed with a clear ref-not-found error.
5. If one or more SemVer tags already point at the target commit, Mint must choose the highest SemVer tag on the target, set `version_bump=already-tagged`, set `needs_git_tag=false`, set `commit_count=0`, and reuse that tag.
6. If no SemVer tag points at the target, Mint must choose the highest SemVer tag reachable from the target commit as `base_tag`.
7. Tags from unrelated branches that are not reachable from the target commit must not affect resolution.
8. If no reachable SemVer tag exists, Mint must use an implicit base version of `v0.0.0` and an empty `base_tag`.
9. Mint must evaluate commits after `base_tag` through the target commit, oldest first.
10. If there are no commits after the reachable base tag and no SemVer tag points at the target, Mint must reuse the base version, set `version_bump=already-tagged`, set `needs_git_tag=false`, and explain the reuse in `release_notes`.
11. Mint must select the highest bump rank found in the commit range:
    - major rank: breaking change
    - minor rank: `feat`
    - patch rank: `fix`, any other conventional type, and any non-conventional commit
12. Conventional Commit type matching must be case-insensitive.
13. A breaking change is any conventional commit with `!` before the colon or a commit body/footer containing `BREAKING CHANGE:` or `BREAKING-CHANGE:`.
14. A major bump increments major and resets minor and patch to `0`.
15. A minor bump increments minor and resets patch to `0`.
16. A patch bump increments patch.
17. A first release with one `feat:` commit must resolve to `v0.1.0`.
18. A first release with one `fix:` or non-conventional commit must resolve to `v0.0.1`.
19. A first release with one breaking commit must resolve to `v1.0.0`.

## RELEASE NOTES

1. `release_notes` are lightweight tag-annotation notes, not CHANGELOG.md content.
2. `release_notes` must include:
   - resolved release tag
   - base tag or `none`
   - target commit SHA
   - selected bump
   - one line per evaluated commit with short SHA, subject, reason, and bump
3. CHANGELOG.md output remains separate and is generated by `mint changelog`.

## GITHUB ACTION

1. `action.yml` must keep building the Mint CLI from `cmd/mint`.
2. `action.yml` must keep using a fixed command allowlist.
3. `action.yml` must add an allowlisted `release-resolve` command that runs `mint release resolve`.
4. `action.yml` must not accept arbitrary shell commands or unsafely interpolate command input.
5. `action.yml` must expose these first-class outputs:
   - `version_tag`
   - `version_bump`
   - `base_tag`
   - `target_sha`
   - `short_sha`
   - `needs_git_tag`
   - `commit_count`
   - `release_notes`
6. The existing `output` action output may remain as captured stdout for backward compatibility.
7. The action must preserve existing supported commands unless implementation discovers a direct incompatibility that is documented before code changes.

## WORKFLOW GENERATION

1. Generated publish workflows must run on `push`.
2. Generated publish workflows must publish only when the push is to the repository default branch.
3. Generated publish workflows must include `concurrency.group: release-publish` and `cancel-in-progress: false`.
4. Generated publish workflows must check out full history and tags.
5. Generated publish workflows must refresh Git tags before release resolution.
6. Generated publish workflows must resolve the release before creating a tag or publishing images.
7. Generated publish workflows must create and push an annotated Git SemVer tag before any image publish when `needs_git_tag=true`.
8. Generated publish workflows must never move existing tags.
9. If the computed tag exists on the target commit, generated workflows must continue.
10. If the computed tag exists on any other commit, generated workflows must fail before image publishing with a clear recovery message.
11. Generated publish workflows must configure Docker Buildx before image publishing.
12. Generated publish workflows must publish each configured image with the resolved SemVer tag and with `latest`.
13. Generated publish workflows must use direct `docker buildx build --push` shell steps for image publishing to preserve parity with the proven source workflows.
14. Generated publish workflows must not include ECS deployment jobs, ECS variables, ECS service/container names, task definition mutation, `aws ecs` commands, `workflow_dispatch` deployment gates, or environment deployment blocks.

## REGISTRY SUPPORT

1. Registry detection must be based on each image URI host.
2. `ghcr.io/...` image URIs resolve to registry kind `ghcr`.
3. AWS ECR image URIs matching `<account>.dkr.ecr.<region>.amazonaws.com/...` resolve to registry kind `ecr`.
4. Unsupported registry hosts must fail closed with an explicit unsupported-registry error.
5. Mixed registry kinds in one generated workflow must fail closed.
6. GHCR workflows must include permissions needed for tag creation and package publishing.
7. GHCR workflows must authenticate to `ghcr.io` with `${{ github.actor }}` and `${{ secrets.GITHUB_TOKEN }}`.
8. ECR workflows must include permissions needed for tag creation and OIDC authentication.
9. ECR workflows must configure AWS credentials using `aws-actions/configure-aws-credentials`.
10. ECR workflows must log in using `aws-actions/amazon-ecr-login`.
11. ECR workflows must require a repository secret named `AWS_PUBLISH_ROLE_TO_ASSUME`.

## DOCUMENTATION REQUIREMENTS

1. README must document `mint release resolve`.
2. README must document `mint release workflow`.
3. README must include a GitHub Actions quick start for release resolution through the Mint action.
4. README must include a publish workflow example or generator example for GHCR.
5. README must include a publish workflow example or generator example for ECR.
6. README must no longer say all release computation, tagging, and publishing behavior is unimplemented after this feature is implemented.
7. README must still state that GitHub Release creation, ECS deployment, package-manager-specific publishing, and unsupported registries are out of scope.
8. `docs/CONSTITUTION.md` must be updated after implementation to distinguish newly implemented release-resolution/publish-workflow behavior from remaining non-goals.
9. `docs/PROJECT_PROGRESS_SUMMARY.md` must remain aligned with the highest completed artifact.

## ACCEPTANCE

1. [AC-01] `mint release resolve` exists in CLI help.
2. [AC-02] `mint release workflow` exists in CLI help.
3. [AC-03] A temporary Git repository with no SemVer tags and one `feat:` commit resolves to `v0.1.0`.
4. [AC-04] A temporary Git repository with no SemVer tags and one `fix:` commit resolves to `v0.0.1`.
5. [AC-05] A temporary Git repository with no SemVer tags and one non-conventional commit resolves to `v0.0.1`.
6. [AC-06] A temporary Git repository with no SemVer tags and one breaking commit resolves to `v1.0.0`.
7. [AC-07] A temporary Git repository with reachable tag `v1.0.0` and one `fix:` commit resolves to `v1.0.1`.
8. [AC-08] A temporary Git repository with reachable tag `v1.0.0` and one `feat:` commit resolves to `v1.1.0`.
9. [AC-09] A temporary Git repository with reachable tag `v1.0.0` and one breaking commit resolves to `v2.0.0`.
10. [AC-10] A temporary Git repository with reachable tag `v1.0.0`, target-branch commit history, and unrelated branch tag `v9.0.0` resolves against `v1.0.0`, not `v9.0.0`.
11. [AC-11] A target commit already tagged with a SemVer tag returns that tag with `needs_git_tag=false`.
12. [AC-12] Release outputs include `version_tag`, `version_bump`, `base_tag`, `target_sha`, `short_sha`, `needs_git_tag`, `commit_count`, and `release_notes`.
13. [AC-13] `short_sha` is 12 characters.
14. [AC-14] `action.yml` supports an allowlisted `release-resolve` command.
15. [AC-15] `action.yml` exposes first-class release outputs.
16. [AC-16] `action.yml` rejects unsupported command values without running arbitrary shell.
17. [AC-17] GHCR workflow generation succeeds for at least one `ghcr.io/...` image URI.
18. [AC-18] ECR workflow generation succeeds for at least one AWS ECR image URI.
19. [AC-19] Workflow generation supports at least two images in one workflow.
20. [AC-20] Unsupported registry hosts fail with an unsupported-registry error.
21. [AC-21] Mixed registry kinds in one workflow fail closed.
22. [AC-22] Generated workflows include `concurrency.group: release-publish` and `cancel-in-progress: false`.
23. [AC-23] Generated workflows create the Git tag before any Docker image publish step.
24. [AC-24] Generated workflows fail on tag collision with another commit and never move the tag.
25. [AC-25] Generated workflows tag every configured image with both the SemVer tag and `latest`.
26. [AC-26] Generated workflows parse as valid YAML.
27. [AC-27] Generated workflows contain no ECS deployment workflow content.
28. [AC-28] README documents release CLI and action usage.
29. [AC-29] `docs/CONSTITUTION.md` and `docs/PROJECT_PROGRESS_SUMMARY.md` reflect the implemented release scope after implementation.
30. [AC-30] `go test ./...` exits 0.
31. [AC-31] `go vet ./...` exits 0.
32. [AC-32] `make build` exits 0.
33. [AC-33] `git diff --check` exits 0 for touched files.
34. [AC-34] No secrets, `.env` values, private keys, tokens, generated local state, or machine-local config are added.
35. [AC-35] `kit map 0005-pull-forward-v1-features` resolves spec relationships, skills, and references.

## EDGE-CASES

1. If the target ref cannot be resolved, fail closed before reading commits.
2. If the repository has no SemVer tags, start from `v0.0.0`.
3. If multiple SemVer tags point at the target commit, choose the highest SemVer tag.
4. If the target commit is already SemVer-tagged, do not request a new tag.
5. If an unrelated branch has a higher SemVer tag, ignore it unless it is reachable from the target.
6. If all commits are non-conventional, resolve a patch bump.
7. If a commit type is conventional but not `feat` or `fix`, resolve a patch bump unless it is breaking.
8. If any commit is breaking, resolve a major bump regardless of other commits.
9. If an image spec omits `name`, `uri`, or `dockerfile`, fail validation.
10. If two image specs use the same `name`, fail validation.
11. If an image URI includes a tag, fail validation.
12. If image URI registry detection fails, do not generate a workflow.
13. If GHCR and ECR images are mixed in one workflow request, fail validation.
14. If a generated workflow would contain ECS deployment keys or commands, it fails acceptance.
15. If a computed Git tag already exists on a different commit, generated workflows fail before image publish.
16. If `GITHUB_OUTPUT` handling is unavailable, CLI release resolution still prints the resolved version tag on stdout.

## OPEN QUESTIONS

none
