package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunBuildRequiresAssetBundle(t *testing.T) {
	var stderr bytes.Buffer
	if err := run([]string{"build"}, &stderr); err == nil || !strings.Contains(err.Error(), "--asset-bundle") {
		t.Fatalf("run() error = %v, want asset bundle requirement", err)
	}
}

func TestBuildWritesStandaloneSite(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PWD", t.TempDir())
	if err := os.Chdir(os.Getenv("PWD")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	if err := build(fixtureAssetBundle(t)); err != nil {
		t.Fatal(err)
	}
	html, err := os.ReadFile(filepath.Join("public", "en", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	page := string(html)
	if !strings.Contains(page, "Independent software") || !strings.Contains(page, "/assets/styles.css") {
		t.Fatalf("generated HTML misses site content or stylesheet: %s", html)
	}
	for _, landmark := range []string{
		`class="skip-link" href="#main-content"`,
		`<header class="ahairu-header">`,
		`<meta name="theme-color" content="#07111f">`,
		`width="64" height="64"> <span>Arai Hû</span>`,
		`<span>Arai Hû</span>`,
		`<main id="main-content" class="ahairu-shell mx-auto px-6 md:px-10" tabindex="-1">`,
		`href="/en/" aria-current="page"`,
		`aria-label="Primary navigation"`,
		`<h3 class="project-name">X-9</h3>`,
		`<link rel="canonical" href="https://araihu.com/en/">`,
		`<meta property="og:image" content="https://araihu.com/social/brand.png">`,
		`<script id="structured-data" type="application/ld+json">`,
	} {
		if !strings.Contains(page, landmark) {
			t.Errorf("generated HTML misses accessibility landmark %q", landmark)
		}
	}
	for name, want := range map[string]string{
		"robots.txt":       "Sitemap: https://araihu.com/sitemap.xml",
		"sitemap.xml":      "https://araihu.com/license/",
		"site.webmanifest": `"name":"Arai Hû"`,
	} {
		data, err := os.ReadFile(filepath.Join("public", name))
		if err != nil {
			t.Errorf("static discovery file %s missing: %v", name, err)
			continue
		}
		if !strings.Contains(string(data), want) {
			t.Errorf("static discovery file %s misses %q", name, want)
		}
	}
	for _, asset := range []string{
		"/assets/releases/v0.1.1/icons/brand/araihu-icon-adaptive-transparent-optical.svg",
		"/assets/releases/v0.1.1/icons/brand/goshtoso-icon-adaptive-transparent-optical.svg",
		"/assets/releases/v0.1.1/icons/brand/manja-icon-adaptive-transparent-optical.svg",
		"/assets/releases/v0.1.1/icons/brand/paje-icon-adaptive-transparent-optical.svg",
		"/assets/releases/v0.1.1/icons/brand/x9-icon-adaptive-transparent-optical.svg",
	} {
		if !strings.Contains(page, asset) {
			t.Errorf("generated HTML misses brand asset %q", asset)
		}
	}
	if strings.Contains(page, "?rev=a8a9647a") {
		t.Error("generated HTML contains stale V10 revision query")
	}
	if !strings.Contains(page, `href="https://x9.araihu.com"`) {
		t.Error("generated HTML misses the X-9 product URL")
	}
	if strings.Contains(page, "xisnove.dev") {
		t.Error("generated HTML still links to the retired Xisnove domain")
	}
	if info, err := os.Stat(filepath.Join("public", "assets", "styles.css")); err != nil || info.Size() == 0 {
		t.Fatalf("generated stylesheet missing or empty: %v", err)
	}
	if info, err := os.Stat(filepath.Join("public", "assets", "ahairu.css")); err != nil || info.Size() == 0 {
		t.Fatalf("brand stylesheet missing or empty: %v", err)
	}
	if info, err := os.Stat(filepath.Join("public", "assets", "araihu-theme.css")); err != nil || info.Size() == 0 {
		t.Fatalf("Arai Hû theme missing or empty: %v", err)
	}
	for _, name := range []string{"release.json", "catalog.json", "themes.json", "campaigns.json", "checksums.txt"} {
		if info, err := os.Stat(filepath.Join("public", "assets", "releases", "v0.1.1", filepath.FromSlash(name))); err != nil || info.Size() == 0 {
			t.Errorf("brand release asset %s missing or empty: %v", name, err)
		}
	}
	for _, name := range []string{"brand.png", "license.png"} {
		file, err := os.Open(filepath.Join("public", "social", name))
		if err != nil {
			t.Errorf("social preview %s missing: %v", name, err)
			continue
		}
		config, format, decodeErr := image.DecodeConfig(file)
		_ = file.Close()
		if decodeErr != nil || format != "png" || config.Width != 1200 || config.Height != 630 {
			t.Errorf("social preview %s = %s %dx%d, want png 1200x630: %v", name, format, config.Width, config.Height, decodeErr)
		}
	}
	for locale, skipLabel := range map[string]string{"pt-br": "Pular para o conteúdo", "es": "Saltar al contenido"} {
		localizedHTML, err := os.ReadFile(filepath.Join("public", locale, "index.html"))
		if err != nil {
			t.Fatalf("localized page missing for %s: %v", locale, err)
		}
		if !strings.Contains(string(localizedHTML), skipLabel) || !strings.Contains(string(localizedHTML), `aria-current="page"`) {
			t.Errorf("localized page %s misses skip link or locale state", locale)
		}
	}
	for _, path := range []string{"brand", "license", "pt-br/brand", "pt-br/license", "es/brand", "es/license"} {
		data, err := os.ReadFile(filepath.Join("public", path, "index.html"))
		if err != nil {
			t.Errorf("%s missing: %v", path, err)
			continue
		}
		if !strings.Contains(string(data), `<main id="main-content"`) {
			t.Errorf("%s misses shared main landmark", path)
		}
	}
}

func TestBuildRetainsExistingImmutableReleases(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	working := t.TempDir()
	if err := os.Chdir(working); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	historical := filepath.Join("public", "assets", "releases", "v0.1.0", "release.json")
	if err := os.MkdirAll(filepath.Dir(historical), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(historical, []byte("historical release\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join("public", "assets", "releases", "latest.json"), []byte("stale channel\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := build(fixtureAssetBundle(t)); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(historical)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "historical release\n" {
		t.Fatalf("historical release = %q", got)
	}
}

func fixtureAssetBundle(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	releaseRoot := filepath.Join(root, "releases", "v0.1.1")
	for name, contents := range map[string][]byte{
		"catalog.json":    []byte(`{"schemaVersion":1}`),
		"themes.json":     []byte(`{"schemaVersion":1}`),
		"campaigns.json":  []byte(`{"schemaVersion":1,"campaigns":[]}`),
		"themes/base.css": []byte("body{}\n"),
		"campaign/v1.js":  []byte("(() => {})()\n"),
	} {
		file := filepath.Join(root, filepath.FromSlash(name))
		if strings.Contains(name, ".json") || strings.Contains(name, ".css") || name == "campaign/v1.js" {
			file = filepath.Join(releaseRoot, filepath.FromSlash(name))
		}
		if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, contents, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	writeBundleFile(t, filepath.Join(root, "campaign", "v1.js"), []byte("(() => {})()\n"))
	type inventoryFile struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
		Size   int64  `json:"size"`
	}
	var inventory []inventoryFile
	for _, name := range []string{"catalog.json", "themes.json", "campaigns.json", "campaign/v1.js", "themes/base.css"} {
		contents, err := os.ReadFile(filepath.Join(releaseRoot, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
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
	writeBundleFile(t, filepath.Join(releaseRoot, "release.json"), releaseJSON)
	var checksumLines []string
	for _, name := range []string{"release.json", "catalog.json", "themes.json", "campaigns.json", "campaign/v1.js", "themes/base.css"} {
		contents, err := os.ReadFile(filepath.Join(releaseRoot, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(contents)
		checksumLines = append(checksumLines, hex.EncodeToString(sum[:])+"  "+name)
	}
	writeBundleFile(t, filepath.Join(releaseRoot, "checksums.txt"), []byte(strings.Join(checksumLines, "\n")+"\n"))

	type theme struct {
		ID     string `json:"id"`
		CSSURL string `json:"cssUrl"`
	}
	type channel struct {
		SchemaVersion  int    `json:"schemaVersion"`
		RuntimeVersion int    `json:"runtimeVersion"`
		Release        string `json:"release"`
		Source         string `json:"source"`
		Theme          theme  `json:"theme"`
		Digest         string `json:"digest"`
	}
	value := channel{SchemaVersion: 1, RuntimeVersion: 1, Release: "v0.1.1", Source: "default", Theme: theme{ID: "base", CSSURL: "https://araihu.com/assets/releases/v0.1.1/themes/base.css"}}
	channelJSON, err := canonicalJSON(value)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(channelJSON)
	value.Digest = hex.EncodeToString(sum[:])
	channelJSON, err = canonicalJSON(value)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"latest", "default", "current"} {
		writeBundleFile(t, filepath.Join(root, "releases", name+".json"), channelJSON)
	}
	return root
}

func canonicalJSON(value any) ([]byte, error) {
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func writeBundleFile(t *testing.T, name string, contents []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, contents, 0o644); err != nil {
		t.Fatal(err)
	}
}
