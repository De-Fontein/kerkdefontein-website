package build

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/De-Fontein/kerkdefontein-website/internal/calendar"
	"github.com/De-Fontein/kerkdefontein-website/internal/drive"
	"github.com/De-Fontein/kerkdefontein-website/internal/fake"
)

var now = time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)

func setup(t *testing.T, mutate func(*fake.Fixtures)) (Sources, Settings) {
	t.Helper()
	for _, tool := range []string{"vips", "vipsheader", "pdftoppm", "brotli"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not installed", tool)
		}
	}
	fx, err := fake.Sample(now)
	if err != nil {
		t.Fatal(err)
	}
	if mutate != nil {
		mutate(&fx)
	}
	srv := httptest.NewServer(fx.Handler())
	t.Cleanup(srv.Close)
	root := t.TempDir()
	return Sources{
		Drive:       drive.Client{HTTP: srv.Client(), BaseURL: srv.URL},
		Calendar:    calendar.Client{HTTP: srv.Client(), BaseURL: srv.URL},
		Web:         srv.Client(),
		YouTubeBase: srv.URL,
		ThumbBase:   srv.URL,
	}, Settings{
		FlyerFolderID: fake.SampleFlyerFolder, DocumentsFolderID: fake.SampleDocumentsFolder,
		CalendarID: "kerk", YouTubeChannelID: fake.SampleChannel,
		SiteRoot: filepath.Join(root, "site"), CacheDir: filepath.Join(root, "cache"),
		BaseURL: "https://kerkdefontein.nl", Version: "test",
	}
}

var quiet = slog.New(slog.NewTextHandler(io.Discard, nil))

func exists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected %s: %v", path, err)
	}
}

func TestRun_PublishesCompleteSite(t *testing.T) {
	src, s := setup(t, nil)
	out, err := Run(context.Background(), quiet, src, s, now)
	if err != nil {
		t.Fatal(err)
	}
	if !out.Published || len(out.Problems) != 0 {
		t.Fatalf("outcome = %+v", out)
	}
	current := filepath.Join(s.SiteRoot, "current")
	for _, rel := range []string{"index.html", "agenda/index.html", "404.html", "static/js/youtube.js",
		"documenten/doc1.pdf", "agenda/ev1.ics"} {
		exists(t, filepath.Join(current, rel))
	}
	home, _ := os.ReadFile(filepath.Join(current, "index.html"))
	if !strings.Contains(string(home), `data-youtube-id="sample1"`) || strings.Count(string(home), `class="flyer"`) != 3 {
		t.Error("home misses video or flyers")
	}
}

func TestRun_PrecompressesTextFiles(t *testing.T) {
	src, s := setup(t, nil)
	if _, err := Run(context.Background(), quiet, src, s, now); err != nil {
		t.Fatal(err)
	}
	index := filepath.Join(s.SiteRoot, "current", "index.html")
	original, _ := os.ReadFile(index)
	decoded, err := exec.Command("brotli", "-d", "-c", index+".br").Output()
	if err != nil || string(decoded) != string(original) {
		t.Fatalf("index.html.br does not decode to index.html: %v", err)
	}
	exists(t, index+".gz")
	exists(t, filepath.Join(s.SiteRoot, "current", "static", "css", "site.css.br"))
	matches, _ := filepath.Glob(filepath.Join(s.SiteRoot, "current", "media", "*.br"))
	if len(matches) != 0 {
		t.Errorf("images must not be precompressed: %v", matches)
	}
}

