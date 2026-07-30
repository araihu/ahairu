// Command checksite validates generated static-site discovery output.
package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/araihu/ahairu/internal/assetbundle"
	"github.com/araihu/ahairu/site"
	"golang.org/x/net/html"
)

const canonicalOrigin = "https://araihu.com"
const sitemapNamespace = "http://www.sitemaps.org/schemas/sitemap/0.9"

const (
	faviconURL               = "/assets/araihu/v0.1.0/platform/web/araihu/favicon.svg"
	appleTouchIconURL        = "/assets/araihu/v0.1.0/platform/web/araihu/apple-touch-icon-180.png"
	manifestURL              = "/site.webmanifest"
	themeColor               = "#07111f"
	campaignRuntimeIntegrity = "sha384-oPH7l1vK9vKP1Dn+18sO3yEXlz4ts6KzPEQl0SW4Y/+im05gOaamNNaQAf6bGH/n"
)

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

// Check verifies every generated public page against its typed source model.
func Check(root string) error {
	if err := checkAssetBundle(root); err != nil {
		return err
	}
	if err := checkRobots(root); err != nil {
		return err
	}
	if err := checkManifest(root); err != nil {
		return err
	}
	pages := site.Pages()
	if err := checkSitemap(root, pages); err != nil {
		return err
	}
	routes := pageRoutes(pages)
	for _, page := range pages {
		if err := checkPage(root, page, routes); err != nil {
			return err
		}
	}
	return nil
}

func checkAssetBundle(root string) error {
	source := assetBundleTree{source: os.DirFS(filepath.Join(root, "assets"))}
	if _, err := assetbundle.Validate(source); err != nil {
		return fmt.Errorf("asset release bundle: %w", err)
	}
	for _, name := range []string{"latest", "default", "current"} {
		data, err := fs.ReadFile(source, "releases/"+name+".json")
		if err != nil {
			return fmt.Errorf("read asset release channel %q: %w", name, err)
		}
		if err := checkChannelURLs(data); err != nil {
			return fmt.Errorf("asset release channel %q: %w", name, err)
		}
	}
	return nil
}

// assetBundleTree presents the release subtree to the shared bundle validator
// without treating unrelated site assets as bundle inputs.
type assetBundleTree struct{ source fs.FS }

func (tree assetBundleTree) Open(name string) (fs.File, error) {
	if name != "." && name != "campaign" && name != "releases" && !strings.HasPrefix(name, "campaign/") && !strings.HasPrefix(name, "releases/") {
		return nil, fs.ErrNotExist
	}
	return tree.source.Open(name)
}

