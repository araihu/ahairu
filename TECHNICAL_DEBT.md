# Technical debt

## Browser verification runtime

Status: open. Some local environments have the `puppeteer` package but no
downloaded Chrome binary, so `npm run test:visual` cannot run there. CI remains
the release authority after `npm ci`; local contributors should install the
matching browser before treating visual coverage as complete.

## Content Security Policy

Status: open. The Worker currently emits no Content-Security-Policy header.
Before adding one, inventory templ output, the deferred campaign runtime,
integrity attributes, styles, social assets, and Workers asset responses. Add a
strict report-only policy first, then enforce only after browser and production
header contracts prove no required resource is blocked.

## Workflow event execution

Status: open. Local contract tests prove workflow source structure, dispatch
payload separation, artifact endpoint binding, and deployment-version selection.
They cannot execute GitHub's expression engine, GitHub App permission issuance,
artifact API, or protected-environment deployment. Add an isolated GitHub
workflow fixture/reusable-workflow harness when one can exercise those paths
without production credentials.

## Retired Assets flat repository variables

Status: migration open. `ASSETS_RELEASE_URL`, `ASSETS_RELEASE_ID`,
`ASSETS_RELEASE_SHA256`, `ASSETS_CHANNEL_URL`, `ASSETS_CHANNEL_ID`, and
`ASSETS_CHANNEL_SHA256` cannot represent a cumulative `release_artifacts`
handoff and are no longer consumed. Operators using main-push promotion must
migrate to one `ASSETS_RELEASE_HANDOFF_JSON` value matching the documented
`araihu-assets-released` payload before removing the retired variables.
