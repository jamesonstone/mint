# Integrate Mint

Use Mint in GitHub Actions to publish release state or manage deployment of an
existing artifact. The action builds Mint on the runner. Pin the action to the
full commit SHA of a reviewed published release.

## Choose a release path

### Publish release state

<ol class="flow" role="list" aria-label="Release-state publishing">
  <li><strong>Resolve</strong><br>Read Git history for a SemVer and release notes.</li>
  <li><strong>Tag</strong><br>Write the runner's version file and publish the Git tag.</li>
  <li><strong>Publish</strong><br>Create or reuse the GitHub Release.</li>
</ol>

Use [release-state publishing](#start-with-release-state-publishing) for tags
and GitHub Releases. For a package or artifact, keep the workflow that builds
and publishes it. Run `mint changelog` separately if you need to update
`CHANGELOG.md`.

### Manage environment deployments

<ol class="flow" role="list" aria-label="Environment-controlled deployment">
  <li><strong>Build once</strong><br>Your workflow retains an immutable artifact.</li>
  <li><strong>Select</strong><br>Mint records eligibility and freezes the selection.</li>
  <li><strong>Deploy and verify</strong><br>Your adapter uses the frozen digest and configuration.</li>
  <li><strong>Record</strong><br>Mint advances verified history and publishes if configured.</li>
</ol>

Use [environment policy](#add-environment-policy) for promotion without
rebuilding, reviewed releases, or deployment recovery. Choose automatic,
reviewed, or manual per environment. A manual promotion request uses the
reviewed release path. Mint writes the proposed changelog entry in a generated
Release PR. Ordinary automatic deployments skip that PR. If an environment has
`publish: true`, Mint publishes its canonical GitHub Release after deployment
verification. See [Release PRs](release-prs.html) for the review and reporting
steps.

## Optional local CLI

Use a published tag from [Mint releases](https://github.com/jamesonstone/mint/releases), rather than installing an unknown branch head. Mint's current build requires Go 1.26.9 or newer; check `go.mod` when upgrading.

```bash
go install github.com/jamesonstone/mint/cmd/mint@<published-tag>
mint version
mint help
```

Replace `<published-tag>` before running. For an environment integration, use a published Mint commit that supports the
policy fields and adapter contracts you configure. For example, `build_events`
and `activation_variable` require Mint v1.1.0 or later.

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

The action uses Go 1.26.9 by default. Override `go-version` only with a supported,
security-patched Go release. Release API URLs must use HTTPS; redirects are refused
so the release credential stays with the configured API endpoint. Read-only
commands need only `contents: read`; grant write permissions only to trusted
release jobs. Never run write-enabled release commands on untrusted pull-request
code or pass untrusted values as `github-api-url`. Command stdout is captured with
a collision-safe multiline output delimiter.

For environment-controlled deployment, follow the next steps instead of publishing every source build as a canonical release.

## Add environment policy

Create `.mint.yaml` using the [sandbox policy example](environments.html#sandbox-first-policy). Set the real repository, trusted workflow paths and check names. Choose explicit `default_environment` when multiple environments exist.

For a team, use `authorization: repository-write`. Mint checks current effective GitHub write, maintain, or admin access for human requests, including team-derived access. Independent reviewers must also have write access. Bots and read/triage access cannot authorize releases. Optional `assignees` route work but grant no authority.

Legacy policies can retain `human_login`; do not combine it with `authorization`. Review a migration of authorization separately from a target-only release request.

## Release several components from one repository

A monorepo can promote each component through its own environments with its own
version line. Write one schema 2 policy file per component, each with a
`component` block. A `deployment` component also needs its own
`control_workflow`; `package` and `artifact` components run no controller:

```yaml
# .mint/backend.yaml
schema_version: 2
mode: deployment
repository: acme/app
default_branch: main
authorization: repository-write
build_workflow: .github/workflows/backend-build.yaml
control_workflow: .github/workflows/backend-mint.yaml
component:
  name: backend            # lowercase letters, digits, and -
  tag_prefix: backend/     # optional; defaults to <name>/
  paths: [backend]         # Git pathspecs this component owns
  changelog: backend/CHANGELOG.md
environments:
  production:
    scope: shared
    deploy: reviewed
    follow: latest
    promotion_workflow: .github/workflows/backend-deploy.yaml
    publish: true
```

A component policy keeps everything separate from other components and from a
repository-wide policy:

| Concern | Repository-wide | Component `backend` |
| --- | --- | --- |
| Versions | `vX.Y.Z` | `backend/vX.Y.Z`, only for commits touching `paths` |
| Journal | `mint-release-state[-<env>]` | `mint-release-state.backend.<env>` |
| Release-control files | `CHANGELOG.md`, `.mint/proposal.json`, `.mint/summary.md` | `changelog`, `.mint/backend/proposal.json`, `.mint/backend/summary.md` |
| Proposal marker | `<!-- mint:proposal:…` | `<!-- mint:component-proposal:backend:…` |
| Controller concurrency | `mint-release-journal` | `mint-release-journal-backend` |
| GitHub Release | marked Latest | never marked Latest |

Give each component its own build, deployment, and observation workflows, so
callback evidence stays unambiguous. Generate one controller per policy with
`--config .mint/backend.yaml` and `--output` set to that policy's
`control_workflow`. Pass the same file to the compatibility adapters with
`MINT_CONFIG=.mint/backend.yaml`, and to source versioning with
`version-config` (see [component releases](versioning.html#component-releases-in-one-repository)).
A target such as `target: backend/v1.4.0` must carry the component prefix.

## Connect trusted project adapters

| Adapter | Required evidence |
| --- | --- |
| Build producer | Successful trusted run with a single `mint-candidate.json` inside the `mint-candidate` Actions artifact. |
| Deployment | Upload `mint-request.json` as `mint-request` before start; deploy the frozen digests; upload exact verified `mint-deployment.json` as `mint-deployment` after verification. |
| Observation | Upload `mint-observation.json` as `mint-observation` with exact source/version, full artifact, runtime configuration, timestamp, run ID, and run URL. |

Allocate source versions with `mint release version-main` or the Action's
`release-version-main` command. Build the exact tagged source commit and retain
its artifact by digest.

By default, Mint accepts `push` and `workflow_dispatch` as producer events.
To build only after another workflow succeeds, list `workflow_run` in
`build_events` and make the producer verify its triggering run in a job-level
`if:`. For automatic deployments, use Mint v1.1.2 or later and wire the
adapter's completion notification so the controller can settle the deployment
without waiting for the next build. See
[Continuous Delivery](environments.html#continuous-delivery) for both workflows.

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
4. Enable the policy's activation variable (`MINT_RELEASE_ENABLED`, or the `activation_variable` named in policy) and the project's existing deployment gates only after those prerequisites are ready.
5. Test with an eligible candidate. In a reviewed or manually requested release,
   inspect the Release PR and the verified result. In an automatic deployment,
   inspect the adapter run and verified environment history.

Generation, installation, or merging policy does not provision cloud resources, grant IAM, deploy a runtime, or activate an existing consumer.

## Check integration state

```bash
mint deployment policy --config .mint.yaml --environment sandbox
mint deployment status --config .mint.yaml --environment sandbox --format markdown
```

Policy inspection works before activation. Shared status reads authenticated repository state and needs the configured token (default `GH_TOKEN`). Treat missing evidence as unknown. Test a sandbox release and recovery before relying on the controller for a critical environment.

## License

Mint is licensed under [Apache-2.0](https://github.com/jamesonstone/mint/blob/main/LICENSE).
The [NOTICE](https://github.com/jamesonstone/mint/blob/main/NOTICE) preserves project
attribution and identifies the bundled fonts' separate SIL Open Font License.
