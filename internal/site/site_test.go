package site

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/De-Fontein/kerkdefontein-website/internal/calendar"
	"github.com/De-Fontein/kerkdefontein-website/internal/images"
)

func sampleData() Data {
	at := func(day, hour int) time.Time { return time.Date(2026, 10, day, hour, 0, 0, 0, calendar.Amsterdam) }
	var events []EventView
	for i, name := range []string{"Aanbiddingsavond", "Bijbelstudie", "Samenkomst", "Vrouwenochtend"} {
		events = append(events, EventView{
			Event:   calendar.Event{ID: name, Summary: name, Start: at(10+i, 19), End: at(10+i, 21)},
			ICSHref: "/agenda/" + name + ".ics",
		})
	}
	thumb := images.Variant{File: "m1-360.webp", Width: 360, Height: 509}
	thumb2x := images.Variant{File: "m1-720.webp", Width: 720, Height: 1018}
	large := images.Variant{File: "m1-1600.webp", Width: 1131, Height: 1600}
	return Data{
		BaseURL: "https://kerkdefontein.nl",
		Video:   &VideoView{ID: "qeWRxvxJvNs", Title: "Vrolijk zijn in Gods tent!", Thumb: images.Variant{File: "yt-480.webp", Width: 480, Height: 360}},
		Flyers: []FlyerView{
			{ID: "f1", Alt: "Aanbiddingsavond", Link: "https://example.org/aanbidding", Thumb: thumb, Thumb2x: thumb2x, Large: large},
			{ID: "f2", Alt: "Bijbelstudie", Thumb: thumb, Thumb2x: thumb2x, Large: large},
		},
		Events: events,
	}
}

func render(t *testing.T, d Data, file string) string {
	t.Helper()
	dir := t.TempDir()
	if err := Render(dir, d); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, file))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRender_HomeShowsVideoFlyersAndThreeEvents(t *testing.T) {
	html := render(t, sampleData(), "index.html")
	for _, want := range []string{
		`<html lang="nl">`,
		`data-youtube-id="qeWRxvxJvNs"`,
		`popovertarget="flyer-f1"`,
		`href="https://example.org/aanbidding"`,
		`src="/media/m1-360.webp"`,
		`width="360" height="509"`,
		`property="og:title"`,
		`property="og:image" content="https://kerkdefontein.nl/static/img/share.jpg"`,
		`"@type":"Church"`,
		`<link rel="stylesheet" href="/static/css/site.css?v=`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("home misses %s", want)
		}
	}
	if n := strings.Count(html, `class="event"`); n != 3 {
		t.Errorf("home shows %d events, want 3", n)
	}
	if n := strings.Count(html, "Meer info"); n != 1 {
		t.Errorf("Meer info appears %d times, want 1 (only f1 has a link)", n)
	}
	if !strings.Contains(html, "za 10 okt") {
		t.Error("dates must be Dutch")
	}
}

func TestRender_HomeEmptyStates(t *testing.T) {
	html := render(t, Data{BaseURL: "https://kerkdefontein.nl"}, "index.html")
	for _, want := range []string{"Er zijn op dit moment geen actuele berichten.", "Er staan geen activiteiten gepland."} {
		if !strings.Contains(html, want) {
			t.Errorf("missing empty state %q", want)
		}
	}
	if strings.Contains(html, "data-youtube-id") {
		t.Error("no video section expected without a video")
	}
}

func TestRender_EscapesVolunteerText(t *testing.T) {
	d := sampleData()
	d.Flyers[0].Alt = `Avond & "Lof" <script>alert(1)</script> 🎉 ` + strings.Repeat("lang ", 50)
	html := render(t, d, "index.html")
	if strings.Contains(html, "<script>alert(1)") {
		t.Fatal("flyer name was not escaped")
	}
	if !strings.Contains(html, "Avond &amp; &#34;Lof&#34; &lt;script&gt;") {
		t.Error("expected escaped flyer name in output")
	}
}

func TestRender_NotFoundPage(t *testing.T) {
	html := render(t, sampleData(), "404.html")
	if !strings.Contains(html, "Pagina niet gevonden") || !strings.Contains(html, `href="/agenda/"`) {
		t.Error("404 page misses its heading or links")
	}
}

// The ▶ icon must not be part of the button's accessible name, or screen readers announce it. The title is a
// caption below the thumbnail (a long title filled and overflowed it), and stays in the button's aria-label;
// a visually hidden span would not do, as browsers put a space before it ("Afspelen : …").
func TestRender_VideoButtonNameStartsWithAfspelen(t *testing.T) {
	html := render(t, sampleData(), "index.html")
	if !strings.Contains(html, `aria-label="Afspelen: Vrolijk zijn in Gods tent!"`) || !strings.Contains(html, `<span class="video-play" aria-hidden="true">▶&nbsp;Afspelen</span>`) {
		t.Error("the button must be named Afspelen: <title>, with the visible ▶ Afspelen hidden from screen readers")
	}
	if !strings.Contains(html, `<p class="video-title" aria-hidden="true">Vrolijk zijn in Gods tent!</p>`) {
		t.Error("the title must show as a caption below the video")
	}
}

