package site

import _ "embed"

//go:embed brand.css
var brandCSS []byte

// BrandCSS returns Arai Hu's project-specific stylesheet.
func BrandCSS() []byte { return brandCSS }
