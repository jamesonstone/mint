```text
███╗   ███╗██╗███╗   ██╗████████╗
████╗ ████║██║████╗  ██║╚══██╔══╝
██╔████╔██║██║██╔██╗ ██║   ██║
██║╚██╔╝██║██║██║╚██╗██║   ██║
██║ ╚═╝ ██║██║██║ ╚████║   ██║
╚═╝     ╚═╝╚═╝╚═╝  ╚═══╝   ╚═╝
```

**Release Tooling CLI**

🪙 Compute the next version, write the changelog, and mint the release.

Mint is a release tooling CLI. The current implementation provides the same
CLI, README, Makefile, and build patterns used by Kit: a small Go binary under
`cmd/mint`, a reusable `pkg/cli` command package, Cobra-based help/version
handling, conventional-commit CHANGELOG.md generation, SemVer release
resolution, GitHub Release publishing, GHCR/ECR publish workflow generation,
linker-injected versions, repository-level build targets, and a GitHub Action
wrapper that builds and exposes the CLI in workflows.

Mint resolves release metadata and creates GitHub Releases directly. Generated
container workflows own Git tag creation and image publishing. ECS deployment,
package-manager-specific publishing, and registries beyond GHCR/ECR remain out
of scope.

CLI principles:

- 🧰 Kit-style command structure
- 📄 documented behavior before release automation expands
- 🪙 deterministic SemVer release resolution from Git history
- ⚡ small root surface while the domain model is still forming
- 🔍 explicit build and verification commands
- 🔄 version output that works for binaries and module-installed builds
- 📝 deterministic CHANGELOG.md generation from conventional commits
- 🚀 GitHub Release publishing through the GitHub API
- 🏷️ Git-tag-first GHCR/ECR publish workflow generation
- 🧩 public GitHub Action integration that keeps `mint` as the core CLI

## ⚙️ Installation

```bash
go install github.com/jamesonstone/mint/cmd/mint@latest
```

Or build from source:

```bash
git clone https://github.com/jamesonstone/mint.git
cd mint
make build
```

To enable the repository-managed Git hooks for this clone:

```bash
make install-git-hooks
```

This configures `core.hooksPath` to use `.githooks/`, including a `pre-commit`
hook that runs `make build` before every `git commit`. If the build fails, the
commit is blocked.

## 🚀 Quick Start

```bash
# inspect the current command surface
mint --help

# print the installed version
mint version

# equivalent Cobra version flag
mint --version

# prepend a CHANGELOG.md release block from conventional commits
mint \
  --prev-tag v1.0.0 \
  --current-tag v1.1.0 \
  --owner jamesonstone \
  --repo kit \
  --output CHANGELOG.md

# resolve the next SemVer release tag from Git history
mint release resolve --commitish HEAD

# create or reuse a GitHub Release for a resolved tag
GITHUB_TOKEN=... mint release github \
  --owner jamesonstone \
  --repo mint \
  --tag v1.1.0 \
  --target "$(git rev-parse HEAD)" \
  --notes-file CHANGELOG.md

# render a GHCR publish workflow
mint release workflow \
  --image name=api,uri=ghcr.io/jamesonstone/mint-api,dockerfile=Dockerfile.api,context=. \
  --output .github/workflows/release-publish.yml

# build the local binary into bin/mint
make build

# run tests
make test

# run Go vet
make vet
```

## 🧩 GitHub Action Quick Start

Use Mint directly in a GitHub Actions workflow:

```yaml
name: Mint

on:
  workflow_dispatch:

jobs:
  mint:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - name: Run Mint
        id: mint
        uses: jamesonstone/mint@v1
        with:
          command: version

      - name: Show Mint output
        run: echo "${{ steps.mint.outputs.output }}"
```

Replace `v1` with the tag, branch, or commit SHA you publish for the action.
Use a version tag such as `v1` for stable workflows after the action is
released.

The action builds the Go CLI from this repository, adds the generated `mint`
binary to the workflow `PATH`, and then runs one supported command.

Generate a changelog from a workflow after checking out the target repository:

