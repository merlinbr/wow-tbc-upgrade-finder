# Armory Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rework the post-import armory into a character-centered Gear/Stats/Talents view with WoWSims-style item cards (ilvl badge, rarity-colored name, enchant effect line, sockets/gems) and a labeled 3D placeholder stage.

**Architecture:** The Go server keeps owning all item data; three additive response fields (`ilvl`, enchant `description`, `talentsString`) feed the existing Svelte 5 renderer. `ArmoryView` becomes a tabbed layout (Gear with two explicit slot columns + weapons strip around `CharacterStage`, Stats, read-only Talents). No new dependencies; the 3D viewer itself is deferred behind the permission/asset-access gates documented in `docs/armory-redesign-research.md`.

**Tech Stack:** Go (server, `wowsims/tbc` fork), Svelte 5 + Vite (`ui-finder/`), Playwright e2e, Node `node:test`.

## Global Constraints

- Spec: `docs/superpowers/specs/2026-09-05-armory-redesign-design.md`. Research: `docs/armory-redesign-research.md`.
- No new frontend or Go dependencies. No changes to ranking, simulation, validation codes, or job lifecycle.
- The built Vite bundle is committed under `cmd/wowsimcli/cmd/upgrade_ui/`; `go build`/`go test` must not require Node. After every UI change, run `cd ui-finder && rtk npm run build` and commit the regenerated bundle with the source.
- Existing import response fields are never removed or renamed; new fields are additive.
- `data-region="armory-view"` and `data-slot` attributes stay so the Playwright contract keeps working; existing e2e assertions are updated only where the redesign intentionally changes them.
- Run commands with `rtk` (repo convention): `rtk go test ./... -count=1`, `rtk npm run build`, `rtk npm run test:e2e`, `rtk node --test src/lib/<file>.test.js` (from `ui-finder/`).
- Canonical slot names from `cmd/wowsimcli/cmd/upgrades/armory.go`: `Head`, `Neck`, `Shoulder`, `Back`, `Chest`, `Wrist`, `Main Hand`, `Off Hand`, `Hands`, `Waist`, `Legs`, `Feet`, `Finger 1`, `Finger 2`, `Trinket 1`, `Trinket 2`, `Ranged`.
- Go tests run without the `with_db` tag; the `assets/enchants` package embeds its own JSON like `assets/database/loader.go` does.

---

### Task 1: Backend display fields (ilvl, enchant description, talentsString)

**Files:**
- Create: `assets/enchants/descriptions.go`
- Create: `assets/enchants/descriptions_test.go`
- Modify: `cmd/wowsimcli/cmd/upgrades/types.go:43-57` (GearSlotData), `:83-88` (EnchantData)
- Modify: `cmd/wowsimcli/cmd/upgrades/armory.go:1-10` (imports), `:121-126` (enrichItem fields), `:169-180` (enchant block)
- Modify: `cmd/wowsimcli/cmd/upgrades/upgrade_server.go:209-218` (importResponse), `:251-266` (handleImport response)
- Modify: `cmd/wowsimcli/cmd/upgrades/armory_test.go` (new test)
- Modify: `cmd/wowsimcli/cmd/upgrade_server_test.go:129-168` (extend import test)

**Interfaces:**
- Consumes: existing `enrichItem(spec, item, catalog)`, `importResponse`, `EnrichArmory` (all in place).
- Produces: `enchants.Descriptions() map[int32]string` (effect ID → display text, from `assets/enchants/descriptions.json`); `GearSlotData.Ilvl int32` (`json:"ilvl"`); `EnchantData.Description string` (`json:"description"`); `importResponse.TalentsString string` (`json:"talentsString"`).

- [ ] **Step 1: Write the failing tests**

Create `assets/enchants/descriptions_test.go`:

```go
package enchants

import "testing"

func TestDescriptionsResolveKnownEffects(t *testing.T) {
	descriptions := Descriptions()
	if descriptions[2673] != "Mongoose" {
		t.Fatalf("descriptions[2673] = %q, want %q", descriptions[2673], "Mongoose")
	}
	if descriptions[3003] != "+34 Attack Power and +16 Hit Rating" {
		t.Fatalf("descriptions[3003] = %q, want %q", descriptions[3003], "+34 Attack Power and +16 Hit Rating")
	}
}
```

Append to `cmd/wowsimcli/cmd/upgrades/armory_test.go` (reuses existing `mustImportFixture` and `NewCatalog(database.Load())` helpers):

```go
func TestEnrichArmoryExposesIlvlAndEnchantDescription(t *testing.T) {
	armory, err := EnrichArmory(mustImportFixture(t), NewCatalog(database.Load()))
	if err != nil {
		t.Fatal(err)
	}
	for _, slot := range armory.Gear {
		if slot.ItemID == 0 {
			continue
		}
		if slot.Ilvl <= 0 {
			t.Fatalf("slot %s item %d has ilvl %d, want > 0", slot.SlotName, slot.ItemID, slot.Ilvl)
		}
	}
	found := false
	for _, slot := range armory.Gear {
		if slot.Enchant == nil {
			continue
		}
		found = true
		if slot.Enchant.Description == "" {
			t.Fatalf("enchant %s has empty description", slot.Enchant.Name)
		}
	}
	if !found {
		t.Fatal("fixture has no enchanted slot to assert description")
	}
}
```

Extend `TestImportReturnsSummaryWithoutStartingJob` in `cmd/wowsimcli/cmd/upgrade_server_test.go`, after the existing `gear` length assertion (gear is `[]any`, already decoded):

