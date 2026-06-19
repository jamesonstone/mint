---
kit_metadata_version: 1
artifact: spec
feature:
  id: 0004
  slug: changelog-generation
  dir: 0004-changelog-generation
summary: Generate CHANGELOG.md release entries from conventional commits between Git refs.
relationships:
  - type: depends_on
    feature: 0002-cli-patterns
    reason: Changelog generation is exposed through the Mint Go CLI.
references:
  - id: constitution
    name: Project constitution
    type: repo_doc
    target: docs/CONSTITUTION.md
    relation: constrains
    read_policy: must
    used_for: release-domain scope and implementation evidence wording
    status: active
  - id: cli-patterns
    name: CLI pattern specification
    type: feature_doc
    target: docs/specs/0002-cli-patterns/SPEC.md
    relation: depends_on
    read_policy: must
    used_for: existing CLI structure
    status: active
---
# SPEC

## SUMMARY

Mint must generate an idempotent `CHANGELOG.md` release block from conventional commits in a Git tag range, link related GitHub issues and commits, group entries deterministically, and fail closed for invalid tags, empty ranges, duplicate versions, or unparseable changelog state.

## INPUT CONTRACT

1. `prev_tag`: Git ref string, or empty string for a first release.
2. `current_tag`: Git ref string, required and must exist.
3. `repo_owner`: GitHub username or organization string, required.
4. `repo_name`: GitHub repository name string, required.
5. `output_file`: path to `CHANGELOG.md`, default `./CHANGELOG.md`.

## COMMIT PARSING

Mint must parse commits from `prev_tag..current_tag`. If `prev_tag` is empty, Mint must parse all commits reachable from `current_tag`.

Use this conventional-commits regex against the commit subject:

```text
^(feat|fix|perf|refactor|docs|test|chore|build|ci)(\(.+?\))?!?:\s(.+)$
```

Extract:

1. `type`: captured group 1.
2. `scope`: captured group 2, optional, with parentheses stripped.
3. `breaking`: true when `!` appears before the colon or when the body contains `BREAKING CHANGE:`.
4. `subject`: captured group 3, first line only.
5. `body`: remaining commit body.
6. `hash`: first seven characters of the commit hash.
7. `date`: commit author date.

Non-conventional commits must be logged as warnings and skipped.

## ISSUE LINK EXTRACTION

Extract at most one issue number per commit in this order:

1. Body: scan for `closes #<num>`, `fixes #<num>`, or `resolves #<num>`.
2. Subject: scan for a `(#<num>)` suffix.
3. If found, render `[#<num>](https://github.com/<owner>/<repo>/issues/<num>)`.
4. If not found, omit the issue link.
5. If the issue number is extracted from the subject suffix, strip that suffix from the rendered subject to avoid duplicate issue text.

## GROUPING AND SORTING

Group included commits by type:

1. `breaking changes`: any included commit with `breaking=true`.
2. `features`: `type=feat`.
3. `fixes`: `type=fix`.
4. `perf`: `type=perf`.
5. `other`: `type=refactor`, `type=build`, or `type=ci`.

Exclude `docs`, `test`, and `chore` commits from the changelog output.

Within each group, sort by scope alphabetically, then by commit author date oldest first.

## RENDERING

Render a prepended release block:

```markdown
## [<version>](<tag_url>) - <date_iso>

### breaking changes
- **<scope>:** <subject>
  ([#<issue>](<issue_url>))
  ([<hash>](<commit_url>))

### features
- **<scope>:** <subject>
  ([#<issue>](<issue_url>))
  ([<hash>](<commit_url>))
```

Rules:

1. Version strips a leading `v` from `current_tag` and must match `X.Y.Z`.
2. Tag URL is `https://github.com/<owner>/<repo>/releases/tag/<tag>`.
3. Date is `YYYY-MM-DD` from the `current_tag` target commit author date.
4. Commit URL is `https://github.com/<owner>/<repo>/commit/<hash>`.
5. If scope is empty, omit the `**scope:**` prefix.
6. Omit empty group sections.
7. For a missing or empty changelog file, create a top-level `# Changelog` heading before the first release block.
8. Render issue and commit links as indented continuation lines inside each bullet so generated markdown passes markdownlint style checks.

## FILE HANDLING

1. Read existing `CHANGELOG.md` if present.
2. Extract all existing `## [version]...` headers.
3. If the current version already exists, fail with `version <v> already in CHANGELOG`.
4. If the file is empty or missing, create it.
5. Prepend the new release block with a blank-line separator before existing content.
6. Preserve all existing content below the new block.
7. Write atomically by writing a temp file in the target directory and renaming it into place.

## ACCEPTANCE

1. [AC-01] Every included conventional commit in range appears exactly once.
2. [AC-02] Commits with `type=docs`, `type=test`, and `type=chore` do not appear.
3. [AC-03] Breaking commits appear in the first rendered section.
4. [AC-04] Issue links render as `[#<num>](<url>)`.
5. [AC-05] Commit hashes render as markdown links.
6. [AC-06] Release date renders as `YYYY-MM-DD`.
7. [AC-07] Version header renders as `[X.Y.Z]`.
8. [AC-08] Duplicate versions fail without overwriting the file.
9. [AC-09] Missing tags and empty ranges fail closed with the specified messages.
10. [AC-10] `go test ./...`, `go vet ./...`, and `make build` pass.

## FAIL-CLOSED GUARDRAILS

1. If `prev_tag` does not exist and is not empty, fail with `tag not found: <tag>`.
2. If `current_tag` does not exist, fail with `tag not found: <tag>`.
3. If the commit range produces zero commits, fail with `no commits in range <prev>..<current>`.
4. If `CHANGELOG.md` exists but contains malformed release headers, fail with `cannot parse CHANGELOG.md`.
5. If the version already exists, fail without overwriting.
6. If issue extraction finds no issue number, omit the link.

## OUTPUTS

1. Return the path to the modified `CHANGELOG.md` from the generator API.
2. Print `added <tag> with <N> commits, <M> breaking`.
3. Exit `0` on success and `1` on fail-closed conditions.

## CLI

Mint must support:

```bash
mint \
  --prev-tag v1.0.0 \
  --current-tag v1.1.0 \
  --owner jamesonstone \
  --repo kit \
  --output CHANGELOG.md
```

Mint may also expose the same behavior through `mint changelog` for discoverability.

## OPEN QUESTIONS

none
