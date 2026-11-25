# Guided init implementation plan

## SUMMARY

Add a prompt-driven setup command for per-environment deployment triggers.

## APPROACH

1. Extend the existing config reader to accept named environment objects and preserve the legacy JSON shape. Add the guided command and write files only after confirmation.
2. Generate the default-branch build and manual dispatch workflows with one reusable deploy-and-verify job. Keep provider operations in project scripts and publish no release by default.
3. Test the config, CLI, generated job graph and syntax. Document the hook contract and verify the public site build.

## COMPONENTS

Cobra handles prompts and confirmation. Embedded templates render both GitHub Actions workflows. The config reader validates both supported file forms.

## DATA

Each environment has a name and trigger. A retained build record carries the immutable reference, digest and source SHA.

## INTERFACES

Repository scripts supply checks, artifact retention, manual resolution, deployment and runtime verification.

## DEPENDENCIES

The generated jobs run on GitHub-hosted Ubuntu with Bash, jq and the pinned Mint Action. Repository workflows supply provider authentication and permissions.

## RISKS

An incorrect project resolver can admit unverified artifacts. Document the binding it must establish and verify the observed digest before uploading a receipt.

## TESTING

Exercise both config forms, prompt cancellation, file collisions, event routing, shared promotion, and failed verification. Run native tests, workflow lint and the docs build.
