```text
███╗   ███╗██╗███╗   ██╗████████╗
████╗ ████║██║████╗  ██║╚══██╔══╝
██╔████╔██║██║██╔██╗ ██║   ██║
██║╚██╔╝██║██║██║╚██╗██║   ██║
██║ ╚═╝ ██║██║██║ ╚████║   ██║
╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝   ╚═╝
```

**Release Tooling CLI**

<!-- BEGIN KIT-MANAGED README BADGES -->
[![Last commit](https://img.shields.io/github/last-commit/jamesonstone/mint)](https://github.com/jamesonstone/mint/commits) [![Open issues](https://img.shields.io/github/issues/jamesonstone/mint)](https://github.com/jamesonstone/mint/issues) [![Pull requests](https://img.shields.io/github/issues-pr/jamesonstone/mint)](https://github.com/jamesonstone/mint/pulls) [![CI](https://github.com/jamesonstone/mint/actions/workflows/ci.yaml/badge.svg)](https://github.com/jamesonstone/mint/actions/workflows/ci.yaml) [![Release](https://img.shields.io/github/v/release/jamesonstone/mint)](https://github.com/jamesonstone/mint/releases)
<!-- END KIT-MANAGED README BADGES -->

Compute the next version, write the changelog, and mint the release.

Mint is a Go CLI and GitHub Action for release state. Use it to resolve
Semantic Versions from Git history, render a changelog, create immutable Git
tags, and publish GitHub Releases. Each component in a monorepo can have its own
tag prefix and source paths.

## Use Mint in GitHub Actions

Use the action from a repository-owned workflow. This example publishes
release state for an exact source commit. Replace
`REPLACE_WITH_MINT_COMMIT_SHA` with the full SHA of the reviewed Mint release.

```yaml
permissions:
  contents: write
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@11d5960a326750d5838078e36cf38b85af677262
        with:
          fetch-depth: 0
      - uses: jamesonstone/mint@REPLACE_WITH_MINT_COMMIT_SHA
        with:
          command: release-publish
          commitish: ${{ github.sha }}
          owner: ${{ github.repository_owner }}
          repo: ${{ github.event.repository.name }}
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

Run this job from a trusted branch or exact source ref. The action builds Mint,
adds it to the runner path, and exposes command outputs. Use
`command: none` to prepare Mint for later steps without invoking a command.

## Promote the same artifact

Your repository's workflows own deployment and verification.

1. After trusted default-branch checks pass, version the source and build it
   once. Retain the artifact by digest with the project, source SHA, SemVer, and
   producer run.
2. Deploy the digest to staging automatically. Record the verified digest and
   staging results.
3. Let an authorized operator dispatch production for an exact project and
   SemVer. Require successful staging evidence for the same artifact. Call the
   shared promotion job with the retained digest. Production uses the build
   from the first step.
4. After production verification, use `mint release github` to publish the
   selected tag. Your workflow owns changelog updates.

See the [environment promotion guide](https://jamesonstone.github.io/mint/environments.html).

## Local CLI

Mint requires Go 1.26.9 or newer. Set `MINT_VERSION` to a published v2 release
tag, then install the CLI:

```bash
go install "github.com/jamesonstone/mint/v2/cmd/mint@${MINT_VERSION:?Set MINT_VERSION to a published v2 tag}"
```

To build from a source checkout, run `make build`, then use `./bin/mint`.

| Command | Purpose |
| --- | --- |
| `mint release resolve` | Resolve the version for a source commit. |
| `mint release select-tag` | Select an existing version tag on a commit or validate a requested version's syntax. |
| `mint release version-main` | Version untagged first-parent commits. |
| `mint release tag` | Create or reuse an immutable annotated Git tag. |
| `mint release github` | Create or reuse a GitHub Release for a selected tag. |
| `mint release publish` | Resolve, write `.version`, tag, and publish a GitHub Release. |
| `mint release workflow` | Generate an image publishing workflow for GHCR or ECR. |
| `mint changelog` | Write a changelog entry from Conventional Commits. |
| `mint version` | Print the installed version. |

Run `mint <command> --help` for options and examples.

## Documentation

- [Mint documentation](https://jamesonstone.github.io/mint/), [llms.txt](https://jamesonstone.github.io/mint/llms.txt), and [local site development](website/README.md)
- [Agent command reference](agent-instructions/instructions.md) and [agent skill](agent-instructions/skills.md)
- [Upgrade from Mint v1](https://jamesonstone.github.io/mint/release-prs.html). Legacy controller users can pin [v1.1.3](https://github.com/jamesonstone/mint/releases/tag/v1.1.3) while migrating.

## Development

Run `go test ./...`, `go vet ./...`, and `make build` to validate changes.
Use `make install-git-hooks` to enable the repository build hook.
See the [testing reference](docs/references/testing.md) for more checks.

## License

Mint is licensed under [Apache-2.0](LICENSE). See [NOTICE](NOTICE) for attribution.
Bundled JetBrains Mono fonts retain their [SIL Open Font License](website/assets/fonts/OFL.txt).

## Maintainers

Maintained by [jamesonstone](https://github.com/jamesonstone).
