# Integrate Mint

Add Mint to your GitHub Actions workflow to resolve a version, generate release notes, create a source tag, or publish a GitHub Release. Mint builds its CLI on the runner and makes `mint` available to later steps.

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
