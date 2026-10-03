// Package site renders the static pages. All volunteer text passes through html/template escaping.
package site

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/De-Fontein/kerkdefontein-website/internal/calendar"
	"github.com/De-Fontein/kerkdefontein-website/internal/images"
)

type FlyerView struct {
	ID, Alt, Link         string
	Thumb, Thumb2x, Large images.Variant
}

type VideoView struct {
	ID, Title string
	Thumb     images.Variant
}

type EventView struct {
	calendar.Event
	ICSHref string
}

type DocumentView struct {
	Title, Href string
}

type Data struct {
	BaseURL         string
	CalendarICalURL string
	Video           *VideoView
	Flyers          []FlyerView
	Events          []EventView
	Documents       []DocumentView
}

type Page struct {
	Path, Template, Title, Description string
}

type NavItem struct {
	Path, Label string
}

var Pages = []Page{
	{"/", "home", "Baptistengemeente De Fontein Middelburg", "Samenkomst elke zondag om 10:00 in de Hoeksteen, Middelburg. Je bent van harte welkom!"},
	{"/eerste-keer/", "eerste-keer", "Eerste keer? · De Fontein", "Wat je kunt verwachten als je voor het eerst een samenkomst van De Fontein bezoekt."},
	{"/samenkomsten/", "samenkomsten", "Samenkomsten · De Fontein", "Elke zondag om 10:00 komen we samen in de Hoeksteen in Middelburg."},
	{"/activiteiten/", "activiteiten", "Activiteiten · De Fontein", "Bijbelstudies, kinderen, jeugd, kringen, gebed, evangelisatie en pastoraat bij De Fontein."},
	{"/agenda/", "agenda", "Agenda · De Fontein", "Alle komende samenkomsten en activiteiten van De Fontein."},
	{"/over-ons/", "over-ons", "Over ons · De Fontein", "Wie we zijn, waar we voor staan, onze missie en waarden."},
	{"/over-ons/israel/", "israel", "Israël · De Fontein", "Waarom De Fontein zich verbonden voelt met Israël."},
	{"/over-ons/baptisme/", "baptisme", "Ontstaan van het Baptisme · De Fontein", "Een beknopte geschiedenis van de baptistengemeenten, van de Jordaan tot De Fontein in Middelburg."},
	{"/doneren/", "doneren", "Doneren · De Fontein", "Steun het werk van Baptistengemeente De Fontein. ANBI-gegevens en jaarstukken."},
	{"/verhuur/", "verhuur", "Verhuur · De Fontein", "Informatie over het gebruik van gebouw de Hoeksteen."},
	{"/privacy/", "privacy", "Privacy · De Fontein", "Hoe De Fontein omgaat met je gegevens op deze website."},
	{"/404", "notfound", "Pagina niet gevonden · De Fontein", "Deze pagina bestaat niet (meer)."},
}

var Nav = []NavItem{
	{"/eerste-keer/", "Eerste keer?"},
	{"/samenkomsten/", "Samenkomsten"},
	{"/activiteiten/", "Activiteiten"},
	{"/agenda/", "Agenda"},
	{"/over-ons/", "Over ons"},
	{"/doneren/", "Doneren"},
	{"/verhuur/", "Verhuur"},
}

//go:embed templates
var templateFS embed.FS

type view struct {
	Page   Page
	Data   Data
	Nav    []NavItem
	JSONLD template.JS
}

func Render(outDir string, d Data) error {
	for _, p := range Pages {
		if err := renderPage(outDir, p, d); err != nil {
			return fmt.Errorf("render %s: %w", p.Path, err)
		}
	}
	return nil
}

func renderPage(outDir string, p Page, d Data) error {
	tmpl, err := template.New("").Funcs(funcs).ParseFS(templateFS,
		"templates/layout.html", "templates/partials/*.html", "templates/pages/"+p.Template+".html")
	if err != nil {
		return err
	}
	ld, err := jsonLD(p, d)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	v := view{Page: p, Data: d, Nav: Nav, JSONLD: ld}
	if err := tmpl.ExecuteTemplate(&buf, "layout", v); err != nil {
		return err
	}
	path := outputPath(outDir, p.Path)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

func outputPath(outDir, pagePath string) string {
	switch pagePath {
	case "/":
		return filepath.Join(outDir, "index.html")
	case "/404":
		return filepath.Join(outDir, "404.html")
	}
	return filepath.Join(outDir, filepath.FromSlash(strings.Trim(pagePath, "/")), "index.html")
}
