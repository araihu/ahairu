import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { createServer } from "node:http";
import { extname, join } from "node:path";
import { fileURLToPath } from "node:url";
import test from "node:test";

import puppeteer from "puppeteer";

const css = await readFile(new URL("../site/brand.css", import.meta.url), "utf8");
const componentCSS = await readFile(new URL("../public/assets/styles.css", import.meta.url), "utf8");
const shellCSS = await readFile(new URL("../public/landingshell/assets/shell.css", import.meta.url), "utf8");
const publicRoot = fileURLToPath(new URL("../public/", import.meta.url));

async function servePublic() {
  const contentTypes = new Map([
    [".css", "text/css"],
    [".html", "text/html"],
    [".js", "text/javascript"],
    [".mp4", "video/mp4"],
    [".svg", "image/svg+xml"],
	[".webp", "image/webp"],
  ]);
  const server = createServer(async (request, response) => {
    try {
      const pathname = decodeURIComponent(new URL(request.url, "http://127.0.0.1").pathname);
      if (pathname === "/api/project-versions") {
        response.setHeader("content-type", "text/html; charset=utf-8");
        response.end('<span id="goshtoso-version-slot" class="project-version"><a href="https://github.com/araihu/goshtoso/releases/tag/v0.1.7">v0.1.7</a></span><span id="goshtoso-charts-version-slot" class="project-version" hx-swap-oob="outerHTML"><a href="https://github.com/araihu/goshtoso-charts/tree/v0.0.1">v0.0.1</a></span>');
        return;
      }
      const relativePath = pathname.endsWith("/") ? `${pathname.slice(1)}index.html` : pathname.slice(1);
      const filePath = join(publicRoot, relativePath);
      if (!filePath.startsWith(publicRoot)) throw new Error("invalid path");
      response.setHeader("content-type", contentTypes.get(extname(filePath)) ?? "application/octet-stream");
      response.end(await readFile(filePath));
    } catch {
      response.statusCode = 404;
      response.end("not found");
    }
  });
  await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
  const address = server.address();
  return {
    origin: `http://127.0.0.1:${address.port}`,
    close: () => new Promise((resolve, reject) => server.close((error) => error ? reject(error) : resolve())),
  };
}

const fixture = `<!doctype html>
<style>${componentCSS}\n${shellCSS}\n${css}</style>
<header class="ahairu-header"><nav class="ahairu-primary-links" aria-label="Project navigation"><a href="#home">Home</a><a href="#libs">Libs</a><a href="#apps">Apps</a><a href="#blog">Blog</a></nav></header>
<section class="storm-hero"></section>
<a class="featured-visual" href="#"><figure class="featured-demo"><video data-featured-montage muted playsinline></video></figure></a>
<a class="signal-button" href="#">Signal</a>
<a class="project-card" href="#"><article class="shadow-lg transition-[translate,box-shadow] duration-150 hover:translate-y-1.5 hover:shadow-sm active:translate-y-2 active:shadow-none motion-reduce:hover:translate-none motion-reduce:active:translate-none motion-reduce:transition-none project-card-surface"><div class="project-art"><span class="project-art-name">Project</span><span class="project-mark"></span></div></article></a>
<a class="project-card project-card--openapi" href="#"><div class="project-art project-art--openapi"><div class="openapi-viewport"><div class="openapi-stream"><div class="openapi-block"><span><b>openapi:</b> 3.1.0</span></div><div class="openapi-block"><span><b>openapi:</b> 3.1.0</span></div></div></div></div></a>
<a class="more-row" href="#"><div class="more-art more-art--chart">More</div></a>
<article class="more-row more-row--muamba"><a class="more-project-link" href="#">Muamba</a><div class="more-art more-art--muamba"><div class="muamba-drops"><img class="muamba-drop" src="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 64 64'%3E%3Crect width='64' height='64' fill='%23c7ff4a'/%3E%3C/svg%3E" alt=""></div></div></article>
<div class="landing-shell__mobile-navigation storm-mobile-menu">
  <div class="landing-shell__mobile-enhanced">
    <button class="landing-shell__mobile-trigger is-bottom-left storm-mobile-trigger"><span class="landing-shell__mobile-trigger-icon"><i></i><i></i><i></i></span>Menu</button>
    <div class="fixed inset-0 transition-opacity duration-200 motion-reduce:transition-none" aria-hidden="true"></div>
    <aside class="fixed transition-transform duration-200 motion-reduce:transform-none motion-reduce:transition-none storm-mobile-panel">Navigation</aside>
  </div>
  <details class="landing-shell__mobile-fallback"><summary class="landing-shell__mobile-fallback-trigger is-bottom-left storm-mobile-trigger">Menu</summary></details>
</div>`;

