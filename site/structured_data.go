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
		"logo":  CanonicalSiteURL + "/assets/logos/araihu-icon-background.svg?rev=a8a9647a",
	}
	if page.Meta.Kind == PageBrand {
		return map[string]any{
			"@context": "https://schema.org",
			"@graph": []any{
				organization,
				map[string]any{
					"@type":     "Brand",
					"@id":       CanonicalSiteURL + "/#brand",
					"name":      "Arai Hû",
					"url":       page.Meta.CanonicalURL,
					"logo":      organization["logo"],
					"publisher": map[string]any{"@id": organization["@id"]},
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
	if page.Meta.Kind == PageLicense {
		pageNode["dateModified"] = licenseEffectiveDate
		pageNode["version"] = licenseVersion
	}
	return pageNode
}
