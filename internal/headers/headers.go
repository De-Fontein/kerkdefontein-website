// Package headers keeps the CSP in one place for the preview server; deploy/Caddyfile repeats it verbatim.
package headers

const CSP = "default-src 'self'; img-src 'self'; style-src 'self'; script-src 'self'; " +
	"frame-src https://www.youtube-nocookie.com; base-uri 'none'; form-action 'none'; frame-ancestors 'none'"
