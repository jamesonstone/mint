---
kit_metadata_version: 1
artifact: plan
feature:
  id: 0001
  slug: init-project
  dir: 0001-init-project
skills:
  - name: rlm
    source: repo-local guide
    path: docs/agents/RLM.md
    trigger: "broad/noisy repository discovery; analyze codebase; scan all files; large repository analysis; scan repository; recursive language model"
    required: true
relationships: []
references:
  - id: user-pasted-plan-task
    name: Pasted plan completion request
    type: prompt
    target: /Users/jamesonstone/.codex/attachments/7a4ccc07-de75-4014-9f36-4cf271cfca9a/pasted-text.txt
    relation: guides
    read_policy: must
    used_for: plan scope, required sections, reference requirements, confidence gate, and final response contract
    status: active
  - id: upstream-spec
    name: Fixed specification
    type: feature_doc
    target: docs/specs/0001-init-project/SPEC.md
    relation: constrains
    read_policy: must
    used_for: binding requirements, acceptance criteria, non-goals, edge cases, and RLM planning metadata requirement
    status: active
  - id: upstream-brainstorm
    name: Upstream brainstorm
    type: feature_doc
    target: docs/specs/0001-init-project/BRAINSTORM.md
    relation: informs
    read_policy: conditional
    used_for: research context, current repository facts, intended direction, and tradeoff rationale
    status: active
  - id: constitution
    name: Project constitution
    type: repo_doc
    target: docs/CONSTITUTION.md
    relation: implements
    read_policy: must
    used_for: primary implementation artifact, placeholder replacement target, and Kit-managed baseline preservation
    status: active
  - id: project-progress-summary
    name: Project progress summary
    type: repo_doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    relation: implements
    read_policy: must
    used_for: supporting progress artifact that must reflect the highest completed feature artifact
    status: active
  - id: kit-map-init-project
    name: Kit map for init-project
    type: command
    target: kit map 0001-init-project
    relation: verifies
    read_policy: evidence
    used_for: prior-work shortlist, feature phase, relationship absence, and front matter reference resolution
    status: active
  - id: repo-file-inventory
    name: Repository file inventory
    type: command
    target: find . -maxdepth 4 -type f -not -path './.git/*' -print
    relation: informs
    read_policy: evidence
    used_for: confirmed current repository contains docs/config but no source tree or dependency manifest
    status: active
  - id: readme
    name: README
    type: repo_doc
    target: README.md
    relation: informs
    read_policy: must
    used_for: product intent wording for Mint as version, changelog, and release tooling
    status: active
  - id: agent-routing-readme
    name: Agent routing entrypoint
    type: repo_doc
    target: docs/agents/README.md
    relation: guides
    read_policy: must
    used_for: scoped context routing and source-of-truth placement
    status: active
  - id: workflows
    name: Agent workflows
    type: repo_doc
    target: docs/agents/WORKFLOWS.md
    relation: constrains
    read_policy: must
    used_for: spec-driven source-of-truth order, readiness gate, and documentation-before-behavior rule
    status: active
  - id: guardrails
    name: Agent guardrails
    type: repo_doc
    target: docs/agents/GUARDRAILS.md
    relation: constrains
    read_policy: must
    used_for: completion bar, placeholder removal, and documentation-only safety boundary
    status: active
  - id: rlm
    name: RLM context routing
    type: skill
    target: docs/agents/RLM.md
    relation: guides
    read_policy: must
    used_for: selected planning skill and `parallelization_mode: "rlm"` metadata
    status: active
  - id: tooling
    name: Agent tooling guidance
    type: repo_doc
    target: docs/agents/TOOLING.md
    relation: guides
    read_policy: must
    used_for: canonical skill lookup, project-directory workflow, and secondary context ordering
    status: active
  - id: references-readme
    name: References index
    type: repo_doc
    target: docs/references/README.md
    relation: guides
    read_policy: conditional
    used_for: durable reference placement and pointer-loaded ruleset model
    status: active
  - id: testing-reference
    name: Testing reference
    type: repo_doc
    target: docs/references/testing.md
    relation: informs
    read_policy: conditional
    used_for: confirmed no stable project-specific test strategy exists yet
    status: active
  - id: tooling-reference
    name: Tooling reference
    type: repo_doc
    target: docs/references/tooling.md
    relation: informs
    read_policy: conditional
    used_for: confirmed no stable project-specific tooling reference exists yet
    status: active
  - id: external-systems-reference
    name: External systems reference
    type: repo_doc
    target: docs/references/external-systems.md
    relation: informs
    read_policy: conditional
    used_for: confirmed no durable external systems are documented yet
    status: active
  - id: kit-config
    name: Kit project config
    type: config
    target: .kit.yaml
    relation: constrains
    read_policy: must
    used_for: constitution path, feature doc paths, confidence target, no out-of-order progression, and managed rulesets
    status: active
  - id: gitignore
    name: Git ignore rules
    type: config
    target: .gitignore
    relation: informs
    read_policy: evidence
    used_for: Go-oriented signal and generated/local artifact boundaries
    status: active
  - id: envrc
    name: direnv loader
    type: config
    target: .envrc
    relation: informs
    read_policy: evidence
    used_for: local dotenv boundary and private environment handling
    status: active
  - id: coderabbit
    name: CodeRabbit config
    type: config
    target: .coderabbit.yaml
    relation: informs
    read_policy: evidence
    used_for: risk that automated review excludes documentation and agent routing files
    status: active
  - id: safety-guardrails-ruleset
    name: Safety guardrails ruleset
    type: ruleset
    target: docs/references/rules/safety-guardrails.md
    relation: constrains
    read_policy: conditional
    used_for: git/GitHub mutation boundary, secret safety, and failure handling constraints
    status: active
  - id: work-lane-gating-ruleset
    name: Work lane gating ruleset
    type: ruleset
    target: docs/references/rules/work-lane-gating.md
    relation: constrains
    read_policy: conditional
    used_for: confirms this documentation plan remains outside implementation lane gating
    status: active
  - id: github-pr-delivery-ruleset
    name: GitHub PR delivery ruleset
    type: ruleset
    target: docs/references/rules/github-pr-delivery.md
    relation: constrains
    read_policy: conditional
    used_for: delivery detail to reference rather than duplicate in the constitution
    status: active
  - id: kit-capabilities-usage-ruleset
    name: Kit capabilities usage ruleset
    type: ruleset
    target: docs/references/rules/kit-capabilities-usage.md
    relation: guides
    read_policy: conditional
    used_for: future Kit command-discovery uncertainty
    status: active
