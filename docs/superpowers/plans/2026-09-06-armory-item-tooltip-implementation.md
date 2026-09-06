# Armory Item Tooltip Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Goal:** Deliver the reference-style item tooltip with an external item icon, actual equipped gem icons, ordered content, and accessible viewport-aware behavior.

**Architecture:** Extend the existing armory response with available catalog metadata. Format data in a pure JavaScript module, render it through a content component, and coordinate all hover/focus triggers through one app-scoped tooltip controller and floating layer. Keep full armory tooltips and report summaries as distinct data variants.

**Tech Stack:** Existing Go, Svelte 5, Vite, Node test runner, and Playwright. No new runtime dependency.

**Design reference:** [Armory item tooltip redesign](2026-09-06-armory-item-tooltip-redesign.md). This file supersedes its high-level implementation sequence. It does not expand the agreed core scope to unavailable reference metadata.

**Delivery status:** Plan only. Unchecked tasks describe future implementation, not completed changes or verified test results.

## Global constraints

- Preserve imported gear, gems, enchants, simulator stats, and ranking behavior.
- Render actual equipped gems, even when their colors do not match their sockets.
- Keep item base stats, random-suffix stats, gem effects, enchants, and socket bonuses distinct; never add them twice.
- Missing metadata must not become invented values. Item level is not required character level; character level is not item required level.
- Keep hover rendering immediate and local; use existing icon URLs with accessible image fallbacks.
- Treat the seven user screenshots as visual references, not sources for item database values.
- Preserve report summary tooltips while giving them the shared visual shell; full candidate tooltips are a separate scope because report data is only a summary.
- Prefix shell commands with `rtk`. Commands below run from the repository root unless a different directory is specified.
- Do not change the simulator, protobuf schema, item database, ranking filters, or armory layout as part of this work.
- Complete and verify each task before proceeding to its dependents. Commit only files belonging to the task when committing implementation milestones.

## File and dependency map

| Task | Files | Responsibility |
| --- | --- | --- |
| 1 | `cmd/wowsimcli/cmd/upgrades/{types.go,armory.go,armory_test.go}` | Additional catalog-backed tooltip fields |
| 2 | New `ui-finder/src/lib/{itemTooltip.js,itemTooltip.test.js,tooltipLabels.js}` | Stable content model and enum/stat labels |
| 3 | New `ui-finder/src/lib/{tooltipPosition.js,tooltipPosition.test.js,tooltipController.js,tooltipController.test.js}` | Placement and independently testable interaction state |
| 4 | `ItemTooltip.svelte`; new `ItemTooltipLayer.svelte`, `TooltipIcon.svelte` under `ui-finder/src/lib/` | Reference-style content, reusable image fallback, floating DOM layer |
| 5 | New `ItemTooltipTrigger.svelte`; modify `App.svelte`, `GearSlot.svelte`, `ReportView.svelte`, `app.css` | App-scoped controller, trigger integration, touch/keyboard details |
| 6 | New `ui-finder/e2e/item-tooltip.spec.js`; modify `armory.spec.js`, `ui-finder/package.json`, `.github/workflows/upgrade_finder.yml`, `docs/upgrade-finder.md`; rebuild `cmd/wowsimcli/cmd/upgrade_ui/` | Focused browser coverage, CI, documentation, embedded delivery |

Tasks 1 and 3 are independent. Task 2 consumes Task 1's field definitions. Task 4 consumes Tasks 2 and 3; Task 5 integrates Task 4. Task 6 verifies the integrated result. Prefer sequential execution unless separate agents have non-overlapping ownership.

## Task 1: Expose existing catalog metadata

**Consumes:** `proto.UIItem` getters in `enrichItem`. **Produces:** Additive JSON fields on `GearSlotData`. Existing fields and meanings remain intact.

- [ ] Add a focused Go test to `armory_test.go`, using the existing imports and helpers:

