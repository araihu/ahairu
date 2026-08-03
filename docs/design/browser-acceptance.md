# Storm Studio Browser Acceptance

Verified against the static build served from `http://127.0.0.1:4187/` on 2026-08-02.

## Responsive layout

| Viewport | Result |
| --- | --- |
| 1280×800 | Desktop topbar visible, mobile navigation hidden, four Goshtoso Card surfaces rendered, `scrollWidth == clientWidth == 1280`. |
| 768×900 | Desktop taxonomy remains visible, mobile navigation hidden, `scrollWidth == clientWidth == 768`. |
| 650×844 | Desktop Home/Libs/Apps/Blog taxonomy visible, mobile navigation hidden, `scrollWidth == clientWidth == 650`. |
| 390×844 | Desktop topbar hidden, bottom-left App Shell trigger visible, `scrollWidth == clientWidth == 390`. |

Automated Chromium coverage also checks the shared navigation boundary at 640, 641, 700, and 701 CSS pixels. Exactly one desktop/mobile project-navigation surface must be visible at every width.

## Enhanced navigation

- Local Goshtoso runtime loaded from `/assets/js/runtime/*` and `/assets/js/goshtoso.min.js`.
- Trigger opened the Goshtoso top Drawer and changed `aria-expanded` from `false` to `true`.
- Drawer title and close control resolve to the storm-paper foreground.
- Active locale resolved to `EN`, `PT`, and `ES` on the corresponding localized page.
- The scrolled EN proof shows Blog/WIP, GitHub, and the selected locale inside the Drawer.

## Native fallback

The built English document was served in a temporary browser fixture with every script tag removed before loading. No DOM mutation was used to reveal the fallback.

- Root never received `landing-shell-mobile-navigation-ready`.
- Enhanced path stayed hidden.
- Native `<details>` opened from its `<summary>` trigger.
- Active `EN` locale remained visible and selected.
- Layout remained exactly 390 CSS pixels wide with no horizontal overflow.

## Motion

- Browser regression tests verify the Goshtoso pressed Card changes whole-card transform and shadow while storm art moves independently.
- The same tests emulate `prefers-reduced-motion: reduce` at 1280×720 and 390×844 and require no atmospheric animation, no Card transform/transition, no art transform, and instantaneous App Shell/Drawer state changes.
- A theme-aware hero test loads only `storm-dark-v1.mp4` or `storm-light-v1.mp4`, verifies full hero coverage and filtered contrast, exercises scroll parallax, and confirms reduced-motion mode leaves the video paused and hidden.

## Theme backdrops

- Dark and light footage are H.264, 960×534, 20 fps, ten seconds, and contain no audio stream.
- Dark transfer size is 393,028 bytes; light transfer size is 451,771 bytes.
- Both use `preload="none"`; progressive enhancement skips playback for reduced-motion and data-saving users.
- Desktop and 390px browser captures show no horizontal overflow or navigation collision in either color scheme.

## Released component boundary

- Project surfaces compose `github.com/araihu/goshtoso/components/card` from Goshtoso v0.1.7.
- Responsive navigation composes `github.com/araihu/goshtoso-app-shells/landingshell` from App Shells v0.1.3.
- `GOWORK=off npm run check` passed against those public pins, including generated output, Go tests, browser design tests, worker tests, and Wrangler dry-run.

## Evidence

- `docs/design/evidence/araihu-desktop-1280.png`
- `docs/design/evidence/araihu-desktop-light-1280.png`
- `docs/design/evidence/araihu-tablet-768.png`
- `docs/design/evidence/araihu-boundary-650.png`
- `docs/design/evidence/araihu-mobile-closed-en.png`
- `docs/design/evidence/araihu-mobile-light-390.png`
- `docs/design/evidence/araihu-mobile-open-en.png`
- `docs/design/evidence/araihu-mobile-fallback-en.png`
- `docs/design/evidence/araihu-project-grid.png`