---
# PLAN

## SUMMARY

Update `docs/CONSTITUTION.md` as a compact project contract, preserving the Kit-managed baseline while replacing placeholders with durable principles, constraints, non-goals, and definitions. Keep the change documentation-only, use current repository evidence to separate facts from future intent, and validate with direct document review plus `kit map`, whitespace checks, placeholder scans, and a diff-scope check.

## APPROACH

1. Treat `SPEC.md` as binding and `BRAINSTORM.md` as supporting context only.
2. Use discovery-first routing for the implementation plan and downstream execution metadata: `parallelization_mode: "rlm"`.
3. Edit `docs/CONSTITUTION.md` as the primary artifact by replacing placeholder prose, not by expanding top-level agent routing files.
4. Preserve the Kit-managed baseline block byte-for-byte unless Kit regenerates it.
5. Structure the constitution as a high-level contract:
   - principles for correctness, clarity, minimalism, evidence, and document-first traceability;
   - constraints for source-of-truth order, ordered Kit progression, progress summary maintenance, secret/local state safety, generated artifact handling, and fact-vs-intent wording;
   - existing change classification retained and made concrete enough to guide future work;
   - non-goals that block premature release-algorithm, CI, hosted-service, package-manager, and external-integration scope;
   - definitions for the terms future agents need before adding product code.
6. Point detailed procedures to `docs/agents/*` and `docs/references/rules/*` instead of copying those rules into the constitution.
7. Update `docs/PROJECT_PROGRESS_SUMMARY.md` only as supporting documentation to keep the feature phase, summary, approach, and open items aligned with the completed plan.
8. Stop before product implementation; the next phase is task generation.

## COMPONENTS

1. `docs/CONSTITUTION.md`
   - Primary implementation artifact.
   - Owns durable project principles, constraints, workflow classification, non-goals, and definitions.
   - Must preserve the Kit-managed baseline block and avoid implementation claims not supported by repository evidence.
2. `docs/PROJECT_PROGRESS_SUMMARY.md`
   - Supporting artifact.
   - Reflects that `0001-init-project` has reached plan phase and has no open planning questions.
3. Existing pointer-loaded workflow docs
   - `docs/agents/*` remains the detailed agent workflow surface.
   - `docs/references/rules/*` remains the detailed git, GitHub, lane-gating, and command-discovery rules surface.
   - These files are dependencies for wording and references, not edit targets unless a direct inconsistency is found.
4. Validation evidence
   - `kit map 0001-init-project` verifies feature map and front matter reference resolution.
   - `git diff --check` verifies touched documentation whitespace.
   - `rg` placeholder scans and direct diff review verify section completeness and scope.

## DATA

1. Markdown section model for `docs/CONSTITUTION.md`
   - `PRINCIPLES`: durable decision principles.
   - `CONSTRAINTS`: invariant rules, including the preserved Kit-managed baseline block.
   - `CHANGE CLASSIFICATION`: concrete formal and ad hoc workflow guidance.
   - `NON-GOALS`: explicit scope boundaries.
   - `DEFINITIONS`: project vocabulary.
