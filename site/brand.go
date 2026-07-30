package site

import (
	"bufio"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	BrandAssetsPublicPrefix   = "/assets/araihu/v0.1.0/"
	BrandAssetsRelease        = "v0.1.0"
	BrandIconsGeneratorCommit = "d8d58c355a21fc5d17edeb3ef0340a5a3b2d6854"
	BrandCatalogSHA256        = "d83be964fa411e87c61b49f0a0b6a2a1465f33ad43bea7cd93b2e434b59266af"
	BrandChecksumsSHA256      = "2d83421b3a95c75f68c88af7d5618034b4189d42adf3f2e39b2c4c048c553d5d"
	BrandSpriteSHA256         = "e0c98a783cf65cf52b0a57cca47b84704499200a7fdb113b751d8f6c5828ba45"
)

//go:embed brand.css
var brandCSS []byte

//go:embed araihu-theme.css
var brandThemeCSS []byte

//go:embed brand-assets
var embeddedBrandAssets embed.FS

// Catalog describes the pinned Arai Hû assets release.
type Catalog struct {
	SchemaVersion    int            `json:"schemaVersion"`
	Release          string         `json:"release"`
	IdentityRevision int            `json:"identityRevision"`
	Assets           []CatalogAsset `json:"assets"`
}

// CatalogAsset is one distributable file addressed by catalog.json.
type CatalogAsset struct {
	CanonicalName string `json:"canonicalName"`
	Namespace     string `json:"namespace"`
	Path          string `json:"path"`
	Product       string `json:"product"`
	SpriteSymbol  string `json:"spriteSymbol"`
	SHA256        string `json:"sha256"`
}

var releaseSupportPaths = []string{
	"NOTICE",
	"catalog.json",
	"checksums.txt",
	"icons/brand/sprite.svg",
	"licenses/Apache-2.0.txt",
	"licenses/heroicons-MIT.txt",
	"platform/web/araihu/manifest-icons.json",
	"platform/web/goshtoso/manifest-icons.json",
	"platform/web/manja/manifest-icons.json",
	"platform/web/paje/manifest-icons.json",
	"platform/web/x9/manifest-icons.json",
}

// BrandCSS returns Arai Hû's project-specific stylesheet.
func BrandCSS() []byte { return brandCSS }

// BrandThemeCSS returns Arai Hû's site-owned Goshtoso theme.
func BrandThemeCSS() []byte { return brandThemeCSS }

// BrandAssets returns the immutable v0.1.0 release subset rooted at its release paths.
func BrandAssets() fs.FS {
	sub, err := fs.Sub(embeddedBrandAssets, "brand-assets")
	if err != nil {
		panic(err)
	}
	return sub
}

// BrandCatalog parses the exact upstream v0.1.0 catalog.
func BrandCatalog() (Catalog, error) {
	return brandCatalog(BrandAssets())
}

func brandCatalog(fsys fs.FS) (Catalog, error) {
	data, err := fs.ReadFile(fsys, "catalog.json")
	if err != nil {
		return Catalog{}, err
	}
	got := sha256.Sum256(data)
	if hex.EncodeToString(got[:]) != BrandCatalogSHA256 {
		return Catalog{}, fmt.Errorf("catalog checksum = %x, want %s", got, BrandCatalogSHA256)
	}
	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return Catalog{}, fmt.Errorf("parse catalog: %w", err)
	}
	if catalog.SchemaVersion != 1 || catalog.Release != "v0.1.0" || catalog.IdentityRevision != 11 {
		return Catalog{}, fmt.Errorf("unexpected catalog schema/release/revision: %d/%s/%d", catalog.SchemaVersion, catalog.Release, catalog.IdentityRevision)
	}
	return catalog, nil
}

// ReleaseChecksums parses the upstream checksums file without requiring excluded release files.
func ReleaseChecksums() (map[string]string, error) {
	return releaseChecksums(BrandAssets())
}

