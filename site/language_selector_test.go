package site

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
)

func TestLanguageMenuPreservesEveryPageFamilyAndNamesCurrentLanguage(t *testing.T) {
	for _, page := range Pages() {
		menu := languageMenu(page)
		if len(menu.Sections) != 1 || len(menu.Sections[0].Items) != 3 {
			t.Fatalf("%s language menu is incomplete", page.Meta.Path)
		}
		for i, item := range menu.Sections[0].Items {
			want := page.Navigation.Locales[i]
			if item.Href != want.URL || item.Label != want.Locale.Name || item.Icon == nil {
				t.Errorf("%s menu changes page family or loses language identity: %#v", page.Meta.Path, item)
			}
		}
		html := renderComponent(t, LanguageSelector(page))
		if !strings.Contains(html, `aria-label="`+sharedContent(page.Meta.Locale.Key).LanguageLabel+": "+page.Meta.Locale.Name+`"`) || !strings.Contains(html, "<noscript>") {
			t.Errorf("%s language trigger lacks localized identity or native fallback", page.Meta.Path)
		}
	}
}

func TestFlagAssetsMatchLicensedProvenance(t *testing.T) {
	data, err := ProjectVisualAsset("flag-icons-provenance.json")
	if err != nil {
		t.Fatal(err)
	}
	var provenance struct {
		License string            `json:"license"`
		Files   map[string]string `json:"files"`
	}
	if err := json.Unmarshal(data, &provenance); err != nil {
		t.Fatal(err)
	}
	if provenance.License != "MIT" || len(provenance.Files) != 4 {
		t.Fatal("flag license or inventory is incomplete")
	}
	for name, want := range provenance.Files {
		data, err := ProjectVisualAsset(name)
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(data)
		if hex.EncodeToString(digest[:]) != want {
			t.Errorf("flag asset %s differs from licensed source", name)
		}
	}
}
