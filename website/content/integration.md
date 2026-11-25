# Integrate Mint

Add Mint to your GitHub Actions workflow to resolve a version, generate release notes, create a source tag, or publish a GitHub Release. Mint builds its CLI on the runner and makes `mint` available to later steps.

## Set up environments with Mint init

Run `mint init` from the repository root with a Mint release that includes the command. Choose `binary` or `image`, enter environment names in promotion order, and choose `continuous` or `manual` for each one. Press Enter to accept a default. Mint defaults to continuous `dev` and `stg` and manual `prd`.

After you confirm, Mint creates `.mint.yaml`, `.github/workflows/mint-deploy.yml`, and `.github/workflows/mint-promote.yml`. If any of those files already exists, Mint stops. Setup writes local files. The repository workflow owns credentials and dispatch.

```yaml
artifact: image
environments:
  - name: dev
    deployment: continuous
  - name: stg
    deployment: continuous
  - name: prd
    deployment: manual
```

A push to the default branch runs the project checks and retains one build. The continuous environments use that build in their listed order. In **Actions → Mint deployment → Run workflow**, select a manual environment and the exact retained version. The manual run selects the existing artifact and calls the same promotion workflow. For a manual target after the first environment, the resolver receives the preceding environment as the verification prerequisite.

Implement these scripts under `.github/mint/` before enabling the workflows. Each runs with Bash. Send diagnostic output to stderr.

| Script | Arguments | Required behavior |
| --- | --- | --- |
| `check` | None | Run the project's required checks. Exit unsuccessfully if a check fails. |
| `build` | Candidate version, source SHA | Reserve an exact version for the source SHA. Build and retain the artifact once. Record the version, source SHA, digest, and trusted producer run. Print a JSON object with the assigned `version`, immutable `ref`, and `digest` in `sha256:…` form. |
| `resolve` | Target environment, version, prerequisite environment or empty string | Look up the exact retained build. If a prerequisite is set, require a trusted successful verification run and receipt for that same build, version, and digest there. Print the retained `version`, `ref`, `digest`, and full `source_sha` as JSON. |
| `deploy` | Environment, version, ref, digest | Deploy the retained artifact. Use the immutable reference for the operation. |
| `verify` | Environment, version, ref, digest | Query the deployed runtime or destination. Run the required checks and print the observed SHA-256 digest. |

