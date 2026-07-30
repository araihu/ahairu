import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";

import { assertCapture, buildProbeSpecs } from "./seasonal_assets_canary.mjs";

function write(root, name, value = name) {
  const target = path.join(root, name);
  mkdirSync(path.dirname(target), { recursive: true });
  writeFileSync(target, value);
}

test("canary derives documented routes and exact GET bodies from public", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "ahairu-canary-"));
  try {
    for (const name of ["latest", "default", "current"]) write(root, `assets/releases/${name}.json`);
    write(root, "assets/campaign/v1.js");
    write(root, "assets/releases/v0.1.1/release.json");
    write(root, "assets/releases/v0.1.1/catalog.json");
    for (const name of ["brand/index.html", "license/index.html", "pt-br/index.html", "es/brand/index.html"]) write(root, name);

    const { specs, releases } = buildProbeSpecs(root);
    assert.deepEqual(releases, ["v0.1.1"]);
    assert.equal(specs.find(({ route }) => route === "/assets/releases/current").get.cache, "public, max-age=60, must-revalidate");
    assert.equal(specs.find(({ route }) => route === "/assets/releases/v0.1.1/catalog.json").get.cache, "public, max-age=31536000, immutable");
    assert.equal(specs.find(({ route }) => route === "/assets/missing.svg").head.bytes, 0);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("canary rejects mismatched capture fields", () => {
  assert.throws(() => assertCapture(
    { method: "GET", route: "/asset", status: 200, type: null, cache: null, cors: null, bytes: 2, sha256: "bad" },
    { status: 200, type: null, cache: null, cors: null, bytes: 1, sha256: "good" },
  ), /bytes=2, want 1/);
});