2. Managed block markers
   - `<!-- BEGIN KIT-MANAGED BASELINE RULES -->`
   - `<!-- END KIT-MANAGED BASELINE RULES -->`
   - The implementation must treat this block as owned by Kit.
3. Progress summary fields
   - Feature table phase and summary.
   - Feature summary status, intent, approach, open items, pointers, and last-updated timestamp.
4. Current evidence inputs
   - `README.md` supplies product intent.
   - `.gitignore` supplies Go-oriented and local/generated artifact signals.
   - `.envrc` supplies local dotenv behavior.
   - Absence of source files and manifests is evidence that runtime architecture is not implemented.

## INTERFACES

1. Files to edit
   - `docs/CONSTITUTION.md`
   - `docs/PROJECT_PROGRESS_SUMMARY.md` if needed to keep progress accurate
2. Files to inspect but not normally edit
   - `docs/specs/0001-init-project/SPEC.md`
   - `docs/specs/0001-init-project/BRAINSTORM.md`
   - `docs/agents/*`
   - `docs/references/*`
   - `.kit.yaml`
   - `README.md`
   - `.gitignore`
   - `.envrc`
   - `.coderabbit.yaml`
3. Commands and evidence interfaces
   - `kit map 0001-init-project`
   - `git diff --check -- docs/CONSTITUTION.md docs/PROJECT_PROGRESS_SUMMARY.md`
   - `rg --files-without-match -e TODO -e TBD docs/CONSTITUTION.md`
   - `rg --files-without-match -e TODO -e TBD docs/PROJECT_PROGRESS_SUMMARY.md`
   - `git diff -- docs/CONSTITUTION.md docs/PROJECT_PROGRESS_SUMMARY.md`
   - `git status --short --branch`
4. Side effects
   - Documentation content changes only.
   - No source code, tests, runtime configuration, generated artifacts, git staging, commits, pushes, issues, branches, or PRs.

## DEPENDENCIES

References are tracked in front matter. The implementation strategy depends on the fixed `SPEC.md`, upstream research in `BRAINSTORM.md`, current repository evidence, and repo-local Kit workflow/rules documents; it does not depend on runtime libraries, APIs, design assets, datasets, or external systems.

## RISKS

1. Risk: The constitution may accidentally claim runtime architecture, Go implementation, release algorithms, or test strategy that does not exist.
   Mitigation: Use evidence-qualified wording and explicitly separate current facts from future intended direction.
2. Risk: The Kit-managed baseline block could be edited by hand.
   Mitigation: Keep the block intact and verify it remains present after editing.
3. Risk: The constitution could duplicate long procedural rules from `docs/agents/*` or `docs/references/rules/*`.
   Mitigation: State invariants in the constitution and point to the detailed rule files.
4. Risk: `docs/PROJECT_PROGRESS_SUMMARY.md` could remain stale after plan completion.
   Mitigation: Update the summary/status/open-items fields as supporting documentation and verify with direct inspection.
5. Risk: Documentation changes may be skipped by automated review because `.coderabbit.yaml` excludes `docs/**`.
   Mitigation: Rely on direct diff review, placeholder scans, and explicit validation commands.
6. Risk: A future implementation task may expand into product code or runtime configuration.
   Mitigation: The plan leaves product implementation out of scope and instructs future agents to reclassify any such expansion before editing.

## TESTING

1. Front matter and feature-map validation
   - Run `kit map 0001-init-project`.
   - Evidence for SPEC acceptance: confirms feature resolution and reference readability.
2. Whitespace validation
   - Run `git diff --check -- docs/CONSTITUTION.md docs/PROJECT_PROGRESS_SUMMARY.md`.
   - Evidence for SPEC acceptance: no whitespace errors in touched docs.
3. Placeholder validation
   - Run `rg --files-without-match -e TODO -e TBD docs/CONSTITUTION.md`.
   - Run `rg --files-without-match -e TODO -e TBD docs/PROJECT_PROGRESS_SUMMARY.md`.
   - Evidence for SPEC acceptance: no placeholder-only constitution sections remain.
4. Managed-block validation
   - Review the final `docs/CONSTITUTION.md` diff and confirm the Kit-managed baseline markers and rules remain present.
   - Evidence for SPEC acceptance: baseline block preserved.
5. Contract-content validation
   - Review `docs/CONSTITUTION.md` against `SPEC.md` for principles, constraints, product intent, non-goals, definitions, pointer-loaded procedure references, and fact-vs-intent separation.
   - Evidence for SPEC acceptance: constitution satisfies the locked requirements without adding new scope.
6. Scope validation
   - Review `git status --short --branch` and the final diff file list.
   - Evidence for SPEC acceptance: no product code, tests, runtime config, generated artifacts, secrets, or machine-local state were modified by implementation.
