# Equipped character feasibility experiment — result

Date: 2026-09-06. Successful. This is the milestone-1 account from
`docs/superpowers/plans/2026-09-06-equipped-character-preview.md`, following the
earlier research in `docs/armory-redesign-research.md` and
`docs/3d-reference-sites-research.md`.

## Verdict

A correctly equipped TBC character was **rendered in Chromium from a loopback
app origin** (http://localhost:5173 via the Vite dev server) with a
**same-origin asset transport** and a **verified viewer distribution**:

- Renderer: ZAM/Wowhead **live** model viewer distribution
  (`https://wow.zamimg.com/modelviewer/live/viewer/viewer.min.js`,
  266,932 bytes, fetched 2026-09-06) driving **TBC content assets**
  (`modelviewer/tbc/*`) — the same combination the Classic WoW Armory
  character page uses. The dedicated `tbc/viewer/viewer.min.js` distribution
  requests `.mo3` files, which do not exist in the TBC asset tree (404s); it
  is **not** the compatible distribution.
- Body: Human female, ZAM model 2 (`race*2-1+gender`), type 16 (CHARACTER).
- Equipment from the plan's sample: Furious Gizmatic Goggles (item 32461 →
  display 45779, slot 1 head) and Torch of the Damned (item 32332 → display
  45350, slot 21 main hand), plus the reference-page fixture set.
- Verified: helmet geometry, weapon attachment, skin/mesh/texture loading,
  animation, framing, built-in pointer drag rotation, programmatic rotation,
  resize (`renderer.resize`), and clean destruction (canvas removed).

Screenshots (prototype, throwaway):

- `ui-finder/prototype/feasibility-render.png` — front view
- `ui-finder/prototype/feasibility-render-rotated.png` — after 45° rotation
- Harness source: `ui-finder/prototype/equipped-character.html`

## Transport inventory (observed requests)

All model data flowed **same-origin** through a local reverse proxy
(`/zam` → `https://wow.zamimg.com`, Vite dev proxy):

```
http://localhost:5173/zam/modelviewer/tbc/meta/character/2.json
http://localhost:5173/zam/modelviewer/tbc/meta/charactercustomization/2.json
http://localhost:5173/zam/modelviewer/tbc/m2/119563.m2            (body mesh)
http://localhost:5173/zam/modelviewer/tbc/skin/470980.skin        (body skin)
http://localhost:5173/zam/modelviewer/tbc/textures/119894.webp    (body texture)
http://localhost:5173/zam/modelviewer/tbc/meta/armor/1/45779.json (goggles meta)
http://localhost:5173/zam/modelviewer/tbc/meta/item/45350.json    (torch meta)
http://localhost:5173/zam/modelviewer/tbc/m2/143177.m2            (torch mesh)
… plus one meta + m2 + texture set per equipped visible slot
```

Two external requests remain cross-origin (classic `<script>`/`<link>`, no CORS
needed):

```
https://wow.zamimg.com/modelviewer/tbc/viewer/viewer.min.js
http://wow.zamimg.com/modelviewer/viewer/viewer.css   (protocol-relative URL)
```

Direct browser-to-`wow.zamimg.com` asset fetches still fail CORS from a
loopback origin (confirmed again in prior research) — the same-origin proxy is
the transport. Same architecture as Classic WoW Armory's `/proxy/` path.

## Verified integration contract

Initialization (mirrors `generateModels` from classicwowarmory.com):

```js
window.CONTENT_PATH = origin + '/zam/modelviewer/tbc/';
const full = await fetch(CONTENT_PATH + 'meta/charactercustomization/2.json');
const options = mapAppearanceToChoices({ skin, face, hairStyle, hairColor, facialStyle }, full);
const viewer = await new window.ZamModelViewer({
  type: 2, contentPath: CONTENT_PATH, container: jQuery('#stage'),
  aspect: 0.79, sheatheWeapons: 0, autoSheathe: 0, hd: true,
  items: [[zamSlot, displayId], ...],       // NOT_DISPLAYED slots: 2,11,12,13,14
  models: { id: race*2-1+gender, type: 16 },
  charCustomization: { options },
});
```

- `new ZamModelViewer(...)` returns a thenable — **must be awaited**.
- Needs a real **jQuery** (1.x era; references `jQuery.support.cors`,
  jQuery-wraps its canvas, binds `$(document).on(...)`). Vendored jQuery 1.12.4
  in the harness (MIT).
