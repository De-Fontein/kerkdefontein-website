package content

import (
	"fmt"
	"testing"
	"time"

	"github.com/De-Fontein/kerkdefontein-website/internal/drive"
)

func file(id, name, mime string, day int) drive.File {
	return drive.File{ID: id, Name: name, MimeType: mime, Size: 1024, MD5: "md5-" + id,
		CreatedTime: time.Date(2026, 10, day, 9, 0, 0, 0, time.UTC)}
}

func TestSelectFlyers_NewestSixAcceptedFilesWin(t *testing.T) {
	var files []drive.File
	for day := 1; day <= 8; day++ {
		files = append(files, file(fmt.Sprint(day), fmt.Sprintf("Flyer %d.png", day), "image/png", day))
	}
	got := SelectFlyers(files)
	if len(got.Shown) != MaxFlyers {
		t.Fatalf("shown %d, want %d", len(got.Shown), MaxFlyers)
	}
	if got.Shown[0].File.ID != "8" || got.Shown[5].File.ID != "3" {
		t.Errorf("order = %s..%s, want 8..3", got.Shown[0].File.ID, got.Shown[5].File.ID)
	}
	if got.Shown[0].Alt != "Flyer 8" {
		t.Errorf("alt = %q, want extension stripped", got.Shown[0].Alt)
	}
}

func TestSelectFlyers_SkipsUnsupportedAndOversizedFiles(t *testing.T) {
	big := file("big", "Groot.png", "image/png", 2)
	big.Size = MaxBytes + 1
	got := SelectFlyers([]drive.File{
		file("ok", "Goed.pdf", "application/pdf", 1),
		file("doc", "Notities", "application/vnd.google-apps.document", 3),
		big,
	})
	if len(got.Shown) != 1 || got.Shown[0].File.ID != "ok" {
		t.Errorf("shown = %+v", got.Shown)
	}
	want := map[string]bool{"Notities": true, "Groot.png": true}
	if len(got.Skipped) != 2 || !want[got.Skipped[0].Name] || !want[got.Skipped[1].Name] {
		t.Errorf("skipped = %+v", got.Skipped)
	}
}

func TestSelectFlyers_InvalidLinkKeepsFlyerWithoutLink(t *testing.T) {
	f := file("a", "Avond.png", "image/png", 1)
	f.Description = "javascript:alert(1)"
	got := SelectFlyers([]drive.File{f})
	if len(got.Shown) != 1 || got.Shown[0].Link != "" {
		t.Errorf("shown = %+v", got.Shown)
	}
	if len(got.BadLinks) != 1 || got.BadLinks[0].Name != "Avond.png" {
		t.Errorf("badLinks = %+v", got.BadLinks)
	}
}

func TestParseLink(t *testing.T) {
	cases := []struct {
		in       string
		wantLink string
		wantOK   bool
	}{
		{"", "", true},
		{"   ", "", true},
		{" https://kerkdefontein.nl/agenda/ ", "https://kerkdefontein.nl/agenda/", true},
		{"http://example.org", "", false},
		{"javascript:alert(1)", "", false},
		{"https://", "", false},
		{"zie https://example.org", "", false},
		{"https://exa mple.org", "", false},
	}
	for _, c := range cases {
		link, ok := ParseLink(c.in)
		if link != c.wantLink || ok != c.wantOK {
			t.Errorf("ParseLink(%q) = %q, %v; want %q, %v", c.in, link, ok, c.wantLink, c.wantOK)
		}
	}
}

func TestSelectDocuments_PDFsOnlySortedByTitle(t *testing.T) {
	docs, skipped := SelectDocuments([]drive.File{
		file("2", "Jaarverslag 2025.pdf", "application/pdf", 1),
		file("1", "Begroting 2026.pdf", "application/pdf", 2),
		file("3", "foto.jpg", "image/jpeg", 3),
	})
	if len(docs) != 2 || docs[0].Title != "Begroting 2026" || docs[1].Title != "Jaarverslag 2025" {
		t.Errorf("docs = %+v", docs)
	}
	if len(skipped) != 1 || skipped[0].Name != "foto.jpg" {
		t.Errorf("skipped = %+v", skipped)
	}
}

func TestTitleFromName_StripsOnlyKnownFileExtensions(t *testing.T) {
	cases := map[string]string{
		"Aanbiddingsavond vr 19.30":     "Aanbiddingsavond vr 19.30",
		"Aanbiddingsavond vr 19.30.PNG": "Aanbiddingsavond vr 19.30",
		"Jaarverslag 2025.pdf":          "Jaarverslag 2025",
		"foto.jpeg":                     "foto",
		".pdf":                          "Flyer",
	}
	for in, want := range cases {
		if got := titleFromName(in); got != want {
			t.Errorf("titleFromName(%q) = %q, want %q", in, got, want)
		}
	}
}
