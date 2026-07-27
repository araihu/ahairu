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
	for _, locale := range site.Locales() {
		if err := render(locale); err != nil {
			return err
		}
	}
	return nil
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
