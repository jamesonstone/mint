---
name: mint-release-tooling
description: >-
  Use Mint to resolve Semantic Versions, write changelogs, create immutable
  Git tags, publish GitHub Releases, or generate a GHCR or ECR publishing workflow.
---

# Mint release tooling

Read [the command reference](instructions.md). If you change Mint itself, read
`AGENTS.md`, `docs/CONSTITUTION.md`, and the relevant feature specification.
Use [the public integration guide](https://jamesonstone.github.io/mint/integration.html)
when adding Mint to a repository.

## Release state workflow

1. Identify the repository, exact source commit, and optional component scope.
2. Fetch full Git history and tags. Resolve the next version with
   `mint release resolve`, or use `release select-tag` when selecting a tag
   that already points to the commit.
3. Generate a changelog if required. Use `--current-ref` if the version's
   tag has yet to be created.
4. Create or reuse the immutable source tag when authorized. Publish through
   `release github` when the workflow allows release publication. For a direct
   source release, `release publish` can combine versioning and publication.
5. Report the exact source tag and GitHub Release URL if published.

## Deployment workflow

Use your repository's shared promotion job for staging and production. Require
trusted staging verification for the exact project, SemVer, source, and
digest before manual production. Deploy the retained digest from the original
build. Record verified success only after the runtime check passes. Use an
environment alias for discovery and the immutable digest for verification.

When operating a legacy controller, identify its v1 pin and use the
[versioned lifecycle guide](https://github.com/jamesonstone/mint/blob/v1.1.2/docs/references/environment-lifecycle.md).
