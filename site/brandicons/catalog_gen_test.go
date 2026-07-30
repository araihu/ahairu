package brandicons

import "testing"

func TestGeneratedAraihuBrandCatalog(t *testing.T) {
	if SpriteURL != "/assets/araihu/v0.1.0/icons/brand/sprite.svg" {
		t.Fatalf("SpriteURL = %q", SpriteURL)
	}
	if got, want := len(Glyphs), 19; got != want {
		t.Fatalf("generated glyph count = %d, want %d", got, want)
	}
	seenNames := make(map[string]struct{}, len(Glyphs))
	seenSymbols := make(map[string]struct{}, len(Glyphs))
	for _, glyph := range Glyphs {
		if glyph.GoName == "" || glyph.CanonicalName == "" || glyph.Symbol == "" {
			t.Fatalf("generated empty glyph field: %#v", glyph)
		}
		if _, exists := seenNames[glyph.CanonicalName]; exists {
			t.Fatalf("duplicate canonical name %q", glyph.CanonicalName)
		}
		if _, exists := seenSymbols[string(glyph.Symbol)]; exists {
			t.Fatalf("duplicate sprite symbol %q", glyph.Symbol)
		}
		seenNames[glyph.CanonicalName] = struct{}{}
		seenSymbols[string(glyph.Symbol)] = struct{}{}
	}
}
