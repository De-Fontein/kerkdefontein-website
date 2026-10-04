// Package fetch holds the one HTTP GET shape shared by the Drive, Calendar and YouTube readers.
package fetch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func Get(ctx context.Context, hc *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("GET %s: unexpected status %s", req.URL.Redacted(), resp.Status)
	}
	return resp, nil
}

func JSON(ctx context.Context, hc *http.Client, url string, v any) error {
	resp, err := Get(ctx, hc, url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("decode %s: %w", url, err)
	}
	return nil
}
