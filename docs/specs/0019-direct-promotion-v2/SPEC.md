---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "verify"
feature:
  id: "0019"
  slug: "direct-promotion-v2"
  dir: "0019-direct-promotion-v2"
---
# SPEC

## PURPOSE

Mint v2 provides release primitives for repository-owned deployment workflows. Remove the legacy deployment controller and prove build-once promotion on hosted GitHub Actions for a single project and for a frontend and backend in one repository.

## CONTEXT

Mint v1.1.2 couples promotion to a journal, generated release PRs, dispatch and completion workflows, and bundled adapters. Existing controller consumers can stay on an immutable v1 pin. Mint's conventional-commit resolver, changelog, source tags, and GitHub Release publisher work independently.

## REQUIREMENTS

- After passing default-branch checks, build each fixture once and verify its digest in continuous staging. An authorized dispatcher selects an exact project and SemVer for production. Both environments call the same promotion steps, with no production build.
- Bind staging evidence to the project, version, source, digest and trusted completed run. Reject missing, failed, cross-project and wrong-digest evidence. Failed and cancelled verification must produce no verified receipt or publication.
- Remove controller commands, renderer, state, and bundled controller adapters from v2. Preserve usable release primitives. Publish real release state only when a project invokes the appropriate release primitive.
- Run two private fixture repositories under jamesonstone. Provider smoke testing is separate.
- Provide `mint verify --expected --observed` for the shared promotion step. Reject missing, malformed, and mismatched SHA-256 inputs. The workflow obtains the observed digest and owns runtime coverage and evidence.

## ACCEPTED PLAN

1. Configure both private repositories with a checked-in GitHub Actions workflow and deterministic fixtures. Exercise positive and negative job paths on GitHub-hosted runners.
2. Remove the controller implementation and decouple remaining release functions. Mark the Mint pull request as a breaking change. Update the public and agent guides and rerun the native release tests and documentation checks.
3. Record hosted proof results in this specification and the pull request. Keep private run and artifact links in the Work Ledger. Keep provider operations and downstream migrations in separate work.

## DECISIONS

- Remove the controller now, including its CLI and action entry points. Keep v1 consumers on the v1 line until they migrate to their own workflows.
- Use a `/v2` Go module path so the v2 release can be installed and imported without changing the v1 import path.
- An authorized dispatcher is sufficient for the private fixture production path. Exact verified staging evidence is mandatory.
- Use publication receipts in the fixture. Production provider smoke and the post-fixture Mint configuration and Kit scaffolding designs need separate validation.
- Compare supplied SHA-256 digests in the CLI. Keep provider lookup in the repository workflow and use the comparator before writing a runtime receipt.

## DISCOVERIES

- Both testbeds used GitHub-hosted runners and a reusable promotion workflow. The fixture uses a workflow-run-derived SemVer and the GitHub artifact ID. Its deployment copies retained bytes into a fixture runtime directory and hashes those bytes.
- Current Kit reconcile refreshes both full action SHA pins and version tags to the latest Mint GitHub Release. Publishing v2 as latest can therefore break v1 controller consumers even if their workflows are pinned. Require a Kit refresh guard or a release strategy that keeps legacy reconciliation on v1 before v2 becomes the latest release.

## VALIDATION

- GitHub-hosted proof completed on 2026-11-23 UTC. Frontend, backend, and single project each built once and promoted `v1.0.1` through continuous staging and authorized manual production. Downloaded bytes, build manifests, staging receipts, production receipts, and fixture publication markers agreed on the respective SHA-256 and retained artifact ID. The frontend and backend had distinct hashes under the same SemVer. Each manual run had zero build jobs and used the same promotion steps as staging.
- Production refused missing, failed, cancelled, wrong-project, and wrong-digest staging evidence before the promotion job. A failed check or build prevented staging promotion. A forced digest mismatch and an interrupted verification in each environment emitted neither a successful verification receipt nor a publication marker.
- Local pass: `go test ./...`, `go test -race ./...`, `go vet ./...`, `make build VERSION=v2.0.0`, Go formatting, Actions lint, and `make docs-check`. The built CLI reports `v2.0.0` and rejects the removed commands. Fixture Python tests and workflow lint passed in both testbeds.
- `kit check direct-promotion-v2` passed. With Kit v3.0.64, `kit check --project` reports the same 240 blocking findings on the v1.1.3 base and this worktree.
- After refreshing onto v1.1.3, a fresh-context review verified the documentation resolution against both parents. The documentation build passed, and the built-in browser showed the overview, integration, and artifact evidence diagrams in order at desktop and mobile widths.
- GH-49 adds command tests for matching, missing, malformed, and mismatched inputs. The fixture integration tests in both private repositories ran the pinned CLI in GitHub Actions and confirmed that a mismatch leaves no receipt.

## OUTCOME

Both private fixtures proved the direct workflow. Mint v2 removes the legacy controller, its adapters, and deployment state. The release primitives remain available for project-owned workflows. Provider tests and consumer migrations are separate.

## REPOSITORY MEMORY

Update the active lifecycle guides, generated documentation sources, release reference, and Constitution for the v2 boundary.
