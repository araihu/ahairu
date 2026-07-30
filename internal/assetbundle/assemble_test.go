package assetbundle

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"os"
	"path"
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

type bundleFixture struct{ files fstest.MapFS }

func fixtureBundle(t *testing.T) *bundleFixture {
	t.Helper()
	files := fstest.MapFS{
		"campaign/v1.js":                  {Data: []byte("(() => {})()\n")},
		"releases/v0.1.1/catalog.json":    {Data: []byte(`{"schemaVersion":1}`)},
		"releases/v0.1.1/themes.json":     {Data: []byte(`{"schemaVersion":1}`)},
		"releases/v0.1.1/campaigns.json":  {Data: []byte(`{"schemaVersion":1,"campaigns":[]}`)},
		"releases/v0.1.1/themes/base.css": {Data: []byte("body{}\n")},
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
	for _, path := range []string{"catalog.json", "themes.json", "campaigns.json", "themes/base.css"} {
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
		Source:         "default",
		Theme:          resolvedTheme{ID: "base", CSSURL: "/assets/releases/" + release + "/themes/base.css"},
	}
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
