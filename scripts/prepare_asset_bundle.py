#!/usr/bin/env python3
"""Validate, download, and safely assemble an Assets release handoff."""

import argparse
import datetime
import hashlib
import json
import os
import pathlib
import re
import struct
import subprocess
import sys
import tarfile
import tempfile
import zipfile


ASSETS_REPOSITORY = "araihu/assets"
SHA256 = re.compile(r"[0-9a-f]{64}\Z")
GIT_REVISION = re.compile(r"[0-9a-f]{40}\Z")
POSITIVE_ID = re.compile(r"[1-9][0-9]*\Z")
SEMVER = re.compile(
    r"v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)"
    r"(?:-(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*)?"
    r"(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?\Z"
)
STATE_REF = "automation/araihu-assets-state"
STATE_PATH = ".automation/araihu-assets/accepted-channel-v1.json"
CHANNEL_FILES = {
    "campaign/v1.js",
    "releases/latest.json",
    "releases/default.json",
    "releases/current.json",
}
HANDOFF_FIELDS = {
    "assets_repository",
    "assets_revision",
    "release_artifacts",
    "runtime_release",
    "channel_artifact_id",
    "channel_artifact_url",
    "channel_artifact_sha256",
    "candidate_bundle_digest",
    "resolution_date",
    "state_ref",
    "state_path",
}
RELEASE_ARTIFACT_FIELDS = {"release", "release_url", "release_sha256"}


def fail(message):
    raise ValueError(message)


def require_string(value, name):
    if not isinstance(value, str) or not value or "\r" in value or "\n" in value:
        fail(f"{name} is invalid")
    return value


def require_sha256(value, name):
    value = require_string(value, name)
    if not SHA256.fullmatch(value):
        fail(f"{name} is invalid")
    return value


def require_semver(value, name):
    value = require_string(value, name)
    if not SEMVER.fullmatch(value):
        fail(f"{name} is not strict SemVer")
    return value


def require_state_ref(value):
    if value != STATE_REF:
        fail("state_ref is invalid")
    return STATE_REF


def require_state_path(value):
    if value != STATE_PATH:
        fail("state_path is invalid")
    return STATE_PATH


def release_url(release):
    return f"https://github.com/araihu/assets/releases/download/{release}/araihu-assets-{release}.tar.gz"


def channel_api_url(identifier):
    return f"https://api.github.com/repos/araihu/assets/actions/artifacts/{identifier}/zip"


def require_channel_artifact_id(value):
    if isinstance(value, bool) or not isinstance(value, int) or value <= 0:
        fail("channel_artifact_id is invalid")
    return value


def require_channel_artifact_url(url, identifier):
    url = require_string(url, "channel_artifact_url")
    expected = f"https://github.com/araihu/assets/actions/runs/[1-9][0-9]*/artifacts/{identifier}"
    if not re.fullmatch(expected, url):
        fail("channel_artifact_url must be the matching araihu/assets Actions HTML artifact URL")
    return url


def validate_handoff(payload):
    if not isinstance(payload, dict) or set(payload) != HANDOFF_FIELDS:
        fail("handoff fields are invalid")
    if payload["assets_repository"] != ASSETS_REPOSITORY:
        fail("assets_repository is invalid")
    if not GIT_REVISION.fullmatch(require_string(payload["assets_revision"], "assets_revision")):
        fail("assets_revision is invalid")

    artifacts = payload["release_artifacts"]
    if not isinstance(artifacts, list) or not artifacts:
        fail("release_artifacts is invalid")
    validated_artifacts = []
    releases = set()
    for index, artifact in enumerate(artifacts):
        if not isinstance(artifact, dict) or set(artifact) != RELEASE_ARTIFACT_FIELDS:
            fail(f"release_artifacts[{index}] fields are invalid")
        release = require_semver(artifact["release"], f"release_artifacts[{index}].release")
        if release in releases:
            fail("release_artifacts repeats a release")
        releases.add(release)
        if artifact["release_url"] != release_url(release):
            fail("release_artifact URL is invalid")
        validated_artifacts.append(
            {
                "release": release,
                "release_url": artifact["release_url"],
                "release_sha256": require_sha256(artifact["release_sha256"], "release_artifact SHA-256"),
            }
        )

    runtime_release = require_semver(payload["runtime_release"], "runtime_release")
    if runtime_release not in releases:
        fail("runtime_release is not present in release_artifacts")
    identifier = require_channel_artifact_id(payload["channel_artifact_id"])
    require_channel_artifact_url(payload["channel_artifact_url"], identifier)
    digest = require_sha256(payload["candidate_bundle_digest"], "candidate_bundle_digest")
    resolution_date = require_string(payload["resolution_date"], "resolution_date")
    try:
        if datetime.date.fromisoformat(resolution_date).isoformat() != resolution_date:
            fail("resolution_date is invalid")
    except ValueError:
        fail("resolution_date is invalid")

    return {
        "assets_repository": ASSETS_REPOSITORY,
        "assets_revision": payload["assets_revision"],
        "release_artifacts": validated_artifacts,
        "runtime_release": runtime_release,
        "channel_artifact_id": identifier,
        "channel_artifact_url": payload["channel_artifact_url"],
        "channel_artifact_sha256": require_sha256(payload["channel_artifact_sha256"], "channel_artifact SHA-256"),
        "candidate_bundle_digest": digest,
        "resolution_date": resolution_date,
        "state_ref": require_state_ref(payload["state_ref"]),
        "state_path": require_state_path(payload["state_path"]),
    }


