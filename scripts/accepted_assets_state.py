#!/usr/bin/env python3
"""Build the v1 durable accepted-channel state after Worker promotion."""

import argparse
import importlib.util
import json
import pathlib


HERE = pathlib.Path(__file__).resolve().parent
SPEC = importlib.util.spec_from_file_location("prepare_asset_bundle", HERE / "prepare_asset_bundle.py")
prepare = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(prepare)


def accepted_state(handoff):
    runtime = next(item for item in handoff["release_artifacts"] if item["release"] == handoff["runtime_release"])
    return {
        "schemaVersion": 1,
        "bundleDigest": handoff["candidate_bundle_digest"],
        "channelArtifactId": handoff["channel_artifact_id"],
        "channelArtifactSha256": handoff["channel_artifact_sha256"],
        "channelArtifactUrl": handoff["channel_artifact_url"],
        "release": runtime["release"],
        "releaseSha256": runtime["release_sha256"],
        "releaseUrl": runtime["release_url"],
        "resolutionDate": handoff["resolution_date"],
        "sourceRepository": handoff["assets_repository"],
        "sourceRevision": handoff["assets_revision"],
        "sourceWorkflow": handoff["assets_repository"] + "/.github/workflows/campaigns.yml",
    }


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--handoff-file", required=True, type=pathlib.Path)
    parser.add_argument("--output", required=True, type=pathlib.Path)
    args = parser.parse_args()
    handoff = prepare.read_handoff(argparse.Namespace(handoff_file=args.handoff_file, handoff_json=None))
    output = args.output
    if output.exists():
        prepare.fail(f"accepted state output already exists: {output}")
    output.parent.mkdir(parents=True, exist_ok=True)
    with output.open("x", encoding="utf-8") as state:
        json.dump(accepted_state(handoff), state, sort_keys=True, separators=(",", ":"))
        state.write("\n")


if __name__ == "__main__":
    try:
        main()
    except (OSError, ValueError, StopIteration) as error:
        raise SystemExit(f"accepted assets state: {error}")
