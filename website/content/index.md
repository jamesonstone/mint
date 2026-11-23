# Build once. Release with evidence.

Mint is a Go CLI and GitHub Action for Semantic Versioning, changelogs, Git tags, and GitHub Releases. Build your application once, then use your repository workflows to promote the same artifact.

<div class="hero-actions"><a class="button" href="integration.html">Integrate Mint →</a><a class="secondary" href="environments.html">Promote an artifact</a></div>

## One artifact across environments

Your build workflow retains an artifact by digest and records its source SHA and version. Continuous staging verifies that artifact. An authorized production dispatch selects the exact project and version and requires matching staging evidence before it can deploy.

<ol class="flow" role="list" aria-label="Release flow">
  <li><strong>Build once</strong><br>Retain the artifact by digest.</li>
  <li><strong>Verify staging</strong><br>Record the digest that passed staging.</li>
  <li><strong>Promote to production</strong><br>Deploy that same digest and verify it.</li>
  <li><strong>Publish</strong><br>Call Mint after production verification.</li>
</ol>

## Start here

- [How Mint works](how-it-works.html) explains versioning, artifact identity, and publication.
- [Integrate Mint](integration.html) shows how to add Mint to a GitHub Actions workflow.
- [Environment promotion](environments.html) covers continuous staging and manual production.
- [Semantic versioning](versioning.html) explains version rules and component scopes.
- [CHANGELOG integration](changelog.html) covers release notes from Conventional Commits.
- [Upgrade from v1](release-prs.html) provides migration guidance for controller users.
- [Recovery](recovery.html) covers artifact rollback and publication retries.

[llms.txt](llms.txt) and [llms-full.txt](llms-full.txt) provide the guides in Markdown.
