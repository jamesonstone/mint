schema: mint.agent_instructions.v1
format: yaml_markdown
audience:
  primary: coding_agents
  secondary: maintainers
purpose:
  summary: "Define the complete implemented Mint functionality map for coding agents."
  agent_goal: "Let an agent decide what Mint can do, why to use it, when to use it, and which command or action path to execute."
  scope: "Current implemented behavior only; future intent is documented only as constraints or vision."

load_order:
  first:
    - path: "README.md"
      reason: "Human and agent entrypoint; links to this manifest."
    - path: "agent-instructions/instructions.md"
      reason: "Comprehensive agent-first Mint functionality manifest."
  skill_usage:
    - path: "agent-instructions/skills.md"
      reason: "Operational code-agent skill for applying Mint in repositories and workflows."
  repo_rules:
    - path: "AGENTS.md"
      reason: "Short routing table for repo-local instructions."
    - path: "docs/agents/README.md"
      reason: "Classify work and load only the needed repo-local docs."
    - path: "docs/CONSTITUTION.md"
      reason: "Canonical project contract when implementation boundaries or project invariants matter."
  feature_scope:
    rule: "When work is feature-scoped, read the feature canonical front matter skills first; fall back to legacy SPEC.md SKILLS only when front matter is absent."

identity:
  name: mint
  repository: "github.com/jamesonstone/mint"
  module: "github.com/jamesonstone/mint"
  binary: "mint"
  product_type:
    - "Go CLI"
    - "Composite GitHub Action"
  one_line: "Compute the next version, write changelogs or release notes, create immutable release tags, and publish GitHub Releases."
  primary_boundary: "Mint owns release state; application repositories own Docker builds, registry publishing, infrastructure, and deployments."