def read_handoff(args):
    if args.handoff_file:
        try:
            source = args.handoff_file.read_text(encoding="utf-8")
        except OSError as error:
            fail(f"read handoff: {error}")
    else:
        source = args.handoff_json
    try:
        return validate_handoff(json.loads(source))
    except json.JSONDecodeError as error:
        fail(f"handoff JSON is invalid: {error.msg}")


def safe_member(name):
    if not name or "\\" in name or name.startswith("/"):
        fail(f"unsafe archive path {name!r}")
    raw_parts = name.split("/")
    if any(part in {"", ".", ".."} for part in raw_parts):
        fail(f"unsafe archive path {name!r}")
    return pathlib.PurePosixPath(name).parts


def read_archive(path):
    if zipfile.is_zipfile(path):
        with zipfile.ZipFile(path) as archive:
            for info in archive.infolist():
                name = info.filename.rstrip("/")
                if not name:
                    continue
                safe_member(name)
                mode = info.external_attr >> 16
                if info.is_dir():
                    yield name, True, None
                elif mode and (mode & 0o170000) not in {0, 0o100000}:
                    fail(f"archive member {name!r} is a symbolic link or special file")
                else:
                    yield name, False, archive.read(info)
        return
    if tarfile.is_tarfile(path):
        with tarfile.open(path) as archive:
            for info in archive.getmembers():
                name = info.name.rstrip("/")
                if not name:
                    continue
                safe_member(name)
                if info.isdir():
                    yield name, True, None
                elif not info.isreg():
                    fail(f"archive member {name!r} is a symbolic link or special file")
                else:
                    source = archive.extractfile(info)
                    if source is None:
                        fail(f"archive member {name!r} cannot be read")
                    yield name, False, source.read()
        return
    fail("archive must be ZIP or tar")


def strict_json(data, name):
    def no_duplicates(pairs):
        result = {}
        for key, value in pairs:
            if key in result:
                fail(f"{name} has duplicate JSON field {key!r}")
            result[key] = value
        return result

    try:
        value = json.loads(data, object_pairs_hook=no_duplicates)
    except (UnicodeDecodeError, json.JSONDecodeError) as error:
        fail(f"{name} is invalid JSON: {error}")
    if not isinstance(value, dict):
        fail(f"{name} is not a JSON object")
    return value


def parse_checksums(data):
    try:
        source = data.decode("utf-8")
    except UnicodeDecodeError as error:
        fail(f"checksums.txt is not UTF-8: {error}")
    if not source.endswith("\n"):
        fail("checksums.txt is not newline terminated")
    entries = {}
    for line in source[:-1].split("\n"):
        digest, separator, name = line.partition("  ")
        if not separator or not SHA256.fullmatch(digest) or not name or name in entries:
            fail(f"checksums.txt has invalid entry {line!r}")
        safe_member(name)
        entries[name] = digest
    return entries


def validate_flat_release(files, expected_release):
    if "release.json" not in files or "checksums.txt" not in files:
        fail("immutable release archive misses release metadata")
    release = strict_json(files["release.json"], "release.json")
    if release.get("release") != expected_release:
        fail("immutable release archive does not match its handoff release")
    checksums = parse_checksums(files["checksums.txt"])
    expected_files = set(files) - {"checksums.txt"}
    if set(checksums) != expected_files:
        fail("checksums.txt does not cover exactly the immutable release archive")
    for name, data in files.items():
        if name == "checksums.txt":
            continue
        actual = hashlib.sha256(data).hexdigest()
        if checksums[name] != actual:
            fail(f"checksums.txt digest mismatch for {name!r}")


