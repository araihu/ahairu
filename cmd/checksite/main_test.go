package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/ahairu/site"
)

func TestWorkflowsPinToolchainsAndActions(t *testing.T) {
	ci := readWorkflow(t, "ci.yml")
	for _, want := range []string{
		"go-version: 1.26.5",
		"node-version: 24",
	} {
		if !strings.Contains(ci, want) {
			t.Errorf("CI misses %q", want)
		}
	}
	assertPinnedActions(t, "ci.yml", ci)

	acceptedAssets := readWorkflow(t, "accepted-assets.yml")
	for _, want := range []string{
		"repository_dispatch:",
		"go-version: 1.26.5",
		"node-version: 24",
		"actions/create-github-app-token@",
		"actions/upload-artifact@",
	} {
		if !strings.Contains(acceptedAssets, want) {
			t.Errorf("accepted-assets misses %q", want)
		}
	}
	assertPinnedActions(t, "accepted-assets.yml", acceptedAssets)

	deploy := readWorkflow(t, "deploy.yml")
	for _, want := range []string{
		"group: ahairu-production",
		"cancel-in-progress: false",
		"environment: production",
		"actions/create-github-app-token@",
		"actions/download-artifact@",
	} {
		if !strings.Contains(deploy, want) {
			t.Errorf("deploy misses %q", want)
		}
	}
	assertPinnedActions(t, "deploy.yml", deploy)
}

func TestWorkflowPromotionAndDeploymentSecurityContracts(t *testing.T) {
	acceptedAssets := readWorkflow(t, "accepted-assets.yml")
	for _, want := range []string{
		"if: github.event_name == 'repository_dispatch'",
		"if: github.event_name == 'push' && github.ref == 'refs/heads/main'",
		"permission-actions: read",
		"- araihu-assets-released",
		"DISPATCH_EVENT_TYPE: ${{ github.event.action }}",
		"ASSETS_HANDOFF_JSON: ${{ toJSON(github.event.client_payload) }}",
		"ASSETS_HANDOFF_JSON: ${{ vars.ASSETS_RELEASE_HANDOFF_JSON }}",
		"--handoff-json \"$ASSETS_HANDOFF_JSON\"",
		"--accepted-output \"$ACCEPTED_ASSETS\"",
		"npm run test:workflow",
	} {
		if !strings.Contains(acceptedAssets, want) {
			t.Errorf("accepted-assets misses %q", want)
		}
	}
	if strings.Contains(acceptedAssets, "assets-release-promoted") || strings.Contains(acceptedAssets, "assets_release_") || strings.Contains(acceptedAssets, "assets_channel_") {
		t.Error("accepted-assets retains the obsolete flat Assets dispatch schema")
	}
	if strings.Contains(acceptedAssets, "client_payload.assets_") || strings.Contains(acceptedAssets, "|| vars") {
		t.Error("accepted-assets allows repository variables to fill a dispatch payload")
	}
	if got := strings.Count(acceptedAssets, "github.event.client_payload"); got != 1 {
		t.Errorf("accepted-assets has %d dispatch payload references, want exactly one full handoff", got)
	}
	if got := strings.Count(acceptedAssets, "vars.ASSETS_"); got != 2 {
		t.Errorf("accepted-assets has %d main-promotion variable references, want the handoff and its non-empty guard", got)
	}
	if got := strings.Count(acceptedAssets, "permission-actions: read"); got != 1 || strings.Count(acceptedAssets, "permission-contents: read") != 1 {
		t.Error("accepted-assets token does not request least Actions and Contents read")
	}

	deploy := readWorkflow(t, "deploy.yml")
	for _, want := range []string{
		"github.event.workflow_run.event == 'repository_dispatch'",
		"vars.ASSETS_RELEASE_HANDOFF_JSON != ''",
		"permission-actions: read",
		"WRANGLER_OUTPUT_FILE_PATH=\"$RUNNER_TEMP/wrangler-version-upload.jsonl\"",
		"wrangler versions upload",
		"wrangler versions deploy --version-id \"$UPLOADED_VERSION\" --percentage 100 --yes",
		"wrangler deployments status --json",
		"--upload \"$RUNNER_TEMP/wrangler-version-upload.jsonl\"",
		"--uploaded-version \"$UPLOADED_VERSION\"",
		"scripts/select_deployed_version.mjs",
		"set -o pipefail",
		"permission-contents: read",
		"permission-contents: write",
		"scripts/accepted_assets_state.py",
		"AHAIRU_REPOSITORY: araihu/ahairu",
		"STATE_REF: automation/araihu-assets-state",
		"STATE_PATH: .automation/araihu-assets/accepted-channel-v1.json",
		"wrangler deployments status --json",
		"Create dedicated accepted-state ref failed",
		"Create accepted state conflicted or failed",
		"Update accepted state conflicted or failed",
		"if [[ -z \"$state_sha\" ]]; then",
		"test \"$update_status\" = 201",
		"test \"$update_status\" = 200",
	} {
		if !strings.Contains(deploy, want) {
			t.Errorf("deploy misses %q", want)
		}
	}
	for _, forbidden := range []string{"wrangler deploy ", "wrangler versions list", "worker-versions-before.json", "worker-versions-after.json"} {
		if strings.Contains(deploy, forbidden) {
			t.Errorf("deploy retains ambiguous Worker version inference %q", forbidden)
		}
	}
	if got := strings.Count(deploy, "permission-actions: read"); got != 1 || strings.Count(deploy, "permission-contents: read") != 1 || strings.Count(deploy, "permission-contents: write") != 1 {
		t.Error("deploy tokens do not use least Assets read and post-promotion Ahairu write permissions")
	}
	if strings.Index(deploy, "wrangler deployments status --json") > strings.Index(deploy, "Mint post-promotion state token") || strings.Index(deploy, "Mint post-promotion state token") > strings.Index(deploy, "Durably accept verified deployed Assets channel") {
		t.Error("accepted state must be written only after deployed Worker version is verified active")
	}
	if strings.Contains(deploy, "refs/heads/main") || strings.Contains(deploy, "contents/.github/") {
		t.Error("deploy state write must never target main or a mutable workflow path")
	}
	createBranch := strings.Index(deploy, "if [[ -z \"$state_sha\" ]]; then")
	createStatus := strings.Index(deploy, "test \"$update_status\" = 201")
	updateBranch := strings.Index(deploy, "else\n            test \"$update_status\" = 200")
	if createBranch < 0 || createStatus < createBranch || updateBranch < createStatus {
		t.Error("accepted-state Contents PUT must require 201 for creates and 200 for updates")
	}
}

