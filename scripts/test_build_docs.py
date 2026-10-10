"""Regression checks for project-site paths and public publishing boundaries."""
from pathlib import Path
import tempfile
import unittest

from build_docs import build, validate


class DocumentationBuildTest(unittest.TestCase):
    def test_project_and_domain_roots(self):
        for base in ("https://example.test/mint/", "https://example.test/"):
            with self.subTest(base=base), tempfile.TemporaryDirectory() as directory:
                output = Path(directory)
                build(output, base)
                self.assertIn(base + "integration.md", (output / "llms.txt").read_text())
                self.assertTrue((output / "assets/fonts/OFL.txt").is_file())
                self.assertFalse((output / "docs").exists())
                self.assertFalse((output / "AGENTS.md").exists())
                self.assertIn('aria-current="page"', (output / "index.html").read_text())

    def test_missing_anchor_fails_build_validation(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory)
            build(output, "https://example.test/mint/")
            path = output / "index.md"
            path.write_text(path.read_text() + "\n[Broken](recovery.md#does-not-exist)\n")
            with self.assertRaisesRegex(ValueError, "missing anchor"):
                validate(output, "https://example.test/mint/")

    def test_missing_asset_fails_build_validation(self):
        with tempfile.TemporaryDirectory() as directory:
            output = Path(directory)
            build(output, "https://example.test/mint/")
            (output / "assets/mint.svg").unlink()
            with self.assertRaisesRegex(ValueError, "missing or outside-site"):
                validate(output, "https://example.test/mint/")


if __name__ == "__main__":
    unittest.main()
