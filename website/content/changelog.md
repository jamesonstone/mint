# CHANGELOG integration

Mint generates linked Markdown release notes from Git commit history. Good Conventional Commit subjects make both version decisions and changelog entries understandable.

## Generate a release entry

```bash
mint changelog \
  --prev-tag v1.2.3 \
  --current-tag v1.3.0 \
  --owner owner \
  --repo application \
  --output CHANGELOG.md
```

For the first entry, omit `--prev-tag`. When the new version tag does not exist yet, identify the actual range end with `--current-ref`:

```bash
mint changelog \
  --prev-tag v1.2.3 \
  --current-tag v1.3.0 \
  --current-ref HEAD \
  --owner owner \
  --repo application
```

The heading uses the current version and links to its GitHub Release. The date comes from the range-end commit's author date. Entries link to commits. Issue links come from a commit-body closing reference such as `Closes #42`, or a trailing `(#42)` in the subject. A scope such as `GH-42` appears as a scope label; it does not itself create an issue link.

## Linear issues and mixed history

Use the Linear identifier as the conventional commit scope and put its canonical
URL in a separate final footer line. Preserve this footer in squash commits:

```text
feat(ENG-123): :sparkles: add invitations

Refs: https://linear.app/acme/issue/ENG-123/add-invitations
```

Mint renders `[ENG-123](https://linear.app/acme/issue/ENG-123/add-invitations)`
alongside the existing commit link. Accepted full-line references are `Refs:`,
`Issue:`, `Closes:`, `Fixes:`, `Resolves:` (the colon is optional), or a bare
canonical `https://linear.app/<workspace>/issue/<IDENTIFIER>[/slug]` URL.
Repeated identical URLs appear once. A Linear scope alone remains a scope label;
Mint cannot infer its workspace or URL. Keep the full link in commit history.

GitHub closing references and subject `(#42)` suffixes retain their existing
precedence. A commit with both a GitHub reference and a Linear footer preserves
both links. Older GitHub commits need no migration, and release bump detection
already accepts Linear scopes.

Generation is offline: Mint never authenticates to Linear or fetches issue titles,
descriptions, comments, status, or private attachments. Missing credentials,
revoked access, deleted tickets, or a Linear outage do not block generation of an
explicit link. Readers still need their own access; Mint cannot verify that a link
is available. Unrecognized or incomplete URLs do not create an issue link; the
commit link remains. No fallback GitHub issue is invented for a Linear identifier.

Commit messages, ticket identifiers, workspace slugs and URLs become part of the
changelog's visibility. Review these inputs before publishing public notes, and
omit a private reference when even its identifier or URL must remain confidential.
Mint adds no private provider content, but cannot redact confidential text already
written into a commit subject.

## Which commits appear

| Group | Included changes |
| --- | --- |
| breaking changes | Included commit types marked with `!` or a `BREAKING CHANGE:` body. |
| features | Non-breaking `feat`. |
| fixes | Non-breaking `fix` and `hotfix`. |
| perf | Non-breaking `perf`. |
| other | Non-breaking `refactor`, `build`, and `ci`. |

The changelog parser requires a valid Conventional Commit header and warns about unrecognized entries. `docs`, `test`, and `chore` are excluded even if marked breaking; `style` is not a recognized changelog type. The changelog parser recognizes the space form `BREAKING CHANGE:`; the version resolver also recognizes `BREAKING-CHANGE:`. Do not confuse changelog inclusion with version impact: the version resolver still treats other and non-conventional changes as patch changes.

## Per-component changelogs

For a component released with a tag prefix, write its own changelog and limit
entries to its paths:

```bash
mint --prev-tag backend/v1.3.0 --current-tag backend/v1.4.0 --current-ref HEAD \
  --owner acme --repo app --output backend/CHANGELOG.md \
  --tag-prefix backend/ --path backend
```

The header keeps the bare version and links to the prefixed tag:

```text
## [1.4.0](https://github.com/acme/app/releases/tag/backend/v1.4.0) - 2026-10-13
```

Mint rejects a current or previous tag that does not start with the prefix.
In the GitHub Action, use the `tag-prefix` and `paths` inputs with
`command: changelog`.

## Existing changelogs and retries

The command prepends a release block while preserving older entries and file permissions. Existing `##` headings must use Mint's parseable release-heading format. It rejects duplicate versions, malformed release headings, missing range refs, and empty ranges rather than silently overwriting history.

A duplicate-version retry is an error, not an idempotent append. Inspect the existing entry before running again. For compatibility, the root CLI also accepts the changelog flags directly.

## Connect CHANGELOG to the Release PR

In reviewed environment releases, `CHANGELOG.md` is a release-control path alongside `.mint/proposal.json` and `.mint/summary.md`. The controller maintains release metadata while eligible candidates advance. Review the proposed version, included changes, source, complete artifact, and runtime configuration on the current PR head.

Control-only commits are classified separately from application work, avoiding a release loop caused by the generated changelog itself. Keep the configured control paths exact; do not expand them to arbitrary source files.

The standalone changelog command writes a local file; it does not open a PR, publish a GitHub Release, or deploy. `release publish` writes its own resolved release notes and does not automatically commit your CHANGELOG. Choose one owner for changelog updates so competing workflows do not duplicate entries.

## Scope of provider support

The Linear support above reads explicit references already present in Git commit
history. It does not resolve a bare Linear identifier or read a ticket through the
Linear API. There is no title, description, status, comment or assignee enrichment.
This matches the existing GitHub CHANGELOG path, which derives links from commit
references rather than fetching GitHub issue content.

`mint release resolve` and `release publish` produce their separate release notes
from commit subjects; they do not automatically use the CHANGELOG renderer or its
Linear footer links. Mint's release/proposal controller and its control issues
remain GitHub-backed. Selecting a Kit development issue provider does not migrate
those control issues or configure Linear state automation.
