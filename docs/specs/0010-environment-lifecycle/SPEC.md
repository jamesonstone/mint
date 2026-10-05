---
status: implementation
issue: GH-17
skills: []
---
# Generic environment lifecycle

Mint builds a version once through a project-owned workflow. Each named environment follows eligible builds or holds an explicit version; reviewed environments keep one updating release PR, and merging that PR freezes the exact artifacts and runtime configuration for the project adapter. Hotfix and rollback requests come from repository Actions; prefer a corrective or revert hotfix. Rollback restores an existing verified artifact and pauses automatic advancement until explicitly resumed.

`.mint.yaml` is editable deployment intent and policy. The journal stores separately authenticated deployment history and timestamped live observations. Ambient checks on merges and after deployment report drift; they never approve drift, move image aliases, or deploy. Optional aliases such as dev are labels for humans; immutable artifact identities remain the deployment authority.

## Acceptance criteria

- AC1: Strict schema 2 policy supports arbitrary nonempty environment sets, local/shared scope, manual/automatic/reviewed deployment, latest-follow XOR fixed strict version target, optional repository/image aliases, environment dependencies and publication policy. Reject unknown/duplicate fields, aliases/merge keys, unsafe paths, dependency cycles and tag collisions. Schema 1 production policies remain compatible.
- AC2: Per-environment journals, proposals, intents, locks, history and observations cannot collide; preserve the existing production journal. Authenticated neutral build candidates can be reused without changing their source manifest identity or rebuilding. Exact immutable image bundles are verified completely, with runtime configuration separate from build configuration.
- AC3: A target-only `.mint.yaml` PR can itself request one environment deployment, evaluated using its pre-change trusted policy; changing policy cannot weaken its own approval. Repository Actions offer equivalent explicit promote/hotfix/rollback/resume requests. A YAML request must not create a second approval PR. Latest-follow reviewed proposals update on each eligible build before freeze; fixed requests remain fixed.
- AC4: Read-only trusted observation runs report desired/last-verified/live/status/as-of/evidence. Reject foreign, stale, future or conflicting observations. Mixed rollout is deploying; missing/failed proof is unknown. Observation cannot clear fences, change baseline or history, deploy, or move aliases. Local state remains machine-scoped.
- AC5: Every deployment verifies exact candidate artifacts, environment and runtime configuration. Failure without verified unchanged runtime remains locked. Promotion prerequisites require verified upstream artifact identity. Rollback uses previously verified history, pauses automatic progression, and requires explicit resume. Publication is optional, with at most one shared environment owning the repository canonical GitHub Release for a version.
- AC6: Provider operations remain adapters: no AWS-specific Mint core, no registry rebuild or alias mutation in observation. Package/artifact publishing projects have honest policy modes and do not invent running deployment baselines. New runtime environments require explicit bootstrap proof; configuration alone does not provision or activate them.
- AC7: Existing commands retain compatibility, generic CLI/Actions surface is small, policy/status/workflow documentation and Mermaid diagram agree. Unit/adversarial tests exercise isolation, stale evidence, tampered policy, bundle mismatch and replay. New Mint PR has current CI and independent review before authorized merge/publication.
- AC8: Discover current consumers, update published Mint pins and add validated `.mint.yaml` in ready project PRs. Preserve existing adoption lanes and disabled activation gates; do not merge consumer PRs, provision resources, change permissions, or deploy them.

## Source map

Policy: pkg/promotion/config.go. Authenticated artifacts/runs: artifact_evidence.go, github.go. Durable journals: journal.go. Lifecycle: types.go, engine.go, intent.go, declaration.go. Repository controller: pkg/cli/release_production*.go; proposal_github.go; control*.go; action.yml. Deployment implementations remain consumer workflows.

## Design decisions

Schema 2 extends the existing engine rather than replacing trust boundaries. Existing production journal branch stays mint-release-state; other environments use validated mint-release-state-<name>, never a ref below an existing ref. A neutral build scope is explicit; schema 1 environment matching stays strict. Artifact bundles must have a deterministic complete identity, and deployment evidence must cover all components. Policy modes deployment/artifact/package distinguish runtime from publication. No hard-coded production requirement. Local scope never writes shared journal state.

Explicit human reconciliation is a separate recovery operation: reauthenticate a fresh observer after the original promotion run completed; accept only the exact frozen intent or exact prior baseline, retaining the fence for every other result. Ambient observations never invoke this operation. A plain YAML deployment request selects an exact version; a generated latest-follow release PR retains its canonical reviewed artifact declaration. Authority changes invalidate pending approval, and consumed resume/reconciliation keys prevent replay from affecting later locks.

A verified completion supersedes older observations while retaining their evidence; they project as unknown until a new observer completes. Replayed terminal completion cannot change later baselines. A configuration-only redeployment of an existing version records environment history and preserves the original version Release. After explicit recovery, queued work returns to a reviewed release PR, including automatic environments, so exception handling exposes the next selection for review.

## Out of scope

Cloud provisioning, IAM/repository-setting changes, package-manager-specific publication, automatic merge, runtime activation in consumers, migration reversal and data rollback. No fabricated bootstrap baseline or unverified runtime success.

## Validation and review

PASS: full Go suite and race suite, vet, native build, affected-file lint, secret scan, formatting and whitespace checks, generated controller actionlint, policy example parsing, and CLI smoke checks. Independent adversarial review repaired policy-authority smuggling, callback/recovery replay, missing rollback error handling, stale observation ordering, configuration-only release republication, and queued recovery behavior. Provider runtime remains unobserved; consumer activation is separate.
