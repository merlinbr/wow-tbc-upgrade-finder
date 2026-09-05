# Svelte Armory Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the embedded vanilla-JS review page with a committed Svelte armory that displays validated equipped gear and deterministic unbuffed base-plus-gear stats before the existing upgrade-ranking flow.

**Architecture:** Keep Go authoritative for imported settings, database resolution, stat calculation, and HTTP contracts. The browser is a Svelte 5 renderer with one module-scoped rune state object and continues to call the existing loopback API. Build the separate `ui-finder` Vite project into the embedded `cmd/wowsimcli/cmd/upgrade_ui/` directory; the committed result lets Go builds and Go tests remain Node-free.

**Tech Stack:** Go, existing wowsims TBC engine/protobuf database, `net/http`, `embed`, Svelte 5 runes, Vite, `@sveltejs/vite-plugin-svelte`, Playwright.

## Global Constraints

- Use Svelte 5 runes. Shared rune state belongs in `ui-finder/src/lib/stores.svelte.js`, not a plain `.js` module.
- Commit the generated Vite bundle under `cmd/wowsimcli/cmd/upgrade_ui/`; `rtk go build` and Go tests must not invoke Node.
- Preserve the existing loopback-only server, API routes, ranking logic, job lifecycle, polling interval (500 ms), cancellation semantics, report fields, and visible wowsims attribution.
- `POST /api/import` adds only `gear`, `stats`, and `derivedStats` to successful responses.
- Validate every nonzero imported item, random-suffix, gem, and enchant-effect ID before enrichment or ranking. Return typed `incompatible_*` validation errors; never pass an unsupported record to engine equipment construction.
- The armory snapshot uses `core.ComputeStats` on a clone with buffs, consumes, and talents cleared. It runs no simulation iterations and is labeled **unbuffed (base + gear)**.
- `Gear` always has the 17 canonical equipment slots. Empty slots and zero gem IDs are ordinary displayed empties.
- Use `core.ColorIntersects` for socket-bonus matching. Preserve hybrid, prismatic, and meta behavior.
- Raw stat keys are snake_case `stats.StatName()` values with no prefix. Derived percentage keys are exactly `melee_hit_percent`, `spell_hit_percent`, `melee_crit_percent`, `spell_crit_percent`, `ranged_hit_percent`, `ranged_crit_percent`, and `block_percent`.
- Add no Go runtime dependency. New frontend dependencies are build/test-only and locked in `ui-finder/package-lock.json`.

---

## Locked File Structure

| Path | Responsibility |
| --- | --- |
| `cmd/wowsimcli/cmd/upgrades/import.go` | Import-time validation of every simulator-consumed equipment identifier. |
| `cmd/wowsimcli/cmd/upgrades/catalog.go` | Immutable database indexes, including random suffixes. |
| `cmd/wowsimcli/cmd/upgrades/armory.go` | Canonical slot enrichment, socket/suffix metadata, and sanitized engine stats snapshot. |
| `cmd/wowsimcli/cmd/upgrades/types.go` | JSON-ready armory domain types and typed validation errors. |
| `cmd/wowsimcli/cmd/upgrades/import_test.go` | Unsupported-item/suffix/gem/enchant import contracts. |
| `cmd/wowsimcli/cmd/upgrades/armory_test.go` | Armory shape, socket, suffix, and engine-snapshot contracts. |
| `cmd/wowsimcli/cmd/upgrade_server.go` | Add armory data to successful imports and safely serve nested Vite assets. |
| `cmd/wowsimcli/cmd/upgrade_server_test.go` | HTTP import and committed-asset route contracts. |
| `ui-finder/package.json` / `package-lock.json` | Isolated reproducible Svelte/Vite/Playwright toolchain. |
| `ui-finder/vite.config.mts` | Relative-base Vite build to the Go embed directory. |
| `ui-finder/src/lib/api.js` | One JSON HTTP client that returns typed server errors. |
| `ui-finder/src/lib/stores.svelte.js` | Shared Svelte 5 state and import/job actions. |
| `ui-finder/src/lib/*.svelte` | Focused import, armory, gear, stats, ranking, and report components. |
| `ui-finder/src/App.svelte` / `src/main.js` / `src/app.css` | Application composition, bootstrap, and responsive armory styling. |
| `ui-finder/e2e/armory.spec.js` / `playwright.config.js` | Browser-level armory and full-flow smoke coverage. |
| `cmd/wowsimcli/cmd/upgrade_ui/` | Generated, committed Vite output only; never hand-edit it. |
| `docs/upgrade-finder.md` | Rebuild command, validation boundaries, armory semantics, and browser smoke procedure. |

