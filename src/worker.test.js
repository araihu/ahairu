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
        fetched.push(new URL(request.url || request));
        return new Response("ok");
      },
    },
  };

  await worker.fetch(
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

  assert.equal(fetched[0].pathname, "/es/index.html");
  assert.equal(fetched[1].pathname, "/en/index.html");
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
  assert.match(html, /id="goshtoso-charts-version-slot"[^>]*hx-swap-oob="outerHTML"[^>]*>.*v0\.0\.1/s);
  assert.deepEqual(githubRequests, [
    "/repos/araihu/goshtoso/releases/latest",
    "/repos/araihu/goshtoso-charts/releases/latest",
    "/repos/araihu/goshtoso-charts/tags",
  ]);
  await Promise.all(pending);

  const second = await worker.fetch(request, env, ctx);
  assert.match(await second.text(), /v0\.1\.7/);
  assert.equal(githubRequests.length, 3, "cache hit must not call GitHub again");
});
