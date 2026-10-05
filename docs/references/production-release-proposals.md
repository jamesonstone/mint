# Reviewed production releases

Source versioning and production publication have separate lifecycles. `release
version-main` tags every new first-parent main commit, including control commits,
with strict SemVer and patch fallback. It never creates a GitHub Release. Run it
under a repository-wide versioning concurrency group after fetching authoritative
main and tags without force. Use `--target-commitish "$GITHUB_SHA"` to stamp the
artifact from that exact source even when newer main commits have arrived.

`release production` uses `.mint.yaml` repository policy and an isolated
`mint-release-state` branch. Generated Git Data commits use the explicit GitHub Actions identity and
non-force compare-and-swap ref updates. Failed writes report a conflict; replay
the operation from a freshly fetched repository, never force or reuse stale state.
The journal retains candidates, immutable merged intents, verified production
baselines and rollback history after Actions artifacts expire. Protect the journal
against deletion and history rewrites. Artifact storage remains application-owned.

## Activation

Use an immutable, **published** Mint feature release SHA in every application
workflow. The previous v0.2.1 release does not contain these commands. Source code
in a ready Mint PR is not a published upgrade. Keep application activation disabled
until the upstream release exists, approvals are granted, and adapters are tested.

Use job-scoped `GITHUB_TOKEN`, exposed as `GH_TOKEN` for Mint and gh. Installation
credentials are checked against the configured repository, never `GET /user`.
Generated release controls allow only the designated human or `github-actions[bot]`;
application source authorship and independent human review remain governed.
The control job needs contents/issues/pull-requests/Actions write and checks read.
Version/build jobs need contents write and Actions/pull-requests/checks read.
Intent freezing needs contents write and Actions/pull-requests/checks read; native
CI remains read-only. No PAT, App or other persistent credential is required.
Enable the repository setting permitting Actions PR creation before activation;
Mint neither modifies that setting nor approves/merges its generated PRs.
Explicit CI dispatch validates the exact proposal head. GitHub may also create
approval-required runs for token-created PR events; those are not assumed passing.

Import a baseline once using `bootstrap --run-id ID`. An
operator must first reconcile the currently running artifact/configuration and its
successful deployment evidence; latest tags or arbitrary successful builds do not
prove production. The configured legacy baseline workflow must upload one digest-attested `mint-baseline` archive after live verification; source-build identity is separately checked against `baseline_build_workflow` (defaults to `build_workflow`). These baseline workflow settings are needed only when importing production, not for routine release or recovery commands. A local JSON file is never bootstrap evidence. The imported candidate must identify exact source tag, digest,
configuration hash and build. Never re-import to erase a failed/unknown intent.

## Adapter sequence

1. Source build: exact checkout; `version-main` (or reviewed `version-hotfix`);
   build once; persist an immutable artifact reference and content digest (for example, a container image, binary, or versioned bundle); upload one
   `mint-candidate` Actions artifact containing only `mint-candidate.json`.
2. Trusted build completion: `candidate --run-id ID`; `scan` recovers missed main
   build events; `propose` updates the single accumulating normal production PR.
   Commands authenticate the producer run, unique archive digest and strict JSON.
3. Proposal close: `propose --event close` pauses it. Reopen: `--event reopen`
   resumes the same PR, refreshes latest selection and clears a prior pin. Preserve
   `.mint/summary.md`. Normal proposals always track the latest eligible candidate;
   sticky normal pins are rejected. Hotfix selections remain explicit. Only `CHANGELOG.md` and named `.mint` control files can change.
4. Explicitly dispatch native CI with `mint_pr` for token-authored proposal heads.
   `validate-review --pr NUMBER --token-env GITHUB_TOKEN` verifies declarations
   without requiring approval before CI. Freeze enforces every configured required
   successful current-head check and independent current-head review.
