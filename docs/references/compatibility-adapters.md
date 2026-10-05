# Schema 1 Actions adapters

Mint's pinned composite Action installs both `mint` and `mint-adapter` in
`RUNNER_TEMP`, outside the selected application source tree. Use `command: none`
when a workflow needs only installation. The Action builds its own pinned source
with Go caching disabled. The companion needs Python 3 and the GitHub CLI, as
provided by GitHub's Ubuntu runners; `go install` alone does not install it.

The companion consolidates existing schema 1 glue. It calls native Mint
operations for candidate eligibility, proposal review, immutable intent creation,
CAS journal changes, completion and publication. It does not implement provider
operations. Schema 2 uses the native `deployment control` controller instead;
this helper does not migrate a policy or enable deployment.

## Event routing

Call `mint-adapter event source`, `promote`, or `reconcile` from the configured
producer, promotion, or control workflow respectively. Native `adapter-policy`
authenticates the active default-branch run against GitHub before routing it.
Set `control_workflow` explicitly when it differs from Mint's default. Human
requests use the policy's authorization, including repository team permissions.
Callbacks use server-owned workflow paths rather than displayed workflow names.

The schema 1 adapter preserves existing inputs: `HOTFIX_PR` for isolated source,
and `CONTROL_OPERATION` (`reconcile`, `hotfix`, `rollback`, `publish`),
`REQUEST_ISSUE`, `ROLLBACK_SOURCE`, `RELEASE_SUMMARY`, and `REQUEST_INTENT` for
recovery. It reads standard `GITHUB_*` event/run identities, `GH_TOKEN`,
`GITHUB_OUTPUT`, and `RUNNER_TEMP`. Hotfix issues contain the existing
`Verified production baseline ID` and `Reviewed fix SHAs` fields (`AUTHORED`
means a newly authored fix). Normal successful builds will update the existing
release proposal whenever Mint's lifecycle permits it; only one release PR is
used, with outcomes attached to that PR.

## Provider manifests

`mint-adapter manifest` supports `configuration`, `verify-configuration`,
`candidate OUTPUT`, `baseline CANDIDATE OUTPUT`, `deployment OUTPUT`, and
`intent-outputs INTENT`. Inputs retain the existing schema 1 protocol:

- `RELEASE_ENVIRONMENT`: the configured environment.
- `RELEASE_VARIABLES`: JSON object of standing configuration.
- `RELEASE_CONFIG_FILES`: JSON list of repository-relative regular files.
- `RELEASE_CONFIG_EXCLUDE_VARIABLES`: optional JSON list of provider-owned
  exclusions; `MINT_*` control variables are always excluded.
- `SOURCE_SHA`, `VERSION_TAG`, `ARTIFACT_REFERENCE`, `ARTIFACT_DIGEST`:
  immutable candidate identity; optional `CANDIDATE_KIND`, `BASELINE_ID`,
  `SOURCE_PR`, `SOURCE_BUILD_RUN_ID`, and `SOURCE_BUILD_RUN_URL` retain source proof.
- `INTENT_ID`, `EXPECTED_CONFIGURATION`: frozen promotion identity.

Configuration hashing keeps the existing canonical JSON algorithm, so unchanged
inputs retain their identity. File links and traversal are rejected. For baseline
construction, supply exact successful source-build evidence and run provider
verification first. For deployment construction, the provider must first verify
the exact artifact in the live runtime, then set `DEPLOYMENT_VERIFIED=true` in the
subsequent successful attestation step. That flag records the caller's evidence;
it is not a Mint runtime check. Mint independently authenticates and validates the
uploaded manifest when completing the intent. Failed or uncertain verification
must never run this step or unlock the deployment fence.

Install trusted tooling before checking out a selected source for configuration
reads. Keep source execution isolated from publication credentials and shared
runner caches. Providers retain their own artifact retention and runtime checks.
