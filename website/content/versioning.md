# Semantic versioning

Mint uses strict release tags in the form `vMAJOR.MINOR.PATCH`, such as `v2.4.1`. Prerelease and build-metadata tags are not strict release tags for Mint's resolver or selection commands.

## How the next version is chosen

The resolver evaluates commits since the highest reachable strict SemVer base tag. The strongest change wins.

| Commit signal | Bump | From `v1.2.3` |
| --- | --- | --- |
| Subject `!`, or body `BREAKING CHANGE:` / `BREAKING-CHANGE:` | Major | `v2.0.0` |
| `feat:` or `feat(scope):` | Minor | `v1.3.0` |
| `fix:`, `hotfix:`, other types, or non-conventional commits | Patch | `v1.2.4` |

Without a base, breaking changes start at `v1.0.0`, features at `v0.1.0`, and other changes at `v0.0.1`. Breaking changes increment major even for a `v0.x.y` base. There is no special pre-1.0 minor-bump exception.

For example, `feat(GH-42): :sparkles: add preview environments` triggers minor; `fix(api)!: remove legacy response fields` triggers major. A gitmoji in the description does not determine the bump.

## Resolve without publishing

```bash
mint release resolve --commitish HEAD
```

Resolution reads Git history; it does not push tags, publish images, or deploy. If the target already has strict SemVer tags, it reuses the highest and reports `already-tagged`.

Fetch full history and tags in CI. Shallow history cannot reliably establish the base or evaluated range.

## Component releases in one repository

A repository can release several components on independent version lines. Give
each component a tag prefix and the paths it owns:

```bash
mint release resolve --tag-prefix backend/ --path backend
mint release resolve --tag-prefix frontend/ --path frontend --path shared/ui
```

- Tags take the form `<prefix>vMAJOR.MINOR.PATCH`, for example `backend/v1.4.0`.
  A prefix contains letters, digits, `.`, `_`, `/`, or `-` and must be safe in a
  Git ref.
- Each scope only sees its own tags. Unprefixed resolution ignores `backend/…`
  tags, and a `backend/` scope ignores `v…` and `frontend/…` tags, so existing
  whole-repository releases keep working beside component releases.
- `--path` is a repeatable Git pathspec. Only commits that touch those paths
  count toward the bump, the commit count, and the release notes. A commit that
  touches two components counts for both.
- `select-tag`, `tag`, `github`, and `publish` accept `--tag-prefix` and reject
  tags outside the scope. `publish` also accepts `--path` and writes the bare
  `X.Y.Z` to the version file.
- In the GitHub Action, set `tag-prefix` and `paths` (one pathspec per line):

```yaml
- uses: jamesonstone/mint@REPLACE_WITH_MINT_COMMIT_SHA
  with:
    command: release-publish
    tag-prefix: backend/
    paths: |
      backend
```

Replace the placeholder with the full SHA of the reviewed Mint release.

To version a component on each default-branch push, pass its prefix and paths
to `version-main`. A first-parent commit that touches those paths can receive
a `<prefix>vX.Y.Z` tag:

```bash
mint release version-main --main-ref origin/main --tag-prefix backend/ --path backend
```

In the action, set `command: release-version-main`, `tag-prefix`, and
`paths` with one path per line. When the target commit did not touch those
paths, `component_changed` is `false` and `version_tag` is empty. A component
build can use those outputs to skip that commit.

## Source version versus deployed release

`mint release version-main` allocates versions for untagged first-parent commits, oldest first. Serialize writers for each tag namespace. For batch versioning, use the current authoritative default branch. Mint advances the patch number to avoid occupied tags.

Use `control_only` to identify commits that only update release metadata. This defaults to `CHANGELOG.md`; the CLI accepts a different list through `--control-paths`. Source tags establish version and commit identity. Build records and deployment receipts establish artifact eligibility and verification.

## Immutable tags

`mint release tag` creates or reuses an annotated tag at an exact source commit. A tag already pointing at that commit is reusable; one pointing elsewhere is a conflict. Mint never moves a conflicting tag.

Inspect the conflicting tag and source identities before retrying. Do not repoint a published version to repair a deployment. Correct or revert source and release a new version, or request a retained-artifact rollback.

## Select an existing version

```bash
mint release select-tag --commitish HEAD
```

Without `--requested-tag`, Mint selects the highest strict SemVer tag on the exact target commit and fails if none exists. With `--requested-tag v1.2.3`, it validates the version syntax and returns that selection. For deployment, resolve the existing tag and require a matching build record and staging receipt.

## Mint version compatibility

Pin an immutable Mint commit and review release notes before changing it.
For controller-dependent workflows, see [Upgrade from Mint v1](release-prs.html).
