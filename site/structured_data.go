package site

import "strings"

const licenseEffectiveDate = "2026-07-29"
const licenseVersion = "1.0"

// StructuredData returns the JSON-LD node or graph for a public page.
func StructuredData(page Page) any {
	organization := map[string]any{
		"@type":  "Organization",
		"@id":    CanonicalSiteURL + "/#organization",
		"name":   "Arai Hû",
		"url":    CanonicalSiteURL + "/",
		"logo":   CanonicalSiteURL + BrandAssetsPublicPrefix + "platform/web/araihu/apple-touch-icon-180.png",
		"sameAs": []string{"https://github.com/araihu"},
	}
	website := map[string]any{
		"@type":      "WebSite",
		"@id":        CanonicalSiteURL + "/#website",
		"name":       "Arai Hû",
		"url":        CanonicalSiteURL + "/",
		"inLanguage": []string{"en", "pt-BR", "es"},
		"publisher":  map[string]any{"@id": organization["@id"]},
	}
	pageNode := map[string]any{
		"@type":       "WebPage",
		"@id":         page.Meta.CanonicalURL + "#webpage",
		"url":         page.Meta.CanonicalURL,
		"name":        page.Meta.Title,
		"description": page.Meta.Description,
		"inLanguage":  page.Meta.Locale.Language,
		"publisher":   map[string]any{"@id": organization["@id"]},
		"isPartOf":    map[string]any{"@id": website["@id"]},
		"about":       map[string]any{"@id": organization["@id"]},
	}
	graph := []any{organization, website, pageNode}
	blogs := make([]any, 0, len(siteLocales))
	for _, locale := range siteLocales {
		url := homeContent(locale.Key).BlogURL
		id := url + "#website"
		blogs = append(blogs, map[string]any{"@id": id})
		graph = append(graph, map[string]any{
			"@type":      []string{"WebSite", "Blog"},
			"@id":        id,
			"name":       "Arai Hû",
			"url":        url,
			"inLanguage": locale.Language,
			"publisher":  map[string]any{"@id": organization["@id"]},
			"isPartOf":   map[string]any{"@id": website["@id"]},
		})
	}
	website["hasPart"] = blogs
	if page.Meta.Kind == PageBrand {
		organization["brand"] = map[string]any{"@id": CanonicalSiteURL + "/#brand"}
		pageNode["mainEntity"] = map[string]any{"@id": CanonicalSiteURL + "/#brand"}
		graph = append(graph, map[string]any{
			"@type": "Brand",
			"@id":   CanonicalSiteURL + "/#brand",
			"name":  "Arai Hû",
			"url":   page.Meta.CanonicalURL,
			"logo":  organization["logo"],
		})
	}
	if page.Meta.Kind == PageHome && page.Home != nil {
		projects := make([]any, 0, len(page.Home.Projects)+len(page.Home.MoreProjects))
		for _, collection := range [][]Project{page.Home.Projects, page.Home.MoreProjects} {
			for _, project := range collection {
				projectNode := map[string]any{
					"name":               project.Name,
					"url":                project.URL,
					"description":        project.Description,
					"creativeWorkStatus": strings.ToLower(project.Status),
				}
				if project.LibraryRepository != "" {
					projectNode["@type"] = "SoftwareSourceCode"
					projectNode["codeRepository"] = project.LibraryRepository
					projectNode["programmingLanguage"] = "Go"
					projectNode["genre"] = project.Category
					projectNode["version"] = project.Version
				} else {
					projectNode["@type"] = "SoftwareApplication"
					projectNode["applicationCategory"] = project.ApplicationCategory
					projectNode["applicationSubCategory"] = project.Category
					projectNode["operatingSystem"] = project.OperatingSystem
					projectNode["softwareVersion"] = project.Version
				}
				projects = append(projects, projectNode)
			}
		}
		pageNode["mentions"] = projects
		pageNode["mainEntity"] = map[string]any{"@id": organization["@id"]}
	}
	if page.Meta.Kind == PageLicense {
		pageNode["dateModified"] = licenseEffectiveDate
		pageNode["version"] = licenseVersion
	}
	return map[string]any{"@context": "https://schema.org", "@graph": graph}
}
