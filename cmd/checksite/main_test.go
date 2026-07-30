package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/ahairu/site"
)

func TestCheckAcceptsCompleteStaticSite(t *testing.T) {
	root := writeValidPublic(t)
	if err := Check(root); err != nil {
		t.Fatal(err)
	}
}

func TestCheckAcceptsNormalHTMLWithOptionalEndTags(t *testing.T) {
	root := writeValidPublic(t)
	mutatePage(t, root, "/en/", func(document string) string {
		return strings.Replace(document, "</body>", `<ul><li>one<li>two</ul><p>first<p>second<table><tr><td>left<td>right</tr></table><a href="https://example.com">external</a><a href="#local">fragment</a><a href="mailto:hello@example.com">mail</a><a href="tel:+15551212">phone</a>`+"</body>", 1)
	})
	if err := Check(root); err != nil {
		t.Fatal(err)
	}
}

func TestCheckRejectsMissingOrWrongSizedSocialImages(t *testing.T) {
	root := writeValidPublic(t)
	if err := os.Remove(filepath.Join(root, "social", "brand.png")); err != nil {
		t.Fatal(err)
	}
	if err := Check(root); err == nil || !strings.Contains(err.Error(), "social image") {
		t.Fatalf("Check missing social image error = %v", err)
	}

	root = writeValidPublic(t)
	writePNG(t, filepath.Join(root, "social", "license.png"), 1200, 629)
	if err := Check(root); err == nil || !strings.Contains(err.Error(), "1200x630") {
		t.Fatalf("Check wrong social image size error = %v", err)
	}
}

func TestCheckRejectsNonTruecolorSocialImages(t *testing.T) {
	images := map[string]image.Image{
		"grayscale": image.NewGray(image.Rect(0, 0, 1200, 630)),
		"indexed":   image.NewPaletted(image.Rect(0, 0, 1200, 630), color.Palette{color.Black, color.White}),
		"rgba":      image.NewNRGBA(image.Rect(0, 0, 1200, 630)),
	}
	for name, preview := range images {
		t.Run(name, func(t *testing.T) {
			root := writeValidPublic(t)
			writeImagePNG(t, filepath.Join(root, "social", "brand.png"), preview)
			if err := Check(root); err == nil || !strings.Contains(err.Error(), "PNG") {
				t.Fatalf("Check() error = %v, want exact PNG format failure", err)
			}
		})
	}
}

