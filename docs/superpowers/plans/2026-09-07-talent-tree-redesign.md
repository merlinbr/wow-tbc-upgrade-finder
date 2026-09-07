# Talent Tree Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the current text/pip talent presentation with a reference-style, three-panel icon viewer while preserving the imported build and ranking behavior.

**Architecture:** Keep raw TBC trees and existing rank helpers. Generate a narrow static icon manifest, generalize the existing tooltip trigger/layer for item and talent content, then render real talent nodes directly into native CSS Grid positions. No backend changes or new dependencies.

**Tech Stack:** Existing Svelte 5, Vite 7, Node built-ins, node:test and Playwright.

## Status and prerequisites

**Draft for user review; not approved for implementation.** This plan follows the [proposed design](../specs/2026-09-07-talent-tree-redesign-design.md), with measured evidence in [research](../../talent-tree-redesign-research.md). Do not treat these documents as permission to start application edits or commits.

The current source is the starting point, not a clean branch assumption. Preserve unrelated user work. Before executing, read the then-current files and resolve references using the available language server before exported-symbol/file renames; the locations below are scope anchors, not immutable line numbers.

## Global Constraints

- One imported build; read-only inspection. No allocation, saved builds, dual-spec, comparison or baseline mutation.
- No new dependencies. No Go API, protobuf, ranking or import-schema changes.
- Preserve the original talent-array indices, four-column sparse locations and nine source tiers. No sorting/compaction of serialized data.
- Keep `decodeTalentsString`, `rankAt` and `treePoints` semantics unchanged; no implicit validation or clamping.
- Reference-style board: max-width 936 px; three columns at talent-container width ≥888 px with 24 px gaps; stacked below that width. Nodes 52 px wide in the wide layout, 44 px otherwise.
- Reuse TBC remote art and the existing icon error fallback; no new asset mirroring or claimed redistribution rights.
- Keep Gear/Stats/Talents navigation and character header unchanged. New Imported build / Read-only metadata is static, not a new tab or fake selector.
- Reuse one tooltip lifecycle and positioning implementation for item and talent details. No parallel tooltip framework or fake item-shaped talents.
- Talent detail content is name, tree, actual rank/max and allocation state. Full effect descriptions are unavailable and out of scope.
- New behavior needs browser proof; tests cover plausible regressions, not source strings or incidental class names. Do not add a screenshot test dependency.

---

## File ownership and execution order

| Slice | Files | Responsibility |
| --- | --- | --- |
| Icon data | New `ui-finder/scripts/generate-talent-icons.mjs`, new `ui-finder/src/lib/talentIcons.json`; modify `ui-finder/package.json` | Reproducible narrow icon projection and freshness check |
| Tooltip seam | Rename `src/lib/ItemTooltipTrigger.svelte` → `TooltipTrigger.svelte`, `ItemTooltipLayer.svelte` → `TooltipLayer.svelte`; modify `tooltipController.js`, `App.svelte`, `GearSlot.svelte`, `ReportView.svelte`; new `TalentTooltip.svelte` | Reuse lifecycle/positioning for two real content types and explicit activation |
| Talent presentation | Modify `src/lib/TalentTrees.svelte`, `src/app.css` | Imported summary, portrait cards, sparse icon placement, states and responsive layout |
| Verification/docs | Existing `src/lib/tooltipController.test.js`, `src/lib/talents.test.js`, `e2e/armory.spec.js`, `e2e/item-tooltip.spec.js`; new `e2e/talents.spec.js` only for the concrete scenario below; update `docs/upgrade-finder.md`, `docs/STATE.md` after implementation proof | Preserve contracts and record shipped behavior |

Paths abbreviated with `src/` or `e2e/` in this table are relative to `ui-finder/`.

Task 1 and Task 2 can be implemented independently, with no concurrent validation; Task 3 consumes both. Main/integration owner runs validation after shared edits settle. Do not delegate tiny mechanical caller changes separately. The final build owns generated `cmd/wowsimcli/cmd/upgrade_ui/` output; never hand-edit bundled assets.

## Task 1: Ship a reproducible talent-icon projection

**Files:** create `ui-finder/scripts/generate-talent-icons.mjs` and `ui-finder/src/lib/talentIcons.json`; modify `ui-finder/package.json`.