### Task 1: Validate every imported equipment identifier

**Files:**
- Modify: `cmd/wowsimcli/cmd/upgrades/catalog.go`
- Modify: `cmd/wowsimcli/cmd/upgrades/import.go`
- Modify: `cmd/wowsimcli/cmd/upgrades/import_test.go`

**Interfaces:**
- Consumes: `database.Load()`, `proto.UIDatabase`, and `proto.ItemSpec{Id, RandomSuffix, Enchant, Gems}`.
- Produces: `Catalog.RandomSuffixes map[int32]*proto.ItemRandomSuffix` and `Import(link) (*ImportedSettings, error)` that rejects every unsupported nonzero equipment identifier before any rank request exists.
- Error contract: `incompatible_item`, `incompatible_random_suffix`, `incompatible_gem`, and `incompatible_enchant` are `ValidationError` codes with the offending numeric ID in `Message`.

- [x] **Step 1: Add failing import-boundary tests**

Extend `import_test.go` with a table that clones `fixtureIndividualSettings()`, mutates one field, encodes the settings with `encodeIndividualLink`, and asserts the exact error code:

```go
func TestImportRejectsUnsupportedEquipmentReferences(t *testing.T) {
    cases := []struct {
        name string
        mutate func(*proto.ItemSpec)
        code string
    }{
        {"item", func(i *proto.ItemSpec) { i.Id = 999999 }, "incompatible_item"},
        {"random suffix", func(i *proto.ItemSpec) { i.RandomSuffix = 999999 }, "incompatible_random_suffix"},
        {"gem", func(i *proto.ItemSpec) { i.Gems = []int32{999999} }, "incompatible_gem"},
        {"enchant", func(i *proto.ItemSpec) { i.Enchant = 999999 }, "incompatible_enchant"},
    }
    // For each case, assert errors.As(err, &ValidationError{}) and err.Code == code.
}
```

Use an existing equipped fixture slot for suffix/gem/enchant mutations so the test exercises the same data path as a normal imported link.

- [x] **Step 2: Run the new test and observe the missing validation**

Run: `rtk go test ./cmd/wowsimcli/cmd/upgrades -run TestImportRejectsUnsupportedEquipmentReferences -count=1`

Expected: FAIL because `Import` currently validates only `ItemSpec.Id`.

- [x] **Step 3: Index suffixes and validate the complete equipment reference**

Add `RandomSuffixes` to `Catalog` and populate it from `db.GetRandomSuffixes()`. In `import.go`, build four ID sets from one `database.Load()` call: items, random suffixes, gems, and enchant effects. For every non-nil `ItemSpec` with `Id > 0`, validate:

```go
if spec.GetRandomSuffix() != 0 && !knownSuffixes[spec.GetRandomSuffix()] {
    return nil, &ValidationError{Code: "incompatible_random_suffix", Message: fmt.Sprintf("equipped random suffix ID %d not found in item database", spec.GetRandomSuffix())}
}
if spec.GetEnchant() != 0 && !knownEnchants[spec.GetEnchant()] {
    return nil, &ValidationError{Code: "incompatible_enchant", Message: fmt.Sprintf("equipped enchant effect ID %d not found in item database", spec.GetEnchant())}
}
for _, gemID := range spec.GetGems() {
    if gemID != 0 && !knownGems[gemID] {
        return nil, &ValidationError{Code: "incompatible_gem", Message: fmt.Sprintf("equipped gem ID %d not found in item database", gemID)}
    }
}
```

Do not validate a suffix, enchant, or gem on an empty `ItemSpec`; it is ignored with the empty slot. Keep the existing item-ID validation and settings digest behavior unchanged.

- [x] **Step 4: Run the import contract suite**

Run: `rtk go test ./cmd/wowsimcli/cmd/upgrades -run 'TestImport|TestNewRequest' -count=1`

Expected: PASS. Existing fixed imports decode unchanged, all four invalid references produce typed errors, and `NewRequest` still leaves protobuf bytes untouched.

- [x] **Step 5: Commit the safe import boundary**

