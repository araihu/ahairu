# Seasonal assets local canary

## Enabled campaign browser gate

The executable gate accepts only an absolute `ASSET_BUNDLE` containing at
least two checksum-verified immutable releases and a `current` channel resolved
to an enabled campaign. `CANARY_CAMPAIGN_CHECK_DATE` is required. It checks
that the effective campaign identity is active inside the selected release's
inclusive `campaigns.json` interval and that theme, toggle IDs, and tinted
brand IDs agree.

Channel schema v1 does not carry its resolver input date, and that date is not
covered by the channel digest. The check date is therefore consistency
evidence only. This harness does not claim it is the exact date used by Assets
to produce the accepted channel.

```sh
npm ci
ASSET_BUNDLE=/absolute/path/to/verified-assets \
  CANARY_CAMPAIGN_CHECK_DATE=2026-10-31 \
  CANARY_PORT=8788 \
  node scripts/seasonal_assets_canary.mjs \
  | tee /private/tmp/ahairu-enabled-seasonal-canary.ndjson
```

Optional expiry coverage takes a second default channel document for the same
retained release set and a date on which its campaign manifest has no enabled
campaign. The harness strictly recomputes this document's channel digest,
binds its theme to `themes.json` and the complete release inventory, and probes
that immutable theme directly before browser execution. Active theme, brand,
and toggle IDs are likewise bound to `themes.json`/`catalog.json`, their
inventory entries, manifest-declared render modes, UI namespace, and declared
sprite symbols. Sprite mode must use `icons/ui/sprite.svg`; asset mode must use
the catalog member's discrete `icons/ui/` path.

```sh
ASSET_BUNDLE=/absolute/path/to/verified-assets \
  CANARY_CAMPAIGN_CHECK_DATE=2026-10-31 \
  CANARY_EXPIRED_CHANNEL=/absolute/path/to/expired-current.json \
  CANARY_EXPIRED_CHECK_DATE=2026-11-01 \
  node scripts/seasonal_assets_canary.mjs
```

One local Wrangler process serves the whole run. Playwright keeps every
navigation and subresource URL at canonical `https://araihu.com`; request
interception proxies those requests to that Wrangler process. Requests to any
other HTTP origin are blocked and fail the canary. The expiry document is
copied into generated `public/` before Wrangler starts, then its canonical
`current` request is mapped to that local Wrangler route for the final refresh.
No remote canary adapter or relaxed runtime origin rule exists.

`CANARY_TIMEOUT_MS` (default 30000) bounds Wrangler readiness, the held runtime
request, every expected `current` request, and Playwright waits. Missing runtime
or channel fetches fail explicitly. Page close/crash, context close, intercepted
request failure, and proxy failure reject pending request gates immediately.

Browser scenarios:

- immutable SSR theme and local brand URLs captured while the deferred runtime
  script response is held, before campaign code executes or starts a refresh;
- first apply: exact theme, source, campaign ID, active toggle, tinted logo and
  icon, sparkles toggle, anonymous CORS, and active stylesheet;
- explicit preference before deferred runtime execution;
- opt-out, persisted reload, and re-enable with campaign-specific storage;
- reduced-motion lifecycle detail;
- expiry refresh and baseline restoration when both expiry inputs are present.

First navigation and opt-out reload wait for the runtime's automatic deferred
bootstrap. They never call `AraiHuCampaign.refresh()`. The only manual refresh
is the expiry scenario after interception switches `current` to the locally
served default document.

Before opening Chromium, direct local GET probes verify the resolved theme,
brand assets, and both toggle resources against the immutable `release.json`
inventory SHA-256 and byte size. The optional default theme receives the same
probe. NDJSON `channel-evidence`,
`resolved-asset-probe`, `browser-state`, and `canary-pass` records repeat the
campaign check date, release, source, effective campaign, and recomputed channel
digest. The check date is not presented as channel provenance.

```json
{"event":"channel-evidence","campaignCheckDate":"2026-10-31","release":"v0.1.2","source":"campaign","campaign":"halloween-2026","digest":"...","releases":["v0.1.1","v0.1.2"]}
{"event":"resolved-asset-probe","campaignCheckDate":"2026-10-31","release":"v0.1.2","source":"campaign","campaign":"halloween-2026","digest":"...","kind":"brand-logo","id":"araihu-logo-tinted-transparent-optical","url":"https://araihu.com/assets/releases/v0.1.2/...","status":200,"bytes":1234,"sha256":"..."}
{"event":"browser-state","scenario":"ssr-baseline","campaignCheckDate":"2026-10-31","release":"v0.1.2","source":"campaign","campaign":"halloween-2026","digest":"...","theme":"araihu","themeSource":"default","activeCampaign":null,"toggleHidden":true,"togglePressed":"false","reducedMotion":false}
{"event":"browser-state","scenario":"first-apply","campaignCheckDate":"2026-10-31","release":"v0.1.2","source":"campaign","campaign":"halloween-2026","digest":"...","theme":"araihu-halloween","themeSource":"campaign","activeCampaign":"halloween-2026","toggleHidden":false,"togglePressed":"true","reducedMotion":false}
{"event":"canary-pass","campaignCheckDate":"2026-10-31","release":"v0.1.2","source":"campaign","campaign":"halloween-2026","digest":"...","releases":["v0.1.1","v0.1.2"],"expiry":true}
```

