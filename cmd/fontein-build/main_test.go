package main

import (
	"strings"
	"testing"

	"github.com/De-Fontein/kerkdefontein-website/internal/build"
	"github.com/De-Fontein/kerkdefontein-website/internal/content"
)

// describe is the body of the healthcheck ping, i.e. what the webmaster reads in the alert email.
func TestDescribe(t *testing.T) {
	if got := describe(build.Outcome{}); got != "no change" {
		t.Errorf("no-change run = %q", got)
	}
	if got := describe(build.Outcome{Published: true, Release: "/srv/site/releases/x"}); got != "published /srv/site/releases/x" {
		t.Errorf("clean run = %q", got)
	}
	got := describe(build.Outcome{Published: true, Release: "r", Problems: []content.Problem{{Name: "Avond.png", Reason: "larger than 25 MB"}}})
	if !strings.Contains(got, "skipped") || !strings.Contains(got, "- Avond.png: larger than 25 MB") {
		t.Errorf("run with problems = %q", got)
	}
}
