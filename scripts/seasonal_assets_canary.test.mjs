import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";

import {
  assertCapture,
  buildProbeSpecs,
  canonicalProxyURL,
  inspectEnabledBundle,
} from "./seasonal_assets_canary.mjs";

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

function enabledBundleFixture(root) {
  const release = "v0.1.2";
  const prefix = `https://araihu.com/assets/releases/${release}/`;
  const campaign = {
    id: "halloween-2026",
    enabled: true,
    startsOn: "2026-10-30",
    endsOn: "2026-10-31",
    theme: "araihu-signal-night",
    toggle: {
      enabledIcon: { asset: "ui-hi-16-solid-sparkles", mode: "sprite" },
      disabledIcon: { asset: "ui-hi-16-solid-moon", mode: "sprite" },
    },
    brand: {
      logo: "araihu-logo-tinted-transparent-optical",
      icon: "araihu-icon-tinted-transparent-optical",
    },
  };
  const assets = [
    ["themes/araihu-signal-night.css", 11],
    ["brand/araihu/logo/tinted-transparent-optical.svg", 12],
    ["icons/brand/araihu-icon-tinted-transparent-optical.svg", 13],
    ["icons/ui/sprite.svg", 14],
  ];
  write(root, "releases/v0.1.1/release.json", JSON.stringify({ release: "v0.1.1", files: [] }));
  write(root, "releases/v0.1.1/campaigns.json", JSON.stringify({ schemaVersion: 1, campaigns: [] }));
  write(root, `releases/${release}/release.json`, JSON.stringify({
    release,
    files: assets.map(([assetPath, size], index) => ({
      path: assetPath,
      size,
      sha256: String(index + 1).repeat(64),
    })),
  }));
  write(root, `releases/${release}/campaigns.json`, JSON.stringify({ schemaVersion: 1, campaigns: [campaign] }));
  const current = {
    schemaVersion: 1,
    runtimeVersion: 1,
    release,
    source: "campaign",
    theme: { id: campaign.theme, cssUrl: `${prefix}${assets[0][0]}` },
    campaign: {
      id: campaign.id,
      toggle: {
        enabledIcon: {
          id: campaign.toggle.enabledIcon.asset,
          mode: "sprite",
          url: `${prefix}${assets[3][0]}`,
          spriteSymbol: "hi-16-solid-sparkles",
        },
        disabledIcon: {
          id: campaign.toggle.disabledIcon.asset,
          mode: "sprite",
          url: `${prefix}${assets[3][0]}`,
          spriteSymbol: "hi-16-solid-moon",
        },
      },
      brand: {
        logo: { id: campaign.brand.logo, url: `${prefix}${assets[1][0]}` },
        icon: { id: campaign.brand.icon, url: `${prefix}${assets[2][0]}` },
      },
    },
    digest: "a".repeat(64),
  };
  write(root, "releases/current.json", JSON.stringify(current));
  const expired = {
    schemaVersion: 1,
    runtimeVersion: 1,
    release,
    source: "default",
    theme: { id: "araihu", cssUrl: `${prefix}themes/araihu.css` },
    digest: "b".repeat(64),
  };
  write(root, "expired.json", JSON.stringify(expired));
  return { current, expiredPath: path.join(root, "expired.json") };
}

test("enabled browser gate records exact campaign resolution and direct assets", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "ahairu-enabled-canary-"));
  try {
    const fixture = enabledBundleFixture(root);
    const contract = inspectEnabledBundle(root, "2026-10-31", fixture.expiredPath, "2026-11-01");
    assert.deepEqual(contract.releases, ["v0.1.1", "v0.1.2"]);
    assert.equal(contract.current.campaign.id, "halloween-2026");
    assert.equal(contract.assets.length, 5);
    assert.equal(contract.assets.find(({ kind }) => kind === "toggle-enabled").id, "ui-hi-16-solid-sparkles");
    assert.equal(contract.expired.source, "default");
    assert.equal(contract.expired.resolutionDate, "2026-11-01");
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("enabled browser gate rejects incomplete release history", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "ahairu-enabled-canary-"));
  try {
    enabledBundleFixture(root);
    rmSync(path.join(root, "releases", "v0.1.1"), { recursive: true, force: true });
    assert.throws(
      () => inspectEnabledBundle(root, "2026-10-31"),
      /at least two retained immutable releases/,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("canonical browser proxy maps paths locally and rejects alternate origins", () => {
  assert.equal(
    canonicalProxyURL("https://araihu.com/assets/releases/current", "http://127.0.0.1:8788"),
    "http://127.0.0.1:8788/assets/releases/current",
  );
  assert.equal(
    canonicalProxyURL(
      "https://araihu.com/assets/releases/current",
      "http://127.0.0.1:8788",
      "/assets/releases/canary-expired.json",
    ),
    "http://127.0.0.1:8788/assets/releases/canary-expired.json",
  );
  assert.throws(
    () => canonicalProxyURL("https://example.com/asset", "http://127.0.0.1:8788"),
    /escaped canonical origin/,
  );
});
