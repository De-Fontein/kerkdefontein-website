// Package web embeds browser assets so that a deploy is a single binary.
package web

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
)

//go:embed static
var Static embed.FS

// CSSVersion changes whenever the stylesheet does, so browsers may cache it for a year.
var CSSVersion = hashOf("static/css/site.css")

func hashOf(path string) string {
	b, err := Static.ReadFile(path)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:5])
}
