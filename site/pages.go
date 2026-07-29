package site

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
