package drive

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestList_FollowsPagesAndExcludesFoldersInQuery(t *testing.T) {
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query().Get("q"))
		if r.URL.Query().Get("pageToken") == "" {
			w.Write([]byte(`{"nextPageToken":"p2","files":[{"id":"a","name":"Aanbidding.png","description":"https://example.org","mimeType":"image/png","createdTime":"2026-10-01T10:00:00Z","md5Checksum":"m1","size":"2048"}]}`))
			return
		}
		w.Write([]byte(`{"files":[{"id":"b","name":"Studie.pdf","mimeType":"application/pdf","createdTime":"2026-10-02T10:00:00Z","md5Checksum":"m2","size":"4096"}]}`))
	}))
	defer srv.Close()

	files, err := Client{HTTP: srv.Client(), BaseURL: srv.URL}.List(context.Background(), "folder123")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d files, want 2", len(files))
	}
	want := File{ID: "a", Name: "Aanbidding.png", Description: "https://example.org", MimeType: "image/png",
		CreatedTime: time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC), MD5: "m1", Size: 2048}
	if files[0] != want {
		t.Errorf("files[0] = %+v, want %+v", files[0], want)
	}
	q := queries[0]
	if !strings.Contains(q, "'folder123' in parents") || !strings.Contains(q, "mimeType != 'application/vnd.google-apps.folder'") || !strings.Contains(q, "trashed = false") {
		t.Errorf("query %q misses a filter", q)
	}
}

func TestDownload_CopiesBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/drive/v3/files/abc" || r.URL.Query().Get("alt") != "media" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte("PNGDATA"))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	if err := (Client{HTTP: srv.Client(), BaseURL: srv.URL}).Download(context.Background(), "abc", &buf); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "PNGDATA" {
		t.Errorf("body = %q", buf.String())
	}
}