```bash
git add cmd/wowsimcli/cmd/upgrades/catalog.go cmd/wowsimcli/cmd/upgrades/import.go cmd/wowsimcli/cmd/upgrades/import_test.go
git commit -m "fix: validate imported equipment references"
```

### Task 2: Add deterministic armory enrichment and stats

**Files:**
- Create: `cmd/wowsimcli/cmd/upgrades/armory.go`
- Create: `cmd/wowsimcli/cmd/upgrades/armory_test.go`
- Modify: `cmd/wowsimcli/cmd/upgrades/types.go`

**Interfaces:**
- Consumes: validated `*ImportedSettings`, immutable `*Catalog`, `core.ColorIntersects`, and `core.ComputeStats`.
- Produces: `func EnrichArmory(imported *ImportedSettings, catalog *Catalog) (*ArmoryData, error)`.
- `ArmoryData` is `Gear []GearSlotData`, `Stats map[string]float64`, and `DerivedStats map[string]float64`; all fields carry the JSON names from the specification.

- [x] **Step 1: Define the JSON domain types and canonical-slot constants**

Add these public types to `types.go`; use concrete child types rather than `map[string]any`:

```go
type ArmoryData struct {
    Gear         []GearSlotData    `json:"gear"`
    Stats        map[string]float64 `json:"stats"`
    DerivedStats map[string]float64 `json:"derivedStats"`
}
type GearSlotData struct {
    Slot        proto.ItemSlot       `json:"slot"`
    SlotName    string               `json:"slotName"`
    ItemID      int32                `json:"itemId"`
    ItemName    string               `json:"itemName"`
    Quality     proto.ItemQuality    `json:"quality"`
    Icon        string               `json:"icon"`
    Phase       int32                `json:"phase"`
    SetName     string               `json:"setName"`
    Stats       map[string]float64   `json:"stats"`
    RandomSuffix *RandomSuffixData   `json:"randomSuffix"`
    Sockets     []SocketData         `json:"sockets"`
    SocketBonus SocketBonusData       `json:"socketBonus"`
    Enchant     *EnchantData         `json:"enchant"`
}
```

Define `canonicalGearSlots` in `armory.go` in the required left-then-right display order and map each `proto.ItemSlot` to its imported equipment index. Always append all 17 entries, using zero-value item fields for empty equipment.

- [x] **Step 2: Write focused armory contracts before implementation**

Create tests that call `EnrichArmory(mustImportFixture(t), NewCatalog(database.Load()))` and assert these observable facts:

```go
func TestEnrichArmoryReturnsCanonicalSlotsAndResolvedMetadata(t *testing.T) {
    armory, err := EnrichArmory(mustImportFixture(t), NewCatalog(database.Load()))
    if err != nil { t.Fatal(err) }
    if len(armory.Gear) != 17 { t.Fatalf("gear slots = %d, want 17", len(armory.Gear)) }
    if armory.Gear[0].Slot != proto.ItemSlot_ItemSlotHead || armory.Gear[8].Slot != proto.ItemSlot_ItemSlotHands {
        t.Fatalf("unexpected display order: %#v", armory.Gear)
    }
}
```

Add one test fixture item with a purple or prismatic gem in a compatible red/blue/yellow socket and assert `SocketBonus.Active == core.ColorIntersects(socketColor, gemColor)`. Add one suffix fixture and assert its returned contribution equals the engine's `RandomSuffix.Stats.Multiply(float64(item.RandPropPoints)/10000).Floor()` result. Keep these fixture mutations on clones so `ImportedSettings.Settings` remains byte-identical.

- [x] **Step 3: Make the stat snapshot match the engine instead of duplicating formulas**

Write an unexported `armoryComputeRequest(imported *ImportedSettings) *proto.ComputeStatsRequest` that deep-clones the imported player, clears its talent configuration and consumes, creates one party/raid with empty raid and party buffs, clears individual buffs/debuffs, and preserves race, class, professions, equipment, and encounter. Call:

```go
result := core.ComputeStats(armoryComputeRequest(imported))
gearStats := result.GetRaidStats().GetParties()[0].GetPlayers()[0].GetGearStats()
```

Convert `gearStats.Stats` into nonzero raw stat entries with `stats.FromProtoArray` and `stats.StatName()` snake-cased. Convert only the seven specified pseudo-stat values into `DerivedStats`, preserving engine percentage values. Return an error when `ComputeStats` reports an error result or lacks exactly one player gear snapshot; never substitute an approximate value.

