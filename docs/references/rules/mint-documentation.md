---
kind: ruleset
slug: mint-documentation
description: Requires the public Mint documentation site and LLM guides to change with functional repository changes.
status: active
applies_to:
  - mint
  - behavior
  - cli
  - github-action
  - release
  - environment-policy
  - integration
  - documentation
read_policy_default: must
---

# Ruleset: mint-documentation

## Purpose

Keep the public documentation accurate for the implementation that ships. A functional change is incomplete when its documentation still describes the old behavior.

## Applies When

Read this rule before implementing or reviewing any functional repository change. This includes CLI commands, flags, defaults, outputs, errors, GitHub Action inputs/outputs, adapters, policy schemas, authorization, versioning, CHANGELOG generation, Release PRs, deployment workflows, environment support, or recovery behavior.

Formatting-only changes and changes to repository governance do not require a public guide edit when they leave product behavior unchanged.

## Rules

- Update the affected public guides under `website/content/` in the same pull request as each functional change. Do not defer documentation to another issue or release.
- Explain the new behavior, integration steps, prerequisites, compatibility, and relevant failure/recovery paths. Remove obsolete instructions and examples. Use source code and tests as evidence; distinguish supported behavior from planned behavior.
- Keep sandbox and non-production examples current. Explain arbitrary named environments without requiring production. Keep desired policy, source identity, artifact eligibility, verified deployment, publication, and live observation distinct.
- Use `website/content/` as the human/agent guide source. The builder generates HTML, raw Markdown, `llms.txt`, and `llms-full.txt`; never edit or commit `_site/` output. When adding/removing a guide, update `PAGES` in `scripts/build_docs.py` so navigation and both LLM entrypoints stay current.
- Keep command and YAML examples consistent with the changed interface. Identify placeholders and activation prerequisites. Never invent runtime evidence, credentials, or configuration identities to make an example appear ready.
- Preserve the mint-inspired design, self-hosted JetBrains Mono, readable line lengths, keyboard navigation, and responsive code/table scrolling when changing the site.
- Run `make docs-check` with `website/requirements.txt` installed. Do a rendered desktop/mobile review for layout changes. Run `actionlint` when changing the publishing workflow.
- Include the changed guides and observed documentation validation in the pull request. A functional diff without an affected guide update requires correction before completion.
- The `Documentation` workflow validates PRs and deploys merged `main` content to GitHub Pages. Do not report a site update as live until the deployment succeeds and the affected page and site-root `llms.txt` are verified at the public URL.

## Maintenance

This is a project-owned rule created with `kit rules add mint-documentation --must`. Keep its pointer outside the Kit-managed contract in all three agent entrypoints. Kit reconciliation must not replace the rule or its project-owned pointers.

## Verification

1. Map each functional behavior change to an updated public guide in the same PR.
2. Confirm examples and both generated LLM files reflect the implementation.
3. Run documentation checks and relevant product validation; report pending, skipped, or failed results accurately.
4. After an authorized merge, verify publication separately from source/CI success.
