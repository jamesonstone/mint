# Release with confidence.

Mint gives every release a clear path from source to artifact to environment. Build once, review an exact selection, and keep evidence of what actually ran.

<div class="hero-actions"><a class="button" href="integration.html">Integrate Mint →</a><a class="secondary" href="environments.html">Explore environments</a></div>

## One artifact. Any environment.

Mint is a Go CLI and GitHub Action for release state. It versions source, records trusted immutable builds, coordinates reviewed environment releases, and records verified outcomes. Your application owns the build and deployment adapters.

<div class="flow" aria-label="Release lifecycle"><span>Source</span><b>→</b><span>Build</span><b>→</b><span>Review</span><b>→</b><span>Verify</span></div>

A sandbox can use the same digest as staging or a customer demo. No environment named production is required. Each environment has its own policy and deployment history; a multi-component application moves as a complete digest bundle.

## Start here

- [How Mint works](how-it-works.html): source versions, candidates, frozen intent, and evidence.
- [Integrate Mint](integration.html): choose a release model, add policy, and connect trusted adapters.
- [Environments & sandboxes](environments.html): local versus shared scope, arbitrary names, and promotion prerequisites.
- [Semantic versioning](versioning.html): understand what increments a version and what a tag proves.
- [CHANGELOG integration](changelog.html): generate linked release notes from Conventional Commits.
- [Release PRs](release-prs.html): review the selection that will deploy.
- [Rollback & hotfix](recovery.html): recover without losing artifact identity or deployment evidence.

## Know what is proven

| Evidence | What it means |
| --- | --- |
| Source tag | A source commit has a version identity. |
| Trusted candidate | A successful build produced an eligible immutable artifact. |
| Approved Release PR | Humans approved an exact selection and runtime configuration. |
| Verified deployment | The adapter verified that exact selection in the environment. |
| Live observation | Timestamped evidence of the runtime at observation time. |

A merge, dispatch, image alias, or GitHub Release alone does not establish live runtime state.

## Read with an agent

Use [llms.txt](llms.txt) for a concise map, [llms-full.txt](llms-full.txt) for the complete guides, or the **Markdown source** link on any page. The site works without JavaScript and loads its fonts locally.
