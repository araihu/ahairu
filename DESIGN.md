# Arai Hu design system

## Visual world

Archive of a black cloud: a deep midnight field opens into a daylight-after-rain project showcase. Cobalt carries depth, paper-like surfaces carry content, and electric lime remains a signal. The mood is expressive open-source studio, not SaaS landing page.

The canonical motif is **Software for stormy weather.** The supporting promise stays concrete: independent, open tools built to endure difficult work.

Core tokens live in `site/brand.css`: `--storm-night`, `--storm-deep`, `--storm-cloud`, `--storm-rain`, `--storm-mist`, `--storm-paper`, and `--storm-signal`. The hero selects optimized day/night storm footage from the system color scheme, then applies one static site-owned filter layer so text contrast and palette remain stable. No CSS cloud, moon, rain, or lightning imitation sits over the footage; direct Card motion belongs to Goshtoso's `InteractionPressed` behavior.

## Typography

Use the installed Goshtoso system face. Display type is heavy, tight, and oversized; body text stays calm and readable. Project names act as the visual rhythm.

## Composition

The storm field frames one sharp organization statement, then yields quickly to the work. Projects use an uneven editorial grid: each card has its own composition, but shares the same metadata, purpose, and full-card link contract. Goshtoso opens the collection; Manja, Pajé, and X-9 show the ecosystem's range. Locale switcher stays compact and functional.

## Interaction

The weather video uses bounded parallax only when motion is allowed and does not load for reduced-motion or data-saving users. Project cards move toward the surface and compress their shadows on hover, retain visible keyboard focus, and stop moving under reduced-motion preferences. No separate CSS weather animation competes with navigation.

## Accessibility

High-contrast text, semantic headings and lists, reduced-motion support, no essential information inside decorative cloud forms.
