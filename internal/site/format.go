package site

import (
	"fmt"
	"time"

	"github.com/De-Fontein/kerkdefontein-website/internal/calendar"
)

// Go has no locale support; these two tables are all the Dutch the date format needs.
var (
	weekdays = [...]string{"zo", "ma", "di", "wo", "do", "vr", "za"}
	months   = [...]string{"jan", "feb", "mrt", "apr", "mei", "jun", "jul", "aug", "sep", "okt", "nov", "dec"}
)

var funcs = map[string]any{
	"dutchDate": func(t time.Time) string {
		t = t.In(calendar.Amsterdam)
		return fmt.Sprintf("%s %d %s", weekdays[t.Weekday()], t.Day(), months[t.Month()-1])
	},
	"clock":   func(t time.Time) string { return t.In(calendar.Amsterdam).Format("15:04") },
	"isoDate": func(t time.Time) string { return t.In(calendar.Amsterdam).Format("2006-01-02") },
	"limit": func(n int, events []EventView) []EventView {
		if len(events) > n {
			return events[:n]
		}
		return events
	},
}
