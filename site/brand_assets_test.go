package site

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestBundledBrandAssetsMatchReleaseChecksums(t *testing.T) {
	if err := ValidateBundledBrandAssets(); err != nil {
		t.Fatal(err)
	}

	checksums, err := ReleaseChecksums()
	if err != nil {
		t.Fatal(err)
	}
	paths, err := BundledBrandAssetPaths()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(paths), 171; got != want {
		t.Fatalf("bundled asset count = %d, want %d", got, want)
	}
	for _, path := range paths {
		if path == "checksums.txt" {
			continue
		}
		want, ok := checksums[path]
		if !ok {
			t.Errorf("bundled file %q absent from release checksums", path)
			continue
		}
		data, err := fs.ReadFile(BrandAssets(), path)
		if err != nil {
			t.Errorf("read %q: %v", path, err)
			continue
		}
		got := sha256.Sum256(data)
		if hex.EncodeToString(got[:]) != want {
			t.Errorf("checksum %q = %x, want %s", path, got, want)
		}
	}
}

func TestCopyBundledBrandAssetsRejectsHistoricalDestination(t *testing.T) {
	destination := t.TempDir()
	historical := filepath.Join(destination, "concepts", "v10")
	if err := os.MkdirAll(historical, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(historical, "old.svg"), []byte("history"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CopyBundledBrandAssets(destination); err == nil || !strings.Contains(err.Error(), "unexpected destination path") {
		t.Fatalf("CopyBundledBrandAssets() error = %v, want unexpected destination path", err)
	}
}

func TestBundledBrandAssetsAreReleaseSubsetOnly(t *testing.T) {
	paths, err := BundledBrandAssetPaths()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if strings.HasPrefix(path, "concepts/") || strings.HasPrefix(path, "review/") ||
			strings.HasPrefix(path, "proof/") || strings.HasPrefix(path, "releases/") ||
			strings.HasPrefix(path, "icons/ui/") || strings.HasPrefix(path, "platform/android/") ||
			strings.HasPrefix(path, "platform/apple/") {
			t.Errorf("unexpected non-site release path %q", path)
		}
	}
}

func TestMonochromeBrandSpriteSymbolsInheritCurrentColor(t *testing.T) {
	data, err := fs.ReadFile(BrandAssets(), "icons/brand/sprite.svg")
	if err != nil {
		t.Fatal(err)
	}
	sprite := string(data)
	for _, symbol := range []string{
		"araihu-icon-monochrome-transparent-optical",
		"araihu-logo-monochrome-transparent-optical",
	} {
		start := strings.Index(sprite, `<symbol id="`+symbol+`"`)
		if start < 0 {
			t.Fatalf("sprite misses symbol %q", symbol)
		}
		end := strings.Index(sprite[start:], ">")
		if end < 0 || !strings.Contains(sprite[start:start+end], `fill="currentColor"`) {
			t.Errorf("symbol %q does not inherit currentColor", symbol)
		}
	}
}

func TestCatalogUsesPinnedReleaseAndPublicPrefix(t *testing.T) {
	catalog, err := BrandCatalog()
	if err != nil {
		t.Fatal(err)
	}
	if catalog.Release != "v0.1.0" || catalog.IdentityRevision != 11 {
		t.Fatalf("catalog release = %q revision = %d", catalog.Release, catalog.IdentityRevision)
	}
	if BrandAssetsPublicPrefix != "/assets/araihu/v0.1.0/" {
		t.Fatalf("public prefix = %q", BrandAssetsPublicPrefix)
	}
	if got := BrandCatalogSHA256; got != "d83be964fa411e87c61b49f0a0b6a2a1465f33ad43bea7cd93b2e434b59266af" {
		t.Fatalf("catalog hash = %q", got)
	}
	if BrandAssetsReleaseCandidateCommit != "613335f60877d3cc6affd04c51a31ce7fa0e433c" {
		t.Fatalf("release candidate commit = %q", BrandAssetsReleaseCandidateCommit)
	}
	if BrandIconsGeneratorCommit != "d8d58c355a21fc5d17edeb3ef0340a5a3b2d6854" {
		t.Fatalf("icon generator commit = %q", BrandIconsGeneratorCommit)
	}
	if BrandChecksumsSHA256 != "2d83421b3a95c75f68c88af7d5618034b4189d42adf3f2e39b2c4c048c553d5d" {
		t.Fatalf("checksums hash = %q", BrandChecksumsSHA256)
	}
	if BrandSpriteSHA256 != "e0c98a783cf65cf52b0a57cca47b84704499200a7fdb113b751d8f6c5828ba45" {
		t.Fatalf("brand sprite hash = %q", BrandSpriteSHA256)
	}
}

func TestBrandAssetValidationRejectsMutationAndUnexpectedHistory(t *testing.T) {
	files := make(fstest.MapFS)
	if err := fs.WalkDir(BrandAssets(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := fs.ReadFile(BrandAssets(), path)
		if err != nil {
			return err
		}
		files[path] = &fstest.MapFile{Data: data}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	originalChecksums := files["checksums.txt"].Data
	lines := strings.Split(string(originalChecksums), "\n")
	lines[0], lines[1] = lines[1], lines[0]
	files["checksums.txt"].Data = []byte(strings.Join(lines, "\n"))
	if err := validateBrandAssets(files); err == nil || !strings.Contains(err.Error(), "checksums file checksum") {
		t.Fatalf("reordered checksums error = %v, want provenance hash failure", err)
	}
	files["checksums.txt"].Data = originalChecksums

	files["icons/brand/araihu-icon-light-transparent-optical.svg"].Data = []byte("corrupt")
	if err := validateBrandAssets(files); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("mutated asset error = %v, want checksum failure", err)
	}
	delete(files, "icons/brand/araihu-icon-light-transparent-optical.svg")
	files["concepts/v10/old.svg"] = &fstest.MapFile{Data: []byte("history")}
	if err := validateBrandAssets(files); err == nil || !strings.Contains(err.Error(), "unexpected bundled release path") {
		t.Fatalf("historical asset error = %v, want unexpected-path failure", err)
	}
}

func TestEveryHomeDownloadExistsAndMatchesCatalog(t *testing.T) {
	catalog, err := BrandCatalog()
	if err != nil {
		t.Fatal(err)
	}
	byPath := make(map[string]CatalogAsset, len(catalog.Assets))
	for _, asset := range catalog.Assets {
		byPath[asset.Path] = asset
	}
	for _, page := range Pages() {
		if page.Home == nil {
			continue
		}
		for _, project := range page.Home.Projects {
			path := strings.TrimPrefix(project.MarkURL, BrandAssetsPublicPrefix)
			asset, ok := byPath[path]
			if !ok {
				t.Errorf("%s mark %q absent from catalog", project.Name, path)
				continue
			}
			data, err := fs.ReadFile(BrandAssets(), path)
			if err != nil {
				t.Errorf("%s mark %q missing: %v", project.Name, path, err)
				continue
			}
			got := sha256.Sum256(data)
			if hex.EncodeToString(got[:]) != asset.SHA256 {
				t.Errorf("%s mark hash = %x, want %s", project.Name, got, asset.SHA256)
			}
		}
	}
}
