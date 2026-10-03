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
	thumb := images.Variant{File: "m1-480.webp", Width: 480, Height: 679}
	large := images.Variant{File: "m1-1600.webp", Width: 1131, Height: 1600}
	return Data{
		BaseURL: "https://kerkdefontein.nl",
		Video:   &VideoView{ID: "qeWRxvxJvNs", Title: "Vrolijk zijn in Gods tent!", Thumb: images.Variant{File: "yt-480.webp", Width: 480, Height: 360}},
		Flyers: []FlyerView{
			{ID: "f1", Alt: "Aanbiddingsavond", Link: "https://example.org/aanbidding", Thumb: thumb, Large: large},
			{ID: "f2", Alt: "Bijbelstudie", Thumb: thumb, Large: large},
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
		`src="/media/m1-480.webp"`,
		`width="480" height="679"`,
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

// The ▶ icon must not be part of the button's accessible name, or screen readers announce it.
func TestRender_VideoButtonNameStartsWithAfspelen(t *testing.T) {
	html := render(t, sampleData(), "index.html")
	if !strings.Contains(html, `<span aria-hidden="true">▶&nbsp;</span>Afspelen: Vrolijk zijn in Gods tent!`) {
		t.Error("play icon must be aria-hidden, followed by the Afspelen label")
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
	if !regexp.MustCompile(`<button[^>]*\bdata-theme-toggle\b[^>]*\shidden[\s>]`).MatchString(html) {
		t.Error("theme toggle must start hidden, so it never shows without JavaScript")
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
