package fetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSON_DecodesBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"name":"fontein"}`))
	}))
	defer srv.Close()

	var v struct{ Name string }
	if err := JSON(context.Background(), srv.Client(), srv.URL, &v); err != nil {
		t.Fatal(err)
	}
	if v.Name != "fontein" {
		t.Errorf("Name = %q", v.Name)
	}
}

func TestGet_NonOKStatusIsAnErrorWithTheStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	}))
	defer srv.Close()

	_, err := Get(context.Background(), srv.Client(), srv.URL)
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("expected 403 error, got %v", err)
	}
}
