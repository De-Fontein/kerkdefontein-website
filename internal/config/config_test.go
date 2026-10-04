package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad_ValidConfigAppliesDefaults(t *testing.T) {
	path := writeConfig(t, `{
		"flyerFolderId": "f1", "documentsFolderId": "d1", "calendarId": "c1",
		"youtubeChannelId": "UCx", "credentialsFile": "/etc/fontein/sa.json",
		"siteRoot": "/srv/site", "cacheDir": "/var/cache/fontein"
	}`)

	c, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if c.FlyerFolderID != "f1" || c.SiteRoot != "/srv/site" {
		t.Errorf("fields not loaded: %+v", c)
	}
	if c.GoogleAPIBase != "https://www.googleapis.com" || c.YouTubeBase != "https://www.youtube.com" || c.ThumbnailBase != "https://i.ytimg.com" {
		t.Errorf("defaults not applied: %+v", c)
	}
	if c.BaseURL != "https://kerkdefontein.nl" {
		t.Errorf("BaseURL default = %q", c.BaseURL)
	}
}

func TestLoad_MissingRequiredFieldsAreAllReported(t *testing.T) {
	_, err := Load(writeConfig(t, `{"flyerFolderId": "f1"}`))
	if err == nil {
		t.Fatal("expected error")
	}
	for _, name := range []string{"documentsFolderId", "calendarId", "youtubeChannelId", "credentialsFile", "siteRoot", "cacheDir"} {
		if !strings.Contains(err.Error(), name) {
			t.Errorf("error %q does not mention %s", err, name)
		}
	}
}

func TestLoad_UnknownFieldIsRejected(t *testing.T) {
	_, err := Load(writeConfig(t, `{"flyerFolder": "typo"}`))
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}
