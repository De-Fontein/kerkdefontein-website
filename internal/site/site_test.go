package site

import (
	"os"
	"path/filepath"
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
		`<script type="module" src="/static/js/youtube.js"></script>`,
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
