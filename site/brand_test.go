package site

import (
	"bytes"
	"testing"
)

func TestHeaderIconUsesCroppedArtboard(t *testing.T) {
	icon, err := BrandAsset("araihu-icon-transparent.svg")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(icon, []byte(`viewBox="280 289 464 464"`)) {
		t.Fatal("header icon artboard includes excessive whitespace")
	}
}
