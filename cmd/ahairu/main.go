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
	file, err := os.Create(filepath.Join("public", "index.html"))
	if err != nil {
		return err
	}
	defer file.Close()
	return site.Page().Render(context.Background(), file)
}