def archive_files(path, kind, expected_release=None):
    files = {}
    for name, directory, contents in read_archive(path):
        if directory:
            if kind == "release":
                fail(f"immutable release archive has directory {name!r}")
            continue
        if name in files:
            fail(f"archive repeats path {name!r}")
        if kind == "release":
            if name.startswith("releases/"):
                fail(f"immutable release archive must have flat root, got {name!r}")
        elif name not in CHANNEL_FILES:
            fail(f"unexpected archive root {name!r} for channel")
        files[name] = contents
    if kind == "release":
        validate_flat_release(files, expected_release)
    elif set(files) != CHANNEL_FILES:
        fail("channel archive must contain campaign/v1.js and all three channel documents")
    return files


def extract(path, destination, kind, expected_release=None):
    files = archive_files(path, kind, expected_release)
    targets = {}
    for name, contents in files.items():
        target = destination / "releases" / expected_release / name if kind == "release" else destination.joinpath(*safe_member(name))
        if target.exists():
            fail(f"archive would overwrite {name!r}")
        targets[name] = (target, contents)
    for name, (target, contents) in targets.items():
        target.parent.mkdir(parents=True, exist_ok=True)
        with target.open("xb") as output:
            output.write(contents)


def channel_bundle_digest(root):
    digest = hashlib.sha256()
    digest.update(b"araihu-channel-bundle-v1\0")
    for name in ("campaign/v1.js", "releases/latest.json", "releases/default.json", "releases/current.json"):
        data = (root / name).read_bytes()
        encoded = name.encode("utf-8")
        digest.update(struct.pack(">Q", len(encoded)))
        digest.update(encoded)
        digest.update(struct.pack(">Q", len(data)))
        digest.update(data)
    return digest.hexdigest()


def validate_channel_bundle_digest(root, expected):
    if channel_bundle_digest(root) != expected:
        fail("candidate_bundle_digest does not match canonical channel bundle")


def download(url, destination, token):
    command = [
        "curl", "--proto", "=https", "--proto-redir", "=https", "--location", "--fail",
        "--silent", "--show-error", "--output", str(destination), url,
    ]
    if token:
        command[1:1] = ["--header", f"Authorization: Bearer {token}"]
    subprocess.run(command, check=True)


def verified_download(url, filename, expected, temporary, token):
    archive = temporary / filename
    download(url, archive, token)
    with archive.open("rb") as contents:
        actual = hashlib.file_digest(contents, "sha256").hexdigest()
    if actual != expected:
        fail(f"SHA-256 mismatch for asset {filename!r}")
    return archive


def write_handoff(path, handoff):
    if path.exists():
        fail(f"accepted handoff output already exists: {path}")
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("x", encoding="utf-8") as output:
        os.chmod(path, 0o600)
        json.dump(handoff, output, sort_keys=True, separators=(",", ":"))
        output.write("\n")


def main():
    parser = argparse.ArgumentParser()
    source = parser.add_mutually_exclusive_group(required=True)
    source.add_argument("--handoff-json")
    source.add_argument("--handoff-file", type=pathlib.Path)
    parser.add_argument("--accepted-output", type=pathlib.Path)
    parser.add_argument("--output", required=True, type=pathlib.Path)
    args = parser.parse_args()
    if args.output.exists():
        fail(f"asset bundle output already exists: {args.output}")
    handoff = read_handoff(args)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    token = os.environ.get("ASSETS_GITHUB_TOKEN", "")
    with tempfile.TemporaryDirectory(prefix="ahairu-assets-", dir=args.output.parent) as temporary_name:
        temporary = pathlib.Path(temporary_name)
        staged = temporary / "bundle"
        staged.mkdir()
        for artifact in handoff["release_artifacts"]:
            archive = verified_download(
                artifact["release_url"],
                f"release-{artifact['release']}.tar.gz",
                artifact["release_sha256"],
                temporary,
                token,
            )
            extract(archive, staged, "release", artifact["release"])
        channel = verified_download(
            channel_api_url(handoff["channel_artifact_id"]),
            f"channel-{handoff['channel_artifact_id']}.zip",
            handoff["channel_artifact_sha256"],
            temporary,
            token,
        )
        extract(channel, staged, "channel")
        validate_channel_bundle_digest(staged, handoff["candidate_bundle_digest"])
        os.replace(staged, args.output)
    if args.accepted_output:
        write_handoff(args.accepted_output, handoff)


if __name__ == "__main__":
    try:
        main()
    except (OSError, subprocess.CalledProcessError, ValueError, tarfile.TarError, zipfile.BadZipFile) as error:
        print(f"prepare asset bundle: {error}", file=sys.stderr)
        sys.exit(1)
