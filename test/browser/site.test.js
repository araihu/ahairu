import assert from "node:assert/strict";
import { spawn } from "node:child_process";
import { existsSync } from "node:fs";
import net from "node:net";
import { after, before, test } from "node:test";

import puppeteer from "puppeteer";

const canonicalOrigin = "https://araihu.com";
const pages = [
  { path: "/en/", kind: "home", language: "en" },
  { path: "/pt-br/", kind: "home", language: "pt-BR" },
  { path: "/es/", kind: "home", language: "es" },
  { path: "/brand/", kind: "brand", language: "en" },
  { path: "/pt-br/brand/", kind: "brand", language: "pt-BR" },
  { path: "/es/brand/", kind: "brand", language: "es" },
  { path: "/license/", kind: "license", language: "en" },
  { path: "/pt-br/license/", kind: "license", language: "pt-BR" },
  { path: "/es/license/", kind: "license", language: "es" },
];

let browser;
let origin;
let wrangler;
let wranglerExit;
let wranglerOutput = "";
let wranglerFailure;

before(async () => {
  const port = await availablePort();
  origin = `http://127.0.0.1:${port}`;
  wrangler = spawn("node_modules/.bin/wrangler", ["dev", "--ip", "127.0.0.1", "--port", String(port)], {
    stdio: ["ignore", "pipe", "pipe"],
  });
  wrangler.stdout.on("data", (chunk) => (wranglerOutput += chunk));
  wrangler.stderr.on("data", (chunk) => (wranglerOutput += chunk));
  wrangler.on("error", (error) => {
    wranglerFailure = error;
  });
  wranglerExit = childExit(wrangler);
  await waitForWrangler();

  const localChromium = "/opt/homebrew/bin/chromium";
  browser = await puppeteer.launch({
    headless: true,
    executablePath: process.env.PUPPETEER_EXECUTABLE_PATH || (existsSync(localChromium) ? localChromium : undefined),
    args: ["--no-sandbox"],
  });
});

after(async () => {
  let browserError;
  try {
    await browser?.close();
  } catch (error) {
    browserError = error;
  }

  let wranglerError;
  try {
    await stopChildProcess(wrangler, wranglerExit);
  } catch (error) {
    wranglerError = error;
  }

  if (browserError && wranglerError) {
    throw new AggregateError([browserError, wranglerError], "browser and Wrangler teardown failed");
  }
  if (browserError) throw browserError;
  if (wranglerError) throw wranglerError;
});

function childExit(child) {
  return new Promise((resolve) => {
    let settled = false;
    const finish = (result) => {
      if (settled) return;
      settled = true;
      resolve(result);
    };
    child.once("error", (error) => finish({ code: null, signal: null, error }));
    child.once("exit", (code, signal) => finish({ code, signal }));
  });
}

async function stopChildProcess(child, exitPromise, graceMS = 5000) {
  if (!child || !exitPromise) return undefined;
  if (child.exitCode !== null || child.signalCode !== null) return exitPromise;

  child.kill("SIGTERM");
  const result = await Promise.race([
    exitPromise,
    new Promise((resolve) => setTimeout(() => resolve(undefined), graceMS)),
  ]);
  if (result) return result;

  if (child.exitCode === null && child.signalCode === null) child.kill("SIGKILL");
  return exitPromise;
}

async function availablePort() {
  const listener = net.createServer();
  await new Promise((resolve, reject) => {
    listener.once("error", reject);
    listener.listen(0, "127.0.0.1", () => {
      listener.removeListener("error", reject);
      resolve();
    });
  });
  const port = listener.address().port;
  await new Promise((resolve, reject) => listener.close((error) => (error ? reject(error) : resolve())));
  return port;
}