Focused harness tests:

```sh
node --test scripts/seasonal_assets_canary.test.mjs
```

Current v0.1.1 evidence cannot satisfy this gate: it contains one immutable
release and resolves `current` to `source=default` with its campaign disabled.
Keep that fail-closed result until a real accepted v0.1.2 two-release bundle
exists. Do not synthesize or edit a release to claim enabled-campaign evidence.

## Recorded v0.1.1 baseline — 2026-07-30

## Scope and input

- Ahairu commit: `c8d0b9220f452ee67283bfec9a951e0a249a9359`
- Verified local bundle: `/private/tmp/ahairu-parity-assets.5JIiK9`
- Channel target: `v0.1.1`, digest `34193fd6171ed32cb3307124cfef8be7713503942864dc5a67623c174d4a6c4f`
- Runtime SHA-256: `a936193b4fed8120e6cb3423f19d3e2ddb0ba32266dc4e5f02a98f5261853709`

No production deployment, promotion, tag, or push occurred.

## Reproducible local HTTP baseline

The earlier executable baseline built only the supplied bundle, started
exactly one local Wrangler session, waited for `/assets/releases/current`, then
checks every documented GET and HEAD route. It derives expected GET bytes and
SHA-256 values from the just-built `public/` tree, so it rejects a stale route,
header, body, or release mapping rather than relying on this historical table.
It also requires every HEAD response body to be empty.

`ASSET_BUNDLE` must be an absolute path to the accepted, verified Assets
bundle. `CANARY_PORT` is an optional loopback port; `CANARY_TIMEOUT_MS` is an
optional conditional-ready timeout (default 30000 ms). The current executable
also requires the enabled campaign inputs documented above; this command is a
historical transcript recipe, not a current passing gate.

```sh
npm ci
ASSET_BUNDLE=/absolute/path/to/verified-assets \
  CANARY_PORT=8788 \
  node scripts/seasonal_assets_canary.mjs | tee /private/tmp/ahairu-seasonal-canary.ndjson
node --test scripts/seasonal_assets_canary.test.mjs
```

Each transcript line is one JSON object. Retain it with the accepted-bundle
provenance and the pre-production record.

```json
{"event":"session-start","pid":12345,"host":"127.0.0.1","port":8788,"routes":12}
{"event":"session-ready","pid":12345,"baseURL":"http://127.0.0.1:8788"}
{"event":"probe","method":"GET","route":"/assets/releases/current","status":200,"type":"application/json; charset=utf-8","cache":"public, max-age=60, must-revalidate","cors":"*","bytes":288,"sha256":"..."}
{"event":"probe","method":"HEAD","route":"/assets/releases/current","status":200,"type":"application/json; charset=utf-8","cache":"public, max-age=60, must-revalidate","cors":"*","bytes":0,"sha256":"e3b0..."}
{"event":"canary-pass","releases":["v0.1.1"]}
{"event":"session-stop","pid":12345,"reason":"complete","exitCode":143,"signal":null}
```

`session-stop` is emitted from normal completion, failure cleanup, and
`SIGINT`/`SIGTERM` handling. A `canary-fail` line precedes a non-zero exit.
The route set is the three channel documents, campaign runtime, `release.json`
and `catalog.json` for every retained immutable release, four canonical pages,
and the two documented 404s. Each record captures status, Content-Type,
Cache-Control, CORS, body bytes, and SHA-256.

## Complete local gate

Executed once, in this order:

```sh
templ generate
git diff --exit-code -- 'site/*_templ.go'
go test ./... -count=1
node --test src/worker.test.js
ASSET_BUNDLE=/private/tmp/ahairu-parity-assets.5JIiK9 npm run check
git diff --check
```

All commands passed. `npm run check` rebuilt from that bundle, passed the Go
suite, public-tree checker, 18 Worker route tests, nine-page Chromium suite,
and `wrangler deploy --dry-run`.

## Local Worker probes

One `wrangler dev --ip 127.0.0.1 --port 8788` session served all probes. GET
body hashes are SHA-256; HEAD bodies were empty. `CORS=*` means
`Access-Control-Allow-Origin: *`; a dash means no such header.