func TestCheckRejectsRelativeMetadataAndInvalidJSONLD(t *testing.T) {
	root := writeValidPublic(t)
	page := filepath.Join(root, "en", "index.html")
	html, err := os.ReadFile(page)
	if err != nil {
		t.Fatal(err)
	}
	html = []byte(strings.Replace(string(html), `href="https://araihu.com/en/"`, `href="/en/"`, 1))
	if err := os.WriteFile(page, html, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Check(root); err == nil || !strings.Contains(err.Error(), "canonical") {
		t.Fatalf("Check relative canonical error = %v", err)
	}
}

func TestCheckRejectsMissingLocalDocumentResources(t *testing.T) {
	root := writeValidPublic(t)
	page := filepath.Join(root, "en", "index.html")
	html, err := os.ReadFile(page)
	if err != nil {
		t.Fatal(err)
	}
	html = []byte(strings.Replace(string(html), "</body>", `<img src="/assets/missing.svg" alt="">`+"</body>", 1))
	if err := os.WriteFile(page, html, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Check(root); err == nil || !strings.Contains(err.Error(), "missing.svg") {
		t.Fatalf("Check missing local resource error = %v", err)
	}
}

func TestCheckAcceptsExistingDotlessVersionedAsset(t *testing.T) {
	root := writeValidPublic(t)
	mutatePage(t, root, "/en/", func(document string) string {
		return strings.Replace(document, "</body>", `<a href="/assets/araihu/v0.1.0/NOTICE" download>NOTICE</a>`+"</body>", 1)
	})
	if err := Check(root); err != nil {
		t.Fatal(err)
	}
}

func TestCheckRejectsExistingDotlessScriptAndImage(t *testing.T) {
	for name, element := range map[string]string{
		"script": `<script src="/assets/araihu/v0.1.0/NOTICE"></script>`,
		"image":  `<img src="/assets/araihu/v0.1.0/NOTICE" alt="">`,
	} {
		t.Run(name, func(t *testing.T) {
			root := writeValidPublic(t)
			mutatePage(t, root, "/en/", func(document string) string {
				return strings.Replace(document, "</body>", element+"</body>", 1)
			})
			if err := Check(root); err == nil || !strings.Contains(err.Error(), "must name a file") {
				t.Fatalf("Check() error = %v, want must name a file", err)
			}
		})
	}
}

func TestCheckRejectsMissingDotlessAssetAndUnknownExtensionlessPage(t *testing.T) {
	for _, resource := range []string{
		"/assets/araihu/v0.1.0/MISSING",
		"/unknown-extensionless-page",
	} {
		t.Run(resource, func(t *testing.T) {
			root := writeValidPublic(t)
			mutatePage(t, root, "/en/", func(document string) string {
				return strings.Replace(document, "</body>", `<a href="`+resource+`">missing</a>`+"</body>", 1)
			})
			if err := Check(root); err == nil || !strings.Contains(err.Error(), "unknown local page route") {
				t.Fatalf("Check() error = %v, want unknown local page route", err)
			}
		})
	}
}

func TestCheckRejectsMissingLocalScriptDownloadAndPage(t *testing.T) {
	tests := []struct {
		name     string
		fragment string
		want     string
	}{
		{name: "script", fragment: `<script src="/assets/missing.js"></script>`, want: "missing.js"},
		{name: "download", fragment: `<a href="/downloads/missing.zip">download</a>`, want: "missing.zip"},
		{name: "page", fragment: `<a href="/missing-page/">page</a>`, want: "missing-page"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeValidPublic(t)
			mutatePage(t, root, "/en/", func(document string) string {
				return strings.Replace(document, "</body>", test.fragment+"</body>", 1)
			})
			if err := Check(root); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Check() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestCheckRejectsTraversalInLocalResource(t *testing.T) {
	for _, resource := range []string{
		"/assets/%2e%2e/assets/styles.css",
		"/assets%2f..%2fsocial/license.png",
		"/assets%2F.%2e%2Fsocial/license.png",
	} {
		t.Run(resource, func(t *testing.T) {
			root := writeValidPublic(t)
			mutatePage(t, root, "/en/", func(document string) string {
				return strings.Replace(document, "</body>", `<script src="`+resource+`"></script>`+"</body>", 1)
			})
			if err := Check(root); err == nil || !strings.Contains(err.Error(), "traversal") {
				t.Fatalf("Check() error = %v, want traversal failure", err)
			}
		})
	}
}

func TestCheckAcceptsReorderedSitemap(t *testing.T) {
	root := writeValidPublic(t)
	pages := site.Pages()
	for left, right := 0, len(pages)-1; left < right; left, right = left+1, right-1 {
		pages[left], pages[right] = pages[right], pages[left]
	}
	writeFile(t, filepath.Join(root, "sitemap.xml"), sitemapXML(pages, "http://www.sitemaps.org/schemas/sitemap/0.9"))
	if err := Check(root); err != nil {
		t.Fatal(err)
	}
}

func TestCheckRejectsSitemapWithoutNamespaceOrWithDuplicateURL(t *testing.T) {
	t.Run("namespace", func(t *testing.T) {
		root := writeValidPublic(t)
		writeFile(t, filepath.Join(root, "sitemap.xml"), sitemapXML(site.Pages(), ""))
		if err := Check(root); err == nil || !strings.Contains(err.Error(), "namespace") {
			t.Fatalf("Check() error = %v, want namespace failure", err)
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		root := writeValidPublic(t)
		pages := site.Pages()
		pages[len(pages)-1] = pages[0]
		writeFile(t, filepath.Join(root, "sitemap.xml"), sitemapXML(pages, "http://www.sitemaps.org/schemas/sitemap/0.9"))
		if err := Check(root); err == nil || !strings.Contains(err.Error(), "duplicate") {
			t.Fatalf("Check() error = %v, want duplicate failure", err)
		}
	})
	for _, mutation := range []struct {
		name string
		from string
		to   string
	}{
		{name: "url child reset", from: "<url>", to: `<url xmlns="">`},
		{name: "loc child reset", from: "<loc>", to: `<loc xmlns="">`},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			root := writeValidPublic(t)
			path := filepath.Join(root, "sitemap.xml")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, path, []byte(strings.Replace(string(data), mutation.from, mutation.to, 1)))
			if err := Check(root); err == nil || !strings.Contains(err.Error(), "namespace") {
				t.Fatalf("Check() error = %v, want child namespace failure", err)
			}
		})
	}
}

func TestCheckRejectsAdversarialDiscoveryDocuments(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t *testing.T, root string)
		want   string
	}{
		{
			name: "commented canonical tag",
			mutate: func(t *testing.T, root string) {
				mutatePage(t, root, "/en/", func(html string) string {
					html = removeTag(t, html, `<link rel="canonical"`)
					return strings.Replace(html, "</head>", `<!-- <link rel="canonical" href="https://araihu.com/en/"> -->`+"</head>", 1)
				})
			},
			want: "canonical",
		},
		{
			name: "wrong reciprocal hreflang target",
			mutate: func(t *testing.T, root string) {
				mutatePage(t, root, "/en/", func(html string) string {
					return strings.Replace(html, `hreflang="es" href="https://araihu.com/es/"`, `hreflang="es" href="https://araihu.com/en/"`, 1)
				})
			},
			want: "alternate",
		},
		{
			name: "missing open graph title",
			mutate: func(t *testing.T, root string) {
				mutatePage(t, root, "/en/", func(html string) string { return removeTag(t, html, `<meta property="og:title"`) })
			},
			want: "og:title",
		},
		{
			name: "missing X card",
			mutate: func(t *testing.T, root string) {
				mutatePage(t, root, "/en/", func(html string) string { return removeTag(t, html, `<meta name="twitter:card"`) })
			},
			want: "twitter:card",
		},
		{
			name: "wrong document language",
			mutate: func(t *testing.T, root string) {
				mutatePage(t, root, "/en/", func(html string) string { return strings.Replace(html, `<html lang="en"`, `<html lang="fr"`, 1) })
			},
			want: "html lang",
		},
		{
			name: "invalid structured data relationship",
			mutate: func(t *testing.T, root string) {
				mutatePage(t, root, "/brand/", func(html string) string {
					return strings.ReplaceAll(html, `https://araihu.com/#organization`, `https://araihu.com/#other`)
				})
			},
			want: "JSON-LD",
		},
		{
			name: "inexact robots directives",
			mutate: func(t *testing.T, root string) {
				path := filepath.Join(root, "robots.txt")
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				writeFile(t, path, []byte(strings.Replace(string(data), "Allow: /", "Allow: /private", 1)))
			},
			want: "robots.txt",
		},
		{
			name: "invalid manifest icon size",
			mutate: func(t *testing.T, root string) {
				path := filepath.Join(root, "site.webmanifest")
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				writeFile(t, path, []byte(strings.Replace(string(data), `"sizes":"192x192"`, `"sizes":"42x42"`, 1)))
			},
			want: "manifest",
		},
		{
			name: "unexpected sitemap route",
			mutate: func(t *testing.T, root string) {
				const oldURL = "https://araihu.com/es/license/"
				const newURL = "https://araihu.com/es/license-copy/"
				original := filepath.Join(root, "es", "license", "index.html")
				data, err := os.ReadFile(original)
				if err != nil {
					t.Fatal(err)
				}
				html := strings.ReplaceAll(string(data), oldURL, newURL)
				html = strings.Replace(html, "<title>", "<title>Copy ", 1)
				html = strings.Replace(html, `name="description" content="`, `name="description" content="Copy `, 1)
				writeFile(t, filepath.Join(root, "es", "license-copy", "index.html"), []byte(html))
				sitemap := filepath.Join(root, "sitemap.xml")
				data, err = os.ReadFile(sitemap)
				if err != nil {
					t.Fatal(err)
				}
				writeFile(t, sitemap, []byte(strings.Replace(string(data), oldURL, newURL, 1)))
			},
			want: "sitemap",
		},
		{
			name: "duplicate html attribute",
			mutate: func(t *testing.T, root string) {
				mutatePage(t, root, "/en/", func(html string) string {
					return strings.Replace(html, `<html lang="en"`, `<html lang="en" lang="es"`, 1)
				})
			},
			want: "duplicate",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeValidPublic(t)
			test.mutate(t, root)
			if err := Check(root); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Check() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestCheckRejectsInvalidAssetReleaseBundle(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t *testing.T, root string)
		want   string
	}{
		{
			name: "missing current channel",
			mutate: func(t *testing.T, root string) {
				if err := os.Remove(filepath.Join(root, "assets", "releases", "current.json")); err != nil {
					t.Fatal(err)
				}
			},
			want: "current.json",
		},
		{
			name: "missing current runtime",
			mutate: func(t *testing.T, root string) {
				if err := os.Remove(filepath.Join(root, "assets", "campaign", "v1.js")); err != nil {
					t.Fatal(err)
				}
			},
			want: "campaign/v1.js",
		},
		{
			name: "missing immutable release document",
			mutate: func(t *testing.T, root string) {
				if err := os.Remove(filepath.Join(root, "assets", "releases", "v0.1.1", "release.json")); err != nil {
					t.Fatal(err)
				}
			},
			want: "release.json",
		},
		{
			name: "invalid release checksums",
			mutate: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "assets", "releases", "v0.1.1", "checksums.txt"), []byte("not a checksum\n"))
			},
			want: "checksums",
		},
		{
			name: "relative channel URL",
			mutate: func(t *testing.T, root string) {
				writeReleaseChannel(t, root, "current", "/assets/releases/v0.1.1/themes/base.css")
			},
			want: "absolute same-origin URL",
		},
		{
			name: "invalid SemVer directory",
			mutate: func(t *testing.T, root string) {
				if err := os.Rename(filepath.Join(root, "assets", "releases", "v0.1.1"), filepath.Join(root, "assets", "releases", "v0.1.01")); err != nil {
					t.Fatal(err)
				}
			},
			want: "outside known layout",
		},
		{
			name: "missing retained inventory file",
			mutate: func(t *testing.T, root string) {
				if err := os.Remove(filepath.Join(root, "assets", "releases", "v0.1.1", "themes", "base.css")); err != nil {
					t.Fatal(err)
				}
			},
			want: "inventory misses file",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeValidPublic(t)
			test.mutate(t, root)
			if err := Check(root); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Check() error = %v, want %q", err, test.want)
			}
		})
	}
}

