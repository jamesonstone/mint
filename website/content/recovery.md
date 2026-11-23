# Recover a deployment

Recovery belongs to the repository workflow. Keep the retained artifact, its source version, the runtime configuration reference, and the successful verification receipt for each release you can restore.

## Roll forward

Revert or fix the source. Let the normal build and staging checks produce and verify a new version. Promote that version with manual production.

## Restore a retained artifact

Use an authorized recovery dispatch to select the previous verified artifact for the project. Resolve its retained digest, restore that reference, and verify the resulting runtime. Check whether the recovery also requires a separate configuration or data change.

## Recover from interrupted verification

Inspect the running artifact and configuration before another deployment. Compare the result with the intended deployment and the last verified release. Complete verification or explicitly restore the prior artifact before recording success or publishing a release.

## Retry publication

If deployment verification succeeded but the GitHub Release call failed, retry publication for the same tag and source SHA. Reuse the success receipt from the verified run. An existing GitHub Release documents publication, and the verification receipt documents the deployed artifact.