async function waitForWrangler() {
  for (let attempt = 0; attempt < 100; attempt++) {
    if (wranglerFailure) {
      throw new Error(`wrangler dev failed to spawn: ${wranglerFailure.message}`);
    }
    if (wrangler.exitCode !== null) {
      throw new Error(`wrangler dev exited before startup:\n${wranglerOutput}`);
    }
    try {
      const response = await fetch(`${origin}/assets/styles.css`);
      if (response.status === 200) return;
    } catch {
      // The listener is still starting.
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
  throw new Error(`wrangler dev did not become ready:\n${wranglerOutput}`);
}

function alternatePaths(kind) {
  const suffix = kind === "home" ? "" : `${kind}/`;
  return {
    en: kind === "home" ? "/en/" : `/${suffix}`,
    "pt-BR": `/pt-br/${suffix}`,
    es: `/es/${suffix}`,
    "x-default": kind === "home" ? "/en/" : `/${suffix}`,
  };
}

async function openCheckedPage(pathname, options = {}) {
  const page = await browser.newPage();
  await page.setCacheEnabled(false);
  await page.setViewport({ width: options.width || 1280, height: options.height || 900, deviceScaleFactor: 1 });
  if (options.scheme) {
    await page.emulateMediaFeatures([{ name: "prefers-color-scheme", value: options.scheme }]);
  }
  const failures = [];
  page.on("console", (message) => {
    if (message.type() === "error") failures.push(`console: ${message.text()}`);
  });
  page.on("pageerror", (error) => failures.push(`page: ${error.message}`));
  page.on("requestfailed", (request) => failures.push(`request: ${request.url()} ${request.failure()?.errorText}`));
  page.on("response", (response) => {
    if (response.status() >= 400) failures.push(`response: ${response.status()} ${response.url()}`);
  });
  const response = await page.goto(`${origin}${pathname}`, { waitUntil: "networkidle0" });
  assert.equal(response.status(), 200, pathname);
  await waitForStableRender(page);
  assert.deepEqual(failures, [], pathname);
  return page;
}

async function waitForStableRender(page) {
  await page.evaluate(async () => {
    const deadlineMS = 10_000;
    const deadline = (message) => new Promise((_, reject) => {
      setTimeout(() => reject(new Error(message)), deadlineMS);
    });

    await Promise.race([
      document.fonts.ready,
      deadline("fonts did not settle within 10 seconds"),
    ]);

    const images = [...document.images];
    await Promise.race([
      Promise.all(images.map((image) => (image.complete ? image.decode().catch(() => {}) : new Promise((resolve) => {
        image.addEventListener("load", resolve, { once: true });
        image.addEventListener("error", resolve, { once: true });
      })))),
      deadline(`images did not settle within 10 seconds: ${images.filter((image) => !image.complete).map((image) => image.currentSrc || image.src).join(", ")}`),
    ]);
    await new Promise((resolve) => requestAnimationFrame(() => requestAnimationFrame(resolve)));
  });
}

test("all canonical pages expose reciprocal absolute metadata and valid JSON-LD", async (t) => {
  for (const expected of pages) {
    await t.test(expected.path, async () => {
      const page = await openCheckedPage(expected.path);
      try {
        const metadata = await page.evaluate(() => ({
          language: document.documentElement.lang,
          canonical: document.querySelector('link[rel="canonical"]')?.href,
          alternates: Object.fromEntries(
            [...document.querySelectorAll('link[rel="alternate"][hreflang]')].map((link) => [link.hreflang, link.href]),
          ),
          ogURL: document.querySelector('meta[property="og:url"]')?.content,
          ogImage: document.querySelector('meta[property="og:image"]')?.content,
          ogTitle: document.querySelector('meta[property="og:title"]')?.content,
          ogDescription: document.querySelector('meta[property="og:description"]')?.content,
          xImage: document.querySelector('meta[name="twitter:image"]')?.content,
          xTitle: document.querySelector('meta[name="twitter:title"]')?.content,
          xDescription: document.querySelector('meta[name="twitter:description"]')?.content,
          xCard: document.querySelector('meta[name="twitter:card"]')?.content,
          jsonLD: document.querySelector('#structured-data')?.textContent,
        }));
        const canonical = `${canonicalOrigin}${expected.path}`;
        assert.equal(metadata.language, expected.language);
        assert.equal(metadata.canonical, canonical);
        assert.equal(metadata.ogURL, canonical);
        assert.equal(metadata.xCard, "summary_large_image");
        assert.ok(metadata.ogTitle && metadata.ogDescription && metadata.xTitle && metadata.xDescription);
        assert.ok(metadata.ogImage.startsWith(`${canonicalOrigin}/social/`));
        assert.equal(metadata.xImage, metadata.ogImage);
        assert.deepEqual(
          metadata.alternates,
          Object.fromEntries(Object.entries(alternatePaths(expected.kind)).map(([language, route]) => [language, canonicalOrigin + route])),
        );
        assert.doesNotThrow(() => JSON.parse(metadata.jsonLD));
      } finally {
        await page.close();
      }
    });
  }
});

test("brand downloads are complete local responses in every locale", async () => {
  for (const pathname of ["/brand/", "/pt-br/brand/", "/es/brand/"]) {
    const page = await openCheckedPage(pathname);
    try {
      const downloads = await page.evaluate(async () =>
        Promise.all(
          [...document.querySelectorAll("a[download]")].map(async (link) => ({
            path: new URL(link.href).pathname,
            status: (await fetch(link.href)).status,
          })),
        ),
      );
      assert.equal(downloads.length, 8, pathname);
      assert.ok(downloads.every(({ path, status }) => path.startsWith("/assets/araihu/v0.1.0/") && status === 200), pathname);
    } finally {
      await page.close();
    }
  }
});

test("all canonical pages fit a 375px mobile viewport", async () => {
  for (const expected of pages) {
    const page = await openCheckedPage(expected.path, { width: 375, height: 812 });
    try {
      const dimensions = await page.evaluate(() => ({
        client: document.documentElement.clientWidth,
        scroll: document.documentElement.scrollWidth,
      }));
      assert.ok(dimensions.scroll <= dimensions.client, `${expected.path}: ${dimensions.scroll} > ${dimensions.client}`);
    } finally {
      await page.close();
    }
  }
});

test("representative pages satisfy the full responsive scheme matrix", async () => {
  const representatives = [
    { path: "/en/", selector: ".project-mark", count: 4 },
    { path: "/brand/", selector: ".variant-card", count: 4 },
    { path: "/license/", selector: ".terms-panel", count: 2 },
  ];
  for (const expected of representatives) {
    for (const width of [375, 768, 1280, 1440]) {
      for (const scheme of ["light", "dark"]) {
        const label = `${expected.path} ${width}px ${scheme}`;
        const page = await openCheckedPage(expected.path, { width, height: 900, scheme });
        try {
          const layout = await page.evaluate(({ selector }) => {
            const main = document.querySelector("main");
            const heading = main?.querySelector("h1");
            const specimens = [...document.querySelectorAll(selector)];
            return {
              clientWidth: document.documentElement.clientWidth,
              scrollWidth: document.documentElement.scrollWidth,
              mainWidth: main?.getBoundingClientRect().width || 0,
              headingHeight: heading?.getBoundingClientRect().height || 0,
              specimenCount: specimens.length,
              specimensVisible: specimens.every((element) => {
                const box = element.getBoundingClientRect();
                return box.width > 0 && box.height > 0;
              }),
              imagesDecoded: [...document.images].every((image) => image.complete && image.naturalWidth > 0),
              background: getComputedStyle(document.body).backgroundColor,
            };
          }, expected);
          assert.ok(layout.scrollWidth <= layout.clientWidth, `${label}: horizontal overflow`);
          assert.ok(layout.mainWidth > 0 && layout.headingHeight > 0, `${label}: primary content is not painted`);
          assert.equal(layout.specimenCount, expected.count, `${label}: representative count`);
          assert.ok(layout.specimensVisible, `${label}: hidden representative`);
          assert.ok(layout.imagesDecoded, `${label}: undecoded image`);
          assert.notEqual(layout.background, "rgba(0, 0, 0, 0)", `${label}: transparent page surface`);
        } finally {
          await page.close();
        }
      }
    }
  }
});

test("brand specimens retain approved variants and stable geometry", async () => {
  const page = await openCheckedPage("/brand/");
  try {
    const variants = await page.evaluate(() =>
      [...document.querySelectorAll(".variant-card")].map((card) => {
        const art = card.querySelector(".variant-art");
        const image = art.querySelector("img");
        const box = art.getBoundingClientRect();
        return {
          className: card.className,
          source: new URL(image.src).pathname,
          background: getComputedStyle(art).backgroundColor,
          width: box.width,
          height: box.height,
        };
      }),
    );
    assert.equal(variants.length, 4);
    for (const name of ["light", "dark", "signal", "tinted"]) {
      assert.ok(variants.some((variant) => variant.className.includes(`variant-${name}`)), name);
    }
    assert.ok(variants.every((variant) => variant.source.includes("transparent-optical.svg")));
    assert.equal(new Set(variants.map((variant) => variant.background)).size, 4);
    assert.ok(Math.max(...variants.map((variant) => variant.width)) - Math.min(...variants.map((variant) => variant.width)) <= 1);
    assert.ok(Math.max(...variants.map((variant) => variant.height)) - Math.min(...variants.map((variant) => variant.height)) <= 1);
  } finally {
    await page.close();
  }
});

test("light and dark schemes remain distinct", async () => {
  const schemes = [];
  for (const scheme of ["light", "dark"]) {
    const page = await openCheckedPage("/en/", { scheme });
    schemes.push(
      await page.evaluate(() => {
        function ink(mark) {
          const canvas = document.createElement("canvas");
          canvas.width = 64;
          canvas.height = 64;
          const context = canvas.getContext("2d");
          context.drawImage(mark, 0, 0, 64, 64);
          const pixels = context.getImageData(0, 0, 64, 64).data;
          let dark = 0;
          let light = 0;
          for (let index = 0; index < pixels.length; index += 4) {
            if (pixels[index + 3] === 0) continue;
            if (pixels[index] < 60 && pixels[index + 1] < 60 && pixels[index + 2] < 60) dark++;
            if (pixels[index] > 220 && pixels[index + 1] > 220 && pixels[index + 2] > 220) light++;
          }
          return { dark, light };
        }
        return {
          background: getComputedStyle(document.body).backgroundColor,
          header: ink(document.querySelector(".ahairu-brand img")),
          projects: [...document.querySelectorAll("img.project-mark")].map(ink),
        };
      }),
    );
    await page.close();
  }
  assert.notEqual(schemes[0].background, schemes[1].background);
  assert.equal(schemes[0].projects.length, 4);
  assert.equal(schemes[1].projects.length, 4);
  assert.ok(schemes[0].header.dark > 100, `light scheme header has only ${schemes[0].header.dark} dark pixels`);
  assert.ok(schemes[1].header.light > 100, `dark scheme header has only ${schemes[1].header.light} light pixels`);
  schemes[0].projects.forEach((mark, index) => {
    assert.ok(mark.dark > 100, `light scheme project ${index + 1} has only ${mark.dark} dark pixels`);
  });
  schemes[1].projects.forEach((mark, index) => {
    assert.ok(mark.light > 100, `dark scheme project ${index + 1} has only ${mark.light} light pixels`);
  });
});

test("child teardown escalates and awaits stubborn processes", async () => {
  const child = spawn(process.execPath, ["-e", "process.on('SIGTERM', () => {}); console.log('ready'); setInterval(() => {}, 1000)"], {
    stdio: ["ignore", "pipe", "ignore"],
  });
  const exitPromise = childExit(child);
  await new Promise((resolve, reject) => {
    child.once("error", reject);
    child.stdout.once("data", resolve);
  });
  const result = await stopChildProcess(child, exitPromise, 50);
  assert.equal(result.signal, "SIGKILL");
  assert.notEqual(child.signalCode, null);
});

test("dark scheme preserves paper specimen and terms-panel contrast", async () => {
  const brand = await openCheckedPage("/brand/", { scheme: "dark", width: 1280 });
  try {
    const identity = await brand.evaluate(() => {
      const surface = document.querySelector(".identity-master");
      const mark = surface.querySelector("img");
      const canvas = document.createElement("canvas");
      canvas.width = mark.naturalWidth;
      canvas.height = mark.naturalHeight;
      const context = canvas.getContext("2d");
      context.drawImage(mark, 0, 0);
      const pixels = context.getImageData(0, 0, canvas.width, canvas.height).data;
      let cobaltInk = 0;
      for (let index = 0; index < pixels.length; index += 4) {
        if (pixels[index + 3] > 0 && pixels[index + 2] > pixels[index] + 40) cobaltInk++;
      }
      return { background: getComputedStyle(surface).backgroundColor, cobaltInk };
    });
    assert.equal(identity.background, "rgb(243, 242, 233)");
    assert.ok(identity.cobaltInk > 1000, `identity specimen has only ${identity.cobaltInk} cobalt pixels`);
  } finally {
    await brand.close();
  }

  const license = await openCheckedPage("/license/", { scheme: "dark", width: 1280 });
  try {
    const headings = await license.evaluate(() =>
      [...document.querySelectorAll(".terms-panel .terms-heading h3")].map((heading) => ({
        color: getComputedStyle(heading).color,
        text: heading.textContent.trim(),
      })),
    );
    assert.equal(headings.length, 2);
    assert.deepEqual(headings.map(({ color }) => color), ["rgb(7, 17, 31)", "rgb(7, 17, 31)"]);
    assert.ok(headings.every(({ text }) => text.length > 0));
  } finally {
    await license.close();
  }
});

test("keyboard focus is visible and reduced motion removes decorative movement", async () => {
  const page = await openCheckedPage("/en/");
  try {
    await page.keyboard.press("Tab");
    const focus = await page.evaluate(() => {
      const active = document.activeElement;
      const style = getComputedStyle(active);
      return { className: active.className, outlineStyle: style.outlineStyle, outlineWidth: style.outlineWidth };
    });
    assert.equal(focus.className, "skip-link");
    assert.notEqual(focus.outlineStyle, "none");
    assert.notEqual(focus.outlineWidth, "0px");

    await page.emulateMediaFeatures([{ name: "prefers-reduced-motion", value: "reduce" }]);
    await page.reload({ waitUntil: "networkidle0" });
    const motion = await page.evaluate(() => ({
      cloud: getComputedStyle(document.querySelector(".storm-hero"), "::before").animationName,
      row: getComputedStyle(document.querySelector(".project-row")).transitionDuration,
    }));
    assert.equal(motion.cloud, "none");
    assert.equal(motion.row, "0s");
  } finally {
    await page.close();
  }
});

test("Worker redirects and 404s are observable over HTTP", async () => {
  for (const [source, destination] of [
    ["/en/brand?x=one%2Ftwo", "/brand/?x=one%2Ftwo"],
    ["/pt-br/license?x=1", "/pt-br/license/?x=1"],
  ]) {
    const response = await fetch(origin + source, { redirect: "manual" });
    assert.equal(response.status, 308, source);
    assert.equal(response.headers.get("Location"), destination, source);
  }
  assert.equal((await fetch(`${origin}/not-a-page`)).status, 404);
});
