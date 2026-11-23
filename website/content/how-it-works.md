# How Mint works

Mint versions source and publishes release state. Your repository workflows build, retain, deploy, and verify artifacts.

## Release sequence

1. Fetch full Git history and tags. Resolve the next version with Mint and associate it with the exact source commit.
2. Build once after the configured checks pass. Retain the artifact and record the project, full source SHA, SemVer, digest, and producer run.
3. Deploy that digest to staging. Record the deployed identity after runtime verification and staging tests succeed.
4. Dispatch production for the exact project and SemVer. Resolve the retained artifact, check the matching staging evidence, and deploy the same digest through the shared promotion workflow.
5. After production verification succeeds, invoke Mint to publish the GitHub Release for the version you deployed. Your workflow decides whether and when to write or commit a changelog.

Repositories that only publish release state can use [the direct publishing command](integration.html#publish-release-state).

## Evidence and identity

| Evidence | What it establishes |
| --- | --- |
| Immutable Git tag | The version points to a source commit. |
| Build record | A trusted run produced the artifact with that digest. |
| Staging verification | The staging workflow verified that digest. |
| Production verification | The production workflow verified the deployed digest. |
| GitHub Release | The workflow published notes and release metadata for a source tag. |

Use an immutable digest for deployment. A registry tag or a static site's environment reference can point to that digest. Record environment-specific runtime configuration separately from artifact identity. For a bundle, verify all of the artifacts selected together.
