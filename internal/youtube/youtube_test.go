package youtube

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLatest_PicksNewestPublishedEntry(t *testing.T) {
	feed, err := os.ReadFile("testdata/feed.xml")
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/feeds/videos.xml" || r.URL.Query().Get("channel_id") != "UCbQP" {
			http.NotFound(w, r)
			return
		}
		w.Write(feed)
	}))
	defer srv.Close()

	v, err := Latest(context.Background(), srv.Client(), srv.URL, "UCbQP")
	if err != nil {
		t.Fatal(err)
	}
	if v.ID != "qeWRxvxJvNs" || v.Title != "Vrolijk zijn in Gods tent! | Leander Janse | 04-10-2026" {
		t.Errorf("got %+v", v)
	}
}

func TestLatest_EmptyFeedIsErrNoVideos(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<feed xmlns="http://www.w3.org/2005/Atom"><title>x</title></feed>`))
	}))
	defer srv.Close()

	_, err := Latest(context.Background(), srv.Client(), srv.URL, "UCbQP")
	if !errors.Is(err, ErrNoVideos) {
		t.Fatalf("err = %v, want ErrNoVideos", err)
	}
}
