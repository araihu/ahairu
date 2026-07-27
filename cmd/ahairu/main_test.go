package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	for _, asset := range []string{
		"logos/araihu-favicon.svg",
		"logos/araihu-mark.svg",
		"logos/goshtoso-mark.svg",
		"logos/manja-mark.svg",
		"logos/paje-mark.svg",
		">X-9</span>",
	} {
		if !strings.Contains(page, asset) {
			t.Errorf("generated HTML misses brand asset %q", asset)
		}
	}
	if info, err := os.Stat(filepath.Join("public", "assets", "styles.css")); err != nil || info.Size() == 0 {
		t.Fatalf("generated stylesheet missing or empty: %v", err)
	}
	if info, err := os.Stat(filepath.Join("public", "assets", "ahairu.css")); err != nil || info.Size() == 0 {
		t.Fatalf("brand stylesheet missing or empty: %v", err)
	}
	for _, locale := range []string{"pt-br", "es"} {
		if _, err := os.Stat(filepath.Join("public", locale, "index.html")); err != nil {
			t.Fatalf("localized page missing for %s: %v", locale, err)
		}
	}
}
