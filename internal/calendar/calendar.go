// Package calendar reads upcoming events from the Google Calendar that Scipio syncs into.
package calendar

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
	_ "time/tzdata" // the server must not depend on the OS tz database being present

	"github.com/De-Fontein/kerkdefontein-website/internal/fetch"
)

var Amsterdam = mustLoad("Europe/Amsterdam")

func mustLoad(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

type Event struct {
	ID       string
	Summary  string
	Location string
	Start    time.Time
	End      time.Time
	AllDay   bool
	Updated  time.Time
}

type Client struct {
	HTTP    *http.Client
	BaseURL string
}

type apiTime struct {
	DateTime time.Time `json:"dateTime"`
	Date     string    `json:"date"`
}

type apiEvent struct {
	ID       string    `json:"id"`
	Summary  string    `json:"summary"`
	Location string    `json:"location"`
	Updated  time.Time `json:"updated"`
	Start    apiTime   `json:"start"`
	End      apiTime   `json:"end"`
}

// Upcoming asks Google to expand recurring events (singleEvents), so no RRULE handling is needed here.
func (c Client) Upcoming(ctx context.Context, calendarID string, from time.Time, max int) ([]Event, error) {
	q := url.Values{}
	q.Set("singleEvents", "true")
	q.Set("orderBy", "startTime")
	q.Set("timeMin", from.UTC().Format(time.RFC3339))
	q.Set("maxResults", strconv.Itoa(max))
	endpoint := c.BaseURL + "/calendar/v3/calendars/" + url.PathEscape(calendarID) + "/events?" + q.Encode()

	var body struct {
		Items []apiEvent `json:"items"`
	}
	if err := fetch.JSON(ctx, c.HTTP, endpoint, &body); err != nil {
		return nil, fmt.Errorf("list calendar events: %w", err)
	}
	events := make([]Event, 0, len(body.Items))
	for _, item := range body.Items {
		e, err := toEvent(item)
		if err != nil {
			return nil, fmt.Errorf("calendar event %s: %w", item.ID, err)
		}
		events = append(events, e)
	}
	return events, nil
}

func toEvent(item apiEvent) (Event, error) {
	e := Event{ID: item.ID, Summary: item.Summary, Location: item.Location, Updated: item.Updated}
	if item.Start.Date == "" {
		e.Start, e.End = item.Start.DateTime, item.End.DateTime
		return e, nil
	}
	start, err := time.ParseInLocation("2006-01-02", item.Start.Date, Amsterdam)
	if err != nil {
		return Event{}, err
	}
	end, err := time.ParseInLocation("2006-01-02", item.End.Date, Amsterdam)
	if err != nil {
		return Event{}, err
	}
	e.Start, e.End, e.AllDay = start, end, true
	return e, nil
}
