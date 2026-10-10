---
kit_metadata_version: 1
artifact: "plan"
workflow_version: 3
phase: "implement"
feature:
  id: "0013"
  slug: "documentation-site"
  dir: "0013-documentation-site"
skills: []
---
# Plan

1. Inspect current lifecycle, version, changelog, policy and authorization behavior.
2. Author curated public guides and an accessible static template with self-hosted licensed fonts.
3. Generate HTML, raw Markdown, llms.txt and llms-full.txt from one source set; check local links and anchors.
4. Add pinned Pages workflow with PR validation and main-only deployment, plus local Make targets.
5. Review content, desktop/mobile layouts and publishing boundaries; validate and secret-scan.
6. Obtain explicit staging/commit approval required by GUARDRAILS.md, deliver a ready PR, and report live activation separately.

No delegation or cross-repository implementation is required.

## SUMMARY

Publish readable HTML and agent guides from a single curated Markdown source.

## APPROACH

Use a pinned Markdown renderer and an explicit page list; validate local links and anchors before uploading the site artifact.

## COMPONENTS

Curated guides, shared template, CSS, self-hosted fonts, Python generator, regression checks and Pages workflow.

## DATA

Only public documentation and licensed assets. No runtime state, customer data or credentials are published.

## INTERFACES

Make targets provide local build/check/serve commands. GitHub Pages serves HTML, raw Markdown and site-root LLM files.

## DEPENDENCIES

`website/requirements.txt` pins Markdown. GitHub Actions use immutable pins. Font assets retain OFL.txt.

## RISKS

Stale functional guidance, broken project-site paths or incomplete deployment evidence. Same-PR updates, link validation and separate live checks address these risks.

## TESTING

Run `make docs-check`, actionlint for workflow changes, and browser desktop/mobile checks for layout changes. Run product checks for functional changes.
