import { mkdtemp, mkdir, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { spawnSync } from "node:child_process";

import puppeteer from "puppeteer";

const root = resolve(import.meta.dirname, "..");
const outputDirectory = join(root, "site", "visuals");
const temporaryDirectory = await mkdtemp(join(tmpdir(), "ahairu-project-visuals-"));
const ffmpeg = process.env.FFMPEG || "ffmpeg";
const montageOnly = process.argv.includes("--montage-only");
const montageViewport = { width: 640, height: 360, deviceScaleFactor: 1.5 };
const wait = (milliseconds) => new Promise((resolveWait) => setTimeout(resolveWait, milliseconds));

await mkdir(outputDirectory, { recursive: true });

function run(command, args) {
  const result = spawnSync(command, args, { encoding: "utf8" });
  if (result.status !== 0) {
    throw new Error(`${command} failed: ${result.stderr || result.stdout}`);
  }
}

async function dismissStorageNotice(page) {
  await page.evaluate(() => {
    const button = [...document.querySelectorAll("button")].find((candidate) =>
      candidate.textContent.includes("Use without storage"),
    );
    button?.click();
  });
  await wait(180);
}

async function center(page, selector) {
  await page.$eval(selector, (element) => element.scrollIntoView({ block: "center", inline: "center" }));
  await wait(220);
}

async function clickButton(page, label) {
  await page.evaluate((wanted) => {
    const button = [...document.querySelectorAll("button")].find((candidate) =>
      (candidate.getAttribute("aria-label") || candidate.textContent).trim().includes(wanted),
    );
    button?.click();
  }, label);
}

async function assertSidebarHidden(page) {
  const state = await page.evaluate(() => {
    const sidebar = document.querySelector(".component-doc-shell__sidebar");
    if (!sidebar) return { hidden: true, width: 0 };
    const rectangle = sidebar.getBoundingClientRect();
    return {
      hidden: rectangle.right <= 0 || rectangle.left >= innerWidth || getComputedStyle(sidebar).display === "none",
      width: rectangle.width,
    };
  });
  if (!state.hidden) throw new Error(`component docs sidebar remains visible at ${montageViewport.width}px (${state.width}px wide)`);
}

async function recordClip(browser, name, url, prepare, perform) {
  const page = await browser.newPage();
  await page.setViewport(montageViewport);
  await page.emulateMediaFeatures([
    { name: "prefers-color-scheme", value: "dark" },
    { name: "prefers-reduced-motion", value: "no-preference" },
  ]);
  await page.goto(url, { waitUntil: "networkidle2", timeout: 30_000 });
  await dismissStorageNotice(page);
  await assertSidebarHidden(page);
  await prepare(page);

  const path = join(temporaryDirectory, `${name}.webm`);
  const recorder = await page.screencast({
    path,
    ffmpegPath: ffmpeg,
    fps: 24,
    quality: 24,
  });
  await perform(page);
  await wait(400);
  await recorder.stop();
  await page.close();
  return path;
}

const browser = await puppeteer.launch({
  headless: true,
  args: ["--enable-webgl", "--ignore-gpu-blocklist"],
});

try {
  const clips = [];
  clips.push(await recordClip(
    browser,
    "tabs",
    "http://127.0.0.1:8090/components/tabs",
    (page) => center(page, "[role='tablist']"),
    async (page) => {
      await wait(260);
      await clickButton(page, "Likes");
      await wait(430);
      await clickButton(page, "Comments");
      await wait(430);
      await clickButton(page, "Saved");
      await wait(260);
    },
  ));
  clips.push(await recordClip(
    browser,
    "todo",
    "http://127.0.0.1:8090/examples/todo",
    (page) => center(page, "input[name='title']"),
    async (page) => {
      const input = await page.$("input[name='title']");
      await input.type("Ship storm UI", { delay: 38 });
      await wait(180);
      await clickButton(page, "Add task");
      await wait(620);
    },
  ));
  clips.push(await recordClip(
    browser,
    "monitor",
    "http://127.0.0.1:8091/components/line",
    (page) => center(page, "figure[aria-label='HTTPS monitor latency in milliseconds']"),
    async (page) => {
      await wait(350);
      await clickButton(page, "Enter fullscreen for HTTPS monitor latency");
      await wait(720);
      await clickButton(page, "Close");
      await wait(300);
    },
  ));
  clips.push(await recordClip(
    browser,
    "line3d",
    "http://127.0.0.1:8091/components/interactive/line-3d",
    async (page) => {
      await page.waitForFunction(() => document.querySelector("figure[aria-label='auto rotating'] canvas"));
      await center(page, "figure[aria-label='auto rotating']");
    },
    async () => {
      await wait(1_650);
    },
  ));

  const montage = join(outputDirectory, "goshtoso-components-montage-v1.mp4");
  const inputs = clips.flatMap((clip) => ["-i", clip]);
  const filters = clips.map((_, index) => `[${index}:v]trim=duration=1.65,setpts=PTS-STARTPTS,fps=20,scale=960:540:force_original_aspect_ratio=increase,crop=960:540[v${index}]`).join(";");
  const concatenation = clips.map((_, index) => `[v${index}]`).join("");
  run(ffmpeg, [
    "-y", ...inputs,
    "-filter_complex", `${filters};${concatenation}concat=n=${clips.length}:v=1:a=0[out]`,
    "-map", "[out]", "-an", "-c:v", "libx264", "-preset", "slow", "-crf", "29",
    "-pix_fmt", "yuv420p", "-movflags", "+faststart", montage,
  ]);
  run(ffmpeg, [
    "-y", "-ss", "0.45", "-i", montage, "-frames:v", "1", "-c:v", "libwebp", "-quality", "72",
    join(outputDirectory, "goshtoso-components-poster-v1.webp"),
  ]);

  const result = JSON.parse(spawnSync("ffprobe", [
    "-v", "error", "-show_entries", "format=duration,size", "-of", "json", montage,
  ], { encoding: "utf8" }).stdout);
  const supportingOutputs = [
    "goshtoso-components-poster-v1.webp",
  ];
  console.log(JSON.stringify({ montage: result.format, outputs: await Promise.all(supportingOutputs.map(async (name) => ({ name, bytes: (await readFile(join(outputDirectory, name))).length }))) }, null, 2));
} finally {
  await browser.close();
  await rm(temporaryDirectory, { recursive: true, force: true });
}