```go
	if _, ok := payload["talentsString"].(string); !ok {
		t.Fatalf("talentsString = %#v, want string", payload["talentsString"])
	}
	firstGear, ok := gear[0].(map[string]any)
	if !ok {
		t.Fatal("gear[0] is not an object")
	}
	if ilvl, _ := firstGear["ilvl"].(float64); ilvl <= 0 {
		t.Fatalf("gear[0].ilvl = %v, want > 0", firstGear["ilvl"])
	}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `rtk go test ./assets/enchants -count=1 && rtk go test ./cmd/wowsimcli/cmd/upgrades -count=1 && rtk go test ./cmd/wowsimcli/cmd -count=1`

Expected: `assets/enchants` fails to compile (no such package); the armory test fails (`slot.Ilvl <= 0`); the server test fails (`talentsString` missing).

- [ ] **Step 3: Implement the description loader**

Create `assets/enchants/descriptions.go` (package `enchants`, mirroring `assets/database/loader.go`):

```go
package enchants

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed descriptions.json
var descriptionsJSON []byte

// Descriptions maps enchant effect IDs to the effect text shown on the
// wowsims item display (e.g. "2673": "Mongoose").
func Descriptions() map[int32]string {
	result := make(map[int32]string)
	if err := json.Unmarshal(descriptionsJSON, &result); err != nil {
		panic(fmt.Errorf("unmarshal enchant descriptions: %w", err))
	}
	return result
}
```

- [ ] **Step 4: Add the display fields**

In `cmd/wowsimcli/cmd/upgrades/types.go`:

- `GearSlotData` gains, after `Phase`: `Ilvl int32 \`json:"ilvl"\``
- `EnchantData` gains, after `Name`: `Description string \`json:"description"\``

In `cmd/wowsimcli/cmd/upgrades/armory.go`:

- Add imports: `"sync"` and `"github.com/wowsims/tbc/assets/enchants"`.
- Near the top of the file (after `canonicalGearSlots`):

```go
var enchantDescriptions = sync.OnceValue(func() map[int32]string { return enchants.Descriptions() })
```

- In `enrichItem`, after `data.SetName = item.GetSetName()` add: `data.Ilvl = item.GetIlvl()`
- Replace the enchant block's `data.Enchant` assignment with:

```go
	description := enchantDescriptions()[enchant.GetEffectId()]
	if description == "" {
		description = enchant.GetName()
	}
	data.Enchant = &EnchantData{
		ID:          enchant.GetEffectId(),
		Name:        enchant.GetName(),
		Icon:        enchant.GetIcon(),
		Description: description,
		Stats:       statMap(enchant.GetStats()),
	}
```

In `cmd/wowsimcli/cmd/upgrades/upgrade_server.go`:

- `importResponse` gains, after `DerivedStats`: `TalentsString string \`json:"talentsString"\``
- In `handleImport`'s `writeJSON` call, add: `TalentsString: imported.Settings.Player.GetTalentsString(),`

- [ ] **Step 5: Run tests to verify they pass**

Run: `rtk go test ./assets/enchants -count=1 && rtk go test ./cmd/wowsimcli/cmd/upgrades -count=1 && rtk go test ./cmd/wowsimcli/cmd -count=1`

Expected: PASS (descriptions resolve; fixture items all have positive ilvl; enchanted slot has a non-empty description; import response carries `talentsString` and `gear[0].ilvl`).

- [ ] **Step 6: Commit**

```bash
git add assets/enchants/descriptions.go assets/enchants/descriptions_test.go cmd/wowsimcli/cmd/upgrades/types.go cmd/wowsimcli/cmd/upgrades/armory.go cmd/wowsimcli/cmd/upgrades/armory_test.go cmd/wowsimcli/cmd/upgrades/upgrade_server.go cmd/wowsimcli/cmd/upgrade_server_test.go
git commit -m "feat: expose ilvl, enchant description, and talents string for armory redesign"
```

---

### Task 2: Talent string decoder and read-only tree view

**Files:**
- Create: `ui-finder/src/lib/talents.js`
- Create: `ui-finder/src/lib/talents.test.js`
- Create: `ui-finder/src/lib/TalentTrees.svelte`

**Interfaces:**
- Consumes: `imported.talentsString` (Task 1) and `character.class` (e.g. `ClassPaladin`) from the existing import payload; bundled tree JSON under `ui/core/talents/trees/*.json` (name, `backgroundUrl`, `talents[]` with `fancyName`, `location.rowIdx/colIdx`, `maxPoints`, ordered per the upstream `newTalentsConfig` convention).
- Produces: `decodeTalentsString(talentsString, treeCount) → string[]`, `rankAt(treeRanks, index) → number`, `treePoints(treeRanks) → number`; `TalentTrees.svelte` with props `{ class, talentsString }`, rendering one `.talent-tree` section per tree.

- [ ] **Step 1: Write the failing node test**

Create `ui-finder/src/lib/talents.test.js`:

```js
import assert from 'node:assert/strict';
import test from 'node:test';
import { decodeTalentsString, rankAt, treePoints } from './talents.js';

test('decodes hyphen-separated trees and pads missing trees', () => {
  assert.deepEqual(decodeTalentsString('321-5-0'), ['321', '5', '0']);
  assert.deepEqual(decodeTalentsString(''), ['', '', '']);
  assert.deepEqual(decodeTalentsString('321'), ['321', '', '']);
});

test('reads a rank at an index and treats missing digits as zero', () => {
  assert.equal(rankAt('321', 0), 3);
  assert.equal(rankAt('321', 2), 1);
  assert.equal(rankAt('12', 4), 0);
  assert.equal(rankAt('', 0), 0);
});

test('sums points across a tree', () => {
  assert.equal(treePoints('321'), 6);
  assert.equal(treePoints(''), 0);
});
```

- [ ] **Step 2: Run test to verify it fails**

Run (from `ui-finder/`): `rtk node --test src/lib/talents.test.js`

Expected: FAIL — `Cannot find module './talents.js'`.

- [ ] **Step 3: Implement the decoder**

Create `ui-finder/src/lib/talents.js`:

```js
// Splits a Wowhead talent string into one rank string per tree.
// Trees are hyphen-separated; each character is one talent's rank, in the
// array order used by ui/core/talents/trees/*.json. Missing trees become ''.
export function decodeTalentsString(talentsString, treeCount = 3) {
  const parts = String(talentsString ?? '').split('-');
  return Array.from({ length: treeCount }, (_, index) => parts[index] ?? '');
}

// Rank allocated at one talent index; missing digits count as zero.
export function rankAt(treeRanks, index) {
  const value = Number(String(treeRanks ?? '').charAt(index));
  return Number.isInteger(value) && value > 0 ? value : 0;
}

// Total points allocated in one tree's rank string.
export function treePoints(treeRanks) {
  let sum = 0;
  for (const digit of String(treeRanks ?? '')) {
    const value = Number(digit);
    if (Number.isInteger(value)) sum += value;
  }
  return sum;
}
```

- [ ] **Step 4: Run test to verify it passes**

Run (from `ui-finder/`): `rtk node --test src/lib/talents.test.js`

Expected: PASS (3 tests).

- [ ] **Step 5: Implement the read-only tree view**

Create `ui-finder/src/lib/TalentTrees.svelte`:

```svelte
<script>
  import { decodeTalentsString, rankAt, treePoints } from './talents.js';
  import druidTrees from '../../../ui/core/talents/trees/druid.json';
  import hunterTrees from '../../../ui/core/talents/trees/hunter.json';
  import mageTrees from '../../../ui/core/talents/trees/mage.json';
  import paladinTrees from '../../../ui/core/talents/trees/paladin.json';
  import priestTrees from '../../../ui/core/talents/trees/priest.json';
  import rogueTrees from '../../../ui/core/talents/trees/rogue.json';
  import shamanTrees from '../../../ui/core/talents/trees/shaman.json';
  import warlockTrees from '../../../ui/core/talents/trees/warlock.json';
  import warriorTrees from '../../../ui/core/talents/trees/warrior.json';

  const treesByClass = {
    ClassDruid: druidTrees,
    ClassHunter: hunterTrees,
    ClassMage: mageTrees,
    ClassPaladin: paladinTrees,
    ClassPriest: priestTrees,
    ClassRogue: rogueTrees,
    ClassShaman: shamanTrees,
    ClassWarlock: warlockTrees,
    ClassWarrior: warriorTrees,
  };

  let { class: playerClass = '', talentsString = '' } = $props();

  let ranksByTree = $derived(decodeTalentsString(talentsString, 3));

  let treeData = $derived((treesByClass[playerClass] ?? []).map((tree, treeIndex) => {
    const ranks = ranksByTree[treeIndex] ?? '';
    const maxRow = Math.max(...tree.talents.map((talent) => talent.location.rowIdx), 0);
    const maxCol = Math.max(...tree.talents.map((talent) => talent.location.colIdx), 0);
    const byLocation = new Map(
      tree.talents.map((talent, index) => [
        `${talent.location.rowIdx}:${talent.location.colIdx}`,
        { ...talent, index, points: rankAt(ranks, index) },
      ]),
    );
    const cells = [];
    for (let row = 0; row <= maxRow; row++) {
      for (let col = 0; col <= maxCol; col++) {
        cells.push(byLocation.get(`${row}:${col}`) ?? null);
      }
    }
    return {
      name: tree.name,
      backgroundUrl: tree.backgroundUrl,
      points: treePoints(ranks),
      cols: maxCol + 1,
      cells,
    };
  }));
</script>

{#if treeData.length}
  <div class="talent-trees" data-region="talent-trees">
    {#each treeData as tree}
      <section class="talent-tree" style="--talent-cols: {tree.cols}" aria-label="{tree.name} talents">
        <div class="talent-tree-header">
          <h3>{tree.name}</h3>
          <span class="talent-tree-points">{tree.points} points</span>
        </div>
        <div class="talent-tree-body" style:background-image={tree.backgroundUrl ? `url('${tree.backgroundUrl}')` : 'none'}>
          {#each tree.cells as cell, cellIndex (cell?.index ?? `empty-${cellIndex}`)}
            {#if cell}
              <div class="talent-cell" class:empty-rank={cell.points === 0} title="{cell.fancyName} {cell.points}/{cell.maxPoints}">
                <span class="talent-name">{cell.fancyName}</span>
                <span class="rank-pips" aria-label="{cell.points} of {cell.maxPoints} points">
                  {#each Array(cell.maxPoints) as _, pipIndex}
                    <span class:filled={pipIndex < cell.points} class="rank-pip" aria-hidden="true"></span>
                  {/each}
                </span>
              </div>
            {:else}
              <div class="talent-cell talent-cell-empty" aria-hidden="true"></div>
            {/if}
          {/each}
        </div>
      </section>
    {/each}
  </div>
{:else}
  <p class="muted">No talent data for this class.</p>
{/if}
```

- [ ] **Step 6: Add the tree styles**

Append to `ui-finder/src/app.css`:

```css
.talent-trees { display: grid; gap: 1rem; margin-top: 1rem; }
.talent-tree { background: rgba(5, 9, 15, 0.55); border: 1px solid var(--border); border-radius: 7px; padding: 0.9rem; }
.talent-tree-header { align-items: baseline; display: flex; gap: 1rem; justify-content: space-between; margin-bottom: 0.6rem; }
.talent-tree-header h3 { margin: 0; }
.talent-tree-points { color: var(--gold); font-variant-numeric: tabular-nums; }
.talent-tree-body { background-position: center; background-size: cover; border: 1px solid var(--border); border-radius: 5px; display: grid; gap: 2px; grid-template-columns: repeat(var(--talent-cols), minmax(0, 1fr)); padding: 0.4rem; }
.talent-cell { background: rgba(9, 13, 20, 0.88); border: 1px solid var(--border); border-radius: 4px; min-height: 3.4rem; padding: 0.3rem; }
.talent-cell-empty { background: transparent; border-color: transparent; }
.talent-cell.empty-rank { opacity: 0.55; }
.talent-name { color: var(--text); display: block; font-size: 0.72rem; line-height: 1.25; }
.rank-pips { display: flex; gap: 2px; margin-top: 0.25rem; }
.rank-pip { background: #3a4a63; border-radius: 50%; height: 0.5rem; width: 0.5rem; }
.rank-pip.filled { background: var(--gold); }
```

