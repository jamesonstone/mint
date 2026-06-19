---
kit_metadata_version: 1
artifact: plan
feature:
  id: 0004
  slug: changelog-generation
  dir: 0004-changelog-generation
summary: Implementation plan for conventional-commit CHANGELOG.md generation.
relationships:
  - type: depends_on
    feature: 0002-cli-patterns
    reason: Changelog generation is implemented as Mint CLI behavior.
references:
  - id: spec
    name: Changelog generation specification
    type: feature_doc
    target: docs/specs/0004-changelog-generation/SPEC.md
    relation: constrains
    read_policy: must
    used_for: binding behavior, errors, rendering, and acceptance
    status: active
  - id: cli-root
    name: Mint CLI root
    type: source
    target: pkg/cli/root.go
    relation: updates
    read_policy: must
    used_for: root flag routing and command registration
    status: active
---
# PLAN

## APPROACH

1. Add a small `pkg/changelog` package for testable generation behavior: git collection, conventional commit parsing, issue extraction, grouping, rendering, and atomic file prepending.
2. Add `mint changelog` with the required flags, and route root-level `mint --prev-tag ... --current-tag ... --owner ... --repo ... --output ...` to the same implementation to satisfy the requested invocation.
3. Keep raw Git work behind `git` command invocations and make tag/range errors explicit.
4. Add tests using temporary Git repositories with fixture commits and tags.
5. Update help, README, durable docs, and feature progress now that changelog generation exists.

## COMPONENTS

1. `pkg/changelog`
   - Public generator API and focused internal helpers.
2. `pkg/cli/changelog.go`
   - Cobra command, flags, root invocation routing, output summary.
3. `README.md`
   - CLI quick start and command table entry.
4. `action.yml`
   - Optional action command support for changelog generation.
5. `docs/CONSTITUTION.md`, `docs/references/tooling.md`, `docs/references/testing.md`, `docs/PROJECT_PROGRESS_SUMMARY.md`
   - Durable implementation facts and verification guidance.

## RISKS

1. Risk: A non-conventional commit could make generation unusable.
   Mitigation: Warn and skip non-conventional commits; fail only for guardrail conditions.
2. Risk: Existing changelog content could be corrupted.
   Mitigation: Parse release headers before writing and use atomic temp-file rename.
3. Risk: Root-level flags could make help noisy.
   Mitigation: Support them for the requested invocation while keeping `mint changelog` discoverable in help.
4. Risk: Shell-based tests could be brittle.
   Mitigation: Use Go tests with temporary Git repositories and direct command execution.

## TESTING

1. `go test ./...`
2. `go vet ./...`
3. `make build`
4. `./bin/mint --help`
5. `./bin/mint changelog --help`
6. `git diff --check -- README.md action.yml docs pkg`
7. `kit map 0004-changelog-generation`
