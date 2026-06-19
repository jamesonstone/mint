---
kit_metadata_version: 1
artifact: brainstorm
feature:
  id: 0005
  slug: pull-forward-v1-features
  dir: 0005-pull-forward-v1-features
relationships:
  - type: depends_on
    target: 0002-cli-patterns
    reason: Release commands must follow the established Go module, Cobra command, Makefile, README, and package layout.
  - type: depends_on
    target: 0003-github-action
    reason: Release behavior must extend the public composite action without arbitrary command execution.
  - type: builds_on
    target: 0004-changelog-generation
    reason: Release resolution shares Git ref and Conventional Commit parsing concerns with changelog generation.
references:
  - id: feature-notes
    name: Feature notes
    type: notes
    target: docs/notes/0005-pull-forward-v1-features
    relation: informs
    read_policy: conditional
    used_for: optional pre-brainstorm research input; only .gitkeep exists
    status: optional
  - id: constitution
    name: Mint constitution
    type: doc
    target: docs/CONSTITUTION.md
    relation: constrains
    read_policy: must
    used_for: current product scope, non-goals, and document-first workflow rules
    status: active
  - id: progress-summary
    name: Project progress summary
    type: doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    relation: informs
    read_policy: must
    used_for: RLM index, current feature phase, and durable status update
    status: active
  - id: readme
    name: Mint README
    type: doc
    target: README.md
    relation: informs
    read_policy: evidence
    used_for: current user-facing claims and future-scoped release wording
    status: active
  - id: action-yml
    name: Mint composite action
    type: action
    target: action.yml
    relation: implements
    read_policy: evidence
    used_for: current action inputs, outputs, and command allowlist
    status: active
  - id: cli-root
    name: Mint root command
    type: code
    target: pkg/cli/root.go
    selector_type: symbol
    selector: rootCmd
    relation: implements
    read_policy: evidence
    used_for: current CLI wording, version behavior, and future-scoped release contract
    status: active
  - id: cli-changelog
    name: Mint changelog command
    type: code
    target: pkg/cli/changelog.go
    selector_type: symbol
    selector: changelogCmd
    relation: implements
    read_policy: evidence
    used_for: current Cobra command pattern and root-level flag compatibility
    status: active
  - id: changelog-package
    name: Changelog package
    type: code
    target: pkg/changelog
    relation: implements
    read_policy: evidence
    used_for: testable Git-backed package pattern for release-adjacent behavior
    status: active
  - id: cli-patterns-spec
    name: CLI patterns spec
    type: doc
    target: docs/specs/0002-cli-patterns/SPEC.md
    relation: constrains
    read_policy: conditional
    used_for: shared CLI/build/README patterns
    status: active
  - id: github-action-spec
    name: GitHub action spec
    type: doc
    target: docs/specs/0003-github-action/SPEC.md
    relation: constrains
    read_policy: conditional
    used_for: composite action command allowlist and public wrapper scope
    status: active
  - id: changelog-spec
    name: Changelog generation spec
    type: doc
    target: docs/specs/0004-changelog-generation/SPEC.md
    relation: informs
    read_policy: conditional
    used_for: adjacent Git/Conventional Commit parsing behavior
    status: active
  - id: r2-release-script
    name: r2 release resolver
    type: code
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/r2/scripts/resolve-release-version.sh
    relation: informs
    read_policy: evidence
    used_for: proven SemVer resolution, reachable tag lookup, patch fallback, and GitHub outputs
    status: active
  - id: r2-main-workflow
    name: r2 publish workflow
    type: workflow
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/r2/.github/workflows/main.yaml
    relation: informs
    read_policy: evidence
    used_for: single-image ECR publish workflow shape
    status: active
  - id: flowcore-main-workflow
    name: Flowcore publish workflow
    type: workflow
    target: /Users/jamesonstone/go/src/github.com/lsmc-bio/flowcore/.github/workflows/main.yaml
    relation: informs
    read_policy: evidence
    used_for: multi-image ECR publish workflow shape
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
    relation: informs
    read_policy: evidence
    used_for: negative scope boundary for deployment-specific ECS automation
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
    used_for: major/minor/patch increment semantics and immutable release expectation
    status: active
  - id: github-workflow-syntax
    name: GitHub Actions workflow syntax
    type: url
    target: https://docs.github.com/actions/using-workflows/workflow-syntax-for-github-actions
    relation: guides
    read_policy: evidence
    used_for: permissions, jobs, and workflow syntax
    status: active
  - id: github-concurrency
    name: GitHub Actions concurrency
    type: url
    target: https://docs.github.com/actions/writing-workflows/choosing-what-your-workflow-does/control-the-concurrency-of-workflows-and-jobs
    relation: guides
    read_policy: evidence
    used_for: release-publish concurrency rules
    status: active
  - id: github-container-registry
    name: GitHub Container Registry
    type: url
    target: https://docs.github.com/packages/working-with-a-github-packages-registry/working-with-the-container-registry
    relation: guides
    read_policy: evidence
    used_for: GHCR image hosting and GitHub Actions authentication
    status: active
  - id: docker-login-action
    name: Docker login action
    type: url
    target: https://github.com/docker/login-action
    relation: guides
    read_policy: evidence
    used_for: GHCR registry login workflow step
    status: active
  - id: docker-build-push-action
    name: Docker build-push action
    type: url
    target: https://github.com/docker/build-push-action
    relation: guides
    read_policy: evidence
    used_for: optional generated build and push workflow step
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
    used_for: ECR Docker login workflow step
    status: active
  - id: aws-ecr-pattern
    name: AWS ECR GitHub Actions pattern
    type: url
    target: https://docs.aws.amazon.com/prescriptive-guidance/latest/patterns/build-and-push-docker-images-to-amazon-ecr-using-github-actions-and-terraform.html
    relation: informs
    read_policy: evidence
    used_for: ECR build and push workflow model
    status: active
