# Tooltips and Source Filters Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Wowhead-style hover tooltips (armory: full detail; report: summary) and source-kind ranking filter checkboxes.

**Architecture:** UI-only changes in the isolated Svelte 5 app (`ui-finder/`). Shared label maps move from private component consts into `labels.js`. A new `ItemTooltip.svelte` renders two variants from data the server already sends (`GearSlotData`, `UIItemSummary`). `RankingPanel` submits numeric `SourceFilterOption` enum values; the server echoes proto names that `ReportView` humanizes.

**Tech Stack:** Svelte 5 (runes), Vite 7, Playwright (Chromium), `node --test`.

**Spec:** `docs/superpowers/specs/2026-09-06-tooltips-and-source-filters-design.md`

## Global Constraints

- No Go/proto/API changes; no new network calls from the browser (no Wowhead script).
- Source-kind wire format is **numeric enum values** (`[5, 7]`), never strings — `ContentFilters.SourceKinds` is `[]proto.SourceFilterOption`; strings fail JSON unmarshal.
- No item icon in the tooltip header; hover/focus keeps the card content visible.
- Tooltip body: white text on dark panel (`#0a0d14`-style) with dark border; stat/gem/enchant colors per mockups.
- Source-kind checkbox group lists the 11 acquireable kinds; value `0` (Unknown source) stays governed by the existing *Include unknown-source items* toggle.
- Commands: `rtk npm run build` in `ui-finder/`; unit tests `rtk node --test src/lib/<file>` in `ui-finder/`; e2e `rtk npm run test:e2e` in `ui-finder/`.
- The generated Vite bundle is committed under `cmd/wowsimcli/cmd/upgrade_ui/` — the final task must rebuild and commit it.

## File Structure

- `ui-finder/src/lib/labels.js` — shared label maps + humanizers (single source of truth; no component-private duplicates afterward).
- `ui-finder/src/lib/ItemTooltip.svelte` — tooltip popover, `full` (GearSlotData) and `summary` (UIItemSummary) variants.
- `ui-finder/src/lib/GearSlot.svelte` — armory trigger + full tooltip; drops its private `socketColors`.
- `ui-finder/src/lib/ReportView.svelte` — summary tooltip on item cells; consumes shared maps; humanizes the Source filters row; drops its private `qualityNames`/`sourceKinds`.
- `ui-finder/src/lib/RankingPanel.svelte` — source-kind checkbox group; numeric submit.
- `ui-finder/e2e/armory.spec.js` — hover and filter assertions.
- `ui-finder/src/lib/labels.test.js` — map/humanizer unit tests.

---

### Task 1: Shared label maps and humanizers

**Files:**
- Modify: `ui-finder/src/lib/labels.js`
- Test: `ui-finder/src/lib/labels.test.js`

**Interfaces:**
- Consumes: nothing new.
- Produces (used by Tasks 2–5):
  - `statLabel(key: string): string` — `"spell_hit_rating"` → `"Spell Hit Rating"`, `"mp5"` → `"MP5"`, unknown keys humanized generically.
  - `formatStatLine(key: string, value: number): string` — `"+32 Strength"`.
  - `qualityLabel(value: number): string` — label or `Unknown quality (N)` (backed by a module-private `qualityNames` map).
  - `socketColors: Record<number, string>` — `{ 0: 'Unknown', 1: 'Meta', 2: 'Red', 3: 'Blue', 4: 'Yellow', 5: 'Green', 6: 'Orange', 7: 'Purple', 8: 'Prismatic' }`.
  - `sourceKinds: Array<{ value: number, label: string, proto: string }>` — 12 entries, value `0` first.
  - `sourceKindLabel(value: number): string` — label or `Unknown source (N)`.
  - `humanizeSourceKind(name: string): string` — `"SourceCrafting"` → `"Crafting"`; falls back to `humanizeEnum(name, 'Source')`.

- [ ] **Step 1: Write the failing tests**

Append to `ui-finder/src/lib/labels.test.js` (update the import line to pull the new names):