func writeValidPublic(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "robots.txt"), site.Robots())
	sitemap, err := site.Sitemap(site.Pages())
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "sitemap.xml"), sitemap)
	writeFile(t, filepath.Join(root, "site.webmanifest"), site.SiteManifest())
	for _, name := range []string{"styles.css", "ahairu.css", "araihu-theme.css"} {
		writeFile(t, filepath.Join(root, "assets", name), []byte("fixture"))
	}
	if err := site.CopyBundledBrandAssets(filepath.Join(root, "assets", "araihu", "v0.1.0")); err != nil {
		t.Fatal(err)
	}
	writePNG(t, filepath.Join(root, "social", "brand.png"), 1200, 630)
	writePNG(t, filepath.Join(root, "social", "license.png"), 1200, 630)
	for _, page := range site.Pages() {
		var output strings.Builder
		if err := site.Layout(page, templ.NopComponent).Render(context.Background(), &output); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(root, page.Meta.Path, "index.html"), []byte(output.String()))
	}
	writeAssetReleaseBundle(t, root)
	return root
}

func writeAssetReleaseBundle(t *testing.T, root string) {
	t.Helper()
	releaseRoot := filepath.Join(root, "assets", "releases", "v0.1.1")
	files := map[string][]byte{
		"catalog.json":    []byte(`{"schemaVersion":1}`),
		"themes.json":     []byte(`{"schemaVersion":1}`),
		"campaigns.json":  []byte(`{"schemaVersion":1,"campaigns":[]}`),
		"campaign/v1.js":  []byte("(() => {})()\n"),
		"themes/base.css": []byte("body{}\n"),
	}
	ordered := []string{"catalog.json", "themes.json", "campaigns.json", "campaign/v1.js", "themes/base.css"}
	type inventoryFile struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
		Size   int64  `json:"size"`
	}
	inventory := make([]inventoryFile, 0, len(ordered))
	for _, name := range ordered {
		contents := files[name]
		writeFile(t, filepath.Join(releaseRoot, filepath.FromSlash(name)), contents)
		sum := sha256.Sum256(contents)
		inventory = append(inventory, inventoryFile{Path: name, SHA256: hex.EncodeToString(sum[:]), Size: int64(len(contents))})
	}
	document := struct {
		SchemaVersion    int             `json:"schemaVersion"`
		Release          string          `json:"release"`
		IdentityRevision int             `json:"identityRevision"`
		RuntimeVersion   int             `json:"runtimeVersion"`
		CatalogSHA256    string          `json:"catalogSha256"`
		ThemesSHA256     string          `json:"themesSha256"`
		CampaignsSHA256  string          `json:"campaignsSha256"`
		Files            []inventoryFile `json:"files"`
	}{SchemaVersion: 1, Release: "v0.1.1", IdentityRevision: 11, RuntimeVersion: 1, Files: inventory}
	for _, item := range inventory {
		switch item.Path {
		case "catalog.json":
			document.CatalogSHA256 = item.SHA256
		case "themes.json":
			document.ThemesSHA256 = item.SHA256
		case "campaigns.json":
			document.CampaignsSHA256 = item.SHA256
		}
	}
	releaseJSON, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(releaseRoot, "release.json"), releaseJSON)
	checksumNames := append([]string{"release.json"}, ordered...)
	checksums := make([]string, 0, len(checksumNames))
	for _, name := range checksumNames {
		contents, err := os.ReadFile(filepath.Join(releaseRoot, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(contents)
		checksums = append(checksums, hex.EncodeToString(sum[:])+"  "+name)
	}
	writeFile(t, filepath.Join(releaseRoot, "checksums.txt"), []byte(strings.Join(checksums, "\n")+"\n"))
	writeFile(t, filepath.Join(root, "assets", "campaign", "v1.js"), files["campaign/v1.js"])
	for _, channel := range []string{"latest", "default", "current"} {
		writeReleaseChannel(t, root, channel, "https://araihu.com/assets/releases/v0.1.1/themes/base.css")
	}
}

func writeReleaseChannel(t *testing.T, root, channel, cssURL string) {
	t.Helper()
	type theme struct {
		ID     string `json:"id"`
		CSSURL string `json:"cssUrl"`
	}
	value := struct {
		SchemaVersion  int    `json:"schemaVersion"`
		RuntimeVersion int    `json:"runtimeVersion"`
		Release        string `json:"release"`
		Source         string `json:"source"`
		Theme          theme  `json:"theme"`
		Digest         string `json:"digest"`
	}{SchemaVersion: 1, RuntimeVersion: 1, Release: "v0.1.1", Source: "default", Theme: theme{ID: "base", CSSURL: cssURL}}
	payload := canonicalJSON(t, value)
	sum := sha256.Sum256(payload)
	value.Digest = hex.EncodeToString(sum[:])
	writeFile(t, filepath.Join(root, "assets", "releases", channel+".json"), canonicalJSON(t, value))
}

func canonicalJSON(t *testing.T, value any) []byte {
	t.Helper()
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writePNG(t *testing.T, path string, width, height int) {
	t.Helper()
	preview := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.Draw(preview, preview.Bounds(), image.NewUniform(color.NRGBA{R: 7, G: 17, B: 31, A: 255}), image.Point{}, draw.Src)
	writeImagePNG(t, path, preview)
}

func writeImagePNG(t *testing.T, path string, preview image.Image) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, preview); err != nil {
		t.Fatal(err)
	}
}

func mutatePage(t *testing.T, root, route string, mutate func(string) string) {
	t.Helper()
	path := filepath.Join(root, route, "index.html")
	html, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, []byte(mutate(string(html))))
}

func removeTag(t *testing.T, html, prefix string) string {
	t.Helper()
	start := strings.Index(html, prefix)
	if start < 0 {
		t.Fatalf("tag %q not found", prefix)
	}
	end := strings.Index(html[start:], ">")
	if end < 0 {
		t.Fatalf("tag %q is malformed", prefix)
	}
	return html[:start] + html[start+end+1:]
}

func sitemapXML(pages []site.Page, namespace string) []byte {
	var document strings.Builder
	if namespace == "" {
		document.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?><urlset>")
	} else {
		fmt.Fprintf(&document, "<?xml version=\"1.0\" encoding=\"UTF-8\"?><urlset xmlns=\"%s\">", namespace)
	}
	for _, page := range pages {
		fmt.Fprintf(&document, "<url><loc>%s</loc></url>", page.Meta.CanonicalURL)
	}
	document.WriteString("</urlset>")
	return []byte(document.String())
}
