package build

import (
	"compress/gzip"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

var compressible = map[string]bool{".html": true, ".css": true, ".js": true, ".svg": true, ".ics": true}

// precompress writes .br and .gz next to every text file, so Caddy serves Brotli (which stock Caddy cannot
// produce on the fly) and the CPU cost is paid once per build instead of per request.
func precompress(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !compressible[filepath.Ext(path)] {
			return err
		}
		if out, err := exec.Command("brotli", "-q", "11", "-k", "-f", path).CombinedOutput(); err != nil {
			return fmt.Errorf("brotli %s: %w: %s", path, err, out)
		}
		return gzipFile(path)
	})
}

func gzipFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out, err := os.OpenFile(path+".gz", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	zw, err := gzip.NewWriterLevel(out, gzip.BestCompression)
	if err != nil {
		out.Close()
		return err
	}
	if _, err := zw.Write(data); err != nil {
		out.Close()
		return err
	}
	if err := zw.Close(); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