- [x] **Step 4: Enrich items, gems, suffixes, enchants, and socket bonuses**

For each canonical `ItemSpec`, resolve the `UIItem`, `UIGem`, `UIEnchant`, and `ItemRandomSuffix` from `Catalog`. Emit direct raw stat maps with the same stat-key conversion as totals. For a selected suffix, scale its raw stats exactly as `core.ItemEquipmentBaseStats` does. Build one `SocketData` per declared socket; a zero gem ID yields `Gem: nil`. Compute `SocketBonus.Active` only when every declared socket has a nonzero gem and `core.ColorIntersects(socketColor, gemColor)` is true. Set `Enchant: nil` only for a zero enchant ID; Task 1 guarantees all nonzero identifiers resolve.

- [x] **Step 5: Run the armory contracts**

Run: `rtk go test ./cmd/wowsimcli/cmd/upgrades -run TestEnrichArmory -count=1`

Expected: PASS. The armory has 17 ordered slots, hybrids/prismatic gems activate correctly, suffix values are scaled, and totals equal the sanitized engine gear-stage snapshot.

- [x] **Step 6: Commit the armory domain module**

```bash
git add cmd/wowsimcli/cmd/upgrades/armory.go cmd/wowsimcli/cmd/upgrades/armory_test.go cmd/wowsimcli/cmd/upgrades/types.go
git commit -m "feat: enrich imported settings for armory review"
```

### Task 3: Extend the import response and serve nested build assets

**Files:**
- Modify: `cmd/wowsimcli/cmd/upgrade_server.go`
- Modify: `cmd/wowsimcli/cmd/upgrade_server_test.go`

**Interfaces:**
- Consumes: `upgrades.EnrichArmory`, `upgrades.NewCatalog(database.Load())`, and committed files under `upgrade_ui/assets/`.
- Produces: successful `POST /api/import` response fields `gear`, `stats`, and `derivedStats`; `GET /assets/{path...}` safely maps to embedded nested asset paths.

- [x] **Step 1: Add failing HTTP and asset tests**

Extend `TestImportReturnsSummaryWithoutStartingJob` to assert `gear` is an array of length 17 and both `stats` and `derivedStats` are JSON objects. Replace direct requests for `/assets/app.js` and `/assets/app.css` with a helper that extracts local `src` and stylesheet `href` values from the committed `GET /` HTML, requests each extracted `/assets/...` URL, and asserts a successful JavaScript or CSS content type. Also request `/assets/not/a/real-file.js` and assert the existing structured 404 body.

- [x] **Step 2: Run the focused server test and observe the missing fields/path support**

Run: `rtk go test ./cmd/wowsimcli/cmd -run 'TestImportReturnsSummaryWithoutStartingJob|TestUpgradePageAndAssetsAreServed' -count=1`

Expected: FAIL: the import response has no armory fields and the current one-segment route cannot serve Vite's nested asset paths.

- [x] **Step 3: Wire one shared immutable catalog into the server**

Add `catalog *upgrades.Catalog` to `upgradeServer`. Initialize it once from `database.Load()` in `newUpgradeServer`; preserve a test-injectable catalog when a test constructs the server. In `handleImport`, call `upgrades.Import`, then `upgrades.EnrichArmory(imported, s.catalog)`. If enrichment returns an unexpected error, use the existing `internal_error` response; validation errors must already have returned before this point. Add `Gear`, `Stats`, and `DerivedStats` fields with JSON names to `importResponse` and populate them without changing existing response values.

- [x] **Step 4: Replace the one-segment asset route with a validated nested path**

Register `GET /assets/{path...}`. In the handler, obtain the wildcard, require `fs.ValidPath(path)`, and call `serveAsset(w, "assets/"+path)`. Reject an empty or invalid path with the existing structured `not_found` response. Keep `GET /` serving `index.html`, no directory listing, `Cache-Control: no-store`, and extension-specific JavaScript/CSS/HTML content types. Do not allow an arbitrary embedded-file name outside `upgrade_ui/assets/`.

- [x] **Step 5: Run the server contracts**

Run: `rtk go test ./cmd/wowsimcli/cmd -run 'TestImportReturnsSummaryWithoutStartingJob|TestUpgradePageAndAssetsAreServed|TestUnknownPathAndMethodAreStructuredErrors' -count=1`

