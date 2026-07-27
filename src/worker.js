const locales = new Set(["en", "pt-br", "es"]);
const assetRevision = "archive-v5";

function localizedPage(locale, requestURL) {
  return new URL(`/${locale}/index.html?rev=${assetRevision}`, requestURL);
}

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const first = url.pathname.split("/")[1].toLowerCase();

    if (url.pathname === "/" || url.pathname === "/en" || url.pathname === "/en/") {
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
