# Equipped 3D character preview — proposed implementation plan

Date: 2026-09-06. Research and planning requested; the 3D feature is not implemented.

## Outcome

Replace the center Character preview placeholder with an interactive, equipped
TBC character like the cyan-marked area in the supplied ClassicWoWArmory and
SixtyUpgrades screenshots. Keep the existing gear columns and weapon strip.
The character should fill the stage, with its feet and equipped weapon visible,
against a quiet background. Retain the app's current gear, stats, talents,
import, and upgrade-ranking behavior.

This is a new rendering feature, not a refactor of an existing renderer.
`ui-finder/src/lib/CharacterStage.svelte` currently contains only a backdrop,
placeholder text, and a disabled button. `ArmoryView.svelte` already places it
between the gear columns, but passes only race/class/spec, not equipment.

## Evidence and assumptions

- Current application contracts: `cmd/wowsimcli/cmd/upgrades/types.go`,
  `proto/api.proto`, `proto/ui.proto`, and `proto/common.proto`.
- Earlier experiments and data-access findings: `docs/armory-redesign-research.md`.
- Reference-site investigation: `docs/3d-reference-sites-research.md`.
- The screenshots establish the desired composition, not a reusable renderer
  API or proof that another site's assets can be embedded here.
- The imported configuration supplies race and equipped item IDs, but lacks
  body type, face, hair, skin, and item display IDs. Shirt and tabard are not
  among the 17 imported slots. Do not invent a tabard to reproduce the examples.
- The earlier research records the user's acceptance of a default/customizable
  appearance for the initial version. Continue that direction: imported race
  and gear, explicit body preset selection, and a label such as
  "Default appearance · imported gear". Exact in-game cosmetics are a later
  feature; the female Human in the screenshot is a useful acceptance fixture,
  not a reason to infer all characters' appearance.

## Recommended technical direction

Start with a small feasibility experiment using a TBC-capable equipped-character
renderer. The Wowhead/ZAM TBC pipeline is a candidate, subject to confirming an
appropriate integration/distribution arrangement and a working asset path from
our actual localhost origin. A successful JSON request from Go or curl does not
prove a browser can load meshes, textures, and animations. The prior research
recorded browser CORS failure; reproduce and resolve the actual integration
contract before promising a production delivery date.

