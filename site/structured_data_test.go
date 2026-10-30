package site

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

func TestHomeStructuredDataIncludesProjectVersionsAndBalemohInEveryLocale(t *testing.T) {
	for _, page := range Pages() {
		if page.Meta.Kind != PageHome {
			continue
		}
		t.Run(page.Meta.Locale.Key, func(t *testing.T) {
			var data struct {
				Mentions []struct {
					Type            string `json:"@type"`
					Name            string `json:"name"`
					URL             string `json:"url"`
					Description     string `json:"description"`
					Category        string `json:"applicationCategory"`
					OperatingSystem string `json:"operatingSystem"`
					SoftwareVersion string `json:"softwareVersion"`
					Version         string `json:"version"`
					Language        string `json:"programmingLanguage"`
					Repository      string `json:"codeRepository"`
					Status          string `json:"creativeWorkStatus"`
				} `json:"mentions"`
			}
			if err := json.Unmarshal([]byte(scriptJSON(t, renderMetadata(t, page))), &data); err != nil {
				t.Fatal(err)
			}
			if len(data.Mentions) != 8 {
				t.Fatalf("projects = %d, want 8", len(data.Mentions))
			}
			libraryRepositories := map[string]string{
				"Goshtoso":            "https://github.com/araihu/goshtoso",
				"Goshtoso App Shells": "https://github.com/araihu/goshtoso-app-shells",
				"Goshtoso Charts":     "https://github.com/araihu/goshtoso-charts",
			}
			libraries := 0
			count := 0
			for _, application := range data.Mentions {
				version := application.SoftwareVersion
				if repository, library := libraryRepositories[application.Name]; library {
					libraries++
					version = application.Version
					if application.Type != "SoftwareSourceCode" || application.Repository != repository || application.Language != "Go" {
						t.Errorf("%s must describe its Go library source: %#v", application.Name, application)
					}
					if application.OperatingSystem != "" || application.Category != "" || application.SoftwareVersion != "" {
						t.Errorf("%s library must not claim application-only properties: %#v", application.Name, application)
					}
				} else if application.Type != "SoftwareApplication" || application.OperatingSystem == "" || (application.Category != "DeveloperApplication" && application.Category != "UtilitiesApplication") {
					t.Errorf("%s lacks application type, category or platform: %#v", application.Name, application)
				}
				parts := regexp.MustCompile(`^v0\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`).FindStringSubmatch(version)
				if parts == nil {
					t.Fatalf("%s version = %q, want a root v0.x.x tag", application.Name, version)
				}
				wantStatus := "beta"
				if parts[1] == "0" {
					wantStatus = "alpha"
				}
				if application.Status != wantStatus {
					t.Errorf("%s %s stage = %q, want %q", application.Name, version, application.Status, wantStatus)
				}
				projects := append(append([]Project{}, page.Home.Projects...), page.Home.MoreProjects...)
				for _, project := range projects {
					if project.Name == application.Name && (project.Version != version || strings.ToLower(project.Status) != application.Status || !strings.Contains(renderComponent(t, ProjectVersion(project)), version)) {
						t.Errorf("%s schema version/stage must match the server-rendered badge", application.Name)
					}
				}
				if application.Name != "Balemoh" {
					continue
				}
				count++
				if application.OperatingSystem != "Linux" || application.Category != "UtilitiesApplication" {
					t.Errorf("Balemoh platform/category = %#v", application)
				}
				if application.Type != "SoftwareApplication" || application.URL != "https://balemoh.araihu.com" {
					t.Errorf("Balemoh schema identity = %#v", application)
				}
				for _, project := range page.Home.Projects {
					if project.Name == "Balemoh" && application.Description != project.Description {
						t.Errorf("Balemoh schema description = %q, want visible localized description %q", application.Description, project.Description)
					}
				}
			}
			if libraries != 3 {
				t.Errorf("library entries = %d, want 3", libraries)
			}
			if count != 1 {
				t.Errorf("Balemoh schema entries = %d, want 1", count)
			}
		})
	}
}

func TestStructuredDataUsesStableOrganizationAndBrandGraph(t *testing.T) {
	brand := requirePage(t, Pages(), "/brand/", "en", PageBrand)
	data, err := json.Marshal(StructuredData(brand))
	if err != nil {
		t.Fatal(err)
	}
	var graph map[string]any
	if err := json.Unmarshal(data, &graph); err != nil {
		t.Fatal(err)
	}
	if graph["@context"] != "https://schema.org" {
		t.Errorf("context = %#v", graph["@context"])
	}
	nodes, ok := graph["@graph"].([]any)
	if !ok || len(nodes) != 2 {
		t.Fatalf("graph nodes = %#v, want organization and brand", graph["@graph"])
	}
	requireJSONLDID(t, nodes, CanonicalSiteURL+"/#organization")
	requireJSONLDID(t, nodes, CanonicalSiteURL+"/#brand")
	for _, raw := range nodes {
		node := raw.(map[string]any)
		if node["@type"] == "Brand" && node["publisher"] != nil {
			t.Error("Brand must not use unsupported publisher property")
		}
		if node["@type"] == "Organization" {
			brand, ok := node["brand"].(map[string]any)
			if !ok || brand["@id"] != CanonicalSiteURL+"/#brand" {
				t.Errorf("organization brand = %#v", node["brand"])
			}
		}
	}
}

func TestLicenseStructuredDataDescribesCanonicalLocalizedTerms(t *testing.T) {
	license := requirePage(t, Pages(), "/pt-br/license/", "pt-BR", PageLicense)
	data, err := json.Marshal(StructuredData(license))
	if err != nil {
		t.Fatal(err)
	}
	var node map[string]any
	if err := json.Unmarshal(data, &node); err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]string{
		"@type":        "WebPage",
		"url":          license.Meta.CanonicalURL,
		"inLanguage":   license.Meta.Locale.Language,
		"dateModified": "2026-07-29",
	} {
		if node[key] != want {
			t.Errorf("%s = %#v, want %q", key, node[key], want)
		}
	}
	publisher, ok := node["publisher"].(map[string]any)
	if !ok || publisher["@id"] != CanonicalSiteURL+"/#organization" {
		t.Errorf("publisher = %#v", node["publisher"])
	}
}

func requireJSONLDID(t *testing.T, nodes []any, id string) {
	t.Helper()
	for _, raw := range nodes {
		node, ok := raw.(map[string]any)
		if ok && node["@id"] == id {
			return
		}
	}
	t.Errorf("JSON-LD graph lacks %q", id)
}
