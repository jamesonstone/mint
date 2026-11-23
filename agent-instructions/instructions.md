# Mint command reference for coding agents

Mint is a Go CLI and composite GitHub Action for release state. The repository
is `jamesonstone/mint` and the Go module is `github.com/jamesonstone/mint/v2`.
Use the [public guides](https://jamesonstone.github.io/mint/) for integration
and [tooling reference](../docs/references/tooling.md) for command details.

## Select a command

| Need | Command | Effect |
| --- | --- | --- |
| Compute a version for a commit | `mint release resolve --commitish <ref>` | Reads Git history and reports release metadata. |
| Select a tag on a commit | `mint release select-tag --commitish <ref>` | Returns the highest strict SemVer tag on that commit. |
| Validate explicit version syntax | `mint release select-tag --requested-tag <tag>` | Returns a valid requested version. Resolve its ref and artifact separately. |
| Version a first-parent history | `mint release version-main --main-ref <ref>` | Allocates immutable source tags. Serialize writers per namespace. |
| Render linked changelog notes | `mint changelog --prev-tag <tag> --current-tag <tag> --owner <owner> --repo <repo>` | Writes a changelog in the workspace. Omit the previous tag for a first release. |
| Create an annotated source tag | `mint release tag --tag <tag> --target <sha> --notes-file <path>` | Creates or reuses the exact tag. Use `--push=false` for a local-only tag. |
| Publish a selected version | `mint release github --tag <tag> --target <sha> --owner <owner> --repo <repo>` | Creates or reuses the GitHub Release. Accepts an optional notes file. |
| Publish directly from source | `mint release publish --commitish <ref> --owner <owner> --repo <repo>` | Resolves the version, writes `.version`, tags, and publishes a GitHub Release. |
| Generate an image publishing workflow | `mint release workflow --image name=<name>,uri=<uri>,dockerfile=<path>,context=<path>` | Renders Actions YAML for GHCR or ECR. |

Fetch full history and tags before running Git-based release commands. Use
`--tag-prefix` for an independent tag namespace, and pass `--path` for each
component path supported by the command. For `version-main`, a scoped build
can use `component_changed` to skip a commit that did not affect the component.

## GitHub Action

Use `jamesonstone/mint@<immutable-sha>` with one of these `command` inputs:

- `version`, `help`, `none`
- `changelog`
- `release-resolve`, `release-select-tag`, `release-version-main`
- `release-tag`, `github-release`, `release-publish`

The action supplies `mint-path`, captured `output`, and structured release
outputs. Use `tag-prefix` and newline-separated `paths` for component scope
where supported. Pin the exact Action version before choosing input fields.
Give read-only jobs `contents: read` and grant write permission to the trusted
jobs that create tags or publish GitHub Releases.

## Environment promotion

Build once after trusted checks. Record the artifact digest, source SHA,
project, version, and build provenance. Use a shared repository-owned
promotion job for continuous staging and authorized manual production.

Require successful staging evidence for the exact retained production
artifact. Verify the digest in production before recording success or
publishing. Use `release github` for the selected tag and SHA after verification.
The repository owns the provider operations, permission checks, and recovery.

For v1 controller dependencies, pin [v1.1.3](https://github.com/jamesonstone/mint/releases/tag/v1.1.3)
and use its [versioned lifecycle guide](https://github.com/jamesonstone/mint/blob/v1.1.3/docs/references/environment-lifecycle.md).
See [v2 migration](https://jamesonstone.github.io/mint/release-prs.html).
