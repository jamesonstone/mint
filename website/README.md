# Public documentation site

The public site lives at `https://jamesonstone.github.io/mint/`. Its site-root agent
entrypoint is `/mint/llms.txt` on this project URL, with full context at
`/mint/llms-full.txt`. A project deployment cannot install a file at the owning
user site's domain-root `/llms.txt`; that requires the user-site repository or a
custom domain. No custom domain is assumed here.

## Local development

```bash
python3 -m venv .venv-docs
source .venv-docs/bin/activate
python3 -m pip install -r website/requirements.txt
make docs-check
make docs-serve
```

Open `http://127.0.0.1:8000/`. The default build uses the public canonical URL for
agent links. To test those links locally, build with an explicit local base:

```bash
python3 scripts/build_docs.py --base-url http://127.0.0.1:8000/
python3 -m http.server 8000 --bind 127.0.0.1 --directory _site
```

Edit the curated Markdown in `content/`, the shared template, or assets. Navigation,
canonical links, Markdown copies, `llms.txt`, and `llms-full.txt` are generated from
one page list in `scripts/build_docs.py`. The build validates local assets, page
links, and anchors, including agent Markdown. Build output `_site/` is ignored.
Do not add repository governance, feature specs, or internal runbooks to the public
page list. Use a fresh output directory in CI.

## GitHub Pages activation

The `Documentation` workflow validates PRs and uploads/deploys only on `main`.
After review and merge, the repository owner must enable **Settings → Pages →
Build and deployment → Source: GitHub Actions**, then run the workflow or push a
site update. Repository instructions prohibit agents from changing settings or
merging. A successful local build or PR check does not prove live publication.

The deployment job has only `pages: write` and `id-token: write` in addition to
repository read access. The build uses immutable action pins and a pinned Markdown
dependency. No runtime JS, external CSS, CDN fonts, or application secrets are needed.
If adopting a custom domain, update the builder's canonical base URL in the workflow
and configure the domain through the approved repository settings process.

## Typography

JetBrains Mono regular and bold WOFF2 are self-hosted under the SIL Open Font
License. The fonts and `OFL.txt` come from `JetBrains/JetBrainsMono` commit
`19371302b95d218af43299bce79ddbddd0bc364d`. Retain the license when updating them.
