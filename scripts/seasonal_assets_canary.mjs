#!/usr/bin/env node
// Local evidence probe. It deliberately starts one Wrangler session only.
import { createHash } from "node:crypto";
import { copyFileSync, existsSync, readdirSync, readFileSync, rmSync, statSync } from "node:fs";
import { once } from "node:events";
import { spawn, spawnSync } from "node:child_process";
import path from "node:path";
import { chromium } from "playwright";
import { executablePath as puppeteerExecutablePath } from "puppeteer";

const host = "127.0.0.1";
const canonicalOrigin = "https://araihu.com";
const currentChannelPath = "/assets/releases/current";
const expiredChannelPath = "/assets/releases/canary-expired.json";
const emptySHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855";
const releaseName = /^v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$/;
const calendarDate = /^\d{4}-\d{2}-\d{2}$/;
const sha256Hex = /^[0-9a-f]{64}$/;

function fail(message) {
  throw new Error(message);
}

function sha256(bytes) {
  return createHash("sha256").update(bytes).digest("hex");
}

function readJSON(filename, label) {
  let value;
  try {
    value = JSON.parse(readFileSync(filename, "utf8"));
  } catch (error) {
    fail(`${label} is not valid JSON: ${error.message}`);
  }
  return value;
}

function validateDate(value, name) {
  if (!calendarDate.test(value || "")) fail(`${name} must use YYYY-MM-DD`);
  const parsed = new Date(`${value}T00:00:00.000Z`);
  if (Number.isNaN(parsed.valueOf()) || parsed.toISOString().slice(0, 10) !== value) {
    fail(`${name} must use YYYY-MM-DD`);
  }
  return value;
}

function discoverReleases(root) {
  const releasesRoot = path.join(root, "releases");
  if (!existsSync(releasesRoot) || !statSync(releasesRoot).isDirectory()) return [];
  return readdirSync(releasesRoot)
    .filter((entry) => releaseName.test(entry) && statSync(path.join(releasesRoot, entry)).isDirectory())
    .sort();
}

function validateChannelIdentity(channel, label) {
  if (!channel || channel.schemaVersion !== 1 || channel.runtimeVersion !== 1 ||
      !releaseName.test(channel.release || "") ||
      (channel.source !== "default" && channel.source !== "campaign") ||
      !sha256Hex.test(channel.digest || "") ||
      !channel.theme || typeof channel.theme.id !== "string" || typeof channel.theme.cssUrl !== "string") {
    fail(`${label} has invalid channel identity`);
  }
  if ((channel.source === "campaign") !== Boolean(channel.campaign)) {
    fail(`${label} has inconsistent campaign metadata`);
  }
}

function validateCanonicalReleaseURL(raw, release, label) {
  let url;
  try {
    url = new URL(raw);
  } catch (_) {
    fail(`${label} is not an absolute URL`);
  }
  const prefix = `/assets/releases/${release}/`;
  if (url.origin !== canonicalOrigin || url.username || url.password || url.search || url.hash ||
      !url.pathname.startsWith(prefix)) {
    fail(`${label} must target ${canonicalOrigin}${prefix}`);
  }
  return url;
}

function activeCampaign(manifest, resolutionDate) {
  return manifest.campaigns.find((campaign) =>
    campaign.enabled && campaign.startsOn <= resolutionDate && resolutionDate <= campaign.endsOn);
}

