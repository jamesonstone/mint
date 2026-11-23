# Environment promotion

Use the same retained artifact in each environment. The repository defines its environments and controls which jobs and operators can deploy to them.

## Continuous staging

A workflow for a trusted default-branch push can run the following jobs in order:

1. Run the project's checks.
2. Resolve source version and build the artifact once. Record the repository, project, SemVer, full source SHA, digest, artifact locator, and producer workflow run and attempt.
3. Retain the build record and the immutable artifact. Confirm they will be available when an operator requests production.
4. Call the shared promotion job for staging. Run runtime verification and any required staging tests. Retain a receipt that binds the verified digest to the build record and to the staging workflow run.

## Manual production

Use `workflow_dispatch` to accept a project and an exact SemVer from an authorized operator. Restrict dispatch execution and write credentials to the trusted workflow and default branch. Apply repository-required approvals before the deployment.

1. Resolve the selected version to the retained build record. Validate the artifact and its provenance.
2. Require a successful, trusted staging run for the same repository, project, version, source SHA, and digest. Stop if evidence is missing, expired, failed, or mismatched.
3. Call the same promotion job as staging. Pass the exact artifact reference and the production environment. The job deploys the retained bytes without rebuilding.
4. Verify the running artifact and required configuration. Record success and optionally publish a GitHub Release after verification.

Give build and production deployment jobs separate permissions. Read the actual digest from the runtime and use [`mint verify`](integration.html#add-environment-promotion) before the job records success. When verification fails or is interrupted, stop publication and inspect the runtime before retrying.

## References and retention

For a container, use its registry digest. For a static site, retain a bundle or asset tree under a digest-addressed path and update the environment's reference to it. Environment labels such as `stg` and `prd` are optional aliases. Compare the immutable digest and record a separate identity for any runtime configuration.

GitHub Actions artifacts can retain a build until the repository's retention limit. For a longer promotion window, retain the artifact and its provenance in an immutable registry or an object store that supports your retention requirements. A missing artifact or receipt stops promotion.