```yaml
steps:
  - uses: actions/checkout@v4
    with:
      fetch-depth: 0

  - name: Generate changelog
    id: mint
    uses: jamesonstone/mint@v1
    with:
      command: changelog
      prev-tag: v1.0.0
      current-tag: v1.1.0
      owner: jamesonstone
      repo: kit
      output: CHANGELOG.md
```

Resolve a release tag from a workflow after checking out full Git history:

```yaml
steps:
  - uses: actions/checkout@v4
    with:
      fetch-depth: 0

  - name: Resolve release
    id: release
    uses: jamesonstone/mint@v1
    with:
      command: release-resolve
      commitish: ${{ github.sha }}

  - name: Show release tag
    run: echo "${{ steps.release.outputs.version_tag }}"
```

Publish a GitHub Release from a workflow with Mint:

```yaml
name: Release

on:
  push:
    branches:
      - main
  workflow_dispatch:

permissions:
  contents: write

jobs:
  github-release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Resolve release
        id: release
        uses: jamesonstone/mint@v1
        with:
          command: release-resolve
          commitish: ${{ github.sha }}

      - name: Write release notes
        id: notes
        shell: bash
        env:
          RELEASE_NOTES: ${{ steps.release.outputs.release_notes }}
        run: |
          notes_file="$RUNNER_TEMP/release-notes.md"
          printf '%s\n' "$RELEASE_NOTES" > "$notes_file"
          echo "path=$notes_file" >> "$GITHUB_OUTPUT"

      - name: Publish GitHub Release
        id: github_release
        uses: jamesonstone/mint@v1
        with:
          command: github-release
          owner: ${{ github.repository_owner }}
          repo: ${{ github.event.repository.name }}
          release-tag: ${{ steps.release.outputs.version_tag }}
          target-sha: ${{ steps.release.outputs.target_sha }}
          release-title: ${{ steps.release.outputs.version_tag }}
          release-notes-file: ${{ steps.notes.outputs.path }}
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

The workflow needs `contents: write` so the repository token can create the
release and create the tag when the tag does not already exist.

Supported action inputs:

| Input                | Default                  | Description                                                                        |
| -------------------- | ------------------------ | ---------------------------------------------------------------------------------- |
| `command`            | `version`                | Run `version`, `help`, `changelog`, `release-resolve`, `github-release`, or `none` |
| `go-version`         | `1.25.5`                 | Go version used to build the CLI on the runner                                     |
| `prev-tag`           | empty                    | Previous Git tag or ref; empty for first changelog release                         |
| `current-tag`        | empty                    | Current Git tag or ref for changelog generation                                    |
| `owner`              | empty                    | GitHub repository owner for changelog links and GitHub Releases                    |
| `repo`               | empty                    | GitHub repository name for changelog links and GitHub Releases                     |
| `output`             | `CHANGELOG.md`           | Changelog file path for `command: changelog`                                       |
| `commitish`          | `HEAD`                   | Git ref for `command: release-resolve`                                             |
| `release-tag`        | empty                    | Strict `vX.Y.Z` SemVer tag for `command: github-release`                           |
| `target-sha`         | empty                    | Commitish where GitHub should create the tag when missing                          |
| `release-title`      | empty                    | Release title; defaults to `release-tag`                                           |
| `release-notes-file` | empty                    | Optional release notes file for `command: github-release`                          |
| `github-token`       | empty                    | GitHub token for `command: github-release`                                         |
| `github-api-url`     | `https://api.github.com` | GitHub API base URL                                                                |

Supported action outputs:

| Output            | Description                                             |
| ----------------- | ------------------------------------------------------- |
| `mint-path`       | Absolute path to the built `mint` binary                |
| `output`          | Captured stdout from the selected command               |
| `version_tag`     | Resolved strict SemVer release tag                      |
| `version_bump`    | `already-tagged`, `patch`, `minor`, or `major`          |
| `base_tag`        | Reachable base SemVer tag, if one exists                |
| `target_sha`      | Full target commit SHA                                  |
| `short_sha`       | Twelve-character target commit SHA                      |
| `needs_git_tag`   | Whether a generated workflow should create a Git tag    |
| `commit_count`    | Number of commits evaluated for release resolution      |
| `release_notes`   | Lightweight tag annotation notes for generated tags     |
| `release_tag`     | GitHub Release tag from `command: github-release`       |
| `release_url`     | GitHub Release URL from `command: github-release`       |
| `release_created` | Whether `command: github-release` created a new release |