func TestDeployWorkflowAssemblesVerifiedAssetBundleDirectly(t *testing.T) {
	acceptedAssets := readWorkflow(t, "accepted-assets.yml")
	for _, want := range []string{
		"araihu-assets-released",
		"toJSON(github.event.client_payload)",
		"ASSETS_RELEASE_HANDOFF_JSON",
		"scripts/prepare_asset_bundle.py",
		"--handoff-json",
		"--accepted-output",
		"npm run check",
		"name: accepted-assets",
	} {
		if !strings.Contains(acceptedAssets, want) {
			t.Errorf("accepted-assets misses %q", want)
		}
	}
	deploy := readWorkflow(t, "deploy.yml")
	for _, want := range []string{
		"run-id: ${{ github.event.workflow_run.id }}",
		"name: accepted-assets",
		"scripts/prepare_asset_bundle.py",
		"--handoff-file \"$ACCEPTED_ASSETS\"",
		"npm run check",
		"wrangler versions upload",
		"wrangler versions deploy",
		"CLOUDFLARE_API_TOKEN: ${{ secrets.CLOUDFLARE_API_TOKEN }}",
	} {
		if !strings.Contains(deploy, want) {
			t.Errorf("deploy misses %q", want)
		}
	}
	for _, workflow := range []struct {
		name, contents string
	}{{"accepted-assets", acceptedAssets}, {"deploy", deploy}} {
		for _, forbidden := range []string{"CLOUDFLARE_DEPLOY_HOOK", "ASSETS_CLOUDFLARE"} {
			if strings.Contains(workflow.contents, forbidden) {
				t.Errorf("%s retains forbidden secret %q", workflow.name, forbidden)
			}
		}
	}
	if strings.Contains(deploy, "accepted.releaseURL") || strings.Contains(deploy, "accepted.channelURL") {
		t.Error("deploy retains the obsolete flat accepted-assets schema")
	}
}

func readWorkflow(t *testing.T, name string) string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

func assertPinnedActions(t *testing.T, name, workflow string) {
	t.Helper()
	for _, line := range strings.Split(workflow, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "- uses: ") {
			continue
		}
		parts := strings.SplitN(strings.TrimPrefix(line, "- uses: "), "@", 2)
		if len(parts) != 2 || len(parts[1]) != 40 {
			t.Errorf("%s action is not pinned to a full SHA: %q", name, line)
			continue
		}
		for _, character := range parts[1] {
			if !(character >= '0' && character <= '9' || character >= 'a' && character <= 'f') {
				t.Errorf("%s action is not pinned to a lowercase SHA: %q", name, line)
				break
			}
		}
	}
}