- [ ] **Step 7: Build and test**

Run (from `ui-finder/`): `rtk npm run build && rtk node --test src/lib/talents.test.js`

Expected: build succeeds; 3 tests pass. Note: `TalentTrees.svelte` is not rendered anywhere yet, so the JSON/component compilation is exercised fully in Task 4; this task verifies the decoder and that the build stays green.

- [ ] **Step 8: Commit**

```bash
git add ui-finder/src/lib/talents.js ui-finder/src/lib/talents.test.js ui-finder/src/lib/TalentTrees.svelte ui-finder/src/app.css
git commit -m "feat: read-only talent tree view for armory redesign"
```

---

### Task 3: WoWSims-style gear card

**Files:**
- Modify: `ui-finder/src/lib/GearSlot.svelte` (full rewrite of the card)
- Modify: `ui-finder/src/app.css:104-132` (gear card block; delete `.socket-row`, old `.socket`, `.enchant`, `.item-meta`, `.socket-bonus`; add ilvl badge, socket strip, enchant effect, quality text, details)
- Modify: `ui-finder/e2e/armory.spec.js` (enchant + ilvl assertions)
- Rebuild: `cmd/wowsimcli/cmd/upgrade_ui/` (via `npm run build`)

**Interfaces:**
- Consumes: `slot` object from `/api/import` `gear[]` (fields incl. new `ilvl`, `enchant.description` from Task 1); existing `side` usage is introduced in Task 4 — this task adds the optional `side = 'left'` prop now so Task 4 only flips CSS.
- Produces: `.item-ilvl` badge, `.quality-text-N` name, `.enchant-effect` line, `.socket-strip` with `.socket`, `.gear-details` disclosure; keeps `data-slot` and the `{slotName} sockets` aria-label.

- [ ] **Step 1: Rewrite the gear card**

Replace the contents of `ui-finder/src/lib/GearSlot.svelte` with:

```svelte
<script>
  let { slot, side = 'left' } = $props();
  let itemIconFailed = $state(false);
  let failedGemIcons = $state({});

  const socketColors = {
    0: 'Unknown',
    1: 'Meta',
    2: 'Red',
    3: 'Blue',
    4: 'Yellow',
    5: 'Green',
    6: 'Orange',
    7: 'Purple',
    8: 'Prismatic',
  };

  function iconUrl(icon) {
    return `https://wow.zamimg.com/images/wow/icons/large/${icon}.jpg`;
  }

  function displayStats(stats) {
    return Object.entries(stats ?? {}).map(([key, value]) => `${key.replaceAll('_', ' ')} ${Number(value).toFixed(2).replace(/\.00$/, '')}`).join(', ');
  }

  function handleItemImageError() {
    itemIconFailed = true;
  }

  function handleGemImageError(index) {
    failedGemIcons[index] = true;
  }

  function slotInitial() {
    return slot.slotName?.trim()?.[0] || '?';
  }
</script>

