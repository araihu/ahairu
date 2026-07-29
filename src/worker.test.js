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
