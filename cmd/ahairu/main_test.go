package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/araihu/ahairu/site"
)

func TestBuildWritesStandaloneSite(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("PWD", t.TempDir())
	if err := os.Chdir(os.Getenv("PWD")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })

	if err := build(); err != nil {
		t.Fatal(err)
	}
	html, err := os.ReadFile(filepath.Join("public", "en", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	page := string(html)
	if !strings.Contains(page, "Software for stormy weather.") || !strings.Contains(page, "/assets/styles.css") {
		t.Fatalf("generated HTML misses site content or stylesheet: %s", html)
	}
	if !strings.Contains(page, "Independent, open tools built to endure difficult work.") {
		t.Error("generated English page misses the supporting storm promise")
	}
	for _, removedFact := range []string{"Four maintained projects", "Built in Go"} {
		if strings.Contains(page, removedFact) {
			t.Errorf("generated English hero still contains removed fact %q", removedFact)
		}
	}
	for _, landmark := range []string{
		`class="skip-link" href="#main-content"`,
		`<header class="ahairu-header">`,
		`<meta name="theme-color" content="#07111f">`,
		`<title>Arai Hû — Software for stormy weather.</title>`,
		`<meta name="description" content="Independent, open software built to endure difficult work.">`,
		`<link rel="canonical" href="https://araihu.com/en/">`,
		`<meta property="og:type" content="website">`,
		`<meta property="og:image" content="https://araihu.com/assets/social/araihu-storm-v1.jpg">`,
		`<meta property="og:image:width" content="1280">`,
		`<meta property="og:image:height" content="640">`,
		`<meta property="og:image:alt" content="Arai Hû mark over a dark storm at sea.">`,
		`<meta name="twitter:card" content="summary_large_image">`,
		`<meta name="twitter:image:alt" content="Arai Hû mark over a dark storm at sea.">`,
		`class="ahairu-brand-copy"><strong>Arai Hû</strong>`,
		`<main id="main-content" tabindex="-1">`,
		`href="/en/" aria-current="page"`,
		`aria-label="Primary navigation"`,
		`<section id="libs" class="featured-project"`,
		`<video data-featured-montage loop muted playsinline preload="metadata" poster="/assets/visuals/goshtoso-components-poster-v1.webp">`,
		`<source src="/assets/visuals/goshtoso-components-montage-v1.mp4" type="video/mp4">`,
		`Tabs · HTMX · Monitoring · Line 3D`,
		`class="project-grid mt-8"`,
		`aria-describedby="X-9-desc">X-9</h3>`,
		`<script defer src="/assets/js/x9-availability.js?rev=1"></script>`,
		`data-chart-bundle-trigger`,
		`hx-get="/fragments/en/charts.html"`,
		`hx-trigger="revealed once"`,
		`id="paje-chart-slot"`,
		`id="x9-chart-slot"`,
		`id="goshtoso-heart-chart-slot"`,
		`data-chart-placeholder="paje"`,
		`data-chart-placeholder="x9"`,
		`data-chart-placeholder="goshtoso-heart"`,
		`id="goshtoso-version-slot"`,
		`hx-get="/api/project-versions"`,
		`hx-trigger="load delay:750ms"`,
		`hx-swap="outerHTML"`,
		`id="goshtoso-charts-version-slot"`,
		`data-shell-preview`,
		`class="openapi-stream"`,
		`<b>openapi:</b> 3.1.0`,
		`github-rest-api.yaml`,
		`<b>operationId:</b> repos/list-releases`,
		`<b>operationId:</b> repos/create-release`,
		`<b>generate_release_notes:</b>`,
		`class="more-art more-art--muamba"`,
		`class="muamba-drop" src="/assets/logos/muamba-mark.svg"`,
		`class="landing-shell__mobile-navigation storm-mobile-menu"`,
		`class="landing-shell__mobile-trigger is-bottom-left storm-mobile-trigger"`,
		`x-bind:aria-expanded="navigationOpen"`,
		`<details class="landing-shell__mobile-fallback">`,
		`class="landing-shell__mobile-fallback-panel storm-mobile-panel"`,
		`motion-reduce:transition-none`,
		`<script defer src="/assets/js/storm-backdrop.js?rev=2"></script>`,
		`<video class="storm-backdrop" data-storm-backdrop loop muted playsinline preload="none">`,
		`<source media="(prefers-color-scheme: light)" src="/assets/video/storm-light-v1.mp4" type="video/mp4">`,
		`<source media="(prefers-color-scheme: dark)" src="/assets/video/storm-dark-v1.mp4" type="video/mp4">`,
		`<div class="storm-video-filter"></div>`,
		`src="/assets/js/runtime/alpinejs-focus/3.14.9/alpine-focus.min.js"`,
		`href="#libs">Libs</a>`,
		`href="#apps">Apps</a>`,
		`href="#field-notes">Blog`,
		`<input id="mailing-email" type="email" placeholder="you@example.com" disabled>`,
		`<button type="button" disabled>Coming soon</button>`,
		`href="https://github.com/araihu">Visit Arai Hû on GitHub`,
		`href="https://go.dev/">Go</a>`,
		`href="https://goshtoso.araihu.com">Goshtoso</a>`,
	} {
		if !strings.Contains(page, landmark) {
			t.Errorf("generated HTML misses accessibility landmark %q", landmark)
		}
	}
	for _, deferredMarkup := range []string{
		`data-echarts-src="/charts/assets/js/runtime/echarts/5.4.3/echarts.min.js"`,
		`data-three-d-src="/charts/assets/js/runtime/three-d/2.0.9/runtime.min.js"`,
		`src="/assets/js/chart-loader.js?rev=1"`,
		`data-paje-actual-chart`,
		`data-x9-live-availability`,
		`data-goshtoso-heart-chart`,
		`_echarts_instance_`,
	} {
		if strings.Contains(page, deferredMarkup) {
			t.Errorf("initial HTML eagerly includes deferred chart payload %q", deferredMarkup)
		}
	}
	for status, expected := range map[string]int{"BETA": 2, "ALPHA": 2, "WIP": 3} {
		marker := `data-status="` + status + `">` + status + `</span>`
		if count := strings.Count(page, marker); count != expected {
			t.Errorf("generated English page has %d %s labels, want %d", count, status, expected)
		}
	}
	if count := strings.Count(page, `hx-get="/api/project-versions"`); count != 1 {
		t.Errorf("generated English page requests project versions %d times, want once", count)
	}
	chartFragment, err := os.ReadFile(filepath.Join("public", "fragments", "en", "charts.html"))
	if err != nil {
		t.Fatalf("read generated English chart fragment: %v", err)
	}
	fragment := string(chartFragment)
	for _, deferredMarkup := range []string{
		`data-echarts-src="/charts/assets/js/runtime/echarts/5.4.3/echarts.min.js"`,
		`data-three-d-src="/charts/assets/js/runtime/three-d/2.0.9/runtime.min.js"`,
		`src="/assets/js/chart-loader.js?rev=1"`,
		`id="paje-chart-slot"`,
		`data-paje-actual-chart`,
		`aria-label="Pajé — Durable workflows for code changes."`,
		`id="x9-chart-slot"`,
		`data-x9-live-availability`,
		`aria-label="X-9 — Self-hosted monitoring control plane."`,
		`id="goshtoso-heart-chart-slot"`,
		`data-goshtoso-heart-chart`,
		`aria-label="Goshtoso Charts — Three-dimensional parametric heart line."`,
	} {
		if !strings.Contains(fragment, deferredMarkup) {
			t.Errorf("chart fragment misses deferred payload %q", deferredMarkup)
		}
	}
	lastSection := -1
	for _, marker := range []string{`id="home"`, `id="libs"`, `id="apps"`, `class="more-projects"`, `class="open-mission"`, `class="keep-up"`} {
		position := strings.Index(page, marker)
		if position < 0 || position <= lastSection {
			t.Errorf("generated section order breaks at %q", marker)
		}
		lastSection = position
	}
	for _, asset := range []string{
		"logos/araihu-icon-background.svg?rev=a8a9647a",
		"logos/araihu-icon-transparent.svg?rev=a8a9647a",
		"logos/manja-icon-transparent.svg?rev=a8a9647a",
	} {
		if !strings.Contains(page, asset) {
			t.Errorf("generated HTML misses brand asset %q", asset)
		}
	}
	if !strings.Contains(page, `href="https://x9.araihu.com"`) {
		t.Error("generated HTML misses the X-9 product URL")
	}
	if strings.Contains(page, "xisnove.dev") {
		t.Error("generated HTML still links to the retired Xisnove domain")
	}
	if strings.Contains(page, `aria-label="Explore project: Goshtoso"`) {
		t.Error("generated page repeats featured Goshtoso inside the project grid")
	}
	if strings.Contains(page, `paje-graph-v1.webp`) {
		t.Error("generated page still uses the Pajé raster instead of the actual chart")
	}
	if strings.Contains(page, `x9-monitor-chart-v1.svg`) {
		t.Error("generated page still uses the X-9 raster instead of the live availability chart")
	}
	if strings.Contains(page, `goshtoso-heart-line-v1.svg`) {
		t.Error("generated page still uses the projected heart SVG instead of the actual Line3D chart")
	}
	if info, err := os.Stat(filepath.Join("public", "assets", "styles.css")); err != nil || info.Size() == 0 {
		t.Fatalf("generated stylesheet missing or empty: %v", err)
	}
	if info, err := os.Stat(filepath.Join("public", "assets", "ahairu.css")); err != nil || info.Size() == 0 {
		t.Fatalf("brand stylesheet missing or empty: %v", err)
	}
	brandCSS, err := os.ReadFile(filepath.Join("public", "assets", "ahairu.css"))
	if err != nil {
		t.Fatalf("read brand stylesheet: %v", err)
	}
	for _, token := range []string{
		"--storm-night:",
		"--storm-cloud:",
		"--storm-mist:",
		"--storm-signal:",
		"@media (prefers-reduced-motion: reduce)",
	} {
		if !strings.Contains(string(brandCSS), token) {
			t.Errorf("brand stylesheet misses storm-system token %q", token)
		}
	}
	reducedMotionStart := strings.LastIndex(string(brandCSS), "@media (prefers-reduced-motion: reduce)")
	if reducedMotionStart < 0 {
		t.Fatal("brand stylesheet misses reduced-motion block")
	}
	reducedMotionCSS := string(brandCSS)[reducedMotionStart:]
	for _, contract := range []string{
		".storm-mobile-trigger, .storm-mobile-panel, .more-row",
		".signal-button, .project-art::before, .project-art::after, .project-art-name, .project-mark, .project-title-mark, .more-art",
		".openapi-stream { animation: none; transform: none; }",
		".muamba-drop { animation: none;",
		"transition: none",
		"transform: none !important",
	} {
		if !strings.Contains(reducedMotionCSS, contract) {
			t.Errorf("reduced-motion contract misses %q", contract)
		}
	}
	if info, err := os.Stat(filepath.Join("public", "assets", "araihu-theme.css")); err != nil || info.Size() == 0 {
		t.Fatalf("brand theme missing or empty: %v", err)
	}
	for _, generatedAsset := range []string{
		filepath.Join("public", "landingshell", "assets", "shell.css"),
		filepath.Join("public", "charts", "assets", "js", "runtime", "echarts", "5.4.3", "echarts.min.js"),
		filepath.Join("public", "charts", "assets", "js", "runtime", "three-d", "2.0.9", "runtime.min.js"),
		filepath.Join("public", "assets", "js", "storm-backdrop.js"),
		filepath.Join("public", "assets", "js", "x9-availability.js"),
		filepath.Join("public", "assets", "js", "chart-loader.js"),
		filepath.Join("public", "assets", "js", "goshtoso.min.js"),
		filepath.Join("public", "assets", "js", "runtime", "alpinejs-focus", "3.14.9", "alpine-focus.min.js"),
		filepath.Join("public", "assets", "js", "runtime", "alpinejs", "3.14.9", "alpine.min.js"),
		filepath.Join("public", "assets", "social", "araihu-storm-v1.jpg"),
	} {
		if info, err := os.Stat(generatedAsset); err != nil || info.Size() == 0 {
			t.Errorf("component runtime asset %s missing or empty: %v", generatedAsset, err)
		}
	}
	for _, backdrop := range []string{"storm-dark-v1.mp4", "storm-light-v1.mp4"} {
		info, err := os.Stat(filepath.Join("public", "assets", "video", backdrop))
		if err != nil || info.Size() == 0 || info.Size() > 500*1024 {
			t.Errorf("optimized backdrop %s missing or outside 500 KiB budget: size=%d err=%v", backdrop, func() int64 {
				if info == nil {
					return 0
				}
				return info.Size()
			}(), err)
		}
	}
	for _, visual := range site.ProjectVisualAssetNames() {
		info, err := os.Stat(filepath.Join("public", "assets", "visuals", visual))
		if err != nil || info.Size() == 0 || info.Size() > 350*1024 {
			t.Errorf("project visual %s missing or outside 350 KiB budget: size=%d err=%v", visual, func() int64 {
				if info == nil {
					return 0
				}
				return info.Size()
			}(), err)
		}
	}
	for _, name := range site.BrandAssetNames() {
		if info, err := os.Stat(filepath.Join("public", "assets", "logos", name)); err != nil || info.Size() == 0 {
			t.Errorf("brand asset %s missing or empty: %v", name, err)
		}
	}
	for _, localizedContent := range site.Locales() {
		for _, project := range append(append([]site.Project{}, localizedContent.Projects...), localizedContent.MoreProjects...) {
			if project.URL == "" || project.Category == "" || project.Description == "" {
				t.Errorf("localized project %q in %s misses URL, category, or description", project.Name, localizedContent.Language)
			}
		}
	}
	localizedExpectations := map[string]struct {
		skipLabel    string
		tagline      string
		promise      string
		currentHref  string
		removedFacts []string
	}{
		"pt-br": {
			skipLabel:    "Pular para o conteúdo",
			tagline:      "Software para passar a trovoada.",
			promise:      "Ferramentas independentes e abertas, feitas para resistir ao trabalho difícil.",
			currentHref:  `href="/pt-br/" aria-current="page"`,
			removedFacts: []string{"Quatro projetos mantidos", "Criados em Go"},
		},
		"es": {
			skipLabel:    "Saltar al contenido",
			tagline:      "Software para tiempos de tormenta.",
			promise:      "Herramientas independientes y abiertas, creadas para resistir el trabajo difícil.",
			currentHref:  `href="/es/" aria-current="page"`,
			removedFacts: []string{"Cuatro proyectos mantenidos", "Creados en Go"},
		},
	}
	for locale, expectation := range localizedExpectations {
		localizedHTML, err := os.ReadFile(filepath.Join("public", locale, "index.html"))
		if err != nil {
			t.Fatalf("localized page missing for %s: %v", locale, err)
		}
		localizedPage := string(localizedHTML)
		if !strings.Contains(localizedPage, expectation.skipLabel) || !strings.Contains(localizedPage, `aria-current="page"`) {
			t.Errorf("localized page %s misses skip link or locale state", locale)
		}
		if !strings.Contains(localizedPage, expectation.tagline) || !strings.Contains(localizedPage, expectation.promise) {
			t.Errorf("localized page %s misses storm tagline or supporting promise", locale)
		}
		for _, removedFact := range expectation.removedFacts {
			if strings.Contains(localizedPage, removedFact) {
				t.Errorf("localized page %s still contains removed hero fact %q", locale, removedFact)
			}
		}
		if got := strings.Count(localizedPage, expectation.currentHref); got != 3 {
			t.Errorf("localized page %s current-locale count = %d; want desktop, enhanced drawer, and native fallback states", locale, got)
		}
	}
	if got := strings.Count(page, `href="/en/" aria-current="page"`); got != 3 {
		t.Errorf("English current-locale count = %d; want desktop, enhanced drawer, and native fallback states", got)
	}
}
