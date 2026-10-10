# Integrate Mint

Start with the smallest release model your repository needs. Add environment control when you need reviewed promotion, verified history, or recovery across environments.

## Install the CLI

Use a published tag from [Mint releases](https://github.com/jamesonstone/mint/releases), rather than installing an unknown branch head. Mint's current build requires Go 1.25.5 or newer; check `go.mod` when upgrading.

```bash
go install github.com/jamesonstone/mint/cmd/mint@<published-tag>
mint version
mint help
```

Replace `<published-tag>` before running. An environment integration must pin a published immutable Mint commit containing schema 2 and team authorization support. Older releases may not support these features.

## Start with release-state publishing

For repositories that only need tags and GitHub Releases, this minimal workflow runs Mint's existing action. Replace `<published-mint-commit>` with the full commit SHA of a reviewed published release.

```yaml
name: Publish release state
on:
  push:
    branches: [main]
permissions:
  contents: write
concurrency:
  group: release-state
  cancel-in-progress: false
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@11d5960a326750d5838078e36cf38b85af677262
        with:
          fetch-depth: 0
      - uses: jamesonstone/mint@<published-mint-commit>
        with:
          command: release-publish
          commitish: ${{ github.sha }}
          owner: ${{ github.repository_owner }}
          repo: ${{ github.event.repository.name }}
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

This writes `.version` in the runner workspace without the leading `v`. It does not commit that file, build containers, or deploy. Add application-owned build steps explicitly. Use one release publisher; remove competing automatic version/tag publishers when adopting Mint.

For environment-controlled deployment, follow the next steps instead of publishing every source build as a canonical release.

## Add environment policy

Create `.mint.yaml` using the [sandbox policy example](environments.html#sandbox-first-policy). Set the real repository, trusted workflow paths and check names. Choose explicit `default_environment` when multiple environments exist.

For a team, use `authorization: repository-write`. Mint checks current effective GitHub write, maintain, or admin access for human requests, including team-derived access. Independent reviewers must also have write access. Bots and read/triage access cannot authorize releases. Optional `assignees` route work but grant no authority.

Legacy policies can retain `human_login`; do not combine it with `authorization`. Review a migration of authorization separately from a target-only release request.

## Connect trusted project adapters

| Adapter | Required evidence |
| --- | --- |
| Build producer | Successful trusted run with a single `mint-candidate.json` inside the `mint-candidate` Actions artifact. |
| Deployment | Upload `mint-request.json` as `mint-request` before start; deploy the frozen digests; upload exact verified `mint-deployment.json` as `mint-deployment` after verification. |
| Observation | Upload `mint-observation.json` as `mint-observation` with exact source/version, full artifact, runtime configuration, timestamp, run ID, and run URL. |

The deployment workflow receives `environment`, `intent_id`, and serialized `intent` inputs. It must consume these exact identities and avoid rebuilding. Mint validates server-side repository, default branch, workflow path, run identity/completion, archive digest, and manifest contents.

Use the [full adapter reference](https://github.com/jamesonstone/mint/blob/main/docs/references/environment-lifecycle.md#project-adapter-evidence) and [legacy integration contract](https://github.com/jamesonstone/mint/blob/main/docs/references/production-release-proposals.md) for manifest fields and controller wiring. These contracts are integration work: a policy file alone cannot make an existing deploy workflow compliant.

## Generate and review the controller

After the configured build, deploy, and observe workflow files exist, generate the repository controller:

```bash
mint deployment workflow \
  --config .mint.yaml \
  --environment sandbox \
  --mint-ref <published-mint-commit> \
  --output .github/workflows/mint-environments.yaml
```

The generator reads callback workflow names from the files, so create those adapters first. Review the generated default-branch controller and validation integration before merging. The controller needs contents/issues/pull-requests/Actions write and checks read; version/build jobs need contents write and Actions/pull-requests/checks read. Native CI can remain read-only. Use a job-scoped token; never embed credentials in policy or manifests. Configure repository rules to require exact-head checks and independent human review. GitHub must permit the controller token to create PRs when generated Release PRs are used.

## Activate a shared environment

1. Supply a real approved runtime `configuration_sha256` and trusted adapters.
2. Configure `baseline_workflow` (and `baseline_build_workflow` if migrating a different producer). Run the trusted baseline adapter to reconcile the actual running artifact/configuration and upload digest-attested `mint-baseline` evidence. Import it with `mint deployment bootstrap --environment sandbox --run-id <verified-run-id>`. Do not infer a baseline from an image tag, supply a local JSON file, or re-import to erase an uncertain deployment.
3. Review candidate production and retention, complete-bundle verification, required checks, and controller callback paths.
4. Enable `MINT_RELEASE_ENABLED` and the project's existing deployment gates only after those prerequisites are ready.
5. Promote an eligible candidate and inspect the Release PR and verified result end to end.

Generation, installation, or merging policy does not provision cloud resources, grant IAM, deploy a runtime, or activate an existing consumer.

## Check integration state

```bash
mint deployment policy --config .mint.yaml --environment sandbox
mint deployment status --config .mint.yaml --environment sandbox --format markdown
```

Policy inspection works before activation. Shared status reads authenticated repository state and needs the configured token (default `GH_TOKEN`). Treat missing evidence as unknown. Test a sandbox release and recovery before relying on the controller for a critical environment.
