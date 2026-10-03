package build

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/De-Fontein/kerkdefontein-website/internal/calendar"
	"github.com/De-Fontein/kerkdefontein-website/internal/ics"
	"github.com/De-Fontein/kerkdefontein-website/internal/site"
	"github.com/De-Fontein/kerkdefontein-website/web"
)

// Google IDs are already filename-safe; the check guards the filesystem against a surprise.
var safeID = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func writeSite(dir string, s Settings, p prepared, events []calendar.Event, now time.Time) error {
	if err := copyFS(web.Static, "static", filepath.Join(dir, "static")); err != nil {
		return err
	}
	data := site.Data{BaseURL: s.BaseURL, CalendarICalURL: s.CalendarICalURL}

	for _, f := range p.flyers {
		for _, v := range []string{f.pair.Thumb.File, f.pair.Large.File} {
			if err := copyFile(filepath.Join(f.fromDir, v), filepath.Join(dir, "media", v)); err != nil {
				return err
			}
		}
		data.Flyers = append(data.Flyers, site.FlyerView{ID: f.flyer.File.ID, Alt: f.flyer.Alt, Link: f.flyer.Link,
			Thumb: f.pair.Thumb, Large: f.pair.Large})
	}
	for _, d := range p.docs {
		if !safeID.MatchString(d.doc.File.ID) {
			continue
		}
		name := d.doc.File.ID + ".pdf"
		if err := copyFile(d.path, filepath.Join(dir, "documenten", name)); err != nil {
			return err
		}
		data.Documents = append(data.Documents, site.DocumentView{Title: d.doc.Title, Href: "/documenten/" + name})
	}
	if p.video.video.ID != "" {
		data.Video = &site.VideoView{ID: p.video.video.ID, Title: p.video.video.Title, Thumb: p.video.thumb}
		if p.video.thumb.File != "" {
			if err := copyFile(filepath.Join(p.video.fromDir, p.video.thumb.File), filepath.Join(dir, "media", p.video.thumb.File)); err != nil {
				return err
			}
		}
	}
	for _, e := range events {
		if !safeID.MatchString(e.ID) {
			continue
		}
		href := "/agenda/" + e.ID + ".ics"
		if err := writeFile(filepath.Join(dir, "agenda", e.ID+".ics"), ics.Event(e, now)); err != nil {
			return err
		}
		data.Events = append(data.Events, site.EventView{Event: e, ICSHref: href})
	}
	if err := site.Render(dir, data); err != nil {
		return err
	}
	return precompress(dir)
}

func copyFS(fsys fs.FS, root, dst string) error {
	return fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		b, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		return writeFile(filepath.Join(dst, rel), b)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func writeFile(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
