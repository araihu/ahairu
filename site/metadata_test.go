package site

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func TestEveryPageHasCompleteAbsoluteMetadata(t *testing.T) {
	pages := Pages()
	titles := make(map[string]struct{}, len(pages))
	descriptions := make(map[string]struct{}, len(pages))
	for _, page := range pages {
		meta := page.Meta
		for name, value := range map[string]string{
			"canonical URL": meta.CanonicalURL,
			"title":         meta.Title,
			"description":   meta.Description,
			"social image":  meta.SocialImageURL,
			"robots":        meta.Robots,
		} {
			if value == "" {
				t.Errorf("%s %s is empty", meta.Path, name)
			}
		}
		if !strings.HasPrefix(meta.CanonicalURL, CanonicalSiteURL+"/") {
			t.Errorf("%s canonical URL is not absolute: %q", meta.Path, meta.CanonicalURL)
		}
		if !strings.HasPrefix(meta.SocialImageURL, CanonicalSiteURL+"/") {
			t.Errorf("%s social image URL is not absolute: %q", meta.Path, meta.SocialImageURL)
		}
		if _, exists := titles[meta.Title]; exists {
			t.Errorf("duplicate title %q", meta.Title)
		}
		titles[meta.Title] = struct{}{}
		if _, exists := descriptions[meta.Description]; exists {
			t.Errorf("duplicate description %q", meta.Description)
		}
		descriptions[meta.Description] = struct{}{}
	}
}

func TestLayoutUsesMetadataForDocumentShell(t *testing.T) {
	page := requirePage(t, Pages(), "/en/", "en", PageHome)
	var output bytes.Buffer
	child := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := w.Write([]byte(`<main id="main-content">content</main>`))
		return err
	})
	if err := Layout(page, child).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	html := output.String()
	for _, fragment := range []string{
		`<!doctype html>`, `<html lang="en" data-theme="araihu">`, `<title>` + page.Meta.Title + `</title>`,
		`<main id="main-content">content</main>`, `<meta property="og:image" content="` + page.Meta.SocialImageURL + `">`,
	} {
		if !strings.Contains(html, fragment) {
			t.Errorf("layout misses %q", fragment)
		}
	}
}

func TestMetadataRendersLocalizedDiscoveryTagsAndSafeJSONLD(t *testing.T) {
	for _, page := range Pages() {
		html := renderMetadata(t, page)
		for _, fragment := range []string{
			`<link rel="canonical" href="` + page.Meta.CanonicalURL + `">`,
			`<meta property="og:image" content="` + page.Meta.SocialImageURL + `">`,
			`<meta name="twitter:card" content="summary_large_image">`,
			`<meta property="og:locale" content="` + page.Meta.Locale.OGLocale + `">`,
			`<link rel="manifest" href="/site.webmanifest">`,
			`<link rel="apple-touch-icon" href="/assets/logos/araihu-icon-background.svg?rev=a8a9647a">`,
			`<script id="structured-data" type="application/ld+json">`,
		} {
			if !strings.Contains(html, fragment) {
				t.Errorf("%s metadata misses %q", page.Meta.Path, fragment)
			}
		}
		for _, alternate := range page.Meta.Alternates {
			fragment := `hreflang="` + alternate.Language + `" href="` + alternate.URL + `"`
			if !strings.Contains(html, fragment) {
				t.Errorf("%s metadata misses alternate %q", page.Meta.Path, fragment)
			}
		}
		jsonLD := scriptJSON(t, html)
		if !json.Valid([]byte(jsonLD)) {
			t.Errorf("%s JSON-LD is invalid: %s", page.Meta.Path, jsonLD)
		}
	}
}

func renderMetadata(t *testing.T, page Page) string {
	t.Helper()
	var output bytes.Buffer
	if err := Metadata(page).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	return output.String()
}

func scriptJSON(t *testing.T, html string) string {
	t.Helper()
	const start = `<script id="structured-data" type="application/ld+json">`
	_, content, found := strings.Cut(html, start)
	if !found {
		t.Fatal("JSON-LD script missing")
	}
	content, found = strings.CutSuffix(content, "\n</script>")
	if !found {
		t.Fatal("JSON-LD script is not safely closed")
	}
	return content
}
