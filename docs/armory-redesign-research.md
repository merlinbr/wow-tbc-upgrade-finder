# Armory redesign and 3D character research

Date: 2026-09-05. Research and proposed direction only; not an approved implementation specification. No application code or dependencies changed.

## Recommendation

Move toward a compact, character-centered equipment screen, with separate Gear, Stats, and read-only Talents views. Preserve this application's purpose: reviewing the exact imported simulation baseline and finding upgrades, not recreating a social armory website.

Treat the visual redesign and the 3D asset integration as separate decisions. An equipped TBC character renderer is technically credible, but no production-ready, permission-cleared, drop-in integration was established. The strongest technical candidate is the actual Wowhead/ZAM TBC pipeline, conditional on an authorized integration and working browser asset transport. Do not promise a generic npm wrapper will solve this.

User decision: the initial preview does not need to match the in-game face, hair, or other cosmetics. Use the imported race and equipment with a clearly labeled default/customized appearance; automatic Blizzard appearance lookup and a detailed cosmetic editor are not prerequisites. Exact appearance can be revisited separately. Renderer permission and asset-access gates still apply.

## Current application

- Svelte 5 and Vite in `ui-finder/`; Go loopback HTTP server in `cmd/wowsimcli/cmd/upgrade_server.go`; bundled `wowsims/tbc-new` simulator and database.
- Flow: individual-sim link → validated import → server-enriched armory → content/budget controls → local single-item DPS ranking → report. Imported settings remain the baseline while candidates are evaluated.
- `ui-finder/src/App.svelte` stacks ImportPanel, ArmoryView, RankingPanel, and ReportView.
- `ArmoryView.svelte` renders character facts, eight left slots, an empty center, nine right slots, then StatPanels. It currently hardcodes the level-70 label, appropriate to its TBC scope; this is not a separately imported level field.
- `GearSlot.svelte` already renders icons, quality, item/phase/set information, enchant, gems, and socket-bonus state. Icons come from ZAM, with local text fallbacks. Existing icon access does not establish permission or technical support for model assets.
- `app.css` caps the main surface at 1240 px and gives the center only 7rem. Gear cards repeat slot names, item IDs, no-enchant messages, and inactive socket-bonus text. Narrow screens collapse to one column.
- `types.go`, `armory.go`, and `handleImport` own the display data. Stats are calculated for the imported buffs, consumes, and talents and must keep the label **raid buffed (link settings)**. This redesign must not silently switch them to an unbuffed or live-armory snapshot.

### Runtime evidence

Started the existing `wowsimcli.exe` locally and imported `cmd/wowsimcli/cmd/upgrades/testdata/retribution_no_settings_link.txt` through the actual Chromium UI, at a 1568 × 1033 CSS viewport.

Observed:

- Human Retribution Paladin; 17 rendered slots, 16 equipped items.
- Armory panel: 1200 px wide, approximately 2152 px tall including stats.
- Gear grid: 1150 px wide, approximately 1362 px tall; tracks 503 / 112 / 503 px.
- Live `/api/import` returned character, digest, revisions, defaults, gear, stats, and derivedStats; no appearance or talents field.
- The fixture differs from the screenshots in some equipment and phase. It verifies the current layout and data flow, not the user's exact loadout.
- Temporary browser tabs and the inspection server were closed. No ranking run or project test suite was needed for this research.

## Proposed visual direction

### Composition

Use the references' strongest idea—compact gear around a central character—but avoid a bright full-screen city scene behind every line of text.

- Compact identity header: name, race/spec, professions, phase, and a clear Find upgrades action. Collapse the large import form after success, retaining Change import.
- Gear / Stats / Talents navigation. Keep ranking controls and progress accessible rather than burying the app's primary action below a long armory.
- Dark neutral stage with restrained faction/class color, a subtle radial light or subdued background confined to the character area.
- Real center space for a model; gear cards flanking it; Main Hand, Off Hand, and Ranged/Relic in a bottom equipment strip.
- Preserve all 17 supported slots. The imported equipment has no shirt or tabard fields; do not invent them to copy the reference's silhouette or symmetry.
- Explicit slot-based layout, rather than depending on array slices after relocating weapons. Preserve accessible reading order and slot labels.

### Gear cards

User refinement: use the compact WoWSims item presentation as the reference, combined with the central character composition above. This supersedes the earlier suggestion to reserve rarity colors mainly for borders.

