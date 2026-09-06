# Character Identity Header Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restyle the imported-character identity area as a reference-style identity card (class-icon avatar, class-colored name, meta line, real-data chip row, muted sim-import note) with a pill tab bar, without changing any import/ranking data flow.

**Architecture:** Pure frontend. A new `identity.js` module maps class enums to hex colors and ZAM icon URLs and computes average item level from the existing `gear` array; `CharacterHeader.svelte` renders the card; `ArmoryView.svelte` passes `gear` down; `app.css` carries the card/chip/pill-tab styles. Backend and the browser-renders-server-data contract are untouched.

**Tech Stack:** Svelte 5 (`$props`/`$derived`/`$state`), Vite 7, `node:test` unit tests, Playwright E2E.

## Global Constraints

- No new dependencies. No Go changes. No new API fields.
- Class colors and icon names MUST mirror the simulator UI values (`ui/core/player_classes/*.ts`, `ui/core/player_specs/*.ts`).
- Tab markup keeps `role="tab"`, `aria-controls`, `aria-selected`, ids `tab-gear`/`tab-stats`/`tab-talents`, and labels Gear/Stats/Talents verbatim.
- The heading keeps `id="armory-heading"` and the exact name text (E2E depends on it).
- Note copy is exactly `No ratings — local simulation import` (em dash).
- All work happens on branch `feat/character-identity-header`. Commit docs and code to this branch only; do not touch the untracked root files `docs/blizzard-character-api-rights-research.md` and `docs/independent-character-renderer-research.md`.

---

### Task 1: identity.js display-data module and unit tests

**Files:**
- Create: `ui-finder/src/lib/identity.js`
- Test: `ui-finder/src/lib/identity.test.js`

**Interfaces:**
- Produces: `classColor(value) → string` (hex or `''`), `classIcon(value) → string` (absolute ZAM URL or `''`), `avgItemLevel(gear) → number` (0 when no equipped items; averages slots with non-zero `itemId` and `ilvl > 0`, rounded).

- [ ] **Step 1: Write the failing test**

```js
// ui-finder/src/lib/identity.test.js
import test from 'node:test';
import assert from 'node:assert/strict';
import { classColor, classIcon, avgItemLevel } from './identity.js';

test('classColor covers all nine classes', () => {
  const classes = [
    'ClassDruid', 'ClassHunter', 'ClassMage', 'ClassPaladin', 'ClassPriest',
    'ClassRogue', 'ClassShaman', 'ClassWarlock', 'ClassWarrior',
  ];
  for (const klass of classes) assert.match(classColor(klass), /^#[0-9a-f]{6}$/i);
  assert.equal(classColor('UnknownClass'), '');
});

test('classIcon builds a ZAM CDN medium icon URL', () => {
  assert.equal(classIcon('ClassMage'), 'https://wow.zamimg.com/images/wow/icons/medium/class_mage.jpg');
  assert.equal(classIcon(''), '');
});

test('avgItemLevel skips empty slots and rounds', () => {
  const gear = [
    { itemId: 1, ilvl: 100 },
    { itemId: 0, ilvl: 0 },
    { itemId: 2, ilvl: 101 },
  ];
  assert.equal(avgItemLevel(gear), 101);
  assert.equal(avgItemLevel([]), 0);
  assert.equal(avgItemLevel([{ itemId: 0, ilvl: 0 }]), 0);
});
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `cd ui-finder && npm run test:unit`
Expected: FAIL — `identity.js` module not found.

- [ ] **Step 3: Write the minimal implementation**

```js
// ui-finder/src/lib/identity.js
// Class presentation values mirror the simulator UI:
// ui/core/player_classes/*.ts (hexColor) and ui/core/player_specs/*.ts
// (getIcon('medium') icon names).
const classColors = {
  ClassDruid: '#ff7d0a',
  ClassHunter: '#abd473',
  ClassMage: '#69ccf0',
  ClassPaladin: '#f58cba',
  ClassPriest: '#ffffff',
  ClassRogue: '#fff569',
  ClassShaman: '#2459ff',
  ClassWarlock: '#9482c9',
  ClassWarrior: '#c79c6e',
};

