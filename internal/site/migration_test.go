package site

import (
	"strings"
	"testing"
)

// Every page of the old site's menu must have a home here (user, 2026-10-03: Israël and others were missing).
// Pinned by section anchor, not by sentence, so the text itself can still be rewritten later.
func TestRender_EveryOldSiteSectionHasAHome(t *testing.T) {
	want := map[string][]string{
		"samenkomsten/index.html":      {`id="welkom"`, `id="avondmaal"`, `id="dopen"`, `id="opdragen"`},
		"activiteiten/index.html":      {`id="bijbelstudies"`, `id="kinderen"`, `id="tieners-en-jeugd"`, `id="jong-volwassenen"`, `id="kringen"`, `id="vrouwen"`, `id="bidstonden"`, `id="evangelisatie"`, `id="pastoraat"`, `id="kjp"`},
		"over-ons/index.html":          {`id="wie-zijn-wij"`, `href="/over-ons/israel/"`, `href="/over-ons/baptisme/"`},
		"over-ons/israel/index.html":   {"Gods oogappel", "Luc. 21:29-33"},
		"privacy/index.html":           {`href="/static/docs/privacyverklaring.pdf"`},
		"over-ons/baptisme/index.html": {"Ontstaan van het Baptisme", "Billy Graham", `href="/static/docs/baptisme-historie.pdf"`},
	}
	for file, needles := range want {
		html := render(t, sampleData(), file)
		for _, needle := range needles {
			if !strings.Contains(html, needle) {
				t.Errorf("%s misses %s", file, needle)
			}
		}
	}
}
