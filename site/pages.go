package site

import "fmt"

var siteLocales = []Locale{
	{Key: "en", Language: "en", OGLocale: "en_US", Label: "EN"},
	{Key: "pt-br", Language: "pt-BR", OGLocale: "pt_BR", Label: "PT"},
	{Key: "es", Language: "es", OGLocale: "es_ES", Label: "ES"},
}

// Pages returns every canonical localized public page.
func Pages() []Page {
	pages := make([]Page, 0, len(siteLocales)*3)
	for _, kind := range []PageKind{PageHome, PageBrand, PageLicense} {
		for _, locale := range siteLocales {
			pages = append(pages, newPage(kind, locale))
		}
	}
	return pages
}

// HomePages returns the only pages the current builder can render. Brand and
// license pages remain modeled, but their renderers are introduced in Task 4.
func HomePages() ([]Page, error) { return PagesByKind(PageHome) }

// PagesByKind returns the complete, validated localized set for one page kind.
func PagesByKind(kind PageKind) ([]Page, error) {
	if kind != PageHome && kind != PageBrand && kind != PageLicense {
		return nil, fmt.Errorf("unknown page kind %q", kind)
	}

	pages := Pages()
	selected := make([]Page, 0, len(siteLocales))
	for _, page := range pages {
		if err := validatePage(page); err != nil {
			return nil, err
		}
		if page.Meta.Kind == kind {
			selected = append(selected, page)
		}
	}
	if len(selected) != len(siteLocales) {
		return nil, fmt.Errorf("page kind %q has %d locales, want %d", kind, len(selected), len(siteLocales))
	}
	return selected, nil
}

func validatePage(page Page) error {
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
		return fmt.Errorf("page %q has %d content models, want 1", page.Meta.Path, present)
	}
	switch page.Meta.Kind {
	case PageHome:
		if page.Home == nil {
			return fmt.Errorf("home page %q lacks home content", page.Meta.Path)
		}
	case PageBrand:
		if page.Brand == nil {
			return fmt.Errorf("brand page %q lacks brand content", page.Meta.Path)
		}
	case PageLicense:
		if page.License == nil {
			return fmt.Errorf("license page %q lacks license content", page.Meta.Path)
		}
	default:
		return fmt.Errorf("page %q has unknown kind %q", page.Meta.Path, page.Meta.Kind)
	}
	return nil
}

func newPage(kind PageKind, locale Locale) Page {
	path := pathFor(kind, locale.Language)
	page := Page{
		Meta: PageMeta{
			Kind:         kind,
			Locale:       locale,
			Path:         path,
			CanonicalURL: CanonicalSiteURL + path,
			Title:        pageTitle(kind, locale),
			Description:  pageDescription(kind, locale),
			Robots:       "index,follow",
			Alternates:   alternatesFor(kind),
		},
		Navigation: navigationFor(kind),
	}

	switch kind {
	case PageHome:
		content := homeContent(locale.Key)
		page.Home = &content
	case PageBrand:
		page.Brand = &BrandContent{Heading: localizedHeading(kind, locale.Key)}
	case PageLicense:
		page.License = &LicenseContent{Heading: localizedHeading(kind, locale.Key)}
	}
	return page
}

func pathFor(kind PageKind, language string) string {
	localeKey := language
	if localeKey == "pt-BR" {
		localeKey = "pt-br"
	}
	if localeKey == "x-default" {
		localeKey = "en"
	}

	prefix := "/" + localeKey
	if localeKey == "en" && kind != PageHome {
		prefix = ""
	}
	switch kind {
	case PageHome:
		return prefix + "/"
	case PageBrand:
		return prefix + "/brand/"
	case PageLicense:
		return prefix + "/license/"
	default:
		return ""
	}
}

func alternatesFor(kind PageKind) []Alternate {
	return []Alternate{
		{Language: "en", URL: CanonicalSiteURL + pathFor(kind, "en")},
		{Language: "pt-BR", URL: CanonicalSiteURL + pathFor(kind, "pt-BR")},
		{Language: "es", URL: CanonicalSiteURL + pathFor(kind, "es")},
		{Language: "x-default", URL: CanonicalSiteURL + pathFor(kind, "x-default")},
	}
}

func navigationFor(kind PageKind) Navigation {
	links := make([]LocaleLink, 0, len(siteLocales))
	for _, locale := range siteLocales {
		links = append(links, LocaleLink{Locale: locale, URL: pathFor(kind, locale.Language)})
	}
	return Navigation{Locales: links}
}

func pageTitle(kind PageKind, locale Locale) string {
	return localizedHeading(kind, locale.Key) + " | Arai Hû"
}

func pageDescription(kind PageKind, locale Locale) string {
	switch kind {
	case PageHome:
		return homeContent(locale.Key).Intro
	case PageBrand:
		return localizedHeading(kind, locale.Key) + " for Arai Hû."
	case PageLicense:
		return localizedHeading(kind, locale.Key) + " for Arai Hû."
	default:
		return ""
	}
}

func localizedHeading(kind PageKind, localeKey string) string {
	if kind == PageHome {
		return homeContent(localeKey).Tagline
	}
	headings := map[PageKind]map[string]string{
		PageBrand: {
			"en": "Brand guidance", "pt-br": "Guia de marca", "es": "Guía de marca",
		},
		PageLicense: {
			"en": "License", "pt-br": "Licença", "es": "Licencia",
		},
	}
	return headings[kind][localeKey]
}