const classIcons = {
  ClassDruid: 'class_druid.jpg',
  ClassHunter: 'class_hunter.jpg',
  ClassMage: 'class_mage.jpg',
  ClassPaladin: 'class_paladin.jpg',
  // The simulator UI uses this priest icon as the class icon.
  ClassPriest: 'spell_shadow_shadowwordpain.jpg',
  ClassRogue: 'class_rogue.jpg',
  ClassShaman: 'class_shaman.jpg',
  ClassWarlock: 'class_warlock.jpg',
  ClassWarrior: 'class_warrior.jpg',
};

export function classColor(value) {
  return classColors[value] ?? '';
}

export function classIcon(value) {
  const name = classIcons[value];
  return name ? `https://wow.zamimg.com/images/wow/icons/medium/${name}` : '';
}

export function avgItemLevel(gear = []) {
  const levels = gear
    .filter((slot) => (slot.itemId ?? 0) !== 0 && (slot.ilvl ?? 0) > 0)
    .map((slot) => slot.ilvl);
  if (levels.length === 0) return 0;
  return Math.round(levels.reduce((sum, value) => sum + value, 0) / levels.length);
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `cd ui-finder && npm run test:unit`
Expected: PASS (4 tests: 3 new + existing labels/api/talents suites still pass).

- [ ] **Step 5: Commit**

```bash
git add ui-finder/src/lib/identity.js ui-finder/src/lib/identity.test.js
git commit -m "feat(ui): add character identity display-data helpers"
```

---

### Task 2: CharacterHeader identity card and gear prop

**Files:**
- Modify: `ui-finder/src/lib/CharacterHeader.svelte` (full replace)
- Modify: `ui-finder/src/lib/ArmoryView.svelte` (pass `gear`)

**Interfaces:**
- Consumes: `classColor`, `classIcon`, `avgItemLevel` from `identity.js` (Task 1); `humanizeEnum` from `./labels.js`.
- Produces: `CharacterHeader` props `{ character = {}, phase = 0, gear = [], settingsDigest = '', simulatorRevision = '', databaseRevision = '' }`.

- [ ] **Step 1: Replace CharacterHeader.svelte**

```svelte
<script>
  import { avgItemLevel, classColor, classIcon } from './identity.js';
  import { humanizeEnum } from './labels.js';

  let { character = {}, phase = 0, gear = [], settingsDigest = '', simulatorRevision = '', databaseRevision = '' } = $props();

  const professionNames = {
    1: 'Alchemy',
    2: 'Blacksmithing',
    3: 'Enchanting',
    4: 'Engineering',
    5: 'Herbalism',
    6: 'Inscription',
    7: 'Jewelcrafting',
    8: 'Leatherworking',
    9: 'Mining',
    10: 'Skinning',
    11: 'Tailoring',
  };

  let professions = $derived((character.professions ?? []).map((profession) => professionNames[profession] ?? String(profession)));
  let color = $derived(classColor(character.class));
  let avatarUrl = $derived(classIcon(character.class));
  let avatarFailed = $state(false);
  let itemLevel = $derived(avgItemLevel(gear));

  function digest(value) {
    if (!value) return '—';
    return `${value.slice(0, 16)}…`;
  }
</script>

<div class="character-header">
  <div class="character-identity">
    <div class="identity-avatar" aria-hidden="true">
      {#if avatarUrl && !avatarFailed}
        <img src={avatarUrl} alt="" onerror={() => (avatarFailed = true)} />
      {:else}
        <span class="avatar-fallback" style:background={color || '#536b8a'}>{(character.name || '?').trim()[0]?.toUpperCase() || '?'}</span>
      {/if}
    </div>
    <div class="identity-copy">
      <div class="section-kicker">Imported character</div>
      <h2 id="armory-heading" class="character-name" style:color={color || 'var(--text)'}>{character.name || 'Unnamed character'}</h2>
      <p class="character-subtitle">Level 70 {humanizeEnum(character.race, 'Race') || 'Unknown race'} · {character.spec ? humanizeEnum(character.spec) : humanizeEnum(character.class, 'Class') || 'Unknown class'}</p>
      <div class="identity-chips" aria-label="Character facts">
        <span class="chip"><span class="chip-label">Avg ilvl</span><strong>{itemLevel > 0 ? itemLevel : '—'}</strong></span>
        <span class="chip"><span class="chip-label">Phase</span><strong>{phase || '—'}</strong></span>
        <span class="chip"><span class="chip-label">Professions</span><strong>{professions.length ? professions.join(', ') : 'None'}</strong></span>
      </div>
      <p class="identity-note">No ratings — local simulation import</p>
    </div>
  </div>
  <div class="character-actions">
    <a class="find-upgrades" href="#ranking-heading">Find upgrades</a>
    <details class="import-details">
      <summary>Import details</summary>
      <dl>
        <div><dt>Settings digest</dt><dd title={settingsDigest}>{digest(settingsDigest)}</dd></div>
        <div><dt>Simulator revision</dt><dd>{simulatorRevision || '—'}</dd></div>
        <div><dt>Database revision</dt><dd>{databaseRevision || '—'}</dd></div>
      </dl>
    </details>
  </div>
</div>
```

- [ ] **Step 2: Pass gear from ArmoryView**

In `ui-finder/src/lib/ArmoryView.svelte`, change the `CharacterHeader` element:

```svelte
<CharacterHeader {character} {phase} {gear} settingsDigest={imported?.settingsDigest} simulatorRevision={imported?.simulatorRevision} databaseRevision={imported?.databaseRevision} />
```

- [ ] **Step 3: Run the unit tests**

Run: `cd ui-finder && npm run test:unit`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add ui-finder/src/lib/CharacterHeader.svelte ui-finder/src/lib/ArmoryView.svelte
git commit -m "feat(ui): render character identity card with avatar, chips, and sim note"
```

---

### Task 3: Card, chip-row, and pill tab styles

**Files:**
- Modify: `ui-finder/src/app.css`
  - `.character-header`, `.character-identity` block: replace with card layout + identity-layout styles.
  - `.character-facts` block: removed (facts move into chips) — delete its rules and the `.character-facts` references in media queries at existing lines ~192-195.
  - `.armory-tabs` / `.armory-tab` block (existing lines ~226-228): replace underline style with pill bar.

**Interfaces:**
- Consumes: markup class names from Task 2 (`.identity-avatar`, `.avatar-fallback`, `.identity-copy`, `.character-name`, `.identity-chips`, `.chip`, `.chip-label`, `.identity-note`).
- Produces: styles only; no JS surface.

- [ ] **Step 1: Replace the header layout block**

Replace (existing ~lines 93-96):

```css
.character-header, .report-header { align-items: start; display: flex; gap: 1.5rem; justify-content: space-between; }
.character-subtitle { color: var(--accent-strong); font-size: 1.05rem; }
.character-facts, .report-summary { display: grid; gap: 0.7rem 1.5rem; margin: 0; }
.character-facts { grid-template-columns: repeat(3, minmax(0, 1fr)); min-width: min(100%, 40rem); }
```

with:

```css
.character-header, .report-header { align-items: start; display: flex; gap: 1.5rem; justify-content: space-between; }
.character-header {
  background: rgba(6, 10, 16, 0.65);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 1rem;
}
.character-subtitle { color: var(--accent-strong); font-size: 1.05rem; }
.report-summary { display: grid; gap: 0.7rem 1.5rem; margin: 0; }
.character-identity { display: flex; gap: 1rem; min-width: 0; }
.identity-avatar {
  background: #10182a;
  border: 1px solid var(--border-bright);
  border-radius: 8px;
  display: flex;
  flex: 0 0 4rem;
  height: 4rem;
  overflow: hidden;
  width: 4rem;
}
.identity-avatar img { height: 100%; object-fit: cover; width: 100%; }
.avatar-fallback {
  align-items: center;
  color: #0a1420;
  display: flex;
  font-size: 1.8rem;
  font-weight: 800;
  height: 100%;
  justify-content: center;
  width: 100%;
}
.identity-copy { min-width: 0; }
.character-name { margin-bottom: 0.25rem; }
.identity-chips { display: flex; flex-wrap: wrap; gap: 0.5rem; margin: 0.6rem 0 0; }
.chip {
  align-items: baseline;
  background: #0b1018;
  border: 1px solid var(--border);
  border-radius: 6px;
  display: inline-flex;
  gap: 0.45rem;
  padding: 0.3rem 0.6rem;
}
.chip-label {
  color: var(--muted);
  font-size: 0.68rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}
.identity-note { color: var(--muted); font-size: 0.85rem; margin: 0.6rem 0 0; }
```

- [ ] **Step 2: Remove the leftover `.character-facts` rules**

Delete (existing ~lines 95-99, plus the media-query line `  .character-facts { min-width: 100%; }` at ~line 194 and the `.character-identity { min-width: 0; }` line at ~218, which is now superseded). Keep `.character-facts dt`/`dd`-adjacent selectors only if still present in the file after the first replacement — the `dt`/`dd` shared rule at ~96-98 must stay because `.report-summary` and `.stat-list` use it.

- [ ] **Step 3: Replace the tab styles with a pill bar**

Replace (existing ~lines 226-228):

```css
.armory-tabs { border-bottom: 1px solid var(--border); display: flex; gap: 0.5rem; margin-top: 1.25rem; }
.armory-tab { background: transparent; border: none; border-bottom: 2px solid transparent; border-radius: 0; color: var(--muted); padding: 0.55rem 0.9rem; }
.armory-tab.active { border-bottom-color: var(--accent); color: var(--text); }
```

with:

```css
.armory-tabs {
  background: rgba(6, 10, 16, 0.65);
  border: 1px solid var(--border);
  border-radius: 10px;
  display: flex;
  gap: 0.35rem;
  margin-top: 1.25rem;
  padding: 0.35rem;
}
.armory-tab {
  background: transparent;
  border: none;
  border-radius: 8px;
  color: var(--muted);
  padding: 0.45rem 1rem;
}
.armory-tab:hover:not(.active) { background: var(--panel-raised); color: var(--text); }
.armory-tab.active { background: var(--accent); color: #0a1420; }
```

- [ ] **Step 4: Verify styles compile**

Run: `cd ui-finder && npm run build`
Expected: Vite build passes.

- [ ] **Step 5: Commit**

```bash
git add ui-finder/src/app.css
git commit -m "style(ui): identity card, chip row, and pill tab bar"
```

---

### Task 4: E2E coverage and full verification

**Files:**
- Modify: `ui-finder/e2e/armory.spec.js` (extend assertions after the `TestMage` heading check)

**Interfaces:**
- Consumes: rendered markup from Tasks 2-3.
- Produces: E2E proof for avatar, class-colored name, chip row, muted note, and pill tab state.

- [ ] **Step 1: Extend the E2E assertions**

After the existing `expect(page.getByRole('heading', { name: 'TestMage', exact: true })).toBeVisible();` line, add:

```js
await expect(page.locator('.identity-avatar img')).toHaveAttribute('src', expect.stringContaining('class_mage.jpg'));
await expect(page.getByRole('heading', { name: 'TestMage', exact: true })).toHaveCSS('color', 'rgb(105, 204, 240)');
await expect(page.locator('.identity-chips .chip').first()).toContainText(/Avg ilvl \d+/);
await expect(page.getByText('No ratings — local simulation import', { exact: true })).toBeVisible();
await expect(page.locator('.armory-tab.active')).toHaveText('Gear');
```

- [ ] **Step 2: Run the E2E suite**

Run: `cd ui-finder && npm run test:e2e`
Expected: all Playwright projects pass (armory, character stage default, item tooltip; visuals project unchanged).

- [ ] **Step 3: Run unit tests and build once more**

Run: `cd ui-finder && npm run test:unit && npm run build`
Expected: PASS / Vite build passes.

- [ ] **Step 4: Commit and update STATE.md**

```bash
git add ui-finder/e2e/armory.spec.js
git commit -m "test(ui): assert character identity card and pill tabs"
```

Then add a short "Character identity header" paragraph to `docs/STATE.md` under **Working behavior → Frontend** and:

```bash
git add docs/STATE.md
git commit -m "docs: record character identity header behavior"
```

---

## Self-Review

- **Spec coverage:** avatar/class color (T1+T2), meta line + chips + note (T2), pill tabs (T3), fallback letter tile (T2 combined with `.avatar-fallback` in T3), tests (T1+T4), deferred items — none implemented, matching "Non-goals". No Go/API changes, per Global Constraints.
- **Placeholder scan:** no TBD/TODO; every step contains full code or exact text.
- **Type consistency:** `avgItemLevel(gear)` returns number (0 = none); `classColor`/`classIcon` accept the proto enum string (`ClassMage`); CharacterHeader accepts `gear` array of `{ itemId, ilvl }`; chip markup class names match the CSS in Task 3.
- **Risk:** `.character-facts` deletion leaves `dt`/`dd` shared selector usable by `.report-summary`/`.stat-list` — Step 2 of Task 3 explicitly keeps it. E2E class-color value `rgb(105, 204, 240)` = `#69ccf0` for the Mage fixture.