function validateCampaignResolution(root, channel, resolutionDate, label) {
  const releaseRoot = path.join(root, "releases", channel.release);
  if (!existsSync(releaseRoot) || !statSync(releaseRoot).isDirectory()) {
    fail(`${label} points outside retained release ${channel.release}`);
  }
  const manifest = readJSON(path.join(releaseRoot, "campaigns.json"), `${label} campaigns manifest`);
  if (!manifest || manifest.schemaVersion !== 1 || !Array.isArray(manifest.campaigns)) {
    fail(`${label} campaigns manifest has invalid schema`);
  }
  const active = activeCampaign(manifest, resolutionDate);
  if (channel.source === "default") {
    if (active) fail(`${label} source=default but ${active.id} is active on ${resolutionDate}`);
    return;
  }
  if (!active || active.id !== channel.campaign.id) {
    fail(`${label} campaign ${channel.campaign.id} is not active on ${resolutionDate}`);
  }
  const expected = {
    theme: active.theme,
    enabledIcon: active.toggle?.enabledIcon?.asset,
    disabledIcon: active.toggle?.disabledIcon?.asset,
    logo: active.brand?.logo,
    icon: active.brand?.icon,
  };
  const actual = {
    theme: channel.theme.id,
    enabledIcon: channel.campaign.toggle?.enabledIcon?.id,
    disabledIcon: channel.campaign.toggle?.disabledIcon?.id,
    logo: channel.campaign.brand?.logo?.id,
    icon: channel.campaign.brand?.icon?.id,
  };
  for (const key of Object.keys(expected)) {
    if (expected[key] !== actual[key]) {
      fail(`${label} ${key}=${JSON.stringify(actual[key])}, want ${JSON.stringify(expected[key])}`);
    }
  }
  if (!actual.logo.includes("tinted") || !actual.icon.includes("tinted")) {
    fail(`${label} must exercise tinted campaign brand assets`);
  }
  if (!actual.enabledIcon.includes("sparkles")) {
    fail(`${label} must exercise the sparkles enabled toggle`);
  }
}

function resolvedAssets(channel) {
  const entries = [
    { kind: "theme", id: channel.theme.id, url: channel.theme.cssUrl },
  ];
  if (channel.campaign) {
    entries.push(
      { kind: "brand-logo", ...channel.campaign.brand.logo },
      { kind: "brand-icon", ...channel.campaign.brand.icon },
      { kind: "toggle-enabled", ...channel.campaign.toggle.enabledIcon },
      { kind: "toggle-disabled", ...channel.campaign.toggle.disabledIcon },
    );
  }
  return entries;
}

function inventoryFor(root, release) {
  const document = readJSON(path.join(root, "releases", release, "release.json"), `${release} release metadata`);
  if (!document || document.release !== release || !Array.isArray(document.files)) {
    fail(`${release} release metadata has invalid inventory`);
  }
  return new Map(document.files.map((entry) => [entry.path, entry]));
}

/**
 * Inspect the exact accepted input before building. Full checksum/schema
 * verification remains authoritative in the Go assembler invoked by runBuild.
 */
export function inspectEnabledBundle(assetBundle, resolutionDate, expiredInput, expiredResolutionDate) {
  const resolvedOn = validateDate(resolutionDate, "CANARY_RESOLUTION_DATE");
  const releases = discoverReleases(assetBundle);
  if (releases.length < 2) {
    fail("enabled browser canary needs at least two retained immutable releases in ASSET_BUNDLE");
  }
  const current = readJSON(path.join(assetBundle, "releases", "current.json"), "current channel");
  validateChannelIdentity(current, "current channel");
  if (current.source !== "campaign") fail("enabled browser canary needs current channel source=campaign");
  validateCampaignResolution(assetBundle, current, resolvedOn, "current channel");

  const assets = resolvedAssets(current);
  const inventory = inventoryFor(assetBundle, current.release);
  for (const asset of assets) {
    const url = validateCanonicalReleaseURL(asset.url, current.release, `${asset.kind} URL`);
    const relative = url.pathname.slice(`/assets/releases/${current.release}/`.length);
    const entry = inventory.get(relative);
    if (!entry || !sha256Hex.test(entry.sha256 || "") || !Number.isInteger(entry.size)) {
      fail(`${asset.kind} ${relative} is absent from ${current.release} inventory`);
    }
    asset.pathname = url.pathname;
    asset.sha256 = entry.sha256;
    asset.bytes = entry.size;
  }

  let expired = null;
  if (expiredInput || expiredResolutionDate) {
    if (!expiredInput || !path.isAbsolute(expiredInput) || !existsSync(expiredInput) || !statSync(expiredInput).isFile()) {
      fail("CANARY_EXPIRED_CHANNEL must name an existing absolute channel JSON file");
    }
    const expiredOn = validateDate(expiredResolutionDate, "CANARY_EXPIRED_RESOLUTION_DATE");
    expired = readJSON(expiredInput, "expired channel");
    validateChannelIdentity(expired, "expired channel");
    if (expired.source !== "default") fail("expired channel must resolve source=default");
    validateCanonicalReleaseURL(expired.theme.cssUrl, expired.release, "expired theme URL");
    validateCampaignResolution(assetBundle, expired, expiredOn, "expired channel");
    expired.resolutionDate = expiredOn;
    expired.input = expiredInput;
  }

  return { resolutionDate: resolvedOn, releases, current, assets, expired };
}