Expected: PASS using the currently committed UI. The asset test stays valid after each hashed rebuild because it follows generated URLs rather than naming files.

- [x] **Step 6: Commit the HTTP boundary**

```bash
git add cmd/wowsimcli/cmd/upgrade_server.go cmd/wowsimcli/cmd/upgrade_server_test.go
git commit -m "feat: return armory data from imports"
```

### Task 4: Create the reproducible Svelte build pipeline

**Files:**
- Create: `ui-finder/package.json`
- Create: `ui-finder/package-lock.json`
- Create: `ui-finder/vite.config.mts`
- Create: `ui-finder/index.html`
- Create: `ui-finder/src/main.js`
- Create: `ui-finder/src/App.svelte`
- Create: `ui-finder/src/app.css`
- Create: `ui-finder/src/lib/api.js`
- Create: `ui-finder/src/lib/stores.svelte.js`
- Modify/delete generated: `cmd/wowsimcli/cmd/upgrade_ui/index.html`, `cmd/wowsimcli/cmd/upgrade_ui/app.js`, `cmd/wowsimcli/cmd/upgrade_ui/app.css`
- Create generated: `cmd/wowsimcli/cmd/upgrade_ui/assets/*`

**Interfaces:**
- Consumes: `/api/import`, `/api/jobs`, and `/api/jobs/{id}` JSON contracts.
- Produces: `npm run build` from `ui-finder` replaces the embedded directory with Vite's `index.html` and hashed `assets/` files.

- [x] **Step 1: Add the isolated package manifest**

Create `ui-finder/package.json` with `private: true`, `type: "module"`, scripts `build: "vite build"` and `test:e2e: "playwright test"`, runtime dependency `svelte`, and development dependencies `vite`, `@sveltejs/vite-plugin-svelte`, and `@playwright/test`. Run `rtk npm install` from `ui-finder` once and commit the resulting lockfile. Do not modify the repository-root package manifest or lockfile.

- [x] **Step 2: Configure Vite for Go embedding**

Use ESM path resolution so the output cannot accidentally land under `ui-finder`:

```ts
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

const outDir = fileURLToPath(new URL('../cmd/wowsimcli/cmd/upgrade_ui/', import.meta.url));
export default defineConfig({
  base: './',
  plugins: [svelte()],
  build: { outDir, emptyOutDir: true, assetsDir: 'assets' },
});
```

Use a minimal `index.html` with `<div id="app"></div>` and a module reference to `/src/main.js`; `main.js` imports `app.css` and mounts `App` into that node.

- [x] **Step 3: Add the shared state and API seam**

In `api.js`, expose `postJSON`, `getJSON`, and `deleteJSON`; each must parse the server's `{error:{code,message}}` body and throw `{code, message}` on a non-2xx response. In `stores.svelte.js`, export one `$state` object with exactly `imported`, `job`, `report`, `error`, `pollTimer`, and `copyStatus`; export action functions `importLink(link)`, `startRanking(link, input)`, `cancelRanking()`, and `copyReport()`. `cancelRanking()` must clear report state immediately, clear the pending timer, issue `DELETE /api/jobs/{id}`, and restore the ranking control state exactly as the vanilla client does.

- [x] **Step 4: Build the shell and prove the generated output is embeddable**

Make `App.svelte` render the product heading, one live alert, a conditional import section, and the unchanged attribution link. Run from `ui-finder`:

```bash
rtk npm run build
```

Then run:

```bash
rtk go test ./cmd/wowsimcli/cmd -run TestUpgradePageAndAssetsAreServed -count=1
```

Expected: both commands PASS. `cmd/wowsimcli/cmd/upgrade_ui/` now contains Vite's generated `index.html` and nested hashed assets; no hand-authored legacy asset remains.

- [x] **Step 5: Commit the build pipeline and generated baseline**

```bash
git add ui-finder cmd/wowsimcli/cmd/upgrade_ui
git commit -m "feat: add embedded Svelte armory build"
```

### Task 5: Port the complete browser flow into focused Svelte components

