package site

import (
	"embed"
	"strings"
)

//go:embed brand.css
var brandCSS []byte

//go:embed brand-assets/logos/*.svg
var brandAssets embed.FS

//go:embed brand-assets/themes/araihu.css
var brandThemeCSS []byte

//go:embed backdrops/*.mp4
var backdropAssets embed.FS

//go:embed visuals/*
var projectVisualAssets embed.FS

//go:embed storm-backdrop.js
var stormBackdropJS []byte

//go:embed x9-availability.js
var x9AvailabilityJS []byte

//go:embed chart-loader.js
var chartLoaderJS []byte

//go:embed github-releases-openapi.yaml
var githubReleasesOpenAPI string

// OpenAPILine is one syntax-highlightable line from the embedded GitHub spec excerpt.
type OpenAPILine struct {
	Indent string
	Key    string
	Value  string
}

// BrandCSS returns Arai Hu's project-specific stylesheet.
func BrandCSS() []byte { return brandCSS }

// BrandThemeCSS returns the vendored Arai Hû Goshtoso theme.
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
		result = append(result, OpenAPILine{
			Indent: indent,
			Key:    trimmed[:separator+1],
			Value:  trimmed[separator+1:],
		})
	}
	return result
}

// BrandAsset reads one embedded organization or project mark.
func BrandAsset(name string) ([]byte, error) {
	return brandAssets.ReadFile("brand-assets/logos/" + name)
}

// BrandAssetNames returns every asset emitted by the static builder.
func BrandAssetNames() []string {
	return []string{"araihu-icon-background.svg", "araihu-icon-transparent.svg", "goshtoso-icon-transparent.svg", "manja-icon-transparent.svg", "muamba-mark.svg", "paje-icon-transparent.svg", "x9-icon-transparent.svg"}
}

// BackdropAsset reads one optimized storm backdrop.
func BackdropAsset(name string) ([]byte, error) {
	return backdropAssets.ReadFile("backdrops/" + name)
}

// BackdropAssetNames returns every video emitted by the static builder.
func BackdropAssetNames() []string {
	return []string{"storm-dark-v1.mp4", "storm-light-v1.mp4"}
}

// ProjectVisualAsset reads one generated project showcase asset.
func ProjectVisualAsset(name string) ([]byte, error) {
	return projectVisualAssets.ReadFile("visuals/" + name)
}

// ProjectVisualAssetNames returns every generated showcase asset.
func ProjectVisualAssetNames() []string {
	return []string{"goshtoso-components-montage-v1.mp4", "goshtoso-components-poster-v1.webp"}
}
