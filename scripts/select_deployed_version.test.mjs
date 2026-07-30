import assert from "node:assert/strict";
import test from "node:test";

import { captureUploadedVersion, selectJustDeployedVersion } from "./select_deployed_version.mjs";

const uploadedA = "11111111-1111-4111-8111-111111111111";
const uploadedB = "22222222-2222-4222-8222-222222222222";

test("captures exactly Wrangler's version-upload ID from machine-readable output", () => {
  assert.equal(captureUploadedVersion([
    JSON.stringify({ type: "version-upload", version: 1, version_id: uploadedA }),
    JSON.stringify({ type: "version-deploy", version: 1, version_id: uploadedB }),
  ].join("\n")), uploadedA);
});

test("rejects machine-readable output with multiple uploaded versions", () => {
  assert.throws(() => captureUploadedVersion([
    JSON.stringify({ type: "version-upload", version: 1, version_id: uploadedA }),
    JSON.stringify({ type: "version-upload", version: 1, version_id: uploadedB }),
  ].join("\n")), /exactly one version-upload/);
});

test("accepts only captured upload as the sole active version", () => {
  assert.equal(selectJustDeployedVersion({
    uploadedVersion: uploadedA,
    deployment: { versions: [{ version_id: uploadedA, percentage: 100 }] },
  }), uploadedA);
});

test("rejects a split deployment", () => {
  assert.throws(() => selectJustDeployedVersion({
    uploadedVersion: uploadedA,
    deployment: { versions: [{ version_id: uploadedA, percentage: 50 }, { version_id: uploadedB, percentage: 50 }] },
  }), /exactly one version at 100%/);
});

test("rejects concurrent version B even when B is solely active", () => {
  assert.throws(() => selectJustDeployedVersion({
    uploadedVersion: uploadedA,
    deployment: { versions: [{ version_id: uploadedB, percentage: 100 }] },
  }), /does not serve captured uploaded version/);
});
