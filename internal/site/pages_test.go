package site

import (
	"strings"
	"testing"
)

func TestRender_EveryPageExistsWithUniqueTitle(t *testing.T) {
	d := sampleData()
	d.Documents = []DocumentView{{Title: "Jaarverslag 2025", Href: "/documenten/abc.pdf"}}
	seen := map[string]bool{}
	for _, p := range Pages {
		if seen[p.Title] {
			t.Errorf("duplicate title %q", p.Title)
		}
		seen[p.Title] = true
		html := render(t, d, outputPath("", p.Path))
		if !strings.Contains(html, "<h1") {
			t.Errorf("%s has no h1", p.Path)
		}
	}
	for _, path := range []string{"/", "/eerste-keer/", "/samenkomsten/", "/activiteiten/", "/agenda/", "/over-ons/", "/doneren/", "/verhuur/", "/privacy/", "/404"} {
		found := false
		for _, p := range Pages {
			found = found || p.Path == path
		}
		if !found {
			t.Errorf("page %s missing", path)
		}
	}
}

// The old site spells the street Roozenburglaan everywhere; an early summary misspelled it Roojenburglaan.
func TestRender_AddressIsRoozenburglaan(t *testing.T) {
	html := render(t, sampleData(), "index.html")
	if !strings.Contains(html, "Roozenburglaan 22") || strings.Contains(html, "Roojenburglaan") {
		t.Error("address must read Roozenburglaan 22 in footer and schema.org")
	}
}

func TestRender_AgendaListsAllEventsWithICSAndSchema(t *testing.T) {
	d := sampleData()
	d.CalendarICalURL = "https://calendar.google.com/calendar/ical/x/public/basic.ics"
	html := render(t, d, "agenda/index.html")
	if n := strings.Count(html, `class="event"`); n != 4 {
		t.Errorf("agenda shows %d events, want 4", n)
	}
	for _, want := range []string{`href="/agenda/Aanbiddingsavond.ics"`, `"@type":"Event"`, d.CalendarICalURL} {
		if !strings.Contains(html, want) {
			t.Errorf("agenda misses %s", want)
		}
	}
}

func TestRender_AgendaEmptyStateAndNoSubscribeLinkWithoutURL(t *testing.T) {
	html := render(t, Data{BaseURL: "https://kerkdefontein.nl"}, "agenda/index.html")
	if !strings.Contains(html, "Er staan geen activiteiten gepland.") {
		t.Error("missing empty state")
	}
	if strings.Contains(html, "Abonneer") {
		t.Error("subscribe link shown without a calendar URL")
	}
}

func TestRender_DonerenHasANBIDataAndDocuments(t *testing.T) {
	d := sampleData()
	d.Documents = []DocumentView{{Title: "Jaarverslag 2025", Href: "/documenten/abc.pdf"}}
	html := render(t, d, "doneren/index.html")
	for _, want := range []string{"7467990", "50766775", "NL07RABO 0385 3320 41", `href="/documenten/abc.pdf"`, "Jaarverslag 2025",
		"https://payment-links.mollie.com/nl/payment/qCZVRaf4fhm2f37QGayKz/details", "https://link.socie.nl/r/sci/c/HC9DTZ7CKV"} {
		if !strings.Contains(html, want) {
			t.Errorf("doneren misses %s", want)
		}
	}
}
