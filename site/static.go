package site

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

// SiteManifest returns the install metadata served at /site.webmanifest.
func SiteManifest() []byte {
	data, err := json.Marshal(struct {
		Name            string `json:"name"`
		ShortName       string `json:"short_name"`
		StartURL        string `json:"start_url"`
		Display         string `json:"display"`
		BackgroundColor string `json:"background_color"`
		ThemeColor      string `json:"theme_color"`
		Icons           []struct {
			Source  string `json:"src"`
			Type    string `json:"type"`
			Sizes   string `json:"sizes"`
			Purpose string `json:"purpose,omitempty"`
		} `json:"icons"`
	}{
		Name: "Arai Hû", ShortName: "Arai Hû", StartURL: "/en/", Display: "browser",
		BackgroundColor: "#07111f", ThemeColor: "#07111f",
		Icons: []struct {
			Source  string `json:"src"`
			Type    string `json:"type"`
			Sizes   string `json:"sizes"`
			Purpose string `json:"purpose,omitempty"`
		}{
			{Source: BrandAssetsPublicPrefix + "platform/web/araihu/icon-192.png", Type: "image/png", Sizes: "192x192"},
			{Source: BrandAssetsPublicPrefix + "platform/web/araihu/icon-512.png", Type: "image/png", Sizes: "512x512"},
			{Source: BrandAssetsPublicPrefix + "platform/web/araihu/icon-maskable-192.png", Type: "image/png", Sizes: "192x192", Purpose: "maskable"},
			{Source: BrandAssetsPublicPrefix + "platform/web/araihu/icon-maskable-512.png", Type: "image/png", Sizes: "512x512", Purpose: "maskable"},
		},
	})
	if err != nil {
		panic(fmt.Sprintf("marshal site manifest: %v", err))
	}
	return append(data, '\n')
}

// Robots returns the crawler policy served at /robots.txt.
func Robots() []byte {
	return []byte("User-agent: *\nAllow: /\nSitemap: " + CanonicalSiteURL + "/sitemap.xml\n")
}

type sitemapDocument struct {
	XMLName xml.Name          `xml:"urlset"`
	XMLNS   string            `xml:"xmlns,attr"`
	URLs    []sitemapURLEntry `xml:"url"`
}

type sitemapURLEntry struct {
	Location string `xml:"loc"`
}

// Sitemap creates an XML sitemap from canonical page records only.
func Sitemap(pages []Page) ([]byte, error) {
	document := sitemapDocument{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: make([]sitemapURLEntry, 0, len(pages))}
	seen := make(map[string]struct{}, len(pages))
	for _, page := range pages {
		url := page.Meta.CanonicalURL
		if url == "" || len(url) <= len(CanonicalSiteURL) || url[:len(CanonicalSiteURL)] != CanonicalSiteURL {
			return nil, fmt.Errorf("page %q has non-canonical URL %q", page.Meta.Path, url)
		}
		if _, exists := seen[url]; exists {
			return nil, fmt.Errorf("duplicate canonical URL %q", url)
		}
		seen[url] = struct{}{}
		document.URLs = append(document.URLs, sitemapURLEntry{Location: url})
	}
	data, err := xml.MarshalIndent(document, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), append(data, '\n')...), nil
}
