---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "implement"
feature:
  id: "0011"
  slug: "team-release-authorization"
  dir: "0011-team-release-authorization"
---
# SPEC

## PURPOSE

Allow team-owned projects to use Mint without naming one personal operator in release policy. Separate authority from assignment.

## CONTEXT

Mint v0.4.0 requires human_login and uses it for trusted controller authorization and issue/PR assignment. Aquarium, Terrarium and Nautilus are team-owned; their existing upgrade PRs must remove that personal policy after a feature-bearing release.

## REQUIREMENTS

Introduce explicit authorization: repository-write. Resolve effective GitHub repository permission server-side, accepting human write/maintain/admin and rejecting read/triage/bots/API errors. Preserve legacy named-operator behavior. Optional assignees do not grant authority. Retain exact review/head/intent guards, immutable artifacts and activation boundaries.

## ACCEPTED PLAN

Implement strict policy fields and client authorization, propagate policy through every production/environment entry point, separate optional assignment, add deterministic API-backed security/compatibility tests, document design, publish through a ready PR after validation and review, then update existing team consumer PRs.

## DECISIONS

Use effective repository permissions, including team-derived grants, instead of duplicating organization membership lists or choosing an arbitrary team slug. Explicit repository-write and legacy human_login are mutually exclusive. Keep the legacy path compatible; never silently broaden an existing named-user policy. Automatic generated controls retain only the narrow built-in Actions author allowance.

## DISCOVERIES

GitHub collaborator permission evidence is available to repository installation tokens with Metadata read. Controller identity must still be authenticated from its active default-branch workflow run before actor authorization.

## VALIDATION

PASS: full `go test ./...`, promotion/CLI race tests, `go vet ./...`, affected-package golangci-lint (zero issues), `make build`, generated team controller actionlint, gofmt, whitespace and changed-source <=300 checks. Security regressions cover effective permission, bots, identity mismatch, API failures, trusted run provenance, independent writer review with nonqualifying read approval, unassigned issue requests, strict fields, legacy digest compatibility and policy YAML round-trip. Independent adversarial review found digest incompatibility and review poisoning; both were repaired. Final review reports no blockers.

## OUTCOME

Implemented explicit team authorization and optional assignment; legacy restrictions remain. Published-release delivery and three existing consumer PR updates remain. No activation, team membership, repository permission or cloud mutation occurred.

## REPOSITORY MEMORY

This spec and the environment lifecycle reference preserve authorization rationale; agent instructions and README summarize the new policy.