func releaseChecksums(fsys fs.FS) (map[string]string, error) {
	data, err := fs.ReadFile(fsys, "checksums.txt")
	if err != nil {
		return nil, err
	}
	got := sha256.Sum256(data)
	if hex.EncodeToString(got[:]) != BrandChecksumsSHA256 {
		return nil, fmt.Errorf("checksums file checksum = %x, want %s", got, BrandChecksumsSHA256)
	}
	checksums := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 || len(parts[0]) != 64 || !fs.ValidPath(parts[1]) {
			return nil, fmt.Errorf("invalid checksum line %q", line)
		}
		if _, err := hex.DecodeString(parts[0]); err != nil {
			return nil, fmt.Errorf("invalid checksum %q: %w", parts[0], err)
		}
		if _, duplicate := checksums[parts[1]]; duplicate {
			return nil, fmt.Errorf("duplicate checksum path %q", parts[1])
		}
		checksums[parts[1]] = parts[0]
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return checksums, nil
}

// BundledBrandAssetPaths returns catalog-selected website assets plus fixed release support files.
func BundledBrandAssetPaths() ([]string, error) {
	return bundledBrandAssetPaths(BrandAssets())
}

func bundledBrandAssetPaths(fsys fs.FS) ([]string, error) {
	catalog, err := brandCatalog(fsys)
	if err != nil {
		return nil, err
	}
	selected := make(map[string]struct{})
	for _, asset := range catalog.Assets {
		if strings.HasPrefix(asset.Path, "brand/") || strings.HasPrefix(asset.Path, "icons/brand/") || strings.HasPrefix(asset.Path, "platform/web/") {
			selected[asset.Path] = struct{}{}
		}
	}
	for _, path := range releaseSupportPaths {
		selected[path] = struct{}{}
	}
	paths := make([]string, 0, len(selected))
	for path := range selected {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

// ValidateBundledBrandAssets proves exact subset membership and upstream checksums.
func ValidateBundledBrandAssets() error {
	return validateBrandAssets(BrandAssets())
}

func validateBrandAssets(fsys fs.FS) error {
	paths, err := bundledBrandAssetPaths(fsys)
	if err != nil {
		return err
	}
	expected := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		expected[path] = struct{}{}
	}
	var actual []string
	if err := fs.WalkDir(fsys, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			actual = append(actual, path)
		}
		return nil
	}); err != nil {
		return err
	}
	sort.Strings(actual)
	if len(actual) != len(paths) {
		return fmt.Errorf("bundled file count = %d, want %d", len(actual), len(paths))
	}
	for _, path := range actual {
		if _, ok := expected[path]; !ok {
			return fmt.Errorf("unexpected bundled release path %q", path)
		}
	}
	checksums, err := releaseChecksums(fsys)
	if err != nil {
		return err
	}
	for _, path := range paths {
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return fmt.Errorf("read bundled release path %q: %w", path, err)
		}
		if path == "checksums.txt" {
			continue
		}
		want, ok := checksums[path]
		if !ok {
			return fmt.Errorf("bundled release path %q absent from checksums", path)
		}
		got := sha256.Sum256(data)
		if hex.EncodeToString(got[:]) != want {
			return fmt.Errorf("checksum %q = %x, want %s", path, got, want)
		}
	}
	return nil
}

// CopyBundledBrandAssets writes the validated release subset below destination.
func CopyBundledBrandAssets(destination string) error {
	if err := ValidateBundledBrandAssets(); err != nil {
		return err
	}
	paths, err := BundledBrandAssetPaths()
	if err != nil {
		return err
	}
	expected := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		expected[path] = struct{}{}
	}
	if _, err := os.Stat(destination); err == nil {
		if err := filepath.WalkDir(destination, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			relative, err := filepath.Rel(destination, path)
			if err != nil {
				return err
			}
			relative = filepath.ToSlash(relative)
			if _, ok := expected[relative]; !ok {
				return fmt.Errorf("unexpected destination path %q", relative)
			}
			if !entry.Type().IsRegular() {
				return fmt.Errorf("destination path %q is not a regular file", relative)
			}
			return nil
		}); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	for _, path := range paths {
		data, err := fs.ReadFile(BrandAssets(), path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}
