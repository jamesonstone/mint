# Environment lifecycle and adapter contract

Mint separates repository intent, reviewed artifact selection, deployment evidence,
and runtime observation. A successful trusted build produces one immutable
candidate. Any configured environment can reuse that candidate without rebuilding;
a multi-component application uses one complete bundle of component digests.
Environment names are project-owned: `stage`, `live`, `distribution`, or other
names are valid. No production environment is required.

A reviewed release uses one PR that will track eligible builds until approval
freezes the selection. A policy request that changes only one environment's target
and recovery intent can use that same policy PR as its approval. Automatic
deployment follows previously approved policy; manual environments collect
candidates until an operator requests promotion. Completion updates durable
verified history and the original approval PR, never a separate status PR.

## Repository policy

This schema 2 example describes environments before activation. Adapter paths are
examples of project-owned workflows, not workflows that Mint provisions. Add each
shared environment's approved `configuration_sha256` and establish its verified
baseline before requesting deployment; omission permits onboarding but does not
permit deployment.

```yaml
schema_version: 2
mode: deployment
repository: owner/application
default_environment: live
default_branch: main
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
  stage:
    scope: shared
    deploy: automatic
    follow: latest
    image_tag: stage
    promotion_workflow: .github/workflows/deploy.yaml
    observation_workflow: .github/workflows/observe.yaml
  live:
    scope: shared
    deploy: reviewed
    follow: latest
    image_tag: live
    requires: [stage]
    publish: true
    promotion_workflow: .github/workflows/deploy.yaml
    observation_workflow: .github/workflows/observe.yaml
```

Use `follow: latest` **or** `target: v1.2.3`, never both. Runtime configuration is
an exact `sha256:<64 lowercase hex digits>` identity, separate from the build
configuration captured in the candidate. `image_tag` is an optional readable alias;
it is not evidence of deployment or permission to substitute a mutable artifact.
A shared prerequisite must have verified the complete selected artifact and have
no deployment in flight. Shared environments cannot depend on a local cache.

| Mode | Meaning |
| --- | --- |
| `deployment` | Immutable artifacts plus environment deployment, verification, and observation adapters. |
| `artifact` | Artifact-oriented publication policy; runtime operations are rejected. |
| `package` | Package-oriented publication policy; runtime operations are rejected. |

Artifact and package environments need no fabricated runtime baseline, configuration,
or deployment workflow. They describe publication intent and retain the project's
existing publication implementation. For deployment projects, select at most one
shared environment with `publish: true` as the canonical release publisher; other
environments deploy without independently publishing the same version. Schema 1
production policy and the existing command namespace remain supported.

## Team ownership and request authority

Use `authorization: repository-write` for a team-owned project. Mint checks the
actor's current effective repository permission with GitHub: write, maintain or
admin access permits a human request. Team-derived repository access is included;
read/triage access, bots and unavailable or mismatched permission evidence do not
permit requests. Mint still authenticates the active default-branch controller run,
exact PR head, required checks and independent human review. In this mode the
reviewer must also have repository write access and must not be the PR author.

Assignment is optional and independent:

```yaml
authorization: repository-write
# Optional routing only; this does not grant deployment authority.
assignees: [release-coordinator]
```

Omit `assignees` for an unassigned team workflow, or use `assignees: []` to request
no assignment explicitly. No personal login is needed. GitHub team access manages
who can act; Mint does not create memberships or modify repository permissions.

Existing `human_login` policies retain their named-operator authorization and,
unless assignment is specified separately, their existing assignee. Do not combine
`human_login` with `authorization`. Migrating to repository permissions is a
reviewed authority change, not a target-only deployment request. Both schema 1 and
schema 2 accept the new policy in the feature-bearing release; an older Mint pin
must be upgraded before removing its required `human_login` field.