---
# BRAINSTORM

## SUMMARY

Mint should pull the proven `r2`, Flowcore, and event-sink release-publish pattern into a reusable CLI and GitHub Action contract: resolve the next SemVer Git tag from commit history, expose deterministic GitHub Actions outputs, and generate or document GHCR/ECR container publish workflows that push both `:<version>` and `:latest`. The feature should standardize release resolution and image publishing only; ECS deployment workflows remain out of scope because deployment targets and runtime variables are environment-specific.

## USER THESIS

### Context Synthesis

Implement Mint as the reusable release engine for the container CI/CD pattern proven in `r2`, Flowcore, and event-sink: compute the next SemVer tag from Conventional Commits with patch fallback, constrain prior tags to tags reachable from the target commit, create the Git tag before image publishing, emit GitHub Actions outputs, and generate registry-neutral workflow guidance that publishes `:<version>` and `:latest` to container image URIs detected as GHCR or ECR. Affected users are maintainers who want one CLI/action contract for release tagging and container publishing across Go services without copying bespoke shell scripts per repo. Done means `mint` exposes tested release-resolution behavior, action inputs/outputs support the release contract, docs stop claiming release behavior is future-scoped, and generated or documented workflows support GHCR and ECR through image URI detection.

### Source Map

1. `r2`, Flowcore, and event-sink now share a Git-tag-first container publish pattern with SemVer tags, `:latest`, release concurrency, reachable-tag lookup, and patch fallback.
2. `README.md`, `docs/CONSTITUTION.md`, and `pkg/cli/root.go` currently say release computation, tagging, and publishing are future-scoped.
3. `action.yml` currently supports only `version`, `help`, `changelog`, and `none`.
4. The corrected release resolver and publish workflow pattern exist in:
   - `/Users/jamesonstone/go/src/github.com/lsmc-bio/r2/scripts/resolve-release-version.sh`
   - `/Users/jamesonstone/go/src/github.com/lsmc-bio/r2/.github/workflows/main.yaml`
   - `/Users/jamesonstone/go/src/github.com/lsmc-bio/flowcore/.github/workflows/main.yaml`
   - `/Users/jamesonstone/go/src/github.com/lsmc-bio/event-sink/.github/workflows/main.yaml`
5. External constraints come from Conventional Commits 1.0.0, SemVer 2.0.0, GitHub Actions workflow syntax/concurrency/token permissions, GHCR docs, Docker registry actions, and AWS ECR GitHub Actions patterns.

### Acceptance Signals To Preserve

