---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "verify"
feature:
  id: "0021"
  slug: "guided-init"
  dir: "0021-guided-init"
---
# SPEC

## PURPOSE

Provide a guided `mint init` for repositories that want to choose a deployment trigger for each environment and start from the same build-once promotion structure.

## CONTEXT

Mint v2 leaves deployment and authorization to repository workflows. The v2 fixture proof used one promotion path, an exact retained digest, and trusted runtime evidence. Mint v2.1.0 reads a pipeline-wide mode and an ordered list of names. Existing configurations and their JSON output are public inputs.

## REQUIREMENTS

- Guide the operator through the artifact type, ordered environment names, and each trigger. Default to dev and stg continuous and prd manual. Show the selected values and request confirmation before writes.
- Write `.mint.yaml` and reviewable GitHub Actions scaffolding. Preserve all existing files and keep incomplete input from generating any output files.
- Accept either the existing top-level deployment and environment name list or an ordered list of environment objects with a deployment mode each. Reject ambiguous mixtures.
- Build and retain after project checks on a trusted default-branch push. Run continuous environments in their listed relative order. A manual dispatch selects an allowed manual environment and an exact version. It resolves a retained build without running the build job and goes through the same promotion steps.
- For manual environments after the first entry, pass the preceding environment to the project's resolver. Require the resolver hook to check trusted successful verification evidence for the exact build there. Configure GitHub deployment protections on the target environment.
- Keep provider check, build, resolve, deploy, verification, and any publication in the repository. A scaffold without project hooks cannot report a verified deployment.

## ACCEPTED PLAN

1. Extend the configuration reader, preserving legacy validation and output. Implement the prompt, confirmation, and exclusive file writes.
2. Generate the push and dispatch workflow and its shared promotion workflow. Require repository hooks to return immutable artifact identity and runtime evidence. Keep permissions read-only by default.
3. Verify legacy config, guided flows, event routing, generated workflow syntax, and documentation. Open the work for review.

## DECISIONS

- Generate a single-project scaffold with one build workflow and a reusable promotion workflow. Component and provider-specific hooks can build on it separately.
- Write triggers in the generated Actions files so deployment is visible and reviewable. Record the same selections in the generated config.
- A default-branch push still retains a build when all environments are manual. Continuously deployed environments use the preceding continuous job on the same run. A manual environment uses the preceding configured environment as its evidence prerequisite, regardless of that environment's trigger.
- Use an immutable published Mint v2.1.0 action for version resolution and digest comparison. No GitHub Release publishing is configured by default.
- Treat the version resolver's output as a candidate. The repository build hook reserves and returns an exact version for the source, including when successive untagged pushes have the same candidate.

## DISCOVERIES

Build hooks retain artifacts with repository-selected storage. The scaffold verifies returned artifact fields, runs the same promotion job for each trigger, and uploads a verification receipt. The project's resolver must verify the retained build and the prerequisite receipt's trusted run and attempt.

## VALIDATION

- Native config and CLI tests include legacy JSON, custom environment choices, cancellation, collisions, and rollback on a partial write.
- Local generated-workflow tests cover mixed, all-continuous, and all-manual configurations. Actionlint accepts all three.
- A disposable shell fixture used the generated jobs and the actual Mint verifier. Staging and production recorded matching version, source SHA and artifact digest from one build. Missing staging evidence blocked resolution and a digest mismatch prevented a verification receipt. No provider or hosted deployment was run.
- A second fixture retained two pushes at two distinct project-reserved versions despite an identical candidate. Staging and manual production promoted the second retained artifact without a third build.
- `go test ./...`, `go vet ./...`, `make build`, Go formatting, `make docs-check`, and `kit check guided-init` passed. A compiled CLI smoke test created the expected files, read back the selected modes, and refused a second run.

## OUTCOME

The new `mint init` writes reviewed configuration and a repository-owned workflow scaffold. It does not overwrite existing setup. The repository supplies the script hooks, environment protections, version allocation, and release publication.

## REPOSITORY MEMORY

The integration and environment guides describe guided setup and the workflow hook contract. The Constitution lists the new CLI surface.
