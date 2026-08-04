#!/usr/bin/env python3
"""Contracts for durable post-promotion accepted-channel state."""

import importlib.util
import pathlib
import unittest


HERE = pathlib.Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("accepted_assets_state", HERE / "accepted_assets_state.py")
state_module = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(state_module)


class AcceptedAssetsStateTest(unittest.TestCase):
    def test_matches_assets_accepted_channel_v1_schema_and_provenance(self):
        handoff = {
            "assets_repository": "araihu/assets",
            "assets_revision": "a" * 40,
            "release_artifacts": [{
                "release": "v1.2.3",
                "release_url": "https://github.com/araihu/assets/releases/download/v1.2.3/araihu-assets-v1.2.3.tar.gz",
                "release_sha256": "b" * 64,
            }],
            "runtime_release": "v1.2.3",
            "channel_artifact_id": 123456,
            "channel_artifact_url": "https://github.com/araihu/assets/actions/runs/987654/artifacts/123456",
            "channel_artifact_sha256": "c" * 64,
            "candidate_bundle_digest": "d" * 64,
            "resolution_date": "2026-07-30",
            "state_ref": "automation/araihu-assets-state",
            "state_path": ".automation/araihu-assets/accepted-channel-v1.json",
        }
        accepted = state_module.accepted_state(state_module.prepare.validate_handoff(handoff))
        self.assertEqual(set(accepted), {
            "schemaVersion", "bundleDigest", "channelArtifactId", "channelArtifactSha256", "channelArtifactUrl",
            "release", "releaseSha256", "releaseUrl", "resolutionDate", "sourceRepository", "sourceRevision", "sourceWorkflow",
        })
        self.assertEqual(accepted["bundleDigest"], "d" * 64)
        self.assertEqual(accepted["channelArtifactId"], 123456)
        self.assertEqual(accepted["sourceRepository"], "araihu/assets")
        self.assertEqual(accepted["sourceWorkflow"], "araihu/assets/.github/workflows/campaigns.yml")


if __name__ == "__main__":
    unittest.main()