Configure GitHub environment protections and provider permissions before enabling deployment. Restrict each protected environment to the default branch and configure required reviewers where available. See [GitHub deployment protections](https://docs.github.com/en/actions/reference/workflows-and-actions/deployments-and-environments).

After the digest comparison succeeds, the shared workflow uploads `mint-verified.json` in an artifact named `mint-verified-<environment>-<run_attempt>`. When resolving prerequisite evidence, match the receipt to the repository, environment, version, source SHA, artifact reference, and digest. Use GitHub Actions to confirm that the trusted run and attempt reported successful verification. An absent, failed, cancelled, or mismatched verification blocks promotion.

The scaffold resolves a candidate version with `mint release resolve`. Successive pushes can get the same candidate before a source tag exists. Reserve the actual version atomically in the build hook or obtain it from your repository's version allocator. Reuse the reservation and any completed retained build for the same source on retry. Continuous promotion uses the assigned version returned by the build hook. The repository also owns [source tagging](versioning.html#source-version-versus-deployed-release) and GitHub Release publication.

If you change a trigger after setup, update the generated workflow and `.mint.yaml` together. Existing configurations with a top-level deployment mode and a list of environment names remain valid.

## Publish release state

<ol class="flow" role="list" aria-label="Release-state publishing">
  <li><strong>Resolve</strong><br>Read Git history for a version and release notes.</li>
  <li><strong>Tag</strong><br>Write the version file and publish the Git tag.</li>
  <li><strong>Publish</strong><br>Create or reuse the GitHub Release.</li>
</ol>

For a repository that publishes GitHub Releases from source, run this job in a workflow that starts on a trusted default-branch push. Replace `REPLACE_WITH_MINT_COMMIT_SHA` with the full commit SHA of the reviewed Mint release.

```yaml
permissions:
  contents: write
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@11d5960a326750d5838078e36cf38b85af677262
        with:
          fetch-depth: 0
      - uses: jamesonstone/mint@REPLACE_WITH_MINT_COMMIT_SHA
        with:
          command: release-publish
          commitish: ${{ github.sha }}
          owner: ${{ github.repository_owner }}
          repo: ${{ github.event.repository.name }}
          github-token: ${{ secrets.GITHUB_TOKEN }}
```

`release-publish` resolves the version, writes the numeric version to `.version` in the runner workspace, creates or reuses the annotated Git tag, and creates or reuses a GitHub Release. The job owns decisions about committing files and uploading release assets.

Use one release publisher per tag namespace. Give read-only jobs `contents: read`. Grant write access only to trusted jobs that create tags or publish releases. The action accepts an HTTPS GitHub API URL and refuses redirects to protect the token.

## Add environment promotion

<ol class="flow" role="list" aria-label="Repository-owned deployment">
  <li><strong>Build once</strong><br>Retain the artifact under its digest.</li>
  <li><strong>Verify staging</strong><br>Record successful evidence for that digest.</li>
  <li><strong>Promote and verify</strong><br>Manually dispatch production with the exact project and version.</li>
  <li><strong>Publish</strong><br>Call Mint after verification if this workflow publishes.</li>
</ol>

Keep the build and deployment jobs in your repository. Call Mint to version the exact build source before retaining the artifact. Follow the two paths in the [environment promotion guide](environments.html). If you publish after verified production, call `mint release github` for the existing tag. That command publishes the selected release without resolving a new version.

In Mint v2, pass the expected digest from the retained build record and the observed digest from the destination runtime to `mint verify`:

```bash
mint verify --expected "$BUILD_DIGEST" --observed "$RUNTIME_DIGEST"
```

Both inputs must use `sha256:` followed by 64 hexadecimal characters. Hexadecimal letter case does not affect equality. The command exits silently on a match and fails on a missing, malformed, or mismatched digest. The deployment job queries the runtime, checks every required replica or artifact, and records a receipt only after its checks pass.

Check that the source tag points to the verified source SHA. Then publish from
a job that requires successful production verification, with the selected
tag and SHA from that job's verified record:

```bash
mint release github \
  --owner "$REPOSITORY_OWNER" --repo "$REPOSITORY_NAME" \
  --tag "$VERIFIED_TAG" --target "$VERIFIED_SOURCE_SHA" \
  --notes-file "$RELEASE_NOTES_FILE"
```

Set `MINT_GITHUB_TOKEN` from the publishing job's token environment. For a
component tag, add `--tag-prefix` with the component's prefix.

## Publish Mint's own binary

Mint's repository uses a fully automatic release workflow. A push to the default branch triggers dev, stg, and prd in order. Dev runs tests and builds one versioned Linux amd64 executable with Mint v2. Staging downloads the exact Actions artifact, runs the CLI, and verifies its version and SHA-256 digest. Production downloads the same artifact and compares it with the staging digest. The promoted CLI creates or validates the source tag. GitHub Actions uploads the binary to a draft and publishes the release after the uploaded digest matches.

The repository declares its pipeline in `.mint.yaml`:

```yaml
artifact: binary
deployment: continuous
environments: [dev, stg, prd]
```

Use a Mint v2 version that supports `mint config` to print these fields as JSON. Pass `--file` to read an alternate path. The workflow validates the artifact type and stage order before it retains the build. The reader also accepts `artifact: image` and `deployment: manual` for workflows that use those modes. Repository workflows own their event triggers and jobs.

Each release published by this workflow contains a `mint-linux-amd64` asset. Download it from the release, set its executable bit, and run `./mint-linux-amd64 version`. If an upload is interrupted, rerun the failed job to resume the draft. An existing asset must match the same digest.

## Release several components from one repository

Give each independently shipped project a tag prefix and paths. For example:

```bash
mint release resolve --tag-prefix backend/ --path backend
mint release resolve --tag-prefix frontend/ --path frontend --path shared/ui
```

Use the same prefix when you create, select, or publish one project's tag. Keep each project's version-to-artifact mapping, staging evidence, and deployment permissions scoped to that project. See [scope and versioning examples](versioning.html#component-releases-in-one-repository).

## Generate an image publishing workflow

```bash
mint release workflow \
  --image name=api,uri=ghcr.io/acme/api,dockerfile=Dockerfile,context=. \
  --output .github/workflows/image-publish.yaml
```

Use the generator for a repository that builds and publishes containers to GHCR or ECR. Provide each image as an untagged repository URI. One generated workflow uses one registry kind. The generated job resolves a version, creates or reuses its source tag, builds the image, and publishes it with the version tag and `latest`. Add deployment and evidence checks in the workflows that promote the retained image. Set `--mint-ref` to the reviewed Mint commit SHA to pin the action in generated YAML.