**Interface:** generated JSON maps the numeric first-rank spell ID, represented as a JSON object key, to a ZAM icon basename without extension. `node scripts/generate-talent-icons.mjs` writes it; `--check` compares without writing and fails on stale/missing data. Both resolve paths relative to the script, not the caller's cwd.

- [ ] Add the generator using Node built-ins. The complete projection algorithm is:

```js
import { readFile, writeFile } from 'node:fs/promises';

const root = new URL('../../', import.meta.url);
const output = new URL('../src/lib/talentIcons.json', import.meta.url);
const classes = ['druid', 'hunter', 'mage', 'paladin', 'priest', 'rogue', 'shaman', 'warlock', 'warrior'];
const readJson = async (path) => JSON.parse(await readFile(new URL(path, root), 'utf8'));
const [db, ...classTrees] = await Promise.all([
  readJson('assets/database/db.json'),
  ...classes.map((name) => readJson(`ui/core/talents/trees/${name}.json`)),
]);
const source = new Map(db.spellIcons.map(({ id, icon }) => [id, icon]));
const icons = {};
for (const tree of classTrees.flat()) {
  for (const talent of tree.talents) {
    const id = talent.spellIds[0];
    const icon = source.get(id);
    if (!icon) throw new Error(`Missing talent icon: ${talent.fancyName} (spell ${id})`);
    icons[id] = icon;
  }
}
const text = `${JSON.stringify(icons, null, 2)}\n`;
if (process.argv.includes('--check')) {
  const current = await readFile(output, 'utf8').catch((error) => {
    if (error.code === 'ENOENT') return '';
    throw error;
  });
  if (current !== text) throw new Error('Stale talentIcons.json; run node scripts/generate-talent-icons.mjs');
} else {
  await writeFile(output, text);
}
```

- [ ] Run the generator from `ui-finder/`, then `node scripts/generate-talent-icons.mjs --check`. This proves current-source coverage and determinism; no network is involved. The measured dataset has 579 entries, but do not hardcode that count as a future compatibility limit.
- [ ] Add `"prebuild": "node scripts/generate-talent-icons.mjs --check"` and prefix the existing `test:unit` command with `node scripts/generate-talent-icons.mjs --check && `. Preserve the existing test command after the prefix. No runtime database import and no extra package/dependency.
- [ ] Verify the manifest's Divine Strength entry is `"20262": "ability_golemthunderclap"`; test one temporary stale-output edit against `--check`, restore by regeneration, then check again. Do not retain the temporary corruption. This is an experiment, not a source-text test to add permanently.

**Acceptance:** generation fails on absent source icons; `--check` fails on stale output; the browser receives only the projection plus existing tree data, never a full `db.json` import. The command itself remains the runnable data-coverage/freshness guard.

## Task 2: Reuse the existing tooltip lifecycle for talent content

**Files:** the tooltip-seam row above, plus existing tooltip tests if the changed observable behavior requires it. Do not modify `tooltipPosition.js` unless a browser reproduction exposes an actual placement bug.

**Consumes:** existing controller methods `enter(record, source)`, `leave(owner, source)`, `enterPanel`, `leavePanel`, `dismiss`, `release`, `destroy`.

**Produces:**

```js
// TooltipTrigger content prop — exactly two supported content shapes:
{ kind: 'item', item, variant: 'full' } // or 'summary'
{ kind: 'talent', talent: { name, treeName, points, maxPoints } }
// Active controller record:
{ owner, id, content, preferredSide, anchor }
// children snippet receives existing handlers plus onactivate:
{ id, onpointerenter, onpointerleave, onfocus, onblur, onactivate }
```

- [ ] Resolve symbol/file references using LSP when available. Rename the two existing Svelte files and `ITEM_TOOLTIP_CONTEXT` → `TOOLTIP_CONTEXT`; migrate every caller in the same cutover. There are currently callers in `App.svelte`, `GearSlot.svelte` and `ReportView.svelte`. No compatibility files or aliases.
- [ ] Change the trigger props from `item`/`variant` to `content`, retaining `preferredSide` and `children`. `buildRecord(target)` returns `{ owner, id, content, preferredSide, anchor: target }`. The existing cleanup effect observes `content` instead of `item`; keep release on unmount and re-import.
- [ ] Keep existing pointer/focus handlers and add explicit activation. Bind this handler to `onclick` on the real trigger buttons so keyboard Enter/Space uses native button activation too:

