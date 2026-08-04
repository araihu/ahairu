package assetbundle

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"testing/fstest"
)

func TestValidateRejectsChannelOutsideImmutableRelease(t *testing.T) {
	input := fixtureBundle(t)
	input.Write("releases/current.json", []byte(`{"schemaVersion":1,"runtimeVersion":1,"release":"v0.1.1","source":"default","theme":{"id":"base","cssUrl":"/private/theme.css"},"digest":"`+strings.Repeat("0", 64)+`"}`))
	if _, err := Validate(input.FS()); err == nil || !strings.Contains(err.Error(), "immutable release") {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsChecksumMismatch(t *testing.T) {
	input := fixtureBundle(t)
	input.Write("releases/v0.1.1/catalog.json", []byte(`{"changed":true}`))
	if _, err := Validate(input.FS()); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateChannelReleaseURLsAcceptTrustedOriginsOnly(t *testing.T) {
	for _, test := range []struct {
		name  string
		url   string
		valid bool
	}{
		{"root relative", "/assets/releases/v0.1.1/themes/base.css", true},
		{"trusted absolute", "https://araihu.com/assets/releases/v0.1.1/themes/base.css", true},
		{"cross origin", "https://evil.example/assets/releases/v0.1.1/themes/base.css", false},
		{"insecure", "http://araihu.com/assets/releases/v0.1.1/themes/base.css", false},
		{"scheme relative", "//araihu.com/assets/releases/v0.1.1/themes/base.css", false},
		{"script scheme", "javascript:alert(1)", false},
		{"query", "https://araihu.com/assets/releases/v0.1.1/themes/base.css?x=1", false},
		{"control", "https://araihu.com/assets/releases/v0.1.1/themes/base.css\n", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := fixtureBundle(t)
			document := input.channelDocument(t, "current")
			document.Theme.CSSURL = test.url
			input.writeChannel(t, "current", document)
			_, err := Validate(input.FS())
			if test.valid && err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("Validate() error = nil")
			}
		})
	}
}

func TestValidateCampaignEnforcesResolvedIconSemantics(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func(*channelDocument)
	}{
		{"invalid icon id", func(document *channelDocument) { document.Campaign.Toggle.EnabledIcon.ID = "bad icon" }},
		{"invalid icon mode", func(document *channelDocument) { document.Campaign.Toggle.EnabledIcon.Mode = "inline" }},
		{"sprite misses symbol", func(document *channelDocument) { document.Campaign.Toggle.EnabledIcon.SpriteSymbol = "missing-symbol" }},
		{"asset has symbol", func(document *channelDocument) { document.Campaign.Toggle.DisabledIcon.SpriteSymbol = "ui-moon" }},
		{"asset has sprite URL", func(document *channelDocument) {
			document.Campaign.Toggle.DisabledIcon.URL = "https://araihu.com/assets/releases/v0.1.1/icons/ui/sprite.svg"
		}},
		{"sprite has asset URL", func(document *channelDocument) {
			document.Campaign.Toggle.EnabledIcon.URL = "https://araihu.com/assets/releases/v0.1.1/icons/ui/sun.svg"
		}},
		{"invalid brand id", func(document *channelDocument) { document.Campaign.Brand.Logo.ID = "bad logo" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := fixtureBundle(t)
			document := input.channelDocument(t, "current")
			test.mutate(&document)
			input.writeChannel(t, "current", document)
			if _, err := Validate(input.FS()); err == nil {
				t.Fatal("Validate() error = nil")
			}
		})
	}
}

