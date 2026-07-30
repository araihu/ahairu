# Arai Hû

Static multilingual Arai Hû organization, brand, and license site, built with templ, Goshtoso, and Cloudflare Workers static assets. English is the fallback locale.

The public route set is intentionally small and canonical:

- Organization: `/en/`, `/pt-br/`, `/es/`
- Brand guidance: `/brand/`, `/pt-br/brand/`, `/es/brand/`
- License and usage: `/license/`, `/pt-br/license/`, `/es/license/`

English brand and license URLs omit `/en`. Their `/en/...` aliases and every no-slash brand/license alias return permanent `308` redirects with the query string intact. Unknown extensionless routes return `404`. The Worker owns HTML canonicalization, so the Cloudflare asset binding deliberately uses `html_handling: "none"`; enabling automatic trailing-slash handling creates a redirect loop around the Worker's explicit `index.html` mapping.

## Build and verification

The build requires Go 1.26.5, templ 0.3.1020, and Node 24+. It consumes the
released Goshtoso module directly; no local `replace` or workspace file belongs
in the repository.

```sh
go version
npm ci
templ generate
ASSET_BUNDLE=/absolute/path/to/verified-assets npm run build
npm run test:metadata
npm run test:routes
npm run test:visual
ASSET_BUNDLE=/absolute/path/to/verified-assets npm run check
```

`--asset-bundle` is required: the build never silently substitutes the embedded
brand subset for a deployable Assets bundle. `npm run check` performs the
deterministic build, complete Go suite, generated-output checker, Worker routes,
headless-browser experience suite, and an undeployed Wrangler bundle dry-run.
The browser suite invokes the real Worker against generated files and verifies
all nine pages, reciprocal metadata and JSON-LD, downloads, mobile overflow,
light/dark specimens, focus, reduced motion, geometry, redirects, and 404s.

The build writes localized standalone files, Goshtoso CSS, site CSS, two social
previews, and the verified Assets release/channel bundle to ignored `public/`.
`src/worker.js` serves `en`, `pt-br`, and `es`. The root route selects the
closest supported locale from `Accept-Language`; explicit locale routes stay
fixed, and unsupported preferences fall back to English. A successful `main` CI
run deploys the complete checked Worker version directly.

Social previews deliberately use two PageKind assets: home and brand pages share `social/brand.png`, while license pages use `social/license.png`. Each localized page retains its own canonical URL, title, description, and Open Graph locale.

`site/brand-assets` is an exact baseline release subset, not a subtree. The
deployable bundle retains every immutable release below
`/assets/releases/vMAJOR.MINOR.PATCH/`, plus `latest`, `default`, and `current`
channel documents. Build-time validation rejects additions, omissions, and
checksum drift. Brand downloads remain pinned to `/assets/releases/v0.1.1/`.

Current baseline provenance: Arai Hû Assets `v0.1.1`, Goshtoso `v0.1.1`.

## Assets release and delivery contract

Assets dispatches only `araihu-assets-released`. Its `client_payload` is one
exact JSON handoff: `assets_repository` (`araihu/assets`), 40-character lowercase
`assets_revision`, nonempty `release_artifacts`, `runtime_release`,
`channel_artifact_id`, `channel_artifact_url`, `channel_artifact_sha256`,
`candidate_bundle_digest`, `resolution_date`, `state_ref`, and `state_path`.
Every `release_artifacts` element has only `release`, `release_url`, and
`release_sha256`. Release tags are strict `v`-prefixed SemVer and each URL must
be exactly
`https://github.com/araihu/assets/releases/download/<tag>/araihu-assets-<tag>.tar.gz`.
`runtime_release` must occur in that array. The channel URL must be exactly the
matching `https://github.com/araihu/assets/actions/runs/<positive-run-id>/artifacts/<positive-id>`
HTML URL; Ahairu validates it, then constructs the trusted GitHub API ZIP URL
from its ID locally. The candidate digest is canonical `sha256:<lowercase-hex>`;
the downloadable release and channel archive hashes are verified before use.

Each release archive contains exactly its declared
`releases/vMAJOR.MINOR.PATCH/` root, including `release.json` and
`checksums.txt`. Ahairu extracts every declared release into one cumulative,
immutable bundle. The channel archive contains exactly `campaign/v1.js`,
`releases/latest.json`, `releases/default.json`, and `releases/current.json`.
Symlinks, special files, absolute paths, traversal, duplicate paths, and
unexpected roots are rejected before extraction.

For local parity, save the complete handoff JSON, then materialize a new bundle
before running the full gate. Set `ASSETS_GITHUB_TOKEN` only when the selected
Assets repository requires it; do not put it in shell history or source control.

```sh
python3 scripts/prepare_asset_bundle.py \
  --handoff-file /absolute/path/to/assets-handoff.json \
  --output .asset-bundle
ASSET_BUNDLE="$PWD/.asset-bundle" npm run check
```

CI mints a short-lived GitHub App token limited to `araihu/assets`, validates
the whole dispatch handoff, checks the complete bundle, then hands the exact
normalized JSON artifact to deploy. Deploy validates it again, reacquires and
hash-verifies every archive, then runs the same full check before `wrangler`
deploy in protected `production`. Dispatch never reads repository variables.
For a `main` push, CI may use only one documented repository variable,
`ASSETS_RELEASE_HANDOFF_JSON`, containing that same complete JSON object; it
never accepts the retired six flat variables. Cloudflare credentials are only
`CLOUDFLARE_API_TOKEN` and `CLOUDFLARE_ACCOUNT_ID` there; Assets never owns a
Cloudflare secret.

Release documents publish the immutable catalog, theme manifest and CSS, and
campaign manifest/runtime together. Consumers may cache versioned files for a
year; channel documents are short-lived selectors and must resolve only to an
included immutable release.

## Deferred shell adoption

This remains a static-first site. Goshtoso App Shells `v0.1.0` is deliberately
not a dependency while pages need no persistent application shell, client state,
or duplicate preference store. Re-evaluate only if an application surface is
introduced; campaign preference stays owned by the released Assets runtime.
