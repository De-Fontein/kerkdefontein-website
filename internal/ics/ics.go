// Package ics writes single-event RFC 5545 files, small enough to hand-roll instead of adding a dependency.
package ics

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/De-Fontein/kerkdefontein-website/internal/calendar"
)

const maxLineOctets = 75

func Event(e calendar.Event, stamp time.Time) []byte {
	var b strings.Builder
	line := func(s string) { b.WriteString(fold(s)); b.WriteString("\r\n") }

	line("BEGIN:VCALENDAR")
	line("VERSION:2.0")
	line("PRODID:-//De Fontein//kerkdefontein.nl//NL")
	line("BEGIN:VEVENT")
	line("UID:" + e.ID + "@kerkdefontein.nl")
	line("DTSTAMP:" + utc(stamp))
	if e.AllDay {
		line("DTSTART;VALUE=DATE:" + e.Start.In(calendar.Amsterdam).Format("20060102"))
		line("DTEND;VALUE=DATE:" + e.End.In(calendar.Amsterdam).Format("20060102"))
	} else {
		line("DTSTART:" + utc(e.Start))
		line("DTEND:" + utc(e.End))
	}
	line("SUMMARY:" + escape(e.Summary))
	if e.Location != "" {
		line("LOCATION:" + escape(e.Location))
	}
	line("END:VEVENT")
	line("END:VCALENDAR")
	return []byte(b.String())
}

func utc(t time.Time) string { return t.UTC().Format("20060102T150405Z") }

func escape(s string) string {
	return strings.NewReplacer(`\`, `\\`, ";", `\;`, ",", `\,`, "\r\n", `\n`, "\n", `\n`).Replace(s)
}

// fold splits at 75 octets without cutting a multi-byte rune; continuation lines start with one space.
func fold(s string) string {
	var b strings.Builder
	limit := maxLineOctets
	for len(s) > limit {
		cut := limit
		for cut > 0 && !utf8.RuneStart(s[cut]) {
			cut--
		}
		b.WriteString(s[:cut])
		b.WriteString("\r\n ")
		s = s[cut:]
		limit = maxLineOctets - 1
	}
	b.WriteString(s)
	return b.String()
}
