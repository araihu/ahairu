#!/usr/bin/env python3
"""Security and schema contracts for the Assets handoff materializer."""

import importlib.util
import io
import hashlib
import json
import pathlib
import tarfile
import tempfile
import unittest
import zipfile


MODULE_PATH = pathlib.Path(__file__).with_name("prepare_asset_bundle.py")
SPEC = importlib.util.spec_from_file_location("prepare_asset_bundle", MODULE_PATH)
prepare = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(prepare)


def handoff():
    release = "v1.2.3"
    return {
        "assets_repository": "araihu/assets",
        "assets_revision": "a" * 40,
        "release_artifacts": [{
            "release": release,
            "release_url": prepare.release_url(release),
            "release_sha256": "b" * 64,
        }],
        "runtime_release": release,
        "channel_artifact_id": 123456,
        "channel_artifact_url": "https://github.com/araihu/assets/actions/runs/987654/artifacts/123456",
        "channel_artifact_sha256": "c" * 64,
        "candidate_bundle_digest": "d" * 64,
        "resolution_date": "2026-07-30",
        "state_ref": "automation/araihu-assets-state",
        "state_path": ".automation/araihu-assets/accepted-channel-v1.json",
    }


def write_release_archive(path, release):
    release_json = json.dumps({"release": release}, separators=(",", ":")).encode() + b"\n"
    files = {"release.json": release_json, "campaign/v1.js": b"runtime\n"}
    checksums = b"".join(
        hashlib.sha256(contents).hexdigest().encode() + b"  " + name.encode() + b"\n"
        for name, contents in sorted(files.items())
    )
    with tarfile.open(path, "w:gz") as archive:
        for filename, contents in {**files, "checksums.txt": checksums}.items():
            info = tarfile.TarInfo(filename)
            info.size = len(contents)
            archive.addfile(info, io.BytesIO(contents))


def write_channel_archive(path):
    with zipfile.ZipFile(path, "w") as archive:
        for filename in prepare.CHANNEL_FILES:
            archive.writestr(filename, "{}\n")


