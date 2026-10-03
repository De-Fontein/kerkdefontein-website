package calendar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const eventsJSON = `{"items":[
 {"id":"e1","summary":"Aanbiddingsavond","location":"de Hoeksteen","updated":"2026-10-01T08:00:00Z",
  "start":{"dateTime":"2026-10-24T19:30:00+02:00"},"end":{"dateTime":"2026-10-24T21:30:00+02:00"}},
 {"id":"e2","summary":"Gemeentekamp","updated":"2026-10-01T08:00:00Z",
  "start":{"date":"2026-10-24"},"end":{"date":"2026-10-27"}},
 {"id":"e3","summary":"Bijbelstudie","updated":"2026-10-01T08:00:00Z",
  "start":{"dateTime":"2026-10-28T19:30:00+01:00"},"end":{"dateTime":"2026-10-28T21:00:00+01:00"}}
]}`

func TestUpcoming_ParsesTimedAndAllDayEventsAcrossDST(t *testing.T) {
	var got *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Write([]byte(eventsJSON))
	}))
	defer srv.Close()

	from := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	events, err := Client{HTTP: srv.Client(), BaseURL: srv.URL}.Upcoming(context.Background(), "kerk@group.calendar.google.com", from, 50)
	if err != nil {
		t.Fatal(err)
	}
	if got.URL.Path != "/calendar/v3/calendars/kerk@group.calendar.google.com/events" {
		t.Errorf("path = %s", got.URL.Path)
	}
	q := got.URL.Query()
	if q.Get("singleEvents") != "true" || q.Get("orderBy") != "startTime" || q.Get("timeMin") != "2026-10-03T12:00:00Z" || q.Get("maxResults") != "50" {
		t.Errorf("query = %s", got.URL.RawQuery)
	}
	if len(events) != 3 {
		t.Fatalf("got %d events", len(events))
	}

	timed := events[0]
	if timed.AllDay || timed.Start.In(Amsterdam).Format("2006-01-02 15:04") != "2026-10-24 19:30" || timed.Location != "de Hoeksteen" {
		t.Errorf("timed event = %+v", timed)
	}
	allDay := events[1]
	wantStart := time.Date(2026, 10, 24, 0, 0, 0, 0, Amsterdam)
	wantEnd := time.Date(2026, 10, 27, 0, 0, 0, 0, Amsterdam)
	if !allDay.AllDay || !allDay.Start.Equal(wantStart) || !allDay.End.Equal(wantEnd) {
		t.Errorf("all-day event = %+v", allDay)
	}
	// After the 25 Oct DST switch the UTC offset is +01:00; wall-clock time must stay 19:30.
	if got := events[2].Start.In(Amsterdam).Format("15:04 -07:00"); got != "19:30 +01:00" {
		t.Errorf("post-DST event start = %s", got)
	}
}
