import assert from "node:assert/strict";
import test from "node:test";

import worker, { preferredLocale } from "./worker.js";

test("preferredLocale matches supported language and region tags", () => {
  assert.equal(preferredLocale("pt-BR,pt;q=0.9,en;q=0.8"), "pt-br");
  assert.equal(preferredLocale("pt-PT,en;q=0.8"), "pt-br");
  assert.equal(preferredLocale("es-MX,es;q=0.9,en;q=0.8"), "es");
  assert.equal(preferredLocale("en-US,en;q=0.9"), "en");
});

test("preferredLocale respects quality weights and exclusions", () => {
  assert.equal(preferredLocale("pt-BR;q=0.7,es;q=0.9"), "es");
  assert.equal(preferredLocale("pt-BR;q=0,es;q=0.8"), "es");
});

test("preferredLocale falls back to English", () => {
  assert.equal(preferredLocale("fr-FR,fr;q=0.9"), "en");
  assert.equal(preferredLocale("*"), "en");
  assert.equal(preferredLocale(""), "en");
});

test("root serves detected locale while explicit routes remain fixed", async () => {
  const fetched = [];
  const env = {
    ASSETS: {
      fetch(request) {
        fetched.push(new Request(request));
        return new Response("ok", { headers: { Vary: "Accept-Encoding" } });
      },
    },
  };

  const negotiated = await worker.fetch(
    new Request("https://araihu.com/", {
      headers: { "Accept-Language": "es-MX,es;q=0.9,en;q=0.8" },
    }),
    env,
  );
  await worker.fetch(
    new Request("https://araihu.com/en/", {
      headers: { "Accept-Language": "pt-BR" },
    }),
    env,
  );

  assert.equal(new URL(fetched[0].url).pathname, "/es/index.html");
  assert.equal(new URL(fetched[1].url).pathname, "/en/index.html");
  assert.equal(negotiated.headers.get("Vary"), "Accept-Encoding, Accept-Language");
});

test("locale negotiation normalizes existing Vary tokens", async () => {
  const cases = [
    ["accept-language", "accept-language"],
    ["Accept-Encoding, accept-language, ACCEPT-LANGUAGE", "Accept-Encoding, accept-language"],
    ["*", "*"],
    ["Accept-Encoding, *, accept-language", "*"],
  ];

  for (const [existing, expected] of cases) {
    const env = {
      ASSETS: {
        fetch() {
          return new Response("ok", { headers: { Vary: existing } });
        },
      },
    };
    const response = await worker.fetch(new Request("https://araihu.com/"), env);

    assert.equal(response.headers.get("Vary"), expected, existing);
  }
});

const canonicalPages = [
  ["/brand/", "/brand/index.html"],
  ["/license/", "/license/index.html"],
  ["/pt-br/brand/", "/pt-br/brand/index.html"],
  ["/pt-br/license/", "/pt-br/license/index.html"],
  ["/es/brand/", "/es/brand/index.html"],
  ["/es/license/", "/es/license/index.html"],
];

function recordingAssets() {
  const fetched = [];
  return {
    fetched,
    env: {
      ASSETS: {
        fetch(request) {
          const copy = new Request(request);
          fetched.push(copy);
          return new Response(null, { status: copy.method === "HEAD" ? 204 : 200 });
        },
      },
    },
  };
}

async function request(path, headers = {}) {
  let requested;
  const response = await worker.fetch(
    new Request(`https://araihu.com${path}`, { headers }),
    {
      ASSETS: {
        fetch(assetRequest) {
          requested = new URL(assetRequest.url);
          return new Response("asset", {
            headers: { Vary: "Accept-Encoding" },
          });
        },
      },
    },
  );
  return { response, requested };
}

test("maps the extensionless current channel", async () => {
  const { response, requested } = await request("/assets/releases/current");

  assert.equal(requested.pathname, "/assets/releases/current.json");
  assert.equal(response.headers.get("content-type"), "application/json; charset=utf-8");
  assert.equal(response.headers.get("cache-control"), "public, max-age=60, must-revalidate");
});

test("maps only declared extensionless release channels", async () => {
  for (const [channel, file] of [
    ["latest", "/assets/releases/latest.json"],
    ["default", "/assets/releases/default.json"],
  ]) {
    const { response, requested } = await request(`/assets/releases/${channel}`);

    assert.equal(requested.pathname, file, channel);
    assert.equal(response.status, 200, channel);
  }

  const { response, requested } = await request("/assets/releases/unknown");
  assert.equal(response.status, 404);
  assert.equal(requested, undefined);
});

