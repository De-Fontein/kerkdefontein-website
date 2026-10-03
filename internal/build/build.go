// Package build runs one regeneration: fetch, skip if unchanged, convert, render, publish.
package build

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/De-Fontein/kerkdefontein-website/internal/calendar"
	"github.com/De-Fontein/kerkdefontein-website/internal/content"
	"github.com/De-Fontein/kerkdefontein-website/internal/drive"
	"github.com/De-Fontein/kerkdefontein-website/internal/release"
	"github.com/De-Fontein/kerkdefontein-website/internal/site"
	"github.com/De-Fontein/kerkdefontein-website/internal/youtube"
)

const (
	keepReleases   = 5
	maxEvents      = 50
	fingerprintTxt = "fingerprint"
)

// embeddedContent covers what the version string misses: two builds from the same uncommitted code share a
// version, and a template edit must still republish. A variable so a test can simulate an edit.
var embeddedContent = site.ContentHash

type Sources struct {
	Drive       drive.Client
	Calendar    calendar.Client
	Web         *http.Client // YouTube feed and thumbnails; carries no Google credentials
	YouTubeBase string
	ThumbBase   string
}

type Settings struct {
	FlyerFolderID, DocumentsFolderID, CalendarID, CalendarICalURL, YouTubeChannelID string
	SiteRoot, CacheDir, BaseURL, Version                                            string
}

type Outcome struct {
	Published bool
	Release   string
	Problems  []content.Problem
}

type inputs struct {
	flyers, docs []drive.File
	events       []calendar.Event
	video        youtube.Video
}

func Run(ctx context.Context, log *slog.Logger, src Sources, s Settings, now time.Time) (Outcome, error) {
	in, err := gather(ctx, src, s, now)
	if err != nil {
		log.Error("source failed, keeping current release", "err", err)
		return Outcome{}, err
	}
	fp := Fingerprint(s.Version+"+"+embeddedContent(), in.flyers, in.docs, in.events, in.video)
	if previous, err := os.ReadFile(filepath.Join(s.SiteRoot, fingerprintTxt)); err == nil && string(previous) == fp {
		log.Info("no change", "fingerprint", fp[:12])
		return Outcome{}, nil
	}

	flyers := content.SelectFlyers(in.flyers)
	for _, p := range flyers.BadLinks {
		log.Warn("flyer link ignored", "file", p.Name, "reason", p.Reason)
	}
	docs, docProblems := content.SelectDocuments(in.docs)
	m := media{src: src, cacheDir: s.CacheDir, log: log}
	prepared, mediaProblems, err := m.prepare(ctx, flyers.Shown, docs, in.video)
	if err == nil {
		err = ctx.Err() // a conversion killed by the deadline looks like a broken file; it is not
	}
	if err != nil {
		log.Error("source failed, keeping current release", "err", err)
		return Outcome{}, err
	}
	problems := append(append(flyers.Skipped, docProblems...), mediaProblems...)
	for _, p := range problems {
		log.Warn("file skipped", "file", p.Name, "reason", p.Reason)
	}

	name := now.UTC().Format("20060102T150405Z")
	dir, err := release.Publish(s.SiteRoot, name, keepReleases, func(dir string) error {
		return writeSite(dir, s, prepared, in.events, now)
	})
	if err != nil {
		log.Error("publish failed, keeping current release", "err", err)
		return Outcome{}, err
	}
	if err := recordFingerprint(s.SiteRoot, fp, problems); err != nil {
		return Outcome{}, err
	}
	log.Info("published", "release", name, "flyers", len(prepared.flyers), "events", len(in.events), "documents", len(prepared.docs))
	return Outcome{Published: true, Release: dir, Problems: problems}, nil
}

func gather(ctx context.Context, src Sources, s Settings, now time.Time) (inputs, error) {
	var in inputs
	var err error
	if in.flyers, err = src.Drive.List(ctx, s.FlyerFolderID); err != nil {
		return inputs{}, fmt.Errorf("flyers: %w", err)
	}
	if in.docs, err = src.Drive.List(ctx, s.DocumentsFolderID); err != nil {
		return inputs{}, fmt.Errorf("documents: %w", err)
	}
	if in.events, err = src.Calendar.Upcoming(ctx, s.CalendarID, now, maxEvents); err != nil {
		return inputs{}, fmt.Errorf("events: %w", err)
	}
	if in.video, err = youtube.Latest(ctx, src.Web, src.YouTubeBase, s.YouTubeChannelID); err != nil {
		return inputs{}, fmt.Errorf("video: %w", err)
	}
	return in, nil
}

// recordFingerprint skips the record while files are skipped, so every run retries them and keeps the
// healthcheck alert open until a volunteer fixes the file.
func recordFingerprint(siteRoot, fp string, problems []content.Problem) error {
	path := filepath.Join(siteRoot, fingerprintTxt)
	if len(problems) > 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("clear fingerprint: %w", err)
		}
		return nil
	}
	if err := os.WriteFile(path, []byte(fp), 0o644); err != nil {
		return fmt.Errorf("write fingerprint: %w", err)
	}
	return nil
}

// Fingerprint includes the binary version, so a deploy always triggers a rebuild with the new templates.
func Fingerprint(version string, flyers, docs []drive.File, events []calendar.Event, video youtube.Video) string {
	h := sha256.New()
	json.NewEncoder(h).Encode(struct {
		Version string
		Flyers  []drive.File
		Docs    []drive.File
		Events  []calendar.Event
		Video   youtube.Video
	}{version, flyers, docs, events, video})
	return hex.EncodeToString(h.Sum(nil))
}
