# Rollback & hotfix

Prefer roll-forward when practical: correct or revert source, build it, and release a new version. Use rollback for retained-artifact recovery and hotfix to isolate a correction from queued default-branch work.

## Request recovery in Actions

Open **Actions → Mint environment control → Run workflow**. Select the environment explicitly and choose an operation.

| Operation | Input and result |
| --- | --- |
| `promote` | Optionally choose an eligible exact version and request reviewed promotion. |
| `hotfix` | Supply a fix PR or issue and a reason; prepare an isolated correction against verified deployment. |
| `rollback` | Supply a reason and optionally a retained verified version; otherwise select the previous distinct verified artifact. |
| `resume` | Make a fresh request to clear the pause caused by rollback. |
| `observe` | Request fresh read-only runtime evidence from the trusted observer. |
| `reconcile` | Supply the active intent, or leave it blank for the current fence, and a fresh successful observer run. |

These are repository controller requests. They require an activated integration and authenticated human authority; a local command invocation alone is not permission to deploy.

## Roll back a retained artifact

1. Inspect verified history and confirm the artifact or complete bundle is still retained and available.
2. Request **rollback** with a reason and, optionally, an exact previously verified version.
3. Review the recovery PR's artifact and current runtime configuration; merge after current-head checks and independent approval.
4. Inspect matching adapter verification and the outcome on the same PR.
5. When ready for ordinary releases, make a fresh **resume** request.

Rollback uses **current trusted runtime configuration**, not historical configuration. It does not reverse database migrations, data changes, or external side effects. To recover configuration, review the desired configuration and deploy it explicitly.

After a successful rollback, latest-following releases pause to avoid immediately redeploying the version just rolled back. Replaying an old resume request cannot clear a later pause. A fixed policy target stays fixed after resume.

## Isolate a hotfix

Use an Actions **hotfix** request with a merged reviewed fix PR, or an issue for authoring a new isolated correction. A source PR titled `hotfix: ...`, `hotfix(component): ...`, or `hotfix(GH-123): :firetruck: ...` can also initiate the trusted hotfix path after source review and merge.

With multiple environments, configure `default_environment` for title-driven requests. Mint prepares isolated source against that environment's verified deployment rather than pulling in unrelated queued changes. Review the source correction and then the release selection; the hotfix still needs build and deployment verification.

The version is the next unused patch in the deployed compatibility series. Integrate the correction into the default branch so future ordinary releases retain it.

## Reconcile an uncertain deployment

Missing, mixed, or conflicting deployment evidence leaves a fence in place. A successful dispatch or standalone observation does not clear it.

1. Establish that the original adapter run has completed.
2. Request **observe** and obtain a trusted successful observation taken after that completion.
3. Request **reconcile** using that observer run and the active intent.
4. Mint compares the full observed identity with the frozen selection and previous verified baseline.

An exact match to the frozen selection establishes success. An exact match to the previous baseline establishes a failed deployment with unchanged runtime. Any other state retains the fence for investigation. All bundle members and configuration identity must match.

Successful reconciliation also retries pending publication. Replaying reconciliation may retry publication without redeployment. Reviewed and automatic unpaused environments restore a reviewed PR for queued work after explicit recovery; manual environments still need promotion.

## Read status without overclaiming

```bash
mint deployment status --environment sandbox --format markdown
```

Read desired policy, last verified deployment, and live observation separately. `live`, `drift`, `deploying`, and `unknown` describe authenticated evidence; an observation timestamp limits what it proves. Do not substitute an old verified version for a fresh runtime observation.