func TestRun_UnchangedSourcesDoNotRebuild(t *testing.T) {
	src, s := setup(t, nil)
	if _, err := Run(context.Background(), quiet, src, s, now); err != nil {
		t.Fatal(err)
	}
	out, err := Run(context.Background(), quiet, src, s, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if out.Published {
		t.Error("second run rebuilt although nothing changed")
	}
}

func TestRun_CorruptFlyerIsSkippedAndReported(t *testing.T) {
	src, s := setup(t, func(fx *fake.Fixtures) {
		fx.Files["flyer2"] = []byte("not a png")
		fx.Flyers[1].MD5 = "corrupt"
	})
	out, err := Run(context.Background(), quiet, src, s, now)
	if err != nil {
		t.Fatal(err)
	}
	if !out.Published || len(out.Problems) != 1 || out.Problems[0].Name != "Bijbelstudie Jakobus.png" {
		t.Errorf("outcome = %+v", out)
	}
	home, _ := os.ReadFile(filepath.Join(s.SiteRoot, "current", "index.html"))
	if n := strings.Count(string(home), `class="flyer"`); n != 2 {
		t.Errorf("home shows %d flyers, want 2", n)
	}
}

func TestRun_SourceFailureKeepsPreviousRelease(t *testing.T) {
	src, s := setup(t, nil)
	if _, err := Run(context.Background(), quiet, src, s, now); err != nil {
		t.Fatal(err)
	}
	before, _ := os.Readlink(filepath.Join(s.SiteRoot, "current"))

	failing, _ := setup(t, func(fx *fake.Fixtures) { fx.FailCalendar = true })
	if _, err := Run(context.Background(), quiet, failing, s, now.Add(time.Minute)); err == nil {
		t.Fatal("expected error when the calendar fails")
	}
	after, _ := os.Readlink(filepath.Join(s.SiteRoot, "current"))
	if before != after {
		t.Errorf("current moved from %s to %s after a failed run", before, after)
	}
}

func TestRun_IdenticalFlyersConvertOnceButRenderTwice(t *testing.T) {
	src, s := setup(t, func(fx *fake.Fixtures) {
		fx.Files["flyer2"] = fx.Files["flyer1"]
		fx.Flyers[1].MD5 = fx.Flyers[0].MD5
	})
	if _, err := Run(context.Background(), quiet, src, s, now); err != nil {
		t.Fatal(err)
	}
	cached, err := os.ReadDir(filepath.Join(s.CacheDir, "flyers"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cached) != 2 {
		t.Errorf("%d cached conversions, want 2 (flyer1 and flyer2 share content)", len(cached))
	}
	home, _ := os.ReadFile(filepath.Join(s.SiteRoot, "current", "index.html"))
	if !strings.Contains(string(home), `id="flyer-flyer1"`) || !strings.Contains(string(home), `id="flyer-flyer2"`) {
		t.Error("both flyers must render with their own Drive-ID-based HTML ids")
	}
}

// A transient Drive error must not publish a site with a flyer missing (spec §8: never a partial build).
func TestRun_DownloadFailureAbortsAndKeepsCurrent(t *testing.T) {
	src, s := setup(t, nil)
	if _, err := Run(context.Background(), quiet, src, s, now); err != nil {
		t.Fatal(err)
	}
	before, _ := os.Readlink(filepath.Join(s.SiteRoot, "current"))

	flaky, _ := setup(t, func(fx *fake.Fixtures) {
		fx.Flyers[1].MD5 = "new-content" // forces a download, which the fake answers with 404
		delete(fx.Files, "flyer2")
	})
	if _, err := Run(context.Background(), quiet, flaky, s, now.Add(time.Minute)); err == nil {
		t.Fatal("expected the run to abort on a download failure")
	}
	after, _ := os.Readlink(filepath.Join(s.SiteRoot, "current"))
	if before != after {
		t.Errorf("current moved from %s to %s after a failed download", before, after)
	}
}

// A run that hits its deadline mid-way must abort, not publish whatever it managed to convert.
func TestRun_CancelledContextDoesNotPublish(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	src, s := setup(t, func(fx *fake.Fixtures) { fx.OnDownload = cancel }) // deadline hits after listing
	out, err := Run(ctx, quiet, src, s, now)
	if err == nil || out.Published {
		t.Fatalf("cancelled run: published=%v err=%v", out.Published, err)
	}
	if _, err := os.Stat(filepath.Join(s.SiteRoot, "current")); err == nil {
		t.Error("a cancelled run must not create a release")
	}
}

// A skipped file must keep the alert open: the next run with unchanged sources reports it again.
func TestRun_SkippedFileIsReportedAgainNextRun(t *testing.T) {
	src, s := setup(t, func(fx *fake.Fixtures) {
		fx.Files["flyer2"] = []byte("not a png")
		fx.Flyers[1].MD5 = "corrupt"
	})
	if _, err := Run(context.Background(), quiet, src, s, now); err != nil {
		t.Fatal(err)
	}
	out, err := Run(context.Background(), quiet, src, s, now.Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Problems) != 1 {
		t.Errorf("second run problems = %+v, want the corrupt flyer again", out.Problems)
	}
}