The [GitHub permission API](https://docs.github.com/en/rest/collaborators/collaborators#get-repository-permissions-for-a-user)
reports effective access across repository, team, organization and enterprise
grants. Mint uses the base permission, so a custom role with base write access is
supported without hard-coding the organization's role names. Its API lookup needs
repository Metadata read, available to repository installation credentials.

## Everyday interaction and recovery

Use the generated **Mint environment control** repository workflow. Its environment
choice is explicit; an omitted CLI environment selects `default_environment`, or
the sole configured environment, and otherwise fails rather than guessing.

| Actions operation | Operator input and result |
| --- | --- |
| `promote` | Choose an environment and optionally an eligible exact version; request reviewed promotion. |
| `hotfix` | Supply a fix PR or issue and a reason; prepare an isolated source fix against verified deployment, then review source and release. |
| `rollback` | Supply a reason and optionally a retained verified version; default to the previous distinct verified artifact. Review the recovery release. |
| `resume` | Make a fresh request to clear rollback's latest-following pause. An explicit policy target remains fixed. |
| `observe` | Request a read-only adapter observation; its successful completed run supplies timestamped runtime evidence. |
| `reconcile` | Select the active intent, or leave it blank for the current fence, and a fresh successful observer run; explicitly settle an uncertain completed deployment. |

Prefer roll-forward: merge corrected or reverted source and release it as a new
version. `hotfix: ...`, `hotfix(component): ...`, and
`hotfix(GH-123): :firetruck: ...` PR titles also initiate the isolated hotfix path
through the trusted controller after source review and merge. With multiple
environments, configure an explicit default for this title-driven path. Integrate
isolated fixes into the default branch so ordinary releases retain the correction.

Rollback restores a retained artifact using **current trusted runtime
configuration**, not the historical configuration. It does not reverse migrations,
data changes, or external side effects. A successful rollback pauses latest
following, preventing an immediate return to the version just rolled back. A resume
request has a durable consumed identity: replaying an old resume cannot clear a
later rollback pause. To recover runtime configuration itself, review the desired
configuration and deploy it explicitly.

A YAML request changes one environment's `target`, `follow`, `operation`, and/or
`reason`, for example replacing `follow: latest` with:

```yaml
target: v1.2.3
operation: rollback
reason: Restore the last working authentication artifact
```

The target must exist in authenticated candidate or verified history evidence as
appropriate. Request PRs require an exact target, except `operation: resume`.
Required checks and independent human approval apply to the exact PR head. Mint
checks the policy before the PR's first commit and before merge: a request cannot
weaken checks, change adapters, runtime configuration, publisher, prerequisites,
or other authority and then approve itself. Such authority changes require their
own review. The request PR is the release approval PR; Mint creates no second
approval or status PR for it. A canonical generated latest-following release PR
instead includes its exact machine-owned declaration and remains dynamic until
approval freezes it.

## Desired, verified, and observed state

| Surface | What it establishes |
| --- | --- |
| Desired policy | Latest-following or fixed target, deployment mode, configuration, and prerequisites. |
| Last verified deployment | Durable artifact and configuration history from authenticated successful deployment evidence. |
| Live observation | Last authenticated runtime observation, its timestamp, run URL, and `live`, `drift`, `deploying`, or `unknown` status. |

`mint deployment status --environment live --format markdown` displays these
separately. An old observation is last-known evidence, not a timeless claim about
runtime. Shared environments use isolated durable journals: production retains
`mint-release-state`; other names use `mint-release-state-<environment>`. Local
observations are private machine caches and cannot advance shared deployment
history or prove a shared prerequisite. Local policy does not provision or deploy
a developer runtime.

A dispatch response proves only that a workflow request was accepted. Missing,
unverified, mixed, or conflicting deployment evidence retains the deployment fence.
An observation alone never clears it. For explicit reconciliation, first obtain a
trusted live observation taken **after** the original adapter completed, then run
`reconcile`. The complete observation must exactly match either the frozen intent
(success) or the previous verified baseline (failed deployment with unchanged
runtime). Any other state retains the fence for investigation. Successful reconciliation also
retries pending publication; replay retries publication without redeployment.
Unpaused reviewed and automatic environments restore a reviewed PR for queued work
after explicit recovery. Manual environments still require promotion.

## Project adapter evidence

Adapters remain language-, registry-, package-manager-, and runtime-independent.
Mint authenticates server-side repository, workflow path, default branch, run
identity and completion, then validates the GitHub artifact archive digest and
manifest. A repository payload cannot substitute for those server checks.

| GitHub Actions artifact | Contract |
| --- | --- |
| `mint-candidate` | Successful trusted producer's `mint-candidate.json`: immutable source, version, build evidence, and artifact identity with environment-neutral `build` scope. |
| `mint-request` | Deployment adapter's `mint-request.json` before start: exact frozen environment, intent, source, full artifact, and runtime configuration. Binds the adapter run to the intent. |
| `mint-deployment` | Adapter's `mint-deployment.json` after verification: exact identity, `outcome`, and `verified`. Successful deployment requires successful trusted run and complete matching identity. |
| `mint-observation` | Observer's `mint-observation.json`: repository, environment, full artifact, configuration, source/version, status, `observed_at`, `run_id`, and `run_url`. Timestamp must belong to the authenticated run. |

Each artifact archive contains its single matching JSON manifest. For a bundle,
`artifact.members` records every component's reference, digest, and build
configuration; the top-level digest fingerprints the canonical complete bundle.
Deployment and observation compare every member. Verifying only one component of
a multi-component application cannot establish a successful deployment.

The promotion adapter receives `environment`, `intent_id`, and serialized frozen
`intent` workflow inputs. It must deploy the approved digests without rebuilding,
report start against `mint-request`, and emit runtime verification against
`mint-deployment`. A failed adapter may establish unchanged runtime only by verifying
the exact previous baseline; otherwise the outcome remains uncertain. Publication
is a separate retryable result after verified deployment, not a reason to redeploy.
Redeploying an already published version with new runtime configuration records
deployment history without replacing the original canonical GitHub Release.
Outcome reporting uses an idempotent machine-owned comment on the original PR and
preserves human comments.

## Activation boundary

Policy parsing and workflow generation are preparation, not activation. A consumer
must pin a published feature-bearing Mint commit, review the trusted controller and
project adapters, supply approved runtime configuration identities, establish
baseline evidence for shared deployments, and enable its existing activation gates.
The generated controller uses `MINT_RELEASE_ENABLED`; it does not create cloud
resources, grant IAM permissions, provision a runtime, or fabricate a baseline.
The controller subscribes to configured adapter workflow names (or the repository
workflow path when its name is omitted), authenticating each callback path separately.

Repository merge, accepted dispatch, verified deployment, and live observation are
separate outcomes. This documentation does not claim existing consumers have been
migrated or activated. See the [legacy production adapter reference](production-release-proposals.md)
for schema 1 integration and the [environment feature spec](../specs/0010-environment-lifecycle/SPEC.md)
for the durable design decisions.
