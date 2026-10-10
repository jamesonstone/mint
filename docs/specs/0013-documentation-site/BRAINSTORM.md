---
kit_metadata_version: 1
artifact: "brainstorm"
workflow_version: 3
phase: "implement"
feature:
  id: "0013"
  slug: "documentation-site"
  dir: "0013-documentation-site"
skills: []
---
# Documentation design

A compact static site keeps maintenance close to the Go repository and works for readers and agents without runtime JavaScript. Pale mint surfaces and dark green text provide a calm visual hierarchy. JetBrains Mono is self-hosted, with limited line length, generous spacing, responsive navigation, keyboard focus and scrollable code/tables.

An existing hosted-docs platform or JS application would add infrastructure and dependencies unnecessary for this scope. Raw Markdown and generated llms.txt avoid a second documentation source. There are no remaining design questions; current implementation determines technical claims.

## SUMMARY

Create public implementation-backed guides with a small static builder.

## USER THESIS

Readers need clear integration, sandbox, versioning and recovery guidance in a readable mint-inspired site.

## RELATIONSHIPS

The environment lifecycle and team authorization features establish documented behavior. Consumer adapters retain a separate feature identity (0012).

## CODEBASE FINDINGS

Go release, changelog, policy and controller code defines behavior. The documentation generator publishes only curated website content.

## AFFECTED FILES

`website/`, `scripts/build_docs.py`, `scripts/test_build_docs.py`, Makefile, README and `.github/workflows/docs.yaml`.

## DEPENDENCIES

Python Markdown for rendering; licensed self-hosted JetBrains Mono; GitHub Pages for publication.

## QUESTIONS

No unresolved design questions.

## OPTIONS

A hosted framework adds unnecessary dependencies. A curated static Markdown site provides HTML and agent guides from one source.

## RECOMMENDED STRATEGY

Keep explicit navigation, responsive CSS and generated agent entrypoints. Keep repository governance outside published artifacts.

## NEXT STEP

Maintain the live documentation with functional changes under `mint-documentation.md`.
