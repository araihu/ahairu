package site

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var socialImageNames = []string{"brand.png", "license.png"}

//go:embed social/brand.png social/license.png
var embeddedSocialImages embed.FS

// CopyBundledSocialImages writes the immutable social preview images to destination.
func CopyBundledSocialImages(destination string) error {
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return fmt.Errorf("create social image directory: %w", err)
	}
	for _, name := range socialImageNames {
		data, err := fs.ReadFile(embeddedSocialImages, "social/"+name)
		if err != nil {
			return fmt.Errorf("read bundled social image %q: %w", name, err)
		}
		if err := os.WriteFile(filepath.Join(destination, name), data, 0o644); err != nil {
			return fmt.Errorf("write social image %q: %w", name, err)
		}
	}
	return nil
}