implemented_surface:
  cli:
    entrypoint: "cmd/mint/main.go"
    command_package: "pkg/cli"
    domain_packages:
      - "pkg/changelog"
      - "pkg/release"
    commands:
      - name: "mint --help"
        type: "utility"
        when: "Inspect the command surface."
        why: "Discover available Mint commands without mutation."
        mutates: false
      - name: "mint version"
        type: "utility"
        when: "Print the installed CLI version."
        why: "Verify the binary or action-built CLI version."
        mutates: false
      - name: "mint --version"
        type: "utility"
        when: "Use Cobra-compatible version output."
        why: "Support scripts that expect a version flag."
        mutates: false
      - name: "mint changelog"
        aliases:
          - "mint with root-level changelog flags"
        type: "release_artifact"
        when: "Generate or prepend a CHANGELOG.md release block from conventional commits."
        why: "Create deterministic markdown release history with GitHub issue and commit links."
        mutates: true
        mutation_scope:
          - "Writes the configured changelog output file atomically."
        core_flags:
          - "--prev-tag"
          - "--current-tag"
          - "--current-ref"
          - "--owner"
          - "--repo"
          - "--output"
        output:
          stdout: "added <tag> with <N> commits, <M> breaking"
          files:
            - "CHANGELOG.md by default"
      - name: "mint release resolve"
        type: "release_state"
        when: "Compute the next strict vX.Y.Z release tag from reachable Git history."
        why: "Determine version, bump type, base tag, target SHA, and lightweight release notes without mutating Git."
        mutates: false
        core_flags:
          - "--commitish"
          - "--github-output"
        outputs:
          stdout: "resolved version tag"
          github_output:
            - "version_tag"
            - "version_bump"
            - "base_tag"
            - "target_sha"
            - "short_sha"
            - "needs_git_tag"
            - "commit_count"
            - "release_notes"
      - name: "mint release tag"
        type: "release_state"
        when: "Create or reuse an annotated strict vX.Y.Z Git tag."
        why: "Move Git tag creation out of copied workflow shell and into tested Mint behavior."
        mutates: true
        mutation_scope:
          - "Creates a local annotated tag when absent."
          - "Pushes refs/tags/<tag> when --push is true."
        core_flags:
          - "--tag"
          - "--target"
          - "--notes-file"
          - "--remote"
          - "--push"
          - "--github-output"
        invariants:
          - "Validate tag as strict vX.Y.Z."
          - "Validate target commit exists."
          - "Reuse an existing tag when it points at the same target commit."
          - "Fail when an existing tag points at another commit."
          - "Never move a tag."
          - "Conflicting tag errors include a recovery path."
        outputs:
          stdout:
            created: "created Git tag <tag> <sha>"
            created_pushed: "created and pushed Git tag <tag> <sha>"
            reused: "Git tag <tag> already exists on target commit <sha>"
          github_output:
            - "tag_name"
            - "tag_target_sha"
            - "tag_created"
            - "tag_reused"
            - "tag_pushed"
      - name: "mint release github"
        type: "release_state"
        when: "Create or reuse a GitHub Release after a release tag exists."
        why: "Publish release state through the GitHub API without requiring gh."
        mutates: true
        mutation_scope:
          - "Creates a GitHub Release when one does not already exist for the tag."
        core_flags:
          - "--owner"
          - "--repo"
          - "--tag"
          - "--target"
          - "--title"
          - "--notes-file"
          - "--token-env"
          - "--api-url"
          - "--github-output"
        token_lookup_order:
          default:
            - "MINT_GITHUB_TOKEN"
            - "GITHUB_TOKEN"
            - "GH_TOKEN"
          override: "--token-env"
        outputs:
          stdout:
            created: "created GitHub release <tag> <url>"
            reused: "GitHub release <tag> already exists <url>"
          github_output:
            - "release_tag"
            - "release_url"
            - "release_created"
      - name: "mint release publish"
        type: "release_state"
        when: "Run resolve, tag creation or reuse, tag push, and GitHub Release creation or reuse in one command."
        why: "Give repositories a single release-state operation while keeping Docker and deployment steps project-owned."
        mutates: true
        mutation_scope:
          - "May create and push a Git tag."
          - "May create a GitHub Release."
          - "Writes temporary release notes internally."
        core_flags:
          - "--commitish"
          - "--owner"
          - "--repo"
          - "--title"
          - "--remote"
          - "--push"
          - "--token-env"
          - "--api-url"
          - "--github-output"
        does_not:
          - "Build Docker images."
          - "Authenticate to container registries."
          - "Push containers."
          - "Deploy services."
      - name: "mint release workflow"
        type: "workflow_generation"
        when: "Render a GHCR or ECR publish workflow for an application repository."
        why: "Generate auditable tag-first image publishing workflow YAML while leaving deployments to the app repository."
        mutates: true
        mutation_scope:
          - "Writes workflow YAML when --output is provided."
          - "Prints workflow YAML to stdout when --output is omitted."
        core_flags:
          - "--image"
          - "--output"
          - "--mint-ref"
        image_spec: "name=<name>,uri=<repository-uri>,dockerfile=<path>,context=<path>"
        supported_registries:
          - "GHCR"
          - "AWS ECR"
        invariants:
          - "Image URI must be a repository URI without tag or digest."
          - "One generated workflow cannot mix registry kinds."
          - "Generated workflows delegate release resolution and Git tag creation to the Mint action."
          - "Generated workflows do not include ECS deployment or GitHub Release publishing."
  github_action:
    file: "action.yml"
    type: "composite"
    behavior: "Build cmd/mint from source, add the binary to PATH, and run one allowlisted Mint command."
    security_model:
      - "Allowlisted command switch only."
      - "No arbitrary shell command execution."
      - "No user-provided command strings or eval."
      - "Action shell wraps the CLI; release logic stays in Go."
    supported_commands:
      - "version"
      - "help"
      - "changelog"
      - "release-resolve"
      - "release-tag"
      - "github-release"
      - "release-publish"
      - "none"
    useful_inputs:
      versioning:
        - "go-version"
        - "commitish"
      changelog:
        - "prev-tag"
        - "current-tag"
        - "current-ref"
        - "owner"
        - "repo"
        - "output"
      release_tag:
        - "release-tag"
        - "target-sha"
        - "release-notes-file"
        - "release-remote"
        - "release-push"
      github_release:
        - "github-token"
        - "github-api-url"
        - "release-title"
    outputs:
      general:
        - "mint-path"
        - "output"
      resolve:
        - "version_tag"
        - "version_bump"
        - "base_tag"
        - "target_sha"
        - "short_sha"
        - "needs_git_tag"
        - "commit_count"
        - "release_notes"
      github_release:
        - "release_tag"
        - "release_url"
        - "release_created"
      git_tag:
        - "tag_name"
        - "tag_target_sha"
        - "tag_created"
        - "tag_reused"
        - "tag_pushed"

