# Environments & sandboxes

Mint models environments by policy, not by a fixed development → staging → production ladder. A repository can contain only sandboxes, several shared test environments, local workspaces, or any project-owned combination. Each environment selects manual, reviewed, or automatic deployment; automatic deployment is Continuous Delivery.

## An arbitrary number of named environments

`environments` is a map with no fixed environment-count cap in the policy validator. Add entries as needed. Names must match `[a-z][a-z0-9_-]{0,62}`: a lowercase letter followed by up to 62 lowercase letters, digits, underscores, or hyphens.

This is configurable scale, not unlimited infrastructure. Each shared environment needs its own real runtime, approved configuration identity, verified baseline, adapter routing, retained artifacts, and GitHub Actions capacity. Mint does not provision or automatically delete ephemeral environments. Update policy and regenerate the controller when the environment list or callback workflows change.

## Sandbox-first policy

This onboarding example is valid policy but deliberately omits runtime configuration identities and baselines. Before deploying, set each shared environment's real `configuration_sha256` and establish trusted baseline evidence.

```yaml
schema_version: 2
mode: deployment
repository: owner/application
default_branch: main
default_environment: sandbox
authorization: repository-write
build_workflow: .github/workflows/build.yaml
validation_workflow: .github/workflows/checks.yaml
control_workflow: .github/workflows/mint-environments.yaml
required_checks: [Checks]
artifact:
  repository: ghcr.io/owner/application
environments:
  workspace:
    scope: local
    deploy: manual
    follow: latest
  sandbox:
    scope: shared
    deploy: automatic
    follow: latest
    image_tag: sandbox
    promotion_workflow: .github/workflows/deploy.yaml
    observation_workflow: .github/workflows/observe.yaml
  demo:
    scope: shared
    deploy: reviewed
    follow: latest
    requires: [sandbox]
    promotion_workflow: .github/workflows/deploy.yaml
    observation_workflow: .github/workflows/observe.yaml
  qa-pinned:
    scope: shared
    deploy: manual
    target: v1.2.3
    promotion_workflow: .github/workflows/deploy.yaml
    observation_workflow: .github/workflows/observe.yaml
```

Adapters shared by multiple entries must use the `environment` input to select the correct runtime and report that exact environment. The target `v1.2.3` must exist as authenticated eligible evidence before a request can deploy it. A full activated configuration adds, for example, `configuration_sha256: sha256:<64 lowercase hex digits>` using a genuine reviewed identity, not the placeholder.

## Choose scope and deployment policy

