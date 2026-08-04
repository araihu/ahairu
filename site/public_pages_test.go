package site

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func renderComponent(t *testing.T, component templ.Component) string {
	t.Helper()
	var output strings.Builder
	if err := component.Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	return output.String()
}

func TestBrandDownloadsResolveToPinnedReleaseFiles(t *testing.T) {
	catalog, err := BrandCatalog()
	if err != nil {
		t.Fatal(err)
	}
	byPath := make(map[string]CatalogAsset, len(catalog.Assets))
	for _, asset := range catalog.Assets {
		byPath[asset.Path] = asset
	}
	checksums, err := ReleaseChecksums()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/brand/", "/pt-br/brand/", "/es/brand/"} {
		page := pageForTest(t, path)
		if got, want := len(page.Brand.Downloads), 8; got != want {
			t.Fatalf("%s download count = %d, want %d", path, got, want)
		}
		for _, download := range page.Brand.Downloads {
			if strings.TrimSpace(download.Label) == "" || strings.TrimSpace(download.Detail) == "" {
				t.Errorf("%s has meaningless download metadata: %#v", path, download)
			}
			if !strings.HasPrefix(download.URL, BrandAssetsPublicPrefix) {
				t.Errorf("%s download escapes release prefix: %q", path, download.URL)
				continue
			}
			releasePath := strings.TrimPrefix(download.URL, BrandAssetsPublicPrefix)
			data, err := fs.ReadFile(BrandAssets(), releasePath)
			if err != nil {
				t.Errorf("%s download %q missing: %v", path, releasePath, err)
				continue
			}
			if releasePath == "checksums.txt" {
				continue
			}
			got := sha256.Sum256(data)
			want := checksums[releasePath]
			if want == "" || hex.EncodeToString(got[:]) != want {
				t.Errorf("%s download %q does not match release checksum", path, releasePath)
			}
			if strings.HasPrefix(releasePath, "brand/") {
				asset, ok := byPath[releasePath]
				if !ok || asset.SHA256 != want {
					t.Errorf("%s download %q lacks matching catalog record", path, releasePath)
				}
			}
		}
	}
}

func TestLocalizedChromeHasNoEnglishAccessibilityFallback(t *testing.T) {
	for _, path := range []string{"/pt-br/brand/", "/es/license/"} {
		component, err := PageComponent(pageForTest(t, path))
		if err != nil {
			t.Fatal(err)
		}
		html := renderComponent(t, component)
		for _, unwanted := range []string{`aria-label="Language"`, `aria-label="Primary navigation"`, `aria-label="Authoritative version"`, `mailto:brand@araihu.com`} {
			if strings.Contains(html, unwanted) {
				t.Errorf("%s contains unlocalized or unverified markup %q", path, unwanted)
			}
		}
	}
}

func TestBrandAndLicenseMetadataDescriptionsAreLocalized(t *testing.T) {
	wants := map[string]string{
		"/pt-br/brand/":   "sistema de identidade prático",
		"/pt-br/license/": "ativos de identidade Arai Hû",
		"/es/brand/":      "sistema de identidad práctico",
		"/es/license/":    "activos de identidad Arai Hû",
	}
	for path, want := range wants {
		page := pageForTest(t, path)
		if !strings.Contains(page.Meta.Description, want) {
			t.Errorf("%s description = %q, want localized phrase %q", path, page.Meta.Description, want)
		}
	}
}

func TestSharedChromeProjectsAndMinimumSizeUseAdaptiveAssets(t *testing.T) {
	if araihuIconURL != BrandAssetsPublicPrefix+"icons/brand/araihu-icon-adaptive-transparent-optical.svg" {
		t.Fatalf("shared chrome icon = %q, want adaptive asset", araihuIconURL)
	}
	for _, home := range Locales() {
		for _, project := range home.Projects {
			if !strings.Contains(project.MarkURL, "-icon-adaptive-transparent-optical.svg") {
				t.Errorf("%s project mark = %q, want adaptive asset", project.Name, project.MarkURL)
			}
			if strings.Contains(project.MarkURL, "-icon-light-transparent-optical.svg") {
				t.Errorf("%s project mark remains hardcoded light asset", project.Name)
			}
		}
	}

	brandHTML := renderComponent(t, BrandPage(pageForTest(t, "/brand/")))
	adaptiveIcon := BrandAssetsPublicPrefix + "icons/brand/araihu-icon-adaptive-transparent-optical.svg"
	if got, want := strings.Count(brandHTML, adaptiveIcon), 4; got != want {
		t.Errorf("brand page adaptive chrome/minimum specimens = %d, want %d", got, want)
	}
	if strings.Contains(brandHTML, "icons/brand/araihu-icon-light-transparent-optical.svg") {
		t.Error("brand page retains hardcoded light icon outside explicit light specimen surface")
	}
	if got, want := strings.Count(brandHTML, "brand/araihu/logo/light-transparent-optical.svg"), 3; got != want {
		t.Errorf("brand page explicit light logo specimen/download references = %d, want %d", got, want)
	}

	homeHTML := renderComponent(t, HomePage(pageForTest(t, "/en/")))
	for _, product := range []string{"araihu", "goshtoso", "manja", "paje", "x9"} {
		want := product + "-icon-adaptive-transparent-optical.svg"
		if !strings.Contains(homeHTML, want) {
			t.Errorf("home page misses adaptive %s mark", product)
		}
	}
}