```js
function onactivate(event) {
  const target = event.currentTarget;
  target.focus();
  // Explicit activation reopens a popup dismissed while focus stayed here.
  controller.release(owner);
  controller.enter(buildRecord(target), 'focus');
}
```

Do not synthesize key events or add a separate tap state machine. The focus source retains touch inspection during scroll. Blur and unmount already release that ownership through existing lifecycle behavior.

- [ ] Change the shared layer's content branch only; retain its measurement, scroll/resize observation, Escape handling and hover bridge:

```svelte
{#key active.owner}
  {#if active.content.kind === 'talent'}
    <TalentTooltip talent={active.content.talent} id={active.id} />
  {:else}
    <ItemTooltip item={active.content.item} variant={active.content.variant} id={active.id} />
  {/if}
{/key}
```

- [ ] Add `TalentTooltip.svelte` as a content renderer, not an owner of timers/positioning:

```svelte
<script>
  let { talent, id } = $props();
  let allocation = $derived(talent.points === 0 ? 'Unallocated'
    : talent.points === talent.maxPoints ? 'Maximum rank' : 'Allocated');
</script>

<div class="talent-tooltip" role="tooltip" {id}>
  <strong>{talent.name}</strong>
  <div>{talent.treeName}</div>
  <div>Rank {talent.points} / {talent.maxPoints}</div>
  <div>{allocation} · Imported build · Read-only</div>
</div>

<style>
  .talent-tooltip {
    background: #080d16;
    border: 1px solid #85848b;
    border-radius: 4px;
    color: #e8edf5;
    font-size: 0.85rem;
    max-width: min(20rem, calc(100vw - 1rem));
    overflow-wrap: anywhere;
    padding: 0.65rem 0.8rem;
  }
  strong { color: #c7ef31; }
</style>
```

- [ ] Migrate item consumers without changing their formatters:

```svelte
<!-- GearSlot: retain the existing icon/name children and events -->
<TooltipTrigger content={{ kind: 'item', item: slot, variant: 'full' }} {preferredSide}>
<!-- ReportView: retain the existing report button and events -->
<TooltipTrigger content={{ kind: 'item', item: upgrade.item, variant: 'summary' }} preferredSide="right">
```

The snippets above are the opening tags being replaced, not new nested wrappers. Update closing tags, imports and children destructuring; wire `onclick={onactivate}` on both gear triggers and the report trigger. `App.svelte` still creates exactly one controller/context and renders exactly one layer.

- [ ] Update controller documentation to the new record shape. Leave the state-machine methods unchanged. Run existing controller/position unit tests and item-tooltip browser checks after concurrent edits settle. Keep or extend a regression only if it proves a real behavior: Escape → explicit activation reopens; unmount/re-import prevents stale delayed reopening.

**Acceptance:** existing item details remain identical; hover/focus/Escape/edge positioning still work; explicit click/tap works without allocating anything. The new talent renderer uses the same overlay, not a second timer/position controller.

## Task 3: Replace text/pip panels with the icon-tree composition

**Files:** `ui-finder/src/lib/TalentTrees.svelte`, `ui-finder/src/app.css`.

**Consumes:** existing `{ class, talentsString }` props and rank helpers; `talentIcons.json`; new `TooltipTrigger`; existing `TooltipIcon.svelte`.

**Produces:** unchanged `data-region="talent-trees"` surface, real sparse inspection nodes, static build summary and responsive portrait panels. `ArmoryView.svelte` needs no new prop.

- [ ] Keep the existing nine imports/class map. Import the icon manifest and tooltip modules. Replace dense `cells` construction with data that preserves the original `talents` array and the rank string:

```js
let treeData = $derived((treesByClass[playerClass] ?? []).map((tree, treeIndex) => {
  const ranks = ranksByTree[treeIndex] ?? '';
  return {
    ...tree,
    ranks,
    points: treePoints(ranks),
    rows: Math.max(...tree.talents.map(t => t.location.rowIdx), 0) + 1,
    cols: Math.max(...tree.talents.map(t => t.location.colIdx), 0) + 1,
    headerIcon: talentIcons[tree.talents.at(-1)?.spellIds[0]] ?? '',
  };
}));
```

The `at(-1)` artwork is the measured capstone in every current tree. It is representative art, not imported specialization metadata. Avoid any `PlayerClasses.specs[treeIndex]` lookup.

