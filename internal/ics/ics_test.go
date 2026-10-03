package ics

import (
	"strings"
	"testing"
	"time"

	"github.com/De-Fontein/kerkdefontein-website/internal/calendar"
)

var stamp = time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)

func TestEvent_TimedEventInUTC(t *testing.T) {
	e := calendar.Event{ID: "e1", Summary: "Aanbiddingsavond", Location: "de Hoeksteen, Middelburg",
		Start: time.Date(2026, 10, 24, 19, 30, 0, 0, calendar.Amsterdam),
		End:   time.Date(2026, 10, 24, 21, 30, 0, 0, calendar.Amsterdam)}

	want := "BEGIN:VCALENDAR\r\nVERSION:2.0\r\nPRODID:-//De Fontein//kerkdefontein.nl//NL\r\n" +
		"BEGIN:VEVENT\r\nUID:e1@kerkdefontein.nl\r\nDTSTAMP:20261003T120000Z\r\n" +
		"DTSTART:20261024T173000Z\r\nDTEND:20261024T193000Z\r\n" +
		"SUMMARY:Aanbiddingsavond\r\nLOCATION:de Hoeksteen\\, Middelburg\r\n" +
		"END:VEVENT\r\nEND:VCALENDAR\r\n"
	if got := string(Event(e, stamp)); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestEvent_AllDayUsesDateValuesWithExclusiveEnd(t *testing.T) {
	e := calendar.Event{ID: "e2", Summary: "Gemeentekamp", AllDay: true,
		Start: time.Date(2026, 10, 24, 0, 0, 0, 0, calendar.Amsterdam),
		End:   time.Date(2026, 10, 27, 0, 0, 0, 0, calendar.Amsterdam)}
	got := string(Event(e, stamp))
	if !strings.Contains(got, "DTSTART;VALUE=DATE:20261024\r\n") || !strings.Contains(got, "DTEND;VALUE=DATE:20261027\r\n") {
		t.Errorf("got:\n%s", got)
	}
	if strings.Contains(got, "LOCATION") {
		t.Error("empty location must be omitted")
	}
}

func TestEvent_EscapesAndFoldsLongLinesOnRuneBoundaries(t *testing.T) {
	e := calendar.Event{ID: "e3", Summary: strings.Repeat("Aanbidding; lofprijzing — ", 5) + "\nnieuwe regel",
		Start: stamp, End: stamp.Add(time.Hour)}
	got := string(Event(e, stamp))
	for _, line := range strings.Split(strings.TrimSuffix(got, "\r\n"), "\r\n") {
		if len(line) > 75 {
			t.Errorf("line longer than 75 octets: %q", line)
		}
		if !strings.HasPrefix(line, " ") && strings.ContainsRune(line, '�') {
			t.Errorf("broken rune in %q", line)
		}
	}
	unfolded := strings.ReplaceAll(got, "\r\n ", "")
	if !strings.Contains(unfolded, `Aanbidding\; lofprijzing — `) || !strings.Contains(unfolded, `\nnieuwe regel`) {
		t.Errorf("escaping wrong:\n%s", unfolded)
	}
}
