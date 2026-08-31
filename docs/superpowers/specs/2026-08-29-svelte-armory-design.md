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
| Stat fidelity | Race/class base at 70 + equipped items + gems + enchants + active socket bonuses; labeled "unbuffed" | Honest and deterministic; buff/talent/sim-computed parity is explicitly not a goal. |
| Icons | `wow.zamimg.com` CDN with quality-colored placeholder fallback | Same source wowsims uses; zero storage; server itself stays loopback-only. |
| Layout | Two-column gear list around an empty center gap; stat panel footer | Reproduces the reference screenshot minus the 3D model. |
| Scope | Gear columns + gems/enchants + stat panels; no talents tab, no tooltips | Matches the user's stated v1. |

## Non-goals

- 3D character model.
- Talents tab / talent tree display.
- Item detail tooltips (hover cards).
- Buffed or sim-computed stats.
- Any change to ranking logic, job lifecycle, or the existing HTTP contracts
  beyond additive response fields.
- New Go runtime dependencies.

## Architecture

```text
ui-finder/                     Svelte 5 + Vite project (UI source of truth)
  package.json / vite.config.mts
  src/main.js, src/App.svelte
  src/lib/{api.js, stores.js}
  src/lib/{ImportPanel,ArmoryView,GearSlot,StatPanels,RankingPanel,ReportView}.svelte
  src/app.css
        │  npm run build (output committed)
        ▼
cmd/wowsimcli/cmd/upgrade_ui/  embed.FS target (index.html + hashed assets)
        │  go:embed (unchanged)
        ▼
wowsimcli binary  ──  localhost JSON API (unchanged routes)
```

- `vite.config.mts` sets `base: './'`, `build.outDir` to `upgrade_ui/`,
  `emptyOutDir: true`. The old `index.html`/`app.js`/`app.css` are replaced by
  the build output (vite emits its own `index.html`).
- Server change: `GET /assets/{name}` serves any file present in the embedded
  FS by name (404 otherwise, content-type by extension). No directory listing.
  `GET /` still serves `index.html`. API routes and loopback binding unchanged.
- `docs/upgrade-finder.md` gains the UI rebuild command
  (`cd ui-finder && npm install && npm run build`).

## Backend additions

New `cmd/wowsimcli/cmd/upgrades/armory.go`:

- `EnrichArmory(imported *ImportedSettings, catalog *Catalog) (*ArmoryData, error)`
  called by the import handler after `upgrades.Import`.
- `ArmoryData{Gear []GearSlotData, Stats map[string]float64}`.

`GearSlotData` per equipped slot:

```json
{
  "slot": 4, "slotName": "Chest",
  "itemId": 29077, "itemName": "Robes of the Aldor", "quality": 4,
  "icon": "inv_chest_cloth_03", "phase": 3, "setName": "",
  "stats": {"item_stat_spell_power": 37},
  "sockets": [
    {"color": "Blue", "gem": {"id": 32227, "name": "Sparkling Azure Moonstone",
                              "icon": "inv_jewelcrafting_gem_18", "color": "Blue"}},
    {"color": "Yellow", "gem": null}
  ],
  "socketBonus": {"stats": {"item_stat_spell_power": 4}, "active": false},
  "enchant": null
}
```

- Slots in the reference order: left column Head, Neck, Shoulder, Back, Chest,
  Wrist, MainHand, OffHand; right column Hands, Waist, Legs, Feet, Finger1,
  Finger2, Trinket1, Trinket2, Ranged.
- `enchant: null` renders as "No Enchant". Missing gem → `"gem": null` renders
  as an empty socket. Missing item/gem/enchant records never fail the import;
  they degrade to placeholders.
Stats computation (pure Go, no sim run):

1. Start from race/class base stats at level 70 (reuse the engine's base-stats
   tables; if the needed accessor is unexported, use the existing generated
   table directly).
2. Add `UIItem.Stats`, gem `UIGem.Stats`, enchant `UIEnchant.Stats`. All stats
   maps use the `proto.Stat` enum's Go name in snake_case as the JSON key
   (e.g. `agility`, `spell_power`, `hit_rating`), consistently for item,
   socket-bonus, and total maps.
3. Socket bonus counts only when every non-meta socket holds a gem whose color
   matches the socket color.
4. Derived values use the engine's `stats` package conversion constants where
   they exist (e.g. crit/hit rating → percent, stamina → health). If a
   conversion is not cleanly available, the field is omitted rather than
   approximated. Health/Mana: base plus derived-from-stamina/intellect using
   engine constants.
5. The panel is labeled "unbuffed (base + gear)". Exact buffed wowsims parity
   is out of scope.

`POST /api/import` response gains `gear` and `stats` alongside the existing
fields; existing fields and error contract are unchanged.

## UI flow

1. **Import** — paste link, POST `/api/import`, typed errors in an alert
   region (unchanged codes).
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

- Import validation errors: unchanged typed codes, rendered in the live alert
  region.
- Unknown item/gem/enchant ID during enrichment: placeholder rendering, never
  a 500.
- Offline/blocked CDN: `<img>` `onerror` swaps to a quality-colored square
  with the slot name initial.
- Asset 404s remain structured JSON errors.

## Verification

- Go contract tests: enriched fixture import returns 17 gear entries with
  correct item names/qualities for known slots; gem/enchant resolution from
  the bundled DB; hand-computed stat totals for at least two raw stats and one
  derived stat; baseline byte-identity tests keep passing.
- Server route tests updated for hashed asset names (index references an
  asset; fetching that asset returns JavaScript/CSS).
- Playwright smoke: armory renders for the fixture mage (all 17 slots, gem
  squares, enchant lines, stat panels), then the full ranking flow behaves as
  before (progress, report, copy, cancel).

## Deferred

Talents tab, item hover tooltips, buffed-stat parity, 3D model, pricing and
multi-item optimization (all unchanged from the base spec's deferral list).
