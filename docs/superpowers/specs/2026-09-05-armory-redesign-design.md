# Armory redesign design

## Status

Approved direction 2026-09-05. This spec amends the layout and presentation
scope of `2026-08-29-svelte-armory-design.md` for the armory review view.
Ranking logic, job lifecycle, import validation, and the simulated baseline
are unchanged.

Research and feasibility evidence: `docs/armory-redesign-research.md`.

## Goal

After a successful import, show a character-centered armory: a compact
identity header, WoWSims-style item cards flanking a central character stage,
a weapons strip beneath the stage, and separate Gear / Stats / Talents views.
Each equipped item card shows the item-level badge, rarity-colored name,
enchant effect line, and sockets with gems. The 3D model itself ships as a
labeled placeholder behind a verified integration gate.

## Decisions

| Area | Decision | Reason |
| --- | --- | --- |
| Framework | Existing Svelte 5 + Vite; restructure `ArmoryView` into view components | No new runtime dependencies. |
| Layout | Two gear columns around a central character stage; Main Hand / Off Hand / Ranged strip under the stage; compact header; Gear / Stats / Talents tabs | Approved direction; strongest idea of the references. |
| Item card | Ilvl badge top-left of the icon; item name in rarity color; green enchant effect line; sockets with gem artwork along the icon bottom; details panel for ID, phase, full stats, set, socket bonus | User refinement for WoWSims item presentation. |
| Enchant text | Effect description resolved by effect ID from the bundled `assets/enchants/descriptions.json`, falling back to the enchant name; proc-only entries keep their name | Matches the WoWSims item display; never invents proc stats. |
| Data source | Server keeps owning all item data; import response gains ilvl, talents string, and enchant description | Browser remains a renderer. |
| Talents view | Read-only three-tree rendering from the bundled `ui/core/talents/trees/*.json` plus the exposed imported talents string | Shows the simulated build without a new talent service. |
| 3D model | Character stage renders bundled class/spec backdrop art and a labeled placeholder now; viewer integration is gated on renderer permission and verified browser asset access | Browser CORS failure is confirmed; provider terms require authorization. |
| Appearance | Imported race plus default or user-selected cosmetics, labeled as a customized preview | User decision: no exact face/hair match in the first attempt. |

## Non-goals

- Exact appearance lookup or any Blizzard API integration now.
- Shipping a 3D viewer before the permission and browser asset-access gates pass.
- Talent editing, hover tooltip service, social controls, PvP tab, or
  shirt/tabard slots (no data exists for them).
- Any change to ranking, simulation, validation codes, or job lifecycle.
- New frontend or Go dependencies.

## Architecture

The Svelte/Go split stays as in the 2026-08-29 spec. Within `ui-finder`:

```text
ArmoryView.svelte        tabs, layout, character stage
  CharacterHeader.svelte compact identity + Find upgrades action
  CharacterStage.svelte  backdrop + labeled placeholder (3D seam)
  GearSlot.svelte        reworked item card (ilvl, name, enchant, sockets, details)
  StatPanels.svelte      unchanged data contract, Stats tab
  TalentTrees.svelte     read-only three-tree view
```

The `data-region="armory-view"` surface and the existing `POST /api/import`
response fields stay valid; new fields are additive.

### Backend additions (`cmd/wowsimcli/cmd/upgrades`)

- `GearSlotData` gains `ilvl int32` from the enriched `UIItem.GetIlvl()`.
- `EnchantData` gains `description string`, resolved during enrichment from a
  startup-loaded effect-ID → text map built from
  `assets/enchants/descriptions.json`, falling back to the enchant name. The
  file is loaded the same way the existing database blob is loaded at startup.
- The import response gains `talentsString` from
  `imported.Settings.Player.GetTalentsString()`.
- No existing field is removed or renamed.

### UI behavior

- **Header:** name, level 70 race and spec, professions, phase; settings
  digest moves into import details. A visible Find upgrades action jumps to
  the ranking controls.
- **Item card:** icon with ilvl badge (top-left) and socket strip (bottom);
  rarity-colored item name; green enchant effect line, e.g.
  `+34 Attack Power and +16 Hit Rating`, or `Mongoose` for proc-only effects;
  no enchant means no line. A quiet slot caption appears only where it adds
  meaning (e.g. empty slots). Details panel on click/keyboard focus: item ID,
  phase, complete stats, set name, socket bonus and its active/inactive state.
- **Center stage:** restrained backdrop using existing bundled art and a
  labeled `Character preview — appearance not imported` placeholder. The
  `Activate 3D` control is present but informational until the gated viewer
  integration exists.
- **Weapons strip:** Main Hand, Off Hand, Ranged beneath the stage. All 17
  supported slots render; placement is explicit per slot, never inferred from
  array slices.
- **Tabs:** Gear / Stats / Talents. Stats keeps the exact server data and the
  **raid buffed (link settings)** label. Talents is read-only.
- **Ranking controls** remain below the armory and always reachable.
- **Responsive:** at ≤ 900 px the stage becomes an optional compact panel;
  at ≤ 640 px gear renders as one readable column.

### Error handling and fallback

- Icon CDN failure keeps the existing quality-colored initial fallback.
- Unknown enchant effect ID in the description map shows the enchant name.
- Missing 3D support, WebGL failure, or the ungated viewer never affects
  gear, stats, talents, or ranking.

## Verification

- Go: extend the armory tests — ilvl present for a known enriched item,
  enchant description resolved for a fixture enchant, import response carries
  the talents string for the fixture.
- End-to-end: update the existing Playwright test for the new layout and
  verify all 17 slot cards, ilvl badges, enchant effect lines, socket/gem
  strips, tab switching, and the unchanged ranking flow (progress, report,
  copy, cancel).
- Manual smoke: import the Retribution fixture; confirm the layout at a wide
  viewport and the narrow collapse, the stats tab keeps the raid-buffed
  label, and the talents tab renders three trees for the fixture Paladin.

## Deferred

- 3D viewer integration (gated; acceptance criteria in
  `docs/armory-redesign-research.md`).
- Exact appearance lookup.
- Talent editing and baseline/preview visual comparison.