release_state_boundary:
  mint_owns:
    - "SemVer release resolution from Git history."
    - "Conventional-commit changelog and release-note generation."
    - "Immutable annotated Git tag creation and reuse."
    - "Git tag push when explicitly requested or when action defaults apply."
    - "GitHub Release creation and idempotent reuse."
    - "GitHub Actions output files for release-state fields."
  application_repositories_own:
    - "Docker image builds."
    - "GHCR, ECR, or other registry authentication."
    - "Image publishing beyond generated workflow YAML."
    - "ECS, Kubernetes, VM, PaaS, or service deployment."
    - "Infrastructure updates."
    - "Package-manager publishing."
    - "Release asset uploads."
  decision_rule: "Use Mint up to the release-state boundary, then hand off to repository-specific workflow steps."

domain_contracts:
  changelog:
    parse_contract: "Conventional commits with supported types feat, fix, perf, refactor, docs, test, chore, build, ci."
    include_types:
      - "feat"
      - "fix"
      - "perf"
      - "refactor"
      - "build"
      - "ci"
    exclude_types:
      - "docs"
      - "test"
      - "chore"
    grouping_order:
      - "breaking changes"
      - "features"
      - "fixes"
      - "perf"
      - "other"
    idempotency:
      - "Fail if the version already exists in CHANGELOG.md."
      - "Fail if an existing changelog cannot be parsed."
      - "Prepend new release content and preserve existing content below."
  release_resolution:
    semver: "Strict vX.Y.Z tags only."
    target_default: "HEAD"
    first_release_defaults:
      breaking: "v1.0.0"
      feature: "v0.1.0"
      otherwise: "v0.0.1"
    bump_rules:
      - "Already tagged target returns version_bump=already-tagged and needs_git_tag=false."
      - "Breaking commits produce a major bump."
      - "feat commits produce a minor bump."
      - "fix, other conventional, and non-conventional commits produce a patch bump."
  git_tags:
    strict_tag_pattern: "vX.Y.Z"
    annotated: true
    default_remote: "origin"
    push_default: true
    never_move_tags: true
    same_commit_existing_tag: "success"
    conflicting_existing_tag: "fail closed with recovery path"
  github_releases:
    idempotent_existing_tag_release: true
    api_default: "https://api.github.com"
    token_envs:
      - "MINT_GITHUB_TOKEN"
      - "GITHUB_TOKEN"
      - "GH_TOKEN"