| Path | GET: status, type, cache, bytes, SHA-256 | HEAD: status, type, cache | CORS |
| --- | --- | --- | --- |
| `/assets/releases/latest` | `200`, `application/json; charset=utf-8`, `public, max-age=60, must-revalidate`, `288`, `8bab7b3817378b5aa2aa6512fa70f536ec888c289dbc2d9f9e09bcf62d56db0d` | `200`, same type/cache | `*` |
| `/assets/releases/default` | `200`, `application/json; charset=utf-8`, `public, max-age=60, must-revalidate`, `288`, `8bab7b3817378b5aa2aa6512fa70f536ec888c289dbc2d9f9e09bcf62d56db0d` | `200`, same type/cache | `*` |
| `/assets/releases/current` | `200`, `application/json; charset=utf-8`, `public, max-age=60, must-revalidate`, `288`, `8bab7b3817378b5aa2aa6512fa70f536ec888c289dbc2d9f9e09bcf62d56db0d` | `200`, same type/cache | `*` |
| `/assets/campaign/v1.js` | `200`, `text/javascript; charset=utf-8`, `public, max-age=0, must-revalidate`, `21528`, `a936193b4fed8120e6cb3423f19d3e2ddb0ba32266dc4e5f02a98f5261853709` | `200`, same type/cache | `*` |
| `/assets/releases/v0.1.1/release.json` | `200`, `application/json`, `public, max-age=31536000, immutable`, `134392`, `eb2f556224ce1bcab979e3f1c8c8f05813dc0c3381b30ae757df32216027ebb9` | `200`, same type/cache | `*` |
| `/assets/releases/v0.1.1/catalog.json` | `200`, `application/json`, `public, max-age=31536000, immutable`, `207662`, `bca54f24af0529ebe988c901c6786110f2006a5bcedbab5928ba2795e1cf7d7c` | `200`, same type/cache | `*` |
| `/brand/` | `200`, `text/html; charset=utf-8`, `public, max-age=0, must-revalidate`, `15610`, `c07b7dc0c47977228a9825c59909bc75fb8e98f431f7e69f93bfbc9b05b7c9c6` | `200`, same type/cache | — |
| `/license/` | `200`, `text/html; charset=utf-8`, `public, max-age=0, must-revalidate`, `8893`, `a5faa667756fde3c80958884f779db18f24d43514383f472f56d16354a3fb98a` | `200`, same type/cache | — |
| `/pt-br/` | `200`, `text/html; charset=utf-8`, `public, max-age=0, must-revalidate`, `6847`, `5505faffddd0048fdd68676fd72204d40d25c7824937e6a9f8ba9ff8d2dbfb89` | `200`, same type/cache | — |
| `/es/brand/` | `200`, `text/html; charset=utf-8`, `public, max-age=0, must-revalidate`, `15770`, `c1643680ab69cebd758284ac6474594c21a737e3b4f6ca334b5b68f94d10b849` | `200`, same type/cache | — |
| `/not-a-page` | `404`, `text/plain;charset=UTF-8`, no cache header, `9`, `e3ebaa16dd9d9b9fc107c42183fb6cf9d22927e1af03dbbdfa0ccc38e4e4ac31` | `404`, no type/cache | — |
| `/assets/missing.svg` | `404`, no type/cache, `0`, `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` | `404`, no type/cache | — |

Only immutable release `v0.1.1` was present. Historical release `v0.1.0` was
never deployed through this Worker; it is acceptable debt not to retrofit it
into the local canary bundle. That exception applies only to a release that was
never live. It does not permit removal of any immutable release already
published to consumers.

## Browser evidence and open canary debt

The existing Ahairu Chromium suite passed during `npm run check`: baseline
page rendering, complete localized brand downloads, responsive/fixed specimen
geometry, and reduced-motion CSS behavior.

Additional live Chromium checks against the same Worker session held the
channel fetch open before first paint. `/brand/` stayed at
`data-theme="araihu"`, `data-theme-source="default"`, no `data-campaign`, and
the campaign toggle stayed hidden. Its replaceable logo retained fixed
`width="720" height="134"`. With reduced motion enabled, `/en/` computed
`.storm-hero::before` animation as `none` and `.project-row` transition as
`0s`.

The supplied accepted channel is `source: default`; its only campaign is
disabled. It therefore cannot exercise active campaign apply, saved-preference
precedence, opt-out/reload persistence, campaign CSS/image failure fallback, or
expiry restoration in Ahairu's live browser without altering the verified
bundle. That alteration would invalidate canary evidence.

Assets runtime evidence, separate from this Worker harness: checkout
`/private/tmp/araihu-assets-seasonal-channels-spec`, revision
`410c3f3b7f4686199b8dda2c41ccacdb6147759f`; command
`node --test /private/tmp/araihu-assets-seasonal-channels-spec/runtime/campaign/v1.test.js`
passed `40/40`. It uses a deterministic runtime fixture, not Ahairu's live
Chromium harness.

The enabled gate above now implements these previously missing live browser
scenarios. Its execution remains blocked only by the absent real v0.1.2 input,
not by missing Ahairu harness code.

## Rollback

1. Inspect production history with `wrangler deployments list --name ahairu`.
   Select the immediately preceding complete Worker version, confirmed against
   its recorded Ahairu commit and complete accepted Assets release/channel
   artifact hashes.
2. Restore exactly that version with:

   ```sh
   wrangler rollback <preceding-version-id> --name ahairu \
     --message "Rollback seasonal-assets canary" --yes
   ```

3. Re-run GET and HEAD probes for the channels, runtime, retained immutable
   releases, and canonical pages. A corrective deployment or rollback is
   **expected** to converge within 60 seconds. This is an unmeasured production
   expectation pending deployment, not a proven SLO or canary result. Keep the
   prior healthy version as rollback target until probes agree with recorded
   hashes.
