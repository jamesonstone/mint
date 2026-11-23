---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "verify"
feature:
  id: "0020"
  slug: "binary-self-release"
  dir: "0020-binary-self-release"
---
# SPEC

## PURPOSE

Mint releases its own versioned CLI binary through continuous delivery. Dev builds the artifact once, staging verifies and executes it, and production publishes that exact binary to GitHub Releases.

## CONTEXT

Mint v2 provides source versioning, immutable tags, and digest comparison. The existing self-release workflow produces release state with no binary asset. The repository owns artifact retention and promotion through GitHub Actions.

## REQUIREMENTS

- Set `artifact: binary`, continuous delivery, and the ordered dev, stg, and prd path in `.mint.yaml`.
- Add a CLI reader that validates these configuration fields and provides machine-readable output. Use it in the workflow to check the actual configured artifact and path.
- Use the same versioned Linux amd64 CLI binary in each stage. Verify its SHA-256 against the previous stage and check the executable version.
- Start from a trusted default-branch push. After the build and staging checks pass, create or validate the immutable source tag with the promoted CLI. Publish the release only after GitHub reports the expected binary digest.
- Run the read-only development and staging path in pull requests. Use only the production job to change GitHub Release state.

## ACCEPTED PLAN

1. Define and test the v2 configuration reader.
2. Replace the self-release job with sequential dev, stg, and prd jobs. Verify the retained binary in stg and prd.
3. Create a verified draft release, attach the binary, then publish the draft. Resume on retries without replacing an existing asset.
4. Update the public integration guide and run local and hosted validation.

## DECISIONS

- Publish a raw Linux amd64 CLI asset first. Use the same hosted runner platform to run staging and production checks.
- Use the published Mint v2.0.0 CLI to resolve the version, and the candidate CLI to check its own version, compare verified digests, and create the immutable tag. The workflow owns GitHub draft and asset operations.
- Use read permissions for development and staging. Limit release write permissions to the trusted default-branch production job.

## DISCOVERIES

- GitHub Actions artifact download can strip executable permissions. The verifier checks the bytes independently before restoring permission and running Mint.
- GitHub Releases reports a SHA-256 digest for the uploaded asset. Verify it while the release is a draft, then confirm the published release still contains the same asset.

## VALIDATION

- Local Go tests, vet, versioned build, config output, workflow lint and documentation checks passed.
- Exercised the production script with the built CLI, a local fixture Git remote, and a GitHub API stand-in. The initial publish, already-published retry and interrupted draft resume succeeded. A conflicting asset or wrong staging digest stopped publication.

## OUTCOME

The workflow uses the new config reader and retains one versioned binary per run. Pull requests run dev and staging with read access. The trusted main push runs production with release write access.

## REPOSITORY MEMORY

Update the public integration guide, release workflow reference, and active command documentation.
