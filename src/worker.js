const locales = new Set(["en", "pt-br", "es"]);
const versionCacheSeconds = 24 * 60 * 60;
const versionCacheRevision = "2";
const versionRepositories = [
  { repo: "goshtoso", slot: "goshtoso-version-slot", outOfBand: false },
  { repo: "goshtoso-charts", slot: "goshtoso-charts-version-slot", outOfBand: true },
];
const canonicalPages = new Map([
  ["/brand/", "/brand/index.html"],
  ["/license/", "/license/index.html"],
  ["/pt-br/brand/", "/pt-br/brand/index.html"],
  ["/pt-br/license/", "/pt-br/license/index.html"],
  ["/es/brand/", "/es/brand/index.html"],
  ["/es/license/", "/es/license/index.html"],
]);
const releaseChannels = new Map([
  ["/assets/releases/latest", "/assets/releases/latest.json"],
  ["/assets/releases/default", "/assets/releases/default.json"],
  ["/assets/releases/current", "/assets/releases/current.json"],
]);
const immutableReleasePath = /^\/assets\/releases\/v(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)(?:-(?:(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9A-Za-z-]*[A-Za-z-][0-9A-Za-z-]*))*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?\//;

function assetRequest(request, pathname) {
  const url = new URL(request.url);
  url.pathname = pathname;
  return new Request(url, request);
}

function redirect(requestURL, pathname) {
  const url = new URL(requestURL);
  return new Response(null, {
    status: 308,
    headers: { Location: `${pathname}${url.search}` },
  });
}

function withVary(response, value) {
  const headers = new Headers(response.headers);
  const tokens = (headers.get("Vary") || "")
    .split(",")
    .map((entry) => entry.trim())
    .filter(Boolean);
  const vary = [];
  const seen = new Set();

  if (tokens.includes("*")) {
    vary.push("*");
  } else {
    for (const token of tokens) {
      const normalized = token.toLowerCase();
      if (!seen.has(normalized)) {
        seen.add(normalized);
        vary.push(token);
      }
    }
  }

  if (vary[0] !== "*" && !seen.has(value.toLowerCase())) {
    vary.push(value);
  }
  headers.set("Vary", vary.join(", "));

  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers,
  });
}

function withAssetHeaders(response, pathname) {
  const headers = new Headers(response.headers);
  if (releaseChannels.has(pathname)) {
    headers.set("Content-Type", "application/json; charset=utf-8");
    headers.set("Cache-Control", "public, max-age=60, must-revalidate");
  } else if (immutableReleasePath.test(pathname)) {
    headers.set("Cache-Control", "public, max-age=31536000, immutable");
  }
  if (pathname.startsWith("/assets/releases/") || pathname.startsWith("/assets/campaign/")) {
    headers.set("Access-Control-Allow-Origin", "*");
    headers.delete("Access-Control-Allow-Credentials");
    const vary = (headers.get("Vary") || "")
      .split(",")
      .map((entry) => entry.trim())
      .filter((entry) => entry && entry.toLowerCase() !== "origin");
    if (vary.length === 0) {
      headers.delete("Vary");
    } else {
      headers.set("Vary", vary.join(", "));
    }
    headers.set("Cross-Origin-Resource-Policy", "cross-origin");
  }
  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers,
  });
}

function githubRequest(path) {
  return new Request(`https://api.github.com${path}`, {
    headers: {
      Accept: "application/vnd.github+json",
      "User-Agent": "araihu.com-project-versions",
      "X-GitHub-Api-Version": "2026-03-10",
    },
  });
}

async function latestVersion(repo, fetchImpl) {
  const release = await fetchImpl(githubRequest(`/repos/araihu/${repo}/releases/latest`));
  if (release.ok) {
    const data = await release.json();
    if (data.tag_name && data.html_url) return { version: data.tag_name, url: data.html_url };
  } else if (release.status !== 404) {
    throw new Error(`GitHub release request failed with ${release.status}`);
  }
  const tags = await fetchImpl(githubRequest(`/repos/araihu/${repo}/tags?per_page=1`));
  if (!tags.ok) throw new Error(`GitHub tag request failed with ${tags.status}`);
  const [latest] = await tags.json();
  if (!latest?.name) throw new Error(`GitHub repository ${repo} has no release or tag`);
  return { version: latest.name, url: `https://github.com/araihu/${repo}/tree/${encodeURIComponent(latest.name)}` };
}

function escapeHTML(value) {
  return String(value).replaceAll("&", "&amp;").replaceAll('"', "&quot;").replaceAll("<", "&lt;").replaceAll(">", "&gt;");
}