1. `feat:` produces a minor bump; `fix:` produces a patch bump.
2. Any other conventional or non-conventional commit defaults to patch.
3. `!` before the Conventional Commit colon or `BREAKING CHANGE:`/`BREAKING-CHANGE:` in the body produces a major bump.
4. Previous SemVer tags are selected only from tags reachable from the target commit.
5. A target commit already tagged with a SemVer tag returns `needs_git_tag=false`.
6. A generated or documented publish workflow creates the Git tag before image publishing.
7. A generated or documented publish workflow never moves an existing tag.
8. A generated or documented publish workflow pushes both the SemVer image tag and `latest`.
9. GHCR and ECR are selected by image URI host detection.
10. Unsupported registry hosts fail closed with an explicit unsupported-registry error.
11. No ECS workflow generation, ECS variables, ECS actions, or ECS deploy command is added.

## RELATIONSHIPS

Relationships are tracked in front matter:

1. `depends_on: 0002-cli-patterns` because release commands must fit the existing Go/Cobra/Makefile/README structure.
2. `depends_on: 0003-github-action` because action support must extend the existing composite action without allowing arbitrary shell execution.
3. `builds_on: 0004-changelog-generation` because release resolution shares Git ref collection and Conventional Commit concepts with the changelog package.

## CODEBASE FINDINGS

1. `docs/CONSTITUTION.md` currently allows Mint to compute the next version, write a changelog, and mint releases as product intent, but explicitly says release algorithms, tagging, publishing, external integrations, and package-manager-specific behavior are not implemented yet. This feature is the first proper spec vehicle for that release contract.
2. `README.md` presents Mint as a release tooling CLI but still tells users that release computation, tagging, publishing, GitHub releases, and package-manager-specific behavior are future-scoped. After implementation, README wording must narrow that future-scoped list to anything still out of scope, especially GitHub Releases, ECS deployment, and package-manager-specific releases.
3. `pkg/cli/root.go` exposes a small Cobra root command with version handling and help text. Its long description says the current command surface includes version reporting and changelog generation only. Release commands should update that text only when implementation exists.
4. `pkg/cli/changelog.go` is the current command pattern to reuse: a small flag struct, `cobra.NoArgs`, focused `RunE`, output written through `cmd.OutOrStdout()`, errors returned rather than printed inline, and root-level compatibility flags routed into the same package call.
5. `pkg/changelog` is the package pattern to reuse for release resolution: testable options/result structs, Git command execution behind focused helpers, deterministic parsing/rendering functions, and temporary Git repository tests in `pkg/changelog/generator_test.go`.
6. `action.yml` is a composite action that builds `cmd/mint` into `$RUNNER_TEMP`, adds the binary to `PATH`, and runs only allowlisted commands. Release action support should extend the allowlist and explicit inputs/outputs. It should not accept arbitrary command strings or shell fragments.
7. The current action output model has only `mint-path` and captured `output`. Release resolution needs first-class outputs: `version_tag`, `version_bump`, `base_tag`, `target_sha`, `short_sha`, `needs_git_tag`, `commit_count`, `release_notes`, and likely `container_image`/`registry_kind` when registry detection is requested.
8. `docs/specs/0002-cli-patterns/SPEC.md` established the Go module, `cmd/mint`, `pkg/cli`, Makefile, hooks, and README patterns while keeping release algorithms out of scope. This feature should not rewrite those patterns.
9. `docs/specs/0003-github-action/SPEC.md` established the action as a CLI wrapper with a fixed safe command allowlist. This feature can add release commands and outputs but should preserve the public-action safety posture.
10. `docs/specs/0004-changelog-generation/SPEC.md` implemented changelog generation from Git refs and Conventional Commits, but not release computation, tag creation, or image publishing. A release package can reuse design lessons but should remain separate from `pkg/changelog` unless concrete code reuse is obvious during implementation.
11. `/Users/jamesonstone/go/src/github.com/lsmc-bio/r2/scripts/resolve-release-version.sh` resolves `commitish` to a commit, finds the latest reachable `vX.Y.Z` tag using `git tag --merged "$target_sha" --sort=-v:refname`, and finds a SemVer tag already pointing at the target with `git tag --points-at "$target_sha"`.
12. The resolver script emits GitHub Actions outputs through `$GITHUB_OUTPUT`: `version_tag`, `version_bump`, `base_tag`, `target_sha`, `short_sha`, `needs_git_tag`, `commit_count`, and multiline `release_notes`.
13. The resolver script starts from `v0.0.0` when no reachable SemVer tag exists. A first `feat:` therefore resolves to `v0.1.0`; a first `fix:` or malformed commit would resolve to `v0.0.1`; a first breaking change would resolve to `v1.0.0` under the proven ranking rules.
14. The resolver script uses highest-rank bump selection across all commits after the base tag: patch rank 1, minor rank 2, major rank 3. It walks commits oldest-first for release note lines.
15. The resolver script treats a SemVer tag already on the target commit as `already-tagged`, emits `needs_git_tag=false`, `commit_count=0`, and reuses the existing tag rather than creating a new one.
16. `r2/.github/workflows/main.yaml`, `flowcore/.github/workflows/main.yaml`, and `event-sink/.github/workflows/main.yaml` all run on default-branch pushes, use `contents: write`, use release concurrency group `release-publish` with `cancel-in-progress: false`, check out full history/tags, refresh tags, resolve the release version, create and push an annotated Git tag before publishing images, then build and push images.
17. The publish workflows never move tags. If the computed tag exists on the same commit, the workflow continues. If it exists on a different commit, the workflow fails with a recovery instruction to inspect/fix tag state or push a follow-up commit after correction.
18. `r2` and event-sink are single-image ECR examples using `IMAGE_NAME` plus `Dockerfile`. Flowcore is a multi-image ECR example using `API_IMAGE`/`Dockerfile.api` and `WORKER_IMAGE`/`Dockerfile.worker`. Mint should either support multiple image specs in v1 or explicitly defer Flowcore parity; silent single-image-only support would not satisfy the source thesis.
19. The source workflows publish ECR images by configuring AWS credentials and using `aws-actions/amazon-ecr-login`, then running `docker buildx build --push` with both `:$VERSION_TAG` and `:latest`.
20. `/Users/jamesonstone/go/src/github.com/lsmc-bio/flowcore/.github/workflows/deploy-ecs.yaml` is intentionally deployment-specific. It uses `workflow_dispatch`, GitHub environments, ECS service/container variables, AWS CLI, `jq`, task definition mutation, and service stability waits. Mint should not generate this class of workflow in the v1 release feature.
21. Conventional Commits 1.0.0 maps `fix` to patch, `feat` to minor, and `BREAKING CHANGE` or `!` to major while allowing other types. The proven resolver's patch fallback for other types is a local policy layered on top of that standard.
22. SemVer 2.0.0 defines major/minor/patch increment meaning and says released version contents must not be modified. That supports the workflow rule to never move an existing SemVer tag.
23. GitHub Actions permissions must be explicit. GHCR publishing likely needs `contents: read`, `packages: write`, and tag creation needs `contents: write`; ECR publishing with OIDC needs `id-token: write` and repo content access for checkout/tagging.
24. GHCR docs and `docker/login-action` support authenticating to `ghcr.io` using `${{ secrets.GITHUB_TOKEN }}` for repository-associated packages. ECR docs and `aws-actions/amazon-ecr-login` support AWS credential configuration followed by ECR Docker login.
25. `docs/PROJECT_PROGRESS_SUMMARY.md` had stale project intent text referring to Kit rather than Mint and a placeholder 0005 feature summary. This supporting doc should be corrected during the brainstorm phase so downstream agents do not inherit stale project context.