async function reducedMotionPage(browser, viewport) {
  const page = await browser.newPage();
  await page.setViewport(viewport);
  await page.emulateMediaFeatures([{ name: "prefers-reduced-motion", value: "reduce" }]);
  await page.setContent(fixture);
  return page;
}

async function revealCharts(page) {
  await page.$eval("[data-chart-bundle-trigger]", (element) => element.scrollIntoView({ block: "center" }));
  await page.waitForFunction(() => [
    "[data-paje-actual-chart] canvas",
    "[data-x9-live-availability] canvas",
    "[data-goshtoso-heart-chart] canvas",
  ].every((selector) => document.querySelector(selector)), { timeout: 15_000 });
}

test("desktop and mobile project navigation meet at the 640px shell boundary", { timeout: 30_000 }, async () => {
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.setContent(fixture);

    for (const width of [640, 641, 700, 701]) {
      await page.setViewport({ width, height: 844 });
      const visibility = await page.evaluate(() => {
        const visible = (element) => getComputedStyle(element).display !== "none" && element.getClientRects().length > 0;
        return {
          desktop: visible(document.querySelector(".ahairu-primary-links")),
          mobile: visible(document.querySelector(".landing-shell__mobile-navigation")),
        };
      });

      assert.equal(Number(visibility.desktop) + Number(visibility.mobile), 1, `${width}px must expose exactly one project navigation surface`);
      assert.equal(visibility.mobile, width <= 640, `${width}px must follow the App Shell 640px mobile boundary`);
    }
  } finally {
    await browser.close();
  }
});

test("storm backdrop selects one theme video, covers the hero, and respects reduced motion", { timeout: 60_000 }, async () => {
  const server = await servePublic();
  const browser = await puppeteer.launch({ headless: true });
  try {
    for (const theme of ["dark", "light"]) {
      const page = await browser.newPage();
      const requestedVideos = [];
      page.on("request", (request) => {
        if (request.url().includes("/assets/video/")) requestedVideos.push(request.url());
      });
      await page.setViewport({ width: theme === "dark" ? 2048 : 1280, height: 720 });
      await page.emulateMediaFeatures([
        { name: "prefers-color-scheme", value: theme },
        { name: "prefers-reduced-motion", value: "no-preference" },
      ]);
      await page.goto(`${server.origin}/en/`, { waitUntil: "networkidle0" });
      await page.waitForFunction(() => {
        const video = document.querySelector("[data-storm-backdrop]");
        return video?.currentSrc && !video.paused;
      });
      const state = await page.evaluate(() => {
        const hero = document.querySelector(".storm-hero").getBoundingClientRect();
        const stage = document.querySelector(".storm-video-stage").getBoundingClientRect();
        const navigation = document.querySelector(".ahairu-nav");
        const navigationRect = navigation.getBoundingClientRect();
        const video = document.querySelector("[data-storm-backdrop]");
        const videoRect = video.getBoundingClientRect();
        return {
          currentSrc: video.currentSrc,
          filter: getComputedStyle(document.querySelector(".storm-video-filter")).backgroundImage,
          coversHero: videoRect.left <= hero.left && videoRect.right >= hero.right && videoRect.top <= hero.top && videoRect.bottom >= hero.bottom,
          coversViewport: stage.left <= 0 && stage.right >= innerWidth && videoRect.left <= 0 && videoRect.right >= innerWidth,
          topbarFullWidth: navigationRect.left === 0 && navigationRect.right === innerWidth,
          topbarPadding: parseFloat(getComputedStyle(navigation).paddingInlineStart),
        };
      });
      assert.match(state.currentSrc, new RegExp(`storm-${theme}-v1\\.mp4$`));
      assert.equal(state.coversHero, true);
      assert.equal(state.coversViewport, true);
      assert.equal(state.topbarFullWidth, true);
      assert.ok(state.topbarPadding >= 24, `expected at least 24px topbar padding, received ${state.topbarPadding}px`);
      assert.notEqual(state.filter, "none");
      assert.deepEqual([...new Set(requestedVideos.map((url) => new URL(url).pathname))], [`/assets/video/storm-${theme}-v1.mp4`]);

      await page.evaluate(() => window.scrollTo(0, 360));
      await new Promise((resolve) => setTimeout(resolve, 100));
      assert.notEqual(await page.$eval(".storm-hero", (hero) => hero.style.getPropertyValue("--storm-parallax")), "0.0px");
      await page.close();
    }

    const reducedPage = await browser.newPage();
    await reducedPage.emulateMediaFeatures([
      { name: "prefers-color-scheme", value: "dark" },
      { name: "prefers-reduced-motion", value: "reduce" },
    ]);
    await reducedPage.goto(`${server.origin}/en/`, { waitUntil: "networkidle0" });
    const reducedState = await reducedPage.evaluate(() => {
      const video = document.querySelector("[data-storm-backdrop]");
      return {
        paused: video.paused,
		montagePaused: document.querySelector("[data-featured-montage]").paused,
        stage: getComputedStyle(document.querySelector(".storm-video-stage")).display,
        parallax: document.querySelector(".storm-hero").style.getPropertyValue("--storm-parallax"),
      };
    });
    assert.deepEqual(reducedState, { paused: true, montagePaused: true, stage: "none", parallax: "" });
  } finally {
    await browser.close();
    await server.close();
  }
});