func TestCheckAcceptsCompleteStaticSite(t *testing.T) {
	root := writeValidPublic(t)
	if err := Check(root); err != nil {
		t.Fatal(err)
	}
}

func TestCheckRejectsCampaignCanaryContractViolations(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		mutate func(string) string
		want   string
	}{
		{
			name: "theme source",
			mutate: func(document string) string {
				return strings.Replace(document, ` data-theme-source="default"`, "", 1)
			},
			want: "theme source",
		},
		{
			name: "runtime channel",
			mutate: func(document string) string {
				return strings.Replace(document, `data-channel="/assets/releases/current"`, `data-channel="/assets/releases/latest"`, 1)
			},
			want: "campaign runtime",
		},
		{
			name: "image dimensions",
			mutate: func(document string) string {
				return strings.Replace(document, `width="64" height="64"`, `width="64"`, 1)
			},
			want: "image dimensions",
		},
		{
			name: "icon hook target",
			mutate: func(document string) string {
				document = strings.Replace(document, `data-asset-brand="icon"`, "", 1)
				return strings.Replace(document, `width="64" height="64"`, `width="64" height="64" data-asset-brand="icon"`, 1)
			},
			want: "icon hook must target link[href]",
		},
		{
			name: "logo hook target",
			path: "/brand/",
			mutate: func(document string) string {
				document = strings.Replace(document, `data-asset-brand="logo"`, "", 1)
				return strings.Replace(document, `rel="manifest"`, `rel="manifest" data-asset-brand="logo"`, 1)
			},
			want: "logo hook must target img[src]",
		},
		{
			name: "exact icon hook count",
			mutate: func(document string) string {
				return strings.Replace(document, `rel="manifest"`, `rel="icon" data-asset-brand="icon"`, 1)
			},
			want: "campaign brand hooks",
		},
		{
			name: "exact logo hook count",
			mutate: func(document string) string {
				return strings.Replace(document, `width="64" height="64"`, `width="64" height="64" data-asset-brand="logo"`, 1)
			},
			want: "campaign brand hooks",
		},
		{
			name: "campaign toggle count",
			mutate: func(document string) string {
				const toggle = `<button class="ahairu-campaign-toggle" type="button" hidden data-campaign-toggle aria-pressed="false"><span data-campaign-toggle-icon aria-hidden="true"></span> <span class="sr-only">Use the standard Arai Hû appearance</span></button>`
				return strings.Replace(document, toggle, toggle+toggle, 1)
			},
			want: "campaign toggle count",
		},
		{
			name: "campaign toggle hidden",
			mutate: func(document string) string {
				return strings.Replace(document, ` hidden data-campaign-toggle`, ` data-campaign-toggle`, 1)
			},
			want: "campaign toggle must have hidden attribute",
		},
		{
			name: "campaign toggle element",
			mutate: func(document string) string {
				document = strings.Replace(document, `<button class="ahairu-campaign-toggle"`, `<div class="ahairu-campaign-toggle"`, 1)
				return strings.Replace(document, `</button>`, `</div>`, 1)
			},
			want: "campaign toggle must be button",
		},
		{
			name: "campaign toggle type",
			mutate: func(document string) string {
				return strings.Replace(document, `type="button" hidden data-campaign-toggle`, `type="submit" hidden data-campaign-toggle`, 1)
			},
			want: "campaign toggle type must be button",
		},
		{
			name: "campaign toggle pressed state",
			mutate: func(document string) string {
				return strings.Replace(document, `data-campaign-toggle aria-pressed="false"`, `data-campaign-toggle aria-pressed="true"`, 1)
			},
			want: "campaign toggle aria-pressed must be false",
		},
		{
			name: "campaign toggle icon child",
			mutate: func(document string) string {
				return strings.Replace(document, `<span data-campaign-toggle-icon aria-hidden="true"></span>`, ``, 1)
			},
			want: "campaign toggle icon children",
		},
		{
			name: "campaign toggle exact icon child count",
			mutate: func(document string) string {
				const icon = `<span data-campaign-toggle-icon aria-hidden="true"></span>`
				return strings.Replace(document, icon, icon+icon, 1)
			},
			want: "campaign toggle icon children",
		},
		{
			name: "campaign toggle accessible name",
			mutate: func(document string) string {
				return strings.Replace(document, `>Use the standard Arai Hû appearance</span>`, `></span>`, 1)
			},
			want: "campaign toggle must have an accessible name",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeValidPublic(t)
			path := test.path
			if path == "" {
				path = "/en/"
			}
			mutatePage(t, root, path, test.mutate)
			if err := Check(root); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Check() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestCheckAcceptsNormalHTMLWithOptionalEndTags(t *testing.T) {
	root := writeValidPublic(t)
	mutatePage(t, root, "/en/", func(document string) string {
		return strings.Replace(document, "</body>", `<ul><li>one<li>two</ul><p>first<p>second<table><tr><td>left<td>right</tr></table><a href="https://example.com">external</a><a href="#local">fragment</a><a href="mailto:hello@example.com">mail</a><a href="tel:+15551212">phone</a>`+"</body>", 1)
	})
	if err := Check(root); err != nil {
		t.Fatal(err)
	}
}

func TestRejectDuplicateHTMLAttributesUsesRawTokens(t *testing.T) {
	for _, source := range []string{
		`<html lang="en" LANG="es"></html>`,
		`<input disabled disabled>`,
		`<img alt=first alt=second/>`,
	} {
		if err := rejectDuplicateHTMLAttributes([]byte(source)); err == nil || !strings.Contains(err.Error(), "duplicate") {
			t.Errorf("rejectDuplicateHTMLAttributes(%q) error = %v, want duplicate", source, err)
		}
	}
	for _, source := range []string{
		`<div title="lang=en lang=es" data-value='x=y'></div>`,
		`<img src=/assets/icons/a/b.svg alt=icon>`,
	} {
		if err := rejectDuplicateHTMLAttributes([]byte(source)); err != nil {
			t.Errorf("rejectDuplicateHTMLAttributes(%q) error = %v", source, err)
		}
	}
}

func TestCheckRejectsMissingOrWrongSizedSocialImages(t *testing.T) {
	root := writeValidPublic(t)
	if err := os.Remove(filepath.Join(root, "social", "brand.png")); err != nil {
		t.Fatal(err)
	}
	if err := Check(root); err == nil || !strings.Contains(err.Error(), "social image") {
		t.Fatalf("Check missing social image error = %v", err)
	}

	root = writeValidPublic(t)
	writePNG(t, filepath.Join(root, "social", "license.png"), 1200, 629)
	if err := Check(root); err == nil || !strings.Contains(err.Error(), "1200x630") {
		t.Fatalf("Check wrong social image size error = %v", err)
	}
}

func TestCheckRejectsNonTruecolorSocialImages(t *testing.T) {
	images := map[string]image.Image{
		"grayscale": image.NewGray(image.Rect(0, 0, 1200, 630)),
		"indexed":   image.NewPaletted(image.Rect(0, 0, 1200, 630), color.Palette{color.Black, color.White}),
		"rgba":      image.NewNRGBA(image.Rect(0, 0, 1200, 630)),
	}
	for name, preview := range images {
		t.Run(name, func(t *testing.T) {
			root := writeValidPublic(t)
			writeImagePNG(t, filepath.Join(root, "social", "brand.png"), preview)
			if err := Check(root); err == nil || !strings.Contains(err.Error(), "PNG") {
				t.Fatalf("Check() error = %v, want exact PNG format failure", err)
			}
		})
	}
}

func TestCheckRejectsRelativeMetadataAndInvalidJSONLD(t *testing.T) {
	root := writeValidPublic(t)
	page := filepath.Join(root, "en", "index.html")
	html, err := os.ReadFile(page)
	if err != nil {
		t.Fatal(err)
	}
	html = []byte(strings.Replace(string(html), `href="https://araihu.com/en/"`, `href="/en/"`, 1))
	if err := os.WriteFile(page, html, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Check(root); err == nil || !strings.Contains(err.Error(), "canonical") {
		t.Fatalf("Check relative canonical error = %v", err)
	}
}

func TestCheckRejectsMissingLocalDocumentResources(t *testing.T) {
	root := writeValidPublic(t)
	page := filepath.Join(root, "en", "index.html")
	html, err := os.ReadFile(page)
	if err != nil {
		t.Fatal(err)
	}
	html = []byte(strings.Replace(string(html), "</body>", `<img src="/assets/missing.svg" alt="" width="1" height="1">`+"</body>", 1))
	if err := os.WriteFile(page, html, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Check(root); err == nil || !strings.Contains(err.Error(), "missing.svg") {
		t.Fatalf("Check missing local resource error = %v", err)
	}
}

func TestCheckAcceptsExistingDotlessVersionedAsset(t *testing.T) {
	root := writeValidPublic(t)
	mutatePage(t, root, "/en/", func(document string) string {
		return strings.Replace(document, "</body>", `<a href="/assets/releases/v0.1.1/NOTICE" download>NOTICE</a>`+"</body>", 1)
	})
	if err := Check(root); err != nil {
		t.Fatal(err)
	}
}

func TestCheckRejectsExistingDotlessScriptAndImage(t *testing.T) {
	for name, element := range map[string]string{
		"script": `<script src="/assets/releases/v0.1.1/NOTICE"></script>`,
		"image":  `<img src="/assets/releases/v0.1.1/NOTICE" alt="" width="1" height="1">`,
	} {
		t.Run(name, func(t *testing.T) {
			root := writeValidPublic(t)
			mutatePage(t, root, "/en/", func(document string) string {
				return strings.Replace(document, "</body>", element+"</body>", 1)
			})
			if err := Check(root); err == nil || !strings.Contains(err.Error(), "must name a file") {
				t.Fatalf("Check() error = %v, want must name a file", err)
			}
		})
	}
}

func TestCheckRejectsMissingDotlessAssetAndUnknownExtensionlessPage(t *testing.T) {
	for _, resource := range []string{
		"/assets/releases/v0.1.1/MISSING",
		"/unknown-extensionless-page",
	} {
		t.Run(resource, func(t *testing.T) {
			root := writeValidPublic(t)
			mutatePage(t, root, "/en/", func(document string) string {
				return strings.Replace(document, "</body>", `<a href="`+resource+`">missing</a>`+"</body>", 1)
			})
			if err := Check(root); err == nil || !strings.Contains(err.Error(), "unknown local page route") {
				t.Fatalf("Check() error = %v, want unknown local page route", err)
			}
		})
	}
}

func TestCheckRejectsMissingLocalScriptDownloadAndPage(t *testing.T) {
	tests := []struct {
		name     string
		fragment string
		want     string
	}{
		{name: "script", fragment: `<script src="/assets/missing.js"></script>`, want: "missing.js"},
		{name: "download", fragment: `<a href="/downloads/missing.zip">download</a>`, want: "missing.zip"},
		{name: "page", fragment: `<a href="/missing-page/">page</a>`, want: "missing-page"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeValidPublic(t)
			mutatePage(t, root, "/en/", func(document string) string {
				return strings.Replace(document, "</body>", test.fragment+"</body>", 1)
			})
			if err := Check(root); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Check() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestCheckRejectsTraversalInLocalResource(t *testing.T) {
	for _, resource := range []string{
		"/assets/%2e%2e/assets/styles.css",
		"/assets%2f..%2fsocial/license.png",
		"/assets%2F.%2e%2Fsocial/license.png",
	} {
		t.Run(resource, func(t *testing.T) {
			root := writeValidPublic(t)
			mutatePage(t, root, "/en/", func(document string) string {
				return strings.Replace(document, "</body>", `<script src="`+resource+`"></script>`+"</body>", 1)
			})
			if err := Check(root); err == nil || !strings.Contains(err.Error(), "traversal") {
				t.Fatalf("Check() error = %v, want traversal failure", err)
			}
		})
	}
}

func TestCheckAcceptsReorderedSitemap(t *testing.T) {
	root := writeValidPublic(t)
	pages := site.Pages()
	for left, right := 0, len(pages)-1; left < right; left, right = left+1, right-1 {
		pages[left], pages[right] = pages[right], pages[left]
	}
	writeFile(t, filepath.Join(root, "sitemap.xml"), sitemapXML(pages, "http://www.sitemaps.org/schemas/sitemap/0.9"))
	if err := Check(root); err != nil {
		t.Fatal(err)
	}
}

func TestCheckRejectsSitemapWithoutNamespaceOrWithDuplicateURL(t *testing.T) {
	t.Run("namespace", func(t *testing.T) {
		root := writeValidPublic(t)
		writeFile(t, filepath.Join(root, "sitemap.xml"), sitemapXML(site.Pages(), ""))
		if err := Check(root); err == nil || !strings.Contains(err.Error(), "namespace") {
			t.Fatalf("Check() error = %v, want namespace failure", err)
		}
	})
	t.Run("duplicate", func(t *testing.T) {
		root := writeValidPublic(t)
		pages := site.Pages()
		pages[len(pages)-1] = pages[0]
		writeFile(t, filepath.Join(root, "sitemap.xml"), sitemapXML(pages, "http://www.sitemaps.org/schemas/sitemap/0.9"))
		if err := Check(root); err == nil || !strings.Contains(err.Error(), "duplicate") {
			t.Fatalf("Check() error = %v, want duplicate failure", err)
		}
	})
	for _, mutation := range []struct {
		name string
		from string
		to   string
	}{
		{name: "url child reset", from: "<url>", to: `<url xmlns="">`},
		{name: "loc child reset", from: "<loc>", to: `<loc xmlns="">`},
	} {
		t.Run(mutation.name, func(t *testing.T) {
			root := writeValidPublic(t)
			path := filepath.Join(root, "sitemap.xml")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, path, []byte(strings.Replace(string(data), mutation.from, mutation.to, 1)))
			if err := Check(root); err == nil || !strings.Contains(err.Error(), "namespace") {
				t.Fatalf("Check() error = %v, want child namespace failure", err)
			}
		})
	}
}

func TestCheckRejectsAdversarialDiscoveryDocuments(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t *testing.T, root string)
		want   string
	}{
		{
			name: "commented canonical tag",
			mutate: func(t *testing.T, root string) {
				mutatePage(t, root, "/en/", func(html string) string {
					html = removeTag(t, html, `<link rel="canonical"`)
					return strings.Replace(html, "</head>", `<!-- <link rel="canonical" href="https://araihu.com/en/"> -->`+"</head>", 1)
				})
			},
			want: "canonical",
		},
		{
			name: "wrong reciprocal hreflang target",
			mutate: func(t *testing.T, root string) {
				mutatePage(t, root, "/en/", func(html string) string {
					return strings.Replace(html, `hreflang="es" href="https://araihu.com/es/"`, `hreflang="es" href="https://araihu.com/en/"`, 1)
				})
			},
			want: "alternate",
		},
		{
			name: "missing open graph title",
			mutate: func(t *testing.T, root string) {
				mutatePage(t, root, "/en/", func(html string) string { return removeTag(t, html, `<meta property="og:title"`) })
			},
			want: "og:title",
		},
		{
			name: "missing X card",
			mutate: func(t *testing.T, root string) {
				mutatePage(t, root, "/en/", func(html string) string { return removeTag(t, html, `<meta name="twitter:card"`) })
			},
			want: "twitter:card",
		},
		{
			name: "wrong document language",
			mutate: func(t *testing.T, root string) {
				mutatePage(t, root, "/en/", func(html string) string { return strings.Replace(html, `<html lang="en"`, `<html lang="fr"`, 1) })
			},
			want: "html lang",
		},
		{
			name: "invalid structured data relationship",
			mutate: func(t *testing.T, root string) {
				mutatePage(t, root, "/brand/", func(html string) string {
					return strings.ReplaceAll(html, `https://araihu.com/#organization`, `https://araihu.com/#other`)
				})
			},
			want: "JSON-LD",
		},
		{
			name: "inexact robots directives",
			mutate: func(t *testing.T, root string) {
				path := filepath.Join(root, "robots.txt")
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				writeFile(t, path, []byte(strings.Replace(string(data), "Allow: /", "Allow: /private", 1)))
			},
			want: "robots.txt",
		},
		{
			name: "invalid manifest icon size",
			mutate: func(t *testing.T, root string) {
				path := filepath.Join(root, "site.webmanifest")
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				writeFile(t, path, []byte(strings.Replace(string(data), `"sizes":"192x192"`, `"sizes":"42x42"`, 1)))
			},
			want: "manifest",
		},
		{
			name: "unexpected sitemap route",
			mutate: func(t *testing.T, root string) {
				const oldURL = "https://araihu.com/es/license/"
				const newURL = "https://araihu.com/es/license-copy/"
				original := filepath.Join(root, "es", "license", "index.html")
				data, err := os.ReadFile(original)
				if err != nil {
					t.Fatal(err)
				}
				html := strings.ReplaceAll(string(data), oldURL, newURL)
				html = strings.Replace(html, "<title>", "<title>Copy ", 1)
				html = strings.Replace(html, `name="description" content="`, `name="description" content="Copy `, 1)
				writeFile(t, filepath.Join(root, "es", "license-copy", "index.html"), []byte(html))
				sitemap := filepath.Join(root, "sitemap.xml")
				data, err = os.ReadFile(sitemap)
				if err != nil {
					t.Fatal(err)
				}
				writeFile(t, sitemap, []byte(strings.Replace(string(data), oldURL, newURL, 1)))
			},
			want: "sitemap",
		},
		{
			name: "duplicate html attribute",
			mutate: func(t *testing.T, root string) {
				mutatePage(t, root, "/en/", func(html string) string {
					return strings.Replace(html, `<html lang="en"`, `<html lang="en" lang="es"`, 1)
				})
			},
			want: "duplicate",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeValidPublic(t)
			test.mutate(t, root)
			if err := Check(root); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Check() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestCheckRejectsInvalidAssetReleaseBundle(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t *testing.T, root string)
		want   string
	}{
		{
			name: "missing current channel",
			mutate: func(t *testing.T, root string) {
				if err := os.Remove(filepath.Join(root, "assets", "releases", "current.json")); err != nil {
					t.Fatal(err)
				}
			},
			want: "current.json",
		},
		{
			name: "missing current runtime",
			mutate: func(t *testing.T, root string) {
				if err := os.Remove(filepath.Join(root, "assets", "campaign", "v1.js")); err != nil {
					t.Fatal(err)
				}
			},
			want: "campaign/v1.js",
		},
		{
			name: "missing immutable release document",
			mutate: func(t *testing.T, root string) {
				if err := os.Remove(filepath.Join(root, "assets", "releases", "v0.1.1", "release.json")); err != nil {
					t.Fatal(err)
				}
			},
			want: "release.json",
		},
		{
			name: "invalid release checksums",
			mutate: func(t *testing.T, root string) {
				writeFile(t, filepath.Join(root, "assets", "releases", "v0.1.1", "checksums.txt"), []byte("not a checksum\n"))
			},
			want: "checksums",
		},
		{
			name: "relative channel URL",
			mutate: func(t *testing.T, root string) {
				writeReleaseChannel(t, root, "current", "/assets/releases/v0.1.1/themes/base.css")
			},
			want: "absolute same-origin URL",
		},
		{
			name: "invalid SemVer directory",
			mutate: func(t *testing.T, root string) {
				if err := os.Rename(filepath.Join(root, "assets", "releases", "v0.1.1"), filepath.Join(root, "assets", "releases", "v0.1.01")); err != nil {
					t.Fatal(err)
				}
			},
			want: "outside known layout",
		},
		{
			name: "missing retained inventory file",
			mutate: func(t *testing.T, root string) {
				if err := os.Remove(filepath.Join(root, "assets", "releases", "v0.1.1", "themes", "base.css")); err != nil {
					t.Fatal(err)
				}
			},
			want: "inventory misses file",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := writeValidPublic(t)
			test.mutate(t, root)
			if err := Check(root); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Check() error = %v, want %q", err, test.want)
			}
		})
	}
}

func writeValidPublic(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "robots.txt"), site.Robots())
	sitemap, err := site.Sitemap(site.Pages())
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "sitemap.xml"), sitemap)
	writeFile(t, filepath.Join(root, "site.webmanifest"), site.SiteManifest())
	for _, name := range []string{"styles.css", "ahairu.css", "araihu-theme.css"} {
		writeFile(t, filepath.Join(root, "assets", name), []byte("fixture"))
	}
	writePNG(t, filepath.Join(root, "social", "brand.png"), 1200, 630)
	writePNG(t, filepath.Join(root, "social", "license.png"), 1200, 630)
	for _, page := range site.Pages() {
		var content templ.Component = templ.NopComponent
		if page.Meta.Kind == site.PageBrand {
			content = templ.Raw(`<img src="/assets/releases/v0.1.1/brand/araihu/logo/tinted-transparent-optical.svg" alt="Arai Hû" width="720" height="134" data-asset-brand="logo">`)
		}
		var output strings.Builder
		if err := site.Layout(page, content).Render(context.Background(), &output); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(root, page.Meta.Path, "index.html"), []byte(output.String()))
	}
	writeAssetReleaseBundle(t, root)
	return root
}