- [ ] Add a `.talent-view` wrapper establishing the sizing container. Add the summary `Imported build`, a static `Read-only` badge, and a labeled split. For example:

```svelte
<div class="talent-build-summary">
  <strong>Imported build</strong>
  <span class="talent-readonly">Read-only</span>
  <span aria-label={treeData.map(t => `${t.name}: ${t.points} points`).join(', ')}>
    {treeData.map(t => t.points).join(' / ')}
  </span>
</div>
```

Keep this inside the supported-class branch. Preserve the current unsupported-class message; absent rank digits still decode to zero. Do not label a dominant tree or manufacture another build button.

- [ ] Keep semantic sections, headers and totals. Put a 32 px decorative `TooltipIcon` beside each header name. The body renders actual nodes directly; remove null placeholders, text names and pips:

```svelte
{#each tree.talents as talent, index (talent.spellIds[0])}
  {@const points = rankAt(tree.ranks, index)}
  <TooltipTrigger
    content={{ kind: 'talent', talent: { name: talent.fancyName, treeName: tree.name, points, maxPoints: talent.maxPoints } }}
  >
    {#snippet children({ id, onpointerenter, onpointerleave, onfocus, onblur, onactivate })}
      <button
        type="button"
        class="talent-node"
        class:unallocated={points === 0}
        class:allocated={points > 0}
        class:maxed={points === talent.maxPoints}
        style:grid-row={talent.location.rowIdx + 1}
        style:grid-column={talent.location.colIdx + 1}
        aria-label={`Inspect ${talent.fancyName}, rank ${points} of ${talent.maxPoints}`}
        aria-describedby={id}
        {onpointerenter} {onpointerleave} {onfocus} {onblur}
        onclick={onactivate}
      >
        <span class="talent-node-art" aria-hidden="true">
          <TooltipIcon icon={talentIcons[talent.spellIds[0]] ?? ''} name={talent.fancyName} size={40} />
        </span>
        <span class="talent-rank" aria-hidden="true">{points}/{talent.maxPoints}</span>
      </button>
    {/snippet}
  </TooltipTrigger>
{/each}
```

Keep artwork centered at 40 px inside the minimum 44 px target. In wide mode enlarge only its rendered box using the local `.talent-node-art` CSS override to 48 px, not every tooltip icon in the app. The CDN large image is 56 px for the sampled icon. Do not alter `TooltipIcon`'s interface solely to add a responsive-size framework.

- [ ] Replace the old `.talent-*`/`.rank-pip*` section in `app.css`, rather than adding competing selectors elsewhere. Core layout rules:

```css
.talent-view { container: talent-view / inline-size; }
.talent-build-summary { align-items: center; display: flex; flex-wrap: wrap; gap: .65rem; justify-content: center; margin: 1rem 0; }
.talent-trees { display: grid; gap: 24px; grid-template-columns: minmax(0, 1fr); margin-inline: auto; max-width: 360px; }
.talent-tree { --node-size: 44px; --node-gap: 10px; background: #0b1018; border: 1px solid var(--border); border-radius: 7px; min-width: 0; }
.talent-tree-header { align-items: center; background: rgba(5, 8, 12, .85); display: flex; flex-wrap: wrap; gap: .5rem; min-height: 56px; padding: .65rem .8rem; }
.talent-tree-header h3 { flex: 1; margin: 0; }
.talent-tree-points { color: var(--gold); font-variant-numeric: tabular-nums; }
.talent-tree-body { background-position: center; background-repeat: no-repeat; background-size: 100% 100%; border-radius: 0 0 6px 6px; display: grid; gap: var(--node-gap); grid-template-columns: repeat(var(--talent-cols), var(--node-size)); grid-template-rows: repeat(var(--talent-rows), var(--node-size)); justify-content: center; padding: 16px 12px; }
.talent-node { align-items: center; background: #0b1018; border: 2px solid #89939f; display: flex; height: var(--node-size); justify-content: center; min-width: 0; padding: 0; position: relative; width: var(--node-size); }
.talent-node:hover:not(:disabled) { background: #0b1018; }
.talent-node.allocated { border-color: #c7ef31; }
.talent-node.maxed { box-shadow: inset 0 0 0 1px #c7ef31; }
.talent-node.unallocated .talent-node-art { filter: grayscale(1) brightness(.48); }
.talent-node:focus-visible { outline: 3px solid var(--gold); outline-offset: 3px; }
.talent-node-art { display: flex; }
.talent-rank { background: #080d16; bottom: 0; color: #e8edf5; font-size: 12px; font-variant-numeric: tabular-nums; line-height: 1.2; padding: 1px 3px; position: absolute; right: 0; }
.talent-node.allocated .talent-rank { color: #c7ef31; }
.talent-readonly { background: #182419; border: 1px solid #758c37; border-radius: 3px; color: #c7ef31; font-size: .75rem; padding: .1rem .4rem; }
@container talent-view (min-width: 888px) {
  .talent-trees { grid-template-columns: repeat(3, minmax(0, 1fr)); max-width: 936px; }
  .talent-tree { --node-size: 52px; --node-gap: 12px; }
  .talent-node-art > * { height: 48px !important; width: 48px !important; }
}
```

