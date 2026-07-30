package site

import (
	"encoding/json"
	"testing"
)

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
