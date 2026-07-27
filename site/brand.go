package site

import (
	"embed"
)

//go:embed brand.css
var brandCSS []byte

//go:embed brand-assets/logos/*.svg
var brandAssets embed.FS

//go:embed brand-assets/themes/araihu.css
var brandThemeCSS []byte

// BrandCSS returns Arai Hu's project-specific stylesheet.
func BrandCSS() []byte { return brandCSS }

// BrandThemeCSS returns the vendored Arai Hû Goshtoso theme.
func BrandThemeCSS() []byte { return brandThemeCSS }

// BrandAsset reads one embedded organization or project mark.
func BrandAsset(name string) ([]byte, error) {
	return brandAssets.ReadFile("brand-assets/logos/" + name)
}

// BrandAssetNames returns every asset emitted by the static builder.
func BrandAssetNames() []string {
	return []string{"araihu-favicon.svg", "araihu-mark.svg", "goshtoso-mark.svg", "manja-mark.svg", "paje-mark.svg"}
}
