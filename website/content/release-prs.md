# Upgrade from Mint v1

Mint v2 removes the deployment controller. Use an immutable v1 pin for existing workflows that still require controller commands or bundled adapters.

## Move deployment to repository workflows

1. Inventory the v1 controller, its adapters, release publishers, and deployed artifact references.
2. Set up [continuous staging and manual production](environments.html) with one shared promotion job.
3. Prove both paths with the same immutable artifact. Verify who can dispatch production and how the workflow validates staging evidence.
4. Switch each environment and release tag namespace to one writer. Retain the previous controller history for audit and recovery.
5. Upgrade the release steps to a reviewed Mint v2 commit after the project's deployment path is ready.

Source versioning has explicit flags for component scope. Replace a `version-main --config` or action `version-config` argument with a tag prefix and paths. See [component versioning](versioning.html#component-releases-in-one-repository).

Before running a workflow updater on a legacy controller repository, check that it respects the v1 pin. Kit versions that refresh Mint references to the latest release can replace a v1 tag or SHA with a v2 reference.

## v1 references

Pin [Mint v1.1.2](https://github.com/jamesonstone/mint/releases/tag/v1.1.2) while operating the legacy controller. Use the versioned [environment lifecycle](https://github.com/jamesonstone/mint/blob/v1.1.2/docs/references/environment-lifecycle.md) and [compatibility adapter guide](https://github.com/jamesonstone/mint/blob/v1.1.2/docs/references/compatibility-adapters.md).
