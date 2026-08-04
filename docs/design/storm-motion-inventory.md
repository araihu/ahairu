# Storm Motion Inventory

Every moving surface must convey atmosphere, hierarchy, or direct manipulation. Reduced-motion mode keeps state changes but removes interpolation and spatial movement.

| Surface | Normal behavior | Purpose | Reduced-motion behavior |
| --- | --- | --- | --- |
| `.storm-backdrop` | Muted theme-selected video loops while a bounded scroll offset adds shallow parallax | Put the storm footage behind the hero without exposing edges | Video is not loaded or played; CSS fallback remains |
| `.storm-video-filter` | Static dark/light gradient selected by the system color scheme | Preserve text contrast and bind footage to the brand palette | Unchanged |
| `.featured-demo video` | Six-second montage plays only while near the viewport | Show real Goshtoso and Goshtoso Charts usage | Poster remains; video is paused |
| `.openapi-stream` | OpenAPI YAML moves upward through the Manja card | Show the publishing input as working material | First YAML frame remains stationary |
| `.muamba-drop` | Small build artifacts fall at staggered speeds inside the Muamba thumbnail | Visualize vendoring remote files into a local build | Artifacts remain as a static scattered composition |
| `.signal-button` | Small lift and signal shadow | Primary-action affordance | No transition or transform |
| Goshtoso Card with `InteractionPressed` | Moves down through Tailwind's individual `translate` property; shadow compresses further when active | Shared physical whole-card affordance | Goshtoso disables transition and resets `translate` |
| `.project-art::before/::after` | Independent scale/translation | Layer project-specific art | No transition or transform |
| `.project-art-name` | Small horizontal shift | Tie lettering to card hover | No transition or transform |
| `.project-mark` | Project-specific rotation/scale | Give each project distinct rhythm | No transition or transform |
| `.more-row` | Surface/color interpolation | Identify the active secondary project | Instant color change |
| `.more-art` | Rotation/scale | Give quiet rows one expressive element | No transition or transform |
| App Shell `.landing-shell__mobile-trigger` plus `.storm-mobile-trigger` skin | Shadow compression and translation | Physical floating-button affordance | Instant state; no transform |
| Goshtoso Drawer backdrop | Opacity interpolation | Shift focus to navigation | Instant opacity |
| Goshtoso top Drawer panel | Slides from top | Reveal responsive navigation | Instant open/close |
| App Shell native `<details>` fallback | Opens without JavaScript | Preserve navigation when runtime is unavailable | Instant open/close |

Automated browser coverage emulates `prefers-reduced-motion: reduce` at 1280×720 and 390×844, then checks computed animation, transition, opacity, and transform states before and after hover/open interactions.
