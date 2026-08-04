# Pajé Lifecycle DAG, Project Lockups, and Social Metadata

**Date:** 2026-08-03
**Status:** Design approved; written specification awaiting review

## Context

The application cards currently place each project mark loose on the far right of the text region. The marks have inconsistent contrast and look detached from the project names. Pajé's current force-directed chart is visually small, randomly arranged, and communicates a generic network instead of Pajé's durable software-development workflow.

The public locale routes also lack the complete initial-HTML social metadata contract required for reliable search and link previews.

## Goals

- Give Manja, Pajé, and X-9 one consistent, high-contrast title lockup.
- Replace Pajé's random force graph with a large, deterministic directed acyclic graph representing a realistic company development lifecycle.
- Keep expensive charts deferred and interaction motion hover-triggered.
- Preserve reduced-motion behavior and readable fallbacks.
- Make `/en/`, `/pt-br/`, and `/es/` complete social-sharing surfaces in their initial generated HTML.

## Non-goals

- No Goshtoso Charts API extension. Existing fixed-position graph support is sufficient.
- No adoption of the new Charts `Surface3D` heart in this slice. That upstream capability remains separately release- and consumer-gated.
- No redesign of project descriptions, destinations, card ordering, or the storm hero.
- No client-side metadata injection.

## Project Title Lockup

Manja, Pajé, and X-9 use one shared heading composition:

- project mark appears immediately left of the project name;
- mark sits inside a `44px` square light plate with a blue border and enough contrast on light, blue, and lime card backgrounds;
- plate receives a small resting rotation consistent with the site's irregular card language;
- hover straightens and slightly translates the plate without moving the title text;
- reduced-motion mode removes transition and transform;
- mark is decorative (`alt=""`) because the adjacent heading already names the project;
- current loose right-side mark is removed.

Markup remains content-first: section eyebrow, title lockup, description, status badge, and project action. The icon must not become an independent focus target.

## Pajé Lifecycle DAG

### Rendering model

Use Goshtoso Charts `interactive.Graph` with `GraphLayoutNone` and explicit node coordinates. Do not use the Tree component: the workflow branches and later merges, while a tree cannot represent shared downstream nodes without duplication. Do not use force layout: replay must remain deterministic and fully framed.

### Workflow

Nodes:

1. Discovery
2. Web research
3. Specification
4. Specification approval
5. Implementation
6. Unit tests
7. Integration tests
8. Documentation
9. CI
10. Adversarial review
11. Release approval
12. Publish

Directed links:

1. Discovery to Web research
2. Web research to Specification
3. Specification to Specification approval
4. Specification approval to Implementation
5. Implementation to Unit tests
6. Implementation to Integration tests
7. Implementation to Documentation
8. Unit tests to CI
9. Integration tests to CI
10. CI to Adversarial review
11. Documentation to Adversarial review
12. Adversarial review to Release approval
13. Release approval to Publish

This topology is acyclic, branches after implementation, and joins validation and documentation before release approval.

### Visual hierarchy

- Layout runs left to right in layers, with the three implementation branches vertically separated.
- Every node has a visible localized label.
- Work nodes use blue/orange tones, validation nodes use pink, and approval/publish gates use lime.
- Edges remain subordinate but visible at site contrast targets.
- Pajé chart fills the card media region rather than floating as a small center cluster.
- At 390px, 884px, and 1280px widths, no label or node may clip or overlap card content.
- Chart is still rendered through the existing deferred HTMX fragment path, preserving first paint.
- Hover creates a fresh chart instance or equivalent deterministic entry replay. Coordinates and topology do not change.
- Pointer exit leaves a stable final frame.
- `prefers-reduced-motion: reduce` renders the final state immediately and does not replay.

### Localization ownership

Lifecycle labels belong to locale content, not the chart renderer. EN, PT-BR, and ES receive complete labels through a dedicated content structure passed into the chart builder. Chart construction owns topology, coordinates, node categories, and presentation only.

### Failure behavior

Before the HTMX fragment arrives, the card keeps its existing lightweight placeholder. Fragment failure must not remove the project name, description, status, icon lockup, or destination. A chart runtime failure may leave the styled media fallback; it must not block navigation.

## Social Metadata

Each public locale route emits metadata exactly once in initial generated HTML, preferably through `github.com/araihu/goshtoso/components/head`:

- localized `<title>`;
- localized `meta[name="description"]`;
- route-specific production canonical URL;
- `og:url`, `og:type=website`, localized `og:title`, localized `og:description`, `og:site_name`, and locale;
- `og:image`, MIME type, `1280x640` dimensions, and localized useful alt text;
- explicit `twitter:card=summary_large_image`, title, description, image, and image alt.

Canonical routes:

- `https://araihu.com/en/`
- `https://araihu.com/pt-br/`
- `https://araihu.com/es/`

Use one intentional storm-branded landscape preview without language-specific embedded text, so all locales can share artwork without misleading copy. Store it at `site/assets/social/araihu-storm-v1.jpg` and publish it as `https://araihu.com/assets/social/araihu-storm-v1.jpg`. It uses JPEG with the matching MIME type, is exactly `1280x640`, and remains below `1 MB`. Keep logo and focal storm elements away from crop edges. Do not invent a Twitter/X account.

HTMX fragments contain no document metadata. Shared layout renders the metadata primitive once.

## Verification

### Go and generated-output tests

- Assert Pajé fragment serializes a fixed-position graph, not force layout.
- Assert exactly 12 nodes and 13 directed links with stable IDs.
- Assert all three locales supply non-empty lifecycle labels.
- Assert application headings render mark before name and contain no loose right-side mark.
- Regenerate templ output and verify no drift.

### Browser tests

- At 390px, 884px, and 1280px, assert every Pajé node and label stays inside media bounds.
- Assert graph occupies the intended media area and project text remains visible.
- Hover twice and assert entry replay occurs while node coordinates remain identical.
- Under reduced motion, assert no hover replay or title-plate transform.
- Assert all three title lockups preserve contrast and alignment on their distinct card backgrounds.
- Assert fragment failure leaves navigable card content.

### Metadata gate

- Fetch generated `/en/`, `/pt-br/`, and `/es/` HTML.
- Assert every required title, description, canonical, Open Graph, image property, image alt, and explicit Twitter/X tag appears exactly once.
- Assert canonical and `og:url` match the current locale route and use production HTTPS URLs.
- Fetch the preview asset and verify `200`, MIME type, `1280x640` dimensions, and size below `1 MB`.
- Run the complete existing `npm run check` and Wrangler dry-run.

## Acceptance Criteria

- Manja, Pajé, and X-9 marks read as part of their names and remain clear on every card surface.
- Pajé visibly communicates the full approved lifecycle with deterministic branches and merges.
- Deferred chart loading does not regress first paint.
- Hover motion replays only when allowed; reduced-motion users receive a stable final graph.
- Locale HTML is social-ready before JavaScript executes.
- All generated-output, browser, metadata, worker, and packaging checks pass.
