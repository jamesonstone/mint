# Release PRs

A Release PR is the review surface for an exact artifact selection and runtime configuration. It is distinct from a source PR: it approves deployment of an already built candidate.

## One updating PR per reviewed environment

As successful eligible builds arrive, the normal Release PR updates to follow the environment's policy. Its declaration and summary are machine-owned release metadata. Independent human approval and required checks must match the exact current head.

Before approval, the selection can change. Approval freezes the complete selected artifact and configuration; later candidates cannot silently replace that frozen intent. Mint retains queued work for a later release.

## Review checklist

- Correct environment and operation: ordinary release, rollback, or hotfix.
- Exact source/version and complete immutable artifact digest or bundle.
- Approved runtime configuration identity and satisfied prerequisites.
- CHANGELOG and summary cover the intended changes.
- Required checks and independent approval apply to this exact head.

For `authorization: repository-write`, requests and independent human reviews require current effective write, maintain, or admin permission. Assignment alone grants no authority. Legacy named-operator policies retain their existing authorization model.

## Target-only policy requests

A PR can change one environment's `target`, `follow`, `operation`, or `reason` and act as its own deployment approval. Requests need an exact target except for `operation: resume`.

```yaml
target: v1.2.3
operation: rollback
reason: Restore the last verified authentication artifact
```

Replace `follow: latest` rather than keeping both. The target must exist in authenticated candidate or verified history as appropriate.

Mint compares policy before the first PR commit and before merge. A target request cannot weaken checks, change adapters, configuration, authorization, publisher, or prerequisites and then approve itself. Review those authority changes separately. Mint creates no second approval PR or separate status PR for a target-only request.

## After merge

The controller verifies approval, freezes intent, and dispatches the project adapter. The adapter reports start and verifies the exact environment/artifact/configuration tuple. Only matching trusted successful evidence advances deployment history.

Mint reports the result on the original approval PR using an idempotent machine-owned comment; human comments are preserved. Canonical publication is a separate retryable outcome after verification. A failed publication can be retried without redeploying the artifact.

A merged PR or accepted workflow dispatch is not proof of deployment. If evidence is incomplete or contradictory, the environment remains fenced until [reconciliation](recovery.html#reconcile-an-uncertain-deployment).

## Manual and automatic environments

Manual environments wait for an operator's **promote** request and then use the reviewed request path. Automatic environments deploy eligible candidates under previously approved policy; ordinary releases do not require a new review PR. Both still require trusted artifacts, exact configuration and verification.
