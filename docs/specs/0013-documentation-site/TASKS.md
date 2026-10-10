---
kit_metadata_version: 1
artifact: "tasks"
workflow_version: 3
phase: "implement"
feature:
  id: "0013"
  slug: "documentation-site"
  dir: "0013-documentation-site"
skills: []
---
# Tasks

## TASK LIST

- [x] T001: Inspect current implementation and allocate human-assigned issue #23 / GH-23.
- [x] T002: Author eight public guides and shared mint-inspired responsive design.
- [x] T003: Bundle JetBrains Mono and its OFL license.
- [x] T004: Add explicit static build, Markdown/LLM outputs and project-path validation.
- [x] T005: Add Pages workflow and local development instructions.
- [x] T006: Complete browser, site, workflow, Go and secret-scan validation.

PASS: `make docs-check` (3 tests and eight HTML/Markdown guides), project/domain-root checks, policy example parsed by Mint, `actionlint`, full `go test ./...`, `go vet ./...`, `make build`, diff check and Gitleaks. Local `/mint/llms.txt` and `/mint/llms-full.txt` return HTTP 200 text/plain. Playwright desktop 1440px and mobile 390px review confirms locally loaded font and no page overflow; code/table regions scroll independently. Pages was subsequently enabled with workflow publishing and HTTPS enforcement; live publication was verified after successful deployment.

- [x] T007: Obtain explicit staging/commit/push/ready-PR approval.
- [x] T008: Deliver ready PR #24: https://github.com/jamesonstone/mint/pull/24 (GH-23, assigned to jamesonstone).
- [x] T009: Repository owner authorized merge and Pages activation; live publication verified.

## PROGRESS TABLE

| ID | TASK | STATUS | OWNER | DEPENDENCIES |
| -- | ---- | ------ | ----- | ------------ |
| T001 | Inspect current implementation and allocate human-assigned issue #23 / GH-23. | done | Jameson Stone |  |
| T002 | Author eight public guides and shared mint-inspired responsive design. | done | Jameson Stone | T001 |
| T003 | Bundle JetBrains Mono and its OFL license. | done | Jameson Stone | T002 |
| T004 | Add explicit static build, Markdown/LLM outputs and project-path validation. | done | Jameson Stone | T003 |
| T005 | Add Pages workflow and local development instructions. | done | Jameson Stone | T004 |
| T006 | Complete browser, site, workflow, Go and secret-scan validation. | done | Jameson Stone | T005 |
| T007 | Obtain explicit staging/commit/push/ready-PR approval. | done | Jameson Stone | T006 |
| T008 | Deliver ready PR #24: https://github.com/jamesonstone/mint/pull/24 (GH-23, assigned to jamesonstone). | done | Jameson Stone | T007 |
| T009 | Repository owner authorized merge and Pages activation; live publication verified. | done | Jameson Stone | T008 |

## TASK DETAILS

PR #24 merged as 438fba1. Pages run 38080602267 succeeded; all eight pages, Markdown, fonts and both LLM files returned HTTP 200. HTTPS is enforced.

### T001

- **GOAL**: Inspect current implementation and allocate human-assigned issue #23 / GH-23.
- **VERIFY**: Site checks and recorded delivery evidence above.
- **EXPECTED FILES**: Curated website, build/workflow sources or recorded approval/delivery evidence appropriate to this task.
- **RISK**: Documentation drift or an unsupported publication claim.
- **ROLLBACK**: Restore the preceding reviewed documentation through a new PR.

### T002

- **GOAL**: Author eight public guides and shared mint-inspired responsive design.
- **VERIFY**: Site checks and recorded delivery evidence above.
- **EXPECTED FILES**: Curated website, build/workflow sources or recorded approval/delivery evidence appropriate to this task.
- **RISK**: Documentation drift or an unsupported publication claim.
- **ROLLBACK**: Restore the preceding reviewed documentation through a new PR.

### T003

- **GOAL**: Bundle JetBrains Mono and its OFL license.
- **VERIFY**: Site checks and recorded delivery evidence above.
- **EXPECTED FILES**: Curated website, build/workflow sources or recorded approval/delivery evidence appropriate to this task.
- **RISK**: Documentation drift or an unsupported publication claim.
- **ROLLBACK**: Restore the preceding reviewed documentation through a new PR.

### T004

- **GOAL**: Add explicit static build, Markdown/LLM outputs and project-path validation.
- **VERIFY**: Site checks and recorded delivery evidence above.
- **EXPECTED FILES**: Curated website, build/workflow sources or recorded approval/delivery evidence appropriate to this task.
- **RISK**: Documentation drift or an unsupported publication claim.
- **ROLLBACK**: Restore the preceding reviewed documentation through a new PR.

### T005

- **GOAL**: Add Pages workflow and local development instructions.
- **VERIFY**: Site checks and recorded delivery evidence above.
- **EXPECTED FILES**: Curated website, build/workflow sources or recorded approval/delivery evidence appropriate to this task.
- **RISK**: Documentation drift or an unsupported publication claim.
- **ROLLBACK**: Restore the preceding reviewed documentation through a new PR.

### T006

- **GOAL**: Complete browser, site, workflow, Go and secret-scan validation.
- **VERIFY**: Site checks and recorded delivery evidence above.
- **EXPECTED FILES**: Curated website, build/workflow sources or recorded approval/delivery evidence appropriate to this task.
- **RISK**: Documentation drift or an unsupported publication claim.
- **ROLLBACK**: Restore the preceding reviewed documentation through a new PR.

### T007

- **GOAL**: Obtain explicit staging/commit/push/ready-PR approval.
- **VERIFY**: Site checks and recorded delivery evidence above.
- **EXPECTED FILES**: Curated website, build/workflow sources or recorded approval/delivery evidence appropriate to this task.
- **RISK**: Documentation drift or an unsupported publication claim.
- **ROLLBACK**: Restore the preceding reviewed documentation through a new PR.

### T008

- **GOAL**: Deliver ready PR #24: https://github.com/jamesonstone/mint/pull/24 (GH-23, assigned to jamesonstone).
- **VERIFY**: Site checks and recorded delivery evidence above.
- **EXPECTED FILES**: Curated website, build/workflow sources or recorded approval/delivery evidence appropriate to this task.
- **RISK**: Documentation drift or an unsupported publication claim.
- **ROLLBACK**: Restore the preceding reviewed documentation through a new PR.

### T009

- **GOAL**: Repository owner authorized merge and Pages activation; live publication verified.
- **VERIFY**: Site checks and recorded delivery evidence above.
- **EXPECTED FILES**: Curated website, build/workflow sources or recorded approval/delivery evidence appropriate to this task.
- **RISK**: Documentation drift or an unsupported publication claim.
- **ROLLBACK**: Restore the preceding reviewed documentation through a new PR.

## DEPENDENCIES

The repository owner authorized merge and Pages settings. No remaining activation dependency.

## NOTES

Legacy feature numbering was corrected from 0012 to 0013 during issue #25 reconciliation. Historical project-wide spec findings are separate from site publication.