test("featured Goshtoso montage owns the full showcase width without the former CSS window", { timeout: 30_000 }, async () => {
  const server = await servePublic();
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.setViewport({ width: 1280, height: 720 });
    await page.goto(`${server.origin}/en/`, { waitUntil: "domcontentloaded" });
    await page.$eval(".featured-demo", (element) => element.scrollIntoView({ block: "center" }));
    await page.waitForFunction(() => document.querySelector("[data-featured-montage]").currentSrc);
    const state = await page.evaluate(() => {
      const visual = document.querySelector(".featured-visual").getBoundingClientRect();
      const demo = document.querySelector(".featured-demo").getBoundingClientRect();
      return {
        sameWidth: Math.abs(visual.width - demo.width) < 1,
        ratio: demo.width / demo.height,
        formerWindow: Boolean(document.querySelector(".featured-window")),
        source: document.querySelector("[data-featured-montage]").currentSrc,
      };
    });
    assert.equal(state.sameWidth, true);
    assert.ok(Math.abs(state.ratio - 16 / 9) < 0.02);
    assert.equal(state.formerWindow, false);
    assert.match(state.source, /goshtoso-components-montage-v1\.mp4$/);
    assert.doesNotMatch(css, /\.featured-window/);
  } finally {
    await browser.close();
    await server.close();
  }
});

test("project maturity labels stay attached to the correct products", { timeout: 30_000 }, async () => {
  const server = await servePublic();
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await browser.newPage();
    let versionRequests = 0;
    page.on("request", (request) => {
      if (new URL(request.url()).pathname === "/api/project-versions") versionRequests += 1;
    });
    await page.setViewport({ width: 1280, height: 720 });
    await page.goto(`${server.origin}/en/`, { waitUntil: "domcontentloaded" });
    await page.waitForSelector("#goshtoso-version-slot a");
    await page.waitForSelector("#goshtoso-charts-version-slot a");
    const labels = await page.evaluate(() => ({
      goshtoso: document.querySelector(".featured-copy [data-status]")?.textContent.trim(),
      manja: document.querySelector(".project-tile--2 [data-status]")?.textContent.trim(),
      paje: document.querySelector(".project-tile--3 [data-status]")?.textContent.trim(),
      x9: document.querySelector(".project-tile--4 [data-status]")?.textContent.trim(),
      charts: document.querySelector(".more-list li:nth-child(2) [data-status]")?.textContent.trim(),
      goshtosoVersion: document.querySelector("#goshtoso-version-slot")?.textContent.trim(),
      chartsVersion: document.querySelector("#goshtoso-charts-version-slot")?.textContent.trim(),
      unlabeledSecondary: document.querySelectorAll(".more-list li:first-child [data-status], .more-list li:nth-child(3) [data-status]").length,
    }));
    assert.deepEqual(labels, {
      goshtoso: "BETA",
      manja: "WIP",
      paje: "WIP",
      x9: "WIP",
      charts: "ALPHA",
      goshtosoVersion: "v0.1.7",
      chartsVersion: "v0.0.1",
      unlabeledSecondary: 0,
    });
    assert.equal(versionRequests, 1);
  } finally {
    await browser.close();
    await server.close();
  }
});