func pageForTest(t *testing.T, path string) Page {
	t.Helper()
	for _, page := range Pages() {
		if page.Meta.Path == path {
			return page
		}
	}
	t.Fatalf("page %q not found", path)
	return Page{}
}

func TestEveryTypedPageHasStaticRendererAndSharedChrome(t *testing.T) {
	for _, page := range Pages() {
		component, err := PageComponent(page)
		if err != nil {
			t.Fatalf("PageComponent(%q): %v", page.Meta.Path, err)
		}
		html := renderComponent(t, component)
		for _, want := range []string{
			`<html lang="` + page.Meta.Locale.Language + `"`,
			`class="skip-link" href="#main-content"`,
			`class="ahairu-header"`,
			`href="` + pathFor(PageHome, page.Meta.Locale.Language) + `"`,
			`href="` + pathFor(PageBrand, page.Meta.Locale.Language) + `"`,
			`href="` + pathFor(PageLicense, page.Meta.Locale.Language) + `"`,
			`<main id="main-content"`,
			`class="ahairu-shell ahairu-footer`,
		} {
			if !strings.Contains(html, want) {
				t.Errorf("page %q misses shared chrome %q", page.Meta.Path, want)
			}
		}
		if strings.Count(html, `<h1`) != 1 {
			t.Errorf("page %q has %d h1 elements, want 1", page.Meta.Path, strings.Count(html, `<h1`))
		}
		if page.Meta.Kind == PageHome {
			for _, runtime := range []string{"htmx", "alpine"} {
				if !strings.Contains(strings.ToLower(html), runtime) {
					t.Errorf("interactive home page %q misses client runtime %q", page.Meta.Path, runtime)
				}
			}
		} else {
			for _, unwanted := range []string{"htmx", "alpine"} {
				if strings.Contains(strings.ToLower(html), unwanted) {
					t.Errorf("static auxiliary page %q unexpectedly loads client runtime %q", page.Meta.Path, unwanted)
				}
			}
		}
	}
}

func TestBrandPageContainsGuidanceDownloadsAndLicenseLink(t *testing.T) {
	page := pageForTest(t, "/brand/")
	html := renderComponent(t, BrandPage(page))
	for _, text := range []string{
		"Minimum size", "Clear space", "Incorrect use", "Download",
		"catalog.json", "checksums.txt", "sprite.svg", "Goshtoso", "Attribution",
	} {
		if !strings.Contains(html, text) {
			t.Errorf("brand page misses %q", text)
		}
	}
	for _, appearance := range []string{"Light", "Dark", "Monochrome", "Tinted"} {
		if !strings.Contains(html, appearance) {
			t.Errorf("brand page misses %s variant", appearance)
		}
	}
	if !strings.Contains(html, `href="/license/"`) {
		t.Error("brand page misses prominent canonical license link")
	}
	if !strings.Contains(html, BrandAssetsPublicPrefix+"icons/brand/sprite.svg#araihu-icon-monochrome-transparent-optical") {
		t.Error("brand page does not use generic Goshtoso icon component against brand sprite")
	}
	if strings.Count(html, ` download`) < 5 {
		t.Errorf("brand page has %d download links, want at least 5", strings.Count(html, ` download`))
	}
}

func TestBrandGuidanceIsLocalized(t *testing.T) {
	tests := []struct {
		path string
		want []string
	}{
		{path: "/pt-br/brand/", want: []string{"Tamanho mínimo", "Área de proteção", "Usos incorretos", "Baixar"}},
		{path: "/es/brand/", want: []string{"Tamaño mínimo", "Área de protección", "Usos incorrectos", "Descargar"}},
	}
	for _, test := range tests {
		html := renderComponent(t, BrandPage(pageForTest(t, test.path)))
		for _, want := range test.want {
			if !strings.Contains(html, want) {
				t.Errorf("%s misses localized guidance %q", test.path, want)
			}
		}
	}
}

func TestLicenseSeparatesBrandTermsAndHeroiconsMIT(t *testing.T) {
	html := renderComponent(t, LicensePage(pageForTest(t, "/license/")))
	for _, text := range []string{
		"Arai Hû brand terms", "Heroicons MIT license",
		"Unmodified integration", "documentation", "notices", "No endorsement",
		"Modified marks", "Standalone redistribution", "Merchandise", "another identity", "implied affiliation",
	} {
		if !strings.Contains(html, text) {
			t.Errorf("license page misses %q", text)
		}
	}
	for _, markup := range []string{`<dt>Version</dt><dd>1.0</dd>`, `<dt>Effective</dt><dd><time datetime="2026-07-29">29 July 2026</time>`} {
		if !strings.Contains(html, markup) {
			t.Errorf("license page misses visible release metadata %q", markup)
		}
	}
}

func TestLocalizedLicenseShowsEnglishAuthorityNoticeBeforeTerms(t *testing.T) {
	tests := []struct {
		path, notice string
	}{
		{path: "/pt-br/license/", notice: "A versão em inglês rege estes termos"},
		{path: "/es/license/", notice: "La versión en inglés rige estos términos"},
	}
	for _, test := range tests {
		html := renderComponent(t, LicensePage(pageForTest(t, test.path)))
		notice := strings.Index(html, test.notice)
		terms := strings.Index(html, `data-license-terms`)
		if notice < 0 || terms < 0 || notice > terms {
			t.Errorf("%s authority notice must precede terms", test.path)
		}
		if !strings.Contains(html, `href="/license/"`) {
			t.Errorf("%s authority notice misses canonical English link", test.path)
		}
	}
}
