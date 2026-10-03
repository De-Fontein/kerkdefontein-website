// Package config loads the server-only settings. Secrets and IDs never live in the repository.
package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type Config struct {
	FlyerFolderID     string `json:"flyerFolderId"`
	DocumentsFolderID string `json:"documentsFolderId"`
	CalendarID        string `json:"calendarId"`
	CalendarICalURL   string `json:"calendarIcalUrl"`
	YouTubeChannelID  string `json:"youtubeChannelId"`
	CredentialsFile   string `json:"credentialsFile"`
	HealthcheckURL    string `json:"healthcheckUrl"`
	SiteRoot          string `json:"siteRoot"`
	CacheDir          string `json:"cacheDir"`
	BaseURL           string `json:"baseUrl"`
	GoogleAPIBase     string `json:"googleApiBase"`
	YouTubeBase       string `json:"youtubeBase"`
	ThumbnailBase     string `json:"thumbnailBase"`
}

func Load(path string) (Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	var c Config
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	c.applyDefaults()
	return c, c.validate()
}

func (c *Config) applyDefaults() {
	if c.BaseURL == "" {
		c.BaseURL = "https://kerkdefontein.nl"
	}
	if c.GoogleAPIBase == "" {
		c.GoogleAPIBase = "https://www.googleapis.com"
	}
	if c.YouTubeBase == "" {
		c.YouTubeBase = "https://www.youtube.com"
	}
	if c.ThumbnailBase == "" {
		c.ThumbnailBase = "https://i.ytimg.com"
	}
}

func (c Config) validate() error {
	required := []struct{ name, value string }{
		{"flyerFolderId", c.FlyerFolderID},
		{"documentsFolderId", c.DocumentsFolderID},
		{"calendarId", c.CalendarID},
		{"youtubeChannelId", c.YouTubeChannelID},
		{"credentialsFile", c.CredentialsFile},
		{"siteRoot", c.SiteRoot},
		{"cacheDir", c.CacheDir},
	}
	var missing []error
	for _, field := range required {
		if field.value == "" {
			missing = append(missing, fmt.Errorf("config: %s is required", field.name))
		}
	}
	return errors.Join(missing...)
}