func writeAssetReleaseBundle(t *testing.T, root string) {
	t.Helper()
	releaseRoot := filepath.Join(root, "assets", "releases", "v0.1.1")
	files := map[string][]byte{
		"catalog.json":    []byte(`{"schemaVersion":1}`),
		"themes.json":     []byte(`{"schemaVersion":1}`),
		"campaigns.json":  []byte(`{"schemaVersion":1,"campaigns":[]}`),
		"campaign/v1.js":  []byte("(() => {})()\n"),
		"themes/base.css": []byte("body{}\n"),
	}
	paths, err := site.BundledBrandAssetPaths()
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range paths {
		if name == "catalog.json" || name == "themes.json" || name == "campaigns.json" || name == "release.json" || name == "checksums.txt" {
			continue
		}
		contents, err := fs.ReadFile(site.BrandAssets(), name)
		if err != nil {
			t.Fatal(err)
		}
		files[name] = contents
	}
	ordered := make([]string, 0, len(files))
	for name := range files {
		ordered = append(ordered, name)
	}
	sort.Slice(ordered, func(left, right int) bool {
		order := map[string]int{"catalog.json": 0, "themes.json": 1, "campaigns.json": 2}
		leftOrder, leftKnown := order[ordered[left]]
		rightOrder, rightKnown := order[ordered[right]]
		if leftKnown && rightKnown {
			return leftOrder < rightOrder
		}
		if leftKnown {
			return true
		}
		if rightKnown {
			return false
		}
		return ordered[left] < ordered[right]
	})
	type inventoryFile struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
		Size   int64  `json:"size"`
	}
	inventory := make([]inventoryFile, 0, len(ordered))
	for _, name := range ordered {
		contents := files[name]
		writeFile(t, filepath.Join(releaseRoot, filepath.FromSlash(name)), contents)
		sum := sha256.Sum256(contents)
		inventory = append(inventory, inventoryFile{Path: name, SHA256: hex.EncodeToString(sum[:]), Size: int64(len(contents))})
	}
	document := struct {
		SchemaVersion    int             `json:"schemaVersion"`
		Release          string          `json:"release"`
		IdentityRevision int             `json:"identityRevision"`
		RuntimeVersion   int             `json:"runtimeVersion"`
		CatalogSHA256    string          `json:"catalogSha256"`
		ThemesSHA256     string          `json:"themesSha256"`
		CampaignsSHA256  string          `json:"campaignsSha256"`
		Files            []inventoryFile `json:"files"`
	}{SchemaVersion: 1, Release: "v0.1.1", IdentityRevision: 11, RuntimeVersion: 1, Files: inventory}
	for _, item := range inventory {
		switch item.Path {
		case "catalog.json":
			document.CatalogSHA256 = item.SHA256
		case "themes.json":
			document.ThemesSHA256 = item.SHA256
		case "campaigns.json":
			document.CampaignsSHA256 = item.SHA256
		}
	}
	releaseJSON, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(releaseRoot, "release.json"), releaseJSON)
	checksumNames := append([]string{"release.json"}, ordered...)
	checksums := make([]string, 0, len(checksumNames))
	for _, name := range checksumNames {
		contents, err := os.ReadFile(filepath.Join(releaseRoot, filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(contents)
		checksums = append(checksums, hex.EncodeToString(sum[:])+"  "+name)
	}
	writeFile(t, filepath.Join(releaseRoot, "checksums.txt"), []byte(strings.Join(checksums, "\n")+"\n"))
	writeFile(t, filepath.Join(root, "assets", "campaign", "v1.js"), files["campaign/v1.js"])
	for _, channel := range []string{"latest", "default", "current"} {
		writeReleaseChannel(t, root, channel, "https://araihu.com/assets/releases/v0.1.1/themes/base.css")
	}
}

func writeReleaseChannel(t *testing.T, root, channel, cssURL string) {
	t.Helper()
	type theme struct {
		ID     string `json:"id"`
		CSSURL string `json:"cssUrl"`
	}
	value := struct {
		SchemaVersion  int    `json:"schemaVersion"`
		RuntimeVersion int    `json:"runtimeVersion"`
		Release        string `json:"release"`
		Source         string `json:"source"`
		Theme          theme  `json:"theme"`
		Digest         string `json:"digest"`
	}{SchemaVersion: 1, RuntimeVersion: 1, Release: "v0.1.1", Source: "default", Theme: theme{ID: "base", CSSURL: cssURL}}
	payload := canonicalJSON(t, value)
	sum := sha256.Sum256(payload)
	value.Digest = hex.EncodeToString(sum[:])
	writeFile(t, filepath.Join(root, "assets", "releases", channel+".json"), canonicalJSON(t, value))
}

func canonicalJSON(t *testing.T, value any) []byte {
	t.Helper()
	var output bytes.Buffer
	encoder := json.NewEncoder(&output)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writePNG(t *testing.T, path string, width, height int) {
	t.Helper()
	preview := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.Draw(preview, preview.Bounds(), image.NewUniform(color.NRGBA{R: 7, G: 17, B: 31, A: 255}), image.Point{}, draw.Src)
	writeImagePNG(t, path, preview)
}

func writeImagePNG(t *testing.T, path string, preview image.Image) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, preview); err != nil {
		t.Fatal(err)
	}
}

func mutatePage(t *testing.T, root, route string, mutate func(string) string) {
	t.Helper()
	path := filepath.Join(root, route, "index.html")
	html, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, []byte(mutate(string(html))))
}

func removeTag(t *testing.T, html, prefix string) string {
	t.Helper()
	start := strings.Index(html, prefix)
	if start < 0 {
		t.Fatalf("tag %q not found", prefix)
	}
	end := strings.Index(html[start:], ">")
	if end < 0 {
		t.Fatalf("tag %q is malformed", prefix)
	}
	return html[:start] + html[start+end+1:]
}

func sitemapXML(pages []site.Page, namespace string) []byte {
	var document strings.Builder
	if namespace == "" {
		document.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?><urlset>")
	} else {
		fmt.Fprintf(&document, "<?xml version=\"1.0\" encoding=\"UTF-8\"?><urlset xmlns=\"%s\">", namespace)
	}
	for _, page := range pages {
		fmt.Fprintf(&document, "<url><loc>%s</loc></url>", page.Meta.CanonicalURL)
	}
	document.WriteString("</urlset>")
	return []byte(document.String())
}
