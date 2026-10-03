# Reviewed production releases

Source versioning and production publication have separate lifecycles. `release
version-main` tags every new first-parent main commit, including control commits,
with strict SemVer and patch fallback. It never creates a GitHub Release. Run it
under a repository-wide versioning concurrency group after fetching authoritative
main and tags without force. Use `--target-commitish "$GITHUB_SHA"` to stamp the
artifact from that exact source even when newer main commits have arrived.

`release production` uses `.mint.yaml` repository policy and an isolated
`mint-release-state` branch. Git Data commits use configured human authorship and
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

Provision a narrowly scoped, approved human-owned `MINT_RELEASE_TOKEN`; no bot
identity is allowed by the current contract. It needs these repositories only:
contents, issues, pull requests and Actions read/write. Native CI uses the ordinary
read-only GitHub token. The token owner, configured Git identity and trusted event
actor are checked before mutations. No credential is created by Mint.

Import a baseline once using `bootstrap --input baseline.json --run-id ID`. An
operator must first reconcile the currently running artifact/configuration and its
successful deployment evidence; latest tags or arbitrary successful builds do not
prove production. The imported candidate must identify exact source tag, digest,
configuration hash and build. Never re-import to erase a failed/unknown intent.

## Adapter sequence

1. Source build: exact checkout; `version-main` (or reviewed `version-hotfix`);
   build once; persist immutable ECR digest or versioned S3 bundle; upload one
   `mint-candidate` Actions artifact containing only `mint-candidate.json`.
2. Trusted build completion: `candidate --run-id ID`; `scan` recovers missed main
   build events; `propose` updates the single accumulating normal production PR.
   Commands authenticate the producer run, unique archive digest and strict JSON.
3. Proposal close: `propose --event close` pauses it. Reopen: `--event reopen`
   resumes the same PR, refreshes latest selection and clears a prior pin. Preserve
   `.mint/summary.md`; `--pin SOURCE_SHA` makes explicit selection sticky until
   reopen. Only `CHANGELOG.md` and named `.mint` control files can change.
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
   keep the lock. Failed/cancelled runs keep the previous production baseline.
7. `publish --intent-id ID` publishes canonical notes and `mint-production.json`
   only after success is durable. A publication retry cannot redeploy. Conflicting
   existing tags, notes or manifests fail closed. `status-pr --intent-id ID` opens
   a separately reviewed changelog/status update, including failed outcomes.

A candidate manifest uses the JSON fields defined by `promotion.Candidate`; its
`artifact` contains `reference`, `digest` and `configuration_sha256`. A deployment
manifest contains `intent_id`, `source_sha`, `artifact_digest`,
`configuration_sha256`, and `verified: true`. Only successful authenticated
producer runs can supply these manifests. Durable ECR/S3 objects must be retained
for the rollback window and verified before promotion; Actions artifacts are
transport evidence, not the durable application bundle.

## Hotfixes and rollback

An approved issue-form request is converted by the repository workflow into typed
JSON `{ "Issue": 123, "BaselineID": "...", "Fixes": ["full SHA"] }`. Verify issue
ownership/assignment; `prepare-hotfix --input request.json` creates GH-123 from
current production, cherry-picks only explicit single-parent fixes and opens a
human-owned source PR against its immutable production base branch. Empty fixes
prepare a branch for newly authored code; metadata alone cannot build a candidate.
Interrupted requests recover the same branch and PR. Conflicts preserve the local
checkout for resolution. Never include queued main changes to repair a conflict.

Merge/review the hotfix source separately. `version-hotfix --pr NUMBER` allocates
an unused patch tag in production's deployed major/minor series, skipping main tag
collisions. Build that merged source once and register its `kind: hotfix` candidate
with the current `baseline_id` and `source_pr`. `propose --kind hotfix` makes the
separate hotfix deployment proposal while the normal queue remains intact.
Stale baseline, extra cherry-picked changes and concurrent promotion fail closed.
Original patch IDs and source versions retain provenance. Later ordinary proposals
require forward integration of every shipped fix (or an explicit actual inverse
patch); equivalent fixes are not repeated in shipped notes. Evolved overlapping
patches require reviewed reconciliation when exact Git proof cannot establish it.

`propose-rollback --pin PREVIOUS_SOURCE_SHA --summary "Reason"` selects only a
previously verified production artifact and creates a separately reviewed
control proposal. Normal freeze/start/finish apply. Successful rollback restores
that version's shipped set and source baseline as a new history event, without
retagging or another GitHub Release. Manual infrastructure/maintenance changes
outside this lifecycle require separate operational authorization and reconciliation.