export function canonicalProxyURL(rawURL, baseURL, pathnameOverride = "") {
  const remote = new URL(rawURL);
  if (remote.origin !== canonicalOrigin) fail(`browser request escaped canonical origin: ${remote.origin}`);
  const local = new URL(baseURL);
  local.pathname = pathnameOverride || remote.pathname;
  local.search = remote.search;
  return local.href;
}

function readPublic(publicRoot, relativePath) {
  const filename = path.join(publicRoot, relativePath);
  if (!existsSync(filename)) fail(`missing built public file ${relativePath}`);
  return readFileSync(filename);
}

function staticSpec(route, relativePath, headers) {
  const body = readPublic(this.publicRoot, relativePath);
  return {
    route,
    get: { status: 200, ...headers, bytes: body.byteLength, sha256: sha256(body) },
    head: { status: 200, ...headers, bytes: 0, sha256: emptySHA256 },
  };
}

/** Build expected GET and HEAD contracts from the exact generated public tree. */
export function buildProbeSpecs(publicRoot) {
  const makeStaticSpec = staticSpec.bind({ publicRoot });
  const cors = "*";
  const specs = [
    ...["latest", "default", "current"].map((channel) => makeStaticSpec(
      `/assets/releases/${channel}`,
      `assets/releases/${channel}.json`,
      { type: "application/json; charset=utf-8", cache: "public, max-age=60, must-revalidate", cors },
    )),
    makeStaticSpec(
      "/assets/campaign/v1.js",
      "assets/campaign/v1.js",
      { type: "text/javascript; charset=utf-8", cache: "public, max-age=0, must-revalidate", cors },
    ),
  ];

  const releasesRoot = path.join(publicRoot, "assets", "releases");
  if (!existsSync(releasesRoot)) fail("missing built immutable releases");
  const releases = readdirSync(releasesRoot)
    .filter((entry) => releaseName.test(entry) && statSync(path.join(releasesRoot, entry)).isDirectory())
    .sort();
  if (releases.length === 0) fail("built public tree has no immutable release");
  for (const release of releases) {
    for (const document of ["release.json", "catalog.json"]) {
      specs.push(makeStaticSpec(
        `/assets/releases/${release}/${document}`,
        `assets/releases/${release}/${document}`,
        { type: "application/json", cache: "public, max-age=31536000, immutable", cors },
      ));
    }
  }

  for (const [route, source] of [
    ["/brand/", "brand/index.html"],
    ["/license/", "license/index.html"],
    ["/pt-br/", "pt-br/index.html"],
    ["/es/brand/", "es/brand/index.html"],
  ]) {
    specs.push(makeStaticSpec(route, source, {
      type: "text/html; charset=utf-8", cache: "public, max-age=0, must-revalidate", cors: null,
    }));
  }

  specs.push(
    {
      route: "/not-a-page",
      get: { status: 404, type: "text/plain;charset=UTF-8", cache: null, cors: null, bytes: 9, sha256: sha256("Not found") },
      head: { status: 404, type: null, cache: null, cors: null, bytes: 0, sha256: emptySHA256 },
    },
    {
      route: "/assets/missing.svg",
      get: { status: 404, type: null, cache: null, cors: null, bytes: 0, sha256: emptySHA256 },
      head: { status: 404, type: null, cache: null, cors: null, bytes: 0, sha256: emptySHA256 },
    },
  );
  return { specs, releases };
}

function captureHeaders(response) {
  return {
    status: response.status,
    type: response.headers.get("content-type"),
    cache: response.headers.get("cache-control"),
    cors: response.headers.get("access-control-allow-origin"),
  };
}

