# AraiHu

Static AraiHu website, built with templ and Goshtoso.

```sh
templ generate
go run ./cmd/ahairu build
```

The build writes deployable files to `public/`: `index.html` plus the Goshtoso stylesheet it needs. No server or client runtime is required.
