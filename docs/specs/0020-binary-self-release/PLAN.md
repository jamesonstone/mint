---
kit_metadata_version: 1
artifact: plan
feature:
  id: "0020"
  slug: binary-self-release
  dir: 0020-binary-self-release
summary: Publish the same Mint CLI binary through dev, stg, and prd.
---
# PLAN

## SUMMARY

Use Mint v2 and GitHub Actions to promote the Mint binary automatically.

## DEPENDENCIES

Use the published v2.0.0 bootstrap, Go toolchain from `go.mod`, and GitHub Actions artifacts.

## APPROACH

Validate repository configuration, build once in dev, verify in stg, and publish the verified asset from prd.

## COMPONENTS

Add a config reader to the CLI and shared verification and publication scripts to the release workflow.

## DATA

Pass the artifact ID, release tag, and SHA-256 digest through the job graph. Keep notes with the retained binary.

## INTERFACES

`mint config` prints the validated repository configuration as JSON. GitHub Releases receives the `mint-linux-amd64` binary.

## TESTING

Check config validation, workflow dependency and permission boundaries, digest mismatch rejection, and draft recovery. Run the dev and stg jobs on pull requests.

## RISKS

An interrupted upload leaves a draft. A retry verifies any existing asset and rejects conflicts before publishing.
