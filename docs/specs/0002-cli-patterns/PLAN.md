---
kit_metadata_version: 1
artifact: plan
feature:
  id: 0002
  slug: cli-patterns
  dir: 0002-cli-patterns
summary: Implementation plan for adopting Kit-style CLI and build scaffolding in Mint.
skills:
  - name: rlm
    source: repo-local guide
    path: docs/agents/RLM.md
    trigger: "compare another repository; adopt external repo patterns; broad repository context"
    required: true
relationships:
  - type: depends_on
    feature: 0001-init-project
    reason: The constitution created by init-project defines the source-of-truth and implementation evidence constraints.
references:
  - id: spec
    name: CLI pattern specification
    type: feature_doc
    target: docs/specs/0002-cli-patterns/SPEC.md
    relation: constrains
    read_policy: must
    used_for: binding requirements and acceptance criteria
    status: active
  - id: kit-makefile
    name: Kit Makefile
    type: config
    target: /Users/jamesonstone/go/src/github.com/jamesonstone/kit/Makefile
    relation: informs
    read_policy: must
    used_for: Makefile target names, version linker flag, and bin output pattern
    status: active
  - id: kit-cli
    name: Kit CLI package
    type: repo_doc
    target: /Users/jamesonstone/go/src/github.com/jamesonstone/kit/pkg/cli
    relation: informs
    read_policy: must
    used_for: Cobra root, help rendering, version command, and terminal style patterns
    status: active
---
# PLAN

## APPROACH

1. Add the Go module and dependencies needed for the same CLI style as Kit: Cobra plus terminal-aware styling.
2. Create `cmd/mint/main.go` as a thin binary entrypoint that delegates to `pkg/cli.Execute()`.
3. Create a small `pkg/cli` surface:
   - `root.go` for root command, banner, version variable, and error handling.
   - `root_help.go` for sectioned Kit-style help rendering.
   - `human_output.go` for terminal-aware labels and help templates.
   - `version.go` for `mint version` and build-info fallback.
4. Add focused tests for version resolution and help rendering.
5. Add Kit-style Makefile and pre-commit hook.
6. Rewrite README in Kit's style while keeping release behavior explicitly future-scoped.
7. Update constitution and repo-wide references now that Go CLI/build tooling exists.

## COMPONENTS

1. `cmd/mint/main.go`
   - Binary entrypoint only.
2. `pkg/cli/*`
   - CLI command definitions, human output style, and tests.
3. `Makefile`
   - Build/test/development entrypoints aligned with Kit.
4. `.githooks/pre-commit`
   - Optional repository-managed hook that runs `make build`.
5. `README.md`
   - Kit-style project overview, installation, quick start, commands, and development guidance.
6. `docs/CONSTITUTION.md`, `docs/references/tooling.md`, `docs/references/testing.md`, `docs/PROJECT_PROGRESS_SUMMARY.md`
   - Durable docs updated to match the new implemented facts.

## RISKS

1. Risk: The new CLI could imply release behavior exists.
   Mitigation: Keep commands limited to help/version and document release algorithms as future work.
2. Risk: Copying too much Kit internals could overfit Mint before its release domain is specified.
   Mitigation: Copy only the reusable CLI/build skeleton, not Kit's workflow commands.
3. Risk: Makefile targets could drift from source package names.
   Mitigation: Use `BINARY_NAME=mint`, `./cmd/mint`, and linker path `github.com/jamesonstone/mint/pkg/cli.Version`.

## TESTING

1. `go test ./...`
2. `go vet ./...`
3. `make build`
4. `./bin/mint version`
5. `./bin/mint --help`
6. `git diff --check -- README.md Makefile go.mod go.sum cmd pkg docs .gitignore .githooks`
7. `kit map 0002-cli-patterns`
