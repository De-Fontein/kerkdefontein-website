package headers

import (
	"os"
	"strings"
	"testing"
)

func TestCSP_AllowsTheScipioGivingFrame(t *testing.T) {
	if !strings.Contains(CSP, "frame-src https://www.youtube-nocookie.com https://referral.socie.nl;") {
		t.Errorf("CSP must allow framing YouTube and Scipio only: %s", CSP)
	}
}

// The preview server sends CSP and Caddy sends its own copy; e2e tests only prove the preview's.
func TestCSP_CaddyfileSendsTheSamePolicy(t *testing.T) {
	caddyfile, err := os.ReadFile("../../deploy/Caddyfile")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(caddyfile), `Content-Security-Policy "`+CSP+`"`) {
		t.Error("deploy/Caddyfile CSP differs from headers.CSP")
	}
}
