package site

import (
	"context"
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestRenderedPagesEnrollCampaignCanaryWithFixedImages(t *testing.T) {
	for _, page := range Pages() {
		t.Run(page.Meta.Path, func(t *testing.T) {
			component, err := PageComponent(page)
			if err != nil {
				t.Fatal(err)
			}
			var output strings.Builder
			if err := component.Render(context.Background(), &output); err != nil {
				t.Fatal(err)
			}
			htmlText := output.String()
			for _, want := range []string{
				`data-theme="araihu"`,
				`data-theme-source="default"`,
				`src="/assets/campaign/v1.js"`,
				`data-channel="/assets/releases/current"`,
				`integrity="sha384-oPH7l1vK9vKP1Dn+18sO3yEXlz4ts6KzPEQl0SW4Y/+im05gOaamNNaQAf6bGH/n"`,
				`crossorigin="anonymous"`,
				`data-campaign-toggle`,
				`data-campaign-toggle-icon`,
			} {
				if !strings.Contains(htmlText, want) {
					t.Errorf("rendered page misses %s", want)
				}
			}
			if runtime, styles := strings.Index(htmlText, `src="/assets/campaign/v1.js"`), strings.LastIndex(htmlText, `rel="stylesheet"`); runtime < styles {
				t.Error("campaign runtime precedes baseline styles")
			}
			if source, body := strings.Index(htmlText, `data-theme-source="default"`), strings.Index(htmlText, "<body"); source < 0 || source > body {
				t.Error("theme source is not available before body content")
			}

			document, err := html.Parse(strings.NewReader(htmlText))
			if err != nil {
				t.Fatal(err)
			}
			var hooks brandHookCounts
			walkRenderedElements(t, document, &hooks)
			wantLogoHooks := 0
			if page.Meta.Kind == PageBrand {
				wantLogoHooks = 1
			}
			if hooks.logo != wantLogoHooks || hooks.icon != 1 {
				t.Errorf("campaign brand hooks = logo:%d icon:%d, want logo:%d icon:1", hooks.logo, hooks.icon, wantLogoHooks)
			}
			requireCampaignToggle(t, hooks.campaignToggles)
		})
	}
}

type brandHookCounts struct {
	logo            int
	icon            int
	campaignToggles []*html.Node
}

func walkRenderedElements(t *testing.T, node *html.Node, hooks *brandHookCounts) {
	t.Helper()
	if node.Type == html.ElementNode {
		attributes := make(map[string]string, len(node.Attr))
		for _, attribute := range node.Attr {
			attributes[attribute.Key] = attribute.Val
		}
		if node.Data == "img" && (attributes["width"] == "" || attributes["height"] == "") {
			t.Errorf("replaceable image %q lacks fixed width and height", attributes["src"])
		}
		switch attributes["data-asset-brand"] {
		case "":
		case "logo":
			if node.Data != "img" || attributes["src"] == "" {
				t.Errorf("logo hook must target img[src]")
			} else {
				hooks.logo++
			}
		case "icon":
			if node.Data != "link" || attributes["href"] == "" {
				t.Errorf("icon hook must target link[href]")
			} else if attributes["rel"] != "icon" {
				t.Errorf("icon hook must target rel=icon link")
			} else {
				hooks.icon++
			}
		default:
			t.Errorf("invalid asset brand hook %q", attributes["data-asset-brand"])
		}
		if _, toggle := attributes["data-campaign-toggle"]; toggle {
			hooks.campaignToggles = append(hooks.campaignToggles, node)
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		walkRenderedElements(t, child, hooks)
	}
}

func requireCampaignToggle(t *testing.T, toggles []*html.Node) {
	t.Helper()
	if len(toggles) != 1 {
		t.Errorf("campaign toggle count = %d, want 1", len(toggles))
		return
	}
	toggle := toggles[0]
	attributes := renderedAttributes(toggle)
	if toggle.Data != "button" {
		t.Errorf("campaign toggle is <%s>, want <button>", toggle.Data)
	}
	if _, hidden := attributes["hidden"]; !hidden {
		t.Error("campaign toggle lacks hidden attribute")
	}
	if attributes["type"] != "button" {
		t.Errorf("campaign toggle type = %q, want button", attributes["type"])
	}
	if attributes["aria-pressed"] != "false" {
		t.Errorf("campaign toggle aria-pressed = %q, want false", attributes["aria-pressed"])
	}
	var iconChildren int
	var srOnlyText strings.Builder
	for child := toggle.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}
		childAttributes := renderedAttributes(child)
		if _, icon := childAttributes["data-campaign-toggle-icon"]; icon {
			iconChildren++
		}
		if hasRenderedClass(childAttributes["class"], "sr-only") {
			srOnlyText.WriteString(renderedText(child))
		}
	}
	if iconChildren != 1 {
		t.Errorf("campaign toggle icon children = %d, want 1", iconChildren)
	}
	if strings.TrimSpace(attributes["aria-label"]) == "" && strings.TrimSpace(srOnlyText.String()) == "" {
		t.Error("campaign toggle lacks accessible name")
	}
}

func renderedAttributes(node *html.Node) map[string]string {
	attributes := make(map[string]string, len(node.Attr))
	for _, attribute := range node.Attr {
		attributes[attribute.Key] = attribute.Val
	}
	return attributes
}

func hasRenderedClass(value, want string) bool {
	for _, class := range strings.Fields(value) {
		if class == want {
			return true
		}
	}
	return false
}