The narrow `!important` overrides only the reused icon module's existing inline size; do not widen it to global icons. If the repo's style rules disallow this, pass a CSS custom-property size through the existing `size` prop instead and use LSP references before changing that module's interface; do not duplicate its image-failure behavior.

Set `--talent-cols`/`--talent-rows` from `tree.cols`/`tree.rows` on the section. The grid's inline background uses the existing URL beneath a dark overlay:

```svelte
style:background-image={tree.backgroundUrl
  ? `linear-gradient(rgba(5, 8, 12, .18), rgba(5, 8, 12, .40)), url('${tree.backgroundUrl}')`
  : 'linear-gradient(#182130, #080d16)'}
```

Add the inset vignette as a **decorative pseudo-element** on the positioned body with `pointer-events: none`; put nodes above it with `position: relative; z-index: 1`. Do not apply the vignette or reduced opacity to badge text. Keep focus outlines unclipped; no blanket `overflow: hidden` on the tree.

```css
.talent-tree-body { position: relative; }
.talent-tree-body::before {
  border-radius: inherit;
  box-shadow: inset 0 0 42px 12px rgba(0, 0, 0, .65);
  content: '';
  inset: 0;
  pointer-events: none;
  position: absolute;
}
.talent-node { z-index: 1; }
```

- [ ] Visually tune overlay/frame contrast against real art, not a blank fixture. The provided color/spacing rules are a concrete starting implementation; acceptance includes measured contrast and absence of overlap, not an exact CSS-text assertion.

**Acceptance:** the real Paladin fixture has three portrait panels, 64 visible inspection nodes, original sparse positions, and 5 / 11 / 45 totals; no name/pip tiles remain. All nine classes use the same data-driven renderer. Narrow layout preserves readable targets and full tree topology.

## Task 4: Prove behavior and the actual visual surface

**Files:** focused `ui-finder/e2e/talents.spec.js` if needed; existing tooltip/rank tests; docs after the browser proof.

- [ ] Keep one behavior-focused regression scenario using the real import fixture, matching existing Playwright conventions. This is the core observable contract, not a test of private field copies:

```js
import { expect, test } from '@playwright/test';
import { readFileSync } from 'node:fs';
const link = readFileSync(new URL('../../cmd/wowsimcli/cmd/upgrades/testdata/retribution_no_settings_link.txt', import.meta.url), 'utf8').trim();

test('talent inspection preserves the imported build and dismisses cleanly', async ({ page }) => {
  await page.goto('/');
  await page.getByLabel('wowsims export link').fill(link);
  await page.getByRole('button', { name: 'Import settings', exact: true }).click();
  await page.getByRole('tab', { name: 'Talents', exact: true }).click();
  const trees = page.locator('[data-region="talent-trees"]');
  const before = await trees.getByRole('button').evaluateAll(nodes => nodes.map(node => node.getAttribute('aria-label')));
  const node = page.getByRole('button', { name: 'Inspect Divine Strength, rank 5 of 5', exact: true });
  await node.focus();
  await expect(page.getByRole('tooltip')).toContainText('Divine Strength');
  await page.keyboard.press('Escape');
  await expect(page.getByRole('tooltip')).toHaveCount(0);
  await page.keyboard.press('Enter');
  await expect(page.getByRole('tooltip')).toContainText('Rank 5 / 5');
  expect(await trees.getByRole('button').evaluateAll(nodes => nodes.map(node => node.getAttribute('aria-label')))).toEqual(before);
  await page.getByRole('tab', { name: 'Gear', exact: true }).click();
  await expect(page.getByRole('tooltip')).toHaveCount(0);
});
```

