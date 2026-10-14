---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "complete"
feature:
  id: "0015"
  slug: "component-promotion"
  dir: "0015-component-promotion"
---
# SPEC

## PURPOSE

Let several independently versioned components of one repository promote
through Mint's schema 2 environment lifecycle without sharing versions, journals,
proposals, or control files. Tracking issue: jamesonstone/mint#28. First consumer:
appliedsymbolics/ambient (Go backend and Expo frontend; program ledger
`docs/programs/mint-adoption/PROGRAM.md` there). Builds on spec 0014 (scoped
release commands, v0.7.0).

## CONTEXT

Research of the lifecycle at `18e8db9` found that two policies could not coexist:
journals are keyed by environment only (`mint-release-state[-env]`), nine sites
validate versions with an unprefixed `^v…$` pattern, `CollectChanges` fails on any
untagged first-parent commit (another component's commits), controllers treat any
`<!-- mint:proposal:` PR as their own, every policy writes `.mint/proposal.json`,
`.mint/summary.md`, and `CHANGELOG.md`, the control workflow uses one concurrency
group, adapters call Mint without `--config`, and publication sends unescaped tags
in `releases/tags/` paths and marks every release Latest.

## REQUIREMENTS

- Optional schema 2 `component: {name, tag_prefix, paths, changelog}`; one policy
  file per component; a deployment component requires its own `control_workflow`
  (package and artifact components run no controller).
- Component versions `<prefix>vX.Y.Z` everywhere (targets, candidates, changes,
  rollback, hotfix, version-main), strictly enforced per journal.
- Isolated journals, local state, control files, proposal markers, workflow names,
  and control concurrency per component.
- Path-scoped `version-main`, `VersionIdentity`, `CollectChanges`, hotfix tags.
- Adapters forward the policy file; manifests accept prefixed versions.
- Policies without `component` keep identical behavior; all existing tests pass.
- Non-goals: several components in one policy file; multi-member bundle manifests
  from the shipped adapter (not needed by per-component builds).

## ACCEPTED PLAN

Add `promotion.Component` (naming, version, control-path, marker, journal,
state-key, concurrency helpers). Store it on `Policy`, `Config`, and `Client`, and
store name/prefix on `State` (new `NewComponentState`/`DecodeComponentState`;
old constructors delegate with the zero component). Replace `versionPattern` with
`strictVersionPattern` + `Component.ValidVersion` and a loose `anyVersionPattern`
for evidence that is later checked against the journal. Thread
`release.Scope` through `VersionMain`, `VersionIdentity`, and
`AllocateHotfixVersion`; add `version-main --config` and Action `version-config`.
Update the adapters, publication, workflows, docs, and tests.

## DECISIONS

- **One policy file per component** rather than a `components:` map in one file:
  each controller, journal, and adapter run already loads one policy, so isolation
  follows from configuration instead of a multi-component engine. Rejected: a
  single file with many components (requires every command to select a component
  and fans out every controller run).
- **Distinct marker namespace** `mint:component-proposal:<name>:` so the
  repository-wide `mint:proposal:` detection never matches component proposals.
- **Component journals always include the environment** (`mint-release-state.<name>.production`);
  only the unscoped production journal keeps the legacy bare name.
- **Strict prefix at the journal**, loose shape at evidence boundaries: adapters and
  `VerifyTag` accept any prefixed version; `validateCandidate`, rollback, policy
  targets, `CollectChanges`, and hotfix provenance require the exact component.
- **`version-main` with no new component commits** reports the newest versioned
  component commit (its own SHA), or nothing when the component has no history.
  For unscoped repositories this equals the previous behavior.
- **Component Releases set `make_latest: false`**; the repository-wide Latest badge
  belongs to the repository-wide line.
- **Independent review (before merge) changed four decisions:** component journals
  use `.` separators (`mint-release-state.<name>.<env>`, local `<name>.<env>`)
  because `-` appears in both name patterns and collided; hotfix base branches
  (`mint-hotfix-base/<name>/GH-N`), source markers, and request markers are
  component-scoped and must carry the journal's own baseline ID;
  `Config.Component` uses `omitzero` so unscoped authority digests and JSON stay
  byte-identical (pinned by a test against `main`'s digest); component
  `version-main` reports `component_changed` instead of failing for commits
  outside its paths.
- **`version-main --config` is opt-in** (Action input `version-config`, default
  empty), so existing callers see no change.

## DISCOVERIES

- Publication used `releases/tags/<version>` without escaping; a slash in a
  component tag would have produced a different API path. Escaped with
  `url.PathEscape`; `git/ref/tags/<ref>` keeps literal slashes (ref path).
- `attributeReviewedRevert` assumes a verified baseline; normal provenance only
  runs after activation, so tests must model one.
- Review findings fixed with regression tests: unscoped `AuthorityDigest` drift
  from a new JSON key (would refuse in-flight intents after upgrade), unscoped
  hotfixes acted on by every component, journal/state name collisions,
  `version-main` failing on commits outside a component, `image_tag` aliases like
  `stable-v1.2.3` newly rejected, other scopes' PRs failing `validate-review`,
  unsafe changelog paths (`.mint/…`, `./x.md`), `--config` dropping
  `.mint/status.json`, and loosened observation versions.
- Second review pass: no blocking issues. Follow-ups applied: component
  environments cannot be named `lock` (journal ref would end in `.lock`); the
  recovery guide documents that hotfix sources are bound to their prepared
  baseline (a behavior change for repository-wide users: a stale hotfix is now
  rejected rather than re-attributed) and one component per hotfix issue.
- Existing `make lint` findings on `main` remain (6, unrelated); this change adds
  none.

## VALIDATION

Local, 2026-10-14:

- PASS: `make test` (Go + 20 adapter tests), `make vet`, `make build`,
  `gofmt -l cmd pkg` (empty), `go test -race ./pkg/...`,
  `golangci-lint run --new-from-rev=origin/main` (0 issues), `make docs-check`,
  `actionlint`, shellcheck of the Action script.
- PASS: new tests `pkg/promotion/component_test.go` (8),
  `pkg/release/version_main_scope_test.go` (3),
  `pkg/cli/release_version_main_scope_test.go` (2), 5 adapter tests; all
  pre-existing tests unchanged and passing.
- PASS: Action script run against a stub CLI: `release-version-main` without and
  with `version-config`.
- PENDING: hosted CI and review on the pull request.

## OUTCOME

Built as planned. Ambient can declare `.mint/backend.yaml` and
`.mint/frontend.yaml` and promote each independently. Merging this PR publishes the
next Mint release through Mint's own `Release` workflow (program step M3).

## REPOSITORY MEMORY

- Created: this SPEC.
- Updated: `website/content/integration.md` (multi-component policies),
  `versioning.md` (component `version-main`), `how-it-works.md` (journal naming),
  `docs/references/environment-lifecycle.md` (component contract),
  `agent-instructions/instructions.md`.
- Unchanged: Constitution (no project-wide invariant changed).