function versionBadge(project, result) {
  const outOfBand = project.outOfBand ? ' hx-swap-oob="outerHTML"' : "";
  if (!result) return `<span id="${project.slot}" class="project-version"${outOfBand}></span>`;
  return `<span id="${project.slot}" class="project-version"${outOfBand}><a href="${escapeHTML(result.url)}" rel="noopener noreferrer">${escapeHTML(result.version)}</a></span>`;
}

async function projectVersions(request, env, ctx) {
  const cache = env.VERSION_CACHE || globalThis.caches?.default;
  const cacheURL = new URL("/api/project-versions", request.url);
  cacheURL.searchParams.set("rev", versionCacheRevision);
  const cacheKey = new Request(cacheURL.toString());
  const cached = await cache?.match(cacheKey);
  if (cached) return cached;

  const fetchImpl = env.GITHUB_FETCH || fetch;
  const versions = await Promise.all(versionRepositories.map(async (project) => {
    try {
      return await latestVersion(project.repo, fetchImpl);
    } catch (error) {
      console.warn(`project version lookup failed for ${project.repo}`, error);
      return null;
    }
  }));
  const response = new Response(versionRepositories.map((project, index) => versionBadge(project, versions[index])).join(""), {
    headers: { "Content-Type": "text/html; charset=utf-8", "Cache-Control": `public, max-age=0, s-maxage=${versionCacheSeconds}` },
  });
  if (cache && versions.some(Boolean)) ctx?.waitUntil(cache.put(cacheKey, response.clone()));
  return response;
}

export function preferredLocale(acceptLanguage) {
  const preferences = (acceptLanguage || "")
    .split(",")
    .map((entry, order) => {
      const [rawTag, ...parameters] = entry.trim().toLowerCase().split(";");
      let quality = 1;

      for (const parameter of parameters) {
        const match = parameter.trim().match(/^q=(0(?:\.\d{0,3})?|1(?:\.0{0,3})?)$/);
        if (match) {
          quality = Number(match[1]);
        }
      }

      return { tag: rawTag, quality, order };
    })
    .filter(({ tag, quality }) => tag && tag !== "*" && quality > 0)
    .sort((left, right) => right.quality - left.quality || left.order - right.order);

  for (const { tag } of preferences) {
    if (locales.has(tag)) {
      return tag;
    }

    const base = tag.split("-")[0];
    if (base === "pt") {
      return "pt-br";
    }
    if (locales.has(base)) {
      return base;
    }
  }

  return "en";
}

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const first = url.pathname.split("/")[1].toLowerCase();

    if (url.pathname === "/api/project-versions" && request.method === "GET") {
      return projectVersions(request, env, ctx);
    }
    if (url.pathname === "/") {
      const locale = preferredLocale(request.headers.get("Accept-Language"));
      const response = await env.ASSETS.fetch(assetRequest(request, `/${locale}/index.html`));
      return withVary(response, "Accept-Language");
    }
    if (url.pathname === "/en" || url.pathname === "/en/") {
      return env.ASSETS.fetch(assetRequest(request, "/en/index.html"));
    }
    if (locales.has(first) && (url.pathname === `/${first}` || url.pathname === `/${first}/`)) {
      return env.ASSETS.fetch(assetRequest(request, `/${first}/index.html`));
    }

    if (url.pathname === "/en/brand" || url.pathname === "/en/brand/") {
      return redirect(request.url, "/brand/");
    }
    if (url.pathname === "/en/license" || url.pathname === "/en/license/") {
      return redirect(request.url, "/license/");
    }

    const staticPage = canonicalPages.get(url.pathname);
    if (staticPage) {
      return env.ASSETS.fetch(assetRequest(request, staticPage));
    }
    if (canonicalPages.has(`${url.pathname}/`)) {
      return redirect(request.url, `${url.pathname}/`);
    }

    const releaseChannel = releaseChannels.get(url.pathname);
    if (releaseChannel && (request.method === "GET" || request.method === "HEAD")) {
      const response = await env.ASSETS.fetch(assetRequest(request, releaseChannel));
      return withAssetHeaders(response, url.pathname);
    }

    if (url.pathname.startsWith("/fragments/")) {
      return env.ASSETS.fetch(request);
    }

    if (!url.pathname.includes(".")) {
      return new Response(request.method === "HEAD" ? null : "Not found", { status: 404 });
    }
    const response = await env.ASSETS.fetch(request);
    if (request.method === "GET" || request.method === "HEAD") {
      return withAssetHeaders(response, url.pathname);
    }
    return response;
  },
};
