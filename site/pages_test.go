package site

import "testing"

func TestPagesContainNineCanonicalLocalizedPages(t *testing.T) {
	pages := Pages()
	if len(pages) != 9 {
		t.Fatalf("Pages() returned %d pages, want 9", len(pages))
	}
	for _, want := range []struct {
		path     string
		language string
		kind     PageKind
	}{
		{"/en/", "en", PageHome},
		{"/pt-br/", "pt-BR", PageHome},
		{"/es/", "es", PageHome},
		{"/brand/", "en", PageBrand},
		{"/pt-br/brand/", "pt-BR", PageBrand},
		{"/es/brand/", "es", PageBrand},
		{"/license/", "en", PageLicense},
		{"/pt-br/license/", "pt-BR", PageLicense},
		{"/es/license/", "es", PageLicense},
	} {
		requirePage(t, pages, want.path, want.language, want.kind)
	}
}

func TestLocaleNavigationPreservesPageKind(t *testing.T) {
	page := requirePage(t, Pages(), "/pt-br/brand/", "pt-BR", PageBrand)
	requireLocaleLink(t, page.Navigation, "es", "/es/brand/")
}

func TestPagesUseCanonicalURLsAndReciprocalAlternates(t *testing.T) {
	wantPaths := map[PageKind]map[string]string{
		PageHome: {
			"en": "/en/", "pt-BR": "/pt-br/", "es": "/es/", "x-default": "/en/",
		},
		PageBrand: {
			"en": "/brand/", "pt-BR": "/pt-br/brand/", "es": "/es/brand/", "x-default": "/brand/",
		},
		PageLicense: {
			"en": "/license/", "pt-BR": "/pt-br/license/", "es": "/es/license/", "x-default": "/license/",
		},
	}
	for _, page := range Pages() {
		if page.Meta.CanonicalURL != "https://araihu.com"+page.Meta.Path {
			t.Errorf("%s canonical URL = %q, want %q", page.Meta.Path, page.Meta.CanonicalURL, "https://araihu.com"+page.Meta.Path)
		}
		for language, path := range wantPaths[page.Meta.Kind] {
			requireAlternate(t, page.Meta.Alternates, language, "https://araihu.com"+path)
		}
	}
}

func requirePage(t *testing.T, pages []Page, path, language string, kind PageKind) Page {
	t.Helper()
	for _, page := range pages {
		if page.Meta.Path == path && page.Meta.Locale.Language == language && page.Meta.Kind == kind {
			return page
		}
	}
	t.Fatalf("page %s (%s, %s) missing", path, language, kind)
	return Page{}
}

func requireLocaleLink(t *testing.T, navigation Navigation, language, path string) {
	t.Helper()
	for _, link := range navigation.Locales {
		if link.Locale.Language == language && link.URL == path {
			return
		}
	}
	t.Fatalf("locale navigation link %s to %s missing", language, path)
}

func requireAlternate(t *testing.T, alternates []Alternate, language, url string) {
	t.Helper()
	for _, alternate := range alternates {
		if alternate.Language == language && alternate.URL == url {
			return
		}
	}
	t.Fatalf("alternate %s to %s missing", language, url)
}
