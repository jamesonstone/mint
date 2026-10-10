# Environments & sandboxes

Mint models environments by policy, not by a fixed development → staging → production ladder. A repository can contain only sandboxes, several shared test environments, local workspaces, or any project-owned combination.

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
| `deploy: automatic` | Follow previously approved policy without requiring a new Release PR for each ordinary candidate. |

Manual promotion enters the reviewed request path. Automatic mode still requires eligible trusted artifacts, activation, exact configuration, prerequisites, and verification. Use reviewed or manual mode for environments where a person must approve each change.

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
