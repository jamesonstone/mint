---
kind: ruleset
slug: mint-integration
description: Keeps the public integration guide aligned with Mint's shipped behavior.
status: active
registry_scope: downstream
applies_to:
  - mint
  - integration
  - documentation
  - github-action
  - cli
  - deployment
read_policy_default: must
---

# Ruleset: mint-integration

## Purpose

Keep [Integrate Mint](https://jamesonstone.io/mint/integration.html) accurate for
people who install, configure, or operate Mint.

## Applies When

Read this rule for every repository change. Changes to Mint behavior or to how
users install, configure, or run Mint must update
`website/content/integration.md` in the same pull request.

## Rules

- Verify claims and examples against the shipped CLI, GitHub Action, policy
  parser, workflow generator, and tests that govern the changed behavior. Give
  the required Mint version or pin when support differs across versions.
- Show the supported path from setup through verification. Explain what Mint
  owns, what the application workflows own, and which actions require
  configuration, credentials, or human approval.
- Keep source tagging, artifact creation, deployment, runtime verification,
  changelog updates, and GitHub Release publication distinct. Ensure diagrams
  show the same steps and conditions as the examples.
- Describe only behavior available to the reader. When the pull request adds
  support that requires a new Mint release, identify that requirement and
  preserve instructions that still apply to supported older releases.
- If the change has no effect on integration, review the guide for accuracy
  without introducing unrelated edits.

## Verification

- Run `make docs-check`. Review the generated integration page, its Markdown
  source, and the relevant entries in `llms.txt` and `llms-full.txt`.
- Review changed diagrams at desktop and mobile widths. Confirm the order and
  meaning remain clear without color, JavaScript, or decorative arrows.
- After an authorized merge, verify the public integration URL and agent guides
  after the documentation deployment succeeds.
