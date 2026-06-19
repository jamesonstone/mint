---
kit_metadata_version: 1
artifact: spec
feature:
  id: 0002
  slug: cli-patterns
  dir: 0002-cli-patterns
summary: Adopt Kit-style CLI, README, Makefile, and build patterns for Mint without implementing release algorithms yet.
skills:
  - name: rlm
    source: repo-local guide
    path: docs/agents/RLM.md
    trigger: "compare another repository; adopt external repo patterns; broad repository context"
    required: true
relationships:
  - type: depends_on
    feature: 0001-init-project
    reason: The completed constitution defines the documentation-first boundary and fact-versus-intent rules for adding Mint's first runtime scaffold.
references:
  - id: constitution
    name: Project constitution
    type: repo_doc
    target: docs/CONSTITUTION.md
    relation: constrains
    read_policy: must
    used_for: source-of-truth order, spec-before-product behavior, and fact-versus-intent wording
    status: active
  - id: kit-readme
    name: Kit README
    type: repo_doc
    target: /Users/jamesonstone/go/src/github.com/jamesonstone/kit/README.md
    relation: informs
    read_policy: must
    used_for: README structure, command table style, quick-start shape, and build instructions
    status: active
  - id: kit-makefile
    name: Kit Makefile
    type: config
    target: /Users/jamesonstone/go/src/github.com/jamesonstone/kit/Makefile
    relation: informs
    read_policy: must
    used_for: build, install, test, lint, fmt, vet, clean, tidy, and git-hook targets
    status: active
  - id: kit-cli-entrypoint
    name: Kit CLI entrypoint
    type: repo_doc
    target: /Users/jamesonstone/go/src/github.com/jamesonstone/kit/cmd/kit/main.go
    relation: informs
    read_policy: must
    used_for: cmd package delegates to pkg/cli Execute pattern
    status: active
  - id: kit-root-cli
    name: Kit root CLI package
    type: repo_doc
    target: /Users/jamesonstone/go/src/github.com/jamesonstone/kit/pkg/cli/root.go
    relation: informs
    read_policy: must
    used_for: Cobra root command, version variable, custom version template, and Execute error handling pattern
    status: active
---
# SPEC

## SUMMARY

Mint should adopt Kit's Go CLI project shape: a `cmd/mint` binary entrypoint, reusable `pkg/cli` command package, Cobra-based root command, version command, Kit-style help/readme presentation, Makefile build targets, and a repository-managed pre-commit build hook.

## PROBLEM

Mint now has a completed constitution and Kit scaffold, but no source code, Go module, Makefile, README structure, or build/test entrypoints. Future release-tooling work needs the same lightweight operator ergonomics Kit uses before release behavior is specified.

## GOALS

1. Add a minimal Go CLI scaffold that follows Kit's `cmd/<binary>` plus `pkg/cli` pattern.
2. Add Kit-style version handling with linker-injected `Version`, `mint version`, and `mint --version`.
3. Add a Kit-style Makefile with `build`, `build-windows`, `install`, `install-git-hooks`, `test`, `lint`, `fmt`, `vet`, `clean`, `tidy`, and `all`.
4. Add a `.githooks/pre-commit` hook that runs `make build`.
5. Rewrite `README.md` in Kit's presentation style while accurately describing Mint's current implemented surface.
6. Update durable project docs so the constitution, progress summary, tooling reference, and testing reference reflect the new runtime/build facts.

## NON-GOALS

1. Do not implement version calculation, changelog generation, tag creation, publishing, GitHub release behavior, or package-manager-specific release logic.
2. Do not add external services, CI workflows, Dockerfiles, or release automation.
3. Do not broaden the command surface beyond root help, version reporting, and build/test tooling.
4. Do not change Kit-managed baseline rules or generated local `.kit` state.

## REQUIREMENTS

1. [SPEC-01] The repository must contain a Go module for `github.com/jamesonstone/mint`.
2. [SPEC-02] The CLI entrypoint must live at `cmd/mint/main.go` and delegate to `pkg/cli.Execute()`.
3. [SPEC-03] The root command must use Cobra and expose Kit-style root help with a banner, sections, custom help templates, and terminal-aware styling.
4. [SPEC-04] The CLI must expose `mint version` and `mint --version`, preferring linker-injected versions and falling back to Go build info or `dev`.
5. [SPEC-05] The Makefile must mirror Kit's build/install/test/lint/fmt/vet/clean/tidy/all pattern and inject `pkg/cli.Version`.
6. [SPEC-06] The README must use Kit-style visual structure, installation/build instructions, quick start, command table, and development section.
7. [SPEC-07] Stable build and test commands must be recorded in repo-wide references.
8. [SPEC-08] Documentation must distinguish implemented CLI/build scaffold from future unimplemented release behavior.

## ACCEPTANCE

1. [AC-01] `go test ./...` passes.
2. [AC-02] `go vet ./...` passes.
3. [AC-03] `make build` creates `bin/mint`.
4. [AC-04] `./bin/mint version` prints a non-empty version string.
5. [AC-05] `./bin/mint --help` renders a Kit-style help surface.
6. [AC-06] `git diff --check` reports no whitespace errors for touched files.
7. [AC-07] `kit map 0002-cli-patterns` resolves the feature.
8. [AC-08] No release algorithm or publishing behavior is introduced.

## OPEN QUESTIONS

none
