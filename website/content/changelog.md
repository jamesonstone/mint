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

## Which commits appear

| Group | Included changes |
| --- | --- |
| breaking changes | Included commit types marked with `!` or a `BREAKING CHANGE:` body. |
| features | Non-breaking `feat`. |
| fixes | Non-breaking `fix` and `hotfix`. |
| perf | Non-breaking `perf`. |
| other | Non-breaking `refactor`, `build`, and `ci`. |

The changelog parser requires a valid Conventional Commit header and warns about unrecognized entries. `docs`, `test`, and `chore` are excluded even if marked breaking; `style` is not a recognized changelog type. The changelog parser recognizes the space form `BREAKING CHANGE:`; the version resolver also recognizes `BREAKING-CHANGE:`. Do not confuse changelog inclusion with version impact: the version resolver still treats other and non-conventional changes as patch changes.

## Existing changelogs and retries

The command prepends a release block while preserving older entries and file permissions. Existing `##` headings must use Mint's parseable release-heading format. It rejects duplicate versions, malformed release headings, missing range refs, and empty ranges rather than silently overwriting history.

A duplicate-version retry is an error, not an idempotent append. Inspect the existing entry before running again. For compatibility, the root CLI also accepts the changelog flags directly.

## Connect CHANGELOG to the Release PR

In reviewed environment releases, `CHANGELOG.md` is a release-control path alongside `.mint/proposal.json` and `.mint/summary.md`. The controller maintains release metadata while eligible candidates advance. Review the proposed version, included changes, source, complete artifact, and runtime configuration on the current PR head.

Control-only commits are classified separately from application work, avoiding a release loop caused by the generated changelog itself. Keep the configured control paths exact; do not expand them to arbitrary source files.

The standalone changelog command writes a local file; it does not open a PR, publish a GitHub Release, or deploy. `release publish` writes its own resolved release notes and does not automatically commit your CHANGELOG. Choose one owner for changelog updates so competing workflows do not duplicate entries.