<article class="gear-slot quality-{slot.quality ?? 0}" class:mirror={side === 'right'} data-slot={slot.slotName}>
  <div class="gear-icon-wrap">
    {#if slot.ilvl}
      <span class="item-ilvl">{slot.ilvl}</span>
    {/if}
    {#if slot.icon && !itemIconFailed}
      <img class="gear-icon" src={iconUrl(slot.icon)} alt="{slot.itemName || 'Empty'} icon" onerror={handleItemImageError} />
    {:else}
      <span class="gear-icon-fallback" role="img" aria-label="{slot.slotName} icon unavailable">{slotInitial()}</span>
    {/if}
    {#if slot.sockets?.length}
      <div class="socket-strip" aria-label={`${slot.slotName} sockets`}>
        {#each slot.sockets as socket, index}
          <span class:empty-socket={!socket.gem} class="socket" title={`${socketColors[socket.color] || 'Unknown'} socket${socket.gem ? `: ${socket.gem.name}` : ': empty'}`}>
            {#if socket.gem?.icon && !failedGemIcons[index]}
              <img src={iconUrl(socket.gem.icon)} alt="{socket.gem.name} gem" onerror={() => handleGemImageError(index)} />
            {:else}
              <span aria-hidden="true">◇</span>
            {/if}
          </span>
        {/each}
      </div>
    {/if}
  </div>
  <div class="gear-copy">
    {#if slot.itemName}
      <h3 class="item-name quality-text-{slot.quality ?? 0}">{slot.itemName}</h3>
      {#if slot.enchant}
        <p class="enchant-effect">{slot.enchant.description || slot.enchant.name}</p>
      {/if}
    {:else}
      <p class="slot-caption">{slot.slotName}</p>
      <p class="item-name muted">Empty slot</p>
    {/if}
    <details class="gear-details">
      <summary>Details</summary>
      <dl class="gear-detail-list">
        <div><dt>Slot</dt><dd>{slot.slotName}</dd></div>
        {#if slot.itemId}
          <div><dt>Item</dt><dd>{slot.itemId} · Phase {slot.phase || '—'}</dd></div>
        {/if}
        {#if slot.setName}
          <div><dt>Set</dt><dd>{slot.setName}</dd></div>
        {/if}
        {#if slot.enchant}
          <div><dt>Enchant</dt><dd>{slot.enchant.name}</dd></div>
        {/if}
        {#if displayStats(slot.stats)}
          <div><dt>Stats</dt><dd>{displayStats(slot.stats)}</dd></div>
        {/if}
        {#if slot.socketBonus?.stats}
          <div><dt>Socket bonus</dt><dd class:bonus-active={slot.socketBonus.active} class:bonus-inactive={!slot.socketBonus.active}>{slot.socketBonus.active ? 'active' : 'inactive'}{displayStats(slot.socketBonus.stats) ? ` · ${displayStats(slot.socketBonus.stats)}` : ''}</dd></div>
        {/if}
      </dl>
    </details>
  </div>
</article>
```

- [ ] **Step 2: Update the gear card styles**

In `ui-finder/src/app.css`, replace the block from `.gear-slot {` through `.empty-socket { ... }` (the card, socket row, and socket rules) with:

```css
.gear-slot { align-items: start; background: rgba(6, 10, 16, 0.65); border: 1px solid var(--border); border-left: 3px solid var(--border-bright); border-radius: 7px; display: flex; gap: 0.7rem; min-height: 5rem; padding: 0.65rem; }
.gear-slot.quality-0 { border-color: #9d9d9d; border-left-color: #9d9d9d; }
.gear-slot.quality-1 { border-color: #ffffff; border-left-color: #ffffff; }
.gear-slot.quality-2 { border-color: #1eff00; border-left-color: #1eff00; }
.gear-slot.quality-3 { border-color: #0070dd; border-left-color: #0070dd; }
.gear-slot.quality-4 { border-color: #a335ee; border-left-color: #a335ee; }
.gear-slot.quality-5 { border-color: #ff8000; border-left-color: #ff8000; }
.gear-slot.quality-6 { border-color: #e6cc80; border-left-color: #e6cc80; }
.gear-slot.quality-7 { border-color: #00ccff; border-left-color: #00ccff; }
.gear-slot.mirror { flex-direction: row-reverse; }
.gear-icon-wrap { flex: 0 0 3.4rem; height: 3.4rem; position: relative; }
.gear-icon, .gear-icon-fallback { border: 1px solid var(--border-bright); border-radius: 4px; display: block; height: 3.4rem; object-fit: cover; width: 3.4rem; }
.gear-icon-fallback { align-items: center; background: #263247; color: var(--gold); display: flex; font-size: 1.3rem; justify-content: center; }
.quality-0 .gear-icon-fallback { background: #454545; }
.quality-1 .gear-icon-fallback { background: #777; color: #fff; }
.quality-2 .gear-icon-fallback { background: #176b13; }
.quality-3 .gear-icon-fallback { background: #124a88; }
.quality-4 .gear-icon-fallback { background: #572080; }
.quality-5 .gear-icon-fallback { background: #8c4900; }
.quality-6 .gear-icon-fallback { background: #77652d; }
.quality-7 .gear-icon-fallback { background: #126b80; }
.item-ilvl { background: #101826; border: 1px solid var(--gold); border-radius: 4px; color: var(--gold); font-size: 0.68rem; font-weight: 700; left: -6px; line-height: 1; padding: 2px 4px; position: absolute; top: -6px; z-index: 1; }
.socket-strip { bottom: -5px; display: flex; gap: 2px; justify-content: center; left: 0; position: absolute; right: 0; }
.socket { align-items: center; background: #263247; border: 1px solid var(--border-bright); border-radius: 3px; display: inline-flex; height: 1rem; justify-content: center; overflow: hidden; width: 1rem; }
.socket img { height: 100%; object-fit: cover; width: 100%; }
.empty-socket { background: transparent; border: 1px dashed var(--gold); color: var(--gold); }
.gear-copy { min-width: 0; }
.item-name { font-weight: 700; margin: 0 0 0.12rem; overflow-wrap: anywhere; }
.quality-text-0 { color: #9d9d9d; }
.quality-text-1 { color: #ffffff; }
.quality-text-2 { color: #1eff00; }
.quality-text-3 { color: #0070dd; }
.quality-text-4 { color: #a335ee; }
.quality-text-5 { color: #ff8000; }
.quality-text-6 { color: #e6cc80; }
.quality-text-7 { color: #00ccff; }
.slot-caption { color: var(--muted); font-size: 0.72rem; letter-spacing: 0.05em; margin: 0; text-transform: uppercase; }
.enchant-effect { color: #1eff00; font-size: 0.8rem; margin: 0.1rem 0 0; overflow-wrap: anywhere; }
.gear-details { color: var(--muted); font-size: 0.78rem; margin-top: 0.35rem; }
.gear-details summary { cursor: pointer; }
.gear-detail-list { display: grid; gap: 0.15rem; margin: 0.4rem 0 0; }
.gear-detail-list div { display: flex; gap: 0.6rem; justify-content: space-between; }
.gear-detail-list dt { color: var(--muted); }
.gear-detail-list dd { margin: 0; overflow-wrap: anywhere; text-align: right; }
.bonus-active { color: var(--success); }
.bonus-inactive { color: #dcae78; }
```

Delete the now-unused old rules if any remain: `.item-meta`, `.enchant`, `.socket-bonus`, `.socket-row`.

- [ ] **Step 3: Update the e2e assertions**

In `ui-finder/e2e/armory.spec.js`, replace:

```js
  await expect(page.getByText(/^Enchant:/).first()).toBeVisible();
```

with:

```js
  await expect(page.locator('.enchant-effect').first()).toBeVisible();
  await expect(page.locator('.item-ilvl').first()).toBeVisible();
```

Keep the existing `page.getByLabel(/sockets$/i)` assertion (the `{slotName} sockets` aria-label is preserved on `.socket-strip`).

- [ ] **Step 4: Rebuild the bundle and run the e2e suite**

Run (from repo root): `cd ui-finder && rtk npm run build && rtk npm run test:e2e`

Expected: build succeeds; the single Chromium test passes (armory renders with ilvl badges and enchant effect lines, then the full ranking/copy/cancel flow).

- [ ] **Step 5: Commit**

```bash
git add ui-finder/src/lib/GearSlot.svelte ui-finder/src/app.css ui-finder/e2e/armory.spec.js cmd/wowsimcli/cmd/upgrade_ui
git commit -m "feat: wowsims-style gear cards with ilvl, enchant effects, and sockets"
```

---

### Task 4: Character-centered armory layout with tabs

**Files:**
- Create: `ui-finder/src/lib/CharacterHeader.svelte`
- Create: `ui-finder/src/lib/CharacterStage.svelte`
- Modify: `ui-finder/src/lib/ArmoryView.svelte` (full rewrite: header, tabs, stage, explicit columns, weapons strip)
- Modify: `ui-finder/src/app.css` (header actions, tabs, stage, weapons strip, grid; delete `.gear-center` rules)
- Modify: `ui-finder/e2e/armory.spec.js` (tab navigation assertions)
- Rebuild: `cmd/wowsimcli/cmd/upgrade_ui/`

**Interfaces:**
- Consumes: `imported` (character, defaults.maxPhase, settingsDigest, simulatorRevision, databaseRevision, gear, stats, derivedStats, talentsString); `GearSlot` with `side` prop (Task 3); `TalentTrees` (Task 2).
- Produces: `.character-header`, `#armory-heading` (keeps `TestMage` heading contract), `a.find-upgrades[href="#ranking-heading"]`, `.armory-tabs` with tabs Gear/Stats/Talents, `.character-stage[data-region="character-stage"]` with `Activate 3D` disabled control, `.weapon-strip`, gear columns keyed by explicit slot names.

- [ ] **Step 1: Create the character header**

Create `ui-finder/src/lib/CharacterHeader.svelte`:

```svelte
<script>
  import { humanizeEnum } from './labels.js';

  let { character = {}, phase = 0, settingsDigest = '', simulatorRevision = '', databaseRevision = '' } = $props();

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

  function digest(value) {
    if (!value) return '—';
    return `${value.slice(0, 16)}…`;
  }
</script>

<div class="character-header">
  <div class="character-identity">
    <div class="section-kicker">Imported character</div>
    <h2 id="armory-heading">{character.name || 'Unnamed character'}</h2>
    <p class="character-subtitle">Level 70 {humanizeEnum(character.race, 'Race') || 'Unknown race'} · {character.spec ? humanizeEnum(character.spec) : humanizeEnum(character.class, 'Class') || 'Unknown class'}</p>
    <dl class="character-facts">
      <div><dt>Professions</dt><dd>{professions.length ? professions.join(', ') : 'None'}</dd></div>
      <div><dt>Phase</dt><dd>{phase || '—'}</dd></div>
    </dl>
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

- [ ] **Step 2: Create the character stage**

Create `ui-finder/src/lib/CharacterStage.svelte`:

```svelte
<script>
  const backdrops = import.meta.glob('../../../assets/img/*.jpg', { eager: true, import: 'default' });

  const backdropFor = {
    BalanceDruid: 'balance_druid_background.jpg',
    FeralCatDruid: 'feral_druid_background.jpg',
    FeralBearDruid: 'feral_druid_tank_background.jpg',
    RestorationDruid: 'resto_druid_background.jpg',
    Hunter: 'hunter_background.jpg',
    Mage: 'mage_background.jpg',
    HolyPaladin: 'holy_paladin.jpg',
    ProtectionPaladin: 'prot_paladin.jpg',
    RetributionPaladin: 'retribution_paladin.jpg',
    Priest: 'healing_priest_background.jpg',
    Rogue: 'rogue_background.jpg',
    ElementalShaman: 'elemental_shaman_background.jpg',
    EnhancementShaman: 'enhancement_shaman_background.jpg',
    RestorationShaman: 'resto_shaman_background.jpg',
    Warlock: 'warlock_background.jpg',
    DpsWarrior: 'warrior_background.jpg',
    ProtectionWarrior: 'warrior_background.jpg',
  };

  let { race = '', class: playerClass = '', spec = '' } = $props();

  let backdropUrl = $derived.by(() => {
    const fileName = backdropFor[spec] ?? backdropFor[playerClass];
    if (!fileName) return '';
    const entry = Object.entries(backdrops).find(([path]) => path.endsWith(`/${fileName}`));
    return entry?.[1] ?? '';
  });
</script>

<div class="character-stage" data-region="character-stage">
  <div class="stage-backdrop" aria-hidden="true" style:background-image={backdropUrl ? `url('${backdropUrl}')` : 'none'}></div>
  <div class="stage-placeholder">
    <span class="stage-kicker">Character preview</span>
    <span class="stage-note">Appearance not imported</span>
    <button type="button" class="secondary-button" disabled title="3D model integration is not available yet">Activate 3D</button>
  </div>
</div>
```

Note: the unused `race` prop is accepted intentionally; a future gated viewer integration selects the model by race. The `<button>` is disabled by design — it documents the deferred capability without promising function.

- [ ] **Step 3: Rewrite the armory view**

Replace the contents of `ui-finder/src/lib/ArmoryView.svelte` with:

```svelte
<script>
  import CharacterHeader from './CharacterHeader.svelte';
  import CharacterStage from './CharacterStage.svelte';
  import GearSlot from './GearSlot.svelte';
  import StatPanels from './StatPanels.svelte';
  import TalentTrees from './TalentTrees.svelte';

  let { imported } = $props();
  let character = $derived(imported?.character ?? {});
  let gear = $derived(Array.isArray(imported?.gear) ? imported.gear : []);
  let phase = $derived(imported?.defaults?.maxPhase ?? 0);
  let activeTab = $state('gear');

  const tabs = [
    { id: 'gear', label: 'Gear' },
    { id: 'stats', label: 'Stats' },
    { id: 'talents', label: 'Talents' },
  ];

  const leftColumn = ['Head', 'Neck', 'Shoulder', 'Back', 'Chest', 'Wrist', 'Hands'];
  const rightColumn = ['Waist', 'Legs', 'Feet', 'Finger 1', 'Finger 2', 'Trinket 1', 'Trinket 2'];
  const weaponSlots = ['Main Hand', 'Off Hand', 'Ranged'];

  let gearBySlot = $derived(new Map(gear.map((slot) => [slot.slotName, slot])));

  function slotsFor(names) {
    return names.map((name) => gearBySlot.get(name)).filter(Boolean);
  }

  function gearKey(slot) {
    const gems = (slot.sockets ?? []).map((socket) => socket.gem?.id ?? '').join(',');
    return `${slot.slotName}:${slot.itemId ?? ''}:${slot.icon ?? ''}:${gems}:${slot.enchant?.id ?? ''}`;
  }
</script>

<section class="panel armory-panel" aria-labelledby="armory-heading" data-region="armory-view">
  <CharacterHeader {character} {phase} settingsDigest={imported?.settingsDigest} simulatorRevision={imported?.simulatorRevision} databaseRevision={imported?.databaseRevision} />

  <div class="armory-tabs" role="tablist" aria-label="Armory views">
    {#each tabs as tab}
      <button id="tab-{tab.id}" class="armory-tab" class:active={activeTab === tab.id} role="tab" aria-controls="panel-{tab.id}" aria-selected={activeTab === tab.id} onclick={() => (activeTab = tab.id)}>{tab.label}</button>
    {/each}
  </div>

  {#if activeTab === 'gear'}
    <div id="panel-gear" role="tabpanel" aria-labelledby="tab-gear">
      <div class="gear-grid" aria-label="Equipped gear">
        <div class="gear-column gear-column-left">
          {#each slotsFor(leftColumn) as slot (gearKey(slot))}
            <GearSlot {slot} side="left" />
          {/each}
        </div>
        <CharacterStage race={character.race} class={character.class} spec={character.spec} />
        <div class="gear-column gear-column-right">
          {#each slotsFor(rightColumn) as slot (gearKey(slot))}
            <GearSlot {slot} side="right" />
          {/each}
        </div>
      </div>
      <div class="weapon-strip" aria-label="Weapons">
        {#each slotsFor(weaponSlots) as slot (gearKey(slot))}
          <GearSlot {slot} />
        {/each}
      </div>
    </div>
  {:else if activeTab === 'stats'}
    <div id="panel-stats" role="tabpanel" aria-labelledby="tab-stats">
      <StatPanels stats={imported?.stats} derivedStats={imported?.derivedStats} />
    </div>
  {:else}
    <div id="panel-talents" role="tabpanel" aria-labelledby="tab-talents">
      <TalentTrees class={character.class} talentsString={imported?.talentsString} />
    </div>
  {/if}
</section>
```

- [ ] **Step 4: Update the layout styles**

In `ui-finder/src/app.css`:

- Replace the `.gear-grid` and `.gear-center` rules (lines ~101-103) with:

```css
.gear-grid { align-items: start; display: grid; gap: 1rem; grid-template-columns: minmax(0, 1fr) minmax(16rem, 22rem) minmax(0, 1fr); margin-top: 1rem; }
.gear-column { display: grid; align-content: start; gap: 0.7rem; }
.weapon-strip { display: grid; gap: 0.7rem; grid-template-columns: repeat(3, minmax(0, 1fr)); margin-top: 1rem; }
```

- Append the header/action/tabs/stage styles:

```css
.character-identity { min-width: 0; }
.character-actions { display: grid; gap: 0.75rem; justify-items: end; }
.find-upgrades { background: var(--accent); border-radius: 6px; color: #0a1420; font-weight: 700; padding: 0.5rem 1rem; text-decoration: none; }
.import-details { color: var(--muted); font-size: 0.8rem; max-width: 16rem; }
.import-details summary { cursor: pointer; }
.import-details dl { display: grid; gap: 0.2rem; margin: 0.4rem 0 0; }
.import-details dt { color: var(--muted); font-size: 0.72rem; text-transform: uppercase; }
.import-details dd { margin: 0; overflow-wrap: anywhere; }
.armory-tabs { border-bottom: 1px solid var(--border); display: flex; gap: 0.5rem; margin-top: 1.25rem; }
.armory-tab { background: transparent; border: none; border-bottom: 2px solid transparent; border-radius: 0; color: var(--muted); padding: 0.55rem 0.9rem; }
.armory-tab.active { border-bottom-color: var(--accent); color: var(--text); }
.character-stage { background: linear-gradient(145deg, #182130, #0b1019); border: 1px solid var(--border); border-radius: 10px; min-height: clamp(14rem, 34vw, 26rem); overflow: hidden; position: relative; }
.stage-backdrop { background-position: center; background-size: cover; inset: 0; opacity: 0.35; position: absolute; }
.stage-placeholder { display: grid; gap: 0.3rem; inset: 0; place-content: center; position: absolute; text-align: center; }
.stage-kicker { color: var(--gold); font-weight: 700; }
.stage-note { color: var(--muted); font-size: 0.85rem; }
```

- Replace the `@media (max-width: 900px)` block's `.gear-grid` rule with:

```css
  .gear-grid { grid-template-columns: minmax(0, 1fr) 14rem minmax(0, 1fr); }
  .character-header { flex-direction: column; }
  .character-actions { justify-items: start; }
```

(Keep the existing `.control-grid` and `.character-facts` rules in that block.)

- In the `@media (max-width: 640px)` block, replace the `.gear-grid` rule with:

```css
  .gear-grid { grid-template-columns: minmax(0, 1fr); }
  .character-stage { order: 2; }
```

and add after it:

```css
  .weapon-strip { grid-template-columns: minmax(0, 1fr); }
```

- Delete the now-dead `.gear-center` and `.gear-column-right` rules from both media blocks (the center track no longer exists).

- Simplify the existing ≤900px responsive rule for the stage: the stage compacts via `clamp()` instead of a show/hide toggle. This is a deliberate simplification of the spec's "optional compact panel" wording — no toggle UI; the stage always compacts at narrow widths.

- [ ] **Step 5: Update the e2e test for tabs**

In `ui-finder/e2e/armory.spec.js`, replace the stat assertions block:

```js
  await expect(page.getByRole('heading', { name: 'Raw stats', exact: true })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Derived percentages', exact: true })).toBeVisible();
  await expect(page.getByText('raid buffed (link settings)', { exact: true })).toBeVisible();
```

with:

```js
  await expect(page.getByRole('link', { name: 'Find upgrades', exact: true })).toBeVisible();
  await page.getByRole('tab', { name: 'Stats', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Raw stats', exact: true })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Derived percentages', exact: true })).toBeVisible();
  await expect(page.getByText('raid buffed (link settings)', { exact: true })).toBeVisible();
  await page.getByRole('tab', { name: 'Talents', exact: true }).click();
  await expect(page.locator('.talent-tree')).toHaveCount(3);
  await page.getByRole('tab', { name: 'Gear', exact: true }).click();
```

The `[data-slot]` count assertion (17) runs before these tab switches on the default Gear tab and stays valid.

- [ ] **Step 6: Rebuild the bundle and run the e2e suite**

Run (from repo root): `cd ui-finder && rtk npm run build && rtk npm run test:e2e`

Expected: build succeeds (nine talent JSONs and the backdrop globs resolve); the e2e test passes (header, tabs, 17 slots, stats/talents tabs, ranking flow).

- [ ] **Step 7: Commit**

```bash
git add ui-finder/src/lib/CharacterHeader.svelte ui-finder/src/lib/CharacterStage.svelte ui-finder/src/lib/ArmoryView.svelte ui-finder/src/app.css ui-finder/e2e/armory.spec.js cmd/wowsimcli/cmd/upgrade_ui
git commit -m "feat: character-centered armory layout with gear, stats, and talents tabs"
```

---

### Task 5: Documentation, full verification, and manual smoke

**Files:**
- Modify: `docs/upgrade-finder.md:46-60` (Armory review section)
- Modify: `docs/STATE.md` (current status and intentional boundaries)

**Interfaces:**
- Consumes: the completed redesign from Tasks 1–4.
- Produces: accurate operator documentation for the new armory surface.

- [ ] **Step 1: Update the armory docs**

In `docs/upgrade-finder.md`, replace the "Armory review" section's description with:

```markdown
After import, review the complete canonical armory before ranking: compact
character header, all 17 gear slots around a labeled character-preview stage,
and Gear / Stats / Talents tabs. Each equipped item card shows the item-level
badge, rarity-colored name, enchant effect line (e.g. `+34 Attack Power and
+16 Hit Rating`), and socketed gems; item IDs, phase, set, and full stats are
in the card's Details disclosure. Stats shows the deterministic **raid buffed
(link settings)** snapshot. Talents renders the imported build read-only from
the bundled TBC talent trees. The 3D character model is not available yet;
the stage shows a labeled placeholder until the gated viewer integration
lands (see `docs/armory-redesign-research.md`).
```

- [ ] **Step 2: Update the project state notes**

In `docs/STATE.md`:

- Under "Current status", add: "The armory redesign is implemented: character-centered layout with Gear/Stats/Talents tabs, WoWSims-style item cards (ilvl badge, rarity-colored names, enchant effect lines, sockets/gems, details disclosures), a labeled 3D placeholder stage with per-spec backdrop art, and a read-only talents view driven by the newly exposed talents string."
- Under "Intentional boundaries", replace "No 3D model, talents tab, item hover tooltips, ..." with "No interactive 3D model (labeled placeholder stage only; viewer gated on renderer permission and browser asset access), no talent editing, no item hover tooltips, ..." keeping the rest of the line.

- [ ] **Step 3: Run the full verification suite**

Run (from repo root):

```bash
rtk go test ./assets/enchants -count=1
rtk go test ./cmd/wowsimcli/cmd/upgrades -count=1
rtk go test ./cmd/wowsimcli/cmd -count=1
rtk go build -o wowsimcli ./cmd/wowsimcli
cd ui-finder
rtk npm run build
rtk node --test src/lib/talents.test.js
rtk npm run test:e2e
```

Expected: all Go tests, build, node tests, and the Chromium e2e pass.

- [ ] **Step 4: Manual smoke test**

Run `rtk go run ./cmd/wowsimcli rank-upgrades --addr 127.0.0.1:43199 --no-browser` and open the printed URL:

1. Import `cmd/wowsimcli/cmd/upgrades/testdata/retribution_no_settings_link.txt`.
2. Gear tab: header shows Player · Retribution Paladin, 17 slot cards across two columns plus a three-card weapons strip, ilvl badges and rarity-colored names visible, enchant lines show effect text, sockets render under icons.
3. Stats tab: **raid buffed (link settings)** label present.
4. Talents tab: three Paladin trees render with allocated rank pips.
5. Narrow the window below 640 px: gear collapses to one readable column, stage compacts between columns.
6. Complete a minimal ranking run (phase/screening/confirmation `1`) to confirm the flow is unchanged.

- [ ] **Step 5: Commit**

```bash
git add docs/upgrade-finder.md docs/STATE.md cmd/wowsimcli/cmd/upgrade_ui
git commit -m "docs: armory redesign operator notes and final bundle"
```

(If the smoke test surfaced no changes, `cmd/wowsimcli/cmd/upgrade_ui` is already committed by Tasks 3–4; `git add` on unchanged paths is a no-op.)
