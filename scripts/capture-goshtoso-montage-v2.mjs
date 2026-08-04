import { mkdtemp, mkdir, readFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";
import { spawnSync } from "node:child_process";

const puppeteerModule = process.env.PUPPETEER_MODULE || "puppeteer";
const { default: puppeteer } = await import(puppeteerModule);

const outputDirectory = resolve(
  process.env.MONTAGE_OUTPUT || join(import.meta.dirname, "..", "site", "visuals"),
);
const temporaryDirectory = await mkdtemp(join(tmpdir(), "goshtoso-montage-v2-"));
const ffmpeg = process.env.FFMPEG || "ffmpeg";
const ffprobe = process.env.FFPROBE || "ffprobe";
const goshtosoBase = process.env.GOSHTOSO_BASE || "http://127.0.0.1:8090";
const chartsBase = process.env.CHARTS_BASE || "http://127.0.0.1:8091";
const viewport = { width: 1280, height: 720, deviceScaleFactor: 1 };
const clipDuration = 1.75;
const loopBridgeDuration = 0.4;
const loopTransitionDuration = 0.3;
const wait = (milliseconds) => new Promise((resolveWait) => setTimeout(resolveWait, milliseconds));

await mkdir(outputDirectory, { recursive: true });

function run(command, args) {
  const result = spawnSync(command, args, { encoding: "utf8" });
  if (result.status !== 0) {
    throw new Error(`${command} failed: ${result.stderr || result.stdout}`);
  }
  return result.stdout;
}

async function dismissStorageNotice(page) {
  await page.evaluate(() => {
    const button = [...document.querySelectorAll("button")].find((candidate) =>
      candidate.textContent.includes("Use without storage"),
    );
    button?.click();
  });
  await wait(120);
}

async function stage(page, selector, options = {}) {
  await page.waitForSelector(selector, { visible: true, timeout: 30_000 });
  await page.$eval(
    selector,
    (matched, config) => {
      const content = config.closest ? matched.closest(config.closest) : matched;
      if (!content) throw new Error(`No stage content for ${config.closest}`);

      document.documentElement.classList.add("dark");
      document.documentElement.style.colorScheme = "dark";
      document.documentElement.style.background = "#080d16";
      document.body.style.margin = "0";
      document.body.style.overflow = "hidden";
      document.body.style.background = "#080d16";

      const shell = document.createElement("main");
      shell.dataset.recordingStage = "true";
      shell.setAttribute("aria-label", config.label);

      const label = document.createElement("div");
      label.className = "recording-label";
      label.textContent = config.label;

      const surface = document.createElement("section");
      surface.className = `recording-surface recording-surface--${config.kind}`;
      surface.append(content);
      shell.append(label, surface);
      document.body.replaceChildren(shell);

      const style = document.createElement("style");
      style.textContent = `
        * { box-sizing: border-box; }
        [data-recording-stage] {
          position: fixed;
          inset: 0;
          display: grid;
          place-items: center;
          padding: 76px 72px 58px;
          overflow: hidden;
          background:
            radial-gradient(circle at 15% 18%, rgba(167, 243, 45, .08), transparent 31%),
            radial-gradient(circle at 84% 78%, rgba(96, 165, 250, .09), transparent 34%),
            #080d16;
          color: #f3f4f6;
          font-family: ui-sans-serif, system-ui, sans-serif;
        }
        .recording-label {
          position: absolute;
          top: 28px;
          left: 40px;
          color: #bef264;
          font-size: 18px;
          font-weight: 800;
          letter-spacing: .16em;
          text-transform: uppercase;
        }
        .recording-surface {
          width: min(1080px, 100%);
          max-height: 100%;
          margin: 0;
          padding: 34px;
          overflow: hidden;
          border: 1px solid rgba(148, 163, 184, .32);
          border-radius: 18px;
          background: rgba(13, 21, 35, .96);
          box-shadow: 0 28px 80px rgba(0, 0, 0, .42);
        }
        .recording-surface > * {
          width: 100% !important;
          max-width: none !important;
          margin: 0 !important;
        }
        .recording-surface--tabs {
          width: min(980px, 100%);
          padding: 62px 56px;
        }
        .recording-surface--todo {
          width: min(1020px, 100%);
          padding: 24px;
        }
        .recording-surface--todo #todo-fragment > header { display: none !important; }
        .recording-surface--chart {
          width: min(1120px, 100%);
          height: 570px;
          padding: 18px 24px 12px;
        }
        .recording-surface--chart .goshtoso-charts-control-wrapper,
        .recording-surface--chart .goshtoso-charts-control-content,
        .recording-surface--chart figure {
          width: 100% !important;
          max-width: none !important;
          height: 100% !important;
          margin: 0 !important;
        }
        .recording-surface--chart .goshtoso-charts-hidden-expand-modal {
          display: none !important;
        }
        .recording-surface--chart .goshtoso-charts-controls > [data-action-group-primary] {
          display: none !important;
        }
        .recording-surface--chart figure > div,
        .recording-surface--chart canvas {
          max-width: 100% !important;
        }
        @media (prefers-reduced-motion: reduce) {
          *, *::before, *::after {
            animation-duration: .001ms !important;
            animation-iteration-count: 1 !important;
            transition-duration: .001ms !important;
          }
        }
      `;
      document.head.append(style);
      window.dispatchEvent(new Event("resize"));
    },
    options,
  );
  await wait(420);
}

async function clickUnique(page, selector) {
  const elements = await page.$$(selector);
  if (elements.length !== 1) {
    throw new Error(`${selector} matched ${elements.length} elements`);
  }
  await elements[0].click();
}

async function recordClip(browser, { name, url, selector, stageOptions, prepare, perform }) {
  const page = await browser.newPage();
  await page.setViewport(viewport);
  await page.emulateMediaFeatures([
    { name: "prefers-color-scheme", value: "dark" },
    { name: "prefers-reduced-motion", value: "no-preference" },
  ]);
  await page.goto(url, { waitUntil: "networkidle2", timeout: 30_000 });
  await dismissStorageNotice(page);
  if (prepare) await prepare(page);
  await stage(page, selector, stageOptions);

  const path = join(temporaryDirectory, `${name}.webm`);
  const recorder = await page.screencast({
    path,
    ffmpegPath: ffmpeg,
    fps: 30,
    quality: 18,
  });
  await perform(page);
  await recorder.stop();

  const errors = await page.evaluate(() => window.__recordingErrors || []);
  if (errors.length > 0) throw new Error(`${name} browser errors: ${errors.join(" | ")}`);
  await page.close();
  return path;
}

const browser = await puppeteer.launch({
  headless: true,
  args: ["--enable-webgl", "--ignore-gpu-blocklist", "--use-angle=swiftshader"],
});

try {
  const clips = [];

  clips.push(await recordClip(browser, {
    name: "tabs-htmx",
    url: `${goshtosoBase}/components/tabs`,
    selector: "#tabs-htmx",
    stageOptions: { label: "Tabs · HTMX", kind: "tabs" },
    prepare: async (page) => {
      await page.evaluate(() => {
        window.__recordingErrors = [];
        addEventListener("error", (event) => window.__recordingErrors.push(event.message));
        addEventListener("unhandledrejection", (event) => window.__recordingErrors.push(String(event.reason)));
      });
    },
    perform: async (page) => {
      await wait(240);
      await clickUnique(page, "#tabs-htmx [role='tab'][aria-controls='tabpanelhtmxdetails']");
      await page.waitForFunction(() => document.querySelector("#tabs-htmx")?.textContent.includes("Details (Lazy Loaded)"));
      await wait(430);
      await clickUnique(page, "#tabs-htmx [role='tab'][aria-controls='tabpanelhtmxactivity']");
      await wait(680);
    },
  }));

  clips.push(await recordClip(browser, {
    name: "todo-htmx",
    url: `${goshtosoBase}/examples/todo`,
    selector: "#todo-fragment",
    stageOptions: { label: "Components · Real HTMX", kind: "todo" },
    prepare: async (page) => {
      await page.evaluate(() => {
        window.__recordingErrors = [];
        addEventListener("error", (event) => window.__recordingErrors.push(event.message));
        addEventListener("unhandledrejection", (event) => window.__recordingErrors.push(String(event.reason)));
      });
    },
    perform: async (page) => {
      const input = await page.$("input[name='title']");
      await input.type("Ship storm UI", { delay: 34 });
      await wait(140);
      await clickUnique(page, "#todo-fragment button[type='submit']");
      await page.waitForFunction(() => document.querySelector("#todo-list")?.textContent.includes("Ship storm UI"));
      await wait(560);
    },
  }));

  clips.push(await recordClip(browser, {
    name: "monitoring",
    url: `${chartsBase}/examples/live-availability`,
    selector: "figure[aria-label='Live availability from server-sent events']",
    stageOptions: { label: "Charts · Monitoring", kind: "chart", closest: ".goshtoso-charts-control-wrapper" },
    prepare: async (page) => {
      await page.evaluate(() => {
        window.__recordingErrors = [];
        addEventListener("error", (event) => window.__recordingErrors.push(event.message));
        addEventListener("unhandledrejection", (event) => window.__recordingErrors.push(String(event.reason)));
      });
    },
    perform: async (page) => {
      await wait(1_640);
    },
  }));

  clips.push(await recordClip(browser, {
    name: "line3d",
    url: `${chartsBase}/components/interactive/line-3d`,
    selector: "figure[aria-label='auto rotating']",
    stageOptions: { label: "Charts · Line 3D", kind: "chart", closest: ".goshtoso-charts-control-wrapper" },
    prepare: async (page) => {
      await page.evaluate(() => {
        window.__recordingErrors = [];
        addEventListener("error", (event) => window.__recordingErrors.push(event.message));
        addEventListener("unhandledrejection", (event) => window.__recordingErrors.push(String(event.reason)));
      });
      await page.waitForFunction(() => document.querySelector("figure[aria-label='auto rotating'] canvas"));
    },
    perform: async (page) => {
      await wait(1_640);
    },
  }));

  const hardCutMaster = join(temporaryDirectory, "goshtoso-family-montage-v2-hardcut-master.mkv");
  const master = join(temporaryDirectory, "goshtoso-family-montage-v2-seamless-master.mkv");
  const inputs = clips.flatMap((clip) => ["-i", clip]);
  const filters = clips
    .map((_, index) =>
      `[${index}:v]trim=duration=${clipDuration},setpts=PTS-STARTPTS,fps=30,scale=1280:720:force_original_aspect_ratio=increase,crop=1280:720,format=yuv420p[v${index}]`,
    )
    .join(";");
  const concatenation = clips.map((_, index) => `[v${index}]`).join("");
  run(ffmpeg, [
    "-y",
    ...inputs,
    "-filter_complex",
    `${filters};${concatenation}concat=n=${clips.length}:v=1:a=0[out]`,
    "-map",
    "[out]",
    "-an",
    "-c:v",
    "ffv1",
    "-level",
    "3",
    hardCutMaster,
  ]);

  const hardCutDuration = Number.parseFloat(
    run(ffprobe, [
      "-v",
      "error",
      "-show_entries",
      "format=duration",
      "-of",
      "default=noprint_wrappers=1:nokey=1",
      hardCutMaster,
    ]).trim(),
  );
  const bridgeStart = (hardCutDuration - loopBridgeDuration).toFixed(6);
  const transitionOffset = (loopBridgeDuration - loopTransitionDuration).toFixed(6);
  run(ffmpeg, [
    "-y",
    "-i",
    hardCutMaster,
    "-filter_complex",
    `[0:v]fps=30,format=yuv420p,split=3[body0][tail0][head0];` +
      `[body0]trim=start=${loopBridgeDuration}:end=${bridgeStart},setpts=PTS-STARTPTS[body];` +
      `[tail0]trim=start=${bridgeStart},setpts=PTS-STARTPTS[tail];` +
      `[head0]trim=duration=${loopBridgeDuration},setpts=PTS-STARTPTS[head];` +
      `[tail][head]xfade=transition=fade:duration=${loopTransitionDuration}:offset=${transitionOffset}[bridge];` +
      `[body][bridge]concat=n=2:v=1:a=0[out]`,
    "-map",
    "[out]",
    "-an",
    "-c:v",
    "ffv1",
    "-level",
    "3",
    master,
  ]);
  await rm(hardCutMaster);

  const poster = join(outputDirectory, "goshtoso-components-poster-v2.webp");
  run(ffmpeg, [
    "-y",
    "-ss",
    "0.95",
    "-i",
    master,
    "-frames:v",
    "1",
    "-c:v",
    "libwebp",
    "-quality",
    "80",
    poster,
  ]);

  const h264 = join(outputDirectory, "goshtoso-components-montage-v2.mp4");
  run(ffmpeg, [
    "-y", "-i", master, "-an", "-vf", "fps=24",
    "-c:v", "libx264", "-preset", "slow", "-tune", "animation", "-crf", "28",
    "-profile:v", "high", "-level:v", "4.0", "-pix_fmt", "yuv420p",
    "-movflags", "+faststart", "-g", "120", "-keyint_min", "120", "-sc_threshold", "0",
    h264,
  ]);

  const av1 = join(outputDirectory, "goshtoso-components-montage-v2-av1.mp4");
  run(ffmpeg, [
    "-y", "-i", master, "-an", "-vf", "fps=24",
    "-c:v", "libsvtav1", "-preset", "6", "-crf", "44",
    "-svtav1-params", "tune=0", "-pix_fmt", "yuv420p", "-g", "120",
    av1,
  ]);

  const clipSizes = await Promise.all(
    clips.map(async (clip) => ({ name: clip.split("/").at(-1), bytes: (await readFile(clip)).length })),
  );
  const outputs = await Promise.all(
    [h264, av1, poster].map(async (path) => ({ path, bytes: (await readFile(path)).length })),
  );
  console.log(JSON.stringify({ outputs, clipSizes }, null, 2));
} finally {
  await browser.close();
  if (process.env.KEEP_RECORDING_CLIPS !== "1") {
    await rm(temporaryDirectory, { recursive: true, force: true });
  } else {
    console.log(`clips: ${temporaryDirectory}`);
  }
}
