#!/usr/bin/env python3
"""Download, verify, and safely assemble one Assets release/channel bundle."""

import argparse
import hashlib
import os
import pathlib
import re
import subprocess
import sys
import tarfile
import tempfile
import zipfile


SHA256 = re.compile(r"[0-9a-f]{64}\Z")
IDENTIFIER = re.compile(r"[A-Za-z0-9][A-Za-z0-9._-]{0,127}\Z")
RELEASE = re.compile(r"v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?\Z")
CHANNEL_FILES = {
    "campaign/v1.js",
    "releases/latest.json",
    "releases/default.json",
    "releases/current.json",
}


def fail(message):
    raise ValueError(message)


def safe_member(name):
    if not name or "\\" in name or name.startswith("/"):
        fail(f"unsafe archive path {name!r}")
    parts = pathlib.PurePosixPath(name).parts
    if any(part in {"", ".", ".."} for part in parts):
        fail(f"unsafe archive path {name!r}")
    return parts


def read_archive(path):
    if zipfile.is_zipfile(path):
        with zipfile.ZipFile(path) as archive:
            for info in archive.infolist():
                name = info.filename.rstrip("/")
                if not name:
                    continue
                parts = safe_member(name)
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


def extract(path, destination, kind):
    seen = set()
    files = set()
    release_roots = set()
    for name, directory, contents in read_archive(path):
        if name in seen:
            fail(f"archive repeats path {name!r}")
        seen.add(name)
        parts = safe_member(name)
        if kind == "release":
            if len(parts) < 3 or parts[0] != "releases" or not RELEASE.fullmatch(parts[1]):
                fail(f"unexpected archive root {name!r} for immutable release")
            release_roots.add(parts[1])
        elif name not in CHANNEL_FILES and not directory:
            fail(f"unexpected archive root {name!r} for channel")
        elif parts[0] not in {"campaign", "releases"}:
            fail(f"unexpected archive root {name!r} for channel")
        target = destination.joinpath(*parts)
        if directory:
            target.mkdir(parents=True, exist_ok=True)
            continue
        if target.exists():
            fail(f"archive would overwrite {name!r}")
        target.parent.mkdir(parents=True, exist_ok=True)
        with target.open("xb") as output:
            output.write(contents)
        files.add(name)
    if kind == "release":
        if len(release_roots) != 1:
            fail("immutable release archive must contain exactly one release root")
        release = next(iter(release_roots))
        if f"releases/{release}/release.json" not in files or f"releases/{release}/checksums.txt" not in files:
            fail("immutable release archive misses release metadata")
    elif files != CHANNEL_FILES:
        fail("channel archive must contain campaign/v1.js and all three channel documents")


def download(url, destination, token):
    command = [
        "curl",
        "--proto",
        "=https",
        "--proto-redir",
        "=https",
        "--location",
        "--fail",
        "--silent",
        "--show-error",
        "--output",
        str(destination),
        url,
    ]
    if token:
        command[1:1] = ["--header", f"Authorization: Bearer {token}"]
    subprocess.run(command, check=True)


def verified_download(url, identifier, expected, temporary, token):
    if not IDENTIFIER.fullmatch(identifier):
        fail("asset identifier is invalid")
    if not SHA256.fullmatch(expected):
        fail("asset SHA-256 is invalid")
    if not url.startswith("https://github.com/") and not url.startswith("https://api.github.com/"):
        fail("asset URL must be an HTTPS GitHub URL")
    archive = temporary / identifier
    download(url, archive, token)
    with archive.open("rb") as contents:
        actual = hashlib.file_digest(contents, "sha256").hexdigest()
    if actual != expected:
        fail(f"SHA-256 mismatch for asset {identifier!r}")
    return archive


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--release-url", required=True)
    parser.add_argument("--release-id", required=True)
    parser.add_argument("--release-sha256", required=True)
    parser.add_argument("--channel-url", required=True)
    parser.add_argument("--channel-id", required=True)
    parser.add_argument("--channel-sha256", required=True)
    parser.add_argument("--output", required=True, type=pathlib.Path)
    args = parser.parse_args()
    if args.output.exists():
        fail(f"asset bundle output already exists: {args.output}")
    args.output.parent.mkdir(parents=True, exist_ok=True)
    token = os.environ.get("ASSETS_GITHUB_TOKEN", "")
    with tempfile.TemporaryDirectory(prefix="ahairu-assets-", dir=args.output.parent) as temporary_name:
        temporary = pathlib.Path(temporary_name)
        release = verified_download(args.release_url, args.release_id, args.release_sha256, temporary, token)
        channel = verified_download(args.channel_url, args.channel_id, args.channel_sha256, temporary, token)
        staged = temporary / "bundle"
        staged.mkdir()
        extract(release, staged, "release")
        extract(channel, staged, "channel")
        os.replace(staged, args.output)


if __name__ == "__main__":
    try:
        main()
    except (OSError, subprocess.CalledProcessError, ValueError, tarfile.TarError, zipfile.BadZipFile) as error:
        print(f"prepare asset bundle: {error}", file=sys.stderr)
        sys.exit(1)
