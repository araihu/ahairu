// Command ahairu builds the static AraiHu website.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/araihu/ahairu/internal/assetbundle"
	"github.com/araihu/ahairu/site"
	"github.com/araihu/goshtoso/assets"
)

func main() {
	if err := run(os.Args[1:], os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "build site: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string, stderr io.Writer) error {
	if len(args) == 0 || args[0] != "build" {
		return errors.New("usage: ahairu build --asset-bundle <directory>")
	}
	flags := flag.NewFlagSet("build", flag.ContinueOnError)
	flags.SetOutput(stderr)
	assetBundle := flags.String("asset-bundle", "", "verified Assets bundle directory")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("usage: ahairu build --asset-bundle <directory>")
	}
	if *assetBundle == "" {
		return errors.New("build: --asset-bundle is required")
	}
	return build(*assetBundle)
}

func build(assetBundle string) error {
	input, err := os.Stat(assetBundle)
	if err != nil {
		return fmt.Errorf("stat asset bundle %q: %w", assetBundle, err)
	}
	if !input.IsDir() {
		return fmt.Errorf("asset bundle %q is not a directory", assetBundle)
	}
	staging, err := os.MkdirTemp(".", ".ahairu-build-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)
	if err := buildAt(staging, os.DirFS(assetBundle)); err != nil {
		return err
	}
	return replacePublic(filepath.Join(staging, "public"))
}

func replacePublic(staging string) error {
	if _, err := os.Lstat("public"); errors.Is(err, fs.ErrNotExist) {
		return os.Rename(staging, "public")
	} else if err != nil {
		return fmt.Errorf("inspect prior public tree: %w", err)
	}
	backup, err := os.MkdirTemp(".", ".ahairu-public-")
	if err != nil {
		return err
	}
	if err := os.Remove(backup); err != nil {
		return err
	}
	if err := os.Rename("public", backup); err != nil {
		return fmt.Errorf("stage prior public tree: %w", err)
	}
	if err := os.Rename(staging, "public"); err != nil {
		if restoreErr := os.Rename(backup, "public"); restoreErr != nil {
			return fmt.Errorf("promote staged public tree: %w; restore prior public tree: %v", err, restoreErr)
		}
		return fmt.Errorf("promote staged public tree: %w", err)
	}
	if err := os.RemoveAll(backup); err != nil {
		return fmt.Errorf("remove prior public tree: %w", err)
	}
	return nil
}

func buildAt(output string, bundle fs.FS) error {
	assetsDirectory := filepath.Join(output, "public", "assets")
	if err := os.MkdirAll(assetsDirectory, 0o755); err != nil {
		return err
	}
	css, err := assets.StylesCSS()
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(assetsDirectory, "styles.css"), css, 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(assetsDirectory, "ahairu.css"), site.BrandCSS(), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(assetsDirectory, "araihu-theme.css"), site.BrandThemeCSS(), 0o644); err != nil {
		return err
	}
	root, err := os.OpenRoot(assetsDirectory)
	if err != nil {
		return err
	}
	defer root.Close()
	if err := assetbundle.Assemble(context.Background(), bundle, root); err != nil {
		return fmt.Errorf("assemble verified asset bundle: %w", err)
	}
	releaseDir := filepath.Join(assetsDirectory, "araihu", "v0.1.0")
	if err := site.CopyBundledBrandAssets(releaseDir); err != nil {
		return fmt.Errorf("copy Arai Hû assets v0.1.0: %w", err)
	}
	if err := site.CopyBundledSocialImages(filepath.Join(output, "public", "social")); err != nil {
		return fmt.Errorf("copy social previews: %w", err)
	}
	for _, page := range site.Pages() {
		if err := render(output, page); err != nil {
			return err
		}
	}
	if err := writeStaticPages(output, site.Pages()); err != nil {
		return err
	}
	return nil
}

func render(output string, page site.Page) error {
	destination := filepath.Join(output, "public", page.Meta.Path, "index.html")
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

func writeStaticPages(output string, pages []site.Page) error {
	sitemap, err := site.Sitemap(pages)
	if err != nil {
		return err
	}
	for name, contents := range map[string][]byte{
		"robots.txt":       site.Robots(),
		"sitemap.xml":      sitemap,
		"site.webmanifest": site.SiteManifest(),
	} {
		if err := os.WriteFile(filepath.Join(output, "public", name), contents, 0o644); err != nil {
			return err
		}
	}
	return nil
}
