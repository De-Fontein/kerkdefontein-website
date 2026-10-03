package site

import (
	"encoding/json"
	"html/template"
	"time"
)

// jsonLD describes the church on every page and, on /agenda/, each event (schema.org, read by search engines
// and AI assistants). json.Marshal escapes <, > and &, so the output is safe inside <script>.
func jsonLD(p Page, d Data) (template.JS, error) {
	churchID := d.BaseURL + "/#kerk"
	graph := []any{map[string]any{
		"@type": "Church", "@id": churchID,
		"name": "Baptistengemeente De Fontein", "url": d.BaseURL + "/", "email": "info@kerkdefontein.nl",
		"address": map[string]any{
			"@type": "PostalAddress", "streetAddress": "Roozenburglaan 22", "postalCode": "4337 JH",
			"addressLocality": "Middelburg", "addressCountry": "NL",
		},
		"event": map[string]any{
			"@type": "Event", "name": "Samenkomst", "location": map[string]any{"@id": churchID},
			"eventSchedule": map[string]any{
				"@type": "Schedule", "byDay": "https://schema.org/Sunday", "startTime": "10:00",
				"repeatFrequency": "P1W", "scheduleTimezone": "Europe/Amsterdam",
			},
		},
	}}
	if p.Path == "/agenda/" {
		for _, e := range d.Events {
			graph = append(graph, eventLD(e, churchID))
		}
	}
	b, err := json.Marshal(map[string]any{"@context": "https://schema.org", "@graph": graph})
	return template.JS(b), err
}

func eventLD(e EventView, churchID string) map[string]any {
	layout := time.RFC3339
	if e.AllDay {
		layout = "2006-01-02"
	}
	var location any = map[string]any{"@id": churchID}
	if e.Location != "" {
		location = map[string]any{"@type": "Place", "name": e.Location}
	}
	return map[string]any{
		"@type": "Event", "name": e.Summary, "location": location,
		"startDate": e.Start.Format(layout), "endDate": e.End.Format(layout),
	}
}