- **Item level:** a readable numeric badge at the top-left of the equipment icon, not the item ID or content phase. Preserve room for gems along the bottom. The bundled database leaves `UIItem.ilvl` unset and stores the item level per scaling option; the implementation reads the equipped item's scaling bucket first and falls back to `ilvl`, rather than inferring it from phase.
- **Item name:** use the item's rarity color (epic purple, rare blue, etc.) on the name as requested, supported by icon quality treatment. Use a sufficiently dark, quiet background and verify text contrast; do not silently replace the requested rarity-colored names with uniformly white text.
- **Enchant:** a concise green effect line beneath the item name, e.g. `+34 Attack Power and +16 Hit Rating`, `+15 Strength`, or `+6 All Stats`, rather than the recipe/enchant name. Keep the full enchant identity available in accessible details. Proc effects such as `Mongoose` retain a recognizable effect name rather than misleadingly displaying their temporary proc as permanent stats. No enchant means no filler line.
- **Sockets and gems:** small sockets along the bottom edge of the item icon, with equipped gem artwork in socket order. Keep empty sockets visible and distinguish socket color from inserted gem color; provide the gem/socket details on hover and keyboard focus or tap. Items without sockets have no gem strip.
- **Details:** move item IDs, phase, complete stats, set/source information where available, and socket-bonus explanations into a focusable/clickable detail panel. Preserve information access rather than deleting it to save space.

The existing upstream implementation already provides a reuse path: `ui/core/components/gear_picker/gear_picker.tsx` renders the item-level overlay, rarity-colored name, enchant description, and gem strip. `ui/core/proto_utils/enchants.ts` resolves descriptions by enchant effect ID from the bundled `assets/enchants/descriptions.json`, falling back to the name. Verified entries include `3003: +34 Attack Power and +16 Hit Rating`, `684: +15 Strength`, `2661: +6 All Stats`, and `2673: Mongoose`. Reuse this data, not the old imperative UI or its hardcoded `/tbc/` asset route. Some descriptions remain names (e.g. Savagery): stat-only enchant labels must use their actual flat effects where the existing description is insufficient; never invent numeric effects for procs.

Hide irrelevant empty metadata, not meaningful missing enhancements. An item with no sockets should not announce an inactive socket bonus. Do not classify every unenchanted slot as a problem: enchant applicability, professions, and the imported baseline matter. Proposed alignment follows WoWSims: icon outside and text toward the center, mirrored on the right; this is layout direction, not authorization to implement.

A later preview of a ranked replacement could visually distinguish Baseline from Preview, while leaving simulator input untouched. This is an optional product idea, not an implementation requirement inferred from the screenshots.

### Stats and talents

Stats gets a separate, grouped view using existing server values. Retain raid-buffed assumptions visibly; do not infer caps or diagnoses without encounter-specific logic.

Talents can resemble the three-tree reference without a new talent service: expose the already-retained `Settings.Player.TalentsString`, then use the nine bundled class JSON files under `ui/core/talents/trees/`. They contain tree names, node positions, maximum ranks, spell IDs, prerequisite locations, and TBC background URLs. The existing decoder splits trees on `-`, reads ranks in array order, and treats missing digits as zero. Reuse that convention and metadata, not the entire older imperative simulator UI framework.

Start read-only so displayed talents remain the imported simulated build. Do not add a second-spec toggle, PvP tab, character claiming, or social controls without corresponding data and purpose. The JSON is not a complete offline spell-tooltip database; icons/descriptions/background rights and loading remain separate considerations.

### Interaction and fallback

- Explicit Activate 3D, rotate, reset, and pause controls; no scroll-wheel capture until the viewer is deliberately engaged.
- Respect reduced motion and pause when hidden; lazy-load the renderer so import/ranking is not dependent on it.
- On narrow screens, show a compact optional model panel and readable gear list rather than squeezing three desktop columns.
- Offline, WebGL failure, missing assets, or unavailable appearance must leave gear, stats, talents, and ranking usable. Clearly label an approximate/customized model; never pass it off as imported appearance.

## Three implementation approaches

| Approach | Benefits | Costs / limits | Position |
| --- | --- | --- | --- |
| Authorized TBC equipment renderer | Interactive character; potential local equipment preview; closest to reference | Provider rights, CORS/asset transport, display-ID mapping, version compatibility, lifecycle | Preferred 3D direction, conditional on feasibility gates |
| Static portrait or character image | Much simpler presentation; no WebGL | Not interactive; cannot show arbitrary replacement equipment; Blizzard image availability and credentials still need checking | Explicit fallback or separately accepted compromise, not a substitute for requested 3D |
| Independent renderer with legally usable converted assets | Control of hosting/runtime; potential offline operation | Asset extraction/conversion, character assembly, textures, animation, TBC coverage, rights; much larger maintenance burden | Not recommended for this app's current scope |

## 3D data and renderer findings