test("marks strict SemVer versioned assets immutable", async () => {
  const { response } = await request("/assets/releases/v0.1.1-rc.1+build.42/catalog.json");

  assert.equal(response.headers.get("cache-control"), "public, max-age=31536000, immutable");
});

test("does not mark leading-zero SemVer identifiers immutable", async () => {
  for (const version of ["v01.1.1", "v0.1.1-01"]) {
    const { response } = await request(`/assets/releases/${version}/catalog.json`);
    assert.equal(response.headers.get("cache-control"), null, version);
  }
});

test("permits anonymous public asset consumption", async () => {
  const { response } = await request("/assets/releases/current", {
    Origin: "https://goshtoso.araihu.com",
  });

  assert.equal(response.headers.get("access-control-allow-origin"), "*");
  assert.equal(response.headers.get("access-control-allow-credentials"), null);
  assert.equal(response.headers.get("cross-origin-resource-policy"), "cross-origin");
  assert.equal(response.headers.get("Vary"), "Accept-Encoding");
});

test("removes credentialed CORS headers and Origin Vary from asset upstreams", async () => {
  const response = await worker.fetch(new Request("https://araihu.com/assets/releases/v0.1.1/catalog.json"), {
    ASSETS: {
      fetch() {
        return new Response("asset", {
          headers: {
            "Access-Control-Allow-Credentials": "true",
            Vary: "Accept-Encoding, Origin, Accept-Language",
          },
        });
      },
    },
  });

  assert.equal(response.headers.get("access-control-allow-origin"), "*");
  assert.equal(response.headers.get("access-control-allow-credentials"), null);
  assert.equal(response.headers.get("Vary"), "Accept-Encoding, Accept-Language");
});

test("preserves HEAD when mapping a release channel", async () => {
  const { fetched, env } = recordingAssets();
  await worker.fetch(new Request("https://araihu.com/assets/releases/current", { method: "HEAD" }), env);

  assert.equal(fetched.length, 1);
  assert.equal(fetched[0].method, "HEAD");
  assert.equal(new URL(fetched[0].url).pathname, "/assets/releases/current.json");
});

test("canonical brand pages map explicitly to static files", async () => {
  for (const [route, asset] of canonicalPages) {
    const { fetched, env } = recordingAssets();
    await worker.fetch(new Request(`https://araihu.com${route}?preview=1`), env);

    assert.equal(fetched.length, 1, route);
    const fetchedURL = new URL(fetched[0].url);
    assert.equal(fetchedURL.pathname, asset, route);
    assert.equal(fetchedURL.search, "?preview=1", route);
  }
});

test("no-slash brand page aliases redirect permanently and preserve raw query", async () => {
  for (const [route] of canonicalPages) {
    const alias = route.slice(0, -1);
    const { fetched, env } = recordingAssets();
    const response = await worker.fetch(
      new Request(`https://araihu.com${alias}?x=one%2Ftwo&x=three+four&empty=`),
      env,
    );

    assert.equal(response.status, 308, alias);
    assert.equal(
      response.headers.get("Location"),
      `${route}?x=one%2Ftwo&x=three+four&empty=`,
      alias,
    );
    assert.equal(fetched.length, 0, alias);
  }
});

test("English-prefixed brand aliases redirect to canonical English routes", async () => {
  for (const route of ["/en/brand", "/en/brand/", "/en/license", "/en/license/"]) {
    const { fetched, env } = recordingAssets();
    const response = await worker.fetch(new Request(`https://araihu.com${route}?ref=en%2Dus`), env);
    const page = route.includes("license") ? "license" : "brand";

    assert.equal(response.status, 308, route);
    assert.equal(response.headers.get("Location"), `/${page}/?ref=en%2Dus`, route);
    assert.equal(fetched.length, 0, route);
  }
});

test("unknown extensionless paths return 404 instead of the English home page", async () => {
  for (const path of ["/unknown", "/deep/unknown", "/pt-br/unknown"]) {
    const { fetched, env } = recordingAssets();
    const response = await worker.fetch(new Request(`https://araihu.com${path}`), env);

    assert.equal(response.status, 404, path);
    assert.equal(fetched.length, 0, path);
  }
});

