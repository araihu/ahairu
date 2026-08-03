// Command ahairu builds the static AraiHu website.
package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/araihu/ahairu/site"
	shellassets "github.com/araihu/goshtoso-app-shells/landingshell/assets"
	chartassets "github.com/araihu/goshtoso-charts/assets"
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
	for _, runtimeAsset := range assets.DefaultRuntimeManifest().Dependencies {
		if runtimeAsset.Enabled {
			if err := writeHandlerAsset("public", assets.Handler(), runtimeAsset.LocalURL); err != nil {
				return err
			}
		}
	}
	if err := writeHandlerAsset("public", shellassets.Handler("/landingshell/assets/"), shellassets.StylesheetURL("/landingshell/assets/")); err != nil {
		return err
	}
	if err := writeHandlerAsset("public", chartassets.Handler(), chartassets.RuntimeURL); err != nil {
		return err
	}
	if err := writeHandlerAsset("public", chartassets.Handler(), chartassets.ThreeDRuntimeURL); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join("public", "assets", "ahairu.css"), site.BrandCSS(), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join("public", "assets", "araihu-theme.css"), site.BrandThemeCSS(), 0o644); err != nil {
		return err
	}
	jsDir := filepath.Join("public", "assets", "js")
	if err := os.MkdirAll(jsDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(jsDir, "storm-backdrop.js"), site.StormBackdropJS(), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(jsDir, "x9-availability.js"), site.X9AvailabilityJS(), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(jsDir, "chart-loader.js"), site.ChartLoaderJS(), 0o644); err != nil {
		return err
	}
	logosDir := filepath.Join("public", "assets", "logos")
	if err := os.MkdirAll(logosDir, 0o755); err != nil {
		return err
	}
	for _, name := range site.BrandAssetNames() {
		asset, err := site.BrandAsset(name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(logosDir, name), asset, 0o644); err != nil {
			return err
		}
	}
	videoDir := filepath.Join("public", "assets", "video")
	if err := os.MkdirAll(videoDir, 0o755); err != nil {
		return err
	}
	for _, name := range site.BackdropAssetNames() {
		asset, err := site.BackdropAsset(name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(videoDir, name), asset, 0o644); err != nil {
			return err
		}
	}
	visualsDir := filepath.Join("public", "assets", "visuals")
	if err := os.MkdirAll(visualsDir, 0o755); err != nil {
		return err
	}
	for _, name := range site.ProjectVisualAssetNames() {
		asset, err := site.ProjectVisualAsset(name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(visualsDir, name), asset, 0o644); err != nil {
			return err
		}
	}
	for _, locale := range site.Locales() {
		if err := render(locale); err != nil {
			return err
		}
		if err := renderChartFragment(locale); err != nil {
			return err
		}
	}
	return nil
}

func writeHandlerAsset(publicRoot string, handler http.Handler, assetURL string) error {
	parsed, err := url.Parse(assetURL)
	if err != nil {
		return fmt.Errorf("parse asset URL %q: %w", assetURL, err)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, parsed.RequestURI(), nil))
	if recorder.Code != http.StatusOK {
		return fmt.Errorf("read embedded asset %q: status %d", assetURL, recorder.Code)
	}
	destination := filepath.Join(publicRoot, filepath.FromSlash(strings.TrimPrefix(parsed.Path, "/")))
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	return os.WriteFile(destination, recorder.Body.Bytes(), 0o644)
}

func render(content site.Content) error {
	destination := filepath.Join("public", content.Path, "index.html")
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()
	return site.Page(content).Render(context.Background(), file)
}

func renderChartFragment(content site.Content) error {
	destination := filepath.Join("public", "fragments", strings.Trim(content.Path, "/"), "charts.html")
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	file, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer file.Close()
	return site.ChartFragment(content).Render(context.Background(), file)
}