**Files:**
- Create: `ui-finder/src/lib/ImportPanel.svelte`
- Create: `ui-finder/src/lib/ArmoryView.svelte`
- Create: `ui-finder/src/lib/GearSlot.svelte`
- Create: `ui-finder/src/lib/StatPanels.svelte`
- Create: `ui-finder/src/lib/RankingPanel.svelte`
- Create: `ui-finder/src/lib/ReportView.svelte`
- Modify: `ui-finder/src/App.svelte`
- Modify: `ui-finder/src/app.css`
- Modify generated: `cmd/wowsimcli/cmd/upgrade_ui/*`

**Interfaces:**
- Consumes: the shared state/actions from `stores.svelte.js`, plus import response `character`, `gear`, `stats`, `derivedStats`, and `defaults`.
- Produces: an accessible armory review before ranking and the unchanged observable ranking/report/copy/cancel behavior.

- [x] **Step 1: Implement the import panel and flow ownership in `App.svelte`**

`ImportPanel` owns only the link input and submit button; it calls `importLink(link)` and receives the shared error as a prop. `App.svelte` imports the shared state and uses conditional blocks: import always visible; `ArmoryView` and `RankingPanel` only when `state.imported`; `ReportView` only when `state.report`. Render errors in one `role="alert" aria-live="assertive"` element and progress in `role="status" aria-live="polite"`. Do not reintroduce DOM ID lookups or imperative `innerHTML` rendering.

- [x] **Step 2: Render the armory from server-owned data only**

`ArmoryView` renders the character header (name, level 70 race/class, professions, phase, and abbreviated digest), eight left-column slots and nine right-column slots from the supplied ordered `gear`, then `StatPanels`. `GearSlot` receives one `GearSlotData`; it renders the item icon/name/quality, `No Enchant` for `null` enchant, declared sockets with empty-square treatment for `null` gems, and an active/inactive socket-bonus line. Icon `<img>` uses the ZAM image URL and an `onerror` callback that replaces it with a quality-colored square whose accessible label includes the slot name. It must not fetch item data or calculate colors/stats in the browser.

- [x] **Step 3: Render consistent raw and derived stats**

`StatPanels` accepts `stats` and `derivedStats`, filters zero values, sorts by display label, and renders raw values separately from percentage values. Maintain one local label map keyed by the exact API keys; unknown keys render their snake_case key rather than disappearing. Format raw fractional values with at most two decimal places and derived values as `${value.toFixed(2)}%`. Do not invent `spell_power`, generic `hit_rating`, or a client-side stat conversion.

- [x] **Step 4: Preserve ranking, polling, report, copy, and cancellation semantics**

`RankingPanel` exposes the existing max-phase, unknown-source, screening, and confirmation controls. It validates `screening > 0` and `confirmation >= screening` before calling `startRanking`, passes the original pasted link, and disables Start while a job is queued/running. The store polls `GET /api/jobs/{id}` every 500 ms, updates stage/completed/total progress, retains a completed report, and clears report state for canceled or failed jobs. `ReportView` ports every existing report field and the copy button; `copyReport()` writes the same JSON payload and announces success/failure through `copyStatus`.

- [x] **Step 5: Apply responsive armory styling and rebuild**

In `app.css`, use a CSS grid for the two gear columns and a deliberately empty center track; collapse to one column without changing canonical order on narrow viewports. Apply WoW-quality colors, visible keyboard focus, legible empty socket placeholders, stat-panel grouping, and the persistent attribution footer. Build with `rtk npm run build` from `ui-finder`, then run `rtk go test ./cmd/wowsimcli/cmd -run TestUpgradePageAndAssetsAreServed -count=1`.

Expected: the generated UI passes the server asset test and contains no manual browser implementation.

- [x] **Step 6: Commit the Svelte flow**

```bash
git add ui-finder/src cmd/wowsimcli/cmd/upgrade_ui
git commit -m "feat: render armory review in Svelte"
```

### Task 6: Add end-to-end proof and update operator documentation

**Files:**
- Create: `ui-finder/playwright.config.js`
- Create: `ui-finder/e2e/armory.spec.js`
- Modify: `docs/upgrade-finder.md`
- Modify generated: `cmd/wowsimcli/cmd/upgrade_ui/*`

**Interfaces:**
- Consumes: the generated Vite bundle, the real `rank-upgrades --no-browser` command, and `cmd/wowsimcli/cmd/upgrades/testdata/fixed_individual_link.txt`.
- Produces: a reproducible browser smoke test and operational documentation for rebuild, import validation, armory semantics, and the complete ranking workflow.