test("featured Goshtoso montage plays only while its visual is hovered", { timeout: 30_000 }, async () => {
  const server = await servePublic();
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.setViewport({ width: 1280, height: 720 });
    await page.goto(`${server.origin}/en/`, { waitUntil: "domcontentloaded" });
    await page.$eval(".featured-demo", (element) => element.scrollIntoView({ block: "center" }));
    await new Promise((resolve) => setTimeout(resolve, 300));
    assert.equal(await page.$eval("[data-featured-montage]", (video) => video.paused), true);

    await page.hover(".featured-visual");
    await page.waitForFunction(() => !document.querySelector("[data-featured-montage]").paused);
    await page.mouse.move(0, 0);
    await page.waitForFunction(() => document.querySelector("[data-featured-montage]").paused);
  } finally {
    await browser.close();
    await server.close();
  }
});

test("chart runtimes and payload stay out of first paint until the HTMX reveal trigger", { timeout: 30_000 }, async () => {
  const server = await servePublic();
  const browser = await puppeteer.launch({ headless: true, args: ["--enable-webgl", "--ignore-gpu-blocklist"] });
  try {
    const page = await browser.newPage();
    const requested = [];
    page.on("request", (request) => requested.push(new URL(request.url()).pathname));
    await page.setViewport({ width: 1280, height: 720 });
    await page.goto(`${server.origin}/en/`, { waitUntil: "domcontentloaded" });
    await new Promise((resolve) => setTimeout(resolve, 250));

    const initial = await page.evaluate(() => ({
      placeholders: document.querySelectorAll("[data-chart-placeholder]").length,
      actualCharts: document.querySelectorAll("[data-paje-actual-chart], [data-x9-live-availability], [data-goshtoso-heart-chart]").length,
      canvases: document.querySelectorAll("canvas").length,
      echarts: Boolean(window.echarts),
    }));
    assert.deepEqual(initial, { placeholders: 3, actualCharts: 0, canvases: 0, echarts: false });
    assert.equal(requested.some((path) => path.includes("/charts/assets/js/runtime/")), false);
    assert.equal(requested.includes("/fragments/en/charts.html"), false);

    await revealCharts(page);
    const chartRequests = requested.filter((path) => path.includes("/charts/assets/js/runtime/"));
    assert.deepEqual(chartRequests.sort(), [
      "/charts/assets/js/runtime/echarts/5.4.3/echarts.min.js",
      "/charts/assets/js/runtime/three-d/2.0.9/runtime.min.js",
    ]);
    assert.equal(requested.filter((path) => path === "/fragments/en/charts.html").length, 1);
    assert.equal(await page.$$eval("[data-chart-placeholder]", (elements) => elements.length), 0);
  } finally {
    await browser.close();
    await server.close();
  }
});

test("Pajé uses an undecorated actual chart that fills its media area", { timeout: 30_000 }, async () => {
  const server = await servePublic();
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.setViewport({ width: 884, height: 781 });
    await page.goto(`${server.origin}/en/`, { waitUntil: "domcontentloaded" });
    await revealCharts(page);
    await page.$eval("[data-paje-actual-chart]", (element) => element.scrollIntoView({ block: "center" }));
    await page.waitForSelector("[data-paje-actual-chart] canvas");
    const state = await page.$eval("[data-paje-actual-chart]", (art) => {
      const artRect = art.getBoundingClientRect();
      const canvasRect = art.querySelector("canvas").getBoundingClientRect();
      return {
        widthDelta: Math.abs(artRect.width - canvasRect.width),
        heightDelta: Math.abs(artRect.height - canvasRect.height),
        surface: getComputedStyle(art.querySelector("figure")).getPropertyValue("--color-chart-surface").trim(),
        images: art.querySelectorAll("img").length,
        buttons: art.querySelectorAll("button").length,
        captions: art.querySelectorAll("figcaption").length,
      };
    });
    assert.ok(state.widthDelta < 1);
    assert.ok(state.heightDelta < 1);
    assert.equal(state.surface, "#07111f");
    assert.deepEqual({ images: state.images, buttons: state.buttons, captions: state.captions }, { images: 0, buttons: 0, captions: 0 });
  } finally {
    await browser.close();
    await server.close();
  }
});

