package site

import (
	"fmt"

	"github.com/a-h/templ"
)

// PageComponent dispatches one validated typed page to its static renderer.
func PageComponent(page Page) (templ.Component, error) {
	if err := validatePage(page); err != nil {
		return nil, err
	}
	switch page.Meta.Kind {
	case PageHome:
		return HomePage(page), nil
	case PageBrand:
		return BrandPage(page), nil
	case PageLicense:
		return LicensePage(page), nil
	default:
		return nil, fmt.Errorf("page %q has unknown kind %q", page.Meta.Path, page.Meta.Kind)
	}
}

func formatIndex(index int) string { return fmt.Sprintf("%02d", index+1) }
