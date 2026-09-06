# Character identity header design

## Status

Proposed 2026-09-07. This amends the presentation scope of
`2026-09-05-armory-redesign-design.md` for the character identity area only: the
header card and the tab bar. Import flow, armory content, ranking, and
simulation are unchanged. No backend changes: every field used is already in
the `/api/import` response or derivable from it.

Reference: third-party armory-style character page (screenshot 2) — identity
card with avatar, class-colored name, level/race/class/spec meta line, a
stat-chip row, a muted absence note, and a pill-style tab bar.

## Goal

Restyle the imported-character identity area to match the reference's
structure: an identity card with an avatar, class-colored character name, a
meta line, a chip row of real import-derived facts, one muted note for data
the sim link cannot carry, then the Gear / Stats / Talents tabs restyled as a
pill bar. The card must never display data the import does not contain.

## Decisions (user-confirmed 2026-09-07)

| Area | Decision | Reason |
| --- | --- | --- |
| Scope | Identity card **and** tab bar restyle; keep the existing Gear / Stats / Talents tabs | Captures the reference's circled area without inventing a Raid Progression/PvP surface for which we have no data. |
| Missing data | Real values only, plus one muted note; no field for server, arena rating, achievements, or portrait | The wowsims export carries none of them; placeholders would fabricate identity. |
| Avatar | Class icon from the ZAM CDN (`class_paladin.jpg` etc.), quality fallback like gear icons | Same CDN as existing item icons; no new dependency or asset download. |
| Name color | Class hex color, same values as the simulator UI (`ui/core/player_classes/*.ts`) | Matches the reference's class-colored name convention. |

## Non-goals

- Server/realm name, arena ratings, achievements, guild, portrait, or any
  Blizzard profile lookup (no data source in the sim link).
- Spec selector / "ACTIVE" pill and talent-tree spec switching (display-only
  state that would imply a second spec exists).
- Any change to `/api/import`, Go enrichment, ranking, or the 17-slot armory.
- New frontend or Go dependencies; no asset downloads.
- Raid Progression / PvP tabs.

## Architecture

All work is in `ui-finder/`. Source of display facts stays the browser:

- `src/lib/identity.js` (new): class hex colors, class/spec icon URL
  maps (mirroring `ui/core/player_classes/*.ts` and `ui/core/player_specs/*.ts`
  values), and `avgItemLevel(gear)` — average of item level over slots whose
  `itemId` is non-zero (skips empty slots), rounded to an integer.
- `CharacterHeader.svelte`: restructured identity card. Props gain `gear`.
  Renders: avatar (class icon with letter fallback), name in class color,
  meta line (level 70 · race · spec · class), chip row (avg ilvl, phase,
  professions), muted note, existing `Find upgrades` + `Import details`
  actions unchanged.
- `ArmoryView.svelte`: passes `gear` to `CharacterHeader`.
- `app.css`: card layout, chip row, and pill tab bar styles replacing the
  current underline tab style. Tab markup, ids, roles, and labels stay fixed
  (E2E and a11y depend on them).

### Identity card content

- Avatar: `https://wow.zamimg.com/images/wow/icons/medium/<class-icon>.jpg`;
  on load error, a letter tile with the class color background — same
  fallback pattern as `GearSlot`.
- Name: `character.name`, colored `classColor(character.class)`.
- Meta line: `Level 70 {race} · {spec} {class}` via `humanizeEnum` (same
  labels as today).
- Chips: `Avg ilvl {n}` (from `gear`), `Phase {phase}`, professions (or
  `None`).
- Muted note: `No ratings — local simulation import` (the sim link carries no
  ratings; mirrors the reference's "No ratings yet." placement).

### Tab bar

Pill style: rounded buttons on a dark track, active tab filled with the
accent and readable text; same `role="tab"` / `aria-controls` /
`aria-selected` contract, same tab ids and labels.

## Error handling and fallback

- Class icon CDN failure → class-colored letter tile (established fallback).
- Unknown class/spec enum → name stays uncolored, avatar letter tile; no
  crash.
- Missing gear / zero equipped items → `Avg ilvl —`, note and chips still
  render.

## Verification

- Unit: `identity.test.js` — avg-item-level skips empty slots and rounds;
  class color/icon map covers all nine classes; unknown values return empty
  without throwing.
- E2E (`armory.spec.js`): existing assertions (heading `TestMage`, 17 slots,
  tab switching, Find upgrades, report) must pass unchanged; add assertions
  for the avatar, class-colored name style, chip row values, muted note, and
  pill tabs.
- Manual smoke: import the Retribution fixture; check wide + narrow layout,
  that the header shows avg ilvl matching the fixture's non-empty slots, and
  the talents tab still renders three trees.

## Deferred

- Optional realm input, spec ACTIVE chips, portraits, Blizzard profile data.