## AFFECTED FILES

1. `pkg/release` or equivalent new package: implement release resolution, SemVer parsing, reachable tag discovery, commit bump ranking, release note modeling, registry detection, and workflow data modeling. Recommended default is `pkg/release` because it is release-domain behavior rather than changelog behavior.
2. `pkg/cli/release.go` or equivalent new command file: expose release behavior through Cobra while matching `pkg/cli/changelog.go` style.
3. `pkg/cli/root.go`: update root help/long text only after release computation is implemented; remove the obsolete claim that all release computation/tagging/publishing is future-scoped.
4. `action.yml`: add release-related inputs, outputs, and allowlisted command cases without allowing arbitrary shell commands.
5. `README.md`: update command docs and GitHub Action quick start after implementation; distinguish supported release resolution/publish workflow behavior from still-out-of-scope GitHub Releases, ECS deployment, and package-manager release behavior.
6. `docs/CONSTITUTION.md`: update current implementation and non-goal language after implementation so release computation/publish workflow behavior is no longer described as absent.
7. `docs/specs/0005-pull-forward-v1-features/SPEC.md`: next durable artifact to define binding command names, inputs, outputs, registry behavior, and non-goals.
8. `docs/specs/0005-pull-forward-v1-features/PLAN.md`: later implementation sequencing across package, CLI, action, docs, and tests.
9. `docs/specs/0005-pull-forward-v1-features/TASKS.md`: later binary-verifiable tasks.
10. `pkg/release/*_test.go`: temporary Git repository tests for no-tag, patch, minor, major, non-conventional patch fallback, reachable-tag filtering, already-tagged target, and tag collision behavior.
11. `pkg/cli/*_test.go`: command parsing/output tests after command names are decided.
12. Action/workflow fixture tests: generated GHCR and ECR publish workflows should parse as YAML and contain the required permissions, concurrency, tag-first, and image-tagging steps. They should also assert absence of ECS-specific keys and actions.
13. No product implementation file should be edited during this brainstorm phase.

