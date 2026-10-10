---
kit_metadata_version: 1
artifact: "spec"
workflow_version: 3
phase: "implement"
feature:
  id: "0012"
  slug: "documentation-site"
  dir: "0012-documentation-site"
references:
  - path: "../../references/environment-lifecycle.md"
    status: active
    read_policy: conditional
skills: []
---
# Documentation site

## Purpose

Publish readable public Mint documentation through GitHub Pages, grounded in the current implementation.

## Requirements

Use a mint-inspired responsive design and self-hosted JetBrains Mono. Explain source/artifact/environment evidence, integration and activation, arbitrary named non-production and sandbox environments, CHANGELOG, strict SemVer, Release PR review, rollback, hotfix, and reconciliation. Generate site-root llms.txt, full context, and raw Markdown guides. Validate PR builds and deploy only default-branch artifacts.

## Decisions

Use Python Markdown with a small explicit static builder, shared HTML template and CSS. No browser JavaScript is required. Upload only curated website sources and assets; repository governance/specs remain private to the repository. Default project URL is https://jamesonstone.github.io/mint/; llms.txt resides under that project site root. No domain-root publishing or custom domain is assumed.

## Readiness

PASS: the user scope is clear, current implementation and references establish behavior, and the clean new issue lane is approved. Policy examples are explicitly onboarding examples; no baseline/configuration is invented. No CLI behavior changes, runtime activation, cloud provisioning or consumer migration is required.

## Validation

Site regression checks, local link/anchor checks, root/project base URLs, browser desktop/mobile review, workflow actionlint, Go tests/vet/build, diff review and secret scan. Final observed results are tracked in TASKS.md.

## Outcome

Implementation is scoped to issue #23 and GH-23. Repository owner must merge after review and enable Pages with GitHub Actions as source before live publishing. Repository settings and merge are prohibited by local delivery rules.

## Additional information

No additional requirements or cross-repository dependencies.
