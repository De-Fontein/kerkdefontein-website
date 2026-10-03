// Package youtube reads the channel's public Atom feed; no API key is involved.
package youtube

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/De-Fontein/kerkdefontein-website/internal/fetch"
)

var ErrNoVideos = errors.New("youtube feed has no videos")

type Video struct {
	ID        string
	Title     string
	Published time.Time
}

type feed struct {
	Entries []struct {
		VideoID   string    `xml:"http://www.youtube.com/xml/schemas/2015 videoId"`
		Title     string    `xml:"http://www.w3.org/2005/Atom title"`
		Published time.Time `xml:"http://www.w3.org/2005/Atom published"`
	} `xml:"http://www.w3.org/2005/Atom entry"`
}

// Latest picks the newest entry by published time. Accepted limitation (spec §2): published is the creation
// time, so a Bible study uploaded after a Sunday stream was scheduled outranks it.
func Latest(ctx context.Context, hc *http.Client, baseURL, channelID string) (Video, error) {
	resp, err := fetch.Get(ctx, hc, baseURL+"/feeds/videos.xml?channel_id="+url.QueryEscape(channelID))
	if err != nil {
		return Video{}, fmt.Errorf("youtube feed: %w", err)
	}
	defer resp.Body.Close()

	var f feed
	if err := xml.NewDecoder(resp.Body).Decode(&f); err != nil {
		return Video{}, fmt.Errorf("parse youtube feed: %w", err)
	}
	var newest Video
	for _, e := range f.Entries {
		if e.VideoID != "" && e.Published.After(newest.Published) {
			newest = Video{ID: e.VideoID, Title: e.Title, Published: e.Published}
		}
	}
	if newest.ID == "" {
		return Video{}, ErrNoVideos
	}
	return newest, nil
}