test("Pajé restarts its force layout on every card hover", { timeout: 30_000 }, async () => {
  const server = await servePublic();
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.setViewport({ width: 884, height: 781 });
    await page.goto(`${server.origin}/en/`, { waitUntil: "domcontentloaded" });
    await revealCharts(page);
    await page.waitForSelector("[data-paje-actual-chart] [_echarts_instance_]");
    await page.$eval("[data-paje-actual-chart]", (element) => element.scrollIntoView({ block: "center" }));

    const hostSelector = "[data-paje-actual-chart] [_echarts_instance_]";
    const instanceID = () => page.$eval(hostSelector, (element) => element.getAttribute("_echarts_instance_"));
    const initialID = await instanceID();

    await page.hover(".project-tile--3 .project-card");
    await page.waitForFunction((previousID) => {
      const host = document.querySelector("[data-paje-actual-chart] [_echarts_instance_]");
      return host && host.getAttribute("_echarts_instance_") !== previousID;
    }, {}, initialID);
    const firstHoverID = await instanceID();

    await page.mouse.move(1, 1);
    await page.hover(".project-tile--3 .project-card");
    await page.waitForFunction((previousID) => {
      const host = document.querySelector("[data-paje-actual-chart] [_echarts_instance_]");
      return host && host.getAttribute("_echarts_instance_") !== previousID;
    }, {}, firstHoverID);

    const state = await page.$eval(hostSelector, (host) => ({
      instanceID: host.getAttribute("_echarts_instance_"),
      layout: window.echarts.getInstanceByDom(host).getOption().series[0].layout,
    }));
    assert.notEqual(state.instanceID, firstHoverID);
    assert.equal(state.layout, "force");
  } finally {
    await browser.close();
    await server.close();
  }
});

test("X-9 availability chart fills its media area", { timeout: 30_000 }, async () => {
  const server = await servePublic();
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.setViewport({ width: 884, height: 781 });
    await page.goto(`${server.origin}/en/`, { waitUntil: "domcontentloaded" });
    await revealCharts(page);
    await page.waitForSelector("[data-x9-live-availability] canvas");
    const state = await page.$eval("[data-x9-live-availability]", (art) => {
      const artRect = art.getBoundingClientRect();
      const canvasRect = art.querySelector("canvas").getBoundingClientRect();
      return {
        tick: art.dataset.x9Tick,
        categories: art.dataset.x9Categories,
        widthDelta: Math.abs(artRect.width - canvasRect.width),
        heightDelta: Math.abs(artRect.height - canvasRect.height),
        images: art.querySelectorAll("img").length,
        buttons: art.querySelectorAll("button").length,
        captions: art.querySelectorAll("figcaption").length,
      };
    });
    assert.equal(state.tick, "0");
    assert.equal(state.categories, "36");
    assert.ok(state.widthDelta < 1);
    assert.ok(state.heightDelta < 1);
    assert.deepEqual({ images: state.images, buttons: state.buttons, captions: state.captions }, { images: 0, buttons: 0, captions: 0 });
  } finally {
    await browser.close();
    await server.close();
  }
});

test("X-9 ticks only while its card is hovered", { timeout: 30_000 }, async () => {
  const server = await servePublic();
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.setViewport({ width: 884, height: 781 });
    await page.goto(`${server.origin}/en/`, { waitUntil: "domcontentloaded" });
    await revealCharts(page);
    await page.waitForFunction(() => document.querySelector("[data-x9-live-availability]")?.dataset.x9Tick);
    const idleTick = await page.$eval("[data-x9-live-availability]", (art) => art.dataset.x9Tick);
    await new Promise((resolve) => setTimeout(resolve, 2_200));
    assert.equal(await page.$eval("[data-x9-live-availability]", (art) => art.dataset.x9Tick), idleTick);

    await page.hover(".project-tile--4 .project-card");
    await page.waitForFunction((tick) => document.querySelector("[data-x9-live-availability]").dataset.x9Tick !== tick, {}, idleTick);
    const activeTick = await page.$eval("[data-x9-live-availability]", (art) => art.dataset.x9Tick);
    await page.mouse.move(0, 0);
    await new Promise((resolve) => setTimeout(resolve, 2_200));
    assert.equal(await page.$eval("[data-x9-live-availability]", (art) => art.dataset.x9Tick), activeTick);
  } finally {
    await browser.close();
    await server.close();
  }
});