test("asset-like paths pass through unchanged and retain GET or HEAD", async () => {
  for (const method of ["GET", "HEAD"]) {
    const { fetched, env } = recordingAssets();
    const request = new Request("https://araihu.com/assets/logo.svg?variant=dark", { method });
    const response = await worker.fetch(request, env);

    assert.equal(fetched.length, 1, method);
    assert.equal(fetched[0].url, request.url, method);
    assert.equal(fetched[0].method, method, method);
    assert.notEqual(response.status, 308, method);
  }
});

test("canonical static page mapping retains HEAD method", async () => {
  const { fetched, env } = recordingAssets();
  await worker.fetch(new Request("https://araihu.com/brand/?preview=1", { method: "HEAD" }), env);

  assert.equal(fetched.length, 1);
  assert.equal(fetched[0].method, "HEAD");
  assert.equal(new URL(fetched[0].url).pathname, "/brand/index.html");
});

test("extensionless chart fragment routes pass through to static assets", async () => {
  const fetched = [];
  const env = {
    ASSETS: {
      fetch(request) {
        const url = new URL(request.url || request);
        fetched.push(url);
        return new Response("chart fragment");
      },
    },
  };

  const response = await worker.fetch(
    new Request("https://araihu.com/fragments/pt-br/charts"),
    env,
  );

  assert.equal(await response.text(), "chart fragment");
  assert.equal(fetched[0].pathname, "/fragments/pt-br/charts");
});

test("project versions use latest releases, tag fallback, and a 24 hour edge cache", async () => {
  const githubRequests = [];
  const cacheEntries = new Map();
  const pending = [];
  const env = {
    ASSETS: { fetch: () => new Response("asset") },
    GITHUB_FETCH: async (request) => {
      const url = new URL(request.url || request);
      githubRequests.push(url.pathname);
      if (url.pathname === "/repos/araihu/goshtoso/releases/latest") {
        return Response.json({
          tag_name: "v0.1.7",
          html_url: "https://github.com/araihu/goshtoso/releases/tag/v0.1.7",
        });
      }
      if (url.pathname === "/repos/araihu/goshtoso-app-shells/releases/latest") {
        return Response.json({
          tag_name: "v0.1.3",
          html_url: "https://github.com/araihu/goshtoso-app-shells/releases/tag/v0.1.3",
        });
      }
      if (url.pathname === "/repos/araihu/goshtoso-charts/releases/latest") {
        return new Response("missing", { status: 404 });
      }
      if (url.pathname === "/repos/araihu/goshtoso-charts/tags") {
        return Response.json([{ name: "v0.0.1" }]);
      }
      return new Response("unexpected", { status: 500 });
    },
    VERSION_CACHE: {
      async match(request) {
        return cacheEntries.get(request.url)?.clone();
      },
      async put(request, response) {
        cacheEntries.set(request.url, response.clone());
      },
    },
  };
  const ctx = { waitUntil: (promise) => pending.push(promise) };
  const request = new Request("https://araihu.com/api/project-versions");

  const first = await worker.fetch(request, env, ctx);
  const html = await first.text();
  assert.equal(first.headers.get("cache-control"), "public, max-age=0, s-maxage=86400");
  assert.match(html, /id="goshtoso-version-slot"[^>]*>.*v0\.1\.7/s);
  assert.match(html, /id="goshtoso-app-shells-version-slot"[^>]*hx-swap-oob="outerHTML"[^>]*>.*v0\.1\.3/s);
  assert.match(html, /id="goshtoso-charts-version-slot"[^>]*hx-swap-oob="outerHTML"[^>]*>.*v0\.0\.1/s);
  assert.doesNotMatch(html, /<a\b/);
  assert.deepEqual(githubRequests, [
    "/repos/araihu/goshtoso/releases/latest",
    "/repos/araihu/goshtoso-app-shells/releases/latest",
    "/repos/araihu/goshtoso-charts/releases/latest",
    "/repos/araihu/goshtoso-charts/tags",
  ]);
  await Promise.all(pending);

  const second = await worker.fetch(request, env, ctx);
  assert.match(await second.text(), /v0\.1\.7/);
  assert.equal(githubRequests.length, 4, "cache hit must not call GitHub again");
});
