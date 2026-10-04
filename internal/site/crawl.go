package site

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
)

// writeCrawlerFiles writes sitemap.xml and robots.txt, so search engines find every page soon after the move.
func writeCrawlerFiles(outDir, baseURL string) error {
	var sitemap strings.Builder
	sitemap.WriteString(xml.Header + `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, p := range Pages {
		if p.Path == "/404" {
			continue
		}
		sitemap.WriteString("  <url><loc>")
		if err := xml.EscapeText(&sitemap, []byte(baseURL+p.Path)); err != nil {
			return err
		}
		sitemap.WriteString("</loc></url>\n")
	}
	sitemap.WriteString("</urlset>\n")
	if err := os.WriteFile(filepath.Join(outDir, "sitemap.xml"), []byte(sitemap.String()), 0o644); err != nil {
		return err
	}
	robots := "User-agent: *\nAllow: /\n\nSitemap: " + baseURL + "/sitemap.xml\n"
	return os.WriteFile(filepath.Join(outDir, "robots.txt"), []byte(robots), 0o644)
}
