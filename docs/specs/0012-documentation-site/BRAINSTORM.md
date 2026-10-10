---
kit_metadata_version: 1
artifact: "brainstorm"
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
# Documentation design

A compact static site keeps maintenance close to the Go repository and works for readers and agents without runtime JavaScript. Pale mint surfaces and dark green text provide a calm visual hierarchy. JetBrains Mono is self-hosted, with limited line length, generous spacing, responsive navigation, keyboard focus and scrollable code/tables.

An existing hosted-docs platform or JS application would add infrastructure and dependencies unnecessary for this scope. Raw Markdown and generated llms.txt avoid a second documentation source. There are no remaining design questions; current implementation determines technical claims.
