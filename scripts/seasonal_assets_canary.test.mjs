import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import test from "node:test";

import {
  assertCapture,
  buildProbeSpecs,
  canonicalProxyURL,
  computeChannelDigest,
  createBrowserRequestGate,
  inspectEnabledBundle,
  snapshotSSRBaseline,
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

test("browser request gates reject missing runtime and current fetches", async () => {
  await assert.rejects(
    createBrowserRequestGate(10).waitForRuntime(),
    /campaign runtime request was not observed within 10ms/,
  );
  await assert.rejects(
    createBrowserRequestGate(10).waitForCurrent(),
    /current channel request 1 was not observed within 10ms/,
  );
  const failed = createBrowserRequestGate(1000);
  failed.fail(new Error("page crashed"));
  await assert.rejects(failed.waitForCurrent(), /page crashed/);
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
    ["themes/araihu.css", 15],
    ["icons/ui/heroicons/16-solid-sparkles.svg", 16],
    ["icons/ui/heroicons/16-solid-moon.svg", 17],
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
  write(root, `releases/${release}/themes.json`, JSON.stringify({
    schemaVersion: 1,
    release,
    themes: [
      { id: "araihu-signal-night", cssPath: assets[0][0], sha256: "1".repeat(64) },
      { id: "araihu", cssPath: assets[4][0], sha256: "5".repeat(64) },
    ],
  }));
  write(root, `releases/${release}/catalog.json`, JSON.stringify({
    schemaVersion: 1,
    release,
    assets: [
      { canonicalName: campaign.brand.logo, namespace: "brand", artwork: "logo", path: assets[1][0], sha256: "2".repeat(64) },
      { canonicalName: campaign.brand.icon, namespace: "brand", artwork: "icon", path: assets[2][0], sha256: "3".repeat(64) },
      { canonicalName: campaign.toggle.enabledIcon.asset, namespace: "ui", artwork: "icon", path: assets[5][0], spriteSymbol: "hi-16-solid-sparkles", sha256: "6".repeat(64) },
      { canonicalName: campaign.toggle.disabledIcon.asset, namespace: "ui", artwork: "icon", path: assets[6][0], spriteSymbol: "hi-16-solid-moon", sha256: "7".repeat(64) },
    ],
  }));
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
    digest: "",
  };
  current.digest = computeChannelDigest(current);
  write(root, "releases/current.json", JSON.stringify(current));
  const expired = {
    schemaVersion: 1,
    runtimeVersion: 1,
    release,
    source: "default",
    theme: { id: "araihu", cssUrl: `${prefix}themes/araihu.css` },
    digest: "",
  };
  expired.digest = computeChannelDigest(expired);
  write(root, "expired.json", JSON.stringify(expired));
  return { current, expiredPath: path.join(root, "expired.json") };
}

test("enabled browser gate checks campaign identity and all direct assets", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "ahairu-enabled-canary-"));
  try {
    const fixture = enabledBundleFixture(root);
    const contract = inspectEnabledBundle(root, "2026-10-31", fixture.expiredPath, "2026-11-01");
    assert.deepEqual(contract.releases, ["v0.1.1", "v0.1.2"]);
    assert.equal(contract.current.campaign.id, "halloween-2026");
    assert.equal(contract.assets.length, 5);
    assert.equal(contract.assets.find(({ kind }) => kind === "toggle-enabled").id, "ui-hi-16-solid-sparkles");
    assert.equal(contract.expired.source, "default");
    assert.equal(contract.campaignCheckDate, "2026-10-31");
    assert.equal(contract.expired.campaignCheckDate, "2026-11-01");
    assert.equal(contract.expired.assets.length, 1);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("expired channel rejects tampered digest and missing theme inventory", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "ahairu-enabled-canary-"));
  try {
    const fixture = enabledBundleFixture(root);
    const expired = JSON.parse(readFileSync(fixture.expiredPath, "utf8"));
    expired.digest = "f".repeat(64);
    writeFileSync(fixture.expiredPath, JSON.stringify(expired));
    assert.throws(
      () => inspectEnabledBundle(root, "2026-10-31", fixture.expiredPath, "2026-11-01"),
      /expired channel digest=.*recomputed=/,
    );

    expired.digest = computeChannelDigest(expired);
    writeFileSync(fixture.expiredPath, JSON.stringify(expired));
    const releasePath = path.join(root, "releases", "v0.1.2", "release.json");
    const release = JSON.parse(readFileSync(releasePath, "utf8"));
    release.files = release.files.filter(({ path: assetPath }) => assetPath !== "themes/araihu.css");
    writeFileSync(releasePath, JSON.stringify(release));
    assert.throws(
      () => inspectEnabledBundle(root, "2026-10-31", fixture.expiredPath, "2026-11-01"),
      /expired channel theme themes\/araihu\.css is absent from v0\.1\.2 inventory/,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("toggle binding rejects sprite URL declared as asset mode", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "ahairu-enabled-canary-"));
  try {
    const fixture = enabledBundleFixture(root);
    const current = structuredClone(fixture.current);
    current.campaign.toggle.enabledIcon.mode = "asset";
    delete current.campaign.toggle.enabledIcon.spriteSymbol;
    current.digest = computeChannelDigest(current);
    writeFileSync(path.join(root, "releases", "current.json"), JSON.stringify(current));
    assert.throws(
      () => inspectEnabledBundle(root, "2026-10-31"),
      /enabledMode="asset", want "sprite"/,
    );

    const campaignsPath = path.join(root, "releases", "v0.1.2", "campaigns.json");
    const campaigns = JSON.parse(readFileSync(campaignsPath, "utf8"));
    campaigns.campaigns[0].toggle.enabledIcon.mode = "asset";
    writeFileSync(campaignsPath, JSON.stringify(campaigns));
    assert.throws(
      () => inspectEnabledBundle(root, "2026-10-31"),
      /toggle-enabled ui-hi-16-solid-sparkles asset mode requires its discrete UI asset path/,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("toggle binding rejects brand namespace drift", () => {
  const root = mkdtempSync(path.join(os.tmpdir(), "ahairu-enabled-canary-"));
  try {
    enabledBundleFixture(root);
    const catalogPath = path.join(root, "releases", "v0.1.2", "catalog.json");
    const catalog = JSON.parse(readFileSync(catalogPath, "utf8"));
    catalog.assets.find(({ canonicalName }) => canonicalName === "ui-hi-16-solid-sparkles").namespace = "brand";
    writeFileSync(catalogPath, JSON.stringify(catalog));
    assert.throws(
      () => inspectEnabledBundle(root, "2026-10-31"),
      /toggle-enabled ui-hi-16-solid-sparkles requires namespace ui/,
    );
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("SSR baseline snapshot cannot be poisoned by later runtime brand mutation", () => {
  const expected = {
    theme: "araihu",
    source: "default",
    campaign: null,
    logo: "https://araihu.com/assets/releases/v0.1.1/brand/araihu/logo/tinted-transparent-optical.svg",
    logoCrossOrigin: null,
    icon: "https://araihu.com/assets/releases/v0.1.1/platform/web/araihu/favicon.svg",
    iconCrossOrigin: null,
    toggleHidden: true,
    togglePressed: "false",
  };
  const state = {
    ...expected,
    href: "https://araihu.com/brand/",
    events: [],
  };
  const baseline = snapshotSSRBaseline(state, expected);
  const poisoned = {
    ...state,
    source: "campaign",
    campaign: "halloween-2026",
    logo: "https://araihu.com/assets/releases/v0.1.2/brand/araihu/logo/tinted-transparent-optical.svg",
  };
  assert.equal(baseline.logo, expected.logo);
  assert.throws(() => snapshotSSRBaseline(poisoned, expected), /SSR baseline source="campaign"/);
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
