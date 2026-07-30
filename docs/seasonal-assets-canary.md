# Seasonal assets local canary — 2026-07-30

## Scope and input

- Ahairu commit: `c8d0b9220f452ee67283bfec9a951e0a249a9359`
- Verified local bundle: `/private/tmp/ahairu-parity-assets.5JIiK9`
- Channel target: `v0.1.1`, digest `34193fd6171ed32cb3307124cfef8be7713503942864dc5a67623c174d4a6c4f`
- Runtime SHA-256: `a936193b4fed8120e6cb3423f19d3e2ddb0ba32266dc4e5f02a98f5261853709`

No production deployment, promotion, tag, or push occurred.

## Reproducible local HTTP canary

The executable canary builds only the supplied bundle, starts exactly one local
Wrangler session, conditionally waits for `/assets/releases/current`, then
checks every documented GET and HEAD route. It derives expected GET bytes and
SHA-256 values from the just-built `public/` tree, so it rejects a stale route,
header, body, or release mapping rather than relying on this historical table.
It also requires every HEAD response body to be empty.

`ASSET_BUNDLE` must be an absolute path to the accepted, verified Assets
bundle. `CANARY_PORT` is an optional loopback port; `CANARY_TIMEOUT_MS` is an
optional conditional-ready timeout (default 30000 ms). This is not `npm run
check`: it does not run Go, browser, or deploy dry-run suites.

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

**Enabled-campaign pre-production gate (blocking):** do not treat this as a
completed enabled-campaign canary. Before an enabled campaign reaches
production, supply an accepted bundle with two retained immutable releases and
a time-bounded enabled `current` campaign, then run the HTTP preflight with its
explicit structural gates and repeat live browser scenarios against that exact,
unmodified bundle:

```sh
ASSET_BUNDLE=/absolute/path/to/verified-assets \
  CANARY_REQUIRE_TWO_RELEASES=1 CANARY_REQUIRE_ENABLED_CAMPAIGN=1 \
  node scripts/seasonal_assets_canary.mjs
```

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
