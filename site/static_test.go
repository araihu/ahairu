package site

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"
)

func TestSitemapHasNineCanonicalPagesAndNoRedirects(t *testing.T) {
	data, err := Sitemap(Pages())
	if err != nil {
		t.Fatal(err)
	}
	var document sitemapDocument
	if err := xml.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if len(document.URLs) != 9 {
		t.Fatalf("sitemap has %d URLs, want 9", len(document.URLs))
	}
	seen := make(map[string]struct{}, len(document.URLs))
	for _, entry := range document.URLs {
		if !strings.HasPrefix(entry.Location, CanonicalSiteURL+"/") {
			t.Errorf("sitemap URL is not absolute: %q", entry.Location)
		}
		if strings.Contains(entry.Location, "/en/brand") || strings.Contains(entry.Location, "/en/license") {
			t.Errorf("sitemap includes redirect URL %q", entry.Location)
		}
		if _, duplicate := seen[entry.Location]; duplicate {
			t.Errorf("sitemap repeats %q", entry.Location)
		}
		seen[entry.Location] = struct{}{}
	}
	for _, page := range Pages() {
		if _, found := seen[page.Meta.CanonicalURL]; !found {
			t.Errorf("sitemap misses %q", page.Meta.CanonicalURL)
		}
	}
}

func TestStaticDiscoveryFilesAreAbsoluteAndUsePublishedIcons(t *testing.T) {
	if got, want := string(Robots()), "User-agent: *\nAllow: /\nSitemap: https://araihu.com/sitemap.xml\n"; got != want {
		t.Errorf("robots = %q, want %q", got, want)
	}
	var manifest map[string]any
	if err := json.Unmarshal(SiteManifest(), &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest["name"] != "Arai Hû" || manifest["start_url"] != "/en/" {
		t.Errorf("manifest identity = %#v", manifest)
	}
	icons, ok := manifest["icons"].([]any)
	if !ok || len(icons) == 0 {
		t.Fatalf("manifest icons = %#v", manifest["icons"])
	}
	for _, raw := range icons {
		icon, ok := raw.(map[string]any)
		if !ok || !strings.HasPrefix(icon["src"].(string), "/assets/logos/") {
			t.Errorf("manifest icon is not published local asset: %#v", raw)
		}
	}
}
