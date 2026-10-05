# Plan and readiness

Reuse authenticated promotion state with explicit schema 2 policy; keep legacy behavior. Implement policy/journal isolation first, lifecycle identities/configuration next, trusted observations and YAML request authorization next, then CLI/Actions and consumer templates. Validate the complete integrated contract before delivery. Generic artifact/package mode supports existing publishing commands and policy validation without claiming runtime control.

## Agent team plan

Supervisor owns Git/GitHub, canonical docs, integration, state/engine/intent/declaration, CLI/Actions request flow, validation and final delivery. Policy lane (proposal_outcomes) owns new environment policy files, config loading integration and journal isolation only. Observation lane (hotfix_prefix, after policy interface settled) owns new observation domain/adapter/tests only. Inventory lane (rollback_history) stays read-only, then verifies consumers. No overlapping file writers; shared types are supervisor-owned. Max concurrency 3 during implementation; independent adversarial read-only review after integration. Serialization: prerequisites, new branch/spec, policy interfaces, root lifecycle, observer, integration, CI/review/merge, published version, consumer delivery. Agent Git/GitHub mutations prohibited.

Readiness gate: confirmed primary workflows include image bundles and package/artifact publication. Explicit modes avoid fabricated runtime assumptions. Initial runtime activation requires authenticated bootstrap evidence. Pre-change policy governs edited YAML; no weakening on its request PR. Adapter permissions/activation stay separate.