export async function probe(baseURL, method, spec) {
  const response = await fetch(`${baseURL}${spec.route}`, { method, redirect: "manual" });
  const body = Buffer.from(await response.arrayBuffer());
  return {
    method,
    route: spec.route,
    ...captureHeaders(response),
    bytes: body.byteLength,
    sha256: sha256(body),
  };
}

export function assertCapture(capture, expected) {
  for (const field of ["status", "type", "cache", "cors", "bytes", "sha256"]) {
    if (capture[field] !== expected[field]) {
      fail(`${capture.method} ${capture.route}: ${field}=${JSON.stringify(capture[field])}, want ${JSON.stringify(expected[field])}`);
    }
  }
}

async function probeResolvedAssets(baseURL, contract) {
  for (const asset of contract.assets) {
    const response = await fetch(`${baseURL}${asset.pathname}`, { redirect: "manual" });
    const body = Buffer.from(await response.arrayBuffer());
    const capture = {
      event: "resolved-asset-probe",
      resolutionDate: contract.resolutionDate,
      release: contract.current.release,
      source: contract.current.source,
      campaign: contract.current.campaign.id,
      digest: contract.current.digest,
      kind: asset.kind,
      id: asset.id,
      url: asset.url,
      status: response.status,
      type: response.headers.get("content-type"),
      cache: response.headers.get("cache-control"),
      cors: response.headers.get("access-control-allow-origin"),
      bytes: body.byteLength,
      sha256: sha256(body),
    };
    if (capture.status !== 200 || capture.bytes !== asset.bytes || capture.sha256 !== asset.sha256) {
      fail(`${asset.kind} direct probe does not match ${contract.current.release} inventory`);
    }
    event(capture);
  }
}

function eventInitScript(preference) {
  window.__araihuCanaryEvents = [];
  [
    "araihu:campaign:before-apply",
    "araihu:campaign:applied",
    "araihu:campaign:restored",
    "araihu:campaign:error",
  ].forEach((type) => {
    document.addEventListener(type, (entry) => {
      window.__araihuCanaryEvents.push({ type, detail: entry.detail });
    });
  });
  if (preference) {
    document.documentElement.setAttribute("data-theme", preference.theme);
    document.documentElement.setAttribute("data-theme-source", "preference");
  }
}

async function captureBrowserState(page) {
  return page.evaluate(() => {
    const root = document.documentElement;
    const logo = document.querySelector('[data-asset-brand="logo"]');
    const icon = document.querySelector('[data-asset-brand="icon"]');
    const toggle = document.querySelector("[data-campaign-toggle]");
    const toggleContainer = document.querySelector("[data-campaign-toggle-icon]");
    const toggleChild = toggleContainer?.firstElementChild || null;
    const themeStyles = Array.from(document.querySelectorAll('link[rel="stylesheet"]')).map((node) => ({
      href: node.href,
      media: node.media,
      crossOrigin: node.getAttribute("crossorigin"),
    }));
    return {
      href: window.location.href,
      theme: root.getAttribute("data-theme"),
      source: root.getAttribute("data-theme-source"),
      campaign: root.getAttribute("data-campaign"),
      logo: logo?.src || null,
      logoCrossOrigin: logo?.getAttribute("crossorigin") ?? null,
      icon: icon?.href || null,
      iconCrossOrigin: icon?.getAttribute("crossorigin") ?? null,
      toggleHidden: toggle?.hidden ?? null,
      togglePressed: toggle?.getAttribute("aria-pressed") ?? null,
      toggleChild: toggleChild ? {
        tag: toggleChild.localName,
        src: toggleChild.src || null,
        viewBox: toggleChild.getAttribute("viewBox"),
        children: toggleChild.children.length,
      } : null,
      themeStyles,
      storage: Object.fromEntries(Object.keys(localStorage).sort().map((key) => [key, localStorage.getItem(key)])),
      events: window.__araihuCanaryEvents || [],
      reducedMotion: window.matchMedia("(prefers-reduced-motion: reduce)").matches,
    };
  });
}

