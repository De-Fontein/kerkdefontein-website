package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func site(t *testing.T) http.Handler {
	t.Helper()
	root := t.TempDir()
	os.WriteFile(filepath.Join(root, "404.html"), []byte("niet gevonden"), 0o644)
	os.WriteFile(filepath.Join(root, "index.html"), []byte("home"), 0o644)
	os.WriteFile(filepath.Join(root, "index.html.br"), []byte("BROTLI"), 0o644)
	return serveSite(root)
}

func get(h http.Handler, path, acceptEncoding string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if acceptEncoding != "" {
		req.Header.Set("Accept-Encoding", acceptEncoding)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestServeSite_MissingPathServes404PageWithStatus(t *testing.T) {
	rec := get(site(t), "/bestaat-niet", "")
	if rec.Code != http.StatusNotFound || rec.Body.String() != "niet gevonden" {
		t.Errorf("got %d %q", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Security-Policy") == "" {
		t.Error("preview must send the production CSP so e2e tests catch violations")
	}
}

func TestServeSite_ServesPlainFileWithoutBrotliSupport(t *testing.T) {
	rec := get(site(t), "/", "gzip")
	if rec.Code != http.StatusOK || rec.Body.String() != "home" || rec.Header().Get("Content-Encoding") != "" {
		t.Errorf("got %d %q %q", rec.Code, rec.Body.String(), rec.Header().Get("Content-Encoding"))
	}
}

// Mirrors Caddy's `precompressed br`, so e2e page-weight budgets measure what visitors download.
func TestServeSite_ServesPrecompressedBrotli(t *testing.T) {
	rec := get(site(t), "/", "gzip, deflate, br")
	if rec.Body.String() != "BROTLI" || rec.Header().Get("Content-Encoding") != "br" {
		t.Errorf("got %q encoding %q", rec.Body.String(), rec.Header().Get("Content-Encoding"))
	}
	if rec.Header().Get("Content-Type") != "text/html; charset=utf-8" || rec.Header().Get("Vary") != "Accept-Encoding" {
		t.Errorf("headers = %v", rec.Header())
	}
}
