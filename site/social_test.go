package site

import (
	"bytes"
	"image"
	_ "image/png"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

func TestBundledSocialImagesAreExactOpenGraphPNGs(t *testing.T) {
	for _, name := range []string{"brand.png", "license.png"} {
		data, err := fs.ReadFile(embeddedSocialImages, "social/"+name)
		if err != nil {
			t.Fatal(err)
		}
		config, format, err := image.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		if format != "png" || config.Width != 1200 || config.Height != 630 {
			t.Errorf("%s = %s %dx%d, want png 1200x630", name, format, config.Width, config.Height)
		}
	}
}

func TestBundledSocialImagesFillTheBottomEdge(t *testing.T) {
	tests := []struct {
		name  string
		x     int
		check func(r, g, b uint32) bool
	}{
		{name: "brand.png", x: 1199, check: func(r, g, b uint32) bool { return b > g && g > r && b > 0x7000 }},
		{name: "license.png", x: 0, check: func(r, g, b uint32) bool { return r < 0x3000 && g < 0x3000 && b < 0x4000 }},
	}
	for _, test := range tests {
		data, err := fs.ReadFile(embeddedSocialImages, "social/"+test.name)
		if err != nil {
			t.Fatal(err)
		}
		preview, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		r, g, b, _ := preview.At(test.x, 629).RGBA()
		if !test.check(r, g, b) {
			t.Errorf("%s bottom edge pixel = #%04x%04x%04x, artwork does not fill row 629", test.name, r, g, b)
		}
	}
}

func TestCopyBundledSocialImagesWritesExactEmbeddedBytes(t *testing.T) {
	destination := t.TempDir()
	if err := CopyBundledSocialImages(destination); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"brand.png", "license.png"} {
		want, err := fs.ReadFile(embeddedSocialImages, "social/"+name)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(destination, name)
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("%s copied bytes differ from embedded source", name)
		}

		if err := os.WriteFile(path, []byte("stale"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := CopyBundledSocialImages(destination); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"brand.png", "license.png"} {
		want, _ := fs.ReadFile(embeddedSocialImages, "social/"+name)
		got, _ := os.ReadFile(filepath.Join(destination, name))
		if !bytes.Equal(got, want) {
			t.Errorf("%s is not deterministically replaced", name)
		}
	}
}