## DEPENDENCIES

1. Existing Go/Cobra stack from the CLI scaffold remains sufficient for command exposure.
2. The Git CLI remains a runtime/test dependency for release resolution, matching `pkg/changelog` and the source shell scripts.
3. No new external Go dependency is obviously required for SemVer `vX.Y.Z` parsing; a small explicit parser should be enough unless implementation needs pre-release/build metadata support.
4. Generated or documented workflows should continue to use `actions/checkout` with full history/tags because reachable tag discovery and tag creation require complete refs.
5. GHCR workflow support depends on `docker/login-action` with registry `ghcr.io`, `${{ github.actor }}`, and `${{ secrets.GITHUB_TOKEN }}` or an equivalent documented token flow.
6. ECR workflow support depends on `aws-actions/configure-aws-credentials` and `aws-actions/amazon-ecr-login`.
7. Docker image publishing can use either direct `docker buildx build --push` commands, which exactly mirrors the source repos, or `docker/build-push-action`, which matches current Docker action guidance and gives a more declarative generated workflow.
8. GitHub workflow permissions must include `contents: write` for tag creation. GHCR also needs `packages: write`; ECR with OIDC needs `id-token: write`.
9. ECS deploy dependencies such as AWS CLI ECS commands, `jq`, `ECS_CLUSTER`, service/container variables, GitHub environments, and service stability waits are explicitly out of scope.
10. Feature notes are not a dependency because `docs/notes/0005-pull-forward-v1-features` contains only `.gitkeep`.

## QUESTIONS

1. Should the primary command names be `mint release resolve` and `mint release workflow`? Recommended default: yes. `mint release resolve` keeps version computation separate from workflow rendering, while `mint changelog` remains the dedicated changelog command.
2. Should Mint generate workflow YAML or only document workflow snippets? Recommended default: generate workflow YAML to stdout by default and to a file with `--output`, while README documents the generated shape. This turns Mint into reusable tooling instead of another copied documentation block.
3. Should v1 support multiple image specs so Flowcore parity is real? Recommended default: yes. Use repeatable image specs such as `--image name=api,uri=<image-uri>,dockerfile=Dockerfile.api,context=.` and fail validation when any required field is missing.
4. Should `mint release resolve` ever create or push Git tags directly? Recommended default: no. The CLI resolver should be pure and emit `needs_git_tag`; generated workflows should perform tag creation before image publish. This preserves dry-run/testability and keeps mutation inside an explicit workflow step.
5. Should registry detection come only from full image URI hostnames? Recommended default: yes. `ghcr.io/...` resolves to GHCR, AWS ECR hostnames like `<account>.dkr.ecr.<region>.amazonaws.com/...` resolve to ECR, and all other hosts fail with `unsupported registry`.
6. Should first-release fallback start from `v0.0.0` exactly like the source script? Recommended default: yes. This yields `v0.1.0` for first `feat:`, `v0.0.1` for first `fix:` or malformed commit, and `v1.0.0` for first breaking commit.
7. Should generated workflows use direct `docker buildx build --push` shell steps or `docker/build-push-action`? Recommended default: use direct `docker buildx build --push` for strict parity with the proven source repos, unless the SPEC chooses a declarative workflow action for maintainability.
8. Should release notes be lightweight resolver output or reuse `pkg/changelog` rendering? Recommended default: keep lightweight resolver notes for tag annotations and action outputs; keep CHANGELOG.md generation in `mint changelog`.
9. Should GitHub Release creation be included? Recommended default: no. This feature should stop at SemVer Git tags and container image publishing because GitHub Releases were not part of the proven source pattern.
10. Should unsupported registries be documented as a hard failure rather than extensibility points? Recommended default: yes. GHCR and ECR are the only v1 registry kinds until a future feature specifies more.

