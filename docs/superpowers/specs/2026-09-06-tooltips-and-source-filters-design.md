# Tooltips and Source Filters — Design

Date: 2026-09-06
Status: Approved

Two display-layer features for the TBC Upgrade Finder:

1. **Item hover tooltips** in the armory and report, built from data the
   server already sends.
2. **Source-kind ranking filters** in the ranking panel, wired to the
   already-implemented server-side `sourceKinds` filter.

Both are UI-only. No Go changes, no API changes, no new data flow.

## 1. Item hover tooltips

### Behavior

Hovering (or keyboard-focusing) a gear card's icon/name in `GearSlot.svelte`
shows a popover with:

- item name (quality-colored) and item level;
- humanized stat lines — proper stat names ("Hit Rating"), not raw
  snake_case keys (current `displayStats` output);
- enchant effect line;
- gems with socket colors;
- socket bonus with active/inactive state;
- phase and quality name.

Report table rows (`ReportView.svelte`) reuse the same popover on the item
cell.

Empty slots show no tooltip. The existing Details disclosure stays; the
tooltip is the quick-glance layer.

### Implementation

- New component `ui-finder/src/lib/ItemTooltip.svelte` taking an
  item-shaped object (the same shape `GearSlot` slots and report
  `upgrade.item` rows already carry).
- Stat-name mapping added to `ui-finder/src/lib/labels.js` alongside the
  existing `humanizeEnum` helpers; `GearSlot` and the tooltip share it.
- Visibility via CSS `:hover` + `:focus-within`; positioned above the card;
  rendered inline (no portal). Keyboard focus requires the trigger to be a
  real focusable element (button or link on the icon/name).

### Non-goals

- No Wowhead script, no external network dependency (keeps the
  "browser performs no item lookup" boundary).
- No flavor text, vendor, or acquisition info — not in the armory payload.

## 2. Source-kind ranking filters

### Behavior

A checkbox group in `RankingPanel.svelte` listing the source kinds already
known to the server: Crafting, Quest, Reputation, PvP, Dungeon, Heroic
dungeon, Raid, Heroic raid, Raid finder, Flexible raid, Sold by vendor
(12 enum values; Unknown source is governed by the existing
*Include unknown-source items* toggle, not this group).

Semantics: checked kinds are **included**; none checked = all sources (the
server treats an empty `sourceKinds` array as no filter).

Selection is sent as `filters.sourceKinds` in the ranking request. The
report already displays applied source filters (`Source filters` summary
row and exclusion counts) — no report changes needed.

### Implementation

- State in `RankingPanel.svelte`; submitted through the existing
  `startRanking` payload (currently hardcodes `sourceKinds: []`).
- The `sourceKinds` label map moves from `ReportView.svelte` (private
  const) to `labels.js` so panel and report share one mapping.
- Invalid kinds are rejected by the server with `invalid_options`, surfaced
  by the existing `onValidationError` path — no client-side duplication.

### Non-goals

- `sourceNames` filtering (specific boss/dungeon names): the server supports
  it, but the catalog source-name vocabulary is not exposed to the browser.
  Deferred until that metadata endpoint exists.

## Testing

- Unit test for the stat-name and source-kind label maps (`labels.test.js`).
- Extend `ui-finder/e2e/armory.spec.js`: hover a gear card and assert
  tooltip content; check a source-kind box, run ranking (minimal
  iterations), and assert the report's `Source filters` row reflects it.
- Server-side source filtering already has Go coverage; unchanged.
