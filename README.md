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

Mint is a release tooling CLI. The current implementation establishes the same
CLI, README, Makefile, and build patterns used by Kit: a small Go binary under
`cmd/mint`, a reusable `pkg/cli` command package, Cobra-based help/version
handling, conventional-commit CHANGELOG.md generation, linker-injected
versions, repository-level build targets, and a GitHub Action wrapper that
builds and exposes the CLI in workflows.

Release computation, tagging, publishing, GitHub releases, and
package-manager-specific behavior are intentionally future-scoped until their
contracts are specified.

CLI principles:

- 🧰 Kit-style command structure
- 📄 documented behavior before release automation
- 🪙 release intent without invented release semantics
- ⚡ small root surface while the domain model is still forming
- 🔍 explicit build and verification commands
- 🔄 version output that works for binaries and module-installed builds
- 📝 deterministic CHANGELOG.md generation from conventional commits
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

Supported action inputs:

| Input         | Default        | Description                                             |
| ------------- | -------------- | ------------------------------------------------------- |
| `command`     | `version`      | Run `version`, `help`, `changelog`, or `none`           |
| `go-version`  | `1.25.5`       | Go version used to build the CLI on the runner          |
| `prev-tag`    | empty          | Previous Git tag or ref; empty for first release        |
| `current-tag` | empty          | Current Git tag or ref for changelog generation         |
| `owner`       | empty          | GitHub repository owner for changelog links             |
| `repo`        | empty          | GitHub repository name for changelog links              |
| `output`      | `CHANGELOG.md` | Changelog file path for `command: changelog`            |

Supported action outputs:

| Output      | Description                              |
| ----------- | ---------------------------------------- |
| `mint-path` | Absolute path to the built `mint` binary |
| `output`    | Captured stdout from the selected command |

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

| Command          | Description                                    |
| ---------------- | ---------------------------------------------- |
| `mint changelog` | Generate CHANGELOG.md from conventional commits |

The root command also accepts the changelog flags directly for script-friendly
usage:

```bash
mint --prev-tag v1.0.0 --current-tag v1.1.0 --owner jamesonstone --repo kit
```

### 🔧 Utilities

| Command           | Description                         |
| ----------------- | ----------------------------------- |
| `mint version`    | Print the installed Mint version    |
| `mint completion` | Generate shell autocompletion       |
| `mint help`       | Help about Mint or a Mint command   |

Future release commands will be added after their feature specs define the
versioning, tag, and publishing contracts.

## 🛠️ Development

| Target                   | Description                                      |
| ------------------------ | ------------------------------------------------ |
| `make build`             | Build `bin/mint` with linker-injected version    |
| `make build-windows`     | Build `bin/mint.exe` for Windows amd64           |
| `make install`           | Install `mint` with linker-injected version      |
| `make install-git-hooks` | Use `.githooks/` as this clone's Git hooks path  |
| `make test`              | Run `go test -v ./...`                           |
| `make lint`              | Run `golangci-lint run ./...`                    |
| `make fmt`               | Run `go fmt ./...`                               |
| `make vet`               | Run `go vet ./...`                               |
| `make tidy`              | Run `go mod tidy`                                |
| `make clean`             | Remove local build output and run `go clean`     |
| `make all`               | Run `fmt`, `vet`, `test`, and `build`            |

The Makefile mirrors Kit's local build pattern. Release-domain behavior should
be added through feature specs before product commands are implemented.
