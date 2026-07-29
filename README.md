# Arai Hû

Static multilingual Arai Hû organization, brand, and license site, built with templ, Goshtoso, and Cloudflare Workers static assets. English is the fallback locale.

The public route set is intentionally small and canonical:

- Organization: `/en/`, `/pt-br/`, `/es/`
- Brand guidance: `/brand/`, `/pt-br/brand/`, `/es/brand/`
- License and usage: `/license/`, `/pt-br/license/`, `/es/license/`

English brand and license URLs omit `/en`. Their `/en/...` aliases and every no-slash brand/license alias return permanent `308` redirects with the query string intact. Unknown extensionless routes return `404`. The Worker owns HTML canonicalization, so the Cloudflare asset binding deliberately uses `html_handling: "none"`; enabling automatic trailing-slash handling creates a redirect loop around the Worker's explicit `index.html` mapping.

## Build and verification

The current pre-release compatibility build requires Go 1.26.5, templ 0.3.1020, Node 24+, and a temporary Go workspace containing this repository plus reviewed Goshtoso source at `91e7225452aa74243a3840412b9d00c91e8cac87`. No local `replace` or workspace file belongs in the repository; Task 8 replaces this bridge with the approved public module release.

```sh
export GOWORK=/path/to/temporary/go.work
go version
npm ci
templ generate
go run ./cmd/ahairu build
npm run test:metadata
npm run test:routes
npm run test:visual
go run ./cmd/checksite public
```

`npm run check` performs the deterministic build, complete Go suite, generated-output checker, Worker routes, headless-browser experience suite, and an undeployed Wrangler bundle dry-run. The browser suite invokes the real Worker against generated files and verifies all nine pages, reciprocal metadata and JSON-LD, downloads, mobile overflow, light/dark specimens, focus, reduced motion, geometry, redirects, and 404s.

The build writes localized standalone files, Goshtoso CSS, site CSS, two social previews, and the pinned Arai Hû v0.1.0 web/brand distribution to ignored `public/`. `src/worker.js` serves `en`, `pt-br`, and `es`. The root route selects the closest supported locale from `Accept-Language`; explicit locale routes stay fixed, and unsupported preferences fall back to English. Production deployment remains a separate approval checkpoint.

Social previews deliberately use two PageKind assets: home and brand pages share `social/brand.png`, while license pages use `social/license.png`. Each localized page retains its own canonical URL, title, description, and Open Graph locale.

`site/brand-assets` is an exact release subset, not a subtree. It contains only `catalog.json`, `checksums.txt`, `NOTICE`, `brand/**`, `icons/brand/**`, `platform/web/**`, and `licenses/**`. Build-time validation rejects additions, omissions, and checksum drift. Public files retain release paths below `/assets/araihu/v0.1.0/`.

Current provenance: assets release candidate `613335f60877d3cc6affd04c51a31ce7fa0e433c`, catalog SHA-256 `d83be964fa411e87c61b49f0a0b6a2a1465f33ad43bea7cd93b2e434b59266af`, checksums SHA-256 `2d83421b3a95c75f68c88af7d5618034b4189d42adf3f2e39b2c4c048c553d5d`, brand sprite SHA-256 `e0c98a783cf65cf52b0a57cca47b84704499200a7fdb113b751d8f6c5828ba45`, and reviewed Goshtoso compatibility source `91e7225452aa74243a3840412b9d00c91e8cac87`. Release tags replace candidate commit provenance only after final approval.