Current understanding: 90%. The unresolved product choices are command naming, workflow generation shape, multi-image input syntax, and direct buildx versus `docker/build-push-action`.

## OPTIONS

1. Option A: release resolver only.
   - Pros: smallest implementation and easiest to test.
   - Cons: does not satisfy the workflow standardization thesis; teams still copy publish workflow logic.
   - Fit: insufficient for this feature unless workflow generation/documentation is explicitly deferred.
2. Option B: release resolver plus generated publish workflow YAML.
   - Pros: turns the proven source pattern into a reusable Mint deliverable; supports GHCR/ECR branching and multi-image publishing; keeps ECS out of scope.
   - Cons: requires a clear image-spec input contract and workflow fixture tests.
   - Fit: recommended.
3. Option C: release resolver plus action-only publish orchestration.
   - Pros: simplest consumer workflow if `jamesonstone/mint` owns more orchestration.
   - Cons: composite actions cannot fully own caller workflow permissions, checkout depth, concurrency, and tag-push semantics as transparently as generated workflow YAML; harder to make tag-first publishing auditable.
   - Fit: useful as a wrapper around `mint release resolve`, but not enough by itself for publish workflow generation.
4. Option D: include ECS deploy workflow generation.
   - Pros: copies the full Flowcore release/deploy experience.
   - Cons: conflicts with the user thesis and introduces target-specific ECS configuration, AWS CLI mutation, environments, service names, and runtime deployment policy.
   - Fit: rejected for v1.

## RECOMMENDED STRATEGY

1. Add a `pkg/release` package with explicit `Options` and `Result` structs for release resolution. Include `VersionTag`, `VersionBump`, `BaseTag`, `TargetSHA`, `ShortSHA`, `NeedsGitTag`, `CommitCount`, and `ReleaseNotes` fields.
2. Implement release resolution by porting behavior from `scripts/resolve-release-version.sh`, not by shelling out to the script:
   - resolve `commitish^{commit}`;
   - find SemVer tags reachable from the target commit;
   - detect SemVer tags already pointing at the target;
   - walk commits oldest-first from base tag to target;
   - rank bump severity as patch/minor/major;
   - default ambiguous commits to patch;
   - produce deterministic release notes and GitHub-output-compatible fields.
3. Keep tag mutation out of `mint release resolve`. It should report whether a tag is needed; generated workflows should create the annotated tag first and never move existing tags.
4. Add `mint release resolve` as the first release CLI command. Use root command style and error handling from `pkg/cli/changelog.go`.
5. Add `mint release workflow` if workflow generation is confirmed. Default output should be stdout, with `--output` for file writes. The generator should fail closed for unsupported registry hosts and malformed image specs.
6. Support GHCR and ECR by detecting image URI host:
   - `ghcr.io/...` uses GHCR login and GitHub package permissions.
   - `<account>.dkr.ecr.<region>.amazonaws.com/...` uses AWS OIDC credentials and ECR login.
   - anything else fails with an explicit unsupported-registry error.
7. Require the generated workflow to:
   - trigger on default-branch push unless overridden in a future spec;
   - use full checkout history and tags;
   - refresh tags before resolution;
   - run `mint release resolve`;
   - create and push an annotated Git tag before image publish when `needs_git_tag=true`;
   - never move an existing tag;
   - publish every configured image with both `:${{ steps.release.outputs.version_tag }}` and `:latest`;
   - include `concurrency.group: release-publish` and `cancel-in-progress: false`;
   - omit all ECS deploy behavior.
8. Extend `action.yml` with an allowlisted release command and first-class release outputs. The action should remain a CLI wrapper and should not accept arbitrary shell commands.
9. Update README and constitution only after implementation to distinguish implemented release resolution/publish workflow support from out-of-scope GitHub Releases, package-manager releases, and environment deployment.
10. Test with temporary Git repositories and workflow fixtures before implementation is considered complete. Required validation should include `go test ./...`, `go vet ./...`, `make build`, YAML parsing for generated workflows, and `git diff --check`.

## NEXT STEP

Resolve the open questions above, then run `kit spec pull-forward-v1-features` to convert this brainstorm into a binding specification.
