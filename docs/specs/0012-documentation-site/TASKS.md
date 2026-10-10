---
kit_metadata_version: 1
artifact: "tasks"
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
# Tasks

- [x] Inspect current implementation and allocate human-assigned issue #23 / GH-23.
- [x] Author eight public guides and shared mint-inspired responsive design.
- [x] Bundle JetBrains Mono and its OFL license.
- [x] Add explicit static build, Markdown/LLM outputs and project-path validation.
- [x] Add Pages workflow and local development instructions.
- [x] Complete browser, site, workflow, Go and secret-scan validation.

PASS: `make docs-check` (3 tests and eight HTML/Markdown guides), project/domain-root checks, policy example parsed by Mint, `actionlint`, full `go test ./...`, `go vet ./...`, `make build`, diff check and Gitleaks. Local `/mint/llms.txt` and `/mint/llms-full.txt` return HTTP 200 text/plain. Playwright desktop 1440px and mobile 390px review confirms locally loaded font and no page overflow; code/table regions scroll independently. GitHub Pages is not enabled (API HTTP 404); live publication is UNOBSERVED.

- [x] Obtain explicit staging/commit/push/ready-PR approval.
- [x] Deliver ready PR #24: https://github.com/jamesonstone/mint/pull/24 (GH-23, assigned to jamesonstone).
- [ ] Repository owner merges and enables Pages; verify live publication afterward.
