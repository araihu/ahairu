const locales = new Set(["en", "pt-br", "es"]);

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const first = url.pathname.split("/")[1].toLowerCase();

    if (url.pathname === "/" || url.pathname === "/en" || url.pathname === "/en/") {
      return env.ASSETS.fetch(new URL("/en/index.html", request.url));
    }
    if (locales.has(first) && (url.pathname === `/${first}` || url.pathname === `/${first}/`)) {
      return env.ASSETS.fetch(new URL(`/${first}/index.html`, request.url));
    }
    if (!url.pathname.includes(".")) {
      return env.ASSETS.fetch(new URL("/en/index.html", request.url));
    }
    return env.ASSETS.fetch(request);
  },
};