func (tree assetBundleTree) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(tree.source, name)
	if err != nil || name != "." {
		return entries, err
	}
	filtered := make([]fs.DirEntry, 0, 2)
	for _, entry := range entries {
		if entry.Name() == "campaign" || entry.Name() == "releases" {
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}

func checkChannelURLs(data []byte) error {
	var document any
	if err := json.Unmarshal(data, &document); err != nil {
		return err
	}
	return checkURLs(document)
}

func checkURLs(value any) error {
	switch value := value.(type) {
	case map[string]any:
		for name, child := range value {
			if name == "url" || name == "cssUrl" {
				urlValue, ok := child.(string)
				if !ok {
					return fmt.Errorf("%s is not a string URL", name)
				}
				if err := requireAbsoluteSameOriginURL(urlValue); err != nil {
					return err
				}
			}
			if err := checkURLs(child); err != nil {
				return err
			}
		}
	case []any:
		for _, child := range value {
			if err := checkURLs(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func requireAbsoluteSameOriginURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "araihu.com" || parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || parsed.RawPath != "" {
		return fmt.Errorf("%q must be an absolute same-origin URL", value)
	}
	return nil
}

func checkRobots(root string) error {
	data, err := os.ReadFile(filepath.Join(root, "robots.txt"))
	if err != nil {
		return fmt.Errorf("read robots.txt: %w", err)
	}
	if string(data) != string(site.Robots()) {
		return fmt.Errorf("robots.txt directives do not exactly match generated policy")
	}
	return nil
}

func checkManifest(root string) error {
	data, err := os.ReadFile(filepath.Join(root, "site.webmanifest"))
	if err != nil {
		return fmt.Errorf("read site.webmanifest: %w", err)
	}
	var manifest struct {
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
			Purpose string `json:"purpose"`
		} `json:"icons"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("parse site.webmanifest: %w", err)
	}
	if manifest.Name != "Arai Hû" || manifest.ShortName != "Arai Hû" || manifest.StartURL != "/en/" || manifest.Display != "browser" || manifest.BackgroundColor != themeColor || manifest.ThemeColor != themeColor {
		return fmt.Errorf("manifest semantics are invalid")
	}
	if len(manifest.Icons) != 4 {
		return fmt.Errorf("manifest has %d icons, want 4", len(manifest.Icons))
	}
	expectedIcons := map[string]struct{ sizes, purpose string }{
		"/assets/araihu/v0.1.0/platform/web/araihu/icon-192.png":          {sizes: "192x192"},
		"/assets/araihu/v0.1.0/platform/web/araihu/icon-512.png":          {sizes: "512x512"},
		"/assets/araihu/v0.1.0/platform/web/araihu/icon-maskable-192.png": {sizes: "192x192", purpose: "maskable"},
		"/assets/araihu/v0.1.0/platform/web/araihu/icon-maskable-512.png": {sizes: "512x512", purpose: "maskable"},
	}
	for _, icon := range manifest.Icons {
		expected, ok := expectedIcons[icon.Source]
		if !ok || icon.Type != "image/png" || icon.Sizes != expected.sizes || icon.Purpose != expected.purpose {
			return fmt.Errorf("manifest icon metadata is invalid")
		}
		delete(expectedIcons, icon.Source)
		if err := requireLocalURL(root, icon.Source); err != nil {
			return fmt.Errorf("manifest icon: %w", err)
		}
	}
	if len(expectedIcons) != 0 {
		return fmt.Errorf("manifest icon set is incomplete")
	}
	return nil
}

func checkSitemap(root string, pages []site.Page) error {
	data, err := os.ReadFile(filepath.Join(root, "sitemap.xml"))
	if err != nil {
		return fmt.Errorf("read sitemap.xml: %w", err)
	}
	var document struct {
		XMLName xml.Name `xml:"urlset"`
		URLs    []struct {
			XMLName  xml.Name `xml:"url"`
			Location struct {
				XMLName xml.Name `xml:"loc"`
				Value   string   `xml:",chardata"`
			} `xml:"loc"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(data, &document); err != nil {
		return fmt.Errorf("parse sitemap.xml: %w", err)
	}
	if document.XMLName.Space != sitemapNamespace || document.XMLName.Local != "urlset" {
		return fmt.Errorf("sitemap root namespace = {%s}%s, want {%s}urlset", document.XMLName.Space, document.XMLName.Local, sitemapNamespace)
	}
	if len(document.URLs) != len(pages) {
		return fmt.Errorf("sitemap has %d URLs, want %d", len(document.URLs), len(pages))
	}
	want := make(map[string]struct{}, len(pages))
	for _, page := range pages {
		want[page.Meta.CanonicalURL] = struct{}{}
	}
	seen := make(map[string]struct{}, len(document.URLs))
	for _, entry := range document.URLs {
		if entry.XMLName.Space != sitemapNamespace || entry.XMLName.Local != "url" {
			return fmt.Errorf("sitemap url namespace = {%s}%s, want {%s}url", entry.XMLName.Space, entry.XMLName.Local, sitemapNamespace)
		}
		if entry.Location.XMLName.Space != sitemapNamespace || entry.Location.XMLName.Local != "loc" {
			return fmt.Errorf("sitemap loc namespace = {%s}%s, want {%s}loc", entry.Location.XMLName.Space, entry.Location.XMLName.Local, sitemapNamespace)
		}
		location := entry.Location.Value
		if _, duplicate := seen[location]; duplicate {
			return fmt.Errorf("sitemap contains duplicate URL %q", location)
		}
		seen[location] = struct{}{}
		if _, expected := want[location]; !expected {
			return fmt.Errorf("sitemap contains unexpected URL %q", location)
		}
	}
	for expected := range want {
		if _, found := seen[expected]; !found {
			return fmt.Errorf("sitemap is missing canonical URL %q", expected)
		}
	}
	return nil
}

func pageRoutes(pages []site.Page) map[string]struct{} {
	routes := make(map[string]struct{}, len(pages))
	for _, page := range pages {
		routes[page.Meta.Path] = struct{}{}
	}
	return routes
}

func checkPage(root string, page site.Page, routes map[string]struct{}) error {
	file, err := localFile(root, page.Meta.CanonicalURL)
	if err != nil {
		return fmt.Errorf("page %q: %w", page.Meta.Path, err)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("read page %q: %w", page.Meta.Path, err)
	}
	document, err := parseHTML(data)
	if err != nil {
		return fmt.Errorf("page %q: malformed HTML: %w", page.Meta.Path, err)
	}
	if document.language != page.Meta.Locale.Language {
		return fmt.Errorf("page %q html lang = %q, want %q", page.Meta.Path, document.language, page.Meta.Locale.Language)
	}
	if err := requireCampaignCanary(document, page); err != nil {
		return pageError(page, err)
	}
	if err := requireExactly("title", document.titles, page.Meta.Title); err != nil {
		return pageError(page, err)
	}
	if err := requireMeta(document, "name", "description", page.Meta.Description); err != nil {
		return pageError(page, err)
	}
	if err := requireMeta(document, "name", "robots", page.Meta.Robots); err != nil {
		return pageError(page, err)
	}
	if err := requireMeta(document, "name", "theme-color", themeColor); err != nil {
		return pageError(page, err)
	}
	if err := requireLink(document, "canonical", page.Meta.CanonicalURL, nil); err != nil {
		return pageError(page, err)
	}
	if err := requireLink(document, "icon", faviconURL, map[string]string{"type": "image/svg+xml"}); err != nil {
		return pageError(page, err)
	}
	if err := requireLink(document, "apple-touch-icon", appleTouchIconURL, nil); err != nil {
		return pageError(page, err)
	}
	if err := requireLink(document, "manifest", manifestURL, nil); err != nil {
		return pageError(page, err)
	}
	if err := requireAlternates(document, page); err != nil {
		return pageError(page, err)
	}
	for property, want := range map[string]string{
		"og:type":        "website",
		"og:title":       page.Meta.Title,
		"og:description": page.Meta.Description,
		"og:url":         page.Meta.CanonicalURL,
		"og:locale":      page.Meta.Locale.OGLocale,
		"og:image":       page.Meta.SocialImageURL,
	} {
		if err := requireMeta(document, "property", property, want); err != nil {
			return pageError(page, err)
		}
	}
	if err := requireOGAlternates(document, page); err != nil {
		return pageError(page, err)
	}
	for name, want := range map[string]string{
		"twitter:card":        "summary_large_image",
		"twitter:title":       page.Meta.Title,
		"twitter:description": page.Meta.Description,
		"twitter:image":       page.Meta.SocialImageURL,
	} {
		if err := requireMeta(document, "name", name, want); err != nil {
			return pageError(page, err)
		}
	}
	if err := checkSocialImage(root, page.Meta.SocialImageURL); err != nil {
		return fmt.Errorf("page %q social image: %w", page.Meta.Path, err)
	}
	if err := checkJSONLD(document, page); err != nil {
		return pageError(page, err)
	}
	for _, resource := range document.resources {
		if err := requireDocumentResource(root, resource, routes); err != nil {
			return fmt.Errorf("page %q local resource: %w", page.Meta.Path, err)
		}
	}
	return nil
}

func pageError(page site.Page, err error) error {
	return fmt.Errorf("page %q: %w", page.Meta.Path, err)
}

func requireCampaignCanary(document htmlDocument, page site.Page) error {
	if document.htmlAttributes["data-theme"] != "araihu" || document.htmlAttributes["data-theme-source"] != "default" {
		return fmt.Errorf("theme source must declare the Arai Hû default before body content")
	}

	var runtime *htmlElement
	for index := range document.scripts {
		script := &document.scripts[index]
		if script.attributes["src"] != "/assets/campaign/v1.js" {
			continue
		}
		if runtime != nil {
			return fmt.Errorf("campaign runtime appears more than once")
		}
		runtime = script
	}
	if runtime == nil || runtime.attributes["data-channel"] != "/assets/releases/current" || runtime.attributes["integrity"] != campaignRuntimeIntegrity || runtime.attributes["crossorigin"] != "anonymous" {
		return fmt.Errorf("campaign runtime contract is invalid")
	}
	if _, deferred := runtime.attributes["defer"]; !deferred {
		return fmt.Errorf("campaign runtime must defer")
	}
	for _, link := range document.links {
		if link.attributes["rel"] == "stylesheet" && link.position > runtime.position {
			return fmt.Errorf("campaign runtime must follow baseline styles")
		}
	}
	for _, image := range document.images {
		if image.attributes["src"] == "" || image.attributes["width"] == "" || image.attributes["height"] == "" {
			return fmt.Errorf("image dimensions must be explicit for every replaceable image")
		}
	}
	var logoHooks, iconHooks int
	for _, hook := range document.brandHooks {
		switch hook.attributes["data-asset-brand"] {
		case "logo":
			if hook.name != "img" || hook.attributes["src"] == "" {
				return fmt.Errorf("logo hook must target img[src]")
			}
			logoHooks++
		case "icon":
			if hook.name != "link" || hook.attributes["href"] == "" {
				return fmt.Errorf("icon hook must target link[href]")
			}
			if hook.attributes["rel"] != "icon" {
				return fmt.Errorf("icon hook must target rel=icon link")
			}
			iconHooks++
		default:
			return fmt.Errorf("asset brand hook %q is invalid", hook.attributes["data-asset-brand"])
		}
	}
	wantLogoHooks := 0
	if page.Meta.Kind == site.PageBrand {
		wantLogoHooks = 1
	}
	if logoHooks != wantLogoHooks || iconHooks != 1 {
		return fmt.Errorf("campaign brand hooks = logo:%d icon:%d, want logo:%d icon:1", logoHooks, iconHooks, wantLogoHooks)
	}
	return nil
}

func requireExactly(name string, values []string, want string) error {
	if len(values) != 1 || values[0] != want {
		return fmt.Errorf("%s = %#v, want exactly %q", name, values, want)
	}
	return nil
}

func requireMeta(document htmlDocument, selector, name, want string) error {
	values := document.metaValues(selector, name)
	return requireExactly(name, values, want)
}

func requireLink(document htmlDocument, rel, wantHref string, wantAttributes map[string]string) error {
	values := make([]string, 0, 1)
	for _, link := range document.links {
		if link.attributes["rel"] != rel {
			continue
		}
		matches := true
		for attribute, want := range wantAttributes {
			if link.attributes[attribute] != want {
				matches = false
				break
			}
		}
		if !matches {
			continue
		}
		values = append(values, link.attributes["href"])
	}
	return requireExactly("link rel="+rel, values, wantHref)
}

func requireAlternates(document htmlDocument, page site.Page) error {
	want := make(map[string]string, len(page.Meta.Alternates))
	for _, alternate := range page.Meta.Alternates {
		want[alternate.Language] = alternate.URL
	}
	actual := map[string]string{}
	for _, link := range document.links {
		if link.attributes["rel"] != "alternate" {
			continue
		}
		language, target := link.attributes["hreflang"], link.attributes["href"]
		if language == "" || target == "" {
			return fmt.Errorf("alternate tag is incomplete")
		}
		if _, duplicate := actual[language]; duplicate {
			return fmt.Errorf("duplicate alternate language %q", language)
		}
		actual[language] = target
	}
	if len(actual) != len(want) {
		return fmt.Errorf("alternate set has %d entries, want %d", len(actual), len(want))
	}
	for language, target := range want {
		if actual[language] != target {
			return fmt.Errorf("alternate %q = %q, want %q", language, actual[language], target)
		}
	}
	return nil
}

func requireOGAlternates(document htmlDocument, page site.Page) error {
	want := map[string]bool{}
	for _, alternate := range page.Meta.Alternates {
		if alternate.Language != "x-default" && alternate.Language != page.Meta.Locale.Language {
			want[ogLocale(alternate.Language)] = true
		}
	}
	actual := map[string]bool{}
	for _, value := range document.metaValues("property", "og:locale:alternate") {
		if actual[value] {
			return fmt.Errorf("duplicate og:locale:alternate %q", value)
		}
		actual[value] = true
	}
	if len(actual) != len(want) {
		return fmt.Errorf("og:locale:alternate set has %d entries, want %d", len(actual), len(want))
	}
	for value := range want {
		if !actual[value] {
			return fmt.Errorf("missing og:locale:alternate %q", value)
		}
	}
	return nil
}

func ogLocale(language string) string {
	for _, candidate := range site.Pages() {
		if candidate.Meta.Locale.Language == language {
			return candidate.Meta.Locale.OGLocale
		}
	}
	return ""
}

func checkJSONLD(document htmlDocument, page site.Page) error {
	if len(document.jsonLD) != 1 {
		return fmt.Errorf("JSON-LD script count = %d, want 1", len(document.jsonLD))
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(document.jsonLD[0]), &value); err != nil {
		return fmt.Errorf("invalid JSON-LD: %w", err)
	}
	if value["@context"] != "https://schema.org" {
		return fmt.Errorf("JSON-LD context is invalid")
	}
	organizationID := canonicalOrigin + "/#organization"
	if page.Meta.Kind == site.PageBrand {
		graph, ok := value["@graph"].([]any)
		if !ok || len(graph) != 2 {
			return fmt.Errorf("JSON-LD brand graph is invalid")
		}
		organization := findJSONLDNode(graph, organizationID)
		brand := findJSONLDNode(graph, canonicalOrigin+"/#brand")
		if organization == nil || organization["@type"] != "Organization" || organization["logo"] != faviconAbsoluteURL() {
			return fmt.Errorf("JSON-LD organization is invalid")
		}
		if brand == nil || brand["@type"] != "Brand" || brand["url"] != page.Meta.CanonicalURL || brand["logo"] != organization["logo"] || jsonLDReference(brand["publisher"]) != organizationID {
			return fmt.Errorf("JSON-LD brand relationship is invalid")
		}
		return nil
	}
	if value["@type"] != "WebPage" || value["@id"] != page.Meta.CanonicalURL+"#webpage" || value["url"] != page.Meta.CanonicalURL || value["inLanguage"] != page.Meta.Locale.Language || jsonLDReference(value["publisher"]) != organizationID {
		return fmt.Errorf("JSON-LD webpage relationship is invalid")
	}
	if page.Meta.Kind == site.PageLicense {
		if value["version"] != page.License.Version || value["dateModified"] != page.License.EffectiveDate {
			return fmt.Errorf("JSON-LD license version or effective date is invalid")
		}
	}
	return nil
}

func findJSONLDNode(graph []any, id string) map[string]any {
	for _, raw := range graph {
		node, ok := raw.(map[string]any)
		if ok && node["@id"] == id {
			return node
		}
	}
	return nil
}

func jsonLDReference(value any) string {
	reference, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	id, _ := reference["@id"].(string)
	return id
}

func faviconAbsoluteURL() string { return canonicalOrigin + faviconURL }

func checkSocialImage(root, imageURL string) error {
	file, err := localFile(root, imageURL)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	if err := site.ValidateSocialPreviewPNG(data); err != nil {
		return err
	}
	return nil
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

func requireDocumentResource(root string, resource htmlResource, routes map[string]struct{}) error {
	if resource.URL == "" || strings.HasPrefix(resource.URL, "#") || strings.HasPrefix(resource.URL, "mailto:") || strings.HasPrefix(resource.URL, "tel:") {
		return nil
	}
	parsed, err := url.Parse(resource.URL)
	if err != nil {
		return fmt.Errorf("%q: %w", resource.URL, err)
	}
	if parsed.Host != "" && parsed.Host != "araihu.com" {
		return nil
	}
	if parsed.Host == "araihu.com" && parsed.Scheme != "https" {
		return fmt.Errorf("%q: local URLs must use HTTPS", resource.URL)
	}
	if parsed.Host == "" && parsed.Scheme != "" {
		return fmt.Errorf("%q: unsupported local URL scheme", resource.URL)
	}
	exactFile, err := exactLocalFile(root, resource.URL)
	if err != nil {
		return fmt.Errorf("%q: %w", resource.URL, err)
	}
	extensionless := filepath.Ext(strings.TrimSuffix(parsed.Path, "/")) == ""
	if extensionless && resource.Kind != "anchor" && resource.Kind != "link" {
		return fmt.Errorf("%q: local %s must name a file", resource.URL, resource.Kind)
	}
	if info, statErr := os.Stat(exactFile); statErr == nil && info.Mode().IsRegular() {
		return nil
	} else if statErr != nil && !os.IsNotExist(statErr) {
		return fmt.Errorf("%q: %w", resource.URL, statErr)
	}
	if extensionless {
		if _, known := routes[parsed.Path]; !known {
			return fmt.Errorf("%q: unknown local page route", resource.URL)
		}
	}
	if err := requireLocalURL(root, resource.URL); err != nil {
		return err
	}
	return nil
}

func localFile(root, resource string) (string, error) {
	file, err := exactLocalFile(root, resource)
	if err != nil {
		return "", err
	}
	parsed, err := url.Parse(resource)
	if err != nil {
		return "", err
	}
	cleanPath := path.Clean(parsed.Path)
	if strings.HasSuffix(parsed.Path, "/") || filepath.Ext(cleanPath) == "" {
		file = filepath.Join(file, "index.html")
	}
	return file, nil
}

func exactLocalFile(root, resource string) (string, error) {
	parsed, err := url.Parse(resource)
	if err != nil {
		return "", err
	}
	if parsed.Host != "" && (parsed.Scheme != "https" || parsed.Host != "araihu.com") {
		return "", fmt.Errorf("not a local Arai Hû URL")
	}
	if parsed.Host == "" && !strings.HasPrefix(parsed.Path, "/") {
		return "", fmt.Errorf("relative local path")
	}
	for _, segment := range strings.Split(parsed.Path, "/") {
		if segment == ".." {
			return "", fmt.Errorf("path traversal")
		}
	}
	cleanPath := path.Clean(parsed.Path)
	if !strings.HasPrefix(cleanPath, "/") || strings.HasPrefix(cleanPath, "/../") {
		return "", fmt.Errorf("invalid local path")
	}
	file := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(cleanPath, "/")))
	relative, err := filepath.Rel(root, file)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes public directory")
	}
	return file, nil
}

type htmlDocument struct {
	language       string
	htmlAttributes map[string]string
	titles         []string
	metas          []htmlElement
	links          []htmlElement
	scripts        []htmlElement
	images         []htmlElement
	brandHooks     []htmlElement
	jsonLD         []string
	resources      []htmlResource
	nextPosition   int
}

type htmlElement struct {
	name       string
	attributes map[string]string
	position   int
}

type htmlResource struct {
	URL  string
	Kind string
}

func (document htmlDocument) metaValues(selector, name string) []string {
	values := make([]string, 0, 1)
	for _, meta := range document.metas {
		if meta.attributes[selector] == name {
			values = append(values, meta.attributes["content"])
		}
	}
	return values
}

func parseHTML(source []byte) (htmlDocument, error) {
	root, err := html.Parse(bytes.NewReader(source))
	if err != nil {
		return htmlDocument{}, err
	}
	var document htmlDocument
	if err := collectHTML(root, &document); err != nil {
		return htmlDocument{}, err
	}
	return document, nil
}

func collectHTML(node *html.Node, document *htmlDocument) error {
	if node.Type == html.ElementNode {
		name := strings.ToLower(node.Data)
		attributes, err := nodeAttributes(node.Attr)
		if err != nil {
			return fmt.Errorf("<%s>: %w", name, err)
		}
		document.nextPosition++
		element := htmlElement{name: name, attributes: attributes, position: document.nextPosition}
		if attributes["data-asset-brand"] != "" {
			document.brandHooks = append(document.brandHooks, element)
		}
		switch name {
		case "html":
			if document.language != "" {
				return fmt.Errorf("duplicate <html>")
			}
			document.language = attributes["lang"]
			document.htmlAttributes = attributes
		case "title":
			document.titles = append(document.titles, nodeText(node))
		case "meta":
			document.metas = append(document.metas, element)
		case "link":
			document.links = append(document.links, element)
			appendResource(document, attributes["href"], "link")
		case "img":
			document.images = append(document.images, element)
			appendResource(document, attributes["src"], "image")
		case "script":
			document.scripts = append(document.scripts, element)
			appendResource(document, attributes["src"], "script")
			if attributes["type"] == "application/ld+json" {
				if attributes["id"] != "structured-data" {
					return fmt.Errorf("JSON-LD script has unexpected id")
				}
				document.jsonLD = append(document.jsonLD, nodeText(node))
			}
		case "a":
			appendResource(document, attributes["href"], "anchor")
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if err := collectHTML(child, document); err != nil {
			return err
		}
	}
	return nil
}

func appendResource(document *htmlDocument, resource, kind string) {
	if resource != "" {
		document.resources = append(document.resources, htmlResource{URL: resource, Kind: kind})
	}
}

func nodeText(node *html.Node) string {
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

func nodeAttributes(attributes []html.Attribute) (map[string]string, error) {
	result := make(map[string]string, len(attributes))
	for _, attribute := range attributes {
		name := strings.ToLower(attribute.Key)
		if _, duplicate := result[name]; duplicate {
			return nil, fmt.Errorf("duplicate attribute %q", name)
		}
		result[name] = attribute.Val
	}
	return result, nil
}
