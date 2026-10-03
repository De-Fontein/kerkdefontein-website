package calendar

import (
	"context"
	"os"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const liveCalendarID = "qqhcq4s0adfsihie9345rnrsq00gbsf4@import.calendar.google.com"

// Probes the real Calendar API with the service account. Opt-in, needs credentials.
func TestLive_ServiceAccountReadsPublicCalendar(t *testing.T) {
	keyFile := os.Getenv("FONTEIN_LIVE_KEY")
	if keyFile == "" {
		t.Skip("set FONTEIN_LIVE_KEY to run")
	}
	key, err := os.ReadFile(keyFile)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	creds, err := google.CredentialsFromJSON(ctx, key, "https://www.googleapis.com/auth/calendar.readonly")
	if err != nil {
		t.Fatal(err)
	}
	client := Client{HTTP: oauth2.NewClient(ctx, creds.TokenSource), BaseURL: "https://www.googleapis.com"}
	events, err := client.Upcoming(ctx, liveCalendarID, time.Now(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) == 0 {
		t.Fatal("no upcoming events; the calendar normally has many")
	}
	t.Logf("first event: %s at %s", events[0].Summary, events[0].Start.In(Amsterdam))
}