5. Merged proposal: `intent --pr NUMBER --merge-sha SHA`; `start --intent-id ID
   --run-id ID`. Read the declaration at the exact merge SHA; never select latest
   main or rebuild. One production lock covers normal, hotfix and rollback.
6. Adapter promotes and verifies exact digest, configuration and running source;
   upload only `mint-deployment.json` as `mint-deployment`. A separate trusted
   completion workflow calls `finish --intent-id ID --run-id ID`. Unknown outcomes
   keep the lock. Failed/cancelled runs retain the deployment fence unless a digest-attested `mint-deployment` manifest verifies the unchanged prior baseline with `outcome: unchanged`; only then is a failed attempt safely finalized.
7. `publish --intent-id ID` publishes canonical notes and `mint-production.json`
   only after success is durable. A publication retry cannot redeploy. Conflicting
   existing tags, notes or manifests fail closed. `report --intent-id ID` attaches an idempotent machine-owned outcome comment
   to the original merged release PR, including failed outcomes. No status PR,
   branch or direct default-branch commit is created. `status-pr` remains a
   deprecated compatibility alias with this same behavior.

A candidate manifest uses the JSON fields defined by `promotion.Candidate`; its
`artifact` contains `reference`, `digest` and `configuration_sha256`. A deployment
manifest contains `intent_id`, `source_sha`, `artifact_digest`,
`configuration_sha256`, and `verified: true`. Only successful authenticated
producer runs can supply these manifests. Durable application artifacts must be retained
for the rollback window and verified before promotion; Actions artifacts are
transport evidence, not the durable application bundle.

## Repository Actions: hotfix and rollback

The routine user interface is the repository's **Production release control**
workflow. Mint executes underneath it; operators and agents need no local Mint
installation, baseline IDs, full SHAs, or JSON manifests.

- **Hotfix:** choose `hotfix`, supply a merged fix PR number or an issue number
  for a newly authored fix, and a reason. Exactly one source input is required.
- **Automatic hotfix:** merge a reviewed same-repository main PR titled
  `hotfix: ...`, `hotfix(component): ...`, or `hotfix(GH-123): :firetruck: ...`. The trusted controller fetches its server
  identity and starts preparation. The emoji is display text, not routing.
- **Rollback:** choose `rollback`, enter a reason, and optionally select a strict
  deployed version. Blank target selects the previous distinct verified deployed
  artifact through deployment ancestry, never tag or build order. Legacy history
  without ancestry requires an explicit known deployed version.

Prefer roll-forward: repair or revert the offending code as a new hotfix and
release a new version. Rollback remains supported when restoring a retained
artifact is the appropriate recovery. Neither operation reverses database/data
changes. Adapter configuration compatibility and artifact/runtime verification
remain mandatory. Requesting preparation never authorizes merge or deployment.

### Maintainer installation

Generate the recovery entry point using a published feature-bearing immutable
Mint SHA, then deliver it through the project's normal review process:

```sh
mint release production workflow --mint-ref "$PUBLISHED_MINT_SHA" \
  --output .github/workflows/mint-recovery.yaml
```

That is the default trusted controller path. A custom filename must also set
`control_workflow` to its exact `.github/workflows/...` path in `.mint.yaml`.
The Action's `production-control` command authenticates the active default-branch
workflow, configured human operator, repository, and event against GitHub.
It reads request inputs as data; it never executes issue/PR text as code.
The generated workflow defaults to hotfix and remains behind
`MINT_RELEASE_ENABLED`. Installing it does not activate production or change
repository settings.

The recovery workflow complements the existing application candidate and
production completion controllers. Those adapters must dispatch the default-branch
producer with `hotfix_pr` when a registered production-base source PR merges,
then register the exact candidate and reconcile its production PR. Keep this one
owner for source completion to avoid duplicate dispatches. This generated entry
point does not replace application builds, deployments, or completion handlers.

### Hotfix isolation and review

