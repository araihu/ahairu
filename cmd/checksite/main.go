// Command checksite validates generated static-site discovery output.
package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/png"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const canonicalOrigin = "https://araihu.com"

var tagPattern = regexp.MustCompile(`(?is)<(link|meta|script|img)\b[^>]*>`)
var attributePattern = regexp.MustCompile(`(?i)\b([a-z:-]+)="([^"]*)"`)
var titlePattern = regexp.MustCompile(`(?is)<title>(.*?)</title>`)
var jsonLDPattern = regexp.MustCompile(`(?is)<script\b[^>]*\btype="application/ld\+json"[^>]*>(.*?)</script>`)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: checksite <public-dir>")
		os.Exit(2)
	}
	if err := Check(os.Args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "check site: %v\n", err)
		os.Exit(1)
	}
}

// Check verifies static metadata, discovery documents, and linked local files.
func Check(root string) error {
	if err := checkRobots(root); err != nil {
		return err
	}
	if err := checkManifest(root); err != nil {
		return err
	}
	locations, err := sitemapLocations(root)
	if err != nil {
		return err
	}
	if len(locations) != 9 {
		return fmt.Errorf("sitemap has %d URLs, want 9", len(locations))
	}
	titles := map[string]struct{}{}
	descriptions := map[string]struct{}{}
	for _, location := range locations {
		if err := checkPage(root, location, titles, descriptions); err != nil {
			return err
		}
	}
	return nil
}

func checkRobots(root string) error {
	data, err := os.ReadFile(filepath.Join(root, "robots.txt"))
	if err != nil {
		return fmt.Errorf("read robots.txt: %w", err)
	}
	if !strings.Contains(string(data), "Sitemap: "+canonicalOrigin+"/sitemap.xml") {
		return fmt.Errorf("robots.txt lacks absolute sitemap URL")
	}
	return nil
}

func checkManifest(root string) error {
	data, err := os.ReadFile(filepath.Join(root, "site.webmanifest"))
	if err != nil {
		return fmt.Errorf("read site.webmanifest: %w", err)
	}
	var manifest struct {
		Icons []struct {
			Source string `json:"src"`
		} `json:"icons"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("parse site.webmanifest: %w", err)
	}
	if len(manifest.Icons) == 0 {
		return fmt.Errorf("site.webmanifest has no icons")
	}
	for _, icon := range manifest.Icons {
		if err := requireLocalURL(root, icon.Source); err != nil {
			return fmt.Errorf("manifest icon: %w", err)
		}
	}
	return nil
}

func sitemapLocations(root string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(root, "sitemap.xml"))
	if err != nil {
		return nil, fmt.Errorf("read sitemap.xml: %w", err)
	}
	var document struct {
		URLs []struct {
			Location string `xml:"loc"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(data, &document); err != nil {
		return nil, fmt.Errorf("parse sitemap.xml: %w", err)
	}
	seen := make(map[string]struct{}, len(document.URLs))
	locations := make([]string, 0, len(document.URLs))
	for _, entry := range document.URLs {
		if !strings.HasPrefix(entry.Location, canonicalOrigin+"/") {
			return nil, fmt.Errorf("sitemap URL is not absolute canonical URL %q", entry.Location)
		}
		if strings.Contains(entry.Location, "/en/brand") || strings.Contains(entry.Location, "/en/license") {
			return nil, fmt.Errorf("sitemap contains redirect URL %q", entry.Location)
		}
		if _, duplicate := seen[entry.Location]; duplicate {
			return nil, fmt.Errorf("sitemap repeats URL %q", entry.Location)
		}
		seen[entry.Location] = struct{}{}
		locations = append(locations, entry.Location)
	}
	return locations, nil
}

func checkPage(root, location string, titles, descriptions map[string]struct{}) error {
	pageFile, err := localFile(root, location)
	if err != nil {
		return fmt.Errorf("sitemap page %q: %w", location, err)
	}
	html, err := os.ReadFile(pageFile)
	if err != nil {
		return fmt.Errorf("read sitemap page %q: %w", location, err)
	}
	document := string(html)
	canonical := matchingAttribute(document, "link", "rel", "canonical", "href")
	if len(canonical) != 1 || canonical[0] != location {
		return fmt.Errorf("page %q has invalid canonical URL %#v", location, canonical)
	}
	title := titlePattern.FindStringSubmatch(document)
	if len(title) != 2 || strings.TrimSpace(title[1]) == "" {
		return fmt.Errorf("page %q lacks title", location)
	}
	if _, duplicate := titles[title[1]]; duplicate {
		return fmt.Errorf("page %q duplicates title %q", location, title[1])
	}
	titles[title[1]] = struct{}{}
	description := matchingAttribute(document, "meta", "name", "description", "content")
	if len(description) != 1 || strings.TrimSpace(description[0]) == "" {
		return fmt.Errorf("page %q lacks description", location)
	}
	if _, duplicate := descriptions[description[0]]; duplicate {
		return fmt.Errorf("page %q duplicates description %q", location, description[0])
	}
	descriptions[description[0]] = struct{}{}
	if err := requireAlternates(document); err != nil {
		return fmt.Errorf("page %q: %w", location, err)
	}
	for _, imageURL := range append(matchingAttribute(document, "meta", "property", "og:image", "content"), matchingAttribute(document, "meta", "name", "twitter:image", "content")...) {
		if !strings.HasPrefix(imageURL, canonicalOrigin+"/") {
			return fmt.Errorf("page %q has relative social image URL %q", location, imageURL)
		}
		if err := checkSocialImage(root, imageURL); err != nil {
			return fmt.Errorf("page %q social image: %w", location, err)
		}
	}
	if len(matchingAttribute(document, "meta", "property", "og:image", "content")) != 1 || len(matchingAttribute(document, "meta", "name", "twitter:image", "content")) != 1 {
		return fmt.Errorf("page %q lacks complete social image metadata", location)
	}
	jsonLD := jsonLDPattern.FindAllStringSubmatch(document, -1)
	if len(jsonLD) != 1 || !json.Valid([]byte(jsonLD[0][1])) {
		return fmt.Errorf("page %q has invalid JSON-LD", location)
	}
	for _, resource := range localResources(document) {
		if err := requireLocalURL(root, resource); err != nil {
			return fmt.Errorf("page %q local resource: %w", location, err)
		}
	}
	return nil
}

func requireAlternates(document string) error {
	want := map[string]bool{"en": false, "pt-BR": false, "es": false, "x-default": false}
	for _, tag := range tags(document) {
		if tag.name != "link" || tag.attributes["rel"] != "alternate" {
			continue
		}
		language, target := tag.attributes["hreflang"], tag.attributes["href"]
		if _, known := want[language]; !known || !strings.HasPrefix(target, canonicalOrigin+"/") {
			return fmt.Errorf("invalid alternate %q=%q", language, target)
		}
		if want[language] {
			return fmt.Errorf("duplicate alternate language %q", language)
		}
		want[language] = true
	}
	for language, found := range want {
		if !found {
			return fmt.Errorf("missing alternate language %q", language)
		}
	}
	return nil
}

func checkSocialImage(root, imageURL string) error {
	file, err := localFile(root, imageURL)
	if err != nil {
		return err
	}
	handle, err := os.Open(file)
	if err != nil {
		return err
	}
	defer handle.Close()
	config, _, err := image.DecodeConfig(handle)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}
	if config.Width != 1200 || config.Height != 630 {
		return fmt.Errorf("dimensions are %dx%d, want 1200x630", config.Width, config.Height)
	}
	return nil
}

