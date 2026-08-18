import { existsSync } from "node:fs";

const homebrewChromium = "/opt/homebrew/bin/chromium";

export function browserLaunchOptions({
  env = process.env,
  pathExists = existsSync,
  platform = process.platform,
} = {}) {
  const args = ["--no-sandbox"];
  if (platform === "darwin") {
    args.push("--disable-features=MacAppCodeSignClone");
  }
  return {
    headless: true,
    executablePath: env.PUPPETEER_EXECUTABLE_PATH || (pathExists(homebrewChromium) ? homebrewChromium : undefined),
    args,
  };
}
