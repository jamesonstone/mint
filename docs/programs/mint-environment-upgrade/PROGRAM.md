# Mint environment upgrade

Coordinator: jamesonstone/mint. Human authorization: merge prerequisite Mint PRs, implement/review/merge the environment lifecycle, publish Mint, then deliver consumer upgrade PRs with `.mint.yaml`. Consumer merges, runtime activation, cloud resources, credentials and repository settings are outside this batch.

## Dependencies and gates

Mint #12, #14 and #16 are merged in that order. GH-17 implements schema 2; current validation and independent adversarial review must pass before its ready PR merges. A published immutable Mint commit is required before consumer pins change. Consumer PRs remain unmerged; configuration does not prove deployment readiness or live state.

## Workstreams

- Mint: issue #17, branch GH-17; implementation and verification in progress.
- Existing adoption lanes to preserve: lsmc-bio/labcore #1134 (GH-1131), lsmc-bio/labcore-ui #824 (GH-823).
- New consumer lanes after publication: lsmc-bio/r2, status, lsmc-lims-connector, aquarium, terrarium, event-sink, flowcore, nautilus; appliedsymbolics/sigint; jamesonstone/kit.
- Mint itself uses its local action and a package-mode `.mint.yaml` in GH-17.

Ready frontier: Mint implementation and review. Consumer inventory is complete; consumer mutations wait for the published version. No runtime evidence or deployment claim is part of this program.

## Design boundaries

Deployment projects describe their existing environments without activating new controllers. Package and artifact publishers retain their publishing workflows and do not invent runtime baselines. Optional registry metadata derives from tracked defaults; unresolved repository variables are not asserted as actual cloud values. API/web bundles must be treated as complete artifact sets when lifecycle adapters are activated. Existing build-and-deploy workflows remain explicit migration work rather than being mislabeled as build-once promotion.
