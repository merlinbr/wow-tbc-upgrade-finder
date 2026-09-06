# Reference sites: actual 3D rendering pipelines

Inspected 2026-09-06. Follow-up to [earlier research](armory-redesign-research.md). Research only; this is the sole file edited. No additional agents, application changes, dependency installs, or remote JavaScript execution in a shell. Public JS was read as inert text; the sites executed normally in Chromium. **Verified** means observed source, HTTP response, or browser behavior; **inference** and remaining uncertainties are called out separately.

## Findings at a glance

| | Classic WoW Armory | Sixty Upgrades |
| --- | --- | --- |
| Renderer | ZAM/Wowhead `ZamModelViewer`, extended by site-hosted ES modules | Three.js/WebGL with site-bundled WoW assembly, materials, textures, skeleton and animation code |
| Asset transport | Same-origin `/proxy/https://wow.zamimg.com/modelviewer/tbc/…` | Direct `https://cdn.sixtyupgrades.com/model-viewer/…` |
| Model representation | ZAM metadata JSON, `.m2`, `.skin`, WebP textures | Converted geometry/animation JSON and PNG textures |
| Equipment input | Inventory-slot/display-ID pairs embedded in character-page HTML | Slot-indexed full item objects containing `displayInfo` from the site's GraphQL data layer |
| Appearance input | Numeric race/body and cosmetic indexes embedded in HTML, translated through ZAM customization metadata | Race/body strings and five cosmetic indexes; model JSON supplies variations |
| Specific reference verified? | Equipped character visibly rendered; asset requests observed | Requested URL redirected to `/tbc/characters`; specific set/appearance not obtained |

