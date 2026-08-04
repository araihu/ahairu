#!/usr/bin/env node
import fs from "node:fs";

function fail(message) {
  throw new Error(`deployed Worker version: ${message}`);
}

function versionID(value, source) {
  if (typeof value !== "string" || value.length === 0) {
    fail(`${source} has no version ID`);
  }
  return value;
}

export function captureUploadedVersion(output) {
  if (typeof output !== "string") fail("Wrangler upload output is not text");
  const entries = output.split("\n").filter((line) => line.trim().length > 0).map((line, index) => {
    try {
      return JSON.parse(line);
    } catch {
      fail(`Wrangler upload output line ${index + 1} is not JSON`);
    }
  });
  const uploads = entries.filter((entry) => entry?.type === "version-upload" && entry.version === 1);
  if (uploads.length !== 1) fail("Wrangler upload output must contain exactly one version-upload result");
  return versionID(uploads[0].version_id, "version-upload result");
}

export function selectJustDeployedVersion({ uploadedVersion, deployment }) {
  const id = versionID(uploadedVersion, "captured uploaded version");
  if (!deployment || !Array.isArray(deployment.versions)) {
    fail("latest deployment has no versions");
  }
  const active = deployment.versions.filter(
    (version) => version && version.percentage === 100 && typeof version.version_id === "string" && version.version_id.length > 0,
  );
  if (active.length !== 1) {
    fail("latest deployment must serve exactly one version at 100%");
  }
  if (active[0].version_id !== id) {
    fail(`latest deployment does not serve captured uploaded version ${id} at 100%`);
  }
  return id;
}

function readJSON(path) {
  return JSON.parse(fs.readFileSync(path, "utf8"));
}

function pathsFromArguments(args) {
  return Object.fromEntries(args.reduce((pairs, value, index, values) => {
    if (value.startsWith("--") && values[index + 1]) pairs.push([value.slice(2), values[index + 1]]);
    return pairs;
  }, []));
}

if (import.meta.main) {
  const paths = pathsFromArguments(process.argv.slice(2));
  if (paths.upload) {
    process.stdout.write(`${captureUploadedVersion(fs.readFileSync(paths.upload, "utf8"))}\n`);
  } else {
    for (const name of ["uploaded-version", "deployment"]) {
      if (!paths[name]) fail(`missing --${name}`);
    }
    const id = selectJustDeployedVersion({
      uploadedVersion: paths["uploaded-version"],
      deployment: readJSON(paths.deployment),
    });
    fs.appendFileSync(process.env.GITHUB_STEP_SUMMARY, `Deployed Worker version: ${id}\n`);
    process.stdout.write(`${id}\n`);
  }
}
