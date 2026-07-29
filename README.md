# Arai Hû

Static multilingual Arai Hû organization site, built with templ, Goshtoso, and Cloudflare Workers static assets. English is the fallback locale.

```sh
templ generate
go run ./cmd/ahairu build
```

The build writes localized standalone files, Goshtoso CSS, site CSS, and the pinned Arai Hû v0.1.0 web/brand distribution to `public/`. `src/worker.js` serves `en`, `pt-br`, and `es`. The root route selects the closest supported locale from `Accept-Language`; explicit locale routes stay fixed, and unsupported preferences fall back to English.

Social previews deliberately use two PageKind assets: home and brand pages share `social/brand.png`, while license pages use `social/license.png`. Each localized page retains its own canonical URL, title, description, and Open Graph locale.

`site/brand-assets` is an exact release subset, not a subtree. It contains only `catalog.json`, `checksums.txt`, `NOTICE`, `brand/**`, `icons/brand/**`, `platform/web/**`, and `licenses/**`. Build-time validation rejects additions, omissions, and checksum drift. Public files retain release paths below `/assets/araihu/v0.1.0/`.

Current provenance: assets release candidate `613335f60877d3cc6affd04c51a31ce7fa0e433c`, catalog SHA-256 `d83be964fa411e87c61b49f0a0b6a2a1465f33ad43bea7cd93b2e434b59266af`, checksums SHA-256 `2d83421b3a95c75f68c88af7d5618034b4189d42adf3f2e39b2c4c048c553d5d`, brand sprite SHA-256 `e0c98a783cf65cf52b0a57cca47b84704499200a7fdb113b751d8f6c5828ba45`, and reviewed Goshtoso generator `d8d58c355a21fc5d17edeb3ef0340a5a3b2d6854`. Release tags replace candidate commit provenance only after final approval.