test("Goshtoso Charts uses an actual undecorated Line3D heart", { timeout: 30_000 }, async () => {
  const server = await servePublic();
  const browser = await puppeteer.launch({ headless: true, args: ["--enable-webgl", "--ignore-gpu-blocklist"] });
  try {
    const page = await browser.newPage();
    await page.setViewport({ width: 884, height: 781 });
    await page.goto(`${server.origin}/en/`, { waitUntil: "domcontentloaded" });
    await revealCharts(page);
    await page.waitForSelector("[data-goshtoso-heart-chart] canvas", { timeout: 10_000 });
    const state = await page.$eval("[data-goshtoso-heart-chart]", (art) => {
      const host = art.querySelector("[_echarts_instance_]");
      const option = window.echarts.getInstanceByDom(host).getOption();
      const point = (index) => option.series[0].data[index].value ?? option.series[0].data[index];
      const artRect = art.getBoundingClientRect();
      const hostRect = host.getBoundingClientRect();
      const canvasRect = art.querySelector("canvas").getBoundingClientRect();
      return {
        type: option.series[0].type,
        autoRotate: art.querySelector("figure").getAttribute("data-goshtoso-charts-line3d-auto-rotate"),
        pointCount: option.series[0].data.length,
        first: point(0),
        quarter: point(80),
        widthDelta: Math.abs(artRect.width - hostRect.width),
        heightDelta: Math.abs(artRect.height - hostRect.height),
        canvasCoversHost: canvasRect.width >= hostRect.width && canvasRect.height >= hostRect.height,
        images: art.querySelectorAll("img").length,
        buttons: art.querySelectorAll("button").length,
        detailsDisplay: getComputedStyle(art.querySelector("[data-line3d-exact-data]")).display,
      };
    });
    assert.equal(state.type, "line3D");
    assert.equal(state.autoRotate, "true");
    assert.equal(state.pointCount, 321);
    assert.deepEqual(state.first, [0, 0, 5]);
    assert.ok(Math.abs(state.quarter[0] - 16) < 1e-9);
    assert.ok(Math.abs(state.quarter[1] + 2.4) < 1e-9);
    assert.ok(Math.abs(state.quarter[2] - 4) < 1e-9);
    assert.ok(state.widthDelta < 3);
    assert.ok(state.heightDelta < 3);
    assert.equal(state.canvasCoversHost, true);
    assert.deepEqual({ images: state.images, buttons: state.buttons, detailsDisplay: state.detailsDisplay }, { images: 0, buttons: 0, detailsDisplay: "none" });
  } finally {
    await browser.close();
    await server.close();
  }
});

test("Goshtoso App Shells art previews desktop and mobile shell composition", { timeout: 30_000 }, async () => {
  const server = await servePublic();
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.setViewport({ width: 884, height: 781 });
    await page.goto(`${server.origin}/en/`, { waitUntil: "networkidle0" });
    const selector = "[data-shell-preview]";
    const before = await page.$eval(selector, (art) => ({
      desktop: art.querySelectorAll(".shell-preview--desktop").length,
      mobile: art.querySelectorAll(".shell-preview--mobile").length,
      images: art.querySelectorAll("img").length,
      desktopTransform: getComputedStyle(art.querySelector(".shell-preview--desktop")).transform,
      mobileTransform: getComputedStyle(art.querySelector(".shell-preview--mobile")).transform,
    }));
    await page.hover(".more-list li:first-child .more-row");
    await new Promise((resolve) => setTimeout(resolve, 300));
    const after = await page.$eval(selector, (art) => ({
      desktopTransform: getComputedStyle(art.querySelector(".shell-preview--desktop")).transform,
      mobileTransform: getComputedStyle(art.querySelector(".shell-preview--mobile")).transform,
    }));
    assert.deepEqual({ desktop: before.desktop, mobile: before.mobile, images: before.images }, { desktop: 1, mobile: 1, images: 0 });
    assert.notEqual(after.desktopTransform, before.desktopTransform);
    assert.notEqual(after.mobileTransform, before.mobileTransform);
  } finally {
    await browser.close();
    await server.close();
  }
});

