# CLAUDE.md

kerkdefontein.nl: a static site that a Go program rebuilds every minute from Google Drive (flyers, ANBI documents),
Google Calendar (agenda) and a YouTube feed. **Read `docs/requirements.md` first.** It holds every agreed
requirement and the decision log. Update both when a requirement changes.

## Commands

```bash
brew install vips poppler brotli && npm ci   # once
make preview      # http://localhost:8080 with fake data (internal/fake); no Google account needed
make test         # tsc + go vet + go test
make e2e          # Playwright: axe, behaviour, layout shift, theme, performance; 4 browser projects
make build        # linux/amd64 binary in dist/
make deploy       # build, copy to HOST (default: fontein), run once; see deploy/README.md
```

Run `make test` and `make e2e` before every commit. Both must pass.

## Layout

- `cmd/fontein-build` runs one build: fetch, skip if unchanged, convert, render, publish.
- `cmd/fontein-preview` renders the fake data and serves it the way Caddy does (CSP, Brotli, 404).
- `internal/build`: change detection (fingerprint), media conversion and caching, release writing, precompression.
- `internal/site`: `Pages`, `Nav`, templates (`templates/pages/*.html`, `partials/`), JSON-LD, Dutch date helpers.
- `internal/{drive,calendar,youtube}` are thin REST clients; `internal/content` holds the flyer and document rules.
- `web/static` holds the CSS, images and PDFs, embedded in the binary. `web/ts` compiles to the gitignored `web/static/js`.
- `deploy/` holds the Caddyfile, the systemd units and a runbook. `internal/headers.CSP` must equal the Caddyfile
  CSP; a test enforces it.

## Rules that are easy to break

- **No layout shift.** Anything drawn must stay put while the rest of the page arrives. Give every image
  `width`/`height`. Give header and hero items fixed slots. Don't centre or right-align groups whose width depends
  on content that arrives later. Don't show or hide things after load. `behaviour.spec.ts` cuts every page off at
  every node; keep it green.
- **No third-party requests on page load.** Embeds (YouTube, Scipio) load on click, with a plain-link fallback
  when JS is off. Every new embed origin needs `frame-src` in both CSP copies and a line on `/privacy/`.
- **No personal details** in pages, code, docs or commit messages: the repository is public, and git history
  counts. Only the elders and role addresses (`info@`, `pastoraat@`, `kjp@`, `verhuur@`) are allowed;
  `TestRender_NoPersonalContactDetails` enforces this.
- **Secrets and IDs** (service-account key, Drive folder IDs, healthcheck URL) live only in `/etc/fontein/`, or
  locally in `~/.config/fontein/`. Never in the repository.
- **Text is Dutch** and copied word for word from the old site until the church rewrites it. Keep section anchors
  (`#vrouwen`, `#bijbelstudies`, …): flyers link to them, and `migration_test.go` pins them.
- **A failed source must not publish.** A download failure is a `sourceError` and aborts the run. A broken file
  becomes a `Problem` and is skipped. Never cache a failure.
- **Assets** are referenced through `{{asset "css/site.css"}}`, which adds a content hash. Caddy caches them for
  a year.
- **Budgets:** JS < 10 KB in total; each page plus its CSS ≤ 14 KB compressed (tested in `internal/build`).
- **Themes:** `theme.js` is a blocking script in `<head>` on purpose: it applies the theme and sets `data-js`
  before the first paint. Keep it small, and don't make it `defer` or `module`.

## Gotchas

- The build fingerprint covers the sources, the binary version and the embedded templates and assets, but **not
  the config**. After changing `config.json`, delete `<siteRoot>/fingerprint` to force a rebuild.
- Release names are UTC seconds. Two builds in the same second fail on purpose, so use a later `now` in tests.
- The Playwright `webServer` runs the preview with fake data. Tests that touch YouTube or Scipio must stub them
  with `page.route`.
- `layout-shift` entries exist only in Chromium. The cut-off test in `behaviour.spec.ts` is the cross-browser check.
