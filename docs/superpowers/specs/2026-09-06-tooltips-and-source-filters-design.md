# Tooltips and Source Filters — Design

Date: 2026-09-06 (rev 2, after review)
Status: Approved

Two features for the TBC Upgrade Finder:

1. **Item hover tooltips** in the armory (full detail) and the report
   (summary detail), built from data the server already sends.
2. **Source-kind ranking filters** in the ranking panel, wired to the
   server-side `sourceKinds` filter.

Payload types referenced below live in
`cmd/wowsimcli/cmd/upgrades/types.go`: `GearSlotData` (:43) carries stats,
random suffix, sockets, socket bonus, and enchant; `UIItemSummary` (:138)
carries only id, name, icon, quality, phase, type, slot.

## 1. Item hover tooltips

### Armory tooltip (full)

Hovering (or keyboard-focusing) a gear card's icon/name in `GearSlot.svelte`
shows a popover styled after the classic Wowhead item tooltip (mockups
approved 2026-09-06), top to bottom:

1. Header row: quality-colored item name; right-aligned `Phase N` badge.
2. `Item Level N` in the item-level gold/orange.
3. Slot line (`slotName`).
4. Stat lines in white, humanized (`+32 Strength`, `+45 Stamina`) — proper
   stat names, not raw snake_case keys.
5. Random suffix: when `RandomSuffix` is present, its stats render as
   additional stat lines (suffix stats are not in `slot.stats`, so without
   this random-suffix items show incomplete stats).
6. Gems: per socket, the gem's colored icon followed by its stat
   contribution in green (`+10 Strength`), matching the mockup's socketed
   gem lines; empty sockets render as an empty colored socket icon.
7. `Socket Bonus: …` in green, with inactive bonuses shown dimmed/gray.
8. Enchant as green `Equip: …` lines (`EnchantData.Description`, falling
   back to its name).

Tooltip body text is white on a dark `#0a0d14`-style panel with a dark
border, like the mockups.

Empty slots show no tooltip. The existing Details disclosure stays; the
tooltip is the quick-glance layer.

Mockup lines with no data behind them are omitted: `Binds when picked up`
(constant flavor), armor type (Plate/Leather — `GearSlotData` carries no
item class), weapon damage/speed/DPS, durability, class restriction,
`Requires Level` (always 70), sell price, and the set-piece list with set
bonuses (only `setName` is in the payload). Adding any of these is a
server payload change, out of scope.

### Report tooltip (reduced)

Report rows carry `UIItemSummary` only — no stats, sockets, enchant, or
ilvl, and upgrade candidates are new items absent from the armory's gear
array, so full tooltips are impossible without an API change. The report
tooltip is therefore reduced to what the summary carries: item name
(quality-colored), quality name, phase, and icon.

Adding full tooltips to the report would require adding a `UIItem`-level
field to `ConfirmedUpgrade` — a Go/proto/API change. Deferred; do it only
if full detail in reports is wanted.

### Implementation

- New component `ui-finder/src/lib/ItemTooltip.svelte` with a `variant`:
  `full` (armory, `GearSlotData`) or `summary` (report, `UIItemSummary`).
- `ui-finder/src/lib/labels.js` gains, alongside `humanizeEnum`:
  - stat-name map (snake_case stat key → display name);
  - `sourceKinds` map moved from `ReportView.svelte` (private const);
  - `qualityNames` map moved from `ReportView.svelte` (private const);
  - `socketColors` map moved from `GearSlot.svelte` (private const).
  Tooltip, GearSlot, ReportView, and RankingPanel all consume these — no
  duplicated maps anywhere.
- Visibility via CSS `:hover` + `:focus-within`; positioned above the card;
  rendered inline (no portal). Keyboard focus requires the trigger to be a
  real focusable element (button on the icon/name).

### Non-goals

- No Wowhead script, no external network dependency (keeps the
  "browser performs no item lookup" boundary).
- No flavor text, vendor, or acquisition info — not in the payloads.
- No report full-detail tooltip (would be a Go change; deferred above).

## 2. Source-kind ranking filters

### Wire format

`ContentFilters.SourceKinds` is `[]proto.SourceFilterOption` — a numeric
Go enum alias decoded with `encoding/json`. The checkbox submit therefore
sends **numeric enum values** (`[5, 7]`), not strings; string arrays fail
to unmarshal (`invalid_request`).

Out-of-range numerics decode silently and are never matched — the server
has no `invalid_options` validation for source kinds (only for iteration
counts). No server-side range validation is added: the checkbox group is
generated from the fixed enum map in `labels.js`, so the browser can only
ever send valid values. There is nothing to validate on either end.

### Behavior

A checkbox group in `RankingPanel.svelte` listing the acquireable source
kinds from the shared `labels.js` map (Crafting, Quest, Reputation, PvP,
Dungeon, Heroic dungeon, Raid, Heroic raid, Raid finder, Flexible raid,
Sold by vendor). "Unknown source" is the 12th enum value but is governed by
the existing *Include unknown-source items* toggle and is not part of this
group.

Semantics: checked kinds are **included**; none checked = all sources (the
server treats an empty `sourceKinds` array as no filter).

Selection is sent as numeric `filters.sourceKinds` in the ranking request,
replacing the current hardcoded `sourceKinds: []`.

### Report display

With a non-empty filter, the server echoes `assumptions.sourceKinds` as
proto enum names (`kind.String()` → `"SourceCrafting"`), which
`ReportView.svelte` currently renders raw. The Source filters row gains a
humanizer that maps these enum-name strings to display labels (strip the
`Source` prefix, apply the shared label map; unknown strings render as-is).
No other report changes needed — exclusion counts already display.

### Non-goals

- `sourceNames` filtering (specific boss/dungeon names): the server supports
  it, but the catalog source-name vocabulary is not exposed to the browser.
  Deferred until that metadata endpoint exists.
- Server-side source-kind range validation: unreachable through the
  checkbox UI; add only if the API surface becomes directly scriptable.

## Testing

- Unit tests for the shared label maps in `labels.test.js`: stat names,
  source kinds, quality names, socket colors, and the
  `SourceCrafting → Crafting` humanizer.
- Extend `ui-finder/e2e/armory.spec.js`: hover a gear card and assert
  tooltip content (including an enchant line and a socket line); check a
  source-kind box, run ranking (minimal iterations), and assert the
  report's `Source filters` row shows the humanized label.
- Server-side source filtering already has Go coverage; unchanged. The
  request-decoding change (numeric enums) is exercised through the e2e
  run.
