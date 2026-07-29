// Command checksite validates generated static-site discovery output.
package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/araihu/ahairu/site"
	"golang.org/x/net/html"
)

const canonicalOrigin = "https://araihu.com"

const (
	faviconURL  = "/assets/logos/araihu-icon-background.svg?rev=a8a9647a"
	manifestURL = "/site.webmanifest"
	themeColor  = "#07111f"
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
	for _, page := range pages {
		if err := checkPage(root, page); err != nil {
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
			Source string `json:"src"`
			Type   string `json:"type"`
			Sizes  string `json:"sizes"`
		} `json:"icons"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("parse site.webmanifest: %w", err)
	}
	if manifest.Name != "Arai Hû" || manifest.ShortName != "Arai Hû" || manifest.StartURL != "/en/" || manifest.Display != "browser" || manifest.BackgroundColor != themeColor || manifest.ThemeColor != themeColor {
		return fmt.Errorf("manifest semantics are invalid")
	}
	if len(manifest.Icons) != 1 {
		return fmt.Errorf("manifest has %d icons, want 1", len(manifest.Icons))
	}
	icon := manifest.Icons[0]
	if icon.Source != faviconURL || icon.Type != "image/svg+xml" || icon.Sizes != "any" {
		return fmt.Errorf("manifest icon metadata is invalid")
	}
	if err := requireLocalURL(root, icon.Source); err != nil {
		return fmt.Errorf("manifest icon: %w", err)
	}
	return nil
}

func checkSitemap(root string, pages []site.Page) error {
	data, err := os.ReadFile(filepath.Join(root, "sitemap.xml"))
	if err != nil {
		return fmt.Errorf("read sitemap.xml: %w", err)
	}
	var document struct {
		URLs []struct {
			Location string `xml:"loc"`
		} `xml:"url"`
	}
	if err := xml.Unmarshal(data, &document); err != nil {
		return fmt.Errorf("parse sitemap.xml: %w", err)
	}
	if len(document.URLs) != len(pages) {
		return fmt.Errorf("sitemap has %d URLs, want %d", len(document.URLs), len(pages))
	}
	for index, page := range pages {
		if document.URLs[index].Location != page.Meta.CanonicalURL {
			return fmt.Errorf("sitemap URL %d = %q, want canonical %q", index, document.URLs[index].Location, page.Meta.CanonicalURL)
		}
	}
	return nil
}

func checkPage(root string, page site.Page) error {
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
	if err := requireLink(document, "apple-touch-icon", faviconURL, nil); err != nil {
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
		if err := requireLocalURL(root, resource); err != nil {
			return fmt.Errorf("page %q local resource: %w", page.Meta.Path, err)
		}
	}
	return nil
}

func pageError(page site.Page, err error) error {
	return fmt.Errorf("page %q: %w", page.Meta.Path, err)
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
	if parsed.Host == "" && !strings.HasPrefix(parsed.Path, "/") {
		return "", fmt.Errorf("relative local path")
	}
	cleanPath := path.Clean(parsed.Path)
	if !strings.HasPrefix(cleanPath, "/") || strings.HasPrefix(cleanPath, "/../") {
		return "", fmt.Errorf("invalid local path")
	}
	if strings.HasSuffix(parsed.Path, "/") || filepath.Ext(cleanPath) == "" {
		cleanPath = strings.TrimSuffix(cleanPath, "/") + "/index.html"
	}
	file := filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(cleanPath, "/")))
	relative, err := filepath.Rel(root, file)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes public directory")
	}
	return file, nil
}

type htmlDocument struct {
	language  string
	titles    []string
	metas     []htmlElement
	links     []htmlElement
	jsonLD    []string
	resources []string
}

type htmlElement struct{ attributes map[string]string }

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
	var document htmlDocument
	tokenizer := html.NewTokenizer(bytes.NewReader(source))
	stack := make([]string, 0, 8)
	var textTarget string
	var scriptIsJSONLD bool
	var text strings.Builder
	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			if err := tokenizer.Err(); err != io.EOF {
				return htmlDocument{}, err
			}
			if len(stack) != 0 {
				return htmlDocument{}, fmt.Errorf("unclosed <%s>", stack[len(stack)-1])
			}
			return document, nil
		case html.TextToken:
			if textTarget != "" {
				text.Write(tokenizer.Raw())
			}
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			name := strings.ToLower(token.Data)
			attributes, err := tokenAttributes(token.Attr)
			if err != nil {
				return htmlDocument{}, fmt.Errorf("<%s>: %w", name, err)
			}
			if name == "html" {
				if document.language != "" {
					return htmlDocument{}, fmt.Errorf("duplicate <html>")
				}
				document.language = attributes["lang"]
			}
			switch name {
			case "meta":
				document.metas = append(document.metas, htmlElement{attributes: attributes})
			case "link":
				document.links = append(document.links, htmlElement{attributes: attributes})
				if href := attributes["href"]; href != "" {
					document.resources = append(document.resources, href)
				}
			case "img":
				if source := attributes["src"]; source != "" {
					document.resources = append(document.resources, source)
				}
			case "title", "script":
				if textTarget != "" {
					return htmlDocument{}, fmt.Errorf("nested text element <%s>", name)
				}
				textTarget = name
				text.Reset()
				scriptIsJSONLD = name == "script" && attributes["type"] == "application/ld+json"
			}
			if tokenType != html.SelfClosingTagToken && !isVoidElement(name) {
				stack = append(stack, name)
			}
			if name == "script" && attributes["type"] == "application/ld+json" && attributes["id"] != "structured-data" {
				return htmlDocument{}, fmt.Errorf("JSON-LD script has unexpected id")
			}
		case html.EndTagToken:
			token := tokenizer.Token()
			name := strings.ToLower(token.Data)
			if len(stack) == 0 || stack[len(stack)-1] != name {
				return htmlDocument{}, fmt.Errorf("mismatched closing </%s>", name)
			}
			stack = stack[:len(stack)-1]
			if textTarget == name {
				switch name {
				case "title":
					document.titles = append(document.titles, text.String())
				case "script":
					if scriptIsJSONLD {
						document.jsonLD = append(document.jsonLD, text.String())
					}
				}
				textTarget = ""
				scriptIsJSONLD = false
			}
		}
	}
}

func tokenAttributes(attributes []html.Attribute) (map[string]string, error) {
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

func isVoidElement(name string) bool {
	switch name {
	case "area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "param", "source", "track", "wbr":
		return true
	default:
		return false
	}
}
