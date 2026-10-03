# Requirements

The agreed requirements for kerkdefontein.nl, the website of Baptistengemeente De Fontein in Middelburg.
This file is the reference. When a requirement changes, update it here, and add the decision to the log at the
bottom with its date.

## Goal

The site replaces the Rocketcake site hosted at Mijndomein. The new site is fast, accessible and hosted in the EU.
Volunteers can publish flyers without technical help. The code is open source (MIT) at
`github.com/De-Fontein/kerkdefontein-website`.

The site and email move separately. Moving the domain and the 50 mailboxes (about 20 GB) is a later project.

## Content sources

| Content | Source | Edited by |
|---|---|---|
| Flyers ("Actueel") | Shared Google Drive folder `Website/Flyers` | Volunteers |
| ANBI documents | Drive folder `Documenten` | Treasurer |
| Agenda | Google Calendar, synced from Scipio | Unchanged (Scipio) |
| Service video | Atom feed of YouTube channel `@defonteinmiddelburg` | Unchanged |
| Online giving | Scipio giving form (goals managed in Scipio) | Scipio admins |
| All other text | Templates in this repository | Webmaster (or AI, reviewed) |
| Member-only material | Scipio (invite-only) | Not linked from the site |

### Flyer rules
- The **6 newest** flyers show, sorted by Drive creation time, newest first.
- An `Archief` subfolder hides flyers manually.
- Accepted files are JPEG, PNG, WebP, and PDF (page 1 only), up to 25 MB.
- The **file name is the screen-reader text**. The extension is dropped, and dots inside the name such as "19.30" are kept.
- If the Drive **description** is an absolute `https://` URL, the enlarged flyer shows a "Meer info" button. Any
  other description is ignored and logged.
- A click opens the enlarged flyer in a native `popover`.
- A new or changed flyer is live within **5 minutes**. The build runs every minute.

### Video
- The video is the newest feed entry by `published`. Accepted limitation: a Bible study uploaded after a Sunday
  stream was scheduled outranks it.
- The player loads only after a click (`youtube-nocookie.com`).

## Quality targets

| Area | Target |
|---|---|
| Accessibility | WCAG 2.2 AA on every page, in every theme, on mobile and desktop |
| Speed | LCP ≤ 1.5 s on slow 4G with 4× CPU throttling |
| Layout shift | **None.** Nothing that is drawn may move while the rest of the page loads, on any page, width or browser |
| Page weight | Homepage ≤ 300 KB (flyers excluded). Each page plus its CSS ≤ 14 KB compressed |
| JavaScript | < 10 KB in total. Every page works without JS |
| Third parties | **Zero requests on page load.** YouTube and Scipio load only after a click. No cookie banner |
| Tracking | No analytics, no visitor statistics, no access logs |
| Freshness | A change in Drive, Calendar or YouTube is visible within 5 minutes |
| Browsers | Last 2 versions of Chrome, Safari (including iOS), Firefox and Edge |

## Technology

- **Build:** a Go program (`fontein-build`) runs every minute from a systemd timer. It fetches the sources and
  stops if nothing changed. Otherwise it converts images, renders the HTML, and switches a symlink to the new
  release atomically. A failed source never publishes a partial site: the current release stays live, and the
  next run retries.
- **Server:** Hetzner Cloud (Germany) running Debian, with Caddy serving precompressed Brotli and gzip files. No
  Node and no backend runtime.
- **Browser:** plain HTML and CSS. TypeScript only for small enhancements. Native `popover` for the menu, the
  flyers and the theme menu.
- **Dependencies:** as few as possible. Every new one needs a reason.
- **Fonts:** system fonts only.

## Pages

