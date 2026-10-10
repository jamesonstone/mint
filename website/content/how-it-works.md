# How Mint works

Mint separates release intent from the evidence that makes it safe to act on that intent. Its state lives with your repository and trusted GitHub Actions runs.

## The release lifecycle

1. **Version source.** Serialize version allocation on the current default-branch head. Each untagged first-parent source commit receives an immutable strict SemVer tag.
2. **Build once.** Your trusted producer builds the application and uploads a `mint-candidate` artifact containing `mint-candidate.json`.
3. **Register eligibility.** Mint authenticates the completed producer run and its archive digest, then records source, version, build configuration, and immutable artifact identity.
4. **Select per environment.** Policy chooses latest-following or an exact target. Prerequisites must have verified the complete selected artifact.
5. **Freeze intent.** Reviewed releases wait for current-head checks and independent human approval. Mint freezes the artifact and runtime configuration together.
6. **Deploy and verify.** The project adapter consumes the frozen intent without rebuilding, reports start, and supplies exact verification evidence.
7. **Record and report.** Verified history advances. Mint reports on the original PR and, if configured, publishes the canonical GitHub Release.

## Artifact identity

A readable tag such as `sandbox` is an optional alias. The selected artifact is an immutable digest. For a multi-component application, `artifact.members` includes every component's reference, digest, and build configuration; the top-level digest fingerprints the complete bundle.

The adapter must deploy and verify every member. A successful API deployment does not prove that the accompanying worker or frontend is correct.

Build configuration identifies how the artifact was made. `configuration_sha256` separately identifies the runtime configuration approved for an environment. The same artifact can be deployed with different reviewed runtime configuration without rebuilding it.

## State and concurrency

Shared environments have separate durable journals. The legacy `production` environment uses `mint-release-state`; other names use `mint-release-state-<environment>`. Each environment retains a fence while a deployment is in flight or its outcome is uncertain.

An uncertain outcome cannot be cleared by dispatching another deployment or merely observing the runtime. [Explicit reconciliation](recovery.html#reconcile-an-uncertain-deployment) checks fresh trusted evidence against the frozen intent or previous baseline.

Local scope stores private machine observations. It cannot advance shared deployment history or satisfy a shared environment's prerequisite.

## What belongs to your application

Mint owns version identity, release state, reviewed selection, and evidence validation. Your project owns Docker/package builds, registry authentication, artifact retention, deployment, runtime verification, cloud resources, and infrastructure permissions.

Basic `mint release publish` resolves a version, writes `.version`, creates or reuses the Git tag, and creates or reuses a GitHub Release. That is a release-state publishing command, not the entire environment controller. It does not deploy an application.

## Publication-only projects

Schema 2 supports `mode: artifact` and `mode: package` alongside `mode: deployment`. Publication-only environments describe publication intent and keep the project's existing publication implementation. Runtime operations are rejected; no fake deployment baseline is required.

For deployment mode, at most one shared environment can have `publish: true`. Its verified release is canonical; other environments deploy without independently publishing the same version.