Do not adopt an npm wrapper merely because it displays a WoW character. The
[Miorey wrapper entrypoint](https://raw.githubusercontent.com/Miorey/wow-model-viewer/master/index.js)
has explicit classic/live branching and relies on external globals. Its
[README](https://github.com/Miorey/wow-model-viewer) describes WotLK/Retail and
external asset transport. TBC compatibility remains something to prove.

Generic [model-viewer](https://modelviewer.dev/docs/index.html) can present a
prepared glTF/GLB asset, but the app would still need a WoW-specific character
assembly/export pipeline. A desktop alternative,
[WMVx](https://github.com/Frostshake/WMVx), lists TBC 2.4.3 and FBX export, with
known TBC particle/texture-animation limitations. It is an export-tool option,
not a browser component or proof of Anniversary compatibility. Building and
maintaining our own conversion service is a substantially larger alternative.

## Proposed boundaries

Use a small viewer integration with three responsibilities:

1. **Visual input mapping:** imported race + visible gear + local appearance
   choices become a renderer-neutral description. Keep simulator enums distinct
   from game race IDs, inventory slots, item display IDs, and model IDs.
2. **Visual asset resolution:** resolve versioned TBC item/display/attachment
   metadata through the selected provider. Report missing visual mappings
   separately from invalid simulator equipment.
3. **Viewer lifecycle:** mount, update equipment, resize, rotate/reset, pause,
   and destroy. Hide provider globals and resource ownership inside one adapter.

Proposed files, subject to the experiment:

| File | Responsibility |
| --- | --- |
| `ui-finder/src/lib/CharacterStage.svelte` | Stage presentation, activation, loading/error/retry, appearance controls |
| `ui-finder/src/lib/CharacterViewer.svelte` | Svelte lifecycle, canvas/container, resize and visibility observers |
| `ui-finder/src/lib/characterPreview.js` | Pure mapping of imported race/gear into visual inputs |
| `ui-finder/src/lib/characterViewerAdapter.js` | Provider-specific creation, updates, controls, disposal |
| `ui-finder/src/lib/ArmoryView.svelte` | Pass imported gear and identity to the stage |
| `ui-finder/src/app.css` | Stage dimensions, controls, small-screen composition |
| `cmd/wowsimcli/cmd/upgrade_visuals.go` | Only if selected transport needs a local visual-metadata endpoint |

Keep preview settings local to the preview. Do not add cosmetic selections to
the simulator Player protobuf, ranking request, settings digest, or assumptions
fingerprint. Import success and ranking must not await renderer/network work.
Only necessary visual fields should reach an external provider; no full export
link, encounter, buffs, report, or credentials.

## Delivery sequence and acceptance criteria

### 1. Prove one equipped TBC character

Use an isolated development harness before changing the production preview.
Resolve the chosen renderer's initialization, asset origin, and version contract.
Start with a Human body preset, Furious Gizmatic Goggles (item 32461), and Torch
of the Damned (item 32332). The prior note's display IDs are sample observations,
not a substitute for versioned mappings for the whole catalog.

Require a visible render in Chromium from the actual loopback app origin.
Verify helmet geometry, weapon attachment, texture/animation loading, framing,
rotation, resize, and clean destruction. Save a screenshot and a concise account
of the necessary request origins and files. If provider transport or distribution
is unresolved, report that exact dependency; do not substitute a static picture
and declare the 3D experiment successful.

### 2. Resolve the imported equipment

Implement explicit race and slot mappings and a versioned item-visual resolver.
Use the current imported baseline, including visible armor, cloak, and supported
weapons. Handle robe/chest inventory differences, two-handed weapons, empty
off-hand, shield versus held item, and ranged weapon versus invisible relic.
Jewelry and gems need not produce model attachments; do not create phantom parts.
Weapon enchant effect IDs need a separate visual mapping; omit unsupported
effects with an honest capability note rather than using effect IDs as model IDs.

Verify mappings for every supported race and body variant against the selected
renderer. Unknown race/missing display IDs must yield a clear partial/unavailable
preview, never a different silently chosen race or an import error. Batch lookups
and bounded caching belong here if the provider contract permits them.

### 3. Integrate the center stage

Enable Activate 3D when the integration is available. Lazy-load on activation;
show explicit loading, ready, partial, unavailable, and retry states. Once active,
the character takes the place of the placeholder art. Provide drag rotation,
accessible rotate/reset controls, and pause; keep page scrolling usable.

Pass gear into the stage. On re-import, update or recreate the viewer from the
new baseline and reject stale asynchronous results. Reset incompatible appearance
options when race changes. Gear-tab unmount and document visibility changes must
stop unnecessary rendering; destruction must release GPU resources and listeners.
Respect reduced motion and cap device-pixel ratio based on measured performance.

Use explicit stage height and camera fitting so the full body, cloak, and weapon
remain visible. The current center is 16–22rem wide and at most 26rem minimum
height; select the final dimensions from the equipped fixture, not a generic cube.
On small screens retain a compact optional preview and the existing readable gear
list. Check tooltips above the canvas and controls without introducing input traps.

### 4. Verify and package

Keep normal CI deterministic: use an adapter fake and local test assets for
loading, failure, retries, race/gear changes, and cleanup. Add mapping tests for
the ID boundaries and attachment cases above. Exercise a true touch browser
context for preview controls as well as keyboard and desktop pointer input.

Perform a separate real-provider browser smoke check for the sample Human,
a non-Human, both body variants, robe/cloak, off-hand, and failed asset. Test slow
network, offline mode, WebGL unavailable/context loss, tab switching, re-import
during load, and ranking while the preview is active. Record actual frame/load
measurements before deciding whether performance is acceptable.

Run UI unit/E2E tests, build the embedded Vite bundle, restart Go, and check the
embedded app. Run the focused Go tests if a metadata/asset endpoint is added.
Document external asset requirements, supported visual effects, appearance
defaults, and reproducible provider smoke checks in `docs/upgrade-finder.md`.

## Optional exact appearance later

If requested, investigate a named-character Blizzard appearance lookup or an
explicit appearance import. Blizzard's current
[Classic profile documentation](https://community.developer.battle.net/documentation/world-of-warcraft-classic/profile-apis)
lists character appearance and character-media endpoints. Its
[namespaces guide](https://community.developer.battle.net/documentation/world-of-warcraft-classic/guides/namespaces)
lists `profile-classicann-{region}` for Burning Crusade Classic Anniversary.
Both were rechecked through their first-party documentation JSON on 2026-09-06.
No authenticated response or renderer-compatible cosmetic mapping has been
demonstrated in this task. Region/realm/name must be supplied
explicitly because a simulation name need not identify a live character.

Use that data only for cosmetics. The character's current live gear can differ
from the imported simulation and must never replace it. A full-body media image
would remain a static fallback, not arbitrary equipped 3D. Any credentials belong
in an appropriate backend configuration, not the browser or distributed binary.

## Ready-to-start decision

The next implementation milestone should be the equipped-character feasibility
experiment. The success criterion is a real, correctly equipped TBC character
rendered from our app origin with a known asset/distribution path. The Svelte
integration becomes predictable once that is demonstrated. Exact cosmetics,
shirt/tabard inputs, and ranked-upgrade visual previews can follow separately.
