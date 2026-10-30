package site

const licenseEffectiveDate = "2026-07-29"
const licenseVersion = "1.0"

// StructuredData returns the JSON-LD node or graph for a public page.
func StructuredData(page Page) any {
	organization := map[string]any{
		"@type": "Organization",
		"@id":   CanonicalSiteURL + "/#organization",
		"name":  "Arai Hû",
		"url":   CanonicalSiteURL + "/en/",
		"logo":  CanonicalSiteURL + BrandAssetsPublicPrefix + "platform/web/araihu/favicon.svg",
	}
	if page.Meta.Kind == PageBrand {
		organization["brand"] = map[string]any{"@id": CanonicalSiteURL + "/#brand"}
		return map[string]any{
			"@context": "https://schema.org",
			"@graph": []any{
				organization,
				map[string]any{
					"@type": "Brand",
					"@id":   CanonicalSiteURL + "/#brand",
					"name":  "Arai Hû",
					"url":   page.Meta.CanonicalURL,
					"logo":  organization["logo"],
				},
			},
		}
	}
	pageNode := map[string]any{
		"@context":    "https://schema.org",
		"@type":       "WebPage",
		"@id":         page.Meta.CanonicalURL + "#webpage",
		"url":         page.Meta.CanonicalURL,
		"name":        page.Meta.Title,
		"description": page.Meta.Description,
		"inLanguage":  page.Meta.Locale.Language,
		"publisher":   map[string]any{"@id": organization["@id"]},
	}
	if page.Meta.Kind == PageHome && page.Home != nil {
		projects := make([]any, 0, len(page.Home.Projects)+len(page.Home.MoreProjects))
		for _, collection := range [][]Project{page.Home.Projects, page.Home.MoreProjects} {
			for _, project := range collection {
				projects = append(projects, map[string]any{
					"@type":                  "SoftwareApplication",
					"name":                   project.Name,
					"url":                    project.URL,
					"description":            project.Description,
					"applicationCategory":    project.ApplicationCategory,
					"applicationSubCategory": project.Category,
					"operatingSystem":        project.OperatingSystem,
				})
			}
		}
		pageNode["mentions"] = projects
		pageNode["publisher"] = organization
	}
	if page.Meta.Kind == PageLicense {
		pageNode["dateModified"] = licenseEffectiveDate
		pageNode["version"] = licenseVersion
	}
	return pageNode
}
