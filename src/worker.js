const locales = new Set(["en", "pt-br", "es"]);
const assetRevision = "a8a9647a";

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

export default {
  async fetch(request, env) {
    const url = new URL(request.url);
    const first = url.pathname.split("/")[1].toLowerCase();

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
