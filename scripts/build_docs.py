#!/usr/bin/env python3
"""Build only curated public Markdown; validate every local HTML/Markdown link."""
import argparse
import html
from html.parser import HTMLParser
from pathlib import Path
import re
import shutil
from string import Template
from urllib.parse import unquote, urlsplit

import markdown

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "website"
PAGES = [
    ("index", "Overview", "START HERE"),
    ("how-it-works", "How Mint works", "CONCEPTS"),
    ("integration", "Integrate Mint", "GET STARTED"),
    ("environments", "Environments & sandboxes", "ENVIRONMENT POLICY"),
    ("versioning", "Semantic versioning", "RELEASE IDENTITY"),
    ("changelog", "CHANGELOG integration", "RELEASE NOTES"),
    ("release-prs", "Release PRs", "REVIEW & PROMOTION"),
    ("recovery", "Rollback & hotfix", "RECOVERY"),
]


class Links(HTMLParser):
    def __init__(self):
        super().__init__()
        self.targets = []
        self.ids = set()

    def handle_starttag(self, tag, attrs):
        attrs = dict(attrs)
        if "id" in attrs:
            if attrs["id"] in self.ids:
                raise ValueError(f"duplicate HTML id: {attrs['id']}")
            self.ids.add(attrs["id"])
        for key in ("href", "src"):
            if attrs.get(key):
                self.targets.append(attrs[key])


def validate(output, base_url):
    output = output.resolve()
    documents = {}
    for path in output.glob("*.html"):
        parser = Links()
        parser.feed(path.read_text())
        documents[path] = parser
    for path in output.glob("*.md"):
        parser = Links()
        parser.feed(markdown.markdown(path.read_text(), extensions=["fenced_code", "tables", "toc"]))
        documents[path] = parser
    for path, parser in documents.items():
        for target in parser.targets:
            check_link(output, path, target, base_url, documents)
    for path in output.rglob("*.css"):
        for target in re.findall(r'url\([\'"]?([^\)\'"]+)[\'"]?\)', path.read_text()):
            check_link(output, path, target, base_url, documents)
    # The llms entrypoint is part of the same published contract.
    for path in output.glob("llms*.txt"):
        for target in re.findall(r"\]\(([^)]+)\)", path.read_text()):
            check_link(output, path, target, base_url, documents)
    print(f"PASS: {len(PAGES)} pages, Markdown guides, and all local links/assets validated")


def check_link(output, path, target, base_url, documents):
    if target.startswith(base_url):
        target = target[len(base_url):]
    parsed = urlsplit(target)
    if parsed.scheme or parsed.netloc:
        return
    destination = (path.parent / unquote(parsed.path)).resolve() if parsed.path else path
    if not destination.is_relative_to(output) or not destination.is_file():
        raise ValueError(f"{path.name}: missing or outside-site link {target}")
    if parsed.fragment and destination in documents:
        if unquote(parsed.fragment) not in documents[destination].ids:
            raise ValueError(f"{path.name}: missing anchor {target}")


def build(output, base_url):
    output = output.resolve()
    # Never upload repository governance/specs or private working files.
    output.mkdir(parents=True, exist_ok=True)
    shutil.copytree(SOURCE / "assets", output / "assets", dirs_exist_ok=True)
    template = Template((SOURCE / "template.html").read_text())
    summaries = []
    full_text = []
    for index, (slug, label, category) in enumerate(PAGES):
        text = (SOURCE / "content" / f"{slug}.md").read_text()
        converter = markdown.Markdown(extensions=["fenced_code", "tables", "toc"],
                                      extension_configs={"toc": {"toc_depth": "2-2"}})
        body = converter.convert(text)
        body = body.replace("<table>", '<div class="table-wrap" role="region" aria-label="Reference table" tabindex="0"><table>')
        body = body.replace("</table>", "</table></div>")
        title = text.splitlines()[0].removeprefix("# ")
        description = text.split("\n\n", 2)[1].replace("\n", " ")
        navigation = ""
        for other, name, _ in PAGES:
            current = ' aria-current="page"' if other == slug else ""
            navigation += f'<a href="{other}.html"{current}>{html.escape(name)}</a>'
        previous = adjacent(index - 1, "← ") if index else "<span></span>"
        following = adjacent(index + 1, "", " →") if index + 1 < len(PAGES) else ""
        page = template.substitute(title=html.escape(title), description=html.escape(description, quote=True),
                                   canonical=html.escape(base_url + slug + ".html", quote=True),
                                   category=category, source=slug + ".md", navigation=navigation,
                                   content=body, toc=converter.toc, previous=previous, next=following)
        (output / f"{slug}.html").write_text(page)
        # Raw Markdown links point to Markdown, not HTML, for agent traversal.
        raw = re.sub(r'(?<=\()([a-z-]+)\.html', r'\1.md', text)
        raw = re.sub(r'(?<=href=")([a-z-]+)\.html', r'\1.md', raw)
        (output / f"{slug}.md").write_text(raw)
        summaries.append(f"- [{label}]({base_url}{slug}.md): {description}")
        full_text.append(raw)
    (output / "llms.txt").write_text(
        "# Mint\n\n> Mint is a Go CLI and GitHub Action for source versioning, immutable artifacts, "
        "reviewed environment releases, and verified recovery.\n\n"
        "Mint supports arbitrary named environments; production is optional. Your project owns "
        "build/deployment adapters and infrastructure. Source tags, verified deployment, and live "
        "observations are separate evidence.\n\n## Guides\n\n" + "\n".join(summaries) +
        f"\n\n## Optional\n\n- [Complete documentation]({base_url}llms-full.txt): All guides in Markdown.\n"
        "- [Source and releases](https://github.com/jamesonstone/mint): Implementation and published versions.\n"
    )
    combined = "\n\n---\n\n".join(full_text)
    # Full context uses absolute links; fragments remain attached to their guide.
    combined = re.sub(r'\]\(([a-z-]+\.(?:md|txt)(?:#[^)]+)?)\)',
                      lambda match: "](" + base_url + match[1] + ")", combined)
    combined = re.sub(r'href="([a-z-]+\.md)"',
                      lambda match: 'href="' + base_url + match[1] + '"', combined)
    (output / "llms-full.txt").write_text(combined)
    (output / ".nojekyll").touch()
    validate(output, base_url)


def adjacent(index, prefix, suffix=""):
    slug, label, _ = PAGES[index]
    return f'<a href="{slug}.html">{prefix}{html.escape(label)}{suffix}</a>'


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", type=Path, default=ROOT / "_site")
    parser.add_argument("--base-url", default="https://jamesonstone.github.io/mint/")
    args = parser.parse_args()
    base = urlsplit(args.base_url)
    if base.scheme not in ("http", "https") or not base.netloc or base.query or base.fragment:
        parser.error("base URL must be an absolute HTTP(S) site URL without query or fragment")
    build(args.output.resolve(), args.base_url.rstrip("/") + "/")


if __name__ == "__main__":
    main()