Evidence: [Classic character page](https://classicwowarmory.com/character/EU/thunderstrike/tastynjuicy?game_version=classic), [Classic wrapper entrypoint](https://classicwowarmory.com/wow-model-viewer/index.js), [Sixty main bundle](https://sixtyupgrades.com/static/js/main.ce82b686.chunk.js), [Sixty dependency bundle](https://sixtyupgrades.com/static/js/2.ece08ad6.chunk.js). These are different pipelines; the fact that both link item tooltips to Wowhead does not identify their renderer.

## Classic WoW Armory — verified

### Provider and version selection

The page loads [ZAM's **live** viewer script](https://wow.zamimg.com/modelviewer/live/viewer/viewer.min.js), then imports the site's `generateModels`. Its [viewer class](https://classicwowarmory.com/wow-model-viewer/wow_model_viewer.js) explicitly extends `ZamModelViewer`.

The page sets `viewerEnv="tbc"`. In `resolveEnvConfig`, the deployed [entrypoint](https://classicwowarmory.com/wow-model-viewer/index.js) selects a **TBC content path**, but takes the default `renderEnv: live` branch and passes `hd: true`. Only `classic1x` sets Classic renderer flags. Its separate `classic` branch selects live assets and a WotLK-to-Retail display bridge; that bridge is **not enabled for this TBC page**. Thus “live viewer script + TBC data” is the observed configuration, not the previously investigated TBC viewer distribution.

### Exact page inputs

The inspected [HTML's inline module](https://classicwowarmory.com/character/EU/thunderstrike/tastynjuicy?game_version=classic) supplied:

```json
{"race":1,"gender":1,"skin":7,"face":3,"hairStyle":9,"hairColor":5,"facialStyle":1,"sheathed":false}
```

Its `displayItems` was:

```json
[[1,45779],[3,46353],[5,42306],[6,42651],[7,37449],[8,45729],[9,45399],[10,43549],[16,38976],[21,45350]]
```

The page assigns these to `characterModel.items`, calls `generateModels(1, '#model_3d', characterModel, viewerEnv)`, then conditionally calls `setZoom(-7)`. Visible item links identify goggles item **32461**, represented by head display **45779**, and Torch item **32332**, represented by main-hand display **45350**. No browser-side item-ID lookup is needed by this initialization. The HTML does not reveal the server's mapping database or the upstream origin of these cosmetic values.

[Customization code](https://classicwowarmory.com/wow-model-viewer/character_modeling.js) computes model ID as `race*2-1+gender` (here **2**, type **16**), fetches `meta/charactercustomization/2.json`, matches named appearance options, and turns each supplied index into the corresponding choice ID. Inputs such as `hairStyle:9` are **indexes into Choices**, not directly the provider's choice IDs. Unmapped options use their first choice.

The source also contains transmog-aware equipment helpers, but this page passes its already prepared array directly. The array has no enchant visual field and omits tabard despite a visible Tabard of Flame equipment card. Therefore the reference is not proof of complete equipment or Mongoose-effect fidelity. The wrapper's `updateItemViewer(slot,displayId,enchant)` passes `visual` to `setItems`, but explicitly describes enchant support as experimental/untested. [Equipment/customization helpers](https://classicwowarmory.com/wow-model-viewer/character_modeling.js), [viewer methods](https://classicwowarmory.com/wow-model-viewer/wow_model_viewer.js).

### Transport observed in Chromium

The [entrypoint](https://classicwowarmory.com/wow-model-viewer/index.js) sets both `window.CONTENT_PATH` and the viewer's `contentPath` to its same-origin proxy. Browser resource inventory contained all these representative requests:

- [Character customization JSON](https://classicwowarmory.com/proxy/https://wow.zamimg.com/modelviewer/tbc/meta/charactercustomization/2.json) and [character metadata](https://classicwowarmory.com/proxy/https://wow.zamimg.com/modelviewer/tbc/meta/character/2.json).
- [Goggles metadata](https://classicwowarmory.com/proxy/https://wow.zamimg.com/modelviewer/tbc/meta/armor/1/45779.json) and [Torch metadata](https://classicwowarmory.com/proxy/https://wow.zamimg.com/modelviewer/tbc/meta/item/45350.json).
- [Character M2](https://classicwowarmory.com/proxy/https://wow.zamimg.com/modelviewer/tbc/m2/119563.m2), [skin mesh](https://classicwowarmory.com/proxy/https://wow.zamimg.com/modelviewer/tbc/skin/470980.skin), and [WebP texture](https://classicwowarmory.com/proxy/https://wow.zamimg.com/modelviewer/tbc/textures/119894.webp).

The equipped character and weapon were visibly rendered. The “unlock” control merely changes pointer-event locking; loading occurs beforehand. **Inference:** this proxy explains how this site avoids direct browser-to-ZAM model CORS failures. Backend caching, validation, forwarding implementation, authorization arrangements, and suitability as another application's proxy remain unknown. No cross-origin use of its proxy was tested.

## Sixty Upgrades — verified source; specific character unavailable

The [requested set URL](https://sixtyupgrades.com/tbc/character/tYLuXDWimtUyafpgbagyW1/set/eu-thunderstrike-tastynjuicy) served the public app shell, then Chromium ended at `https://sixtyupgrades.com/tbc/characters`. It did not show the requested character. Cause unknown; this is not evidence that the set is deleted or private.

### Renderer and asset format

The [main bundle](https://sixtyupgrades.com/static/js/main.ce82b686.chunk.js) imports module 4 as `dA`, constructs `dA.eb` as the WebGL renderer, creates skinned buffer geometry/skeletons, and supplies custom shader materials. The [dependency bundle](https://sixtyupgrades.com/static/js/2.ece08ad6.chunk.js) maps `eb` to its Three.js WebGL renderer implementation, containing `THREE.WebGLRenderer` diagnostics. The WoW-specific assembly implementation lives in the site's main bundle; no ZAM viewer call appears in the inspected rendering path. Original authorship or possible upstream derivation of that assembly code is not established.

Useful main-bundle search anchors: `Uv` / `Wv` (model loader, around character offset 554511), `loadModel` (579796), `findAttachment`, `getCurrentColorId`, and `Tv` / `Dv` (texture composition, 526015). Offsets are zero-based decoded string positions, not line numbers.

- Models: `/model-viewer/model/character/{race}{gender}.json`; equipment attachments: `/model-viewer/model/item/{modelId}.json`.
- Additional animations: `/model-viewer/model/character/{race}{gender}_{animation}.json`, loaded on demand; some animations are already in model JSON.
- Textures: `/model-viewer/texture/{character|item}/{fileDataID}.png`. Body composition fetches PNG blobs, creates object URLs, draws regions onto a 2D canvas, and uploads the composed texture to Three.js. Attached models use a texture loader. These are source-observed paths, not a GLB/glTF viewer contract.

All use `https://cdn.sixtyupgrades.com`. [Human female JSON](https://cdn.sixtyupgrades.com/model-viewer/model/character/humanfemale.json) returned HTTP 200 and contains vertices, normals, UVs, materials, skin, bones, weights, attachments, animations, and variations. That JSON and [one character PNG](https://cdn.sixtyupgrades.com/model-viewer/texture/character/119894.png) returned `Access-Control-Allow-Origin: *` when requested with `Origin: http://localhost:3000`; the JSON did so with the site's own Origin too (`Vary: Origin`). Without Origin, the initial JSON response omitted ACAO. **Limit:** these HTTP header probes establish sampled transport behavior, not a full equipped render from localhost.

### Items and cosmetics

The main bundle's `FullItem` GraphQL fragment requests `displayInfo`: `displayInfoId`, models with model/race/gender/class/position IDs, model textures, body texture component sections with male/female texture IDs, geoset groups, helmet visibility, flags and sheathe type. The renderer receives the **full item objects**, selects attachment models by race/body/position, and loads their `modelId`; forwarding a WoW item ID or even one display ID alone is insufficient. Internal equipment indexes include back **15**, main hand **16**, off hand **17**, unlike the Classic Armory viewer's inventory types. [Main bundle](https://sixtyupgrades.com/static/js/main.ce82b686.chunk.js).

Character GraphQL fragments request `race`, `gender` and `variations` containing `skinColor`, `face`, `hair`, `hairColor`, `facialHair`. UI controls and randomization use numeric indexes; model variations determine geosets and texture selection. Race/body strings, e.g. `human` and `female`, also select the model URL. These are saved/customizable character properties; the inspected source does not prove this set's exact cosmetic values or how they originated. [Main bundle](https://sixtyupgrades.com/static/js/main.ce82b686.chunk.js), [model variations](https://cdn.sixtyupgrades.com/model-viewer/model/character/humanfemale.json).

The TBC client uses [the site's GraphQL endpoint](https://api.sixtyupgrades.com/tbc/graphql). Its `Wo` signing helper obtains Cognito credentials and signs requests for AWS `execute-api`; this is not an established anonymous third-party lookup API. No credentials were extracted or signed requests replayed. Source methods support equipment/appearance changes, helm/cloak/ranged toggles, sheathing, animations, PNG export and disposal; their presence does not constitute a public SDK. [Main bundle](https://sixtyupgrades.com/static/js/main.ce82b686.chunk.js).

## Integration options and unresolved dependencies

1. **ZAM-based integration:** the Classic Armory modules provide a concrete technical adapter example, including TBC path selection, display-pair input and cosmetic translation. An independently configured authorized asset transport could use the same architecture. The [Miorey wrapper README](https://github.com/Miorey/wow-model-viewer/blob/master/README.md) documents matching APIs and dependencies; common lineage is an inference, and its package is not verified equivalent to this site's modified TBC configuration. Neither public source nor the observed proxy establishes permission to reuse ZAM assets or depend on Classic Armory's service.
2. **Sixty-style independent assembly:** technically demonstrated by public source and converted assets; requires a compatible renderer, item-display metadata, converted model/animation/texture inventory and appearance mapping. No supported downloadable SDK, embedding contract, public renderer repository, or third-party data-access grant was established. The site's [Use License](https://sixtyupgrades.com/tbc/terms), also embedded in the main bundle, restricts copying/modification, public display and mirroring. Wildcard CORS is not a reuse license. A provider arrangement or separately usable assets/code remains necessary.
3. **External links:** ordinary links to reference character pages are available, but do not supply a locally controlled arbitrary-equipment preview. No documented iframe/postMessage integration was found in the inspected material; iframe feasibility was not tested.

For architecture planning, preserve provider-specific slot/race/body mappings and keep item ID, display ID, model-file ID and cosmetic-choice ID distinct. Exact imported appearance remains optional as already decided. This research establishes two real technical patterns; it does not establish an approved provider, reusable complete asset catalogue, exhaustive effects/gear fidelity, or a working integration inside this application.
