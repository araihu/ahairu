package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/araihu/ahairu/site"
)

func TestBuildWritesStandaloneSite(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PWD", t.TempDir())
	if err := os.Chdir(os.Getenv("PWD")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	if err := build(); err != nil {
		t.Fatal(err)
	}
	html, err := os.ReadFile(filepath.Join("public", "en", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	page := string(html)
	if !strings.Contains(page, "Independent software") || !strings.Contains(page, "/assets/styles.css") {
		t.Fatalf("generated HTML misses site content or stylesheet: %s", html)
	}
	for _, landmark := range []string{
		`class="skip-link" href="#main-content"`,
		`<header class="ahairu-header">`,
		`<meta name="theme-color" content="#07111f">`,
		`width="64" height="64"> <span>Arai Hû</span>`,
		`<span>Arai Hû</span>`,
		`<main id="main-content" class="ahairu-shell mx-auto px-6 md:px-10" tabindex="-1">`,
		`href="/en/" aria-current="page"`,
		`aria-label="Primary navigation"`,
		`<h3 class="project-name">X-9</h3>`,
	} {
		if !strings.Contains(page, landmark) {
			t.Errorf("generated HTML misses accessibility landmark %q", landmark)
		}
	}
	for _, asset := range []string{
		"logos/araihu-icon-background.svg?rev=a8a9647a",
		"logos/araihu-icon-transparent.svg?rev=a8a9647a",
		"logos/goshtoso-icon-transparent.svg?rev=a8a9647a",
		"logos/manja-icon-transparent.svg?rev=a8a9647a",
		"logos/paje-icon-transparent.svg?rev=a8a9647a",
		"logos/x9-icon-transparent.svg?rev=a8a9647a",
	} {
		if !strings.Contains(page, asset) {
			t.Errorf("generated HTML misses brand asset %q", asset)
		}
	}
	if !strings.Contains(page, `href="https://x9.araihu.com"`) {
		t.Error("generated HTML misses the X-9 product URL")
	}
	if strings.Contains(page, "xisnove.dev") {
		t.Error("generated HTML still links to the retired Xisnove domain")
	}
	if info, err := os.Stat(filepath.Join("public", "assets", "styles.css")); err != nil || info.Size() == 0 {
		t.Fatalf("generated stylesheet missing or empty: %v", err)
	}
	if info, err := os.Stat(filepath.Join("public", "assets", "ahairu.css")); err != nil || info.Size() == 0 {
		t.Fatalf("brand stylesheet missing or empty: %v", err)
	}
	if info, err := os.Stat(filepath.Join("public", "assets", "araihu-theme.css")); err != nil || info.Size() == 0 {
		t.Fatalf("brand theme missing or empty: %v", err)
	}
	for _, name := range site.BrandAssetNames() {
		if info, err := os.Stat(filepath.Join("public", "assets", "logos", name)); err != nil || info.Size() == 0 {
			t.Errorf("brand asset %s missing or empty: %v", name, err)
		}
	}
	for locale, skipLabel := range map[string]string{"pt-br": "Pular para o conteúdo", "es": "Saltar al contenido"} {
		localizedHTML, err := os.ReadFile(filepath.Join("public", locale, "index.html"))
		if err != nil {
			t.Fatalf("localized page missing for %s: %v", locale, err)
		}
		if !strings.Contains(string(localizedHTML), skipLabel) || !strings.Contains(string(localizedHTML), `aria-current="page"`) {
			t.Errorf("localized page %s misses skip link or locale state", locale)
		}
	}
}