function assertCanonicalPage(state) {
  if (!state.href.startsWith(`${canonicalOrigin}/`)) {
    fail(`browser navigation lost canonical origin: ${state.href}`);
  }
}

function assertToggleIcon(state, icon) {
  if (!state.toggleChild) fail("campaign toggle has no rendered icon");
  if (icon.mode === "sprite") {
    if (state.toggleChild.tag !== "svg" || state.toggleChild.children < 1) {
      fail(`${icon.id} did not render as a bounded inline SVG`);
    }
  } else if (state.toggleChild.tag !== "img" || state.toggleChild.src !== icon.url) {
    fail(`${icon.id} did not render as the declared image`);
  }
}

function assertAppliedState(state, channel) {
  assertCanonicalPage(state);
  if (state.theme !== channel.theme.id || state.source !== "campaign" ||
      state.campaign !== channel.campaign.id || state.toggleHidden ||
      state.togglePressed !== "true" || state.logo !== channel.campaign.brand.logo.url ||
      state.icon !== channel.campaign.brand.icon.url ||
      state.logoCrossOrigin !== "anonymous" || state.iconCrossOrigin !== "anonymous") {
    fail("browser first apply does not match resolved campaign");
  }
  const style = state.themeStyles.find((entry) => entry.href === channel.theme.cssUrl);
  if (!style || style.media !== "all" || style.crossOrigin !== "anonymous") {
    fail("resolved campaign theme stylesheet is not active with anonymous CORS");
  }
  assertToggleIcon(state, channel.campaign.toggle.enabledIcon);
}

function assertPreferenceState(state, baseline) {
  assertCanonicalPage(state);
  if (state.theme !== "canary-preference" || state.source !== "preference" || state.campaign !== null ||
      !state.toggleHidden || state.logo !== baseline.logo || state.icon !== baseline.icon ||
      state.events.some((entry) => entry.type === "araihu:campaign:applied")) {
    fail("explicit preference did not retain baseline presentation");
  }
}

function assertOptOutState(state, channel, baseline) {
  const key = `araihu.assets.campaign.v1.optout.${channel.campaign.id}`;
  assertCanonicalPage(state);
  if (state.theme !== baseline.theme || state.source !== "campaign-opt-out" ||
      state.campaign !== channel.campaign.id || state.toggleHidden ||
      state.togglePressed !== "false" || state.logo !== baseline.logo ||
      state.icon !== baseline.icon || state.storage[key] !== "1") {
    fail("campaign opt-out did not restore and persist baseline presentation");
  }
  assertToggleIcon(state, channel.campaign.toggle.disabledIcon);
}

function assertExpiredState(state, expired, baseline) {
  assertCanonicalPage(state);
  if (state.theme !== baseline.theme || state.source !== baseline.source || state.campaign !== null ||
      !state.toggleHidden || state.logo !== baseline.logo || state.icon !== baseline.icon ||
      !state.events.some((entry) =>
        entry.type === "araihu:campaign:restored" && entry.detail?.code === "campaign-inactive")) {
    fail(`expired ${expired.release} channel did not restore baseline presentation`);
  }
}

function transcriptState(scenario, contract, state, extra = {}) {
  event({
    event: "browser-state",
    scenario,
    resolutionDate: contract.resolutionDate,
    release: contract.current.release,
    source: contract.current.source,
    campaign: contract.current.campaign.id,
    digest: contract.current.digest,
    theme: state.theme,
    themeSource: state.source,
    activeCampaign: state.campaign,
    toggleHidden: state.toggleHidden,
    togglePressed: state.togglePressed,
    logo: state.logo,
    icon: state.icon,
    reducedMotion: state.reducedMotion,
    events: state.events,
    ...extra,
  });
}

