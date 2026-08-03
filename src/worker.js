const locales = new Set(["en", "pt-br", "es"]);
const assetRevision = "a8a9647a";
const versionCacheSeconds = 24 * 60 * 60;
const versionCacheRevision = "2";
const versionRepositories = [
  { repo: "goshtoso", slot: "goshtoso-version-slot", outOfBand: false },
  { repo: "goshtoso-charts", slot: "goshtoso-charts-version-slot", outOfBand: true },
];

function localizedPage(locale, requestURL) {
  return new URL(`/${locale}/index.html?rev=${assetRevision}`, requestURL);
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
  return {
    version: latest.name,
    url: `https://github.com/araihu/${repo}/tree/${encodeURIComponent(latest.name)}`,
  };
}

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll('"', "&quot;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;");
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
  const html = versionRepositories.map((project, index) => versionBadge(project, versions[index])).join("");
  const response = new Response(html, {
    headers: {
      "Content-Type": "text/html; charset=utf-8",
      "Cache-Control": `public, max-age=0, s-maxage=${versionCacheSeconds}`,
    },
  });
  if (cache && versions.some(Boolean)) ctx?.waitUntil(cache.put(cacheKey, response.clone()));
  return response;
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
      return env.ASSETS.fetch(localizedPage(locale, request.url));
    }
    if (url.pathname === "/en" || url.pathname === "/en/") {
      return env.ASSETS.fetch(localizedPage("en", request.url));
    }
    if (locales.has(first) && (url.pathname === `/${first}` || url.pathname === `/${first}/`)) {
      return env.ASSETS.fetch(localizedPage(first, request.url));
    }
    if (!url.pathname.includes(".")) {
      return env.ASSETS.fetch(localizedPage("en", request.url));
    }
    return env.ASSETS.fetch(request);
  },
};
