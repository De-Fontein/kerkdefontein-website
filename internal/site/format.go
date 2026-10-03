package site

import (
	"fmt"
	"time"

	"github.com/De-Fontein/kerkdefontein-website/internal/calendar"
	"github.com/De-Fontein/kerkdefontein-website/web"
)

// Go has no locale support; these two tables are all the Dutch the date format needs.
var (
	weekdays = [...]string{"zo", "ma", "di", "wo", "do", "vr", "za"}
	months   = [...]string{"jan", "feb", "mrt", "apr", "mei", "jun", "jul", "aug", "sep", "okt", "nov", "dec"}
)

var funcs = map[string]any{
	"dutchDate": dutchDate,
	"dateRange": dateRange,
	"clock":     func(t time.Time) string { return t.In(calendar.Amsterdam).Format("15:04") },
	"isoDate":   isoDate,
	"asset":     web.AssetURL,
	"limit": func(n int, events []EventView) []EventView {
		if len(events) > n {
			return events[:n]
		}
		return events
	},
}

func dutchDate(t time.Time) string {
	t = t.In(calendar.Amsterdam)
	return fmt.Sprintf("%s %d %s", weekdays[t.Weekday()], t.Day(), months[t.Month()-1])
}

func isoDate(t time.Time) string { return t.In(calendar.Amsterdam).Format("2006-01-02") }

// dateRange shows the last day too when an event spans several days. Google's end is exclusive (the next
// day for all-day events, the end instant for timed ones), so step back before taking its date.
func dateRange(e calendar.Event) string {
	last := e.End.Add(-time.Nanosecond)
	if e.AllDay {
		last = e.End.AddDate(0, 0, -1)
	}
	if isoDate(last) <= isoDate(e.Start) {
		return dutchDate(e.Start)
	}
	return dutchDate(e.Start) + " – " + dutchDate(last)
}
