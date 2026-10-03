package fake

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"time"

	"github.com/De-Fontein/kerkdefontein-website/internal/drive"
)

const (
	SampleFlyerFolder     = "flyers"
	SampleDocumentsFolder = "documenten"
	SampleChannel         = "UCsample"
)

// Sample builds believable content relative to now, so the preview always shows upcoming events.
func Sample(now time.Time) (Fixtures, error) {
	f := Fixtures{FlyerFolderID: SampleFlyerFolder, DocumentsFolderID: SampleDocumentsFolder,
		Files: map[string][]byte{}, Thumbs: map[string][]byte{}}

	colours := []color.RGBA{{0x05, 0x5d, 0x75, 0xff}, {0xe0, 0xd3, 0xbb, 0xff}, {0xc6, 0xa7, 0x8e, 0xff}}
	names := []string{"Aanbiddingsavond vr 19.30.png", "Bijbelstudie Jakobus.png", "Gemeentekamp 2027.png"}
	for i, name := range names {
		body, err := encodePNG(800, 1131, colours[i])
		if err != nil {
			return Fixtures{}, err
		}
		id := fmt.Sprintf("flyer%d", i+1)
		f.Files[id] = body
		file := drive.File{ID: id, Name: name, MimeType: "image/png", Size: int64(len(body)),
			MD5: md5Hex(body), CreatedTime: now.Add(-time.Duration(i) * 24 * time.Hour)}
		if i == 0 {
			file.Description = "https://kerkdefontein.nl/agenda/"
		}
		f.Flyers = append(f.Flyers, file)
	}

	pdf := []byte("%PDF-1.4\n1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj\n2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj\n3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 595 842]>>endobj\ntrailer<</Root 1 0 R>>\n%%EOF\n")
	f.Files["doc1"] = pdf
	f.Documents = []drive.File{{ID: "doc1", Name: "Jaarverslag 2025.pdf", MimeType: "application/pdf",
		Size: int64(len(pdf)), MD5: md5Hex(pdf), CreatedTime: now}}

	sunday := now.AddDate(0, 0, (7-int(now.Weekday()))%7)
	day := func(d time.Time, hour int) string {
		return time.Date(d.Year(), d.Month(), d.Day(), hour, 0, 0, 0, time.UTC).Format(time.RFC3339)
	}
	f.EventsJSON = fmt.Sprintf(`{"items":[
	 {"id":"ev1","summary":"Samenkomst","location":"de Hoeksteen","start":{"dateTime":%q},"end":{"dateTime":%q}},
	 {"id":"ev2","summary":"Bijbelstudie","start":{"dateTime":%q},"end":{"dateTime":%q}},
	 {"id":"ev3","summary":"Gemeentekamp","start":{"date":%q},"end":{"date":%q}}]}`,
		day(sunday, 8), day(sunday, 10), day(sunday.AddDate(0, 0, 3), 17), day(sunday.AddDate(0, 0, 3), 19),
		sunday.AddDate(0, 0, 13).Format("2006-01-02"), sunday.AddDate(0, 0, 15).Format("2006-01-02"))

	f.FeedXML = fmt.Sprintf(`<feed xmlns:yt="http://www.youtube.com/xml/schemas/2015" xmlns="http://www.w3.org/2005/Atom">
	 <entry><yt:videoId>sample1</yt:videoId><title>Voorbeelddienst</title><published>%s</published></entry></feed>`,
		now.Add(-48*time.Hour).Format(time.RFC3339))
	thumb, err := encodeJPEG(480, 360, colours[0])
	if err != nil {
		return Fixtures{}, err
	}
	f.Thumbs["sample1"] = thumb
	return f, nil
}

func encodePNG(w, h int, c color.RGBA) ([]byte, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, solid(w, h, c))
	return buf.Bytes(), err
}

func encodeJPEG(w, h int, c color.RGBA) ([]byte, error) {
	var buf bytes.Buffer
	err := jpeg.Encode(&buf, solid(w, h, c), nil)
	return buf.Bytes(), err
}

func solid(w, h int, c color.RGBA) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = c.R, c.G, c.B, c.A
	}
	return img
}

func md5Hex(b []byte) string {
	sum := md5.Sum(b)
	return hex.EncodeToString(sum[:])
}
