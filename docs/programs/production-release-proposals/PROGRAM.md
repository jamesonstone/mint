# Production release proposals

Coordinator: Mint. Implementation authorization from Jameson through release side-conversation handoff, 2026-10-03. No merge, cloud, credential or production activation grant.

| Repository | Issue | Branch | Spec | State | Dependencies |
| --- | --- | --- | --- | --- | --- |
| jamesonstone/mint | #13 | GH-13 | ../../specs/0009-production-release-proposals/SPEC.md | READY_SOURCE (PR #14) | publication/review |
| lsmc-bio/labcore | #1131 | GH-1131 | docs/specs/0111-mint-production-releases/SPEC.md | READY_SOURCE (PR #1134) | tested published Mint release |
| lsmc-bio/labcore-ui | #823 | GH-823 | docs/specs/0095-mint-production-releases/SPEC.md | READY_SOURCE (PR #824) | tested published Mint release |

Ready frontier: Mint PR #14; application source delivery GH-1131/GH-823. All source adapters are prepared; adopting an exact published Mint feature release remains BLOCKED_DEPENDENCY.

Activation gates: verified feature-bearing Mint version and SHA; ready PR checks/review and explicit merge authority; job-scoped built-in token and Actions PR-creation setting; verified current production baseline; approved workflow/cloud activation. Missing PR-creation permission and baseline do not authorize provisioning or inferred records.

Last reconciliation: 2026-10-03 UTC. Mint v0.2.1 latest, no feature lifecycle. All lanes created from freshly fetched default branches. Both Actions settings currently disable PR creation; no new credential is required. Companion internal docs: [docs PR #429](https://github.com/lsmc-bio/docs/pull/429), GH-428. Mint hosted native CI passed at tested feature source 4230ef0b8bfc1bd1e76332de5cb847c818fc7c79; application hosted CI is still in progress. No production actions performed.