test("Goshtoso pressed card owns the whole-card hover response", { timeout: 30_000 }, async () => {
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.setViewport({ width: 1280, height: 720 });
    await page.setContent(fixture);
    const before = await page.evaluate(() => {
      const card = getComputedStyle(document.querySelector(".project-card-surface"));
      return { translate: card.translate, shadow: card.boxShadow };
    });
    await page.hover(".project-card-surface");
    await new Promise((resolve) => setTimeout(resolve, 220));
    const after = await page.evaluate(() => {
      const card = getComputedStyle(document.querySelector(".project-card-surface"));
      return {
        translate: card.translate,
        shadow: card.boxShadow,
        art: getComputedStyle(document.querySelector(".project-art"), "::before").transform,
      };
    });
    assert.notEqual(after.translate, before.translate);
    assert.notEqual(after.shadow, before.shadow);
    assert.notEqual(after.art, "none");
  } finally {
    await browser.close();
  }
});

test("Manja OpenAPI text scrolls only while its card is hovered", { timeout: 30_000 }, async () => {
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.setContent(fixture);
    const before = await page.$eval(".openapi-stream", (element) => ({
      animation: getComputedStyle(element).animationName,
      playState: getComputedStyle(element).animationPlayState,
      transform: getComputedStyle(element).transform,
      viewportHeight: element.closest(".project-art--openapi").getBoundingClientRect().height,
    }));
    await new Promise((resolve) => setTimeout(resolve, 180));
    const idle = await page.$eval(".openapi-stream", (element) => getComputedStyle(element).transform);
    assert.equal(before.animation, "openapi-scroll");
    assert.equal(before.playState, "paused");
    assert.ok(before.viewportHeight <= 512);
    assert.equal(idle, before.transform);

    await page.hover(".project-card--openapi");
    await new Promise((resolve) => setTimeout(resolve, 180));
    const active = await page.$eval(".openapi-stream", (element) => ({
      playState: getComputedStyle(element).animationPlayState,
      transform: getComputedStyle(element).transform,
    }));
    assert.equal(active.playState, "running");
    assert.notEqual(active.transform, idle);
  } finally {
    await browser.close();
  }
});

test("official Muamba crate marks fall only while their row is hovered", { timeout: 30_000 }, async () => {
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.setContent(fixture);
    const before = await page.$eval(".muamba-drop", (element) => ({
      animation: getComputedStyle(element).animationName,
      playState: getComputedStyle(element).animationPlayState,
      transform: getComputedStyle(element).transform,
    }));
    await new Promise((resolve) => setTimeout(resolve, 180));
    const idle = await page.$eval(".muamba-drop", (element) => getComputedStyle(element).transform);
    assert.equal(before.animation, "muamba-fall");
    assert.equal(before.playState, "paused");
    assert.equal(idle, before.transform);

    await page.hover(".more-row--muamba");
    await new Promise((resolve) => setTimeout(resolve, 180));
    const active = await page.$eval(".muamba-drop", (element) => ({
      playState: getComputedStyle(element).animationPlayState,
      transform: getComputedStyle(element).transform,
    }));
    assert.equal(active.playState, "running");
    assert.notEqual(active.transform, idle);
  } finally {
    await browser.close();
  }
});

test("secondary project art gains a visible border on row hover", { timeout: 30_000 }, async () => {
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.setViewport({ width: 1280, height: 720 });
    await page.setContent(fixture);
    const before = await page.$eval(".more-art", (element) => getComputedStyle(element).borderColor);
    await page.hover(".more-row");
    await new Promise((resolve) => setTimeout(resolve, 240));
    const after = await page.$eval(".more-art", (element) => ({
      color: getComputedStyle(element).borderColor,
      style: getComputedStyle(element).borderStyle,
      width: getComputedStyle(element).borderWidth,
    }));
    assert.equal(before, "rgba(0, 0, 0, 0)");
    assert.deepEqual(after, { color: "rgb(49, 88, 143)", style: "solid", width: "1px" });
  } finally {
    await browser.close();
  }
});

