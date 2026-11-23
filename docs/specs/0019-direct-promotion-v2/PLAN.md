---
kit_metadata_version: 1
artifact: plan
feature:
  id: "0019"
  slug: direct-promotion-v2
  dir: 0019-direct-promotion-v2
summary: Prove direct promotion and remove the Mint v1 controller.
---
# PLAN

## SUMMARY

Prove build-once promotion and make Mint v2 a set of release primitives for repository-owned workflows.

## DEPENDENCIES

The two private testbeds need GitHub Actions and a repository writer for manual dispatch. Provider tests and Kit integration have separate milestones.

## APPROACH

Run staging from a default-branch push. Run manual production by project and version. Both call the same promotion workflow and verify the runtime fixture digest. Remove controller-only source, action interfaces, bundled adapters, and CI. Update the public guides and agent references.

## COMPONENTS

Keep `pkg/release`, `pkg/changelog`, the CLI, and the composite action. Remove `pkg/promotion` and the controller commands. Move supported Go imports to the v2 module path.

## DATA

A build manifest binds repository, project, SemVer, exact source SHA, retained artifact digest, and producer run and attempt. Runtime receipts add environment and verification identity.

## INTERFACES

Production uses GitHub workflow dispatch with project and SemVer. Source versioning accepts explicit tag prefix and paths in the CLI and action. Existing controller consumers use the immutable v1 release.

## TESTING

Compare hosted build archives and runtime receipts. Exercise refused staging evidence and failed and interrupted verification. Run Go tests, vet, build, actionlint, and the public documentation checks.

## RISKS

The v2 controller removal breaks v1 consumers that follow an unpinned action ref. The fixture proves hosted workflow behavior with test artifacts. Registry and CloudFront behavior require separate provider validation.
