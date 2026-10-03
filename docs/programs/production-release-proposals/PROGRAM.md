# Production release proposals

Coordinator: Mint. Implementation authorization from Jameson through release side-conversation handoff, 2026-10-03. No merge, cloud, credential or production activation grant.

| Repository | Issue | Branch | Spec | State | Dependencies |
| --- | --- | --- | --- | --- | --- |
| jamesonstone/mint | #13 | GH-13 | ../../specs/0009-production-release-proposals/SPEC.md | READY_SOURCE (PR #14) | publication/review |
| lsmc-bio/labcore | #1131 | GH-1131 | docs/specs/0111-mint-production-releases/SPEC.md | IN_PROGRESS | tested published Mint release |
| lsmc-bio/labcore-ui | #823 | GH-823 | docs/specs/0095-mint-production-releases/SPEC.md | IN_PROGRESS | tested published Mint release |

Ready frontier: Mint PR #14; application source delivery GH-1131/GH-823. All source adapters are prepared; adopting an exact published Mint feature release remains BLOCKED_DEPENDENCY.

Activation gates: verified feature-bearing Mint version and SHA; ready PR checks/review and explicit merge authority; configured human automation credential with narrow permissions; verified current production baseline; approved workflow/cloud activation. Missing token and baseline do not authorize provisioning or inferred records.

Last reconciliation: 2026-10-03 UTC. Mint v0.2.1 latest, no feature lifecycle. All lanes created from freshly fetched default branches. Application credentials observed by names only. No production actions performed.
