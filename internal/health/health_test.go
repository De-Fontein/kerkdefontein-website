package health

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing_SuccessAndFailurePaths(t *testing.T) {
	var gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotPath, gotBody = r.URL.Path, string(b)
	}))
	defer srv.Close()

	if err := Ping(context.Background(), srv.Client(), srv.URL+"/uuid", false, "ok"); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/uuid" || gotBody != "ok" {
		t.Errorf("success ping = %s %q", gotPath, gotBody)
	}
	if err := Ping(context.Background(), srv.Client(), srv.URL+"/uuid", true, "skipped: Flyer.png"); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/uuid/fail" || gotBody != "skipped: Flyer.png" {
		t.Errorf("fail ping = %s %q", gotPath, gotBody)
	}
}

func TestPing_EmptyURLIsNoop(t *testing.T) {
	if err := Ping(context.Background(), http.DefaultClient, "", true, "x"); err != nil {
		t.Fatal(err)
	}
}
