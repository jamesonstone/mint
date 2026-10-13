---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "complete"
feature:
  id: "0014"
  slug: "component-scoped-releases"
  dir: "0014-component-scoped-releases"
---
# SPEC

## PURPOSE

Let one repository release several components on independent version lines.
The first consumer is appliedsymbolics/ambient (Go backend container and Expo
frontend web bundle; program ledger `docs/programs/mint-adoption/PROGRAM.md`
there). Tracking issue: jamesonstone/mint#27.

## CONTEXT

Mint assumed one version line: `strictSemVerTagPattern` was `^v…$`
(`pkg/release/semver.go`), tags formatted as `vX.Y.Z`, and release history and
changelog `git log` calls had no pathspecs. Prefixed tags such as
`backend/v1.2.0` were silently ignored. The human chose independent
per-component versioning over one bundled version (recorded in the Ambient
program ledger).

## REQUIREMENTS

- Optional tag prefix and repeatable pathspecs for `release resolve`,
  `select-tag`, `tag`, `github`, `publish`, and `changelog`, in the CLI and the
  GitHub Action (`tag-prefix`, `paths`).
- A scope only sees its own tags; unscoped behavior, outputs, and error text are
  unchanged.
- Path filters limit bumps, commit counts, release notes, and changelog entries.
- Per-component changelog files keep bare `X.Y.Z` headers linking to prefixed tags.
- Non-goals here: `release version-main`, hotfix allocation, schema 2
  `components:`, and per-component promotion (jamesonstone/mint#28).

## ACCEPTED PLAN

Add `release.Scope{TagPrefix, Paths}` with validation, prefixed parse/format,
and pathspec arguments. Thread it through tag discovery, history loading,
resolution, selection, tagging, GitHub Release validation, the version file, and
publishing. Add `TagPrefix`/`Paths` to changelog options. Share CLI flags through
`pkg/cli/release_scope.go`. Add Action inputs and argument arrays. Update the
versioning and changelog guides, agent instructions, and README.

## DECISIONS

- **Prefix isolation by parsing**: a scope parses only `<prefix>vX.Y.Z`, so the
  unscoped `^v…$` pattern never matches component tags and vice versa. Mixed
  repositories keep working without migration.
- **Pathspecs are Git's own** (`git log … -- <paths>`), so globs and magic
  pathspecs work and a commit touching two components counts for both.
- **`version-main` stays unscoped** (zero `Scope`) because it feeds schema 2
  promotion, which becomes component-aware in jamesonstone/mint#28.
- **Unscoped messages are preserved verbatim**, so existing tests and users see
  no change.
- Rejected: a single combined tag namespace with per-component metadata
  (breaks SemVer ordering per component).

## DISCOVERIES

- `gitHubAPIURL` double-escaped path parts (`backend/v1.2.3` became
  `backend%252Fv1.2.3`), so release lookup by a prefixed tag always missed.
  Fixed by setting decoded `Path` and single-escaped `RawPath`; covered by
  `TestPublishGitHubReleaseScopedTag`. Unprefixed tags were never affected.
- The docs validator treats inline `(…/path)` text as links; examples now use
  fenced blocks.
- Pre-existing `make lint` findings on `main` (5 errcheck, 1 unused
  `tagAuthorDate`) remain; this change adds none (`--new-from-rev=origin/main`:
  0 issues). CI does not run lint.

## VALIDATION

Local, 2026-10-13, at the PR head:

- PASS: `make test` (Go and adapter Python tests), `make vet`, `make build`,
  `gofmt -l cmd pkg` (empty), `actionlint`, `go test -race` for release,
  changelog, and cli, `make docs-check` (8 pages, links validated).
- PASS: new tests `pkg/release/scope_test.go`, `pkg/changelog/scope_test.go`,
  `pkg/cli/release_scope_test.go`; `TestActionYAMLSupportsReleaseCommands`
  extended; all existing tests unchanged and passing.
- PASS: the Action run script, extracted from `action.yml` and run against a
  stub CLI: unscoped, scoped, multi-path with blank lines and spaces, and
  `release-version-main` unchanged; shellcheck clean.
- FAIL (pre-existing, unchanged): `make lint` on the full tree (6 issues on
  `main`); 0 new issues.
- PENDING: hosted CI on the pull request.

## OUTCOME

Component-scoped releases work in the CLI and the Action as planned; default
behavior is unchanged. Remaining: schema 2 components and per-component
promotion (jamesonstone/mint#28), then a Mint release (jamesonstone/mint#29)
before Ambient adopts it.

## REPOSITORY MEMORY

- Created: this SPEC.
- Updated: `website/content/versioning.md` (component releases),
  `website/content/changelog.md` (per-component changelogs),
  `agent-instructions/instructions.md` (flags and Action inputs), README.
- Unchanged: Constitution (no project-wide invariant changed).