async function openProxiedContext(browser, baseURL, { preference = null, reducedMotion = "no-preference" } = {}) {
  const context = await browser.newContext({ reducedMotion, serviceWorkers: "block" });
  const unexpected = [];
  let useExpiredChannel = false;
  await context.addInitScript(eventInitScript, preference);
  await context.route(/^https?:\/\//, async (route) => {
    const request = route.request();
    const remote = new URL(request.url());
    if (remote.origin !== canonicalOrigin) {
      unexpected.push(request.url());
      await route.abort("blockedbyclient");
      return;
    }
    const override = useExpiredChannel && remote.pathname === currentChannelPath ? expiredChannelPath : "";
    const headers = { ...request.headers() };
    delete headers.host;
    const response = await route.fetch({
      url: canonicalProxyURL(request.url(), baseURL, override),
      headers,
      maxRedirects: 0,
    });
    await route.fulfill({ response });
  });
  return {
    context,
    useExpiredChannel() {
      useExpiredChannel = true;
    },
    assertLocalOnly() {
      if (unexpected.length) fail(`browser attempted non-canonical network requests: ${unexpected.join(", ")}`);
    },
  };
}

async function waitForRuntime(page) {
  await page.waitForFunction(() => window.AraiHuCampaign?.version === 1);
  await page.evaluate(() => window.AraiHuCampaign.refresh());
}

async function runBrowserCanary(baseURL, contract) {
  const browserPath = puppeteerExecutablePath();
  if (!browserPath || !existsSync(browserPath)) {
    fail("Puppeteer Chromium is unavailable; run npm ci before the browser canary");
  }
  const browser = await chromium.launch({ headless: true, executablePath: browserPath });
  try {
    const first = await openProxiedContext(browser, baseURL);
    try {
      const page = await first.context.newPage();
      await page.goto(`${canonicalOrigin}/brand/`, { waitUntil: "domcontentloaded" });
      await waitForRuntime(page);
      const state = await captureBrowserState(page);
      assertAppliedState(state, contract.current);
      transcriptState("first-apply", contract, state);
      first.assertLocalOnly();
    } finally {
      await first.context.close();
    }

    const preferred = await openProxiedContext(browser, baseURL, {
      preference: { theme: "canary-preference" },
    });
    let baseline;
    try {
      const page = await preferred.context.newPage();
      await page.goto(`${canonicalOrigin}/brand/`, { waitUntil: "domcontentloaded" });
      await waitForRuntime(page);
      const state = await captureBrowserState(page);
      baseline = { theme: "araihu", source: "default", logo: state.logo, icon: state.icon };
      assertPreferenceState(state, baseline);
      transcriptState("explicit-preference", contract, state);
      preferred.assertLocalOnly();
    } finally {
      await preferred.context.close();
    }

    const toggle = await openProxiedContext(browser, baseURL);
    try {
      const page = await toggle.context.newPage();
      await page.goto(`${canonicalOrigin}/brand/`, { waitUntil: "domcontentloaded" });
      await waitForRuntime(page);
      await page.locator("[data-campaign-toggle]").click();
      await page.waitForFunction(() => document.documentElement.dataset.themeSource === "campaign-opt-out");
      let state = await captureBrowserState(page);
      assertOptOutState(state, contract.current, baseline);
      transcriptState("opt-out", contract, state);

      await page.reload({ waitUntil: "domcontentloaded" });
      await waitForRuntime(page);
      state = await captureBrowserState(page);
      assertOptOutState(state, contract.current, baseline);
      transcriptState("opt-out-reload", contract, state);

      await page.locator("[data-campaign-toggle]").click();
      await page.waitForFunction(() => document.documentElement.dataset.themeSource === "campaign");
      state = await captureBrowserState(page);
      assertAppliedState(state, contract.current);
      const key = `araihu.assets.campaign.v1.optout.${contract.current.campaign.id}`;
      if (Object.prototype.hasOwnProperty.call(state.storage, key)) {
        fail("campaign re-enable retained opt-out storage");
      }
      transcriptState("re-enabled", contract, state);
      toggle.assertLocalOnly();
    } finally {
      await toggle.context.close();
    }

    const reduced = await openProxiedContext(browser, baseURL, { reducedMotion: "reduce" });
    try {
      const page = await reduced.context.newPage();
      await page.goto(`${canonicalOrigin}/brand/`, { waitUntil: "domcontentloaded" });
      await waitForRuntime(page);
      const state = await captureBrowserState(page);
      assertAppliedState(state, contract.current);
      const lifecycle = state.events.filter((entry) =>
        entry.type === "araihu:campaign:before-apply" || entry.type === "araihu:campaign:applied");
      if (!state.reducedMotion || lifecycle.length < 2 ||
          lifecycle.some((entry) => entry.detail?.reducedMotion !== true)) {
        fail("campaign lifecycle did not expose reduced-motion preference");
      }
      transcriptState("reduced-motion", contract, state);
      reduced.assertLocalOnly();
    } finally {
      await reduced.context.close();
    }

    if (contract.expired) {
      const expiry = await openProxiedContext(browser, baseURL);
      try {
        const page = await expiry.context.newPage();
        await page.goto(`${canonicalOrigin}/brand/`, { waitUntil: "domcontentloaded" });
        await waitForRuntime(page);
        expiry.useExpiredChannel();
        await page.evaluate(() => window.AraiHuCampaign.refresh());
        const state = await captureBrowserState(page);
        assertExpiredState(state, contract.expired, baseline);
        transcriptState("expiry-refresh", contract, state, {
          expiredResolutionDate: contract.expired.resolutionDate,
          expiredRelease: contract.expired.release,
          expiredSource: contract.expired.source,
          expiredDigest: contract.expired.digest,
        });
        expiry.assertLocalOnly();
      } finally {
        await expiry.context.close();
      }
    } else {
      event({
        event: "scenario-skip",
        scenario: "expiry-refresh",
        reason: "CANARY_EXPIRED_CHANNEL and CANARY_EXPIRED_RESOLUTION_DATE not supplied",
      });
    }
  } finally {
    await browser.close();
  }
}