```go
func TestEnrichItemExposesTooltipMetadataWithoutMutatingSpec(t *testing.T) {
    spec := &proto.ItemSpec{Id: 900001}
    before := mustMarshal(t, spec)
    item := &proto.UIItem{
        Id: 900001, Name: "Tooltip fixture", ArmorType: proto.ArmorType_ArmorTypePlate,
        WeaponType: proto.WeaponType_WeaponTypeSword,
        HandType: proto.HandType_HandTypeTwoHand,
        WeaponDamageMin: 100, WeaponDamageMax: 200, WeaponSpeed: 3.5,
        ClassAllowlist: []proto.Class{proto.Class_ClassPaladin},
        RequiredProfession: proto.Profession_Blacksmithing, Unique: true,
    }
    got, err := enrichItem(spec, item, NewCatalog(database.Load()))
    if err != nil { t.Fatal(err) }
    if got.ArmorType != item.ArmorType || got.WeaponType != item.WeaponType ||
        got.HandType != item.HandType || got.WeaponDamageMin != 100 ||
        got.WeaponDamageMax != 200 || got.WeaponSpeed != 3.5 ||
        !reflect.DeepEqual(got.ClassAllowlist, item.ClassAllowlist) ||
        got.RequiredProfession != item.RequiredProfession || !got.Unique {
        t.Fatalf("tooltip metadata: %#v", got)
    }
    if !bytes.Equal(before, mustMarshal(t, spec)) { t.Fatal("spec mutated") }
}
```

This synthetic item tests transport fields, not an equip-valid real item. Add a ranged-type case and extend the existing populated-socket assertion to compare gem ID, icon, and stats with `targetGem`.

- [ ] Run `rtk go test ./cmd/wowsimcli/cmd/upgrades -run 'TestEnrichItemExposesTooltipMetadata' -count=1` and confirm the new field assertions fail before implementing.
- [ ] Add these fields to `GearSlotData`:

```go
ArmorType          proto.ArmorType        `json:"armorType"`
WeaponType         proto.WeaponType       `json:"weaponType"`
HandType           proto.HandType         `json:"handType"`
RangedWeaponType   proto.RangedWeaponType `json:"rangedWeaponType"`
WeaponDamageMin    float64                `json:"weaponDamageMin"`
WeaponDamageMax    float64                `json:"weaponDamageMax"`
WeaponSpeed        float64                `json:"weaponSpeed"`
ClassAllowlist     []proto.Class          `json:"classAllowlist"`
RequiredProfession proto.Profession      `json:"requiredProfession"`
Unique             bool                  `json:"unique"`
```

- [ ] Assign the fields after the existing item metadata assignments in `enrichItem`:

```go
data.ArmorType = item.GetArmorType()
data.WeaponType = item.GetWeaponType()
data.HandType = item.GetHandType()
data.RangedWeaponType = item.GetRangedWeaponType()
data.WeaponDamageMin = item.GetWeaponDamageMin()
data.WeaponDamageMax = item.GetWeaponDamageMax()
data.WeaponSpeed = item.GetWeaponSpeed()
data.ClassAllowlist = append([]proto.Class(nil), item.GetClassAllowlist()...)
data.RequiredProfession = item.GetRequiredProfession()
data.Unique = item.GetUnique()
```

- [ ] Run `rtk gofmt -w cmd/wowsimcli/cmd/upgrades/types.go cmd/wowsimcli/cmd/upgrades/armory.go cmd/wowsimcli/cmd/upgrades/armory_test.go`, then `rtk go test ./cmd/wowsimcli/cmd/upgrades -count=1`. Verify the API JSON field spellings with a marshaling assertion if the HTTP tests do not already exercise them.
- [ ] Review checkpoint: new fields are additive; no source stat/gem/enchant calculations changed. Suggested commit: `feat: expose armory tooltip metadata`.

## Task 2: Implement a pure tooltip content model

**Consumes:** Existing `GearSlotData`, Task 1's fields, or existing `UIItemSummary` for summary mode.

**Produces:** `buildItemTooltip(item, variant = 'full')` in `itemTooltip.js`. Return exactly:

```js
// TextLine: { key: string, text: string }
// SocketView: { color: number, gem: object|null, text: string }
// Every array exists, even when empty. Missing text is ''.
{
  name, icon, quality, phase, ilvl, qualityLabel,
  slotLabel, typeLabel, suffixLabel,
  weaponLines: [], baseLines: [], enchantLine: '', sockets: [],
  socketBonus: null, // or { text: string, active: boolean }
  restrictionLines: [], equipLines: [], setName: ''
}
```

- [ ] Add Node tests using `node:test` and `node:assert/strict`. Include this concrete regression case:

```js
test('orders item stats without folding in gems or enchants', () => {
  const input = {
    itemName: 'Chest fixture', slotName: 'Chest', armorType: 4,
    stats: { melee_crit_rating: 31, intellect: 31, strength: 56, armor: 1825, stamina: 48 },
    randomSuffix: { name: 'of Strength', stats: { strength: 2 } },
    enchant: { description: '+6 All Stats', stats: { strength: 6 } },
    sockets: [{ color: 2, gem: { id: 1, name: 'Gem fixture', icon: 'gem_fixture', stats: { strength: 10 } } }],
    socketBonus: { stats: { strength: 4 }, active: false },
  };
  const before = structuredClone(input);
  const result = buildItemTooltip(input);
  assert.deepEqual(result.baseLines.map(x => x.text),
    ['1825 Armor', '+58 Strength', '+48 Stamina', '+31 Intellect']);
  assert.equal(result.typeLabel, 'Plate');
  assert.equal(result.enchantLine, '+6 All Stats');
  assert.equal(result.sockets[0].gem.icon, 'gem_fixture');
  assert.equal(result.socketBonus.active, false);
  assert.deepEqual(result.equipLines.map(x => x.text),
    ['Equip: Improves melee critical strike rating by 31.']);
  assert.deepEqual(input, before);
});
```

- [ ] Run `rtk node --test ui-finder/src/lib/itemTooltip.test.js` before creating the formatter.
- [ ] Create `tooltipLabels.js` with numeric maps matching the checked-in `proto/common.proto`. Zero means no label; an unknown nonzero value displays `Unknown type (N)` or the equivalent restriction label. Do not use array index assumptions for class IDs.

```js
export const armorTypes = { 1: 'Cloth', 2: 'Leather', 3: 'Mail', 4: 'Plate' };
export const weaponTypes = { 1: 'Axe', 2: 'Dagger', 3: 'Fist Weapon', 4: 'Mace', 5: 'Held In Off-hand', 6: 'Polearm', 7: 'Shield', 8: 'Staff', 9: 'Sword' };
export const handTypes = { 1: 'Main Hand', 2: 'One-Hand', 3: 'Off Hand', 4: 'Two-Hand' };
export const rangedTypes = { 1: 'Bow', 2: 'Crossbow', 3: 'Gun', 4: 'Thrown', 5: 'Wand', 6: 'Idol', 7: 'Libram', 8: 'Totem', 9: 'Sigil' };
export const classes = { 1: 'Warrior', 2: 'Paladin', 3: 'Hunter', 4: 'Rogue', 5: 'Priest', 7: 'Shaman', 8: 'Mage', 9: 'Warlock', 11: 'Druid' };
export const professions = { 1: 'Alchemy', 2: 'Blacksmithing', 3: 'Enchanting', 4: 'Engineering', 5: 'Herbalism', 6: 'Inscription', 7: 'Jewelcrafting', 8: 'Leatherworking', 9: 'Mining', 10: 'Skinning', 11: 'Tailoring' };
```

- [ ] Implement stable ordering and tooltip-only formatting. Do not change `formatStatLine` globally, since stats and other views depend on it. Base order is armor, bonus armor, Strength, Agility, Stamina, Intellect, Spirit, Health, Mana, then elemental resistances. Known remaining engine keys are green equip lines; unknown keys are preserved as white lines at the end in lexical order.

```js
const baseOrder = ['armor', 'bonus_armor', 'strength', 'agility', 'stamina',
  'intellect', 'spirit', 'health', 'mana', 'arcane_resistance',
  'fire_resistance', 'frost_resistance', 'nature_resistance', 'shadow_resistance'];
const equipOrder = ['attack_power', 'ranged_attack_power', 'feral_attack_power',
  'healing_power', 'spell_damage', 'arcane_damage', 'fire_damage', 'frost_damage',
  'holy_damage', 'nature_damage', 'shadow_damage', 'physical_damage',
  'melee_hit_rating', 'spell_hit_rating', 'melee_crit_rating', 'spell_crit_rating',
  'melee_haste_rating', 'spell_haste_rating', 'expertise_rating',
  'armor_penetration', 'spell_penetration', 'defense_rating', 'block_rating',
  'block_value', 'dodge_rating', 'parry_rating', 'resilience_rating', 'mp5'];
const numberText = value => Number(value).toFixed(2).replace(/\.?0+$/, '');
// Use toFixed(2) before stripping zeros; do not apply this regex to raw integers.
const signedText = value => `${value >= 0 ? '+' : ''}${numberText(value)}`;
const ratingNames = {
  melee_hit_rating: 'melee hit rating', spell_hit_rating: 'spell hit rating',
  melee_crit_rating: 'melee critical strike rating', spell_crit_rating: 'spell critical strike rating',
  melee_haste_rating: 'melee haste rating', spell_haste_rating: 'spell haste rating',
};
// For positive ratingNames entries use:
// `Equip: Improves ${ratingNames[key]} by ${numberText(value)}.`
// For other equip keys and negative values use:
// `Equip: ${signedText(value)} ${statLabel(key)}`
```

