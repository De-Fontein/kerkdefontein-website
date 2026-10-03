// Package fake serves recorded-shape Drive, Calendar and YouTube responses, for tests and the local preview.
package fake

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/De-Fontein/kerkdefontein-website/internal/drive"
)

type Fixtures struct {
	FlyerFolderID     string
	DocumentsFolderID string
	Flyers            []drive.File
	Documents         []drive.File
	Files             map[string][]byte // Drive file ID → content
	EventsJSON        string
	FeedXML           string
	Thumbs            map[string][]byte // YouTube video ID → hqdefault.jpg
	FailCalendar      bool
	OnDownload        func() // test hook, called before a Drive file is served
}

func (f Fixtures) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /drive/v3/files", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		files := []drive.File{}
		switch {
		case strings.Contains(q, "'"+f.FlyerFolderID+"'"):
			files = f.Flyers
		case strings.Contains(q, "'"+f.DocumentsFolderID+"'"):
			files = f.Documents
		}
		json.NewEncoder(w).Encode(map[string]any{"files": files})
	})
	mux.HandleFunc("GET /drive/v3/files/{id}", func(w http.ResponseWriter, r *http.Request) {
		if f.OnDownload != nil {
			f.OnDownload()
		}
		body, ok := f.Files[r.PathValue("id")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write(body)
	})
	mux.HandleFunc("GET /calendar/v3/calendars/{id}/events", func(w http.ResponseWriter, r *http.Request) {
		if f.FailCalendar {
			http.Error(w, "backend error", http.StatusInternalServerError)
			return
		}
		w.Write([]byte(f.EventsJSON))
	})
	mux.HandleFunc("GET /feeds/videos.xml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(f.FeedXML))
	})
	mux.HandleFunc("GET /vi/{id}/hqdefault.jpg", func(w http.ResponseWriter, r *http.Request) {
		body, ok := f.Thumbs[r.PathValue("id")]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Write(body)
	})
	return mux
}