func renderedText(node *html.Node) string {
	var text strings.Builder
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current.Type == html.TextNode {
			text.WriteString(current.Data)
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return text.String()
}

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

func TestHomePagesSelectExactlyValidatedRenderableHomes(t *testing.T) {
	homePages, err := HomePages()
	if err != nil {
		t.Fatal(err)
	}
	if len(homePages) != 3 {
		t.Fatalf("HomePages() returned %d pages, want 3", len(homePages))
	}
	for _, page := range homePages {
		if page.Meta.Kind != PageHome || page.Home == nil || page.Brand != nil || page.License != nil {
			t.Errorf("HomePages() returned non-renderable home page: %#v", page)
		}
	}

	for _, page := range Pages() {
		if page.Meta.Kind == PageBrand && page.Brand == nil {
			t.Errorf("brand page %s is not modeled", page.Meta.Path)
		}
		if page.Meta.Kind == PageLicense && page.License == nil {
			t.Errorf("license page %s is not modeled", page.Meta.Path)
		}
	}
}

func TestPagesUseCanonicalURLsAlternatesAndNavigation(t *testing.T) {
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
	pages := Pages()
	byPath := make(map[string]Page, len(pages))
	canonicals := make(map[string]struct{}, len(pages))
	titles := make(map[string]struct{}, len(pages))
	for _, page := range pages {
		if _, duplicate := byPath[page.Meta.Path]; duplicate {
			t.Errorf("duplicate path %q", page.Meta.Path)
		}
		byPath[page.Meta.Path] = page
		if _, duplicate := canonicals[page.Meta.CanonicalURL]; duplicate {
			t.Errorf("duplicate canonical URL %q", page.Meta.CanonicalURL)
		}
		canonicals[page.Meta.CanonicalURL] = struct{}{}
		if _, duplicate := titles[page.Meta.Title]; duplicate {
			t.Errorf("duplicate title %q", page.Meta.Title)
		}
		titles[page.Meta.Title] = struct{}{}
	}
	for _, page := range pages {
		if page.Meta.CanonicalURL != "https://araihu.com"+page.Meta.Path {
			t.Errorf("%s canonical URL = %q, want %q", page.Meta.Path, page.Meta.CanonicalURL, "https://araihu.com"+page.Meta.Path)
		}
		if len(page.Meta.Alternates) != 4 {
			t.Errorf("%s has %d alternates, want 4", page.Meta.Path, len(page.Meta.Alternates))
		}
		seenAlternates := make(map[string]struct{}, len(page.Meta.Alternates))
		for language, path := range wantPaths[page.Meta.Kind] {
			requireAlternate(t, page.Meta.Alternates, language, "https://araihu.com"+path)
		}
		for _, alternate := range page.Meta.Alternates {
			if _, duplicate := seenAlternates[alternate.Language]; duplicate {
				t.Errorf("%s repeats alternate language %q", page.Meta.Path, alternate.Language)
			}
			seenAlternates[alternate.Language] = struct{}{}
			alternatePath := strings.TrimPrefix(alternate.URL, "https://araihu.com")
			target, ok := byPath[alternatePath]
			if !ok {
				t.Errorf("%s alternate %q targets unknown path %q", page.Meta.Path, alternate.Language, alternatePath)
				continue
			}
			if target.Meta.CanonicalURL != alternate.URL {
				t.Errorf("%s alternate %q does not target canonical URL", page.Meta.Path, alternate.Language)
			}
			if alternate.Language != "x-default" {
				requireAlternate(t, target.Meta.Alternates, page.Meta.Locale.Language, page.Meta.CanonicalURL)
			}
		}
		if len(page.Navigation.Locales) != 3 {
			t.Errorf("%s has %d locale links, want 3", page.Meta.Path, len(page.Navigation.Locales))
		}
		for _, language := range []string{"en", "pt-BR", "es"} {
			requireLocaleLink(t, page.Navigation, language, wantPaths[page.Meta.Kind][language])
		}
		seenLocaleLinks := make(map[string]struct{}, len(page.Navigation.Locales))
		for _, link := range page.Navigation.Locales {
			if _, duplicate := seenLocaleLinks[link.Locale.Language]; duplicate {
				t.Errorf("%s repeats navigation locale %q", page.Meta.Path, link.Locale.Language)
			}
			seenLocaleLinks[link.Locale.Language] = struct{}{}
			target, ok := byPath[link.URL]
			if !ok {
				t.Errorf("%s navigation targets unknown path %q", page.Meta.Path, link.URL)
				continue
			}
			if target.Meta.Kind != page.Meta.Kind {
				t.Errorf("%s navigation changes kind from %q to %q", page.Meta.Path, page.Meta.Kind, target.Meta.Kind)
			}
			if target.Meta.Locale.Language != link.Locale.Language {
				t.Errorf("%s navigation locale %q targets %q", page.Meta.Path, link.Locale.Language, target.Meta.Locale.Language)
			}
		}
		requireOnlyMatchingContent(t, page)
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

func requireOnlyMatchingContent(t *testing.T, page Page) {
	t.Helper()
	present := 0
	if page.Home != nil {
		present++
	}
	if page.Brand != nil {
		present++
	}
	if page.License != nil {
		present++
	}
	if present != 1 {
		t.Errorf("%s has %d content models, want 1", page.Meta.Path, present)
	}
	switch page.Meta.Kind {
	case PageHome:
		if page.Home == nil {
			t.Errorf("%s home page lacks home content", page.Meta.Path)
		}
	case PageBrand:
		if page.Brand == nil {
			t.Errorf("%s brand page lacks brand content", page.Meta.Path)
		}
	case PageLicense:
		if page.License == nil {
			t.Errorf("%s license page lacks license content", page.Meta.Path)
		}
	}
}
