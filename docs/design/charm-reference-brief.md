# Charm-Inspired Arai Hû Design Reference

## Primary reference

- Live reference: https://charm.land/
- Reference purpose: study information hierarchy, irregular project composition, tactile interaction, atmospheric motion, and responsive navigation.
- Constraint: do not copy Charm's artwork, palette, mascots, typography treatment, product copy, or exact layouts.

## Intended Arai Hû outcome

Arai Hû should feel like an expressive open-source studio whose products work together without losing their independent identities. The canonical motif is **Software for stormy weather.** The visual language is dark cloud, rain, mist, paper after rain, cobalt depth, and electric-lime signal.

## Reference patterns to translate

1. **Clear organization taxonomy**
   - Persistent desktop navigation separates Home, Libs, Apps, and Blog.
   - Arai Hû keeps Blog visibly marked WIP until it is functional.

2. **One dominant featured project**
   - Goshtoso receives a focused editorial section before the complete catalog.
   - The feature includes project-specific art, a concise promise, and one clear action.

3. **Uneven project grid**
   - Mix wide and narrow cards instead of repeating identical tiles.
   - Allow project lettering, marks, media, and split surfaces to cross internal visual regions.
   - Preserve one shared accessibility and full-card-link contract.

4. **Tactile cards**
   - Hover moves a card toward the page and compresses its shadow.
   - Active state moves slightly farther.
   - Project art can shift independently, but names and actions stay legible.

5. **A quieter secondary-project lane**
   - Goshtoso App Shells, Goshtoso Charts, and Muamba appear after the primary products.
   - Alternate solid, outlined, and open rows to avoid a mechanical card wall.

6. **Mission and community closure**
   - End with a concise open-source mission, an honest field-notes WIP surface, and a real GitHub conversation action.
   - Never present a disabled mailing preview as a functioning subscription form.

7. **Atmospheric background**
   - Translate Charm's ambient background role into Arai Hû storm weather: slow clouds, restrained rain, subtle grain, and rare light changes.
   - Atmosphere establishes identity without reducing text contrast or competing with navigation.

8. **Responsive floating navigation**
   - At small widths, replace the desktop link row with a fixed bottom-left trigger.
   - Open navigation as a top drawer with large touch targets, WIP state, GitHub action, and locale links.
   - Site fallback remains native and no-runtime; reusable App Shell composition should use Goshtoso Drawer after release.

## Reusable component boundary

- **Goshtoso Card:** custom media slot and optional physical pressed interaction.
- **Goshtoso Drawer:** left, right, top, and bottom directions with side-appropriate extent presets.
- **Goshtoso App Shells:** owns the breakpoint, trigger placement, and responsive navigation composition.
- **ActionGroup:** remains an action grouping primitive; it does not own floating layout policy.
- **Arai Hû site:** owns storm art direction, project hierarchy, localized copy, and section composition.

## Acceptance evidence

- Desktop proof at 1280px shows complete topbar, storm hero, dominant Goshtoso feature, uneven primary grid, secondary lane, mission, and community closure.
- Mobile proof at 390px shows the floating bottom-left trigger and open top drawer without horizontal overflow.
- Keyboard focus remains visible on navigation, cards, and actions.
- `prefers-reduced-motion: reduce` disables every ambient and interaction animation without hiding content.
- English, Brazilian Portuguese, and Spanish pages preserve hierarchy and status labels.
- Generated templ output, source templates, CSS, tests, and public assets remain synchronized.

## Review question

Does the current Git worktree deliver this reference translation as a coherent Arai Hû design, or does it merely imitate surface details while missing hierarchy, usability, brand distinction, or reusable-component boundaries?
