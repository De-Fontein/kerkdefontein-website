package build

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/De-Fontein/kerkdefontein-website/internal/content"
	"github.com/De-Fontein/kerkdefontein-website/internal/fetch"
	"github.com/De-Fontein/kerkdefontein-website/internal/images"
	"github.com/De-Fontein/kerkdefontein-website/internal/youtube"
)

type media struct {
	src      Sources
	cacheDir string
	log      *slog.Logger
}

type preparedFlyer struct {
	flyer   content.Flyer
	pair    images.Pair
	fromDir string
}

type preparedDoc struct {
	doc  content.Document
	path string
}

type preparedVideo struct {
	video   youtube.Video
	thumb   images.Variant
	fromDir string
}

// sourceError marks a failed download: the source is unreachable, not the file broken, so the whole run
// aborts and retries instead of publishing a site with the file missing (spec §8).
type sourceError struct{ err error }

func (e sourceError) Error() string { return e.err.Error() }
func (e sourceError) Unwrap() error { return e.err }

type prepared struct {
	flyers []preparedFlyer
	docs   []preparedDoc
	video  preparedVideo
}

// prepare returns an error only for source failures; a file that downloads but cannot be converted
// becomes a Problem, so one broken flyer never blocks the others.
func (m media) prepare(ctx context.Context, flyers []content.Flyer, docs []content.Document, video youtube.Video) (prepared, []content.Problem, error) {
	var p prepared
	var problems []content.Problem
	for _, f := range flyers {
		dir := filepath.Join(m.cacheDir, "flyers", f.File.MD5)
		pair, err := m.converted(dir, func(tmp string) (images.Pair, error) {
			return m.convertDriveFile(ctx, tmp, f.File.ID, f.File.MimeType, f.File.MD5)
		})
		if errors.As(err, new(sourceError)) {
			return prepared{}, nil, fmt.Errorf("flyer %s: %w", f.File.Name, err)
		}
		if err != nil {
			problems = append(problems, content.Problem{Name: f.File.Name, Reason: err.Error()})
			continue
		}
		p.flyers = append(p.flyers, preparedFlyer{flyer: f, pair: pair, fromDir: dir})
	}
	for _, d := range docs {
		path, err := m.document(ctx, d)
		if err != nil {
			return prepared{}, nil, fmt.Errorf("document %s: %w", d.File.Name, err)
		}
		p.docs = append(p.docs, preparedDoc{doc: d, path: path})
	}
	p.video = m.videoThumb(ctx, video)
	return p, problems, nil
}

// converted caches by content hash: unchanged flyers are never downloaded or converted again, and a
// cache entry appears only complete (tmp dir renamed into place).
func (m media) converted(dir string, convert func(tmp string) (images.Pair, error)) (images.Pair, error) {
	if raw, err := os.ReadFile(filepath.Join(dir, "pair.json")); err == nil {
		var pair images.Pair
		if json.Unmarshal(raw, &pair) == nil {
			return pair, nil
		}
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return images.Pair{}, err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(dir), ".convert-")
	if err != nil {
		return images.Pair{}, err
	}
	defer os.RemoveAll(tmp)
	pair, err := convert(tmp)
	if err != nil {
		return images.Pair{}, err
	}
	raw, err := json.Marshal(pair)
	if err != nil {
		return images.Pair{}, err
	}
	if err := os.WriteFile(filepath.Join(tmp, "pair.json"), raw, 0o644); err != nil {
		return images.Pair{}, err
	}
	return pair, os.Rename(tmp, dir)
}

func (m media) convertDriveFile(ctx context.Context, tmp, fileID, mimeType, base string) (images.Pair, error) {
	src := filepath.Join(tmp, "source")
	out, err := os.Create(src)
	if err != nil {
		return images.Pair{}, err
	}
	err = m.src.Drive.Download(ctx, fileID, out)
	out.Close()
	if err != nil {
		return images.Pair{}, sourceError{err}
	}
	pair, err := images.Convert(ctx, src, mimeType, tmp, base)
	if err != nil {
		return images.Pair{}, fmt.Errorf("convert: %w", err)
	}
	os.Remove(src)
	return pair, nil
}

func (m media) document(ctx context.Context, d content.Document) (string, error) {
	path := filepath.Join(m.cacheDir, "documents", d.File.MD5+".pdf")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".download-")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	err = m.src.Drive.Download(ctx, d.File.ID, tmp)
	tmp.Close()
	if err != nil {
		return "", err
	}
	return path, os.Rename(tmp.Name(), path)
}

// videoThumb is best effort: without a thumbnail the play button still works on a dark background.
func (m media) videoThumb(ctx context.Context, v youtube.Video) preparedVideo {
	dir := filepath.Join(m.cacheDir, "video", v.ID)
	pair, err := m.converted(dir, func(tmp string) (images.Pair, error) {
		src := filepath.Join(tmp, "source.jpg")
		resp, err := fetch.Get(ctx, m.src.Web, m.src.ThumbBase+"/vi/"+v.ID+"/hqdefault.jpg")
		if err != nil {
			return images.Pair{}, err
		}
		defer resp.Body.Close()
		out, err := os.Create(src)
		if err != nil {
			return images.Pair{}, err
		}
		if _, err := out.ReadFrom(resp.Body); err != nil {
			out.Close()
			return images.Pair{}, err
		}
		out.Close()
		return images.Convert(ctx, src, "image/jpeg", tmp, "yt-"+v.ID)
	})
	if err != nil {
		m.log.Warn("video thumbnail unavailable", "video", v.ID, "err", err)
		return preparedVideo{video: v}
	}
	return preparedVideo{video: v, thumb: pair.Thumb, fromDir: dir}
}
