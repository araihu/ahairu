const locales = new Set(["en", "pt-br", "es"]);
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
  } else if (/^\/assets\/releases\/v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?\//.test(pathname)) {
    headers.set("Cache-Control", "public, max-age=31536000, immutable");
  }
  if (pathname.startsWith("/assets/releases/") || pathname.startsWith("/assets/campaign/")) {
    headers.set("Access-Control-Allow-Origin", "*");
    headers.set("Cross-Origin-Resource-Policy", "cross-origin");
  }
  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers,
  });
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
  async fetch(request, env) {
    const url = new URL(request.url);
    const first = url.pathname.split("/")[1].toLowerCase();

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
