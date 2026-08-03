# Arai Hû Storm Studio Site Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a distinctive, project-first Arai Hû organization site that borrows Charm's clarity, irregular editorial rhythm, and tactile motion while using Arai Hû's own storm-weather identity.

**Architecture:** Keep the organization site static, multilingual, and server-rendered with templ. Put reusable behavior in Goshtoso, responsive public-site composition in Goshtoso App Shells, and organization-specific art direction in Arai Hû. No runtime-heavy background effects are required: CSS layers provide clouds, rain, grain, and restrained lightning with reduced-motion fallbacks.

**Tech Stack:** Go, templ, Goshtoso, CSS, Node test runner, Cloudflare Workers/Wrangler, in-app browser QA.

## Global Constraints

- Canonical motif: `Software for stormy weather.`
- Supporting promise: independent, open tools built to endure difficult work.
- Preserve the approved Arai Hû v11 asset geometry; choose transparent/background variants by production context.
- Navigation taxonomy stays `Home`, `Libs`, `Apps`, and `Blog`; unfinished destinations are visibly marked `WIP`.
- Blog and mailing-list controls must not imply live functionality before a real backend exists.
- All animation must respect `prefers-reduced-motion`.
- Arai Hû owns brand expression; Goshtoso owns component behavior; App Shells owns responsive shell policy.
- Every slice ends with generated-output checks, focused tests, `git diff --check`, and literal desktop/mobile browser evidence.

---

### Task 1: Canonical Storm Motif and Localized Copy

**Files:**
- Modify: `site/content.go`
- Modify: `cmd/ahairu/main_test.go`
- Regenerate: `site/home_templ.go`

**Interfaces:**
- Produces: localized `Content.Tagline` and `Content.Intro` values consumed by metadata, desktop/mobile brand marks, and the hero.

- [x] Update the generated-site test to require the English, Brazilian Portuguese, and Spanish storm taglines.
- [x] Run `go test ./cmd/ahairu -run TestBuildWritesStandaloneSite -count=1` and confirm failure against the old copy.
- [x] Replace the three taglines and supporting introductions in `site/content.go`.
- [x] Run `templ generate` and rerun the focused test.
- [x] Run `git diff --check` and review all three generated locale pages.

### Task 2: Storm Atmosphere and Motion Tokens

**Files:**
- Modify: `site/brand.css`
- Modify: `DESIGN.md`
- Test: `cmd/ahairu/main_test.go`

**Interfaces:**
- Produces: storm cloud, rain, grain, flash, pressed-shadow, and motion-duration tokens for later sections.

- [x] Add markup/style contract assertions for the storm tokens and reduced-motion stylesheet.
- [x] Confirm the assertions fail before the final token names exist.
- [x] Consolidate cloud, rain, grain, and interaction values into named custom properties.
- [x] Document intended motion, contrast, and reduced-motion behavior in `DESIGN.md`.
- [x] Verify desktop hero motion and a reduced-motion browser profile.

### Task 3: Project-First Information Architecture

**Files:**
- Modify: `site/content.go`
- Modify: `site/home.templ`
- Test: `cmd/ahairu/main_test.go`

**Interfaces:**
- Produces: one featured project, separate library/application collections, and a secondary-project lane.

- [x] Test the required section order: hero, featured Goshtoso, libraries, applications, more projects, mission, field notes/chat.
- [x] Test that every project has a real destination, category, and localized description.
- [x] Render Goshtoso as the dominant wide card and Manja, Pajé, and X-9 in an uneven grid.
- [x] Render App Shells, Charts, and Muamba with alternating filled/unfilled rows.
- [x] Verify heading hierarchy and landmark names in all locales.

### Task 4: Expressive Goshtoso Card Primitive

**Files:**
- Modify: `components/card/types.go`
- Modify: `components/card/card.templ`
- Test: `components/card/card_coverage_test.go`
- Modify: `site/internal/pages/demo/componentpages/card/card.templ`

**Interfaces:**
- Produces: `InteractionPressed`, `Media templ.Component`, and `MediaClass string`.

- [x] Test custom-media precedence over `ImageURL` and pressed interaction classes.
- [x] Implement the minimal additive Card API.
- [x] Add a project-card demo with art crossing visual regions.
- [x] Regenerate templ, CSS, and consumer skill references.
- [x] Run root and site Go suites against current source.

### Task 5: Four-Direction Goshtoso Drawer

**Files:**
- Modify: `components/drawer/types.go`
- Test: `components/drawer/drawer_coverage_test.go`
- Modify: `site/internal/pages/demo/componentpages/drawer/drawer.templ`

**Interfaces:**
- Produces: `SideTop`, `SideBottom`, `Height`, and `HeightSM|MD|LG|XL|Full` while preserving left/right defaults.

- [x] Test top/bottom geometry and enter/leave transforms.
- [x] Implement side-specific dimensions, borders, and translation axes.
- [x] Add top-navigation and bottom-action demos.
- [x] Regenerate templ, CSS, and consumer skill references.
- [x] Run root and site Go suites against current source.

### Task 6: Floating Mobile Navigation Composition

**Files:**
- Modify: `site/home.templ`
- Modify: `site/brand.css`
- Test: `cmd/ahairu/main_test.go`
- Later modify after Goshtoso release: `goshtoso-app-shells/landingshell/config.go`
- Later modify after Goshtoso release: `goshtoso-app-shells/landingshell/layout.templ`

**Interfaces:**
- Site produces: native no-runtime `<details>` fallback.
- App Shells later produces: responsive navigation policy composed from a floating trigger and top Drawer.

