# Seasonal assets local canary — 2026-07-30

## Scope and input

- Ahairu commit: `c8d0b9220f452ee67283bfec9a951e0a249a9359`
- Verified local bundle: `/private/tmp/ahairu-parity-assets.5JIiK9`
- Channel target: `v0.1.1`, digest `34193fd6171ed32cb3307124cfef8be7713503942864dc5a67623c174d4a6c4f`
- Runtime SHA-256: `a936193b4fed8120e6cb3423f19d3e2ddb0ba32266dc4e5f02a98f5261853709`

No production deployment, promotion, tag, or push occurred.

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

Only immutable release `v0.1.1` was present. Required second-release probe
`/assets/releases/v0.1.0/release.json` returned `404` for both GET and HEAD;
GET had an empty body hash `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`.
Its cache header was `public, max-age=31536000, immutable` because its pathname
matches the immutable-release route rule.

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
bundle. That alteration would invalidate canary evidence. The Assets runtime
source suite was also run locally: 40/40 passed, including those scenarios and
the reduced-motion lifecycle hook, but it uses its deterministic runtime
fixture rather than Ahairu's live Chromium harness.

**Handoff condition:** do not treat this as a completed enabled-campaign canary.
Before production, supply an accepted bundle retaining two immutable releases
and a time-bounded enabled campaign channel, then repeat the live browser
scenarios above against that unmodified bundle.

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
   expected to converge within 60 seconds; keep the prior healthy version as
   the rollback target until those probes agree with its recorded hashes.