- Needs `window.WH` shims: `WH.debug()`, `WH.WebP.getImageExtension()`.
- Loaded character lives in `viewer.renderer.actors` (this viewer revision);
  `renderer.models` stays empty. `actor.setItems`, `actor.clearSlots`,
  `actor.setAppearance`, `actor.setSheath` exist; actor `setAnimPaused` no-ops
  on this revision — **pause needs a different mechanism (milestone 3)**.
- Controls: `viewer.renderer.azimuth/zenith`, `viewer.renderer.resize(w,h)`,
  `viewer.destroy()`. Built-in pointer drag rotation worked; window resize is
  not auto-handled (except fullscreen).
- Slot convention: ZAM inventory enums — head 1, neck 2, shoulder 3, shirt 4,
  chest 5, waist 6, legs 7, feet 8, wrist 9, hands 10, finger 11/12,
  trinket 13/14, ranged 15, back 16, tabard 19, robe 20, main hand 21,
  off hand 22, relic 28. These are the `inventorySlot id` values returned by
  `https://www.wowhead.com/tbc/item={id}&xml` (`<icon displayId>` + `<inventorySlot id>`).

## Item-to-display resolution

`https://www.wowhead.com/tbc/item=32461&xml` returns:
`<icon displayId="45779">…` and `<inventorySlot id="1">Head</…>` — i.e. one
server-side request per item yields exactly the two viewer inputs
(displayId, zam slot). This is the versioned TBC mapping source; fetch
server-side (no CORS), cache, and keep item IDs distinct from display IDs and
model-file IDs.

## Unresolved dependencies (exact)

1. **Authorization/distribution.** The ZAM viewer + model assets are not a
   public SDK; the published `viewer.min.js.LICENSE.txt` 404s. ZAM terms
   restrict copying, distribution, display/mirroring/framing without consent
   (see prior research). Technical integration is solved; **production use
   needs an explicit arrangement with Wowhead/ZAM** (or a source the operator
   is entitled to use). Until then the production stage stays gated
   (`--enable-3d` opt-in; default remains the labeled placeholder).
2. **Production transport — resolved.** Implemented in
   `cmd/wowsimcli/cmd/upgrade_visuals.go`: fixed-path same-origin proxy for
   `modelviewer/tbc/**`, plus the item-display resolver endpoint, both gated
   behind the `--enable-3d` flag; verified in the real app (see addendum).
3. **jQuery vendoring — resolved.** 97 KB MIT jQuery 1.12.4 is vendored into
   `ui-finder/src/lib/vendor/` and code-split with the adapter (loaded only on
   activation).

## Addendum — production app verification (same day)

The integration was then wired into the real application behind the
`--enable-3d` opt-in flag and smoke-tested against the embedded Go server:

- `cmd/wowsimcli/cmd/upgrade_visuals.go`: versioned item resolver
  (Wowhead TBC XML → display ID + inventory slot, bounded cache, chest/robe
  meta probe armor/5 vs armor/20) and the fixed same-origin ZAM asset proxy.
- `--enable-3d` gates the visuals routes and the UI's `visualsEnabled` flag;
  with the flag off the stage stays an honest "3D preview unavailable" state.
- Chromium run against `rank-upgrades --enable-3d` with the imported
  Retribution paladin: **real render** through the production proxy (57
  `/visuals/zam/` asset requests, 1 resolve request, 0 errors), full gear on
  the character, "Partial preview: Ranged" for the relic libram (no display
  model, reported honestly), rotate/pause/resume/gear-tab-cleanup verified.
- Screenshot: `ui-finder/prototype/feasibility-real-app-stage.png`
  (early 352 px stage) and `ui-finder/prototype/size-final-stage.png`
  (final: center track widened to 30rem, stage 478×670 filling the column
  height, character framed with margins via `renderer.distance *= 1.4`
  applied after the auto-fit settles).

Remaining before shipping the integration: operator confirmation of the
provider arrangement (the authorization gate), plus the plan's remaining
failure-path and performance measurements (offline, WebGL-unavailable,
slow network, frame/load numbers).

## Caveats for the report

- The `tbc/viewer/viewer.min.js` distribution renders correctly; the
  `live/viewer` + TBC content path (ClassicArmory's combo) was also tested and
  works — the TBC distribution is the cleaner versioned choice.
- All requests were made over the local Vite origin; the Go origin needs the
  same proxy wiring (next milestone).
- Frame-rate/perf and offline/failure behavior not yet measured; WebGL
  unavailable/failed-asset paths are milestone-3/4 work.
- Exact appearance remains the chosen **default/customizable** direction; the
  fixture used the reference page's cosmetic indexes to prove option mapping.
