# PLAN

Implement versioning and state primitives first; add transport and CLI adapters; integrate workflows only after their contracts are testable. Preserve legacy commands and serialize mutations.


## Actions-first revision

Refactor existing lifecycle in GH-13/PR #14: recognize hotfix conventional commits; enforce normal auto-tracking; report outcomes on the original PR; record rollback deployment ancestry; add request resolution and authenticated Actions adapter; generate repository controls; validate domain, CLI, Action and workflow contracts; update operator/agent references. No production activation or repository settings changes.

## GH-15 simplification plan

1. Audit portability and redundancy against current GitHub primary documentation.
2. Normalize policy defaults, defer bootstrap-only validation and broaden conventional hotfix intent.
3. Propagate custom config into rendered recovery workflow and validate its destination.
4. Scope CLI flags and reuse authenticated candidate evidence for scan/direct registration.
5. Test policy failures, custom setup, conventional scopes, CLI option rejection and single-download scan with all existing lifecycle tests. Document findings, then deliver a separate ready PR stacked on GH-13.

Readiness: existing artifact/approval boundaries remain explicit; the selected changes do not require schema migration or application-specific behavior.
