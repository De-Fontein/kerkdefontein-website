package web

import (
	"io/fs"
	"testing"
)

func TestStaticContainsRequiredAssets(t *testing.T) {
	for _, path := range []string{"static/css/site.css", "static/js/youtube.js", "static/js/theme.js", "static/img/logo.svg", "static/img/share.jpg"} {
		if _, err := fs.Stat(Static, path); err != nil {
			t.Errorf("%s: %v (run `npx tsc` and add the images)", path, err)
		}
	}
}

func TestBudgets(t *testing.T) {
	// The CSS size budget is checked compressed, with each page, in internal/build (spec: 14 KB on the wire).
	url, err := AssetURL("js/theme.js")
	if err != nil || len(url) != len("/static/js/theme.js?v=")+10 {
		t.Errorf("AssetURL(js/theme.js) = %q, %v", url, err)
	}
	if _, err := AssetURL("js/missing.js"); err == nil {
		t.Error("unknown asset must be an error, so a template typo fails the build")
	}
	js, err := fs.ReadFile(Static, "static/js/youtube.js")
	if err != nil {
		t.Fatal(err)
	}
	if len(js) > 10*1024 {
		t.Errorf("youtube.js is %d bytes, budget 10 KB", len(js))
	}
}
