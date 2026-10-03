---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0009"
  slug: "production-release-proposals"
  dir: "0009-production-release-proposals"
---
# SPEC

## PURPOSE

Implement the approved Mint-managed production proposal lifecycle, including isolated hotfixes.

## CONTEXT

This is a dependent cross-repository implementation. Mint issue #13 owns the engine; LabCore #1131 and LabCore UI #823 own artifact and deployment adapters. The approved specification below is authoritative for this feature. Coordinator: Mint docs/programs/production-release-proposals/PROGRAM.md.

## REQUIREMENTS

# Mint-managed versioning and production release proposals

Status: implementation specification and handoff, 2026-10-03.
Scope: jamesonstone/mint, lsmc-bio/labcore, lsmc-bio/labcore-ui.

## 1. Objective and agreed product behavior

Give every new main-branch commit a deterministic semantic version and immutable Git tag, while reserving new published GitHub Releases in the application repositories for versions that successfully reach production. Keep the entire flow automated and code-driven. The team should initiate normal production releases by reviewing and merging an automatically maintained release PR, without local CLI setup or choosing workflow_dispatch inputs in the Actions UI.

Mint owns release state, deterministic SemVer, tag creation, candidate/proposal reconciliation, notes rendering, and GitHub Release publication. Application repositories own build artifacts, registry/storage operations, infrastructure deployment, and environment-specific verification. GitHub Actions supplies the execution environment; the CLI remains available for programmatic integration and recovery.

No release, deployment, PR update, or operational state mutation has been executed by the authoring side conversation. This file is a specification, not implementation evidence.

## 2. Source observations to verify before implementation

- Mint at /Users/jamesonstone/go/src/github.com/jamesonstone/mint has Cobra commands under pkg/cli and reusable release logic under pkg/release. It already supports release resolve, select-tag, tag, github, publish, changelog, and workflow generation. Keep backward compatibility with those commands unless a documented migration is necessary.
- LabCore .github/workflows/main.yaml responds to main pushes, calls mint release publish at a pinned Mint revision before the image build completes, then builds/pushes a versioned ECR image. .github/workflows/deploy.yaml responds to completion of main and has existing manual dispatch inputs, deployment safeguards, and immutable deployment manifests.
- LabCore UI .github/workflows/main.yaml builds and deploys on every main push, embedding the commit SHA; it does not currently implement the proposed named production-release lifecycle.
- Application workflow integration must replace automatic production deployment on normal main pushes with release-proposal promotion, while preserving existing checks, protected activation settings, identity verification, operational safety, maintenance support, and rollback mechanisms.
- Refresh these facts from the current default branches. Do not copy stale workflow fragments verbatim.

## 3. Vocabulary and evidence boundaries

Version: strict SemVer tag attached immutably to one source SHA. A version is not proof that a build or deployment succeeded.
Candidate: version with a successful, verifiable deployment artifact and required validation evidence.
Proposal: open release PR selecting a candidate for production.
Release intent: the validated, immutable selection read from a merged proposal at its merge SHA.
Production deployment: the deployment operation and its observed outcome.
Published GitHub Release: the production version record published only after deployment verification succeeds.
Product acceptance: separate application/physical/outbound evidence, never implied by infrastructure health.

A successful source version may exist without a candidate; a candidate may never be deployed. Failed or paused production attempts must not be represented as successful GitHub Releases.

## 4. User experience

1. A normal feature/fix PR merges to main.
2. Version automation assigns its version/tag and prepares its artifact.
3. After the artifact and checks are eligible, Mint opens one ready PR titled Release to production: vX.Y.Z, or updates the existing one.
4. The PR displays current deployed version, proposed version, included PR count, short human-editable summary, artifact identity, validation links, and generated cumulative notes. The diff contains only approved release-control files and CHANGELOG.md.
5. The team leaves the PR open to accumulate changes. Mint keeps it current until merge. Target changes require fresh checks and stale approvals must not silently authorize a new target.
6. Merging deploys the exact version selected in that merge, even if a newer main push arrives afterward.
7. After production verification, publish a GitHub Release on that existing tag, with the same notes and linked deployment evidence.
8. The next application candidate starts a new proposal cycle.

