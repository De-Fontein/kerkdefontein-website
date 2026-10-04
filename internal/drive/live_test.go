package drive

import (
	"context"
	"os"
	"testing"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Probes the real Drive API: proves a service account sees file descriptions. Opt-in, needs credentials.
func TestLive_ServiceAccountSeesDescriptions(t *testing.T) {
	keyFile, folder := os.Getenv("FONTEIN_LIVE_KEY"), os.Getenv("FONTEIN_LIVE_FLYER_FOLDER")
	if keyFile == "" || folder == "" {
		t.Skip("set FONTEIN_LIVE_KEY and FONTEIN_LIVE_FLYER_FOLDER to run")
	}
	key, err := os.ReadFile(keyFile)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	creds, err := google.CredentialsFromJSON(ctx, key, "https://www.googleapis.com/auth/drive.readonly")
	if err != nil {
		t.Fatal(err)
	}
	client := Client{HTTP: oauth2.NewClient(ctx, creds.TokenSource), BaseURL: "https://www.googleapis.com"}
	files, err := client.List(ctx, folder)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if f.Description != "" {
			t.Logf("description visible on %q: %q", f.Name, f.Description)
			return
		}
	}
	t.Fatalf("none of %d files has a description; add one to a test flyer and rerun", len(files))
}