To install Mint into the workflow `PATH` without running it immediately:

```yaml
steps:
  - name: Set up Mint
    uses: jamesonstone/mint@v1
    with:
      command: none

  - name: Use the Mint CLI
    run: mint --help
```

## 🧰 Commands

### 🪙 Release

| Command                 | Description                                     |
| ----------------------- | ----------------------------------------------- |
| `mint release`          | Resolve releases and render publish workflows   |
| `mint release resolve`  | Compute a strict `vX.Y.Z` release tag           |
| `mint release github`   | Create or reuse a GitHub Release                |
| `mint release workflow` | Generate a GHCR or ECR publish workflow         |
| `mint changelog`        | Generate CHANGELOG.md from conventional commits |

Resolve a release tag:

```bash
mint release resolve --commitish HEAD
```

The resolver prints the resolved `version_tag` and can write all release fields
in GitHub Actions output format:

```bash
mint release resolve \
  --commitish HEAD \
  --github-output "$GITHUB_OUTPUT"
```

Create or reuse a GitHub Release:

```bash
mint release github \
  --owner jamesonstone \
  --repo mint \
  --tag v1.1.0 \
  --target "$(git rev-parse HEAD)" \
  --notes-file CHANGELOG.md
```

`mint release github` reads a token from `MINT_GITHUB_TOKEN`, `GITHUB_TOKEN`, or
`GH_TOKEN`, unless `--token-env` points at a different environment variable. It
does not require the GitHub CLI.

Generate a GHCR publish workflow:

```bash
mint release workflow \
  --image name=api,uri=ghcr.io/jamesonstone/mint-api,dockerfile=Dockerfile.api,context=. \
  --output .github/workflows/release-publish.yml
```

Generate an ECR publish workflow:

```bash
mint release workflow \
  --image name=api,uri=123456789012.dkr.ecr.us-east-1.amazonaws.com/mint-api,dockerfile=Dockerfile.api,context=. \
  --output .github/workflows/release-publish.yml
```

Generated workflows run on `push`, resolve the release with the Mint action,
create and push an annotated SemVer Git tag before image publishing when needed,
and publish each image with both the resolved version tag and `latest`.

Image URIs must be repository URIs without tags. All images in one generated
workflow must use the same supported registry kind: GHCR or AWS ECR.

The root command also accepts the changelog flags directly for script-friendly
usage:

```bash
mint --prev-tag v1.0.0 --current-tag v1.1.0 --owner jamesonstone --repo kit
```

### 🔧 Utilities

| Command           | Description                       |
| ----------------- | --------------------------------- |
| `mint version`    | Print the installed Mint version  |
| `mint completion` | Generate shell autocompletion     |
| `mint help`       | Help about Mint or a Mint command |

Mint does not deploy ECS services, publish package-manager artifacts, or
support registries beyond GHCR/ECR in this release workflow feature.

## 🛠️ Development

| Target                   | Description                                     |
| ------------------------ | ----------------------------------------------- |
| `make build`             | Build `bin/mint` with linker-injected version   |
| `make build-windows`     | Build `bin/mint.exe` for Windows amd64          |
| `make install`           | Install `mint` with linker-injected version     |
| `make install-git-hooks` | Use `.githooks/` as this clone's Git hooks path |
| `make test`              | Run `go test -v ./...`                          |
| `make lint`              | Run `golangci-lint run ./...`                   |
| `make fmt`               | Run `go fmt ./...`                              |
| `make vet`               | Run `go vet ./...`                              |
| `make tidy`              | Run `go mod tidy`                               |
| `make clean`             | Remove local build output and run `go clean`    |
| `make all`               | Run `fmt`, `vet`, `test`, and `build`           |

The Makefile mirrors Kit's local build pattern. Additional release-domain
behavior should be added through feature specs before product commands are
implemented.
