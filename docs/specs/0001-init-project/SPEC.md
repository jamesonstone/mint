---
kit_metadata_version: 1
artifact: spec
feature:
  id: 0001
  slug: init-project
  dir: 0001-init-project
skills:
  - name: rlm
    source: repo-local guide
    path: docs/agents/RLM.md
    trigger: "analyze codebase; scan all files; large repository analysis; scan repository; recursive language model; broad or noisy repository context"
    required: true
relationships: []
references:
  - id: user-pasted-spec-task
    name: Pasted spec completion request
    type: prompt
    target: /Users/jamesonstone/.codex/attachments/82791ff5-9b8a-453e-a0da-2939b4d1af9e/pasted-text.txt
    relation: guides
    read_policy: must
    used_for: spec completion scope, required sections, metadata requirements, confidence gate, and final response contract
    status: active
  - id: upstream-brainstorm
    name: Upstream brainstorm
    type: feature_doc
    target: docs/specs/0001-init-project/BRAINSTORM.md
    relation: informs
    read_policy: must
    used_for: validated codebase findings, recommended defaults, non-goals, dependencies, and next-step scope
    status: active
  - id: constitution
    name: Project constitution
    type: repo_doc
    target: docs/CONSTITUTION.md
    relation: constrains
    read_policy: must
    used_for: target document, current placeholder state, and Kit-managed baseline rules that must be preserved
    status: active
  - id: project-progress-summary
    name: Project progress summary
    type: repo_doc
    target: docs/PROJECT_PROGRESS_SUMMARY.md
    relation: constrains
    read_policy: must
    used_for: confirms feature phase and enforces highest-completed-artifact tracking
    status: active
  - id: kit-map-init-project
    name: Kit map for init-project
    type: command
    target: kit map 0001-init-project
    relation: verifies
    read_policy: evidence
    used_for: verified feature phase, no prior relationships, and resolvable front matter references
    status: active
  - id: repo-file-inventory
    name: Repository file inventory
    type: command
    target: find . -maxdepth 4 -type f -not -path './.git/*' -print
    relation: informs
    read_policy: evidence
    used_for: established that the repository currently contains docs and config but no source tree or dependency manifest
    status: active
  - id: readme
    name: README
    type: repo_doc
    target: README.md
    relation: informs
    read_policy: must
    used_for: only explicit product intent for Mint as a version, changelog, and release tool
    status: active
  - id: agents-routing
    name: AGENTS routing table
    type: repo_doc
    target: AGENTS.md
    relation: guides
    read_policy: conditional
    used_for: repo-local source-of-truth routing and instruction entrypoint expectations
    status: active
  - id: claude-routing
    name: CLAUDE routing table
    type: repo_doc
    target: CLAUDE.md
    relation: guides
    read_policy: conditional
    used_for: alternate agent entrypoint alignment with repo-local docs
    status: active
  - id: copilot-instructions
    name: GitHub Copilot repository instructions
    type: repo_doc
    target: .github/copilot-instructions.md
    relation: guides
    read_policy: conditional
    used_for: Copilot routing alignment and non-monolithic instruction rule
    status: active
  - id: agent-routing-readme
    name: Agent routing entrypoint
    type: repo_doc
    target: docs/agents/README.md
    relation: guides
    read_policy: must
    used_for: task classification, source-of-truth routing, and scoped document loading
    status: active
  - id: workflows
    name: Agent workflows
    type: repo_doc
    target: docs/agents/WORKFLOWS.md
    relation: constrains
    read_policy: must
    used_for: spec-driven workflow, authority order, and readiness gate expectations
    status: active
  - id: guardrails
    name: Agent guardrails
    type: repo_doc
    target: docs/agents/GUARDRAILS.md
    relation: constrains
    read_policy: must
    used_for: required section population, documentation-only boundary, and front matter maintenance
    status: active
  - id: rlm
    name: RLM context routing
    type: skill
    target: docs/agents/RLM.md
    relation: guides
    read_policy: must
    used_for: selected execution-time skill, just-in-time discovery, and required downstream parallelization mode
    status: active
  - id: tooling
    name: Agent tooling guidance
    type: repo_doc
    target: docs/agents/TOOLING.md
    relation: guides
    read_policy: must
    used_for: canonical skills discovery, project-directory workflow, and secondary global input ordering
    status: active
  - id: references-readme
    name: References index
    type: repo_doc
    target: docs/references/README.md
    relation: guides
    read_policy: conditional
    used_for: durable-reference placement and rule pointer-loading model
    status: active
  - id: testing-reference
    name: Testing reference
    type: repo_doc
    target: docs/references/testing.md
    relation: informs
    read_policy: conditional
    used_for: confirmed no stable project-specific testing strategy exists yet
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
    used_for: confidence threshold, feature doc paths, constitution path, no out-of-order progression, and managed rulesets
    status: active
  - id: gitignore
    name: Git ignore rules
    type: config
    target: .gitignore
    relation: informs
    read_policy: evidence
    used_for: Go-oriented implementation signal and local/generated artifact exclusions
    status: active
  - id: envrc
    name: direnv loader
    type: config
    target: .envrc
    relation: informs
    read_policy: evidence
    used_for: local dotenv loading and private environment boundary
    status: active
  - id: coderabbit
    name: CodeRabbit config
    type: config
    target: .coderabbit.yaml
    relation: informs
    read_policy: evidence
    used_for: evidence that automated review excludes docs and top-level agent routing files
    status: active
  - id: pr-template
    name: Pull request template
    type: repo_doc
    target: .github/pull_request_template.md
    relation: informs
    read_policy: conditional
    used_for: current PR heading structure for any explicit future delivery workflow
    status: active
  - id: safety-guardrails-ruleset
    name: Safety guardrails ruleset
    type: ruleset
    target: docs/references/rules/safety-guardrails.md
    relation: constrains
    read_policy: conditional
    used_for: git, GitHub, identity, secret, protected-branch, and failure-handling boundaries
    status: active
  - id: work-lane-gating-ruleset
    name: Work lane gating ruleset
    type: ruleset
    target: docs/references/rules/work-lane-gating.md
    relation: constrains
    read_policy: conditional
    used_for: confirms standalone Kit spec work is non-implementation and must not force issue, branch, commit, push, or PR workflow
    status: active
  - id: github-pr-delivery-ruleset
    name: GitHub PR delivery ruleset
    type: ruleset
    target: docs/references/rules/github-pr-delivery.md
    relation: constrains
    read_policy: conditional
    used_for: on-demand delivery workflow details that the constitution should reference instead of duplicating
    status: active
  - id: kit-capabilities-usage-ruleset
    name: Kit capabilities usage ruleset
    type: ruleset
    target: docs/references/rules/kit-capabilities-usage.md
    relation: guides
    read_policy: conditional
    used_for: command-discovery guidance for future Kit command uncertainty
    status: active
  - id: secondary-global-inputs
    name: Secondary global inputs
    type: global_context
    target: /Users/jamesonstone/.claude/CLAUDE.md, /Users/jamesonstone/.codex/AGENTS.md, /Users/jamesonstone/.codex/instructions.md, /Users/jamesonstone/.codex/skills/*/SKILL.md
    relation: informs
    read_policy: evidence
    used_for: skills discovery after repo-local docs; no additional execution-time skill selected
    status: optional
---
# SPEC

## SUMMARY

Defines the project constitution for Mint as a Kit-managed, documentation-first release tooling project before runtime code exists. The constitution must lock durable development rules, product intent, constraints, non-goals, and terms without inventing implementation details or duplicating pointer-loaded workflow rules.

## PROBLEM

Mint currently has a short `README.md` product signal and Kit scaffold documentation, but `docs/CONSTITUTION.md` is mostly placeholder text. Without a completed constitution, future agents and humans lack a canonical project contract for source-of-truth order, workflow progression, scope boundaries, implementation posture, and product direction.

## GOALS

1. Replace placeholder constitution content with durable, project-specific rules that future work can rely on.
2. Preserve the existing Kit-managed baseline rules in `docs/CONSTITUTION.md`.
3. Capture Mint's current product intent as a release/version/changelog tool while clearly marking unimplemented runtime architecture as future work.
4. Define non-negotiable constraints for documentation-first workflow, source-of-truth order, progress tracking, local/generated artifact handling, and secret safety.
5. Define project goals, non-goals, and key terms precisely enough for a later implementation agent to draft `PLAN.md` and `TASKS.md` without reopening broad discovery.
6. Keep detailed procedural rules pointer-loaded through `docs/agents/*` and `docs/references/rules/*` instead of copying them into the constitution.
7. Keep `docs/PROJECT_PROGRESS_SUMMARY.md` aligned with the highest completed artifact for `0001-init-project`.

## NON-GOALS

1. Do not implement product code, tests, runtime configuration, CI, release automation, or generated artifacts.
2. Do not introduce or declare a Go module, package manifest, dependency lockfile, command surface, or source tree.
3. Do not define concrete release algorithms such as SemVer bump detection, changelog grouping, tag creation, publishing behavior, or GitHub release behavior.
4. Do not claim that Mint has implemented architecture, runtime dependencies, external integrations, or test strategy before those facts exist in the repository.
5. Do not duplicate the detailed git, GitHub, work-lane, PR delivery, or Kit command-discovery rules already maintained under `docs/agents/*` and `docs/references/rules/*`.
6. Do not broaden the feature into updating `AGENTS.md`, `CLAUDE.md`, `.github/copilot-instructions.md`, starter reference docs, or product `README.md` unless a direct contradiction with the new constitution is found.
7. Do not create an issue, branch, commit, push, or PR as part of this feature unless the user explicitly requests delivery workflow.

## USERS

1. Human maintainers who need a concise project contract before adding Mint's first implementation.
2. Coding agents working through Kit feature phases under `docs/specs/0001-init-project`.
3. Reviewers who need to distinguish current repository facts from future release-tooling intent.
4. Future implementation agents that need stable boundaries for product code, tests, and release behavior.

## SKILLS

Skills are tracked in front matter. The selected execution-time skill is `rlm` via `docs/agents/RLM.md`; no repo-local `.agents/skills/*/SKILL.md` files exist, and inspected secondary global skills were not applicable to this constitution feature.

## RELATIONSHIPS

Relationships are tracked in front matter. This feature has no prior feature dependency because `0001-init-project` is the first and only feature in `docs/specs`.

## DEPENDENCIES

References are tracked in front matter. The feature depends on repo-local docs and evidence from the current repository state, not on runtime libraries or external services.

## REQUIREMENTS

1. [SPEC-01] The implementation must update `docs/CONSTITUTION.md` as the primary artifact and keep the change documentation-only.
2. [SPEC-02] `docs/CONSTITUTION.md` must contain populated sections for principles, constraints, change classification, non-goals, and definitions; no section may remain empty or contain only placeholder/TODO comments.
3. [SPEC-03] The existing `<!-- BEGIN KIT-MANAGED BASELINE RULES -->` through `<!-- END KIT-MANAGED BASELINE RULES -->` block must remain intact unless Kit itself regenerates it.
4. [SPEC-04] The constitution must state that `docs/CONSTITUTION.md` is the canonical project contract, while `docs/specs/<feature>/` controls feature-scoped requirements, plans, and tasks.
5. [SPEC-05] The constitution must preserve the source-of-truth order from repo-local workflow guidance: safety and permissions, current user request, constitution, feature artifacts, then repo conventions.
6. [SPEC-06] The constitution must state that spec-driven work proceeds through Kit artifacts in order and that out-of-order progression is not allowed unless the user explicitly overrides the project workflow.
7. [SPEC-07] The constitution must require `docs/PROJECT_PROGRESS_SUMMARY.md` to reflect the highest completed artifact per feature whenever feature docs advance.
8. [SPEC-08] The constitution must capture Mint's product intent as computing the next version, writing the changelog, and minting the release.
9. [SPEC-09] The constitution may describe Mint as an intended CLI/release automation project with a Go-oriented signal from `.gitignore`, but it must not claim Go implementation exists before `go.mod` or source files exist.
10. [SPEC-10] The constitution must explicitly distinguish current facts from future intended architecture, dependencies, command surface, and release behavior.
11. [SPEC-11] The constitution must require future product behavior, release algorithms, external integrations, and test strategy to be defined in feature specs before implementation.
12. [SPEC-12] The constitution must require generated local state, caches, `.env`, secrets, private keys, tokens, and machine-local config to remain out of version control.
13. [SPEC-13] The constitution must recognize `.kit/*` local runtime artifacts and direnv/dotenv inputs as local developer state, not durable project state.
14. [SPEC-14] The constitution must keep top-level agent entrypoints as routing tables and prohibit turning them into monolithic always-loaded manuals.
15. [SPEC-15] The constitution must direct detailed workflow, RLM, tooling, guardrail, and delivery procedures to `docs/agents/*` and `docs/references/rules/*` rather than duplicating those procedures.
16. [SPEC-16] The constitution must include non-goals that prevent premature scope expansion into package-manager-specific release behavior, hosted release services, broad CI orchestration, external-system integrations, and detailed release algorithms.
17. [SPEC-17] The constitution must define key terms needed for later work, including Mint, constitution, Kit-managed project, feature artifact, release, version, changelog, local generated state, and implementation evidence.
18. [SPEC-18] The constitution must preserve the current documentation file-size exception: the source file size guideline does not apply to `docs/**`, `.kit/**`, or `.kit.yaml`.
19. [SPEC-19] The downstream `PLAN.md` or execution metadata must preserve discovery-first routing and record `parallelization_mode: "rlm"`.
20. [SPEC-20] Any supporting documentation updated during implementation must remain consistent with the final constitution and must not introduce unverified implementation claims.

## ACCEPTANCE

1. [AC-01] `docs/CONSTITUTION.md` has no remaining HTML TODO placeholders in the required sections.
2. [AC-02] The Kit-managed baseline rules block in `docs/CONSTITUTION.md` remains present and unchanged unless regenerated by Kit.
3. [AC-03] `docs/CONSTITUTION.md` clearly states Mint's current product intent and separately states that no runtime implementation, module manifest, source tree, or test strategy exists yet.
4. [AC-04] `docs/CONSTITUTION.md` includes explicit constraints for source-of-truth order, ordered Kit workflow progression, progress summary maintenance, local/generated artifact handling, and secret safety.
5. [AC-05] `docs/CONSTITUTION.md` points to `docs/agents/*` and `docs/references/rules/*` for detailed operational procedures instead of copying their full contents.
6. [AC-06] `docs/CONSTITUTION.md` includes concrete non-goals and definitions matching this specification's scope.
7. [AC-07] `docs/PROJECT_PROGRESS_SUMMARY.md` reflects the highest completed artifact for `0001-init-project` after implementation.
8. [AC-08] `git diff --check -- docs/CONSTITUTION.md docs/PROJECT_PROGRESS_SUMMARY.md` reports no whitespace errors after the implementation.
9. [AC-09] `kit map 0001-init-project` resolves the feature and its references after implementation.
10. [AC-10] The final implementation diff contains no product code, tests, runtime config, generated artifacts, secrets, or machine-local state.

## EDGE-CASES

1. If future implementation evidence conflicts with this spec's inferred Go CLI direction, the implementation agent must treat the evidence as stronger than the inference and update the feature docs before editing the constitution.
2. If `docs/PROJECT_PROGRESS_SUMMARY.md` is stale when implementation starts, update it as supporting documentation so it reflects the highest completed artifact.
3. If a future Kit refresh changes the managed baseline block, preserve the regenerated block and avoid hand-editing inside the Kit-managed markers.
4. If no project-specific testing, tooling, or external-system guidance exists, the constitution must say that those areas require future specs rather than inventing stable rules.
5. If CodeRabbit or another automated reviewer ignores documentation files, acceptance must rely on direct file review and the validation commands in this spec.
6. If the work later expands into source code, build configuration, CI, or release behavior, stop and reclassify the new work as implementation before editing those files.
7. If a user explicitly requests GitHub delivery for this documentation feature, follow the repo-local delivery hard gate and rulesets before any issue, branch, staging, commit, push, or PR mutation.
8. If secondary global instructions conflict with repo-local docs, prefer repo-local docs for this project.
9. If no repo-local `.agents/skills/*/SKILL.md` files exist, keep the selected skill set minimal and use the repo-local RLM guide as the applicable skill-like routing input.

## OPEN-QUESTIONS

none