### Import data is insufficient for exact appearance

`proto/api.proto:21-95` contains race, class, equipment, talents, and simulation configuration but no body type, skin, face, or hair. `proto/ui.proto:49-101` and `GearSlotData` contain item IDs/icons but no item display/model IDs.

Keep these ID domains distinct: simulator race, game race, simulator slot, inventory type, item ID, item display ID, model-file ID, enchant effect ID, and enchant visual ID. For example, local `RaceHuman` is enum 5 (`proto/common.proto`), while the inspected ZAM Human metadata uses race 1. Directly forwarding enum numbers would select the wrong character.

Proposed data flow:

Imported baseline gear + explicitly mapped race → TBC display/inventory mapping → selected or separately fetched appearance → isolated viewer input.

The original import remains authoritative. A live Battle.net character may have different gear/talents and must not overwrite the simulation baseline. Visual settings should not alter its settings digest.

### Wowhead/ZAM: actual TBC pipeline, not a drop-in public SDK

Primary sources:

- [TBC dressing room](https://www.wowhead.com/tbc/dressing-room)
- [Paperdoll integration source](https://wow.zamimg.com/js/Paperdoll.js)
- [TBC viewer distribution](https://wow.zamimg.com/modelviewer/tbc/viewer/viewer.min.js)

The inspected integration uses inventory slots and **display IDs**, with optional enchant visual IDs. It normalizes main-hand equipment to slot 21, including two-hand weapons; it exposes equipment updates and calls `destroy()` when replacing viewers. This suggests a small Svelte lifecycle owner, not a full UI rewrite. Exact Svelte integration, transitive binary/texture loading, effects, and cleanup were not executed.

Read-only metadata checks succeeded:

| Equipment | Item → display ID | Evidence |
| --- | --- | --- |
| Furious Gizmatic Goggles | 32461 → 45779 | [TBC item XML](https://www.wowhead.com/tbc/item=32461&xml), [helmet metadata](https://wow.zamimg.com/modelviewer/tbc/meta/armor/1/45779.json) |
| Torch of the Damned | 32332 → 45350 | [TBC item XML](https://www.wowhead.com/tbc/item=32332&xml), [weapon metadata](https://wow.zamimg.com/modelviewer/tbc/meta/item/45350.json) |

[Human model 1](https://wow.zamimg.com/modelviewer/tbc/meta/character/1.json), [model 2](https://wow.zamimg.com/modelviewer/tbc/meta/character/2.json), and [customization metadata](https://wow.zamimg.com/modelviewer/tbc/meta/charactercustomization/1.json) also returned data in research. Equivalent Classic equipment paths returned 404; TBC and Classic are not interchangeable asset environments.

**Confirmed browser blocker:** requests for both equipment metadata URLs from the running local app failed. Chromium explicitly reported that the weapon response lacked `Access-Control-Allow-Origin`. Server-side/tool HTTP success therefore does not prove browser integration. Any approved transport must cover the complete metadata/model/animation/texture dependency graph, not just an item-display lookup.

**Permission gate:** [ZAM terms](https://corp.fanbyte.com/legal/terms), especially Sections 4 and 8, restrict copying/distribution and display/mirroring/framing without consent. No supported third-party SDK grant was established. Obtain explicit authorization or choose a supported, legally usable asset source; do not treat a proxy, scraped endpoint, or permissive wrapper license as permission. This is a dependency risk, not a legal opinion about a specific proposed agreement.

### Wrapper alternatives

- [Miorey/wow-model-viewer](https://github.com/Miorey/wow-model-viewer): useful reference, but its [entrypoint](https://github.com/Miorey/wow-model-viewer/blob/master/index.js) only branches between `classic` and live; passing `tbc` is not proof of TBC support. It relies on external ZAM globals/assets, requires display mappings, and includes WotLK-to-Retail assumptions. Installing it does not solve CORS, rights, or missing appearance.
- [JollyGrin/turtle-wow-model-viewer](https://github.com/JollyGrin/turtle-wow-model-viewer): independent Three.js approach, but targets 1.12/Turtle assets, takes extracted asset paths/texture configurations rather than imported item IDs, and does not establish TBC/Anniversary coverage. Its asset disclaimer is not a redistribution license.
- [Google model-viewer](https://modelviewer.dev/) loads assembled glTF/GLB models; it does not assemble WoW equipment from race/item IDs. [wow.export](https://github.com/Kruithne/wow.export) is a possible conversion tool, not a ready-made dynamic web armory backend. This route shifts the hard work into assets and assembly.

### Security and lifecycle design constraints

[INFERENCE/design] Mutable third-party JavaScript loaded into the app's page can access that page's data and local API. Prefer an authorized pinned distribution or a deliberately isolated renderer context; supply only necessary visual inputs, never credentials or the full imported configuration.

If an authorized integration requires Go-mediated asset loading, use fixed allowed origins/path patterns, bounded requests and sizes, redirect checks, and no arbitrary-URL proxy. Do not build transport to evade provider restrictions.

One viewer owner should handle initialization, equipment updates, pause, and destruction, including late asynchronous completion after unmount. Browser verification must demonstrate those behaviors rather than assume the library gets them right.

## Blizzard appearance and media: current official support

The current [Classic profile reference](https://community.developer.battle.net/documentation/world-of-warcraft-classic/profile-apis) explicitly lists `/appearance`, `/equipment`, `/character-media`, and `/specializations`. This was verified in its rendered documentation and [first-party backing JSON](https://community.developer.battle.net/api/pages/content/documentation/world-of-warcraft-classic/profile-apis.json), not inferred from Retail APIs.

The [official namespaces guide](https://community.developer.battle.net/documentation/world-of-warcraft-classic/guides/namespaces) and its [backing JSON](https://community.developer.battle.net/api/pages/content/documentation/world-of-warcraft-classic/guides/namespaces.json) explicitly identify Burning Crusade Classic Anniversary:

- `static-classicann-{region}`
- `dynamic-classicann-{region}`
- `profile-classicann-{region}`

The generic `profile-classic-*` default now identifies Mists progression, not Anniversary. Character data updates on logout.

Public named-character lookups use an application OAuth token; account discovery/protected profiles use user consent and `wow.profile`. See [official OAuth documentation](https://community.developer.battle.net/documentation/guides/using-oauth). A named public lookup does not require building a user-login system, but someone must provision developer credentials. A distributed local app must not embed a shared secret in its frontend or binary. Operator-owned credentials in the local Go process are a possible optional setup; a hosted credential service would be a larger product decision.

**Not verified:** an authenticated Anniversary appearance response, exact cosmetic identifiers, full-body media availability, or a usable item-display mapping. The media endpoint describes available images such as an avatar, not an interactive renderer or arbitrary-equipment render request. Even a successful full-body image cannot show a hypothetical local replacement item.

The current Classic game-data docs do not establish the Retail Item Appearance API as an Anniversary display-ID solution. Keep this as an experiment, not an assumed dependency.

An addon appearance sidecar is another possibility, but the existing WoWSims exporter/import/Player contract does not carry full cosmetics. Extending only an addon would lose that data through the current sim-link path. Do not change upstream protobufs merely to make a visual preview possible.

Any adopted API flow also needs to comply with [Blizzard developer API terms](https://www.blizzard.com/en-us/legal/a2989b50-5f16-43b1-abec-2ae17cc09dd6/blizzard-developer-api-terms-of-use), including credential handling, attribution, and applicable data-retention obligations.

## Suggested next decisions and validation

1. Resolved by the user: exact in-game face/hair/cosmetics are not required initially. The equipment presentation must include item-level icon badges, rarity-colored names, concise enchant effects, and visible sockets/gems as in WoWSims.
2. Approve a visual direction before implementation. A visual comparison can then explore subdued character-stage versus more cinematic treatment.
3. Resolve renderer/asset permission and supported transport before committing to a provider. Do not silently reduce 3D to a static image if this gate fails.
4. After approval and allowed access, prove a Human with the goggles and Torch in a real local browser: correct TBC/body variant, correct attachments, orbit/pause/reset, then all visible imported gear. Verify an empty off-hand, cloak/robe behavior, missing asset, offline mode, narrow screen, and mount/unmount cleanup. Enchant visuals require their own mapping; gems, jewelry, trinkets, and relics are not all visible model parts.
5. Defer exact-appearance lookup. If requested later, use operator-owned credentials and a known Anniversary character to inspect `/appearance`, `/character-media`, `/equipment`, and `/status` with the explicit Anniversary namespace. Compare returned cosmetics to renderer identifiers; never replace imported gear with live gear.
6. Implement the approved layout using the current focused Svelte components, preserve the Go simulation/ranking logic, and add only the data fields required by the selected design. Item level is now a required display field, not optional; expose it alongside the required enchant effect description and reuse existing socket/gem data. Expose talentsString for the proposed talents view.

## Scope and evidence limits

This research includes a real current-app import/visual inspection, primary-source review, successful metadata retrieval, and a real browser CORS failure. It does **not** include a working equipped 3D render, an authenticated Blizzard lookup, a licensed provider agreement, a mockup, or an approved implementation plan. Application files and dependencies are unchanged; this note is the sole deliverable file.
