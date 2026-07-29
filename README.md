# Arai Hû

Static multilingual Arai Hû organization site, built with templ, Goshtoso, and Cloudflare Workers static assets. English is the fallback locale.

```sh
templ generate
go run ./cmd/ahairu build
```

The build writes localized standalone files, Goshtoso CSS, the Arai Hû theme, favicon, and project marks to `public/`. `src/worker.js` serves `en`, `pt-br`, and `es`. The root route selects the closest supported locale from `Accept-Language`; explicit locale routes stay fixed, and unsupported preferences fall back to English.

Brand assets live in the `site/brand-assets` git subtree, sourced from `araihu/assets`. Update it with authenticated GitHub access:

```sh
git subtree pull --prefix=site/brand-assets https://github.com/araihu/assets.git main --squash
```
