// Package web embeds browser assets so that a deploy is a single binary.
package web

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"fmt"
	"io/fs"
)

//go:embed static
var Static embed.FS

// versions maps "css/site.css" to a short content hash. Assets are cached for a year, so their URL must change
// whenever their content does, or returning visitors run old scripts against new markup.
var versions = hashAll()

// AssetURL returns the cache-busted URL of an embedded asset, e.g. "/static/js/theme.js?v=1a2b3c4d5e".
func AssetURL(name string) (string, error) {
	v, ok := versions[name]
	if !ok {
		return "", fmt.Errorf("unknown asset %q", name)
	}
	return "/static/" + name + "?v=" + v, nil
}

func hashAll() map[string]string {
	hashes := map[string]string{}
	err := fs.WalkDir(Static, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := Static.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(b)
		hashes[path[len("static/"):]] = hex.EncodeToString(sum[:5])
		return nil
	})
	if err != nil {
		panic(err)
	}
	return hashes
}
