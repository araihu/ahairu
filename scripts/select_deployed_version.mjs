#!/usr/bin/env node
import fs from "node:fs";

function fail(message) {
  throw new Error(`deployed Worker version: ${message}`);
}

function versionIDs(versions) {
  if (!Array.isArray(versions)) fail("versions list is not an array");
  return new Set(versions.map((version) => version?.id).filter((id) => typeof id === "string" && id.length > 0));
}

export function selectJustDeployedVersion({ before, deployment, after }) {
  const beforeIDs = versionIDs(before);
  const afterIDs = versionIDs(after);
  if (!deployment || !Array.isArray(deployment.versions)) {
    fail("latest deployment has no versions");
  }
  const active = deployment.versions.filter(
    (version) => version && version.percentage === 100 && typeof version.version_id === "string" && version.version_id.length > 0,
  );
  if (active.length !== 1 || deployment.versions.length !== 1) {
    fail("latest deployment must serve exactly one version at 100%");
  }
  const id = active[0].version_id;
  if (beforeIDs.has(id)) fail(`${id} was already present before deploy`);
  if (!afterIDs.has(id)) fail(`${id} is absent from versions listed after deploy`);
  return id;
}

function readJSON(path) {
  return JSON.parse(fs.readFileSync(path, "utf8"));
}

if (import.meta.main) {
  const paths = Object.fromEntries(process.argv.slice(2).reduce((pairs, value, index, values) => {
    if (value.startsWith("--") && values[index + 1]) pairs.push([value.slice(2), values[index + 1]]);
    return pairs;
  }, []));
  for (const name of ["before", "deployment", "after"]) {
    if (!paths[name]) fail(`missing --${name}`);
  }
  const id = selectJustDeployedVersion({
    before: readJSON(paths.before),
    deployment: readJSON(paths.deployment),
    after: readJSON(paths.after),
  });
  fs.appendFileSync(process.env.GITHUB_STEP_SUMMARY, `Deployed Worker version: ${id}\n`);
  process.stdout.write(`${id}\n`);
}
