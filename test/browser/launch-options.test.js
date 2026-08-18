import assert from "node:assert/strict";
import test from "node:test";

import { browserLaunchOptions } from "./launch-options.js";

test("macOS browser launch prefers Homebrew Chromium and disables code-sign clones", () => {
  const options = browserLaunchOptions({
    env: {},
    pathExists: (path) => path === "/opt/homebrew/bin/chromium",
    platform: "darwin",
  });

  assert.equal(options.executablePath, "/opt/homebrew/bin/chromium");
  assert.ok(options.args.includes("--disable-features=MacAppCodeSignClone"));
});

test("explicit Puppeteer browser override remains authoritative", () => {
  const options = browserLaunchOptions({
    env: { PUPPETEER_EXECUTABLE_PATH: "/custom/chromium" },
    pathExists: () => {
      throw new Error("filesystem probe must not run for an explicit override");
    },
  });

  assert.equal(options.executablePath, "/custom/chromium");
});

test("non-macOS browser launch retains Puppeteer fallback without the macOS feature flag", () => {
  const options = browserLaunchOptions({
    env: {},
    pathExists: () => false,
    platform: "linux",
  });

  assert.equal(options.executablePath, undefined);
  assert.deepEqual(options.args, ["--no-sandbox"]);
});
