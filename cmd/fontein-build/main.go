// Command fontein-build performs one site regeneration; systemd runs it every minute.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/De-Fontein/kerkdefontein-website/internal/build"
	"github.com/De-Fontein/kerkdefontein-website/internal/calendar"
	"github.com/De-Fontein/kerkdefontein-website/internal/config"
	"github.com/De-Fontein/kerkdefontein-website/internal/drive"
	"github.com/De-Fontein/kerkdefontein-website/internal/health"
)

var version = "dev" // set by -ldflags in the Makefile

func main() { os.Exit(run()) }

func run() int {
	configPath := flag.String("config", "/etc/fontein/config.json", "path to the server config")
	flag.Parse()
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("run_id", newRunID(), "version", version)

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Error("config invalid", "err", err)
		return 1
	}
	// Under the 60 s timer interval, so runs never pile up.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	web := &http.Client{Timeout: 10 * time.Second}

	googleHTTP, err := googleClient(ctx, cfg.CredentialsFile, web)
	if err != nil {
		log.Error("google credentials", "err", err)
		ping(ctx, log, web, cfg.HealthcheckURL, true, "google credentials: "+err.Error())
		return 1
	}
	outcome, err := build.Run(ctx, log, build.Sources{
		Drive:       drive.Client{HTTP: googleHTTP, BaseURL: cfg.GoogleAPIBase},
		Calendar:    calendar.Client{HTTP: googleHTTP, BaseURL: cfg.GoogleAPIBase},
		Web:         web,
		YouTubeBase: cfg.YouTubeBase,
		ThumbBase:   cfg.ThumbnailBase,
	}, build.Settings{
		FlyerFolderID: cfg.FlyerFolderID, DocumentsFolderID: cfg.DocumentsFolderID,
		CalendarID: cfg.CalendarID, CalendarICalURL: cfg.CalendarICalURL, YouTubeChannelID: cfg.YouTubeChannelID,
		SiteRoot: cfg.SiteRoot, CacheDir: cfg.CacheDir, BaseURL: cfg.BaseURL, Version: version,
	}, time.Now())
	if err != nil {
		ping(ctx, log, web, cfg.HealthcheckURL, true, "build aborted: "+err.Error())
		return 1
	}
	ping(ctx, log, web, cfg.HealthcheckURL, len(outcome.Problems) > 0, describe(outcome))
	return 0
}

func googleClient(ctx context.Context, keyFile string, base *http.Client) (*http.Client, error) {
	key, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	creds, err := google.CredentialsFromJSON(ctx, key,
		"https://www.googleapis.com/auth/drive.readonly", "https://www.googleapis.com/auth/calendar.readonly")
	if err != nil {
		return nil, err
	}
	return oauth2.NewClient(context.WithValue(ctx, oauth2.HTTPClient, base), creds.TokenSource), nil
}

func ping(ctx context.Context, log *slog.Logger, hc *http.Client, url string, failed bool, message string) {
	if err := health.Ping(ctx, hc, url, failed, message); err != nil {
		log.Error("healthcheck ping failed", "err", err)
	}
}

func describe(o build.Outcome) string {
	if !o.Published {
		return "no change"
	}
	if len(o.Problems) == 0 {
		return "published " + o.Release
	}
	var b strings.Builder
	fmt.Fprintf(&b, "published %s, but skipped:\n", o.Release)
	for _, p := range o.Problems {
		fmt.Fprintf(&b, "- %s: %s\n", p.Name, p.Reason)
	}
	return b.String()
}

func newRunID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}