agent_workflows:
  inspect_mint:
    steps:
      - "Run mint --help or go run ./cmd/mint --help."
      - "Read README.md for examples."
      - "Read this manifest for command selection."
  resolve_only:
    when: "A workflow needs version metadata but should not mutate GitHub or Git."
    command: "mint release resolve --commitish HEAD"
    github_action_command: "release-resolve"
    required_checkout: "fetch-depth: 0"
  changelog_only:
    when: "A release block must be generated before tag creation."
    command: "mint changelog --prev-tag <previous> --current-tag <version> --current-ref <sha> --owner <owner> --repo <repo> --output CHANGELOG.md"
    note: "Use --current-ref when the release tag has not been created yet."
  tag_then_github_release:
    when: "A workflow wants explicit control over each release-state step."
    commands:
      - "mint release resolve --commitish HEAD"
      - "mint release tag --tag <version> --target <sha> --notes-file <notes> --push=true"
      - "mint release github --owner <owner> --repo <repo> --tag <version> --target <sha> --notes-file <notes>"
  publish_release_state:
    when: "A workflow wants one Mint command to resolve, tag, push, and create or reuse the GitHub Release."
    command: "mint release publish --owner <owner> --repo <repo> --commitish HEAD"
    github_action_command: "release-publish"
    permissions: "contents: write"
  generated_image_workflow:
    when: "An app repository wants generated GHCR or ECR image publishing YAML."
    command: "mint release workflow --image name=<name>,uri=<uri>,dockerfile=<path>,context=. --output .github/workflows/release-publish.yml"
    boundary: "Generated workflow may publish images, but Mint release-state commands still own tag creation."

development_rules:
  source_of_truth_order:
    - "safety and permission constraints"
    - "current user request"
    - "docs/CONSTITUTION.md"
    - "SPEC.md"
    - "PLAN.md"
    - "TASKS.md"
    - "BRAINSTORM.md"
    - "repo conventions"
  docs:
    - "Keep AGENTS.md, CLAUDE.md, and .github/copilot-instructions.md short routing tables."
    - "Put durable workflow detail under docs/agents or docs/references."
    - "Do not add always-loaded monolithic instruction manuals."
    - "PROJECT_PROGRESS_SUMMARY.md must reflect the highest completed artifact per feature."
  code:
    - "Keep cmd/mint thin."
    - "Keep pkg/cli as Cobra adapter code."
    - "Keep domain logic in pkg/changelog and pkg/release."
    - "Prefer explicit errors and table-driven tests."
    - "Do not add dependencies without a clear runtime or test purpose."
  delivery:
    - "Do not stage, commit, push, create issues, or mutate PRs without explicit user approval."
    - "For GitHub delivery in this Kit-managed repository, load docs/agents/GUARDRAILS.md and relevant docs/references/rules first."

validation:
  docs_only:
    recommended:
      - "git diff --check"
      - "markdownlint when available"
  cli_or_release_changes:
    required:
      - "go test ./..."
      - "go vet ./..."
      - "make build"
      - "go run ./cmd/mint release resolve --commitish HEAD"
      - "go run ./cmd/mint release tag --help"
      - "go run ./cmd/mint release github --help"
      - "go run ./cmd/mint release publish --help"
      - "git diff --check"
  action_changes:
    required:
      - "Parse action.yml or run existing action metadata tests."
      - "Assert command allowlist and output names."

non_goals:
  - "Hosted release service."
  - "Web application."
  - "Arbitrary action command execution."
  - "Docker build execution inside Mint release-state commands."
  - "Registry authentication inside Mint release-state commands."
  - "Deployment orchestration."
  - "Release asset uploads."
  - "Package-manager publishing."
  - "Moving existing Git tags."

evidence_files:
  cli:
    - "cmd/mint/main.go"
    - "pkg/cli/root.go"
    - "pkg/cli/changelog.go"
    - "pkg/cli/release.go"
    - "pkg/cli/release_tag.go"
    - "pkg/cli/release_publish.go"
    - "pkg/cli/version.go"
  domains:
    - "pkg/changelog"
    - "pkg/release/resolve.go"
    - "pkg/release/tag.go"
    - "pkg/release/github_release.go"
    - "pkg/release/workflow.go"
  action:
    - "action.yml"
  docs:
    - "README.md"
    - "docs/CONSTITUTION.md"
    - "docs/agents/README.md"
    - "docs/agents/GUARDRAILS.md"
    - "docs/agents/WORKFLOWS.md"
    - "docs/agents/TOOLING.md"
