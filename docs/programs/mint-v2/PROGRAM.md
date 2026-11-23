# Mint v2 direct promotion

Mint coordinates this program. [GH-46](https://github.com/jamesonstone/mint/issues/46) tracks the initial proof and controller removal on branch `GH-46`. [Spec 0019](../../specs/0019-direct-promotion-v2/SPEC.md) records the product contract and observed validation.

| Workstream | State and dependency |
| --- | --- |
| Private GitHub Actions fixtures | Verified 2026-11-23 UTC. Continuous staging and manual production used the same retained digest for all three projects. Keep run and artifact links in Work Ledger WL-59. |
| Mint v2 | Prepared for review under GH-46. Removes the v1 controller and uses the v2 module path. Making v2 the latest release depends on a Kit refresh guard or a release strategy that keeps legacy reconciliation on v1. |
| Digest comparison | [GH-49](https://github.com/jamesonstone/mint/issues/49), based on the v2 controller-removal branch. Both fixture promotion workflows use this command before recording a receipt. |
| Provider smoke | Separate milestone after the hosted fixture proof. Needs an approved disposable registry and static site target. |
| Mint and Kit scaffolding | Separate work after fixture results settle the configuration and scaffold contract. Kit adoption needs a published, immutable Mint release. |

The ready frontier is Mint review and resolution of the latest-release dependency. Merge, consumer migration, and provider activation need their own authorization.