Closing an unmerged proposal pauses it. Subsequent pushes must not reopen it or create a replacement automatically. Reopening immediately catches it up to the latest eligible candidate and regenerates notes, preserving manual summary content. Reopening never deploys. Keep the paused head branch recoverable; automatic delete-branch policies and recovery behavior must be accounted for.

Provide a discoverable Request production release issue form as a recovery entry point when no proposal exists. It must not bypass a paused proposal or create a competing one; when paused, guide the requester to reopen it. Creating an issue is proposal intent only, not deployment authorization. Restrict mutation handling to trusted repository principals; an arbitrary issue label or PR title is never authorization.

## 5. State machine and durable identity

States: idle -> open -> paused -> open; open -> merged/deploying -> deployed or deployment_failed. A successful deployment advances the production baseline. A failed deployment leaves the prior successful baseline unchanged.

Use one active or paused proposal per repository/environment. Identify it with a stable machine marker, repository identity, environment, proposal generation, and recorded PR/head branch identity, rather than title parsing. Search and reconcile both open and closed-unmerged proposals. Merged historical proposals never count as paused.

Use durable, authenticated records for candidate identity, successful production identity, and proposal control state. Implement the smallest mechanism supported by repository/GitHub facilities; do not introduce a database service. Records must survive runner loss, artifact retention expiry where the record is required long term, duplicate events, and ordinary reruns. Do not rely solely on an Actions artifact with short retention as the lasting production baseline.

When a deployment is in flight, collect newer candidates but do not finalize a competing production range until its outcome is reconciled. Serialize production promotions per environment; serialize proposal mutations per environment. Do not cancel an already started deployment merely because a newer main commit exists.

On out-of-order events, never regress an automatically tracked proposal to an older candidate. Recompute from authoritative candidates and ancestry rather than treating the triggering event as necessarily latest.

## 6. Version and candidate generation

- Version each new main SHA, including non-feature/control commits if adhering literally to every push; replays of the same SHA reuse the version/tag.
- Version increments remain conventional-commit driven; establish and test a patch fallback when an ordinary commit would otherwise yield no bump. Never silently omit a requested main version.
- Consider main first-parent ordering, batched pushes, serialized jobs, and the existing tag baseline so two simultaneous events cannot allocate competing versions. Process intervening main SHAs when a push includes several commits; squash merging is the preferred simple PR-to-version mapping.
- Tags cannot be moved or reused for different SHAs. On a conflicting existing tag, fail with actionable evidence rather than force-push.
- Version creation must not publish a GitHub Release.
- Publish candidate metadata only after successful build/validation. It includes repository, source SHA, version, workflow run, immutable artifact reference/hash, and build configuration identity.
- Release-control-only commits receive version identity as required but must not recursively generate new release proposals. Their changes do not become artificial user-facing changelog items. Define and test a trusted path/content classification; do not trust a title or skip marker to suppress arbitrary application changes.
- A mixed application/control change is not a control-only commit and must be handled explicitly, not silently excluded.

## 7. Release declaration and editable summary

Adopt a documented repository-local configuration such as .mint.yaml and a small proposal declaration under an agreed release-control directory. Names/schema are illustrative and should follow existing Mint conventions.

Example declaration:

schema_version: 1
environment: production
version: v0.142.3
source_sha: <full SHA>
artifact: <immutable image/bundle identity>
previous_production_version: v0.140.2
selection: latest

Normal mode tracks the latest eligible version. If a team member explicitly selects an earlier version, represent that as selection: pinned so automation does not silently override their choice. On reopening a paused PR, the agreed default is to catch up to latest; document and visibly report any reset of a prior pin. Reject missing, unbuilt, foreign-repository, untrusted, or incompatible selections. Production attempts must validate source/tag/artifact correspondence.

Store summary separately from generated content, or use strictly delimited manual/generated blocks. Preserve manual text on ordinary updates, reopen, reruns, and recovery. Generate a deterministic factual fallback summary from PR metadata if no curated text exists; do not require an LLM or introduce secrets/services for summaries.

## 8. Changelog and GitHub Release content

For each selected release, render one H2 heading with selected version/date, followed by a short summary and one flat list of included merged PRs. Each entry includes its merged version linked to the GitHub tag/tree and its PR number/title linked to the PR.

Example:

## v0.142.3 — 2026-10-03

Improves receiving defaults and package-content guidance.

- [v0.142.1](https://github.com/OWNER/REPO/tree/v0.142.1) — [#810: Preserve matched-order counts](https://github.com/OWNER/REPO/pull/810)
- [v0.142.2](https://github.com/OWNER/REPO/tree/v0.142.2) — [#814: Numeric tube counts](https://github.com/OWNER/REPO/pull/814)
- [v0.142.3](https://github.com/OWNER/REPO/tree/v0.142.3) — [#821: Clarify contents warning](https://github.com/OWNER/REPO/pull/821)

Range is previous successful production SHA exclusive to selected candidate SHA inclusive. Do not use previous SemVer tag, latest main, latest arbitrary GitHub Release, or previous attempted production release as the baseline. Resolve PR/version mapping authoritatively, deduplicate PRs, paginate APIs, and escape Markdown as needed. Preserve revert PRs visibly; do not invent net-change cancellation semantics. For direct commits with no PR, include an explicit linked commit/version entry so changes cannot vanish. Define behavior for rebase/merge strategies with several versioned commits per PR.

CHANGELOG.md and GitHub Release share the same canonical entry. The merged proposal entry must be explicitly marked production deployment pending until verified. On failure keep it explicitly failed/pending, without advancing successful production state. Success updates the status through normal reviewed repository delivery where required, using control-only automation that cannot redeploy or recursively open proposals. Prefer keeping stable note content plus a small status marker to minimize churn.

Publishing notes on an already existing source tag does not require modifying that tagged source commit. The latest main branch may contain the release documentation while the deployed source tag remains unchanged. Document this distinction.

## 9. Deployment and publication contract

On a validated merged proposal, read its declaration and summary from the exact merge SHA and freeze the resulting intent. Never fetch the moving PR branch or latest main to determine the deployed target. Independently resolve tag SHA and artifact identity and verify them against the declaration.

Reuse already built artifacts where possible. LabCore deploys an ECR image by digest. LabCore UI stores an immutable exported static bundle and records its checksum, source SHA, and public configuration identity; production promotes that exact bundle. Do not recompile with changing dependencies/configuration during promotion and claim it is the same artifact. If a build must occur at promotion, use pinned inputs and record a distinct build identity.

Keep existing app-owned deployment verification. Publish the GitHub Release only after the exact deployment is observed successful. Attach release manifest and artifact/deployment links; verify existing tag/Release targets on retries. No new GitHub Releases per merge; no prerelease objects used as a workaround for candidate inventory.

If deployment succeeds but GitHub Release publication fails, record deployed/publication_pending and retry publication without redeploying. If deployment fails or verification is unknown, do not publish a successful Release. Retries must be idempotent and validate the same intent/artifacts. Rollback to an older already released version is a new deployment attempt/history entry, not a second GitHub Release for the same tag or a forward-version advancement. Keep rollback mechanics intact and do not retag source.

## 10. Authentication, events, and trust

Use narrow job permissions and existing approved credentials. Respect human attribution and repository delivery policies. Do not introduce broader credentials, change branch protection, weaken required checks, or auto-approve proposals to make automation work.

GitHub GITHUB_TOKEN-generated PR/commit/tag events do not generally trigger further Actions workflows. Explicitly design proposal validation/event chaining with reusable workflows or supported dispatch mechanisms, or an approved existing GitHub App token. Do not assume an automated PR push automatically launches required CI. Any required new credential/access is a named activation dependency, not permission to silently provision it.

Reopened events must reconcile closed proposal state, even if no new main push follows. Guard synchronize/update event handlers against self-triggering loops. Use immutable workflow revisions for downstream installations; pin the latest verified Mint release to its exact commit in LabCore and UI, rather than runtime @latest. Document the human-readable Mint version alongside the pin.

## 11. Migration and activation

- Implement and validate Mint first, including backward-compatible CLI/Action exports and documentation.
- Integrate both apps against that exact tested Mint revision; adopt the newest published Mint release containing the feature when available. Do not claim latest integration if only an unpublished local binary is tested.
- Establish baseline from verified current production source/artifact identity. Existing historical GitHub Releases may correspond to every merge, so latest Release alone is not evidence of deployed baseline.
- Preserve historical tags and Releases; do not delete or relabel them automatically. Apply production-only publication to new releases and document the legacy boundary.
- Preserve current production resources and settings. Disable ordinary-main automatic production deployment as part of the reviewed workflow migration, while keeping version/build automation active.
- Preserve existing maintenance and explicit recovery workflows; route future ordinary promotion through the new verified declaration contract.
- Validate first with synthetic fixtures and mocked APIs. Avoid live deployment or persistent operational-data changes for testing without their authorization.
- Implementation/PR authorization does not by itself waive repository merge, cloud mutation, credential, or protected activation approval boundaries. Deliver ready PRs and concrete activation prerequisites, then respect those boundaries.

## 12. Acceptance and validation

Mint tests must cover deterministic bumps; same-SHA replay; immutable tag collision; batch/concurrent/out-of-order main events; successful-build eligibility; one proposal only; cumulative PR/version/link mapping; previous production range; direct commits; pagination; manual summary preservation; manual pinned selection; close/pause with several intervening merges; reopen/catch-up without another push; missing/deleted branch recovery; duplicates; API partial failure; release-control recursion; mixed changes; untrusted triggers; exact merged-declaration pinning; in-flight deployment; failure baseline preservation; successful deploy/publication retry; existing Release target conflict; rollback semantics; GitHub token event limitations.

App workflow tests must prove that a main feature merge creates version/candidate/proposal but performs no production deployment or GitHub Release publication; a merged release proposal deploys exactly its selected artifact; closed/reopened proposals behave as specified; production failure does not publish; required CI genuinely runs against updated proposal heads; existing safety/activation and maintenance contracts remain valid.

Run Mint native tests/build/vet and relevant repo checks. Run LabCore native/contract/workflow validation and LabCore UI tests/typecheck/lint/format/file-length/web-build plus new workflow contract checks. Report existing unrelated failures literally without expanding scope unnecessarily. Inspect generated notes/PR content with a synthetic fixture spanning at least three versions and PRs.

Required end-state report: spec locations, issues/branches/ready PRs for all three repos, exact Mint version/SHA adopted, local and hosted check states, migration baseline evidence or blocker, and separate implementation/merge/deployment/publication/product-acceptance states. Do not call production activation complete on source checks alone.

## 13. Delivery ownership and implementation sequence

Parent chat owns implementation across all three repositories. Preserve primary checkouts and existing work; use governed per-repo issue/branch/worktree lanes and read current repository instructions. Adopt this specification into canonical repo specs before implementation, with a coordinating cross-repository plan and explicit dependency links. No subagent use is requested by this handoff.

Sequence: (1) readiness audit and durable specs; (2) Mint engine/CLI/Action and tests; (3) latest feature-bearing Mint release availability and exact downstream pin; (4) LabCore workflow/artifact/proposal integration; (5) UI immutable bundle/workflow/proposal integration; (6) synthetic complete lifecycle verification and ready PR delivery; (7) authorized merge/activation and production evidence, when granted.

## 14. Hotfix escape hatch (additional user requirement, 2026-10-03)

The user explicitly requires a way to repair production immediately without deploying the queued changes represented by the ordinary Release to production PR. Implement this in Mint and both application workflow integrations, not merely as documentation.

### Product and team UX

A hotfix must start from the verified last successful production source tag/SHA, never latest main by default. A main candidate containing the fix also contains its ancestors, so selecting it cannot exclude earlier queued work. Only when the selected candidate's actual diff is exactly the desired fix is ordinary candidate pinning sufficient.

Provide a discoverable Request production hotfix issue form (or a comparably visible repository-native entry point) that selects production baseline and explicitly identifies the intended fix PR/commit(s), with short rationale. Mint prepares a narrowly scoped hotfix proposal/work branch using current repo delivery rules. Support a newly authored hotfix against the production baseline as well as cherry-picking an existing main fix. Normal protected review and checks still apply; no local CLI requirement and no bypass of required gates. Do not execute untrusted issue text as shell/code or trust an issue label as authorization.

Implement an explicit hotfix build/proposal flow that retains production-only publication rules. It must pin its source baseline, exact fixes, resulting source SHA/tag, immutable artifact, and previous successful production identity. Its review diff and generated notes show ONLY that hotfix's changes. Maintain human attribution, source PR links, and source-version provenance for a cherry-picked main fix. Resolve cherry-pick conflicts visibly in the governed work lane; do not mask conflicts or drag additional unrelated commits into production automatically.

### Isolation and release state

The ordinary main-tracking proposal remains open/paused with its queued changes; do not overwrite its branch, silently merge it, or erase its manual summary to create the hotfix. Distinguish proposal identity by kind: normal or hotfix. Permit the hotfix as an explicit separately scoped exception to the single normal open/paused proposal rule, but retain one active production deployment per repository/environment. Serialize hotfix versus normal promotions and fence every intent against the production baseline identity it was reviewed against.

Once the hotfix succeeds, update the durable production baseline and publish its production GitHub Release. Reconcile the normal proposal against the new production contents; require fresh validation/review. If the ordinary release is in flight when a hotfix is requested, do not interrupt or race it automatically: report the in-flight state, reconcile its outcome, and rebuild/review the hotfix from the resulting production baseline as necessary. If a stale proposal merges after the baseline changes, fail before deployment and request a reconciled proposal; never silently deploy a different version.

### Forward integration and changelog correctness

The fix must also be present in main before the next ordinary promotion, through its original main fix or a reviewed forward integration. Record logical provenance so equivalent cherry-picked changes can be recognized without assuming identical SHAs. Do not auto-drop queued changes or merge the rest of main into the hotfix branch. Prevent an ordinary release from undoing the hotfix; require explicit reconciliation when this cannot be proved.

Hotfix branches can diverge from main. The production SHA may therefore not be an ancestor of the next selected main SHA. A naïve previous-production..main commit range can repeat the already shipped hotfix or omit information. Extend note/range generation with explicit shipped PR/commit provenance, verified patch equivalence where appropriate, and branch-aware comparison. Do not silently infer equivalence from titles or hide a deliberate revert. If equivalence is ambiguous, fail the proposal check with actionable evidence. Each release notes entry reflects the changes newly introduced relative to actual production, preserving links to main PR/version and hotfix source/version as applicable.

### Version allocation

Mint must allocate an immutable unique patch version for the hotfix in the deployed compatibility series, handling collisions with existing per-main tags. A production patch increment may already exist as an unreleased main version and cannot be reused or retagged. Select the next unused valid patch in that series using an explicit tested policy and serialized allocation. Do not claim hotfix source includes features from a queued minor release merely to obtain the numerically largest global tag.

After a hotfix, candidate selection and production baseline logic must not rely solely on largest SemVer, latest created tag, or latest GitHub Release. Existing queued candidate versions may precede the new hotfix number while containing additional work. Document and test a coherent ordinary promotion/version policy that preserves tag immutability and provenance; if a fresh main candidate/version is needed, generate it through reviewed integration/control mechanisms rather than moving a tag. Challenge this policy in the readiness gate before implementation rather than leaving it to a release-time surprise.

### Required acceptance tests

1. Production at version A, queued main changes B and C, main hotfix D: hotfix artifact built from A plus only D, never B/C.
2. New hotfix authored against A without an existing main fix: normal promotion is blocked until verified forward integration.
3. Existing next-patch tag collision: deterministic unused patch allocation, no tag rewriting.
4. Hotfix succeeded: only its production Release is published, baseline advances, normal queue remains intact and is reconciled.
5. Subsequent ordinary release: notes do not count the logically equivalent hotfix twice; final artifact retains the fix and all intended queued work.
6. Conflicting cherry-pick or ambiguous patch equivalence: explicit failure and no production mutation.
7. Normal and hotfix proposals compete/in-flight/stale baseline: serialized deployment and fail-closed intent fencing, no silent target replacement.
8. Failed hotfix: no successful Release/baseline advancement; exact intent retry and queued normal state preserved.
9. Existing safeguards, review checks, trusted issue handling, immutable artifacts, and manual summary preservation remain enforced for both proposal kinds.

Parent implementation must adopt this addendum into Mint's canonical feature spec and both downstream specs before finishing the new workflow. The earlier statement of one proposal per environment is refined to one ordinary proposal plus an explicitly scoped hotfix proposal, with one serialized production promotion at a time.

## ACCEPTED PLAN

1. Adopt the complete specification and hotfix addendum in all three lanes.
2. Implement Mint deterministic versioning, durable state, proposal reconciliation, exact intents, production completion, and hotfix provenance with synthetic tests.
3. Deliver a ready Mint PR. Verify the feature-bearing published release and pin its exact revision downstream when available.
4. Integrate immutable app artifact building and promotion, preserving existing safety and recovery contracts.
5. Run native and workflow checks; deliver ready application PRs. Activation remains separately gated.

## DECISIONS

- Durable release metadata lives on an isolated Git journal branch with compare-and-swap updates; it does not need a database or depend on short-lived Actions artifacts.
- Generated release tags, journal/control commits and PRs, and verified Releases use the narrow GitHub Actions identity exception in section 15. Human source development, independent review, merge and protection remain required.
- Required proposal checks must be explicitly dispatched; GITHUB_TOKEN event suppression cannot be assumed away.
- The production baseline must be explicitly imported from verified deployment evidence. No automatic inference from the latest tag/Release.
- Hotfix selection must be built from production plus explicit fixes. Main ancestry cannot be sliced by pinning a main candidate.
- Same-series hotfix versions skip occupied patches; next main version allocation observes occupied tags without moving any tag. Candidate selection uses ancestry and provenance rather than numeric maximum alone.

## DISCOVERIES

- Current Mint release v0.2.1 does not contain this lifecycle. A published feature-bearing release is unavailable until the Mint PR is merged and its release workflow succeeds.
- Built-in job-scoped GITHUB_TOKEN is the default. Both repositories currently return can_approve_pull_request_reviews=false; Actions PR creation must be enabled separately before activation.

## VALIDATION

PENDING: implementation and synthetic lifecycle tests. No runtime or production validation has been performed.

## OUTCOME

IN_PROGRESS. Merge, credential provisioning, baseline import and production activation are not authorized by implementation scope.

## REPOSITORY MEMORY

This specification preserves the approved requirements and material design decisions. The coordinating program tracks dependent delivery evidence.

## Implementation checkpoint

GH-13 implements source versioning, authenticated run manifest registration,
CAS journal, paginated current-head checks, reviewed proposals and immutable
intents, deployment/publication retry separation, status PRs, isolated hotfix
source preparation/recovery and verified rollback history. Native Go tests, vet
and build pass. Full lint has pre-existing findings; changed-source lint is the
patch gate. Hosted GitHub event and cloud acceptance remain UNOBSERVED.

The feature is not yet a published Mint release. Downstream activation remains
held pending that exact published SHA, approved credentials, baseline import and
application adapter acceptance. No cloud changes, merge or production activation
were performed.

Integration validation also covers control-only merge commits, source-build
versus workflow-code identity for isolated hotfix builds, bootstrap archive and
independent build provenance, replay after baseline changes and ordered revert
accounting. A native pull-request CI workflow is included.


## 15. Simplification/refactor directive (supersedes human-token requirement, 2026-10-03)

User feedback: the personal-token activation dependency and supporting design add unnecessary complexity relative to existing deployment pipelines. User explicitly requests sending a refactor plan to the parent and simplifying the current design. Retain the agreed version/proposal/production-only Release/hotfix lifecycle; refactor its implementation and authentication, not the product requirements.

### A. Audit and preserve existing integration

Compare each proposed workflow against its original base-revision pipeline before editing. Produce a small before/after event/job map. LabCore's current proposal already reuses deploy.yaml through workflow_call; retain that reuse rather than replacing its deployment safeguards. Reuse the existing UI S3/CloudFront/OIDC deployment steps similarly. Existing app-owned build/scan, runtime configuration, account verification, protected activation, rolling deploy and deployment-verification logic should remain the authority. Do not introduce a second production deploy implementation or new AWS authentication path.

### B. Remove separate human PAT requirement

Replace secrets.MINT_RELEASE_TOKEN with the job-scoped built-in github.token/GITHUB_TOKEN for normal GitHub release automation, and set GH_TOKEN from that credential for gh. Mint should accept the supplied GitHub token independently of whether GET /user identifies a human; installation tokens cannot be treated as invalid merely because they are not personal user tokens.

This refactor intentionally allows the GitHub Actions automation identity for narrowly scoped generated version tags, release-state/control commits, release PR maintenance, and post-verification Releases. Preserve human authorship of application development, independent human approval/merge, branch protection and exact-target checks. Document this precise automation exception instead of fabricating human authorship or granting a blanket bot exemption to unrelated code changes. The previous spec's absolute human-token/bot prohibition for this release automation is superseded by this requested simplification.

Use least-privilege per-job permissions derived from actual API calls (contents, pull-requests, issues reads/writes only when necessary, actions only for reads or dispatch when needed). Ordinary PR checks remain read-only. No new PAT, App, persistent credential, broad repo scope, or security/protection weakening by default. Account for repository Actions settings governing PR creation; report any unavailable GitHub setting as a specific activation prerequisite.

### C. Make chaining explicit, without token event assumptions

Built-in token operations must not be assumed to trigger downstream workflows automatically. Keep native checks running on the exact generated proposal head through an explicit supported dispatch/reusable validation path where necessary. Use job dependencies, workflow_call, and exact producer workflow_run completion as appropriate. Avoid adding redundant event orchestration solely to mimic personal-token events. Check current GitHub behavior and actual repository settings; do not rely on an outdated categorical statement that all PR events are suppressed. Never execute untrusted PR code with the privileged release token. Retain safe trusted-default-branch control execution and unprivileged candidate validation.

### D. Small conceptual flow

1. Main source pipeline: version/tag -> existing build -> immutable candidate identity -> reconcile one release proposal.
2. PR lifecycle control: close pauses; reopen refreshes; generated/manual control updates validate without recursive rebuild/redeploy cycles.
3. Merged release proposal: read exact declaration -> existing app deployment of selected artifact -> existing verification -> production GitHub Release and baseline update.
4. Hotfix: production baseline plus explicit fix -> isolated reviewed source/candidate/proposal -> same deployment path. Ordinary queue remains intact and forward integration is enforced.

Keep CLI/API surface focused around these operations. Consolidate thin adapters/duplicate install/auth/manifest code where this materially reduces complexity. Do not split working deployment into additional workflows merely for naming symmetry. Keep versioning and production publication distinct.

### E. Minimize state without discarding essential evidence

Retain only durable state needed for candidate artifact identity, paused/open proposal identity, last verified production baseline, and in-flight/final promotion identity. Reuse existing deployment manifests and GitHub metadata rather than duplicate them into parallel attestations/journals. Simplify bootstrap using verified retained production deployment evidence; a previous tag still does not prove production. Do not remove immutable artifact verification, baseline fencing, idempotency, pause/reopen recovery, exact merged selection or concurrent hotfix/ordinary serialization for cosmetic brevity.

A protected state branch may remain if it is the simplest tested durable solution. Explain its purpose plainly rather than requiring extra infrastructure. Preserve existing historical releases without deletion/reclassification. No production activation or live data mutation is performed merely to test the refactor.

### F. Verification and delivery

Reuse exact existing Mint/application implementation lanes and PRs for this scope-preserving refactor; do not open coordinating/corrective duplicates. Update canonical specs/runbooks and deployment setup instructions, removing the human PAT prerequisite. Pin downstream applications to the latest published feature-bearing Mint version/exact revision when available; retain a clear dependency hold while upstream is unpublished.

Add tests for installation-token authentication, narrowly allowed automation identities, actual per-job permission needs, explicit exact-head validation with GITHUB_TOKEN, and absence of secrets.MINT_RELEASE_TOKEN dependency. Preserve all existing lifecycle/hotfix/failure/rollback/race tests. Run native/adapter/workflow validation. Report any remaining token/settings limitation from evidence, rather than introducing a PAT proactively. Report source implementation, checks, merge, activation, deployment, and production publication separately.

Deliver a concise final explanation of reduced moving parts and the remaining necessary activation steps. No new credential should appear on that list unless a specific tested limitation remains and is separately raised to the user.


Refactor checkpoint: installation-token authentication, narrow automation identity,
independent human approval rejection for bots, explicit head validation, and hotfix
source-build dispatch are tested locally. UI production reuses main.yaml; backend
continues to reuse deploy.yaml. Source remains UNRELEASED; no repository Actions
setting, credential, cloud or production mutation was performed.
