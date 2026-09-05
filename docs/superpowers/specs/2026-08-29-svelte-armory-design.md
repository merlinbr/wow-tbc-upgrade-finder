# Svelte armory UI design

## Status

Approved direction 2026-08-29 (user selected approach A). This spec covers the
Svelte migration and the armory character-review view. It amends the
implementation plan's "no-build browser UI" constraint: the UI may now use a
build toolchain, as long as the compiled bundle is committed and `go build`
needs no Node.

## Goal

After a successful link import, show an armory-style character view (like the
provided wowsims screenshot, without the 3D model): the complete equipped gear
with gems and enchants per slot, plus basic unbuffed stats. This view replaces
the current plain summary list as the "review before ranking" step. The rest
of the flow (filters/policy → ranking → report) is ported into the same Svelte
app.

## Decisions

| Area | Decision | Reason |
| --- | --- | --- |
| Framework | Svelte 5 (runes) + Vite | User decision; reactive state fits gear/stat panels; compiles to plain JS with a tiny runtime. |
| Delivery | Built bundle committed under `cmd/wowsimcli/cmd/upgrade_ui/` | Keeps `rtk go build`/tests Node-free; rebuild UI only when it changes. |
| Armory data source | Server-side enrichment in Go from the bundled `UIDatabase` | The database is the only item-data authority; browser stays a renderer. |
| Stat fidelity | Engine-computed level-70 base + gear snapshot, excluding buffs, consumes, and talents; labeled "unbuffed (base + gear)" | Uses the simulator's deterministic stat pipeline without running iterations, so racial effects, stat dependencies, gems, enchants, socket bonuses, suffixes, and gear effects agree with ranking inputs. |
| Icons | `wow.zamimg.com` CDN with quality-colored placeholder fallback | Same source wowsims uses; zero storage; server itself stays loopback-only. |
| Layout | Two-column gear list around an empty center gap; stat panel footer | Reproduces the reference screenshot minus the 3D model. |
| Scope | Gear columns + gems/enchants + stat panels; no talents tab, no tooltips | Matches the user's stated v1. |

## Non-goals

- 3D character model.
- Talents tab / talent tree display.
- Item detail tooltips (hover cards).
- Buffed stats, talents, consumes, or a simulation run for the armory panel.
- Any change to ranking logic, job lifecycle, or existing successful HTTP fields
  beyond additive response fields.
- New Go runtime dependencies.

## Architecture

```text
ui-finder/                     Svelte 5 + Vite project (UI source of truth)
  package.json / vite.config.mts
  src/main.js, src/App.svelte
  src/lib/{api.js, stores.svelte.js}
  src/lib/{ImportPanel,ArmoryView,GearSlot,StatPanels,RankingPanel,ReportView}.svelte
  src/app.css
        │  npm run build (output committed)
        ▼
cmd/wowsimcli/cmd/upgrade_ui/  embed.FS target (index.html + hashed assets)
        │  go:embed (unchanged)
        ▼
wowsimcli binary  ──  localhost JSON API (unchanged routes)
```

- `vite.config.mts` sets `base: './'`, `build.outDir` to the resolved
  `../cmd/wowsimcli/cmd/upgrade_ui/` path (relative to the `ui-finder` project
  root), and `emptyOutDir: true`. Vite's normal `assets/` directory is retained.
  The old `index.html`/`app.js`/`app.css` are replaced by the build output
  (Vite emits its own `index.html`).
- Server change: `GET /assets/{path...}` serves a nested asset beneath the
  embedded `upgrade_ui/assets/` directory. It accepts only a valid relative
  file path, returns a structured 404 otherwise, sets content-type by
  extension, and never lists a directory. `GET /` still serves `index.html`.
  API routes and loopback binding are unchanged.
- `docs/upgrade-finder.md` gains the UI rebuild command
  (`cd ui-finder && npm ci && npm run build`); `package-lock.json` is
  committed with the UI source.

## Backend additions

New `cmd/wowsimcli/cmd/upgrades/armory.go`:

- `EnrichArmory(imported *ImportedSettings, catalog *Catalog) (*ArmoryData, error)`
  is called by the import handler after `upgrades.Import`.
- `ArmoryData{Gear []GearSlotData, Stats map[string]float64,
  DerivedStats map[string]float64}`. `Gear` always contains the canonical 17
  slots, in display order, including explicit empty-slot entries.
- `Catalog` also indexes `UIDatabase.RandomSuffixes` by ID.

`GearSlotData` per canonical slot:

```json
{
  "slot": 4, "slotName": "Chest",
  "itemId": 29077, "itemName": "Robes of the Aldor", "quality": 4,
  "icon": "inv_chest_cloth_03", "phase": 3, "setName": "",
  "stats": {"spell_damage": 37},
  "randomSuffix": null,
  "sockets": [
    {"color": "Blue", "gem": {"id": 32227, "name": "Sparkling Azure Moonstone",
                              "icon": "inv_jewelcrafting_gem_18", "color": "Blue"}},
    {"color": "Yellow", "gem": null}
  ],
  "socketBonus": {"stats": {"spell_damage": 4}, "active": false},
  "enchant": null
}
```