This defends immutable displayed allocations and the Escape/reactivation/unmount transition. Do not create dozens of per-talent tests. Reuse existing controller timing tests for stale timer behavior.

- [ ] Run focused commands from `ui-finder/` after all implementation edits settle:

```sh
node scripts/generate-talent-icons.mjs --check
node --test src/lib/talents.test.js src/lib/tooltipController.test.js src/lib/tooltipPosition.test.js
npm run build
npx playwright test e2e/talents.spec.js e2e/item-tooltip.spec.js e2e/armory.spec.js
```

The current `playwright.config.js` starts Go servers on ports 43123 and 43124 and exercises the **embedded bundle**, not the Vite source server. Build the UI **before** those servers compile/start; otherwise the test sees stale UI. Stop conflicting fixed-port processes first (`reuseExistingServer` is false). Commands above are planned checks, **not results from this research session**.

- [ ] Launch the source UI with the local API, import the Paladin fixture and inspect the actual Talents tab at 1280, 978, 390 and 320 px widths. Capture after screenshots at desktop and mobile. Compare with the research baseline and supplied image: three approximately 650 px portrait cards on wide screens, legible art, clear ranks, consistent node positions and no hidden horizontal scroll.
- [ ] Inspect a zero-rank, partial-rank and full-rank node. Exercise pointer-to-popup travel, Tab, Enter, Space, Escape then reactivation, touch tap, scrolling and viewport edges. Verify 400%-equivalent reflow and measure text/focus/state contrast. Do not equate node grayscale with disabling inspection.
- [ ] Block CDN images in a dedicated browser check. Verify name/rank inspection and fallback remain usable; unblocking/re-importing can show fresh icons. Do not claim all 579 remote URLs are available based on the one sampled research icon.
- [ ] Re-import the existing Mage fixture while a Paladin popup is open. Verify the new three-tree names/ranks and absence of stale Paladin content. Exercise Gear and report tooltips after the shared refactor, and the existing ranking flow through its current E2E scenario.
- [ ] Smoke-check all nine class datasets using a temporary isolated browser harness that supplies their actual tree data and representative rank strings. This is a rendering experiment, not a claim that unsupported simulator specs can be imported. Check original first/middle/last positions and all three headers; remove the temporary harness after the check.
- [ ] After browser proof, update `docs/upgrade-finder.md` and `docs/STATE.md` with the shipped read-only icon layout, detail limitations, responsive behavior and icon-regeneration command. Keep research baseline screenshots labeled **before**; do not overwrite them with after images.
- [ ] Keep the final Vite-generated embedded bundle consistent with the source. For packaged-surface verification, rebuild `wowsimcli` after that build and launch the rebuilt binary; restarting an older binary does not include new embedded assets. No hand-edited generated JavaScript.

**Final acceptance:** the design's seven acceptance groups are covered with real browser evidence, targeted regression results and a production build. Report precisely what ran and any unverified remote asset/rights limits. No application implementation has been performed by this planning session.

## Design-to-plan coverage

| Design requirement | Plan coverage |
| --- | --- |
| Existing imported build and original indices | Tasks 3 and 4; unchanged helpers and native inspection only |
| Illustrated portrait cards and icon/rank states | Task 3; browser sizes/art contrast in Task 4 |
| Sparse 4 × 9 topology and narrow stacking | Task 3 explicit CSS placement and container query; Task 4 all-class/mobile smoke |
| Complete narrow icon projection | Task 1 generation and freshness guard |
| Shared accessible mouse/keyboard/touch details | Task 2 shared lifecycle/activation; Task 4 transition checks |
| No fabricated descriptions/second builds | Global constraints, Task 2 exact content, Task 3 static summary |
| Failure/re-import/old item-tooltip preservation | Tasks 2 and 4 |
| Documentation and embedded production surface | Task 4 after browser proof |

## Handoff

Review the proposed design first. If approved, implement Tasks 1–3 with the named file ownership, then perform Task 4's verification and documentation. A faithful read-only viewer is the complete scope; a talent calculator is a separate product decision, not an implicit next phase.