function configuredPort(value) {
  const port = Number(value || 8788);
  if (!Number.isInteger(port) || port < 1 || port > 65535) fail("CANARY_PORT must be an integer from 1 through 65535");
  return port;
}

function configuredTimeout(value) {
  const timeout = Number(value || 30000);
  if (!Number.isInteger(timeout) || timeout < 1000 || timeout > 120000) fail("CANARY_TIMEOUT_MS must be an integer from 1000 through 120000");
  return timeout;
}

function event(record) {
  process.stdout.write(`${JSON.stringify(record)}\n`);
}

function runBuild(assetBundle) {
  const result = spawnSync("npm", ["run", "build"], {
    cwd: process.cwd(), env: { ...process.env, ASSET_BUNDLE: assetBundle }, stdio: ["ignore", "ignore", "inherit"],
  });
  if (result.error) fail(`start build: ${result.error.message}`);
  if (result.status !== 0) fail(`build exited ${result.status ?? "without status"}`);
}

function startWrangler(port) {
  const executable = path.join(process.cwd(), "node_modules", ".bin", "wrangler");
  if (!existsSync(executable)) fail("missing node_modules/.bin/wrangler; run npm ci first");
  const logs = [];
  const child = spawn(executable, ["dev", "--ip", host, "--port", String(port)], {
    cwd: process.cwd(), env: process.env, stdio: ["ignore", "pipe", "pipe"],
  });
  for (const stream of [child.stdout, child.stderr]) {
    stream?.on("data", (chunk) => {
      logs.push(chunk.toString());
      if (logs.length > 40) logs.shift();
    });
  }
  return { child, logs };
}

async function waitForReady(baseURL, timeout, child, logs) {
  const deadline = Date.now() + timeout;
  let lastError = "not attempted";
  while (Date.now() < deadline) {
    if (child.exitCode !== null || child.signalCode !== null) {
      fail(`Wrangler exited before ready (code=${child.exitCode}, signal=${child.signalCode}): ${logs.join("").trim()}`);
    }
    try {
      const response = await fetch(`${baseURL}/assets/releases/current`, { redirect: "manual", signal: AbortSignal.timeout(1000) });
      if (response.status === 200) return;
      lastError = `status ${response.status}`;
    } catch (error) {
      lastError = error.message;
    }
    await new Promise((resolve) => setTimeout(resolve, 200));
  }
  fail(`Wrangler did not become ready within ${timeout}ms: ${lastError}; ${logs.join("").trim()}`);
}

