---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "deliver"
feature:
  id: "0012"
  slug: "consumer-adapters"
  dir: "0012-consumer-adapters"
---
# SPEC

## PURPOSE

Remove duplicated release plumbing from consumer repositories without weakening artifact or runtime evidence.

## CONTEXT

LabCore #1138 and LabCore UI #824 contain identical event, manifest, and installation helpers. Both intentionally retain schema 1 while activation is held. Mint already owns lifecycle policy and the journal; project adapters own ECS/S3 operations.

## REQUIREMENTS

Ship shared adapter tooling with the pinned Mint Action. Preserve schema 1 protocol, exact intent selection, immutable artifacts, credential-free source builds and project runtime verification. No activation, merge, infrastructure change, or schema 2 migration in this change.

## ACCEPTED PLAN

Package the existing compatibility event and manifest adapters upstream, call the native Mint CLI for lifecycle decisions, authenticate active Actions runs using native policy/client helpers, remove application copies, retain provider-specific helpers and security fixtures, validate upstream and both integrations, deliver a Mint dependency PR plus updates to existing consumer PRs.

## DECISIONS

Use a small Actions-installed mint-adapter companion (Python 3 and gh, available on supported Ubuntu runners) rather than porting stable glue into a second Go lifecycle engine. This is explicitly schema 1 compatibility tooling; schema 2 keeps its native controller. Configuration inputs and exclusions remain explicit caller data. A deployment manifest requires affirmative provider verification evidence; the helper itself performs no runtime verification.

## DISCOVERIES

The consumers must configure control_workflow explicitly: Mint defaults to mint-recovery.yaml, while these integrations use mint-control.yaml. Source, promotion and control events each require their configured trusted workflow identity.

## VALIDATION

PASS: go test ./..., go vet ./..., make build, changed-code golangci-lint, 15 Python adapter regressions, workflow actionlint, and actual composite Action installation from an external working directory with CLI/PATH/manifest smoke verification. Full lint reports six pre-existing untouched findings. Provider deployments remain UNOBSERVED.

## OUTCOME

Implemented Actions-installed shared compatibility companion and native active-run authorization. Consumer adoption is pinned to the upstream feature commit and held for upstream publication/review. The architecture exception for compatibility routing/serialization is explicit in the Constitution; eligibility and journal state remain Go-owned.

## REPOSITORY MEMORY

This spec records the compatibility boundary. README and adapter documentation will document installation and provider responsibility.
