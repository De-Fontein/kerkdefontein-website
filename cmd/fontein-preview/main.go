// Command fontein-preview renders the site from fake sources and serves it, for local work and e2e tests.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/De-Fontein/kerkdefontein-website/internal/build"
	"github.com/De-Fontein/kerkdefontein-website/internal/calendar"
	"github.com/De-Fontein/kerkdefontein-website/internal/drive"
	"github.com/De-Fontein/kerkdefontein-website/internal/fake"
	"github.com/De-Fontein/kerkdefontein-website/internal/headers"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	out := flag.String("out", "out/preview", "site root")
	flag.Parse()

	now := time.Now()
	fixtures, err := fake.Sample(now)
	if err != nil {
		log.Fatal(err)
	}
	sources := httptest.NewServer(fixtures.Handler())
	defer sources.Close()

	_, err = build.Run(context.Background(), slog.Default(), build.Sources{
		Drive:       drive.Client{HTTP: sources.Client(), BaseURL: sources.URL},
		Calendar:    calendar.Client{HTTP: sources.Client(), BaseURL: sources.URL},
		Web:         sources.Client(),
		YouTubeBase: sources.URL,
		ThumbBase:   sources.URL,
	}, build.Settings{
		FlyerFolderID: fake.SampleFlyerFolder, DocumentsFolderID: fake.SampleDocumentsFolder,
		CalendarID: "preview", YouTubeChannelID: fake.SampleChannel,
		SiteRoot: *out, CacheDir: filepath.Join(*out, "cache"),
		BaseURL: "http://localhost" + *addr, Version: fmt.Sprint(now.UnixNano()),
	}, now)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("serving http://localhost%s\n", *addr)
	log.Fatal(http.ListenAndServe(*addr, serveSite(filepath.Join(*out, "current"))))
}

// serveSite mimics Caddy: the production CSP, precompressed Brotli, and 404.html with status 404.
func serveSite(root string) http.Handler {
	files := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", headers.CSP)
		target := filepath.Join(root, filepath.FromSlash(path.Clean("/"+r.URL.Path)))
		info, err := os.Stat(target)
		if errors.Is(err, fs.ErrNotExist) {
			serveNotFound(w, root)
			return
		}
		if err == nil && info.IsDir() {
			if !strings.HasSuffix(r.URL.Path, "/") {
				files.ServeHTTP(w, r) // redirects /agenda to /agenda/
				return
			}
			target = filepath.Join(target, "index.html")
		}
		if strings.Contains(r.Header.Get("Accept-Encoding"), "br") && serveBrotli(w, target) {
			return
		}
		files.ServeHTTP(w, r)
	})
}

func serveBrotli(w http.ResponseWriter, target string) bool {
	body, err := os.ReadFile(target + ".br")
	if err != nil {
		return false
	}
	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(target)))
	w.Header().Set("Content-Encoding", "br")
	w.Header().Set("Vary", "Accept-Encoding")
	w.Write(body)
	return true
}

func serveNotFound(w http.ResponseWriter, root string) {
	body, err := os.ReadFile(filepath.Join(root, "404.html"))
	if err != nil {
		http.Error(w, "404 page not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	w.Write(body)
}