async function stopWrangler(session, reason) {
  if (!session?.child) return;
  const { child } = session;
  if (child.exitCode === null && child.signalCode === null) {
    child.kill("SIGTERM");
    await Promise.race([once(child, "exit"), new Promise((resolve) => setTimeout(resolve, 5000))]);
    if (child.exitCode === null && child.signalCode === null) child.kill("SIGKILL");
  }
  event({ event: "session-stop", pid: child.pid, reason, exitCode: child.exitCode, signal: child.signalCode });
}

export async function runCanary() {
  const assetBundle = process.env.ASSET_BUNDLE;
  if (!assetBundle || !path.isAbsolute(assetBundle) || !existsSync(assetBundle) || !statSync(assetBundle).isDirectory()) {
    fail("ASSET_BUNDLE must name an existing absolute verified bundle directory");
  }
  const contract = inspectEnabledBundle(
    assetBundle,
    process.env.CANARY_RESOLUTION_DATE,
    process.env.CANARY_EXPIRED_CHANNEL,
    process.env.CANARY_EXPIRED_RESOLUTION_DATE,
  );
  const port = configuredPort(process.env.CANARY_PORT);
  const timeout = configuredTimeout(process.env.CANARY_TIMEOUT_MS);
  runBuild(assetBundle);
  const publicRoot = path.join(process.cwd(), "public");
  const { specs, releases } = buildProbeSpecs(publicRoot);
  const builtCurrent = readJSON(path.join(publicRoot, "assets", "releases", "current.json"), "built current channel");
  if (builtCurrent.release !== contract.current.release || builtCurrent.digest !== contract.current.digest ||
      builtCurrent.source !== contract.current.source || builtCurrent.campaign?.id !== contract.current.campaign.id) {
    fail("built current channel differs from accepted ASSET_BUNDLE");
  }
  let expiredFixture = "";
  if (contract.expired) {
    expiredFixture = path.join(publicRoot, expiredChannelPath.slice(1));
    copyFileSync(contract.expired.input, expiredFixture);
  }
  event({
    event: "resolution",
    resolutionDate: contract.resolutionDate,
    release: contract.current.release,
    source: contract.current.source,
    campaign: contract.current.campaign.id,
    digest: contract.current.digest,
    releases: contract.releases,
  });

  const baseURL = `http://${host}:${port}`;
  const session = startWrangler(port);
  event({ event: "session-start", pid: session.child.pid, host, port, routes: specs.length });
  let stopped = false;
  const shutdown = async (reason) => {
    if (stopped) return;
    stopped = true;
    await stopWrangler(session, reason);
  };
  const interrupt = (signal) => {
    shutdown(signal).finally(() => process.exit(signal === "SIGINT" ? 130 : 143));
  };
  process.once("SIGINT", interrupt);
  process.once("SIGTERM", interrupt);
  try {
    await waitForReady(baseURL, timeout, session.child, session.logs);
    event({ event: "session-ready", pid: session.child.pid, baseURL });
    for (const spec of specs) {
      for (const method of ["GET", "HEAD"]) {
        const capture = await probe(baseURL, method, spec);
        assertCapture(capture, spec[method.toLowerCase()]);
        if (method === "HEAD" && capture.bytes !== 0) fail(`HEAD ${spec.route}: body is not empty`);
        event({ event: "probe", ...capture });
      }
    }
    await probeResolvedAssets(baseURL, contract);
    await runBrowserCanary(baseURL, contract);
    event({
      event: "canary-pass",
      resolutionDate: contract.resolutionDate,
      release: contract.current.release,
      source: contract.current.source,
      campaign: contract.current.campaign.id,
      digest: contract.current.digest,
      releases,
      expiry: Boolean(contract.expired),
    });
  } catch (error) {
    event({ event: "canary-fail", message: error.message });
    throw error;
  } finally {
    process.removeListener("SIGINT", interrupt);
    process.removeListener("SIGTERM", interrupt);
    await shutdown("complete");
    if (expiredFixture) rmSync(expiredFixture, { force: true });
  }
}

if (import.meta.url === new URL(process.argv[1], "file:").href) {
  runCanary().catch((error) => {
    process.stderr.write(`seasonal assets canary: ${error.message}\n`);
    process.exitCode = 1;
  });
}
