package site

import (
	"bytes"
	"embed"
	"encoding/binary"
	"fmt"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
)

var socialImageNames = []string{"brand.png", "license.png"}

//go:embed social/brand.png social/license.png
var embeddedSocialImages embed.FS

// ValidateSocialPreviewPNG verifies the exact social-preview PNG contract.
func ValidateSocialPreviewPNG(data []byte) error {
	if len(data) < 33 || !bytes.Equal(data[:8], []byte("\x89PNG\r\n\x1a\n")) {
		return fmt.Errorf("invalid PNG signature or truncated header")
	}
	if binary.BigEndian.Uint32(data[8:12]) != 13 || string(data[12:16]) != "IHDR" {
		return fmt.Errorf("PNG must begin with a 13-byte IHDR chunk")
	}
	width := binary.BigEndian.Uint32(data[16:20])
	height := binary.BigEndian.Uint32(data[20:24])
	if width != 1200 || height != 630 {
		return fmt.Errorf("dimensions are %dx%d, want 1200x630", width, height)
	}
	if data[24] != 8 {
		return fmt.Errorf("PNG bit depth is %d, want 8", data[24])
	}
	if data[25] != 2 {
		return fmt.Errorf("PNG color type is %d, want 2 (opaque truecolor RGB)", data[25])
	}
	if data[26] != 0 || data[27] != 0 || data[28] > 1 {
		return fmt.Errorf("invalid PNG compression/filter/interlace methods %d/%d/%d", data[26], data[27], data[28])
	}
	if _, err := png.Decode(bytes.NewReader(data)); err != nil {
		return fmt.Errorf("decode PNG: %w", err)
	}
	return nil
}

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
