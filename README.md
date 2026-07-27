# AraiHu

Static multilingual Arai Hu organization site, built with templ, Goshtoso, and Cloudflare Workers static assets. English is the fallback locale.

```sh
templ generate
go run ./cmd/ahairu build
```

The build writes localized standalone files to `public/` and a single Goshtoso stylesheet. `src/worker.js` serves `en`, `pt-br`, and `es`, falling back to English for unsupported paths.