test("Goshtoso Charts art leans left on row hover", { timeout: 30_000 }, async () => {
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await browser.newPage();
    await page.setViewport({ width: 1280, height: 720 });
    await page.setContent(fixture);
    await page.hover(".more-row");
    await new Promise((resolve) => setTimeout(resolve, 280));
    const rotation = await page.$eval(".more-art--chart", (element) => {
      const matrix = new DOMMatrixReadOnly(getComputedStyle(element).transform);
      return Math.atan2(matrix.b, matrix.a) * 180 / Math.PI;
    });
    assert.ok(rotation < 0, `expected negative rotation, received ${rotation}`);
  } finally {
    await browser.close();
  }
});

test("reduced motion removes card movement", { timeout: 30_000 }, async () => {
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await reducedMotionPage(browser, { width: 1280, height: 720 });
    const beforeHover = await page.evaluate(() => ({
      cardTransition: getComputedStyle(document.querySelector(".project-card-surface")).transitionProperty,
      cardTranslate: getComputedStyle(document.querySelector(".project-card-surface")).translate,
      markTransform: getComputedStyle(document.querySelector(".project-mark")).transform,
      moreTransform: getComputedStyle(document.querySelector(".more-art")).transform,
	  openapiAnimation: getComputedStyle(document.querySelector(".openapi-stream")).animationName,
	  muambaAnimation: getComputedStyle(document.querySelector(".muamba-drop")).animationName,
    }));
    assert.deepEqual(beforeHover, {
      cardTransition: "none",
      cardTranslate: "none",
      markTransform: "none",
      moreTransform: "none",
	  openapiAnimation: "none",
	  muambaAnimation: "none",
    });

    await page.hover(".project-card");
    const cardHover = await page.evaluate(() => ({
      card: getComputedStyle(document.querySelector(".project-card-surface")).translate,
      artBefore: getComputedStyle(document.querySelector(".project-art"), "::before").transform,
      artAfter: getComputedStyle(document.querySelector(".project-art"), "::after").transform,
      name: getComputedStyle(document.querySelector(".project-art-name")).transform,
      mark: getComputedStyle(document.querySelector(".project-mark")).transform,
    }));
    assert.deepEqual(cardHover, { card: "none", artBefore: "none", artAfter: "none", name: "none", mark: "none" });

    await page.hover(".more-row");
    assert.equal(await page.$eval(".more-art", (element) => getComputedStyle(element).transform), "none");
  } finally {
    await browser.close();
  }
});

test("reduced motion makes the mobile drawer and trigger instantaneous", { timeout: 30_000 }, async () => {
  const browser = await puppeteer.launch({ headless: true });
  try {
    const page = await reducedMotionPage(browser, { width: 390, height: 844 });
    const closed = await page.evaluate(() => ({
      trigger: getComputedStyle(document.querySelector(".landing-shell__mobile-trigger")).transitionDuration,
      scrim: getComputedStyle(document.querySelector("[aria-hidden='true'].fixed")).transitionDuration,
      drawer: getComputedStyle(document.querySelector(".storm-mobile-panel")).transitionDuration,
      fallback: getComputedStyle(document.querySelector(".landing-shell__mobile-fallback-trigger")).transitionDuration,
    }));
    assert.deepEqual(closed, { trigger: "0s", scrim: "0s", drawer: "0s", fallback: "0s" });

	await page.evaluate(() => document.documentElement.classList.add("landing-shell-mobile-navigation-ready"));
    await page.hover(".landing-shell__mobile-trigger");
    const opened = await page.evaluate(() => ({
      triggerTransform: getComputedStyle(document.querySelector(".landing-shell__mobile-trigger")).transform,
      drawerTransform: getComputedStyle(document.querySelector(".storm-mobile-panel")).transform,
      drawerTransition: getComputedStyle(document.querySelector(".storm-mobile-panel")).transitionDuration,
    }));
    assert.deepEqual(opened, {
      triggerTransform: "none",
      drawerTransform: "none",
      drawerTransition: "0s",
    });
  } finally {
    await browser.close();
  }
});