```js
import {
  humanizeEnum, statLabel, formatStatLine,
  qualityLabel, sourceKindLabel, humanizeSourceKind, sourceKinds, socketColors,
} from './labels.js';

test('statLabel humanizes snake_case stat keys with acronyms', () => {
  assert.equal(statLabel('strength'), 'Strength');
  assert.equal(statLabel('spell_hit_rating'), 'Spell Hit Rating');
  assert.equal(statLabel('mp5'), 'MP5');
  assert.equal(statLabel('armor_penetration_rating'), 'Armor Penetration Rating');
  assert.equal(statLabel('some_future_stat'), 'Some Future Stat');
});

test('formatStatLine signs and formats values', () => {
  assert.equal(formatStatLine('strength', 32), '+32 Strength');
  assert.equal(formatStatLine('armor', 1825), '+1825 Armor');
  assert.equal(formatStatLine('mp5', 7.5), '+7.5 MP5');
});

test('qualityLabel maps known qualities and flags unknown', () => {
  assert.equal(qualityLabel(4), 'Epic');
  assert.equal(qualityLabel(99), 'Unknown quality (99)');
});

test('sourceKindLabel maps known kinds and flags unknown', () => {
  assert.equal(sourceKindLabel(6), 'Heroic dungeon');
  assert.equal(sourceKindLabel(42), 'Unknown source (42)');
});

test('humanizeSourceKind maps proto names to labels', () => {
  assert.equal(humanizeSourceKind('SourceCrafting'), 'Crafting');
  assert.equal(humanizeSourceKind('SourceHeroicDungeon'), 'Heroic dungeon');
  assert.equal(humanizeSourceKind('SourceSomethingNew'), 'Something New');
});

test('sourceKinds lists 12 kinds with Unknown first', () => {
  assert.equal(sourceKinds.length, 12);
  assert.deepEqual(sourceKinds[0], { value: 0, label: 'Unknown source', proto: 'SourceUnknown' });
  assert.equal(sourceKinds.find((k) => k.value === 7).label, 'Raid');
});

test('socketColors covers all gem color enum values', () => {
  assert.equal(socketColors[1], 'Meta');
  assert.equal(socketColors[2], 'Red');
  assert.equal(socketColors[8], 'Prismatic');
  assert.equal(Object.keys(socketColors).length, 9);
});
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd ui-finder && rtk node --test src/lib/labels.test.js`
Expected: FAIL — `statLabel` is not exported.

- [ ] **Step 3: Write the implementation**

Full new content of `ui-finder/src/lib/labels.js`:

```js
// Turns protobuf enum names ("RaceHuman", "ClassPaladin", "RetributionPaladin")
// into display text: strips the enum prefix, then splits camelCase words.
export function humanizeEnum(value, prefix) {
  if (!value) return '';
  const name = prefix && value.startsWith(prefix) ? value.slice(prefix.length) : value;
  return name.replace(/([a-z])([A-Z])/g, '$1 $2');
}

// Item stat keys arrive as snake_case of the engine's StatName() values
// (sim/core/stats/stats.go), e.g. "spell_hit_rating", "mp5".
const upperAcronyms = new Set(['mp5']);

export function statLabel(key) {
  if (!key) return '';
  return key
    .split('_')
    .map((word) => (upperAcronyms.has(word) ? word.toUpperCase() : word.charAt(0).toUpperCase() + word.slice(1)))
    .join(' ');
}

export function formatStatLine(key, value) {
  const formatted = Number(value).toFixed(2).replace(/\.00$/, '').replace(/(\.\d)0$/, '$1');
  return `+${formatted} ${statLabel(key)}`;
}

const qualityNames = {
  0: 'Junk', 1: 'Common', 2: 'Uncommon', 3: 'Rare',
  4: 'Epic', 5: 'Legendary', 6: 'Artifact', 7: 'Heirloom',
};

export function qualityLabel(value) {
  return qualityNames[value] ?? `Unknown quality (${value ?? '—'})`;
}

export const socketColors = {
  0: 'Unknown', 1: 'Meta', 2: 'Red', 3: 'Blue', 4: 'Yellow',
  5: 'Green', 6: 'Orange', 7: 'Purple', 8: 'Prismatic',
};

// proto.SourceFilterOption values, in enum order. Value 0 ("Unknown source")
// is not offered in filter checkboxes; it is governed by the include-unknown
// toggle.
export const sourceKinds = [
  { value: 0, label: 'Unknown source', proto: 'SourceUnknown' },
  { value: 1, label: 'Crafting', proto: 'SourceCrafting' },
  { value: 2, label: 'Quest', proto: 'SourceQuest' },
  { value: 3, label: 'Reputation', proto: 'SourceReputation' },
  { value: 4, label: 'PvP', proto: 'SourcePvP' },
  { value: 5, label: 'Dungeon', proto: 'SourceDungeon' },
  { value: 6, label: 'Heroic dungeon', proto: 'SourceHeroicDungeon' },
  { value: 7, label: 'Raid', proto: 'SourceRaid' },
  { value: 8, label: 'Heroic raid', proto: 'SourceHeroicRaid' },
  { value: 9, label: 'Raid finder', proto: 'SourceRaidFinder' },
  { value: 10, label: 'Flexible raid', proto: 'SourceFlexibleRaid' },
  { value: 11, label: 'Sold by vendor', proto: 'SourceSoldByVendor' },
];

const sourceKindsByValue = new Map(sourceKinds.map((k) => [k.value, k.label]));
const sourceKindsByProto = new Map(sourceKinds.map((k) => [k.proto, k.label]));

export function sourceKindLabel(value) {
  return sourceKindsByValue.get(value) ?? `Unknown source (${value ?? '—'})`;
}

export function humanizeSourceKind(name) {
  if (!name) return '';
  return sourceKindsByProto.get(name) ?? humanizeEnum(name, 'Source');
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd ui-finder && rtk node --test src/lib/labels.test.js`
Expected: all tests PASS (the two existing `humanizeEnum` tests still pass).