- `randomSuffix` is `null` or the selected suffix's ID, name, and scaled stat
  contribution. Its scaling exactly follows `core.ItemEquipmentBaseStats`.
- Slots in the reference order: left column Head, Neck, Shoulder, Back, Chest,
  Wrist, MainHand, OffHand; right column Hands, Waist, Legs, Feet, Finger1,
  Finger2, Trinket1, Trinket2, Ranged.
- `enchant: null` renders as "No Enchant". A zero gem ID renders as an empty
  socket. An empty equipment slot renders as an empty slot, not a missing
  `GearSlotData` entry.
- `Import` validates every nonzero equipped item, random-suffix, gem, and
  enchant-effect ID against the bundled database. An unsupported record returns
  a typed `incompatible_*` validation error before enrichment or ranking; it is
  never converted into a placeholder. This keeps armory review and later
  ranking on the same accepted input.

Stats computation (deterministic Go, no simulation iterations):

1. Clone the imported setup and clear raid, party, and individual buffs,
   consumes, and talent configuration. Keep the selected race, class,
   professions, and equipment.
2. Use `core.ComputeStats` on that clone and expose the player's gear-stage
   stat snapshot. This applies permanent racial effects, stat dependencies,
   game rounding, selected random suffixes, gems, enchants, valid socket
   bonuses, set bonuses, and static gear effects with the same engine rules
   used by ranking.
3. `stats`, gear-item `stats`, suffix `stats`, and socket-bonus `stats` use
   the snake_case form of `stats.StatName()` with no prefix. Examples:
   `agility`, `spell_damage`, `spell_hit_rating`, `melee_hit_rating`, and
   `mp5`. `item_stat_*`, `spell_power`, and ambiguous `hit_rating` keys are
   never emitted.
4. `derivedStats` contains only engine-produced percentage values, keyed
   `melee_hit_percent`, `spell_hit_percent`, `melee_crit_percent`,
   `spell_crit_percent`, `ranged_hit_percent`, `ranged_crit_percent`, and
   `block_percent`; values are percentages in the range 0–100.
5. Socket-bonus activation uses `core.ColorIntersects`, including hybrid,
   prismatic, and meta gem behavior. The panel is labeled
   "unbuffed (base + gear)".

`POST /api/import` response gains `gear`, `stats`, and `derivedStats` alongside
the existing successful fields. Existing validation codes and response shapes
remain unchanged; the new `incompatible_*` codes apply only to previously
unchecked unsupported equipment records.

## UI flow

1. **Import** — paste link, POST `/api/import`, typed validation errors
   (including `incompatible_*`) in an alert region.
2. **Armory (review)** — character header (name, level 70 race/class,
   professions, phase, settings digest); two-column gear list per the order
   above; stat panel footer. "Start ranking" controls remain below the
   armory, as today's controls remain below the summary.
3. **Ranking** — same polling (500 ms) and cancel semantics, Svelte-driven.
4. **Report** — same table/fields, plus the copy button; attribution footer
   stays visible on every screen.

State lives in one Svelte store module (`imported`, `job`, `report`, `error`);
views switch on flow state; canceling clears report state exactly as today.

## Error handling

- Import validation errors, including unsupported item, random-suffix, gem, or
  enchant IDs, use typed `incompatible_*` codes in the live alert region.
- A zero equipment or gem ID is an ordinary empty slot/socket; it is not an
  error and does not produce a placeholder record.
- Offline/blocked CDN: `<img>` `onerror` swaps to a quality-colored square
  with the slot name initial.
- Asset 404s remain structured JSON errors.

## Verification

- Go contract tests: enriched fixture import returns all 17 canonical gear
  entries with correct item names/qualities; resolves gems, enchants, and a
  selected random suffix from the bundled DB; and reports hybrid/prismatic
  socket bonuses with the same result as `core.ColorIntersects`. The armory
  stats and derived percentages match the sanitized `core.ComputeStats`
  fixture snapshot. Unsupported nonzero item, suffix, gem, and enchant IDs
  return typed validation errors; baseline byte-identity tests keep passing.
- Server route tests parse the committed generated index, extract every asset
  URL, and fetch each exact nested hashed URL as JavaScript or CSS.
- Playwright smoke: armory renders for the fixture mage (all 17 slots, gem
  squares, enchant lines, stat panels), then the full ranking flow behaves as
  before (progress, report, copy, cancel).

## Deferred

Talents tab, item hover tooltips, buffed-stat parity, 3D model, pricing and
multi-item optimization (all unchanged from the base spec's deferral list).