| Path | Content |
|---|---|
| `/` | Hero, latest service video, flyers ("Actueel"), next 3 agenda items, giving teaser |
| `/eerste-keer/` | First visit. **Placeholder (lorem ipsum) until the church supplies the text** |
| `/samenkomsten/` | Sunday services, Avondmaal, Dopen, Opdragen |
| `/activiteiten/` | Grouped list beside one flat article; facts (Wanneer / Voor wie / Contact) per activity. Anchors such as `#vrouwen` are linked from flyers. Vrouwen is a **placeholder** |
| `/agenda/` | All upcoming events, an `.ics` download per event, a subscribe link, the nieuwsflits archive |
| `/over-ons/` | Wie zijn wij, Waar staan we voor, Missie, Waarden, and links to the two pages below |
| `/over-ons/israel/` | Israël text, with a link to christenenvoorisrael.nl |
| `/over-ons/baptisme/` | History of the Baptist movement, with the history PDF |
| `/doneren/` | Scipio giving form (embedded on click), bank details, ANBI data and documents, the elders |
| `/verhuur/` | The building is not for rent; contact address |
| `/privacy/` | How the website handles data, and a link to the church's full privacy statement (PDF) |
| `404` | Helpful not-found page |

- **Text:** copied word for word from the old site, apart from personal details (see below) and obvious typos.
  The church will rewrite the text later.
- **Language:** Dutch only for now. Visitors from Ukraine, Germany, Russia and elsewhere make translations a
  future option.
- **Every page** carries Open Graph tags and schema.org JSON-LD (`Church`, plus `Event` on the agenda).
- **Themes:** a dropdown with the options Automatisch, Licht, Donker and OLED (true black). The menu stays top
  right. On desktop the theme button is a borderless icon after the links. The choice is stored in
  `localStorage` and applied before the first paint.

## Privacy and personal details

- The repository and the site contain **no personal contact details**. The only names allowed are the elders
  (raad van oudsten). The only email addresses allowed are role addresses: `info@`, `pastoraat@`, `kjp@` and
  `verhuur@`. Git history matters too, because the repository is public.
- Secrets, folder IDs and the service-account key live only on the server (`/etc/fontein/`), never in the
  repository.
- Logs contain no personal data.

## Out of scope

- Domain transfer and email hosting. These are a later project and need an EU provider.
- Members area (in Scipio), redirects from old URLs, contact form, search, analytics, `llms.txt`, healthchecks.io.

## Decision log

| Date | Decision |
|---|---|
| 2026-10-03 | Build our own site: Go build tool, static output, Hetzner and Caddy |
| 2026-10-03 | No members area and no Scipio link (Scipio is invite-only); no redirects; Dutch only |
| 2026-10-03 | Remove personal details from the site and the git history, except the elders and role addresses |
| 2026-10-03 | Keep the Ruth flyer, including its phone number. **Ask the contact person on the flyer before launch** |
| 2026-10-03 | No healthchecks.io account. The proof of concept runs on a temporary public link |
| 2026-10-03 | Mailchimp: only the archive link. The old sign-up link is dead |
| 2026-10-03 | Theme dropdown with an OLED theme, using native HTML where possible |
| 2026-10-03 | Online giving: embed the Scipio form on click and drop the Mollie payment link |
| 2026-10-03 | Restore all pages missing from the first migration: Israël, Baptisme, every activity, Vieringen |
| 2026-10-03 | Israël page: link to Christenen voor Israël, without embedding their videos |
| 2026-10-03 | Link the church's privacy statement (PDF) from `/privacy/` |
| 2026-10-03 | No layout shift is a hard requirement, tested on every page at 4 widths in 4 browsers |
| 2026-10-03 | Activiteiten: two columns, grouped list that follows along, one flat article (Safari Reader View) |
| 2026-10-03 | Enlarged flyers: ‹ / › buttons, arrow keys and swipes; no auto-rotating carousel (missed content, WCAG 2.2.2) |
| 2026-10-03 | Route plannen opens the visitor's own maps app (Apple Maps, Android's choice), OpenStreetMap otherwise |
| 2026-10-03 | Keep render-blocking CSS and theme script: inlining costs CSP hashes per deploy; LCP is ~0.5 s on slow 4G |

## Open items

- **Text the church still has to supply:** the "Eerste keer?" page and the Vrouwen section.
- **Share image:** a real one. The current image is a crop of the YouTube banner.
- **Privacy statement (2020):** it still mentions a contact form, which the new site doesn't have. Its list of
  processors (Scipio, ING, Rabobank) doesn't include Mollie, Google Workspace, Mailchimp or Hetzner.
- **Launch:** order the Hetzner server, launch at `nieuw.kerkdefontein.nl`, then move the domain.