`hotfix --fix-pr NUMBER --reason "Reason"` is the equivalent CLI request.
It verifies merged same-repository default-branch identity, required checks and
independent source approval, and resolves squash, merge or rebase histories to
exact reviewed single-parent fixes. Fetch full main history and the original PR
head. Ambiguous/non-linear or altered merge resolutions require an isolated
reviewed source fix, not guessed main changes.

Mint allocates a separate human-assigned tracking issue and GH issue branch;
it never overwrites the original fix PR's branch. Registered request markers
bind PR number, merge SHA and production baseline, making manual/prefix replays
converge. A stale registered baseline stops with an action to inspect the request.
`hotfix --issue NUMBER --reason "Reason"` prepares a production-base source lane
for newly authored code; the issue must be open, human-owned and assigned.

The source PR is checked and reviewed against an immutable production base.
Merging it causes the application adapter to build that exact source and make
one hotfix production approval PR. Its human merge promotes the artifact.
The normal queue and manual summary remain intact. Hotfix candidate baselines,
shared production locking and patch provenance prevent queued-main inclusion
and an ordinary release undoing a shipped fix without explicit reconciliation.
Newly authored fixes still need reviewed forward integration into main.
Cherry-pick conflicts publish a metadata-only recoverable source branch/PR;
CI rejects it until all requested application fixes are present. Temporary
runner paths are not the sole recovery surface.

`prepare-hotfix --input request.json` remains a lower-level adapter accepting
`Issue`, `BaselineID`, and `Fixes`. Prefer the PR/issue request surface for users.
`version-hotfix --pr NUMBER` allocates an unused patch in production's deployed
major/minor series. Generated source merges do not recursively trigger requests.

### Rollback review and replay

`rollback --reason "Reason" [--to vX.Y.Z]` is the equivalent CLI request.
It creates one reviewed rollback proposal showing current and target versions,
reason, artifact identity, a roll-forward recommendation, and compatibility/data
limits. Low-level `propose-rollback --pin SOURCE_SHA --summary "Reason"` remains
available. Replaying non-open proposals never modifies their merged branches.
Actions rollback runs persist their initial selection; rerunning the same run
cannot resolve a new "previous" version and undo a completed rollback.

Normal freeze/start/finish gates apply. Successful rollback restores verified
source/artifact contents as a new history event without retagging or publishing
another GitHub Release. The outcome goes on the same release PR. Artifact storage
must retain rollback objects for the configured recovery window. Missing objects,
ambiguous history, stale baselines and unknown runtime evidence fail closed with
an explicit recovery action. Manual infrastructure changes require their own
operational authorization and reconciliation.

## Simple policy and command setup

`control_paths` may be omitted: Mint uses `CHANGELOG.md`, `.mint/proposal.json`, and `.mint/summary.md`, the exact files it generates. An explicit list must include all three; historical `.mint/status.json` remains allowed. Partial, duplicate, or unsafe paths fail before use. Policy is one strict YAML document; YAML merge keys are unsupported.

`baseline_workflow` is required for `bootstrap` only. `baseline_build_workflow` defaults to the configured producer, and may explicitly identify a legacy producer during migration. This default removes repetition without treating an ordinary successful build as proof of deployed production.

For a custom policy file, generate with `--config config/release.yaml`. The generated Action passes that same repository-relative path to Mint. With `--output`, the destination must equal `control_workflow` (or `.github/workflows/mint-recovery.yaml` when omitted); Mint creates missing parent directories. The workflow must be a `.yml` or `.yaml` file directly under `.github/workflows`. Config paths cannot be absolute, escape the repository, or contain Actions expressions. Commit both policy and workflow before enabling them.

Everyday `production --help` shows `status`, `hotfix`, `rollback`, and `workflow`. Existing adapter operations remain callable by name and expose their own relevant options; use `production <operation> --help` for integration work. Invalid options fail rather than being silently ignored. `status` still returns the full authoritative JSON journal for compatibility.