- [ ] **Step 5: Commit**

```bash
git add ui-finder/src/lib/labels.js ui-finder/src/lib/labels.test.js
git commit -m "feat: shared label maps and stat humanizers in labels.js"
```

---

### Task 2: ItemTooltip component

**Files:**
- Create: `ui-finder/src/lib/ItemTooltip.svelte`
- Test: covered by e2e in Task 6 (visual DOM asserted there).

**Interfaces:**
- Consumes: `statLabel`, `formatStatLine`, `qualityLabel`, `socketColors` from `./labels.js` (Task 1).
- Produces: `ItemTooltip` with props `item` (object) and `variant` (`'full' | 'summary'`, default `'full'`). `full` expects the `GearSlotData` shape (`itemName`, `quality`, `ilvl`, `phase`, `slotName`, `stats`, `randomSuffix`, `sockets`, `socketBonus`, `enchant` — all optional). `summary` expects the `UIItemSummary` shape (`name`, `quality`, `phase`, `icon`). Renders nothing when the item is absent or has no name.

- [ ] **Step 1: Write the component**

Create `ui-finder/src/lib/ItemTooltip.svelte`:

```svelte
<script>
  import { formatStatLine, qualityLabel, socketColors } from './labels.js';

  let { item, variant = 'full' } = $props();

  let summary = $derived(variant === 'summary');
  let name = $derived(item?.itemName ?? item?.name ?? '');
  let statEntries = $derived(Object.entries(item?.stats ?? {}));
  let suffixEntries = $derived(item?.randomSuffix ? Object.entries(item.randomSuffix.stats ?? {}) : []);

  function gemText(gem) {
    const lines = Object.entries(gem?.stats ?? {}).map(([key, value]) => formatStatLine(key, value));
    return lines.length ? lines.join(', ') : (gem?.name ?? '');
  }
</script>

{#if name}
  <span class="item-tooltip" role="tooltip">
    {#if summary}
      <span class="tooltip-summary">
        <span class="tooltip-name quality-text-{item.quality ?? 0}">{name}</span>
        {#if item.phase}<span class="tooltip-phase">Phase {item.phase}</span>{/if}
      </span>
      <span class="tooltip-meta">{qualityLabel(item.quality)}</span>
    {:else}
      <span class="tooltip-header">
        <span class="tooltip-name quality-text-{item.quality ?? 0}">{name}</span>
        {#if item.phase}<span class="tooltip-phase">Phase {item.phase}</span>{/if}
      </span>
      {#if item.ilvl}<span class="tooltip-ilvl">Item Level {item.ilvl}</span>{/if}
      <span class="tooltip-meta">{item.slotName}</span>
      {#each statEntries as [key, value]}
        <span class="tooltip-stat">{formatStatLine(key, value)}</span>
      {/each}
      {#if item.randomSuffix}
        <span class="tooltip-meta">{item.randomSuffix.name}</span>
        {#each suffixEntries as [key, value]}
          <span class="tooltip-stat">{formatStatLine(key, value)}</span>
        {/each}
      {/if}
      {#if item.sockets?.length}
        <span class="tooltip-sockets">
          {#each item.sockets as socket}
            <span class="tooltip-socket">
              <span class="socket-dot socket-{socket.color}" title="{socketColors[socket.color] ?? 'Unknown'} socket"></span>
              {#if socket.gem}
                <span class="tooltip-gem">{gemText(socket.gem)}</span>
              {:else}
                <span class="tooltip-gem empty">{socketColors[socket.color] ?? 'Unknown'} socket (empty)</span>
              {/if}
            </span>
          {/each}
        </span>
      {/if}
      {#if item.socketBonus?.stats && Object.keys(item.socketBonus.stats).length}
        <span class="tooltip-socket-bonus" class:inactive={!item.socketBonus.active}>
          Socket Bonus: {Object.entries(item.socketBonus.stats).map(([key, value]) => formatStatLine(key, value)).join(', ')}
        </span>
      {/if}
      {#if item.enchant}
        <span class="tooltip-enchant">Equip: {item.enchant.description || item.enchant.name}</span>
      {/if}
    {/if}
  </span>
{/if}

<style>
  .item-tooltip {
    display: none;
    position: absolute;
    z-index: 30;
    top: 0;
    left: calc(100% + 10px);
    width: max-content;
    max-width: 280px;
    flex-direction: column;
    gap: 2px;
    padding: 8px 10px;
    background: #0a0d14;
    border: 1px solid #2b3444;
    border-radius: 6px;
    color: #f2f4f8;
    font-size: 0.85rem;
    line-height: 1.35;
    pointer-events: none;
    box-shadow: 0 6px 18px rgba(0, 0, 0, 0.55);
  }
  .tooltip-header { display: flex; justify-content: space-between; gap: 12px; }
  .tooltip-phase { color: #b8c0cc; white-space: nowrap; }
  .tooltip-ilvl { color: #e6b23c; }
  .tooltip-meta { color: #d5dae2; }
  .tooltip-stat { color: #ffffff; }
  .tooltip-sockets { display: flex; flex-direction: column; gap: 2px; }
  .tooltip-socket { display: flex; align-items: center; gap: 6px; }
  .tooltip-gem { color: #3fd13f; }
  .tooltip-gem.empty { color: #8a93a3; }
  .socket-dot { width: 10px; height: 10px; border-radius: 2px; display: inline-block; border: 1px solid rgba(255, 255, 255, 0.35); }
  .socket-dot.socket-1 { background: #4a3a6b; }
  .socket-dot.socket-2 { background: #b32424; }
  .socket-dot.socket-3 { background: #2456b3; }
  .socket-dot.socket-4 { background: #d6c520; }
  .socket-dot.socket-5 { background: #2e9e44; }
  .socket-dot.socket-6 { background: #d67a20; }
  .socket-dot.socket-7 { background: #7a2e9e; }
  .socket-dot.socket-8 { background: #c8ccd4; }
  .tooltip-socket-bonus { color: #3fd13f; }
  .tooltip-socket-bonus.inactive { color: #8a93a3; }
  .tooltip-enchant { color: #3fd13f; }
</style>
```