func TestValidateRequiresRootCampaignRuntimeToMatchCurrentRelease(t *testing.T) {
	input := fixtureBundle(t)
	input.Write("campaign/v1.js", []byte("changed runtime\n"))
	if _, err := Validate(input.FS()); err == nil || !strings.Contains(err.Error(), "does not match current immutable release") {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestValidateRejectsReleaseFileOutsideInventory(t *testing.T) {
	input := fixtureBundle(t)
	input.Write("releases/v0.1.1/extra.svg", []byte("unexpected"))
	input.rewriteChecksums(t, "v0.1.1")
	if _, err := Validate(input.FS()); err == nil || !strings.Contains(err.Error(), "inventory") {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestAssembleRetainsEveryImmutableRelease(t *testing.T) {
	destination := rootedFixture(t, "releases/v0.1.0/release.json", []byte("old"))
	if err := Assemble(context.Background(), fixtureV011(t), destination); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"releases/v0.1.0/release.json", "releases/v0.1.1/release.json"} {
		if _, err := destination.Stat(name); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestAssembleRejectsDifferentByteCollision(t *testing.T) {
	destination := rootedFixture(t, "releases/v0.1.1/catalog.json", []byte("old"))
	if err := Assemble(context.Background(), fixtureV011(t), destination); err == nil || !strings.Contains(err.Error(), "collision") || !strings.Contains(err.Error(), "releases/v0.1.1/catalog.json") {
		t.Fatalf("Assemble() error = %v", err)
	}
}

func TestAssemblePublishesVerifiedSnapshotWhenSourceBecomesSymlink(t *testing.T) {
	directory := t.TempDir()
	writeFixtureToDirectory(t, fixtureBundle(t), directory)
	foreign := filepath.Join(t.TempDir(), "foreign-runtime.js")
	if err := os.WriteFile(foreign, []byte("foreign runtime\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	source := &swapAfterOpenFS{FS: os.DirFS(directory), target: filepath.Join(directory, "campaign", "v1.js"), foreign: foreign}
	destination := rootedFixture(t, "placeholder", []byte("unused"))
	if err := Assemble(context.Background(), source, destination); err != nil {
		t.Fatal(err)
	}
	published, err := destination.ReadFile("campaign/v1.js")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(published, []byte("(() => {})()\n")) {
		t.Fatalf("published runtime = %q, want verified runtime", published)
	}
}

type swapAfterOpenFS struct {
	fs.FS
	target, foreign string
	swapped         bool
}

func (source *swapAfterOpenFS) Open(name string) (fs.File, error) {
	file, err := source.FS.Open(name)
	if err != nil || name != "campaign/v1.js" || source.swapped {
		return file, err
	}
	source.swapped = true
	if err := os.Remove(source.target); err != nil {
		_ = file.Close()
		return nil, err
	}
	if err := os.Symlink(source.foreign, source.target); err != nil {
		_ = file.Close()
		return nil, err
	}
	return file, nil
}

type bundleFixture struct{ files fstest.MapFS }

func fixtureBundle(t *testing.T) *bundleFixture {
	t.Helper()
	files := fstest.MapFS{
		"campaign/v1.js":                       {Data: []byte("(() => {})()\n")},
		"releases/v0.1.1/catalog.json":         {Data: []byte(`{"schemaVersion":1}`)},
		"releases/v0.1.1/themes.json":          {Data: []byte(`{"schemaVersion":1}`)},
		"releases/v0.1.1/campaigns.json":       {Data: []byte(`{"schemaVersion":1,"campaigns":[]}`)},
		"releases/v0.1.1/themes/base.css":      {Data: []byte("body{}\n")},
		"releases/v0.1.1/campaign/v1.js":       {Data: []byte("(() => {})()\n")},
		"releases/v0.1.1/icons/ui/sprite.svg":  {Data: []byte(`<svg><symbol id="ui-sun"/></svg>\n`)},
		"releases/v0.1.1/icons/ui/moon.svg":    {Data: []byte("moon\n")},
		"releases/v0.1.1/brand/logo.svg":       {Data: []byte("logo\n")},
		"releases/v0.1.1/icons/brand/icon.svg": {Data: []byte("icon\n")},
	}
	fixture := &bundleFixture{files: files}
	fixture.release(t, "v0.1.1")
	channel := fixture.channel(t, "v0.1.1")
	for _, name := range []string{"latest", "default", "current"} {
		fixture.Write("releases/"+name+".json", channel)
	}
	return fixture
}

func fixtureV011(t *testing.T) fs.FS { return fixtureBundle(t).FS() }

func (f *bundleFixture) FS() fs.FS { return f.files }

func (f *bundleFixture) Write(name string, data []byte) { f.files[name] = &fstest.MapFile{Data: data} }

func (f *bundleFixture) release(t *testing.T, release string) {
	t.Helper()
	prefix := "releases/" + release + "/"
	type file struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
		Size   int64  `json:"size"`
	}
	var inventory []file
	for _, path := range []string{"catalog.json", "themes.json", "campaigns.json", "brand/logo.svg", "campaign/v1.js", "icons/brand/icon.svg", "icons/ui/moon.svg", "icons/ui/sprite.svg", "themes/base.css"} {
		data := f.files[prefix+path].Data
		sum := sha256.Sum256(data)
		inventory = append(inventory, file{Path: path, SHA256: hex.EncodeToString(sum[:]), Size: int64(len(data))})
	}
	document := struct {
		SchemaVersion    int    `json:"schemaVersion"`
		Release          string `json:"release"`
		IdentityRevision int    `json:"identityRevision"`
		RuntimeVersion   int    `json:"runtimeVersion"`
		CatalogSHA256    string `json:"catalogSha256"`
		ThemesSHA256     string `json:"themesSha256"`
		CampaignsSHA256  string `json:"campaignsSha256"`
		Files            []file `json:"files"`
	}{SchemaVersion: 1, Release: release, IdentityRevision: 11, RuntimeVersion: 1, Files: inventory}
	for _, entry := range inventory {
		switch entry.Path {
		case "catalog.json":
			document.CatalogSHA256 = entry.SHA256
		case "themes.json":
			document.ThemesSHA256 = entry.SHA256
		case "campaigns.json":
			document.CampaignsSHA256 = entry.SHA256
		}
	}
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	f.Write(prefix+"release.json", data)

	f.rewriteChecksums(t, release)
}

func (f *bundleFixture) rewriteChecksums(t *testing.T, release string) {
	t.Helper()
	prefix := "releases/" + release + "/"
	var paths []string
	for name := range f.files {
		if strings.HasPrefix(name, prefix) && name != prefix+"checksums.txt" {
			paths = append(paths, strings.TrimPrefix(name, prefix))
		}
	}
	sort.Strings(paths)
	checksums := make([]string, 0, len(paths))
	for _, name := range paths {
		data := f.files[prefix+name].Data
		sum := sha256.Sum256(data)
		checksums = append(checksums, hex.EncodeToString(sum[:])+"  "+name)
	}
	f.Write(prefix+"checksums.txt", []byte(strings.Join(checksums, "\n")+"\n"))
}

func (f *bundleFixture) channel(t *testing.T, release string) []byte {
	t.Helper()
	document := channelDocument{
		SchemaVersion:  1,
		RuntimeVersion: 1,
		Release:        release,
		Source:         "campaign",
		Theme:          resolvedTheme{ID: "base", CSSURL: "https://araihu.com/assets/releases/" + release + "/themes/base.css"},
		Campaign: &resolvedCampaign{ID: "summer-2026", Toggle: resolvedToggle{
			EnabledIcon:  resolvedIcon{ID: "ui-sun", Mode: "sprite", URL: "https://araihu.com/assets/releases/" + release + "/icons/ui/sprite.svg", SpriteSymbol: "ui-sun"},
			DisabledIcon: resolvedIcon{ID: "ui-moon", Mode: "asset", URL: "https://araihu.com/assets/releases/" + release + "/icons/ui/moon.svg"},
		}, Brand: resolvedBrand{
			Logo: resolvedAsset{ID: "araihu-logo", URL: "https://araihu.com/assets/releases/" + release + "/brand/logo.svg"},
			Icon: resolvedAsset{ID: "araihu-icon", URL: "https://araihu.com/assets/releases/" + release + "/icons/brand/icon.svg"},
		}},
	}
	return encodeChannel(t, document)
}

func (f *bundleFixture) writeChannel(t *testing.T, name string, document channelDocument) {
	t.Helper()
	f.Write("releases/"+name+".json", encodeChannel(t, document))
}

func encodeChannel(t *testing.T, document channelDocument) []byte {
	t.Helper()
	document.Digest = ""
	data, err := encodeCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	document.Digest = hex.EncodeToString(sum[:])
	data, err = encodeCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func (f *bundleFixture) channelDocument(t *testing.T, name string) channelDocument {
	t.Helper()
	var document channelDocument
	if err := decodeStrict(f.files["releases/"+name+".json"].Data, &document); err != nil {
		t.Fatal(err)
	}
	return document
}

func writeFixtureToDirectory(t *testing.T, fixture *bundleFixture, directory string) {
	t.Helper()
	for name, file := range fixture.files {
		path := filepath.Join(directory, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, file.Data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func rootedFixture(t *testing.T, name string, data []byte) *os.Root {
	t.Helper()
	directory := t.TempDir()
	root, err := os.OpenRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	if err := root.MkdirAll(path.Dir(name), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := root.WriteFile(name, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}
