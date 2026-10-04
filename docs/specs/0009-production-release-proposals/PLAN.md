# PLAN

Implement versioning and state primitives first; add transport and CLI adapters; integrate workflows only after their contracts are testable. Preserve legacy commands and serialize mutations.


## Actions-first revision

Refactor existing lifecycle in GH-13/PR #14: recognize hotfix conventional commits; enforce normal auto-tracking; report outcomes on the original PR; record rollback deployment ancestry; add request resolution and authenticated Actions adapter; generate repository controls; validate domain, CLI, Action and workflow contracts; update operator/agent references. No production activation or repository settings changes.