| Setting | Meaning |
| --- | --- |
| `scope: local` | Private machine observation cache; deployment remains your project's responsibility. |
| `scope: shared` | Authenticated repository journal and trusted adapters for a shared runtime. |
| `deploy: manual` | Collect candidates until an operator requests promotion. |
| `deploy: reviewed` | Keep one updating Release PR; checks and independent approval freeze its selection. |
| `deploy: automatic` | [Continuous Delivery](#continuous-delivery): follow previously approved policy without requiring a new Release PR for each ordinary candidate. |

Manual promotion enters the reviewed request path. Automatic mode still requires eligible trusted artifacts, activation, exact configuration, prerequisites, and verification. Use reviewed or manual mode for environments where a person must approve each change.

## Continuous Delivery

Continuous Delivery is an option for every environment, whatever its name. Set
`deploy: automatic` with `follow: latest` on the environments that should follow
the latest eligible build, and keep `reviewed` or `manual` on the others:

```yaml
environments:
  sandbox:
    scope: shared
    deploy: automatic      # Continuous Delivery: no Release PR per build
    follow: latest
    promotion_workflow: .github/workflows/deploy.yaml
  live:
    scope: shared
    deploy: reviewed       # one updating Release PR, frozen by approval
    follow: latest
    requires: [sandbox]
    promotion_workflow: .github/workflows/deploy.yaml
```

Mint freezes the exact artifact and runtime configuration of an automatic environment and dispatches your deployment adapter. One deployment runs at a time for an environment: a build that arrives while the deployment run is still active is registered and waits, and after a verified deployment Mint selects the latest eligible build, so an intermediate build can be skipped. After a failed or uncertain deployment the waiting builds stay registered and nothing delivers them until an operator settles the outcome; the controller fails on later builds until then and names the fenced intent. Change the mode in a reviewed policy PR to turn Continuous Delivery on or off for one environment. An explicit promote request, a rollback, and queued work after explicit recovery still go through a reviewed PR.

### Build after checks pass

Mint trusts a producer run only when `push` or `workflow_dispatch` started it, unless policy says otherwise. If your build starts after another workflow succeeds, list the events that may start it:

```yaml
build_workflow: .github/workflows/build.yaml
build_events: [workflow_dispatch, workflow_run]
```

`build_events` is a nonempty unique list of `push`, `workflow_dispatch`, and `workflow_run`. List only the events your producer workflow uses: each one is a way to build, and a `push` or `workflow_dispatch` build does not wait for another workflow. Without `workflow_dispatch`, isolated hotfix builds cannot register. The list also applies to recovery scans and the baseline source build; deployment and observation adapters are unaffected. A run that `workflow_run` started can supply only a normal candidate for its own head commit. Schema 1 policies do not accept the field.

GitHub runs a `workflow_run` workflow from the default branch, with your repository's secrets and token, and reports the default-branch head as its identity for every upstream run, including a pull request from a fork. Mint cannot see which run started your producer, so the producer must check it before any privileged step:

```yaml
name: Build
on:
  workflow_run:
    workflows: [Checks]
    types: [completed]
    branches: [main]
  workflow_dispatch:
permissions:
  contents: read
jobs:
  build:
    if: >-
      (github.event_name == 'workflow_dispatch' &&
       github.ref == format('refs/heads/{0}', github.event.repository.default_branch)) ||
      (github.event.workflow_run.conclusion == 'success' &&
       github.event.workflow_run.event == 'push' &&
       github.event.workflow_run.head_branch == github.event.repository.default_branch &&
       github.event.workflow_run.head_repository.full_name == github.repository &&
       github.event.workflow_run.head_sha == github.sha)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@11d5960a326750d5838078e36cf38b85af677262
        with:
          ref: ${{ github.sha }}
          persist-credentials: false
      # Build github.sha, publish immutable artifacts, then upload
      # mint-candidate. Allocate the source version in a job that has
      # contents: write.
```

- **Refuse at job level, or fail.** A refused run must end `skipped` or `failure`. A run that ends `success` with no `mint-candidate` makes the controller callback fail on each activated environment.
- **Do not interpolate the triggering run into scripts.** Its branch name and commit message are attacker-controlled. Read `github.event.workflow_run` fields only in `if:` expressions or through `env:`; `${{ github.event.workflow_run.* }}` inside `run:` is script injection with the producer's credentials.
- **Build `github.sha` only.** Never check out, execute, or download artifacts from the triggering run.

A producer that skips the check can deliver a default-branch commit whose checks did not pass. A producer that uses content of the triggering run can register that content under a reviewed version, and every environment that reuses the build can then receive it. Mint cannot detect either case.

A chained producer builds a commit only while it is still the default-branch head when its checks finish. When merges arrive faster than the checks run, no candidate appears until the newest commit's checks pass. GitHub limits `workflow_run` chains to three levels; checks, producer, and controller use two. Candidate registration also needs the immutable version tag of the exact source commit, so allocate [source versions](versioning.html) before the build.

### Separate activation switches

The generated controller runs only while a repository variable is `true`. The default is `MINT_RELEASE_ENABLED`. To activate one policy file without another, for example a sandbox policy beside a schema 1 production policy, name its own variable and regenerate the controller:

```yaml
activation_variable: MINT_SANDBOX_ENABLED
```

The name must start with `MINT_` and use upper-case letters, digits, and underscores. The prefix keeps the switch out of the compatibility adapters' runtime configuration identity. The variable gates the whole controller for that policy file.

### Roll out the two fields

`build_events` and `activation_variable` are authority: a target-only request PR cannot change them, and each is part of the frozen authority digest. A policy that uses neither keeps its digest and generated controller.

1. Pin every workflow that reads the policy (build, deployment, baseline, observation, controller) to a Mint release that has the fields. Earlier releases reject a policy that contains them.
2. Change the policy in a reviewed PR while no intent is pending or in flight. The change alters the authority digest of every environment in that policy file, so a pending approval or deployment start is refused.
3. Regenerate the controller, then set the activation variable when the environment is ready.

## Latest-following versus fixed targets

Choose exactly one of `follow: latest` or `target: vX.Y.Z`. Latest means the latest eligible candidate, not an arbitrary registry `latest` tag. A fixed target remains fixed until a reviewed policy change changes it.

`default_environment` determines the environment for omitted CLI selection and title-driven hotfixes. Without a default, omission works only for a single configured environment. Actions requests require an explicit environment.

`image_tag` is optional, must be unique across the policy, and must differ from immutable version tags. Aliases do not establish deployed state.

## Promote the same build

`requires: [sandbox]` makes `demo` depend on the sandbox having verified the complete selected artifact with no deployment in flight. Dependencies must name existing environments, contain no duplicates or cycles, and cannot make shared environments depend on local caches.

You can add `qa-east`, `qa-west`, or a reviewed `preview-42` as separate entries and reuse one candidate across them. Their runtime configurations and outcomes remain independent. A parallel dependency graph is valid; Mint does not force a single linear promotion chain.

## Publishing without production

No `production` entry is required. In deployment mode, optionally mark one shared environment `publish: true` to publish canonical releases after verification. At most one is allowed; naming it `distribution`, `demo`, or `live` does not change its evidence requirements.

For package/artifact projects, use `mode: package` or `mode: artifact` and keep the existing publication adapter. Those modes reject runtime operations and need no fabricated runtime state.


## Linear-tracked manual hotfixes

The schema 1 compatibility controller supports Linear tracking without a GitHub issue or a Linear token. For operation `hotfix`, set `HOTFIX_TICKET_URL` to the canonical `https://linear.app/<workspace>/issue/ENG-123` URL, `HOTFIX_BASELINE_ID` to the verified idle baseline, `HOTFIX_FIXES` to unique full single-parent commit SHAs (or `AUTHORED`), and `HOTFIX_REASON` to the reason. Leave `REQUEST_ISSUE` empty. Repositories may also bind `HOTFIX_TICKET_WORKSPACE` and `HOTFIX_TICKET_TEAM`; both must match the URL when configured. Native `prepare-hotfix --input <request.json> --run-id <active-control-run>` accepts the same `ticket_url`, `BaselineID`, and `Fixes` fields with `Issue: 0`.

The ticket is attribution, not authentication: Mint verifies the active trusted default-branch workflow_dispatch run and authorized human actor before external-ticket preparation. It does not fetch or edit Linear. The operator must select the repository's configured team and record the request in its ticket. Invalid/mixed provider inputs and stale baselines fail closed. New Linear source branches use `mint-hotfix-source/ENG-123` (or `mint-hotfix-source/<component>/ENG-123`), so an existing ordinary `ENG-123` development branch is preserved. Production-base branches and markers retain the exact ticket ID; immutable metadata and source PRs retain the canonical URL. A reused identifier with a different workspace URL, baseline, or fixes is rejected. Preparation never approves, merges, or deploys; existing source-review and provenance checks still apply. Legacy GitHub issue requests remain supported for repositories that select that tracker.

Native hotfix version allocation and candidate verification bind the source marker to the exact ticket ID, component-scoped production-base branch, and verified baseline. Both `ENG-123` and legacy `GH-123` identities use this same check. The normal independent current-head approval and configured checks still run before version allocation; reruns reuse the existing immutable tag.

Prepared Linear requests from older releases remain recoverable on their original bare ENG branches when ticket URL, baseline and fixes match. An ordinary bare development branch without hotfix metadata is left untouched while the new source uses its dedicated namespace. Transport or metadata failures never authorize overwriting either branch. Legacy GitHub request names remain unchanged.
