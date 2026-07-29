package site

import (
	"bytes"
	"io/fs"
	"testing"
)

func TestHeaderIconUsesCroppedArtboard(t *testing.T) {
	icon, err := fs.ReadFile(BrandAssets(), "icons/brand/araihu-icon-light-transparent-optical.svg")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(icon, []byte(`viewBox="262.510 263.185 498.904 498.904"`)) {
		t.Fatal("header icon artboard includes excessive whitespace")
	}
}
