---
kit_metadata_version: 1
artifact: plan
feature:
  id: 0003
  slug: github-action
  dir: 0003-github-action
summary: Implementation plan for exposing the Mint CLI as a public GitHub composite action.
relationships:
  - type: depends_on
    feature: 0002-cli-patterns
    reason: The action delegates to the CLI scaffold and version command from this feature.
references:
  - id: spec
    name: GitHub Action specification
    type: feature_doc
    target: docs/specs/0003-github-action/SPEC.md
    relation: constrains
    read_policy: must
    used_for: binding requirements and acceptance criteria
    status: active
  - id: readme
    name: README
    type: repo_doc
    target: README.md
    relation: updates
    read_policy: must
    used_for: public quick start documentation
    status: active
---
# PLAN

## APPROACH

1. Add `action.yml` as a composite action because the current integration is a reusable sequence of runner steps around a CLI.
2. Use `actions/setup-go` with a configurable `go-version` input so the action can build the Go module on hosted runners.
3. Build `./cmd/mint` into `$RUNNER_TEMP/mint-action/mint`, inject the action ref as `pkg/cli.Version` when the ref is safe, and append the binary directory to `GITHUB_PATH`.
4. Run only a fixed allowlist of current commands after building: `version`, `help`, or `none`.
5. Capture command stdout in the `output` action output and expose the built path as `mint-path`.
6. Update README and durable repo docs so the action is documented as a CLI wrapper, not release automation.

## COMPONENTS

1. `action.yml`
   - Public GitHub Action metadata and composite action steps.
2. `README.md`
   - GitHub Action quick start examples.
3. `docs/CONSTITUTION.md`, `docs/references/tooling.md`, `docs/references/testing.md`, `docs/PROJECT_PROGRESS_SUMMARY.md`
   - Durable project facts and verification guidance.

## RISKS

1. Risk: The action could imply that Mint release automation exists.
   Mitigation: Keep supported action commands limited to current CLI behavior and document future release behavior separately.
2. Risk: Shell command injection through action inputs.
   Mitigation: Use an allowlist for `command` instead of raw command passthrough.
3. Risk: Marketplace consumers might expect a prebuilt binary.
   Mitigation: Make the action explicitly build the CLI from source and expose the resulting binary on `PATH`.

## TESTING

1. Parse `action.yml` as YAML.
2. Run a local build equivalent to the action build step.
3. Run the locally built binary with `version` and `--help`.
4. `go test ./...`
5. `go vet ./...`
6. `make build`
7. `git diff --check -- action.yml README.md docs`
8. `kit map 0003-github-action`
