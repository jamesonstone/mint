# Mint simplification audit — GH-15

Mint's production workflow is described in two paragraphs in the [README](../../README.md#reviewed-production-releases). The generic integration contract is an immutable artifact reference, content digest, configuration digest, source/build identity, and a project-owned deployment verifier. The engine does not build or deploy the application. Container images, binaries and packaged bundles use the same contract.

## Research and selected changes

The audit followed every production command and its callers, policy loading, workflow rendering, candidate scan, journal updates, proposal creation, review checks, hotfix provenance and rollback ancestry. All findings below come from source inspection; GitHub behavior was checked against primary documentation.

| Finding | Implemented simplification | Evidence and boundary |
| --- | --- | --- |
| Every operation exposed all flags; several were irrelevant or unused. | Four visible operator commands; adapters remain callable with only consumed options. Removed unused `--outcome`. | CLI rejects cross-command options before policy/network use. Existing application adapter invocations remain valid. |
| Routine commands required one-time migration settings; omitted control paths could reject Mint's own PR. | Bootstrap-specific validation; producer default for baseline build; complete generated-file default and strict explicit policies. | Missing bootstrap verifier still fails before importing evidence. Explicit unsafe or partial paths fail. |
| Generated workflow discarded custom config and could be written to a path its own authentication rejected. | Preserve repository-relative config; require configured output destination; create directories. | Generated YAML is parsed and the Action input is checked; expressions/path traversal are refused. |
| Automatic hotfix intent required Kit-specific GH issue syntax. | Accept conventional `hotfix:`, `hotfix(scope):`, and existing GH scopes, including breaking markers. | Authoritative merged PR and independent review establish source identity; title supplies intent only. |
| Scan downloaded the same evidence twice and repeatedly searched candidate history. | One shared registrar and one archive download per new run; indexed known run IDs. | Source/tag, run, archive digest, artifact, hotfix review and replay-conflict checks remain. Known historical runs do not need expired artifacts. |
| Authored hotfix isolation referenced `origin/main` even with a custom default branch. | Use the configured default branch for queued-history proof. | A `trunk` fixture accepts an isolated fix and rejects a queued commit; missing branch evidence fails. |
| Exact workflow path comparisons rejected GitHub's documented `path@ref` representation. | Accept optional nonempty ref qualification while matching the full configured filename. | Repository, head repository, event, status and source fences remain independent. |

GitHub's [workflow-run API](https://docs.github.com/en/rest/actions/workflow-runs) documents run identities and a qualified path example. [Workflow trigger guidance](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/trigger-a-workflow) explains token-trigger behavior and dispatch: keep explicit exact-head checks rather than assuming generated PRs automatically have passing CI. [Rerun guidance](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/re-run-workflows-and-jobs) preserves the original run's source and actor privileges; durable rollback run identity must remain. These are transport facts, not permission to deploy or merge.

## Valuable opportunities deferred

- **Optional tracking issues and deterministic Mint branches:** potentially removes GitHub issue allocation and history scans, but requires deliberate identity migration, interrupted-creation recovery and collision tests. Current source/release branch provenance depends on GH lanes. Merely renaming branches would break recovery guarantees.
- **Several allowed recovery operators:** removes single-person operational coupling, but authorization and assignment currently share one identity. Separate those policies explicitly before changing access. This audit does not broaden permissions.
- **No-op journal and dispatch suppression:** reduces commits and runs, but persistence and CI dispatch can fail independently. A durable dispatch recovery contract is needed before suppressing retries; otherwise a saved proposal can lose required validation.
- **Known-PR lookup before history discovery:** promising transport optimization; retain interrupted-creation discovery and strict marker/repository/author checks. Deferred to keep this change focused on the operator/setup and candidate evidence paths.
- **GitHub Enterprise URL support and compact status output:** honest host-derived release links/upload endpoints and a derived human status view warrant separate compatibility tests. Raw journal JSON remains unchanged here.

Keep the journal and separate build, reviewed selection, frozen intent, runtime verification and publication evidence. They prove different events. Combining them into one mutable record would simplify names while obscuring correctness. Existing schema, publication retries, production lock and application deployment ownership remain intact.

## Scope and adoption

This PR builds on the Actions-first production feature in #14. It changes no repository setting, cloud resource, credential, deployed artifact, application adapter, or Kit distribution. Adoption still requires a published feature-bearing Mint revision and project activation checks. This research is not deployment evidence.
