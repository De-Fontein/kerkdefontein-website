package images

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func requireTools(t *testing.T) {
	t.Helper()
	for _, tool := range []string{"vips", "vipsheader", "pdftoppm"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s not installed", tool)
		}
	}
}

func writePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{0x05, 0x5d, 0x75, 0xff})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

// A one-page blank A4 PDF; poppler rebuilds the missing xref table, which keeps the fixture readable.
const blankPDF = "%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\n3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 595 842]>>endobj\ntrailer<</Root 1 0 R>>\n%%EOF\n"

func TestConvert_PNGNeverUpscales(t *testing.T) {
	requireTools(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "in.png")
	writePNG(t, src, 800, 1131)

	set, err := Convert(context.Background(), src, "image/png", dir, "flyer")
	if err != nil {
		t.Fatal(err)
	}
	if set.Thumb != (Variant{File: "flyer-360.webp", Width: 360, Height: 509}) {
		t.Errorf("thumb = %+v", set.Thumb)
	}
	if set.Thumb2x != (Variant{File: "flyer-720.webp", Width: 720, Height: 1018}) {
		t.Errorf("thumb2x = %+v", set.Thumb2x)
	}
	if set.Large != (Variant{File: "flyer-1600.webp", Width: 800, Height: 1131}) {
		t.Errorf("large = %+v (must not upscale)", set.Large)
	}
	for _, f := range []string{"flyer-360.webp", "flyer-720.webp", "flyer-1600.webp"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Error(err)
		}
	}
}

func TestConvert_PDFUsesFirstPage(t *testing.T) {
	requireTools(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "in.pdf")
	if err := os.WriteFile(src, []byte(blankPDF), 0o644); err != nil {
		t.Fatal(err)
	}

	set, err := Convert(context.Background(), src, "application/pdf", dir, "doc")
	if err != nil {
		t.Fatal(err)
	}
	if set.Thumb.Width != 360 || set.Thumb.Height <= set.Thumb.Width {
		t.Errorf("thumb = %+v, want portrait 360 wide", set.Thumb)
	}
}

func TestConvert_CorruptFileIsAnError(t *testing.T) {
	requireTools(t)
	dir := t.TempDir()
	src := filepath.Join(dir, "bad.png")
	if err := os.WriteFile(src, []byte("not an image"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Convert(context.Background(), src, "image/png", dir, "bad"); err == nil {
		t.Fatal("expected error")
	}
}