Use the existing `statLabel` for generic labels. Preserve the current distinction between `healing_power` and `spell_damage`; do not combine separate engine fields into an assumed shared bonus. Filter zero/non-finite numbers and keep negative values correctly signed.

- [ ] Build item stat lines from a new map containing base plus random-suffix values once. Preserve `suffixLabel`. Leave gem/enchant/bonus values in their sections. Resolve name from `itemName ?? name`; summary mode must stop before reading full-item fields.
- [ ] Map sockets in their original order. Keep `gem` and its identity unchanged, compute effect text from ordered gem stats, and fall back to the gem name if no numeric effects exist. Use `socketColors` from `labels.js` for an empty socket label. Bonus text uses only its own stats and the supplied `active` flag.
- [ ] For slot/type, prefer hand type on the left for weapon items, otherwise `slotName`. Type precedence is ranged type, weapon type, armor type. Show weapon damage and speed only when speed is positive and the damage range is valid. Display no invented weapon DPS or durability values. Restriction lines are `Unique`, `Classes: ...`, and `Requires ...` for known nonzero values.
- [ ] Add tests for empty sockets, missing gem stats, unknown stat/type keys, negative stats, zero values, a weapon, all restriction maps, and summary mode. Example summary assertion:

```js
test('summary never invents full-item data', () => {
  const vm = buildItemTooltip({ name: 'Candidate', quality: 4, phase: 3 }, 'summary');
  assert.equal(vm.name, 'Candidate');
  assert.deepEqual(vm.baseLines, []);
  assert.deepEqual(vm.sockets, []);
  assert.equal(vm.enchantLine, '');
});
```

- [ ] Re-run the Node file. Review checkpoint: this module has no DOM, imports no app state, and never mutates its input. Suggested commit: `feat: normalize item tooltip content`.

## Task 3: Implement placement and interaction independently

**Produces:** `positionTooltip(anchor, size, viewport, preferredSide)` in `tooltipPosition.js`; `createTooltipController(onChange, timing = {})` and `ITEM_TOOLTIP_CONTEXT` in `tooltipController.js`.

Placement coordinates use CSS pixels in the same coordinate space as `getBoundingClientRect`. Viewport is `{ left, top, width, height }`, including visual-viewport offsets when present. Size covers the full external-icon-plus-panel group.

- [ ] Add placement tests before implementation:

```js
test('flips at the right edge and clamps vertically', () => {
  const result = positionTooltip(
    { left: 900, right: 954, top: 650 },
    { width: 363, height: 300 },
    { left: 0, top: 0, width: 1000, height: 800 }, 'right');
  assert.deepEqual(result, { left: 527, top: 492 });
});
```

- [ ] Implement the pure placement function:

```js
export function positionTooltip(anchor, size, viewport, preferredSide = 'right') {
  const margin = 8, gap = 10;
  const minX = viewport.left + margin, minY = viewport.top + margin;
  const maxX = Math.max(minX, viewport.left + viewport.width - margin - size.width);
  const maxY = Math.max(minY, viewport.top + viewport.height - margin - size.height);
  const candidates = { right: anchor.right + gap, left: anchor.left - size.width - gap };
  const alternate = preferredSide === 'right' ? 'left' : 'right';
  const fits = x => x >= minX && x <= maxX;
  const x = fits(candidates[preferredSide]) ? candidates[preferredSide]
    : fits(candidates[alternate]) ? candidates[alternate] : candidates[preferredSide];
  return { left: Math.min(maxX, Math.max(minX, x)),
    top: Math.min(maxY, Math.max(minY, anchor.top)) };
}
```

The layer must cap its size before measuring: this algorithm cannot make oversized content smaller. Test both preferences, left/top/bottom edges, and nonzero viewport offsets.

- [ ] Implement the controller with these exact callable methods:

```js
// onChange receives null (closed), or:
// { owner, id, item, variant, preferredSide, anchor }
// owner is a Symbol per wrapper; anchor is the actual hovered/focused element.
// timing defaults: { setTimer: setTimeout, clearTimer: clearTimeout, delay: 120 }
createTooltipController(onChange, timing);
// Returned methods:
// enter(record, source) source is 'pointer' or 'focus'
// leave(owner, source)
// enterPanel(owner)
// leavePanel(owner)
// dismiss()
// release(owner) called when owner is destroyed or its item object changes
// destroy()
export const ITEM_TOOLTIP_CONTEXT = Symbol('item-tooltip');
```

