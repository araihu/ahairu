# Pajé DAG, Project Lockups, and Social Metadata Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace Pajé's random force graph with a large localized lifecycle DAG, integrate project marks into application headings, and complete initial-HTML social metadata for all locale routes.

**Architecture:** Keep locale-owned copy in `site/content.go`, fixed DAG topology and presentation in `site/project_charts.go`, semantic composition in `site/home.templ`, and visual behavior in `site/brand.css`. Continue using deferred HTMX chart fragments. Render social metadata directly in the shared templ head because pinned Goshtoso `v0.1.7` exposes dependency helpers but no metadata primitive.

**Tech Stack:** Go, templ, Goshtoso, Goshtoso Charts `v0.0.1`, HTMX, CSS, static generation, in-app browser.

## Global Constraints

- Execute in current dirty `main` lane because it contains the approved visual work; preserve all unrelated changes.
- Do not use TDD. Implement first, then run focused regression and visual browser checks as requested.
- Do not adopt or pin the unreleased Goshtoso Charts `Surface3D` work.
- Pajé graph uses `interactive.Graph`, `GraphLayoutNone`, 12 localized nodes, and 13 directed links.
- Expensive charts remain absent from initial HTML and arrive through the existing HTMX fragment.
- Hover replays allowed motion; reduced-motion mode stays static.
- Every locale route emits complete social metadata exactly once in initial generated HTML.
- Social image is `site/assets/social/araihu-storm-v1.jpg`, published at `https://araihu.com/assets/social/araihu-storm-v1.jpg`, `1280x640`, JPEG, below `1 MB`.

---

### Task 1: Localized Lifecycle Model and Deterministic DAG

**Files:**
- Modify: `site/content.go`
- Modify: `site/project_charts.go`
- Modify: `site/chart_fragments.templ`
- Regenerate: `site/chart_fragments_templ.go`

**Interfaces:**
- Produces: `PajeLifecycleLabels` on `Content` and `PajeWorkflowGraph(project Project, labels PajeLifecycleLabels) interactive.Instance`.
- Consumes: existing localized `Project` and deferred `ChartFragment` composition.

- [ ] **Step 1: Add localized lifecycle labels**

Define `PajeLifecycleLabels` with fields `Discovery`, `WebResearch`, `Specification`, `SpecificationApproval`, `Implementation`, `UnitTests`, `IntegrationTests`, `Documentation`, `CI`, `AdversarialReview`, `ReleaseApproval`, and `Publish`. Populate EN, PT-BR, and ES content without empty fields.

- [ ] **Step 2: Replace force graph with fixed DAG**

Build 12 nodes with explicit coordinates and stable IDs encoded in each node name, use `GraphLayoutNone`, remove `Force`, show labels, categorize work/validation/gate nodes, and build the 13 links from the approved spec.

- [ ] **Step 3: Pass labels through the fragment**

Update `ChartFragment` to call `PajeWorkflowGraph(content.Projects[2], content.PajeLifecycle)` and keep the current `paje-chart-slot` replacement contract.

- [ ] **Step 4: Regenerate templ output**

Run:

```bash
templ generate
```

Expected: `site/chart_fragments_templ.go` and `site/home_templ.go` match authored templates.

### Task 2: Application Project Title Lockups

**Files:**
- Modify: `site/home.templ`
- Modify: `site/brand.css`
- Regenerate: `site/home_templ.go`

**Interfaces:**
- Produces: a three-column content grid where `.project-title-mark` and the Goshtoso Card heading share row 2.
- Consumes: `Project.MarkURL`, `ProjectFooter`, and Goshtoso Card's existing heading/description relationship.

- [ ] **Step 1: Render mark beside title**

Keep Goshtoso Card's semantic `h3` and `aria-describedby` output. Render the decorative mark through `ProjectFooter`, rename it `.project-title-mark`, and place it in grid column 1 while the heading occupies column 2 on the same row. Remove the former loose right-side placement.

- [ ] **Step 2: Style shared plate**

Add a `44px` light plate, blue border, subtle shadow, and small resting rotation. Keep text aligned across Manja, Pajé, and X-9. On hover/focus-within straighten the plate; under reduced motion remove transition and transform.

- [ ] **Step 3: Regenerate templ output**

Run `templ generate` and inspect generated diffs for only authored-template consequences.

### Task 3: Initial-HTML Social Metadata and Preview Asset

**Files:**
- Modify: `site/content.go`
- Modify: `site/home.templ`
- Modify: `site/brand.go`
- Modify: `cmd/ahairu/main.go`
- Create: `site/assets/social/araihu-storm-v1.jpg`
- Regenerate: `site/home_templ.go`

**Interfaces:**
- Produces: localized `Content.PageTitle`, `Content.SocialDescription`, `Content.SocialImageAlt`, `Content.CanonicalURL`; embedded `SocialAsset`; generated `/assets/social/araihu-storm-v1.jpg`.
- Consumes: final storm identity, locale paths, static build pipeline.

- [ ] **Step 1: Add route metadata values**

Set canonical URLs for EN, PT-BR, and ES. Use localized title, description, `og:locale`, and image alt. Keep one shared text-free storm image.

- [ ] **Step 2: Render metadata once**

In `Page`, replace standalone title/description with one explicit metadata block containing canonical, Open Graph, image structured properties, and Twitter/X large-card tags. Keep `head.Dependencies` for runtime assets. Do not emit metadata in fragments.

- [ ] **Step 3: Produce preview image**

Extract a strong storm frame from `site/backdrops/storm-dark-v1.mp4`, apply the site's navy treatment, and place the Arai Hû mark inside safe crop margins. Encode JPEG at `1280x640` below `1 MB`.

- [ ] **Step 4: Embed and publish asset**

Add the image to the site's embedded asset inventory and static builder so `public/assets/social/araihu-storm-v1.jpg` is always generated.

- [ ] **Step 5: Regenerate templ output**

Run `templ generate`.

### Task 4: Visual Browser Acceptance and Regression Gate

**Files:**
- Modify only if browser evidence reveals defects: `site/brand.css`, `site/home.templ`, `site/project_charts.go`, `site/content.go`
- Regenerate after any template correction: `site/home_templ.go`, `site/chart_fragments_templ.go`

**Interfaces:**
- Consumes: built site served at `http://127.0.0.1:4187/pt-br/` and equivalent locale routes.
- Produces: verified desktop/mobile visual behavior and social metadata evidence.

- [ ] **Step 1: Build and run regression checks**

Run:

```bash
templ generate
GOWORK=off npm run check
git diff --check
```

Expected: all existing Go, browser, worker, generated-output, and Wrangler checks pass.

- [ ] **Step 2: Verify responsive visuals in the in-app browser**

At viewport widths `390`, `884`, and `1280`, inspect Pajé graph framing, visible labels, project title lockups, card content, and horizontal overflow. Capture screenshots at each width.

- [ ] **Step 3: Verify interaction states**

Hover and leave Pajé twice; confirm deterministic replay and stable final coordinates. Hover all title plates. Emulate reduced motion and confirm graph/plates remain static.

- [ ] **Step 4: Verify metadata and preview bytes**

Fetch `/en/`, `/pt-br/`, and `/es/`; confirm each required tag appears once with route-correct production HTTPS values. Fetch the preview and verify HTTP `200`, `image/jpeg`, `1280x640`, and size below `1 MB`.

- [ ] **Step 5: Rebuild after corrections**

Repeat the full command gate and browser checkpoints until clean. Do not call the landing complete with broken visual framing or incomplete metadata.