func TestRender_MultiDayEventsShowTheirDateRange(t *testing.T) {
	d := Data{BaseURL: "https://kerkdefontein.nl", Events: []EventView{
		{Event: calendar.Event{ID: "kamp", Summary: "Gemeentekamp", AllDay: true,
			Start: time.Date(2026, 10, 24, 0, 0, 0, 0, calendar.Amsterdam), End: time.Date(2026, 10, 27, 0, 0, 0, 0, calendar.Amsterdam)}},
		{Event: calendar.Event{ID: "dag", Summary: "Startzondag", AllDay: true,
			Start: time.Date(2026, 11, 1, 0, 0, 0, 0, calendar.Amsterdam), End: time.Date(2026, 11, 2, 0, 0, 0, 0, calendar.Amsterdam)}},
		{Event: calendar.Event{ID: "nacht", Summary: "Gebedsnacht",
			Start: time.Date(2026, 11, 6, 22, 0, 0, 0, calendar.Amsterdam), End: time.Date(2026, 11, 7, 6, 0, 0, 0, calendar.Amsterdam)}},
	}}
	html := render(t, d, "agenda/index.html")
	if !strings.Contains(html, "za 24 okt – ma 26 okt") {
		t.Error("multi-day all-day event must show its last day (Google's end date is exclusive)")
	}
	if !strings.Contains(html, ">zo 1 nov<") {
		t.Error("single-day all-day event must show one date")
	}
	if !strings.Contains(html, "vr 6 nov – za 7 nov") {
		t.Error("a timed event past midnight must show both dates")
	}
}

// The theme script must block rendering (no defer/async/module) so a saved theme is applied before first paint.
func TestRender_ThemeToggleAndBlockingScriptInHead(t *testing.T) {
	html := render(t, sampleData(), "index.html")
	head := html[:strings.Index(html, "</head>")]
	if !regexp.MustCompile(`<script src="/static/js/theme\.js\?v=[0-9a-f]+"></script>`).MatchString(head) {
		t.Error("head must load theme.js as a plain blocking script")
	}
}

// The menu stays in the top-right corner (user, 2026-10-03), so the theme toggle precedes it in the DOM.
func TestRender_MenuIsLastInHeader(t *testing.T) {
	html := render(t, sampleData(), "index.html")
	toggle, nav := strings.Index(html, "data-theme-toggle"), strings.Index(html, `<nav aria-label="Hoofdmenu">`)
	if toggle < 0 || nav < 0 || toggle > nav {
		t.Errorf("theme toggle (at %d) must come before the menu (at %d)", toggle, nav)
	}
}

// Scripts are cached for a year like the stylesheet, so their URLs must change with their content: a stale
// cached theme.js crashed against new markup after a deploy (2026-10-03).
func TestRender_AssetURLsCarryAContentHash(t *testing.T) {
	html := render(t, sampleData(), "index.html")
	for _, asset := range []string{"/static/js/theme.js", "/static/js/youtube.js", "/static/css/site.css"} {
		if !regexp.MustCompile(regexp.QuoteMeta(asset) + `\?v=[0-9a-f]{10}"`).MatchString(html) {
			t.Errorf("%s is not versioned", asset)
		}
	}
}

// The theme menu is native HTML: a popover opened by the button, holding one radio group.
func TestRender_ThemeMenuIsANativePopoverWithFourChoices(t *testing.T) {
	html := render(t, sampleData(), "index.html")
	if !strings.Contains(html, `popovertarget="theme-menu"`) || !strings.Contains(html, `id="theme-menu" class="theme-menu" popover`) {
		t.Error("theme button must open the #theme-menu popover")
	}
	for _, value := range []string{"auto", "light", "dark", "oled"} {
		if !strings.Contains(html, `<input type="radio" name="theme" value="`+value+`"`) {
			t.Errorf("theme menu misses the %s option", value)
		}
	}
}

// Chromium paints a half-arrived page; rel=expect holds the first paint until the header is complete, so header
// items never move as the rest of it arrives (Firefox and Safari ignore it and rely on the fixed CSS slots).
func TestRender_FirstPaintWaitsForTheWholeHeader(t *testing.T) {
	for _, p := range Pages {
		html := render(t, sampleData(), outputPath("", p.Path))
		head := html[:strings.Index(html, "</head>")]
		if !strings.Contains(head, `<link rel="expect" href="#main" blocking="render">`) || !strings.Contains(html, `<main id="main">`) {
			t.Errorf("%s: first paint does not wait for the header", p.Path)
		}
	}
}

// Lighthouse (user, 2026-10-03): a fixed 480 px flyer thumbnail was too big for desktop and too small for retina.
func TestRender_FlyerThumbnailsLetTheBrowserPickASize(t *testing.T) {
	html := render(t, sampleData(), "index.html")
	want := `src="/media/m1-360.webp" srcset="/media/m1-360.webp 360w, /media/m1-720.webp 720w" sizes="(min-width: 72rem) 22rem, (min-width: 48rem) 30vw, 45vw"`
	if !strings.Contains(html, want) {
		t.Errorf("flyer thumbnail misses srcset/sizes: want %s", want)
	}
}

// The video thumbnail is the largest element on desktop, so it loads first; logo and favicon carry a content
// hash so they can be cached for a year.
func TestRender_VideoThumbnailHasHighPriorityAndImagesAreVersioned(t *testing.T) {
	html := render(t, sampleData(), "index.html")
	if !regexp.MustCompile(`<img src="/media/yt-480\.webp"[^>]*fetchpriority="high"`).MatchString(html) {
		t.Error("video thumbnail must have fetchpriority=high")
	}
	if n := len(regexp.MustCompile(`/static/img/logo\.svg\?v=[0-9a-f]{10}`).FindAllString(html, -1)); n != 2 {
		t.Errorf("logo and favicon must use the versioned URL, found %d", n)
	}
}