Use a single pending timer per open/close purpose, and an incrementing generation to invalidate stale callbacks. Track pointer ownership, focus ownership, panel hover, and dismissal suppression separately. A new owner closes the visible old owner immediately; its own hover delay still applies. Focus wins over a pending hover and opens immediately.

| Event | Required result |
| --- | --- |
| Pointer enters non-touch trigger | Open after 120ms if still eligible |
| Focus enters trigger | Cancel competing timers and open immediately |
| Pointer/focus leaves | Close after 120ms only when neither trigger input nor panel hover keeps it open |
| Pointer moves between two triggers of same item | Keep one panel; use the latest anchor without a duplicate instance |
| Pointer enters panel | Cancel close timer |
| Escape | Close immediately; suppress currently active trigger sources until they leave and re-enter |
| Other owner opens | Old owner cannot close/reopen the new panel from a stale callback |
| Item object changes, owner unmounts, or app tears down | Close associated panel and cancel associated timers/listeners |

- [ ] Add controller tests with an injected timer queue instead of wall-clock sleeps. A starting test:

```js
test('focus opens immediately and release closes', () => {
  const changes = [];
  const controller = createTooltipController(value => changes.push(value));
  const owner = Symbol('chest');
  const record = { owner, id: 'chest-tip', item: { itemName: 'Chest' },
    variant: 'full', preferredSide: 'right', anchor: {} };
  controller.enter(record, 'focus');
  assert.equal(changes.at(-1)?.owner, owner);
  controller.release(owner);
  assert.equal(changes.at(-1), null);
  controller.destroy();
});
```

Additional cases must follow the transition table, especially Escape while focus remains, moving to another owner before the old delay expires, and retained panel hover.

- [ ] Run `rtk node --test ui-finder/src/lib/tooltipPosition.test.js ui-finder/src/lib/tooltipController.test.js`. Suggested commit: `feat: add tooltip placement and interaction controller`.

## Task 4: Render the visual panel and floating layer

**Consumes:** Task 2's model and Task 3's controller record/placement. **Produces:** `ItemTooltip` content, `TooltipIcon` fallbacks, and `ItemTooltipLayer` positioned DOM.

- [ ] Create `TooltipIcon.svelte` with props `{ icon, name, size = 16 }`. Use the existing Zamimg URL convention. Render an image when available; on missing/failed image render a same-size framed fallback with `role="img"` and `aria-label="{name} icon unavailable"`. Key the stateful image branch by URL or reset failures when the URL changes. Never hide gem text after an image failure.
- [ ] Refactor `ItemTooltip.svelte` to accept `{ item, variant = 'full', id }` and derive its content through `buildItemTooltip`. Its root is `class="item-tooltip" role="tooltip" {id}` and is always visible when mounted. Remove all positioning, hover visibility rules, and the old `Equip:` enchant prefix from this component.
- [ ] Use this skeleton for socket content; retain each row's gem identity and add deterministic test hooks:

```svelte
{#each vm.sockets as socket, index (index)}
  <div class="tooltip-socket" data-gem-id={socket.gem?.id ?? ''}>
    {#if socket.gem}
      <TooltipIcon icon={socket.gem.icon} name={socket.gem.name} size={16} />
      <span class="tooltip-gem">{socket.text}</span>
    {:else}
      <span class="empty-socket-icon socket-{socket.color}" aria-hidden="true"></span>
      <span class="tooltip-muted">{socket.text}</span>
    {/if}
  </div>
{/each}
```

- [ ] Render all sections in the design's order. Wrap optional sections only when populated so missing content leaves no blank separators. Keep phase `white-space: nowrap`, the title `min-width: 0`, and the entire tooltip `text-align: left`. Use native text interpolation, not raw HTML.
- [ ] Apply these starting visual tokens and adjust only after visual inspection:

```css
.item-tooltip { display: grid; grid-template-columns: 38px minmax(0, 320px); gap: 5px; text-align: left; }
.tooltip-body { box-sizing: border-box; min-width: 0; padding: 8px 10px;
  color: #f2f2f4; background: rgba(18, 21, 38, .98); border: 1px solid #85848b;
  border-radius: 3px; box-shadow: 0 4px 14px #0008; font-size: 13px; line-height: 1.25; }
.tooltip-header { display: grid; grid-template-columns: minmax(0, 1fr) auto; gap: 12px; }
.tooltip-name { font-size: 15px; overflow-wrap: anywhere; }
.tooltip-phase { color: #aaaab4; white-space: nowrap; }
.tooltip-ilvl, .tooltip-set { color: #ffd100; }
.tooltip-enchant, .tooltip-gem, .tooltip-equip, .tooltip-socket-bonus { color: #1eff00; }
.tooltip-muted, .tooltip-socket-bonus.inactive { color: #9d9da8; }
.tooltip-section + .tooltip-section { margin-top: 8px; }
.tooltip-socket { display: flex; align-items: flex-start; gap: 4px; }
```

Keep the item icon outside the body at top-left on both placement sides. At viewport widths below 420px use a single-column footprint and place that same 38px icon inside the header. This avoids hiding the external icon offscreen or duplicating accessible images. Cap total width to viewport width minus 16px and cap body height to viewport height minus 16px; make body overflow scrollable when necessary.

- [ ] Create `ItemTooltipLayer.svelte` accepting `{ active, controller }`. Mount only when `active` exists and `active.anchor.isConnected`; key content by active owner/item so stale icon state cannot survive an item change. Use a Svelte action that appends the floating root to `document.body` and removes it on destroy, allowing the component to keep normal Svelte ownership while escaping clipping ancestors.

```js
function portal(node) {
  document.body.appendChild(node);
  return { destroy() { node.remove(); } };
}
```

- [ ] Give the layer `position: fixed; z-index: 1000`. Initially measure with `visibility: hidden`, then apply `positionTooltip` and reveal. Use `ResizeObserver` on the layer and anchor, capture-phase document scroll, window resize, and visualViewport scroll/resize. Coalesce remeasurements using one animation frame. Reposition after wrapping/icon/content changes. Clean up observers, listeners, and scheduled frames when the active record changes or closes. Release the owner if its anchor is detached.
- [ ] Wire pointer entry/exit on the layer to `enterPanel`/`leavePanel`. Add one Escape listener while open that calls `dismiss`. Keep the wrapper gap inside the hovered layer so moving toward the floating item icon does not immediately close the panel.
- [ ] Build with `rtk npm --prefix ui-finder run build` to catch Svelte errors. Browser interaction assertions land after Task 5 integration; do not claim visual parity from compilation alone. Suggested commit: `feat: render reference-style item tooltip panel`.

## Task 5: Integrate icon/name/report triggers and Details fallback

**Consumes:** Controller, content, and floating layer. **Produces:** One active tooltip across the app, with icon/name trigger parity and no clipped report panel.

- [ ] Set up the controller once in `ui-finder/src/App.svelte`:

```svelte
<script>
  import { setContext, onDestroy } from 'svelte';
  import { createTooltipController, ITEM_TOOLTIP_CONTEXT } from './lib/tooltipController.js';
  import ItemTooltipLayer from './lib/ItemTooltipLayer.svelte';
  let activeTooltip = $state(null);
  const tooltipController = createTooltipController(value => { activeTooltip = value; });
  setContext(ITEM_TOOLTIP_CONTEXT, tooltipController);
  onDestroy(() => tooltipController.destroy());
</script>

<ItemTooltipLayer active={activeTooltip} controller={tooltipController} />
```

Merge these imports and statements into the existing script; do not create a second script block. The layer can follow the existing footer.

- [ ] Create `ItemTooltipTrigger.svelte` with props `{ item, variant = 'full', preferredSide = 'right', children }`. It emits no wrapper DOM. Allocate `owner = Symbol()` and a stable unique tooltip ID per instance. Get the controller from context. Render the children snippet with `{ id, onpointerenter, onpointerleave, onfocus, onblur }`; handlers construct the controller record using `event.currentTarget` as `anchor`. Ignore touch pointer entry. Release on unmount and when the item object changes; do not auto-open replacement content during import.
- [ ] Wrap each occupied `GearSlot` article in one `ItemTooltipTrigger`, reusing its snippet handlers on both the existing icon button and a new name button inside the existing `h3`. Preserve the existing article, socket strip, quality classes, ilvl badge, and mirroring. Prefer right placement for the left column and left placement for the right column. Example of the name's markup inside the snippet:

```svelte
<h3 class="item-name quality-text-{slot.quality ?? 0}">
  <button type="button" class="name-trigger" aria-describedby={trigger.id}
    onpointerenter={trigger.onpointerenter} onpointerleave={trigger.onpointerleave}
    onfocus={trigger.onfocus} onblur={trigger.onblur}>{slot.itemName}</button>
</h3>
```

Use the identical handler/description binding on `.gear-trigger`. Keep empty slots outside tooltip wrappers. Remove both old nested `ItemTooltip` instances. If using `aria-describedby` only while open, expose an `open` snippet field derived from the active owner; otherwise the ID resolves as soon as the panel mounts on focus.

- [ ] Replace each occupied report name span with a summary-mode wrapper and button. Preserve displayed item ID, phase, quality, and ranking content. Do not pass equipped armory data to a candidate tooltip.
- [ ] Remove the `.gear-trigger:hover .item-tooltip`, `.name-trigger:hover .item-tooltip`, `.mirror .item-tooltip`, and report descendant visibility rules from `app.css`. Retain trigger style resets and add a visible `:focus-visible` outline. Confirm no remaining rule overrides the new panel visibility or left alignment.
- [ ] Expand the existing gear Details disclosure with any information newly available only in the tooltip: type, weapon lines, individual gem names/effects, restrictions, and separate suffix stats. Reuse `buildItemTooltip` for text, preserve the existing item ID and raw-stat entries, and keep this disclosure keyboard/touch usable. Do not mount a second `role="tooltip"` inside Details.
- [ ] Update existing `armory.spec.js` assertions to target `page.getByRole('tooltip')` and the trigger's description ID rather than looking for a descendant. Example:

```js
await firstTrigger.hover();
const tooltip = page.getByRole('tooltip');
await expect(tooltip).toBeVisible();
await expect(tooltip).toHaveAttribute('id', await firstTrigger.getAttribute('aria-describedby'));
await expect(tooltip).toContainText('Item Level');
```

- [ ] Build the embedded UI and run the existing browser smoke test before adding new assertions: `rtk npm --prefix ui-finder run build`, then `rtk npm --prefix ui-finder run test:e2e -- armory.spec.js`. The server must be newly started by Playwright, since the Go embed is a build-time snapshot. Suggested commit: `feat: integrate accessible shared item tooltips`.

## Task 6: Add regression coverage and deliver the embedded build

**Files:** `ui-finder/e2e/item-tooltip.spec.js`, existing smoke test, package scripts, CI, documentation, generated UI assets.

- [ ] Use a real existing imported-link fixture for the API envelope and replace only selected gear in the browser response with controlled display fixtures. The response has top-level `gear`, as used by `ArmoryView.svelte`; do not assume an `armory.gear` object. Example fixture setup:

```js
await page.route('**/api/import', async route => {
  const response = await route.fetch();
  const payload = await response.json();
  const chest = payload.gear.find(slot => slot.slotName === 'Chest');
  Object.assign(chest, {
    itemName: 'Tooltip Chest Fixture', icon: 'fixture_chest', quality: 4,
    ilvl: 146, phase: 3, armorType: 4,
    stats: { armor: 1825, strength: 56, stamina: 48, intellect: 31, melee_crit_rating: 31 },
    randomSuffix: null, enchant: { description: '+6 All Stats' },
    sockets: [
      { color: 2, gem: { id: 101, name: 'Red Gem Fixture', icon: 'fixture_red', stats: { strength: 10 } } },
      { color: 4, gem: { id: 102, name: 'Orange Gem Fixture', icon: 'fixture_orange', stats: { strength: 5, melee_crit_rating: 5 } } },
      { color: 3, gem: { id: 103, name: 'Purple Gem Fixture', icon: 'fixture_purple', stats: { stamina: 7, strength: 5 } } },
    ],
    socketBonus: { stats: { spell_damage: 5, healing_power: 5 }, active: true },
  });
  await route.fulfill({ response, json: payload });
});
```

These synthetic display values are not database truth and must never enter production assets or simulator requests. Import the fixture through the real form, as the existing smoke test does.

- [ ] Stub icon URLs with deterministic locally defined SVG bodies in Playwright route handlers for automated layout checks. Abort selected requests for error cases. Separately inspect real gem art in a manual run; color blocks used for deterministic tests are not proof that the production icons look correct.
- [ ] Add exact content regressions:

```js
const chest = page.locator('[data-slot="Chest"]');
await chest.locator('.name-trigger').hover();
const tip = page.getByRole('tooltip');
await expect(tip).toHaveCount(1);
await expect(tip.getByRole('img', { name: 'Tooltip Chest Fixture', exact: true })).toBeVisible();
await expect(tip.locator('.tooltip-socket img')).toHaveCount(3);
await expect(tip.locator('[data-gem-id="102"] img')).toHaveAttribute('src', /fixture_orange\.jpg$/);
await expect(tip.locator('.tooltip-enchant')).toHaveText('+6 All Stats');
await expect(tip.locator('.tooltip-socket-bonus')).not.toHaveClass(/inactive/);
```

- [ ] Add separate cases for an empty socket, mismatched-color equipped gem with inactive bonus, icon failure, re-import with a recovered icon, plain neck, weapon, unknown metadata, and name-only gem. Verify section order using element bounding boxes and confirm there is no `Equip: +6 All Stats` text.
- [ ] Add hover-to-panel persistence, icon-to-name switching, rapid movement across items, keyboard focus, Escape while focused, tab changes/unmount, and report-summary tests. Assert exactly one visible tooltip and that a stale timer never reopens an old item.
- [ ] Check complete footprint bounds for both columns, bottom weapons, page scroll, report-table horizontal scroll, and a long name:

```js
const box = await page.locator('.tooltip-layer').boundingBox();
const viewport = page.viewportSize();
expect(box.x).toBeGreaterThanOrEqual(8);
expect(box.y).toBeGreaterThanOrEqual(8);
expect(box.x + box.width).toBeLessThanOrEqual(viewport.width - 8);
expect(box.y + box.height).toBeLessThanOrEqual(viewport.height - 8);
```

Run at 1280x900 and 390x844. Verify browser zoom at 200% manually and confirm scrollable long content and Details access on a touch-sized viewport.

- [ ] Capture the neck and gemmed chest screenshots as review artifacts using `testInfo.attach` and `locator.screenshot`. Review against user references for actual image content, external icon, border, title wrap, and spacing. Do not create brittle cross-platform pixel baselines unless the project also standardizes the screenshot browser/OS/fonts.
- [ ] Add `"test:unit": "node --test src/lib/*.test.js"` to `ui-finder/package.json`. Add a CI step after UI dependency installation:

```yaml
      - name: Test UI helpers
        working-directory: ui-finder
        run: npm run test:unit
```

Keep CI's existing build-before-Playwright ordering. No root TypeScript script change is required for this Svelte feature.

- [ ] Run the final scoped checks:

```text
rtk go test ./cmd/wowsimcli/cmd/... -count=1
rtk npm --prefix ui-finder run test:unit
rtk npm --prefix ui-finder run build
rtk npm --prefix ui-finder run test:e2e
rtk git diff --check
rtk git status --short
```

Install with `rtk npm --prefix ui-finder ci` only if dependencies are missing; install Chromium with `rtk npm exec --prefix ui-finder -- playwright install chromium` if required. Use the platform's approval mechanism for restricted network/install operations. Report pre-existing failures separately; do not expand into simulator repairs.

- [ ] Update `docs/upgrade-finder.md` to describe actual gem icons, hover/focus/Escape behavior, Details access, and the metadata boundary. Include rebuilt `cmd/wowsimcli/cmd/upgrade_ui/` outputs with the source changes. Suggested commit: `test: verify redesigned tooltips and refresh embedded UI`.

## Completion checklist

- [ ] The external item icon and actual gem icons appear in the production Go-served UI.
- [ ] Base stats, enchants, gems, bonuses, and equip lines are ordered and correctly labeled.
- [ ] Empty sockets and failed image loads cannot be confused.
- [ ] Item names wrap cleanly; phase does not wrap or force an excessively narrow title.
- [ ] The full panel footprint stays in the viewport and text stays left-aligned in either column.
- [ ] Mouse and keyboard have equivalent content; Escape and re-import work reliably.
- [ ] Touch and keyboard users can inspect the same information through Details.
- [ ] Report summary tooltips remain accurate and functional.
- [ ] Targeted Go, Node, browser, and embedded-build checks passed; visual artifacts were inspected.
- [ ] No unsupported binding, durability, level, price, proc, meta-activation, or full set-bonus values were invented.

## Explicitly separate future work

Full reference metadata remains the follow-up described in the design: audited local sources for binding, required level, vendor value, durability, readable proc/use effects, meta-gem requirements, and complete set membership/bonus descriptions. That work needs its own source-coverage investigation and implementation plan. This implementation already reserves the appropriate visual grouping without displaying false placeholder facts.