class AssetHandoffContractTest(unittest.TestCase):
    def test_accepts_authoritative_handoff_schema(self):
        self.assertEqual(prepare.validate_handoff(handoff()), handoff())
        self.assertIsInstance(handoff()["channel_artifact_id"], int)

    def test_rejects_cross_schema_and_malicious_urls(self):
        cases = []
        wrong_repo = handoff()
        wrong_repo["assets_repository"] = "araihu/not-assets"
        cases.append((wrong_repo, "assets_repository"))
        legacy_fields = handoff()
        legacy_fields["assets_release_url"] = "https://example.test"
        cases.append((legacy_fields, "handoff fields"))
        bad_release_url = handoff()
        bad_release_url["release_artifacts"][0]["release_url"] = "https://github.com/other/assets/releases/download/v1.2.3/araihu-assets-v1.2.3.tar.gz"
        cases.append((bad_release_url, "release_artifact URL"))
        bad_channel_url = handoff()
        bad_channel_url["channel_artifact_url"] = "https://github.com/araihu/assets/actions/runs/987654/artifacts/999999"
        cases.append((bad_channel_url, "channel_artifact_url"))
        bad_channel_id = handoff()
        bad_channel_id["channel_artifact_id"] = "123456"
        cases.append((bad_channel_id, "channel_artifact_id"))
        bad_runtime = handoff()
        bad_runtime["runtime_release"] = "v9.9.9"
        cases.append((bad_runtime, "runtime_release"))
        for value, message in cases:
            with self.subTest(message=message):
                with self.assertRaisesRegex(ValueError, message):
                    prepare.validate_handoff(value)

    def test_rejects_noncanonical_field_values(self):
        cases = [
            ("assets_revision", "A" * 40, "assets_revision"),
            ("runtime_release", "v01.2.3", "SemVer"),
            ("runtime_release", "v1.2.3-01", "SemVer"),
            ("candidate_bundle_digest", "D" * 64, "candidate_bundle_digest"),
            ("candidate_bundle_digest", "sha256:" + "d" * 64, "candidate_bundle_digest"),
            ("resolution_date", "2026-02-30", "resolution_date"),
            ("state_ref", "refs/heads/main", "state_ref"),
            ("state_path", "../release-state.json", "state_path"),
        ]
        for key, value, message in cases:
            with self.subTest(key=key):
                candidate = handoff()
                candidate[key] = value
                with self.assertRaisesRegex(ValueError, message):
                    prepare.validate_handoff(candidate)

    def test_multiple_release_archives_materialize_a_cumulative_bundle(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            first = root / "first.tar.gz"
            second = root / "second.tar.gz"
            channel = root / "channel.zip"
            destination = root / "bundle"
            destination.mkdir()
            write_release_archive(first, "v1.2.3")
            write_release_archive(second, "v1.2.4")
            write_channel_archive(channel)
            prepare.extract(first, destination, "release", "v1.2.3")
            prepare.extract(second, destination, "release", "v1.2.4")
            prepare.extract(channel, destination, "channel")
            self.assertTrue((destination / "releases/v1.2.3/release.json").is_file())
            self.assertTrue((destination / "releases/v1.2.4/release.json").is_file())
            self.assertTrue((destination / "campaign/v1.js").is_file())
            self.assertFalse((destination / "releases/v1.2.3/releases").exists())

    def test_canonical_channel_digest_matches_assets_v1(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            (root / "campaign").mkdir()
            (root / "releases").mkdir()
            (root / "campaign/v1.js").write_bytes(b"runtime-v1\n")
            (root / "releases/latest.json").write_bytes(b'{"channel":"latest"}\n')
            (root / "releases/default.json").write_bytes(b'{"channel":"default"}\n')
            (root / "releases/current.json").write_bytes(b'{"channel":"current"}\n')
            self.assertEqual(
                prepare.channel_bundle_digest(root),
                "c5b3d30fba9fa1bf6b932d61d4299e4351637a14727aa845b588e323296178db",
            )
            (root / "releases/latest.json").write_bytes(b'{"channel":"changed"}\n')
            with self.assertRaisesRegex(ValueError, "candidate_bundle_digest"):
                prepare.validate_channel_bundle_digest(
                    root,
                    "c5b3d30fba9fa1bf6b932d61d4299e4351637a14727aa845b588e323296178db",
                )

    def test_rejects_archive_traversal_forms(self):
        for name in ("../release.json", "/release.json", "campaign\\v1.js", "releases/./v1.2.3/release.json"):
            with self.subTest(name=name):
                with self.assertRaisesRegex(ValueError, "unsafe archive path"):
                    prepare.safe_member(name)

    def test_rejects_release_archive_for_a_different_tag(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            archive = root / "release.tar.gz"
            destination = root / "bundle"
            destination.mkdir()
            write_release_archive(archive, "v1.2.4")
            with self.assertRaisesRegex(ValueError, "does not match"):
                prepare.extract(archive, destination, "release", "v1.2.3")

    def test_rejects_wrapped_or_checksum_tampered_flat_release_archive(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            destination = root / "bundle"
            destination.mkdir()
            wrapped = root / "wrapped.tar.gz"
            with tarfile.open(wrapped, "w:gz") as archive:
                contents = b"{}\n"
                info = tarfile.TarInfo("releases/v1.2.3/release.json")
                info.size = len(contents)
                archive.addfile(info, io.BytesIO(contents))
            with self.assertRaisesRegex(ValueError, "flat root"):
                prepare.extract(wrapped, destination, "release", "v1.2.3")

            tampered = root / "tampered.tar.gz"
            write_release_archive(tampered, "v1.2.3")
            with tarfile.open(tampered, "r:gz") as source, tarfile.open(root / "bad.tar.gz", "w:gz") as archive:
                for item in source:
                    data = source.extractfile(item).read() if item.isfile() else None
                    if item.name == "campaign/v1.js":
                        data = b"tampered\n"
                        item.size = len(data)
                    archive.addfile(item, io.BytesIO(data) if data is not None else None)
            with self.assertRaisesRegex(ValueError, "digest mismatch"):
                prepare.extract(root / "bad.tar.gz", destination, "release", "v1.2.3")

    def test_rejects_symbolic_link_archive_members(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory)
            archive = root / "release.tar.gz"
            destination = root / "bundle"
            destination.mkdir()
            with tarfile.open(archive, "w:gz") as contents:
                link = tarfile.TarInfo("releases/v1.2.3/link")
                link.type = tarfile.SYMTYPE
                link.linkname = "release.json"
                contents.addfile(link)
            with self.assertRaisesRegex(ValueError, "symbolic link or special file"):
                prepare.extract(archive, destination, "release", "v1.2.3")


if __name__ == "__main__":
    unittest.main()
