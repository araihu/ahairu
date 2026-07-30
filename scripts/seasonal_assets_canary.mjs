#!/usr/bin/env node
// Local evidence probe. It deliberately starts one Wrangler session only.
import { createHash } from "node:crypto";
import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { once } from "node:events";
import { spawn, spawnSync } from "node:child_process";
import path from "node:path";

const host = "127.0.0.1";
const emptySHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855";
const releaseName = /^v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$/;

function fail(message) {
  throw new Error(message);
}

function sha256(bytes) {
  return createHash("sha256").update(bytes).digest("hex");
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

function assertPreproductionGates(publicRoot, releases) {
  if (process.env.CANARY_REQUIRE_TWO_RELEASES === "1" && releases.length < 2) {
    fail("enabled-campaign pre-production gate needs two retained immutable releases (set only for that gate)");
  }
  if (process.env.CANARY_REQUIRE_ENABLED_CAMPAIGN === "1") {
    const current = JSON.parse(readPublic(publicRoot, "assets/releases/current.json"));
    if (current.source !== "campaign" || !current.campaign) {
      fail("enabled-campaign pre-production gate needs current channel source=campaign with campaign metadata");
    }
  }
}

export async function runCanary() {
  const assetBundle = process.env.ASSET_BUNDLE;
  if (!assetBundle || !path.isAbsolute(assetBundle) || !existsSync(assetBundle) || !statSync(assetBundle).isDirectory()) {
    fail("ASSET_BUNDLE must name an existing absolute verified bundle directory");
  }
  const port = configuredPort(process.env.CANARY_PORT);
  const timeout = configuredTimeout(process.env.CANARY_TIMEOUT_MS);
  runBuild(assetBundle);
  const publicRoot = path.join(process.cwd(), "public");
  const { specs, releases } = buildProbeSpecs(publicRoot);
  assertPreproductionGates(publicRoot, releases);

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
    event({ event: "canary-pass", releases });
  } catch (error) {
    event({ event: "canary-fail", message: error.message });
    throw error;
  } finally {
    process.removeListener("SIGINT", interrupt);
    process.removeListener("SIGTERM", interrupt);
    await shutdown("complete");
  }
}

if (import.meta.url === new URL(process.argv[1], "file:").href) {
  runCanary().catch((error) => {
    process.stderr.write(`seasonal assets canary: ${error.message}\n`);
    process.exitCode = 1;
  });
}