- [ ] **Step 2: Verify it compiles**

Run: `cd ui-finder && rtk npm run build`
Expected: Vite build succeeds (component not yet imported anywhere — bundler tree-shakes it; success means valid syntax).

- [ ] **Step 3: Commit**

```bash
git add ui-finder/src/lib/ItemTooltip.svelte
git commit -m "feat: ItemTooltip popover component with full and summary variants"
```

---

### Task 3: Armory tooltip in GearSlot

**Files:**
- Modify: `ui-finder/src/lib/GearSlot.svelte`
- Modify: `ui-finder/src/lib/app.css` (only if `.gear-icon-wrap` lacks `position: relative`)

**Interfaces:**
- Consumes: `ItemTooltip` (Task 2, `variant="full"`); `socketColors` from `./labels.js` (Task 1).
- Produces: gear cards whose icon block and item name are focusable tooltip triggers. `.gear-trigger` and `.name-trigger` classes wrap the popovers.

- [ ] **Step 1: Wire the shared socketColors and tooltip into GearSlot**

In `ui-finder/src/lib/GearSlot.svelte`:

1. Replace the private `socketColors` const (lines 6–16) with an import:

```js
  import { socketColors } from './labels.js';
  import ItemTooltip from './ItemTooltip.svelte';
```

2. Make the icon block a focusable trigger. Replace the `<div class="gear-icon-wrap">` opening (line 40) and its closing `</div>` (line 62) with:

