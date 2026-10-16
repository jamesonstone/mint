---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "implement"
feature:
  id: "0016"
  slug: "linear-changelog"
  dir: "0016-linear-changelog"
---
# SPEC

## PURPOSE

Preserve Linear issue references in release changelogs alongside GitHub history.

## CONTEXT

Mint is an offline Git parser, not an issue-content enrichment client. Release bump parsing already accepts arbitrary scopes.

## REQUIREMENTS

Preserve GitHub defaults and history. Use Linear identifiers as scopes and full URLs as commit footers. Never query or publish private issue contents; no live test tickets.

## ACCEPTED PLAN

Parse explicit Linear footer URLs, preserve GitHub links and commit attribution, and document privacy and unavailable-provider behavior.

## DECISIONS

Use existing agent capabilities for ticket writes. No new network client, credential store, or automatic private content enrichment. Two directly coupled PRs require no program ledger.

## DISCOVERIES

Kit has no ticket-creation command; issue payload and lifecycle are canonical agent policy. Mint has no provider API dependency.

## VALIDATION

`go test ./...`, `go vet ./...`, build, changelog help smoke test and 22 Python adapter tests pass. `make docs-check` passes using Python 3.14 and pinned Markdown 3.11 in a disposable environment; HTML/raw Markdown/LLM outputs and links generated and validated. Explicit URL, duplicate URL, mixed GitHub/Linear history, invalid URLs and unavailable-auth parsing tests pass. Lint reports seven existing findings in unchanged files (five errcheck, two unused); no changed-code lint finding. Independent review found no Mint defect. No live provider or release operations ran.

## OUTCOME

Implemented for review; no merge or release authorized. Merge Mint first, then Kit; adopt a Mint version containing Linear-link parsing in consumer workflows before activating Linear. There is no Go API dependency or required dependency version bump between these PRs.

## REPOSITORY MEMORY

Feature spec and public configuration/changelog documentation. Historical specs remain unchanged.

## DELIVERY EVIDENCE

Issue: [GH-32](https://github.com/jamesonstone/mint/issues/32). Branch: `GH-32`.
Ready PR: [Mint #33](https://github.com/jamesonstone/mint/pull/33).
Companion: [Kit #265](https://github.com/jamesonstone/kit/pull/265).
Local validation passed except the seven documented unrelated lint findings.
Hosted documentation build passed; deployment was correctly skipped on the PR.
Hosted native CI was pending when written; the live PR is authoritative.
No merge, release, credential configuration or Linear writes performed.

## PROVIDER PARITY AUDIT

The pre-change CHANGELOG generator reads Git subjects and bodies, derives GitHub
issue URLs, and renders commit links. It does not fetch GitHub ticket contents.
Linear support therefore reads canonical full footer URLs already in Git and
renders identifier links without credentials or a provider API. A new Git-fixture
integration test exercises Generate through the Git reader and file writer,
preserves mixed-provider and commit links, and excludes unrelated commit-body text.
That test and the public documentation checks passed on continuation.

This is reference reading, not API enrichment. Bare Linear IDs cannot resolve a
workspace or URL, and Mint does not verify access, existence, title or status.
Explicit URLs/IDs and commit subjects can expose metadata and need visibility
review. Separate release resolve/publish notes remain subject-based; production
release/proposal control issues remain GitHub-backed. The public guide now states
these boundaries. No provider creation, population or lifecycle API test is claimed.
