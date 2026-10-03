// Package health reports each run to healthchecks.io; a missing ping for 15 minutes emails the webmaster.
package health

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

func Ping(ctx context.Context, hc *http.Client, checkURL string, failed bool, message string) error {
	if checkURL == "" {
		return nil
	}
	url := checkURL
	if failed {
		url += "/fail"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(message))
	if err != nil {
		return err
	}
	resp, err := hc.Do(req)
	if err != nil {
		return fmt.Errorf("healthcheck ping: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck ping: status %s", resp.Status)
	}
	return nil
}