func matchingAttribute(document, tagName, selectorName, selectorValue, valueName string) []string {
	var matches []string
	for _, tag := range tags(document) {
		if tag.name == tagName && tag.attributes[selectorName] == selectorValue && tag.attributes[valueName] != "" {
			matches = append(matches, tag.attributes[valueName])
		}
	}
	return matches
}

type htmlTag struct {
	name       string
	attributes map[string]string
}

func tags(document string) []htmlTag {
	matches := tagPattern.FindAllStringSubmatch(document, -1)
	result := make([]htmlTag, 0, len(matches))
	for _, match := range matches {
		attributes := map[string]string{}
		for _, attribute := range attributePattern.FindAllStringSubmatch(match[0], -1) {
			attributes[strings.ToLower(attribute[1])] = attribute[2]
		}
		result = append(result, htmlTag{name: strings.ToLower(match[1]), attributes: attributes})
	}
	return result
}

func localResources(document string) []string {
	resources := make([]string, 0)
	for _, tag := range tags(document) {
		for _, name := range []string{"href", "src"} {
			resource := tag.attributes[name]
			if resource == "" || strings.HasPrefix(resource, "#") || strings.HasPrefix(resource, "mailto:") {
				continue
			}
			parsed, err := url.Parse(resource)
			if err != nil || (parsed.Host != "" && parsed.Scheme != "https") || (parsed.Host != "" && parsed.Scheme == "https" && parsed.Host != "araihu.com") {
				continue
			}
			resources = append(resources, resource)
		}
	}
	return resources
}

func requireLocalURL(root, resource string) error {
	file, err := localFile(root, resource)
	if err != nil {
		return fmt.Errorf("%q: %w", resource, err)
	}
	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("%q: %w", resource, err)
	}
	return nil
}

func localFile(root, resource string) (string, error) {
	parsed, err := url.Parse(resource)
	if err != nil {
		return "", err
	}
	if parsed.Host != "" && (parsed.Scheme != "https" || parsed.Host != "araihu.com") {
		return "", fmt.Errorf("not a local Arai Hû URL")
	}
	path := parsed.EscapedPath()
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "..") {
		return "", fmt.Errorf("invalid local path")
	}
	if strings.HasSuffix(path, "/") || filepath.Ext(path) == "" {
		path = strings.TrimSuffix(path, "/") + "/index.html"
	}
	return filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(path, "/"))), nil
}
