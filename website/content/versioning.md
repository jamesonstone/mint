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

## Source version versus deployed release

`mint release version-main` serially allocates one version for every untagged first-parent default-branch commit, oldest first. Controllers must use current authoritative main rather than an old triggering event SHA. Occupied tags, including isolated hotfix tags, are skipped by advancing patch until a free version is found.

Trusted release-control-only commits also receive a source identity, but do not become application build candidates or add noise to application release notes. Source versioning does not create a GitHub Release or prove artifact eligibility. A selected environment release may include several source versions since its last verified deployment.

## Immutable tags

`mint release tag` creates or reuses an annotated tag at an exact source commit. A tag already pointing at that commit is reusable; one pointing elsewhere is a conflict. Mint never moves a conflicting tag.

Inspect the conflicting tag and source identities before retrying. Do not repoint a published version to repair a deployment. Correct or revert source and release a new version, or request a retained-artifact rollback.

## Select an existing version

```bash
mint release select-tag --commitish HEAD
```

Without `--requested-tag`, Mint selects the highest strict SemVer tag on the exact target commit and fails if none exists. With `--requested-tag v1.2.3`, it validates the version syntax and returns that selection; this alone does not prove the requested artifact exists, matches the commit, or was deployed. Environment requests additionally require authenticated candidate/history evidence.

## Hotfix version series

An isolated hotfix allocates the next unused patch in the verified deployed version's series. If that environment is on `v1.2.3` while ordinary work has reached `v1.3.0`, the fix remains in `v1.2.x`; occupied patches are skipped. Merge the correction back into ordinary source so later releases keep it.
