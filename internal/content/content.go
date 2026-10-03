// Package content holds the volunteer-facing rules: which Drive files become flyers and documents.
package content

import (
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/De-Fontein/kerkdefontein-website/internal/drive"
)

const (
	MaxFlyers = 6
	MaxBytes  = 25 << 20
)

var (
	flyerTypes    = map[string]bool{"image/jpeg": true, "image/png": true, "image/webp": true, "application/pdf": true}
	documentTypes = map[string]bool{"application/pdf": true}
)

type Flyer struct {
	File drive.File
	Alt  string
	Link string
}

type Document struct {
	File  drive.File
	Title string
}

type Problem struct {
	Name   string
	Reason string
}

type FlyerSelection struct {
	Shown    []Flyer
	Skipped  []Problem
	BadLinks []Problem
}

func SelectFlyers(files []drive.File) FlyerSelection {
	var sel FlyerSelection
	var accepted []drive.File
	for _, f := range files {
		if reason := rejectReason(f, flyerTypes); reason != "" {
			sel.Skipped = append(sel.Skipped, Problem{Name: f.Name, Reason: reason})
			continue
		}
		accepted = append(accepted, f)
	}
	sort.SliceStable(accepted, func(i, j int) bool { return accepted[i].CreatedTime.After(accepted[j].CreatedTime) })
	if len(accepted) > MaxFlyers {
		accepted = accepted[:MaxFlyers]
	}
	for _, f := range accepted {
		link, ok := ParseLink(f.Description)
		if !ok {
			sel.BadLinks = append(sel.BadLinks, Problem{Name: f.Name, Reason: "description is not a single https:// link"})
		}
		sel.Shown = append(sel.Shown, Flyer{File: f, Alt: titleFromName(f.Name), Link: link})
	}
	return sel
}

func SelectDocuments(files []drive.File) ([]Document, []Problem) {
	var docs []Document
	var skipped []Problem
	for _, f := range files {
		if reason := rejectReason(f, documentTypes); reason != "" {
			skipped = append(skipped, Problem{Name: f.Name, Reason: reason})
			continue
		}
		docs = append(docs, Document{File: f, Title: titleFromName(f.Name)})
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].Title < docs[j].Title })
	return docs, skipped
}

// ParseLink accepts an empty description (no link) or exactly one absolute https URL; anything else is
// rejected so a typo or a hostile description can never become a javascript: or http: link on the site.
func ParseLink(description string) (string, bool) {
	s := strings.TrimSpace(description)
	if s == "" {
		return "", true
	}
	if strings.ContainsAny(s, " \t\n") {
		return "", false
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return "", false
	}
	return s, true
}

func rejectReason(f drive.File, allowed map[string]bool) string {
	if !allowed[f.MimeType] {
		return "unsupported file type " + f.MimeType
	}
	if f.Size > MaxBytes {
		return "larger than 25 MB"
	}
	return ""
}

// Only real file extensions are stripped: a volunteer naming a flyer "Avond 19.30" must keep the time.
var fileExtensions = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true, ".pdf": true}

func titleFromName(name string) string {
	title := name
	if ext := path.Ext(name); fileExtensions[strings.ToLower(ext)] {
		title = strings.TrimSuffix(name, ext)
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return "Flyer"
	}
	return title
}
