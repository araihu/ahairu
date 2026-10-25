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
	BrandAssetsPublicPrefix   = "/assets/releases/v0.2.5/"
	BrandAssetsRelease        = "v0.2.5"
	BrandIconsGeneratorCommit = "b0f08d46bab278cef30c03cb77a51da8df5d9721"
	BrandCatalogSHA256        = "c448e94d02ee3d8727ffa40fa38fae3319f178547c12376968c3ad58ff0b758b"
	BrandChecksumsSHA256      = "444de3ca511255e9fce04546f8922963e26b6b04b81cff16df79fe813598b014"
	BrandSpriteSHA256         = "2a0832eaa44148c89bb9edc4be768fb2c12058279ad8dec69062f5b0241b5970"
)

//go:embed brand.css
var brandCSS []byte

//go:embed araihu-theme.css
var brandThemeCSS []byte

//go:embed backdrops/*.mp4
var backdropAssets embed.FS

//go:embed visuals/*
var projectVisualAssets embed.FS

//go:embed assets/social/*.jpg
var socialAssets embed.FS

//go:embed storm-backdrop.js
var stormBackdropJS []byte

//go:embed x9-availability.js
var x9AvailabilityJS []byte

//go:embed chart-loader.js
var chartLoaderJS []byte

//go:embed github-releases-openapi.yaml
var githubReleasesOpenAPI string

//go:embed brand-assets
var embeddedBrandAssets embed.FS

// OpenAPILine is one syntax-highlightable line from the embedded GitHub spec excerpt.
type OpenAPILine struct {
	Indent string
	Key    string
	Value  string
}

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
	"release.json",
	"catalog.json",
	"themes.json",
	"campaigns.json",
	"checksums.txt",
	"icons/brand/sprite.svg",
	"licenses/Apache-2.0.txt",
	"licenses/heroicons-MIT.txt",
	"platform/web/araihu/manifest-icons.json",
	"platform/web/balemoh/manifest-icons.json",
	"platform/web/goshtoso/manifest-icons.json",
	"platform/web/manja/manifest-icons.json",
	"platform/web/x9/manifest-icons.json",
}

// BrandCSS returns Arai Hû's project-specific stylesheet.
func BrandCSS() []byte { return brandCSS }

// BrandThemeCSS returns Arai Hû's site-owned Goshtoso theme.
func BrandThemeCSS() []byte { return brandThemeCSS }

// StormBackdropJS returns the progressive enhancement for video playback and parallax.
func StormBackdropJS() []byte { return stormBackdropJS }

// X9AvailabilityJS returns the simulated live-tick enhancement for X-9.
func X9AvailabilityJS() []byte { return x9AvailabilityJS }

// ChartLoaderJS returns the post-reveal chart runtime and swap loader.
func ChartLoaderJS() []byte { return chartLoaderJS }

// GitHubReleaseOpenAPILines returns an official GitHub REST API release-spec excerpt.
func GitHubReleaseOpenAPILines() []OpenAPILine {
	lines := strings.Split(strings.TrimSpace(githubReleasesOpenAPI), "\n")
	result := make([]OpenAPILine, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " ")
		indent := line[:len(line)-len(trimmed)]
		separator := strings.Index(trimmed, ":")
		if separator < 0 {
			result = append(result, OpenAPILine{Indent: indent, Value: trimmed})
			continue
		}
		result = append(result, OpenAPILine{Indent: indent, Key: trimmed[:separator+1], Value: trimmed[separator+1:]})
	}
	return result
}

// BackdropAsset reads one optimized storm backdrop.
func BackdropAsset(name string) ([]byte, error) { return backdropAssets.ReadFile("backdrops/" + name) }

// BackdropAssetNames returns every video emitted by the static builder.
func BackdropAssetNames() []string { return []string{"storm-dark-v2.mp4", "storm-light-v2.mp4"} }

// ProjectVisualAsset reads one project showcase asset.
func ProjectVisualAsset(name string) ([]byte, error) {
	return projectVisualAssets.ReadFile("visuals/" + name)
}

// ProjectVisualAssetNames returns every generated showcase asset.
func ProjectVisualAssetNames() []string {
	return []string{
		"goshtoso-components-montage-v2-av1.mp4",
		"goshtoso-components-montage-v2.mp4",
		"goshtoso-components-poster-v2.webp",
		"margo-icon.svg",
		"muamba-mark.svg",
	}
}

// SocialAsset reads one organization social-preview image.
func SocialAsset(name string) ([]byte, error) { return socialAssets.ReadFile("assets/social/" + name) }

// SocialAssetNames returns every social-preview image emitted by the static builder.
func SocialAssetNames() []string { return []string{"araihu-storm-v1.jpg"} }

// BrandAssets returns the immutable v0.2.5 release subset rooted at its release paths.
func BrandAssets() fs.FS {
	sub, err := fs.Sub(embeddedBrandAssets, "brand-assets")
	if err != nil {
		panic(err)
	}
	return sub
}

// BrandCatalog parses the exact upstream v0.2.5 catalog.
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
	if catalog.SchemaVersion != 2 || catalog.Release != BrandAssetsRelease || catalog.IdentityRevision != 11 {
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
