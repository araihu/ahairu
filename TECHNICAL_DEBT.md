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
