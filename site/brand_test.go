package site

import (
	"bytes"
	"strings"
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

func TestBackdropAssetsStayWithinTransferBudget(t *testing.T) {
	for _, name := range BackdropAssetNames() {
		asset, err := BackdropAsset(name)
		if err != nil {
			t.Fatal(err)
		}
		if len(asset) == 0 || len(asset) > 500*1024 {
			t.Errorf("backdrop %s size = %d bytes; want 1..512000", name, len(asset))
		}
	}
}

func TestProjectVisualAssetsStayWithinTransferBudget(t *testing.T) {
	for _, name := range ProjectVisualAssetNames() {
		asset, err := ProjectVisualAsset(name)
		if err != nil {
			t.Fatal(err)
		}
		if len(asset) == 0 || len(asset) > 350*1024 {
			t.Errorf("project visual %s size = %d bytes; want 1..358400", name, len(asset))
		}
	}
}

func TestGitHubReleaseOpenAPISpecIsSubstantialAndReal(t *testing.T) {
	lines := GitHubReleaseOpenAPILines()
	if len(lines) < 140 {
		t.Fatalf("GitHub release OpenAPI excerpt has %d lines; want at least 140", len(lines))
	}
	var source strings.Builder
	for _, line := range lines {
		source.WriteString(line.Indent + line.Key + line.Value + "\n")
	}
	for _, contract := range []string{
		"title: GitHub v3 REST API",
		`"/repos/{owner}/{repo}/releases":`,
		"operationId: repos/list-releases",
		"operationId: repos/create-release",
		`"$ref": "#/components/schemas/release"`,
	} {
		if !strings.Contains(source.String(), contract) {
			t.Errorf("GitHub release OpenAPI excerpt misses %q", contract)
		}
	}
}

func TestBackdropEnhancementProtectsMotionAndDataPreferences(t *testing.T) {
	script := StormBackdropJS()
	for _, contract := range [][]byte{
		[]byte("prefers-reduced-motion: reduce"),
		[]byte("navigator.connection?.saveData"),
		[]byte("requestAnimationFrame"),
		[]byte("video.pause()"),
		[]byte("IntersectionObserver"),
		[]byte("montage.pause()"),
	} {
		if !bytes.Contains(script, contract) {
			t.Errorf("storm backdrop script misses %q", contract)
		}
	}
}