- [x] Test that desktop keeps full navigation and mobile receives a floating trigger plus top drawer.
- [x] Implement the bottom-left trigger and storm-tinted top drawer without changing document flow.
- [x] Verify keyboard open/close, focus visibility, current locale, and WIP labels at 390px.
- [x] After directional Drawer is released, add an opt-in landing-shell mobile-navigation config.
- [x] Keep `ActionGroup` unchanged; floating placement belongs to shell composition, not action semantics.

### Task 7: Interaction and Animation Proof

**Files:**
- Modify: `site/brand.css`
- Test: `cmd/ahairu/main_test.go`
- Create: `docs/design/storm-motion-inventory.md`

**Interfaces:**
- Produces: an explicit inventory of initial, hover, active, open, close, scroll, and reduced-motion states.

- [x] Record every moving surface and its purpose; remove decorative motion without a hierarchy or affordance role.
- [x] Verify project-card hover compresses shadow and moves toward the page.
- [x] Verify art layers shift independently without obscuring names or actions.
- [x] Verify drawer open/close and trigger state at touch size.
- [x] Capture desktop, mobile-closed, mobile-open, project-grid, and reduced-motion evidence.

### Task 8: Mission, Community, and Honest WIP Surfaces

**Files:**
- Modify: `site/content.go`
- Modify: `site/home.templ`
- Test: `cmd/ahairu/main_test.go`

**Interfaces:**
- Produces: open-source mission, GitHub conversation CTA, and disabled field-notes preview.

- [x] Test that the mailing form is disabled and labeled as coming soon.
- [x] Test that the community CTA points to the organization GitHub page.
- [x] Keep the mission copy short, centered, and specific to public Go software.
- [x] Keep a real mailing provider in a separate backend slice; this site leaves the preview visibly disabled.
- [x] Verify localized status and action labels.

### Task 9: Metadata, Social Preview, and Release Gate

**Files:**
- Modify: `site/home.templ`
- Create: `site/assets/og.png` only after the visual direction is frozen
- Modify: `cmd/ahairu/main_test.go`

**Interfaces:**
- Produces: localized page title/description and one storm-themed social preview.

- [ ] Test canonical title, description, theme color, favicon, and social metadata.
- [ ] Generate one complete landscape social card using the final storm motif and approved assets.
- [ ] Inspect all authored text in the image; retry once only if unusable.
- [x] Run `npm run check` and Wrangler dry-run.
- [x] Integrate Goshtoso feature worktrees in dependency order, then update App Shells and the site dependency pin.
- [x] Perform final 390px, 768px, 1280px, and reduced-motion browser acceptance.

## Current Checkpoint

- [x] Charm reference sections captured and motion behavior inspected.
- [x] Storm hero, uneven project grid, mission, community, and WIP surfaces prototyped locally.
- [x] Native floating mobile menu prototyped and visually verified.
- [x] Expressive Card implemented and tested in an isolated Goshtoso worktree.
- [x] Four-direction Drawer implemented and tested in an isolated Goshtoso worktree.
- [x] Canonical storm motif applied to all locales.
- [x] Goshtoso feature branches reviewed, committed, and integrated.
- [x] Landing Shell composition implemented against released Goshtoso APIs.
- [x] Final multi-breakpoint release acceptance completed against Goshtoso v0.1.7 and App Shells v0.1.3.
- [ ] Social preview remains a separate follow-up after this visual direction is accepted.

## Adversarial Review Round 1 Remediation

- [x] Render desktop and mobile locale controls from one shared component so the active locale remains visible in the open drawer.
- [x] Inventory every transition, animation, and spatial transform; neutralize all of them under `prefers-reduced-motion: reduce`.
- [x] Add real-browser reduced-motion checks for desktop card hover and 390px menu open/close behavior.
- [x] Merge the expressive Card and four-direction Drawer APIs into Goshtoso, pass the two-module release gates, and publish a public release.
- [x] Implement floating mobile-navigation policy in Goshtoso App Shells using the released Drawer, retain a documented native `<details>` fallback, and publish a public release.
- [x] Pin both public releases in Arai Hû; compose project cards from Goshtoso Card and mobile navigation from App Shells instead of owning those behaviors locally.
- [x] Prove the enhanced and no-runtime fallback paths at 390px, plus desktop and reduced-motion behavior, across EN, PT-BR, and ES.
- [x] Freeze the revised Git target, emit all three adversarial-review checkpoints, and resubmit the same target to the independent judge.

## Adversarial Review Round 2 Remediation

- [x] Remove the site-local 700px taxonomy cutoff so the desktop surface remains visible until App Shell mobile navigation takes ownership at 640px.
- [x] Add real-Chromium boundary assertions at 640, 641, 700, and 701px requiring exactly one project-navigation surface.
- [x] Rebuild and capture the served 650px page with Home, Libs, Apps, and Blog visible and no horizontal overflow.
- [x] Re-freeze the revised Git target, emit all three checkpoints, and resubmit the same target to the independent judge.

## Follow-up: Theme-Aware Storm Footage

- [x] Identify the supplied night and day footage and remove their audio streams.
- [x] Crop generation-edge artifacts, resize to 960×534 at 20 fps, and keep each H.264 asset below 500 KiB.
- [x] Select exactly one source from the system color scheme and place a dedicated filter layer above the footage.
- [x] Add bounded scroll parallax while preserving full hero coverage.
- [x] Skip loading and playback for reduced-motion or data-saving users.
- [x] Prove dark/light selection, transfer exclusivity, coverage, parallax, reduced motion, desktop/mobile layout, and Wrangler packaging.
