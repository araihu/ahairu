// Command ahairu builds the static AraiHu website.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/araihu/ahairu/site"
	"github.com/araihu/goshtoso/assets"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] != "build" {
		fmt.Fprintln(os.Stderr, "usage: ahairu build")
		os.Exit(2)
	}
	if err := build(); err != nil {
		fmt.Fprintf(os.Stderr, "build site: %v\n", err)
		os.Exit(1)
	}
}

func build() error {
	if err := os.MkdirAll("public/assets", 0o755); err != nil {
		return err
	}
	css, err := assets.StylesCSS()
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join("public", "assets", "styles.css"), css, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join("public", "assets", "ahairu.css"), site.BrandCSS(), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join("public", "assets", "araihu-theme.css"), site.BrandThemeCSS(), 0o644); err != nil {
		return err
	}
	releaseDir := filepath.Join("public", "assets", "araihu", "v0.1.0")
	if err := site.CopyBundledBrandAssets(releaseDir); err != nil {
		return fmt.Errorf("copy Arai Hû assets v0.1.0: %w", err)
	}
	if err := site.CopyBundledSocialImages(filepath.Join("public", "social")); err != nil {
		return fmt.Errorf("copy social previews: %w", err)
	}
	for _, page := range site.Pages() {
		if err := render(page); err != nil {
			return err
		}
	}
	if err := writeStaticPages(site.Pages()); err != nil {
		return err
	}
	return nil
}

func render(page site.Page) error {
	destination := filepath.Join("public", page.Meta.Path, "index.html")
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()
	component, err := site.PageComponent(page)
	if err != nil {
		return fmt.Errorf("select renderer for %q: %w", page.Meta.Path, err)
	}
	return component.Render(context.Background(), file)
}

func writeStaticPages(pages []site.Page) error {
	sitemap, err := site.Sitemap(pages)
	if err != nil {
		return err
	}
	for name, contents := range map[string][]byte{
		"robots.txt":       site.Robots(),
		"sitemap.xml":      sitemap,
		"site.webmanifest": site.SiteManifest(),
	} {
		if err := os.WriteFile(filepath.Join("public", name), contents, 0o644); err != nil {
			return err
		}
	}
	return nil
}