```svelte
  <div class="gear-icon-wrap">
    <button type="button" class="gear-trigger" aria-describedby={slot.itemName ? 'tip-' + slot.slotName : undefined}>
      …existing ilvl / img / fallback / socket-strip markup unchanged…
      {#if slot.itemName}
        <ItemTooltip item={slot} variant="full" />
      {/if}
    </button>
  </div>
```

Keep the existing children (ilvl badge, icon/fallback, socket strip) inside the button; the `id` on the tooltip must be unique — pass a stable id prop if the component's root span doesn't accept one; simplest: drop `aria-describedby` and rely on the tooltip being inside the focusable trigger (screen readers read the button's accessible content). Prefer this simpler form unless review objects.

3. Make the name a second trigger. Wrap the `<h3 class="item-name …">` (line 65) in:

```svelte
      <span class="name-trigger">
        <h3 class="item-name quality-text-{slot.quality ?? 0}">{slot.itemName}</h3>
        <ItemTooltip item={slot} variant="full" />
      </span>
```

4. Add visibility CSS (component `<style>` block; create one if absent — check `app.css` first for existing `.gear-slot` styles and colocate there if that is the established pattern):

```css
  .gear-trigger, .name-trigger { position: relative; display: inline-flex; background: none; border: 0; padding: 0; cursor: help; }
  .gear-trigger:hover .item-tooltip, .gear-trigger:focus-visible .item-tooltip,
  .name-trigger:hover .item-tooltip, .name-trigger:focus-visible .item-tooltip { display: flex; }
```

For mirrored right-column cards (`article.mirror`), flip the side:

```css
  .mirror .item-tooltip { left: auto; right: calc(100% + 10px); }
```

5. Ensure `.gear-icon-wrap` (or `.gear-trigger`) has `position: relative` so `left: calc(100% …)` anchors to the trigger, not the page.

- [ ] **Step 2: Verify visually with the dev server**

Run: terminal 1 `rtk go run ./cmd/wowsimcli rank-upgrades --addr 127.0.0.1:43123 --no-browser`; terminal 2 `cd ui-finder && rtk npm run dev`; open `http://localhost:5173`, import the retribution fixture link from `cmd/wowsimcli/cmd/upgrades/testdata/retribution_no_settings_link.txt`, hover a gear icon.
Expected: dark popover beside the card with name, `Phase N`, `Item Level`, stats, gems with colored dots, socket bonus, green `Equip:` line; card content does not move. Tab focus shows the same popover. Right-column cards open the tooltip leftward.

- [ ] **Step 3: Commit**

```bash
git add ui-finder/src/lib/GearSlot.svelte ui-finder/src/lib/app.css
git commit -m "feat: full item tooltips on armory gear cards"
```

---

### Task 4: ReportView — shared maps, humanized filters, summary tooltip

**Files:**
- Modify: `ui-finder/src/lib/ReportView.svelte`

**Interfaces:**
- Consumes: `qualityLabel`, `sourceKindLabel`, `humanizeSourceKind` from `./labels.js` (Task 1); `ItemTooltip` (Task 2, `variant="summary"`).

- [ ] **Step 1: Swap private maps for shared ones**

In `ui-finder/src/lib/ReportView.svelte`:

1. Update imports:

```js
- Consumes: `qualityLabel`, `sourceKindLabel`, `humanizeSourceKind` from `./labels.js` (Task 1); `ItemTooltip` (Task 2, `variant="summary"`).
  import { humanizeEnum, qualityLabel, sourceKindLabel, humanizeSourceKind } from './labels.js';
  import ItemTooltip from './ItemTooltip.svelte';

2. Delete the private `qualityNames` const (lines 13–16) and the private `sourceKinds` const (lines 17–21), plus the now-unused `qualityLabel`/`sourceKindLabel` local function definitions (lines 27–32). Leave the private `slotNames` map untouched.
3. Replace the Source filters row (line 87) with humanized labels:

```svelte
    <div><dt>Source filters</dt><dd>{report.assumptions?.sourceKinds?.map(humanizeSourceKind).join(', ') || 'All sources'}{report.assumptions?.sourceNames?.length ? ` · ${report.assumptions.sourceNames.join(', ')}` : ''}</dd></div>
```

4. Wrap the report table item cell (line 133) with the summary tooltip trigger:

```svelte
              <span class="report-item-trigger">
                {upgrade.item?.name || 'Unknown'} ({upgrade.item?.id})
                {#if upgrade.item?.name}
                  <ItemTooltip item={upgrade.item} variant="summary" />
                {/if}
              </span>
            </td>
```

5. Add the trigger CSS (same colocated-style rule as Task 3):

```css
  .report-item-trigger { position: relative; cursor: help; }
  .report-item-trigger:hover .item-tooltip, .report-item-trigger:focus-within .item-tooltip { display: flex; }
```

- [ ] **Step 2: Verify unit tests and build**

Run: `cd ui-finder && rtk node --test src/lib/labels.test.js src/lib/api.test.js && rtk npm run build`
Expected: tests pass; Vite build succeeds.

- [ ] **Step 3: Commit**

```bash
git add ui-finder/src/lib/ReportView.svelte
git commit -m "feat: summary tooltips and humanized source filters in report"
```

---

### Task 5: Source-kind filter checkboxes in RankingPanel

**Files:**
- Modify: `ui-finder/src/lib/RankingPanel.svelte`

**Interfaces:**
- Consumes: `sourceKinds` from `./labels.js` (Task 1).
- Produces: ranking submit payload carries numeric `filters.sourceKinds` (array of checked `value`s; `[]` when none checked = all sources).

- [ ] **Step 1: Add state and submit wiring**

In `ui-finder/src/lib/RankingPanel.svelte`:

1. Import and add state (after the existing `let confirmationIterations = $state(1000);` line 8):

```js
  import { sourceKinds } from './labels.js';
```
```js
  let selectedKinds = $state([]);
```

2. Reset on new import inside the `$effect` block (after line 19 `confirmationIterations = …;`):

```js
    selectedKinds = [];
```

3. In `submitRanking`, replace `"sourceKinds: []"` (line 49) with:

```js
        sourceKinds: [...selectedKinds],
```

- [ ] **Step 2: Add the checkbox group to the template**

Insert after the `includeUnknown` check-control div (after line 82) inside `.control-grid`:

```svelte
      <fieldset class="check-control source-kind-group" disabled={busy}>
        <legend>Include sources</legend>
        {#each sourceKinds.filter((kind) => kind.value !== 0) as kind}
          <label>
            <input
              type="checkbox"
              checked={selectedKinds.includes(kind.value)}
              onchange={(event) => {
                selectedKinds = event.currentTarget.checked
                  ? [...selectedKinds, kind.value]
                  : selectedKinds.filter((v) => v !== kind.value);
              }}
            />
            {kind.label}
          </label>
        {/each}
      </fieldset>
```

With no boxes checked the payload is `sourceKinds: []` — the server treats it as all sources, matching the "none checked = all" spec semantics. Add minimal CSS in the established style location:

```css
  .source-kind-group { display: flex; flex-wrap: wrap; gap: 4px 12px; border: 0; padding: 0; margin: 0; }
  .source-kind-group legend { float: left; }
```

- [ ] **Step 3: Verify unit tests still pass and build**

Run: `cd ui-finder && rtk node --test src/lib/api.test.js && rtk npm run build`
Expected: pass.

- [ ] **Step 4: Commit**

```bash
git add ui-finder/src/lib/RankingPanel.svelte
git commit -m "feat: source-kind filter checkboxes with numeric enum submit"
```

---

### Task 6: E2E coverage, rebuild, docs

**Files:**
- Modify: `ui-finder/e2e/armory.spec.js`
- Rebuild: `cmd/wowsimcli/cmd/upgrade_ui/` (committed bundle)
- Modify: `docs/upgrade-finder.md` (workflow/results wording only)

**Interfaces:**
- Consumes: everything from Tasks 1–5.
- Produces: green e2e suite; rebuilt embedded bundle; doc statements matching behavior.

- [ ] **Step 1: Add e2e assertions**

In `ui-finder/e2e/armory.spec.js`, after the gear-tab assertions (after line 44, still inside the single test before the ranking section), insert:

```js
  // Full tooltip on the armory (hover the first gear icon).
  const firstTrigger = page.locator('.gear-trigger').first();
  await firstTrigger.hover();
  const tooltip = page.locator('.item-tooltip').first();
  await expect(tooltip).toBeVisible();
  await expect(tooltip).toContainText(/Item Level/i);
  await expect(tooltip.locator('.tooltip-enchant, .tooltip-socket').first()).toBeAttached();

  // Summary tooltip in the report appears after ranking; asserted below.
```

Then in the ranking section (after `await confirmation.fill('1');`, line 54), check every source-kind box before starting ranking, so the report's Source filters row shows humanized labels while the candidate set stays effectively unfiltered (all kinds included):

```js
  const kindBoxes = page.locator('.source-kind-group input[type="checkbox"]');
  const kindCount = await kindBoxes.count();
  for (let index = 0; index < kindCount; index += 1) {
    await kindBoxes.nth(index).check();
  }
```

And after the report table becomes visible (after line 60), assert:

```js
  await expect(page.getByText('Source filters')).toBeVisible();
  await expect(page.locator('.report-summary dd').filter({ hasText: 'Crafting' })).toBeVisible();
  const reportTrigger = page.locator('.report-item-trigger').first();
  await reportTrigger.hover();
  await expect(page.locator('.item-tooltip').first()).toBeVisible();
```

- [ ] **Step 2: Run the e2e suite**

Run: `cd ui-finder && rtk npm run test:e2e`
Expected: 1 Chromium test passed (the existing single-test suite with new assertions).

- [ ] **Step 3: Rebuild the embedded bundle and verify the binary serves it**

Run: `cd ui-finder && rtk npm run build && cd .. && go build -o wowsimcli ./cmd/wowsimcli && ./wowsimcli rank-upgrades --addr 127.0.0.1:43199 --no-browser` and open the printed URL.
Expected: the production (embedded) UI shows tooltips and the source-kind checkboxes — not just the dev server.

- [ ] **Step 4: Update docs**

In `docs/upgrade-finder.md`:
- In "Armory review", after the enchant/gems sentence, add: "Hovering or keyboard-focusing an item icon or name shows a tooltip with the item's stats, gems, socket bonus, and enchant."
- In "Workflow" step 3, add: "Optionally restrict ranked sources by checking the desired source kinds; with none checked, all sources are ranked."

- [ ] **Step 5: Commit everything**

```bash
git add ui-finder/e2e/armory.spec.js cmd/wowsimcli/cmd/upgrade_ui docs/upgrade-finder.md
git commit -m "feat: e2e coverage for tooltips and source filters; rebuild embedded bundle"
```

- [ ] **Step 6: Final verification (full required-checks set)**

Run from repo root:

```bash
go test ./cmd/wowsimcli/cmd/upgrades -count=1
go test ./cmd/wowsimcli/cmd -count=1
cd ui-finder && rtk node --test src/lib/labels.test.js && rtk node --test src/lib/api.test.js
```

Expected: all pass. Go suites should be untouched by these UI changes — their passing confirms the no-Go-changes constraint held.