- [x] **Step 1: Configure an isolated real-server Playwright run**

Set `playwright.config.js` to use one fixed loopback port, `http://127.0.0.1:43123`, and a `webServer` command from the repository root:

```js
webServer: {
  command: 'rtk go run ./cmd/wowsimcli rank-upgrades --addr 127.0.0.1:43123 --no-browser',
  url: 'http://127.0.0.1:43123/',
  reuseExistingServer: false,
  cwd: fileURLToPath(new URL('..', import.meta.url)),
}
```

Set Chromium as the sole project. Do not mock `/api/import` or `/api/jobs`; the smoke test must exercise the actual embedded UI and local Go server.

- [x] **Step 2: Write the browser smoke before finalizing UI behavior**

In `e2e/armory.spec.js`, read the fixed link fixture, open `/`, import it, and assert: the header appears; all 17 `[data-slot]` elements render; at least one socket and enchant line render; raw and derived stat panels render; and the visible attribution link points to wowsims. Set max phase to `1`, screening and confirmation iterations to `1`, start ranking, observe a progress status before terminal completion, then assert a report table and successful copy-status message. Start one additional job and click Cancel; assert the canceled status and absence of a report. Use role/label selectors, not generated CSS classes.

- [x] **Step 3: Run the focused browser smoke**

Run from `ui-finder`:

```bash
rtk npm run test:e2e
```

Expected: PASS against the real local process. If the bounded real ranking takes longer than the Playwright default, raise only that test's timeout; do not add a test-only server mode or mock ranking responses.

- [x] **Step 4: Update operational documentation**

Add a **Rebuild UI** section to `docs/upgrade-finder.md`:

```bash
cd ui-finder
rtk npm ci
rtk npm run build
```

Update Input to list unsupported item, random-suffix, gem, and enchant-effect IDs as typed pre-ranking rejections. Replace the decoded-summary workflow step with armory review: complete canonical gear, gems, enchants, socket-bonus status, and deterministic **unbuffed (base + gear)** stat panels. State that it excludes buffs, consumes, and talents but follows the engine for racial effects, stat dependencies, suffixes, socket rules, set bonuses, and static gear effects. Extend the manual smoke procedure with the armory assertions and retain ranking, copy, and cancel checks.

- [x] **Step 5: Run final contract and build verification**

Run:

```bash
rtk go test ./cmd/wowsimcli/cmd/... -count=1
```

Then from `ui-finder` run:

```bash
rtk npm ci
rtk npm run build
rtk npm run test:e2e
```

Expected: Go contracts pass without Node, the committed bundle rebuilds deterministically, and the real browser surface completes import, armory review, ranking, copying, and cancellation.

- [x] **Step 6: Commit verification and documentation**

```bash
git add docs/upgrade-finder.md ui-finder cmd/wowsimcli/cmd/upgrade_ui
git commit -m "test: verify Svelte armory workflow"
```

## Plan Self-Review

- **Spec coverage:** Task 1 prevents the unknown-ID crash path. Task 2 covers all armory domain data, canonical slots, suffixes, socket matching, engine-backed raw/derived stats, and the unbuffed boundary. Task 3 exposes the additive import fields and securely serves Vite output. Tasks 4–5 create, embed, and render the Svelte application while retaining existing ranking/report behavior. Task 6 supplies the requested route, contract, Playwright, and documentation verification.
- **Intentional exclusions:** No 3D model, talent UI, item tooltips, buffed-stat mode, price data, optimizer change, ranking rewrite, account/persistence feature, remote server, or Go runtime dependency is introduced.
- **Consistency:** `ImportedSettings`, `Catalog`, `ArmoryData`, `GearSlotData`, `EnrichArmory`, and the three additive import response fields are the only new cross-task contracts. Frontend state uses the same `imported`, `job`, `report`, and `error` concepts throughout.
- **Placeholder scan:** Every implementation and verification step names the files, symbols, behavior, failure mode, and command required to finish it. Unsupported simulator input is rejected rather than rendered as a misleading placeholder; only ordinary empty slots and zero gem IDs render empty.

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-08-29-svelte-armory.md`. Two execution options:

1. **Subagent-Driven (recommended)** — dispatch a fresh subagent per task and review after each task.
2. **Inline Execution** — execute tasks in this session with checkpoints for review.
