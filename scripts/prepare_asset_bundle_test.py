#!/usr/bin/env python3
"""Security contracts for the Assets artifact materializer."""

import importlib.util
import pathlib
import unittest


MODULE_PATH = pathlib.Path(__file__).with_name("prepare_asset_bundle.py")
SPEC = importlib.util.spec_from_file_location("prepare_asset_bundle", MODULE_PATH)
prepare = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(prepare)


class AssetArtifactContractTest(unittest.TestCase):
    def test_accepts_exact_assets_artifact_endpoint(self):
        prepare.require_assets_artifact_url(
            "https://api.github.com/repos/araihu/assets/actions/artifacts/123456/zip",
            "123456",
        )

    def test_rejects_wrong_artifact_urls_and_ids(self):
        cases = [
            ("https://github.com/araihu/assets/actions/artifacts/123456/zip", "123456"),
            ("https://api.github.com/repos/other/assets/actions/artifacts/123456/zip", "123456"),
            ("https://api.github.com/repos/araihu/other/actions/artifacts/123456/zip", "123456"),
            ("https://api.github.com/repos/araihu/assets/actions/artifacts/999/zip", "123456"),
            ("https://api.github.com/repos/araihu/assets/actions/artifacts/123456", "123456"),
            ("https://api.github.com/repos/araihu/assets/actions/artifacts/123456/zip?download=1", "123456"),
            ("https://api.github.com/repos/araihu/assets/actions/artifacts/123456/zip", "artifact-123456"),
            ("https://api.github.com/repos/araihu/assets/actions/artifacts/0123456/zip", "0123456"),
        ]
        for url, identifier in cases:
            with self.subTest(url=url, identifier=identifier):
                with self.assertRaisesRegex(ValueError, "artifact"):
                    prepare.require_assets_artifact_url(url, identifier)

    def test_rejects_archive_traversal_forms(self):
        for name in ("../release.json", "/release.json", "campaign\\v1.js", "releases/./v1.2.3/release.json"):
            with self.subTest(name=name):
                with self.assertRaisesRegex(ValueError, "unsafe archive path"):
                    prepare.safe_member(name)


if __name__ == "__main__":
    unittest.main()
