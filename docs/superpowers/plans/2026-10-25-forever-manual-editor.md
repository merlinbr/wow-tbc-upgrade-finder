# Forever Manual Editor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the Phase 2 Forever armory: a local, single-user planner where a character is created by race/class/spec/level, owns one Current build and any number of named plans, and every equipment and talent edit is validated against the pinned Forever catalog before it is saved.

**Architecture:** A new Go binary (`cmd/forever-armory`) loads a normalized, pinned catalog from embedded assets and serves a JSON API plus an embedded Svelte 5 SPA. Domain and validation packages are independent of the TBC application; a SQLite store persists characters and builds behind a `Store` interface so Phase 3 can add ownership and publication without changing the domain. The generator in `tools/forever-catalog` is the only code that reads upstream WoWSims Forever payloads.

**Tech Stack:** Go 1.25 (`net/http`, `database/sql`, `modernc.org/sqlite`), Svelte 5 + Vite 7, Playwright, Node 22.

**Spec:** `docs/superpowers/specs/2026-10-01-forever-armory-design.md`
**Read with:** `docs/forever-upstream-data-contract.md` (identifier rules, validation semantics, pinned hashes). The contract governs where this plan and the spec disagree.

## Global Constraints

- Module path is `github.com/wowsims/tbc`. Forever packages never import `sim/`, `cmd/wowsimcli/`, or `assets/database`.
- Catalog pin: commit `0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4`; `db.json` sha256 `9fa8335e7e97f824373ad8225b6c2d82e1a2915b0e71a6f05e2312add442e9e3`; `leftover_db.json` sha256 `eeb3ff7788226b9739e470c27d8a9a4035dbae5b5ad5f5d781483eeef0ce4685`.
- `catalog_revision` is exactly `wowsims/forever@0cfb1a0:<sha256(db.json)>:<sha256(leftover_db.json)>`. It is computed from the raw downloaded bytes, never typed by hand.
- The server binds loopback only by default (`127.0.0.1:0`) and prints its URL. Flags: `--addr`, `--no-browser`, `--db`.
- Phase 2 has no authentication, no publication, and no remote storage. `published` is always `false` and no public route exists.
- Level 60 is the only supported equipment-planning level. Lower levels may be stored and get a talent budget, but never an equipment-legality claim.
- Talent budget is `level - 9` for levels 10–60; below 10 it is 0. Talents are stored by identity, never by position.
- Validation errors use typed `snake_case` codes and are returned with HTTP 4xx; a failed save mutates nothing.
- SQLite is the only persistent store in this phase. The default database path is `./forever-armory.db`; tests use an in-memory database.
- Item and talent icons use `https://wow.zamimg.com/images/wow/icons/medium/{icon}.jpg` with a neutral fallback. No game art is committed.
- UI copy is English. Every task ends with a commit.
- The TBC application, its assets, and its tests are not modified by any task.

## Review Focus

The following are the failure modes most likely to reach a user. Each gets a test in the owning task.

1. **Missing or corrupt catalog asset:** the server must fail to start with a clear error, never serve a partial or empty picker.
2. **Item variant or suffix dropped:** an item whose stats live only in `scalingOptions` must show and save the selected variant's stats, never zeros or another variant's values.
3. **Ambiguous or effect-only enchant:** resolution must key on effect + slot, reject ambiguity by name, and never guess an item/spell identity.
4. **Malformed talent input:** wrong length, unknown key, over-cap or over-budget allocations must be rejected, never clamped into a legal-looking build.
5. **Concurrent save:** a stale `version` must return `409 version_conflict`, preserve both edits, and never silently overwrite.

## File Structure

```text
cmd/forever-armory/
  main.go                 flags, startup, graceful shutdown
  webui/                  committed Vite build output (embedded)
internal/forever/
  domain/
    catalog.go            Item, Enchant, RandomSuffix, TalentClass, TalentTree, Capabilities
    character.go          Character, Build, Equipment, Talents, Slot keys
    errors.go             typed domain errors
  catalog/
    catalog.go            load embedded assets, indexes, lookup helpers
    search.go             item/enchant filtering and pagination
  validate/
    equipment.go          class/race/armor/weapon/slot/unique/limit/faction/profession/DW
    enchant.go            effect+slot resolution, type and class rules
    talents.go            budget, tier, prerequisite, broken-talents
  store/
    store.go              Store interface
    sqlite.go             schema, migrations, CRUD, optimistic versioning
  service/
    service.go            character/build workflows and validation orchestration
  server/
    server.go             routes, embedded UI, JSON helpers
    handlers_catalog.go   /api/meta, /api/items, /api/enchants, /api/talents
    handlers_characters.go /api/characters, /api/builds
assets/forever/
  embed.go                //go:embed
  catalog.json.gz         generated
  manifest.json           generated
  talents/<class>.json    generated
  capabilities.json       generated
tools/forever-catalog/
  main.go                 CLI: fetch, project, verify, write
  fetch.go                pinned raw download and hash verification
  enums.go                parse proto/common.proto enums
  project_items.go        items, enchants, suffixes
  project_talents.go      trees, talent identities
  capabilities.go         class/race + spec dual-wield tables
  manifest.go             revision, counts, write
  testdata/               miniature upstream payload
ui-forever/
  package.json, vite.config.js, index.html, playwright.config.js
  src/main.js, src/App.svelte, src/app.css
  src/lib/api.js, router.svelte.js, stores.svelte.js
  src/lib/CharacterLibrary.svelte, CreateCharacterDialog.svelte
  src/lib/Editor.svelte, IdentityHeader.svelte, GearGrid.svelte, SlotCard.svelte
  src/lib/StatsPanel.svelte, ItemPicker.svelte, EnchantPicker.svelte
  src/lib/TalentEditor.svelte, TalentTree.svelte, TalentNode.svelte
  src/lib/SaveBar.svelte, ViolationList.svelte
  e2e/*.spec.js
docs/superpowers/plans/2026-10-25-forever-manual-editor.md
```

---

## Part 1 — Catalog and domain

### Task 1: Pinned catalog generator (enums, items, enchants, suffixes)

**Files:**
- Create: `internal/forever/domain/catalog.go`
- Create: `internal/forever/domain/errors.go`
- Create: `tools/forever-catalog/main.go`
- Create: `tools/forever-catalog/fetch.go`
- Create: `tools/forever-catalog/enums.go`
- Create: `tools/forever-catalog/project_items.go`
- Create: `tools/forever-catalog/testdata/proto.txt`, `testdata/db.json`, `testdata/leftover_db.json`
- Test: `tools/forever-catalog/project_items_test.go`, `tools/forever-catalog/enums_test.go`

**Interfaces:**
- Produces `domain.Item`, `domain.Stat`, `domain.Variant`, `domain.Enchant`, `domain.RandomSuffix`, `domain.Slot`, `domain.SlotOrder`.
- Produces `FetchPinned(ctx context.Context, baseURL string) (RawPayload, error)` and `ParseEnums(protoText string) (Enums, error)` and `ProjectItems(RawPayload, Enums) (domain.Catalog, error)`.
- `domain.Catalog` fields: `Revision string`, `Items []domain.Item`, `Enchants []domain.Enchant`, `Suffixes []domain.RandomSuffix`, `LeftoverItems []domain.Item`, `LeftoverEnchants []domain.Enchant`.
- `domain.Item.Slots` is derived from `type` + `handType`/`weaponType`: head/neck/shoulder/back/chest/wrist/hands/waist/legs/feet each map to one slot; finger maps to `finger1`,`finger2`; trinket maps to `trinket1`,`trinket2`; ranged to `ranged`; weapon maps to `main_hand`,`off_hand` for one-hand, `off_hand` for shield/off-hand, `main_hand` for two-hand/main-hand.

- [ ] **Step 1: Write the failing enum test**

```go
func TestParseEnums(t *testing.T) {
	enums, err := ParseEnums(string(mustRead(t, "testdata/proto.txt")))
	if err != nil { t.Fatal(err) }
	if got := enums.StatName(27); got != "armor" { t.Fatalf("stat 27 = %q", got) }
	if got := enums.SlotName(0); got != "head" { t.Fatalf("slot 0 = %q", got) }
	if got := enums.ItemTypeName(7); got != "hands" { t.Fatalf("type 7 = %q", got) }
	if got := enums.WeaponName(9); got != "sword" { t.Fatalf("weapon 9 = %q", got) }
}
```

- [ ] **Step 2: Run it and confirm it fails**

Run: `go test ./tools/forever-catalog -run TestParseEnums -v`
Expected: FAIL — `undefined: ParseEnums`.

- [ ] **Step 3: Implement the enum parser and domain catalog types**

Parse the pinned `proto/common.proto` with one regex per enum (`Stat`, `PseudoStat`, `ItemType`, `ArmorType`, `WeaponType`, `HandType`, `RangedWeaponType`, `Class`, `Race`, `ItemSlot`, `ItemQuality`, `Profession`), strip the enum-name prefix from each member, and lower-case to the key form (`ItemSlotMainHand` → `main_hand`, `StatArmor` → `armor`). Return an `Enums` value with `Name(value int) string` helpers per enum. Define the domain structs in `internal/forever/domain/catalog.go` with the JSON tags used by the API.

- [ ] **Step 4: Run the enum test and confirm it passes**

Run: `go test ./tools/forever-catalog -run TestParseEnums -v`
Expected: PASS.

- [ ] **Step 5: Write the failing item-projection test**

Fixture `testdata/db.json` contains two items: one with only `scalingOptions` (stats `{"0":8,"2":7}`, `ilvl` 28, no top-level stats) and one with top-level `stats` and `ilvl`; one enchant with `effectId`, `itemId`, `spellId`, `type`; one random suffix. Assert:

```go
cat, err := ProjectItems(raw, enums)
item := findItem(cat.Items, 720)
if item.Ilvl != 28 { t.Fatalf("ilvl = %d", item.Ilvl) }
if got := statValue(item.Stats, "strength"); got != 8 { t.Fatalf("strength = %v", got) }
if len(item.Slots) != 1 || item.Slots[0] != domain.SlotHands { t.Fatalf("slots = %v", item.Slots) }
if cat.Enchants[0].EffectID != 8219 || cat.Enchants[0].SpellID != 1249058 { t.Fatal("enchant ids") }
if cat.Suffixes[0].ID != 12718 { t.Fatal("suffix") }
```

- [ ] **Step 6: Run it and confirm it fails**

Run: `go test ./tools/forever-catalog -run TestProjectItems -v`
Expected: FAIL — `undefined: ProjectItems`.

- [ ] **Step 7: Implement item/enchant/suffix projection**

Fetch `db.json` and `leftover_db.json` from `https://raw.githubusercontent.com/wowsims/forever/<pin>/assets/database/`, verify both sha256 values, and unmarshal into `RawPayload` with camelCase fields (`scalingOptions`, `randomSuffixOptions`, `classAllowlist`, `requiredProfession`, `factionRestriction`, `setId`, `setName`, `weaponSpeed`, `weaponDamageMin`, `weaponDamageMax`, `stats`, `pseudoStats`, `itemEffects`, `sources`).

Flatten the `scalingOptions["0"]` variant into `Item.Ilvl`, `Item.Stats` and `Item.WeaponDamageMin/Max`; keep any other variant keys in `Item.Variants` with their resolved `Ilvl`. Convert `stats` maps (`{"0":8}`) through `Enums.StatName`. Drop items whose `type` is 0 (shirts/tabbards and other non-statistical rows). Derive `Item.Slots` per the interface block. Keep `LeftoverItems` and `LeftoverEnchants` as separate slices.

- [ ] **Step 8: Run the projection test and confirm it passes**

Run: `go test ./tools/forever-catalog -run 'TestParseEnums|TestProjectItems' -v`
Expected: PASS.

- [ ] **Step 9: Run the generator against the pin and verify counts**

Run: `go run ./tools/forever-catalog -verify`
Expected output includes `items=8599 leftoverItems=1186 enchants=246 leftoverEnchants=20 suffixCount=<n>` and writes `assets/forever/catalog.json.gz`. The command exits non-zero if the raw counts differ from the contract. (Eligible counts are recorded but not asserted; shirts are filtered.)

- [ ] **Step 10: Commit**

```bash
git add internal/forever/domain tools/forever-catalog assets/forever/catalog.json.gz
git commit -m "feat(forever): generate pinned catalog items, enchants and suffixes"
```

### Task 2: Generator talents, capabilities and manifest

**Files:**
- Create: `tools/forever-catalog/project_talents.go`
- Create: `tools/forever-catalog/capabilities.go`
- Create: `tools/forever-catalog/manifest.go`
- Test: `tools/forever-catalog/project_talents_test.go`, `tools/forever-catalog/capabilities_test.go`
- Modify: `tools/forever-catalog/main.go`, `assets/forever/manifest.json`

**Interfaces:**
- Produces `domain.TalentClass` (one class: `Class`, `Trees []domain.TalentTree`), `domain.TalentTree` (`Name`, `Background`, `Talents []domain.Talent`), `domain.Talent`, `domain.Capabilities`, `domain.ClassInfo`, `domain.SpecInfo`, `domain.WeaponEligibility`.
- Produces `ProjectTalents(raw RawTalents, enums Enums) (map[domain.Class]*domain.TalentClass, error)` and `ParseCapabilities(capsTS string, enums Enums, specs map[domain.Class][]domain.SpecInfo) (domain.Capabilities, error)`.
- `Talent.ID` is `def:<definitionId>` when present, otherwise `loc:<class>:<treeIndex>:<row>,<col>`; `Talent.PrereqID` refers to another talent's ID.
- `Capabilities.Specs` is `map[domain.Class][]domain.SpecInfo` with one spec per class tree (`<class>_<tree-slug>`).
- `Capabilities.DualWield` is `map[domain.Spec]bool`, true for every spec of a class present in the pinned `SPEC_CAN_DUAL_WIELD` mapping (`warrior`, `hunter`, `rogue`).

- [ ] **Step 1: Write the failing talent test**

Fixture tree: three talents, one with `definitionId`, one without, one with `prereqLocation`. Assert identity, prerequisite resolution and cap:

```go
trees, err := ProjectTalents(rawTrees, enums)
tal := trees[0].Trees[0].Talents[1]
if tal.ID != "def:135378" { t.Fatalf("id = %q", tal.ID) }
if tal.MaxPoints != 3 { t.Fatalf("max = %d", tal.MaxPoints) }
if trees[0].Trees[0].Talents[2].PrereqID != tal.ID { t.Fatalf("prereq = %q", ...) }
if trees[0].Trees[0].Talents[0].ID == "" { t.Fatal("missing fallback id") }
```

- [ ] **Step 2: Run it and confirm it fails**

Run: `go test ./tools/forever-catalog -run TestProjectTalents -v`
Expected: FAIL.

- [ ] **Step 3: Implement talent projection**

Fetch the nine `ui/sim/talents/trees/<class>.json` files from the pin. Map each tree to a spec key with a slug of the tree name (`Holy` → `holy`, `Feral Combat` → `feral_combat`). Resolve `prereqLocation` to the target talent's ID; fail the run if a prerequisite cannot be resolved. Resolve talent icons from `UIDatabase.spell_icons` by `spellId`; leave `Icon` empty when absent.

- [ ] **Step 4: Run the talent test and confirm it passes**

Run: `go test ./tools/forever-catalog -run TestProjectTalents -v`
Expected: PASS.

- [ ] **Step 5: Write the failing capabilities test**

Using the fixture TS snippet and enum tables, assert Paladin plate/mail armor, Paladin sword eligibility, Mage wand, Warrior/Bow, and that a Paladin spec is not dual-wield while a Warrior spec is:

```go
caps, err := ParseCapabilities(capsTS, enums, specs)
if !slices.Contains(caps.Armor["paladin"], "plate") { t.Fatal("paladin plate") }
if caps.DualWield["paladin_retribution"] { t.Fatal("paladin dw") }
if !caps.DualWield["warrior_fury"] { t.Fatal("warrior dw") }
if !slices.Contains(caps.Races["paladin"], "human") { t.Fatal("paladin human") }
```

- [ ] **Step 6: Run it and confirm it fails**

Run: `go test ./tools/forever-catalog -run TestParseCapabilities -v`
Expected: FAIL.

- [ ] **Step 7: Implement the capabilities parser**

Fetch `ui/sim/player/classes/capabilities_auto_gen.ts`. Parse the four `Record<number, ...>` blocks (`CLASS_ARMOR_TYPES`, `CLASS_WEAPON_TYPES`, `CLASS_RANGED_WEAPON_TYPES`, `CLASS_RACES`) and `SPEC_CAN_DUAL_WIELD` with a regex that matches `[Class.X]: [...]` and `Spec.X` members, translating enum member names through `Enums`. Derive `DualWield` by mapping the upstream spec names (`SpecDpsWarrior`, `SpecProtectionWarrior` → `warrior`; `SpecHunter` → `hunter`; `SpecRogue` → `rogue`) onto every spec of that class. Fail the generator if a member cannot be translated.

- [ ] **Step 8: Run the capabilities test and confirm it passes**

Run: `go test ./tools/forever-catalog -run TestParseCapabilities -v`
Expected: PASS.

- [ ] **Step 9: Write the manifest**

`manifest.json` fields: `catalogRevision`, `sourceCommit`, `sha256` (of both raw payloads), `counts` (`itemsTotal`, `itemsEligible`, `itemsLeftover`, `enchants`, `leftoverEnchants`, `talents`, `classes`, `suffixes`). No timestamp, so regeneration is byte-stable. `-verify` asserts `itemsTotal == 8599`, `itemsLeftover == 1186`, `enchants == 246`, `leftoverEnchants == 20`, `talents == 465`, `classes == 9`.

- [ ] **Step 10: Regenerate all assets and verify**

Run: `go run ./tools/forever-catalog -verify`
Expected: writes `assets/forever/{catalog.json.gz,manifest.json,capabilities.json,talents/*.json}` and prints the counts above.

- [ ] **Step 11: Commit**

```bash
git add tools/forever-catalog assets/forever
git commit -m "feat(forever): generate talents, capabilities and catalog manifest"
```

### Task 3: Server scaffold, embedded assets and `/api/meta`

**Files:**
- Create: `assets/forever/embed.go`
- Create: `internal/forever/catalog/catalog.go`
- Create: `internal/forever/server/server.go`
- Create: `cmd/forever-armory/main.go`
- Test: `internal/forever/server/server_test.go`, `internal/forever/catalog/catalog_test.go`

**Interfaces:**
- `catalog.Load() (*Catalog, error)` loads the embedded assets; `catalog.LoadFrom(fsys fs.FS) (*Catalog, error)` does the same from any filesystem so tests can inject failures.
- `server.New(cat *catalog.Catalog, svc *service.Service) http.Handler` (svc may be nil in this task).
- `GET /api/meta` returns `{"catalogRevision":..., "counts":{...}, "classes":[...], "specs":{...}, "races":[...], "dualWield":{...}}`.

- [ ] **Step 1: Write the failing catalog-load test**

```go
func TestLoadEmbeddedCatalog(t *testing.T) {
	cat, err := Load()
	if err != nil { t.Fatal(err) }
	if cat.Manifest.CatalogRevision == "" { t.Fatal("empty revision") }
	if len(cat.Items) < 1000 { t.Fatalf("items = %d", len(cat.Items)) }
	if cat.Talents["paladin"] == nil || len(cat.Talents["paladin"].Trees) != 3 { t.Fatal("paladin trees") }
}

func TestLoadRejectsCorruptCatalog(t *testing.T) {
	_, err := LoadFrom(fstest.MapFS{"catalog.json.gz": &fstest.MapFile{Data: []byte("not gzip")}})
	if err == nil { t.Fatal("corrupt catalog accepted") }
}
```

- [ ] **Step 2: Run them and confirm they fail**

Run: `go test ./internal/forever/catalog -run TestLoad -v`
Expected: FAIL — `undefined: Load`.

- [ ] **Step 3: Implement the loader and a startup that refuses a bad asset**

`LoadFrom` decompresses `catalog.json.gz`, unmarshals `capabilities.json` and `manifest.json`, and builds a `Catalog` with `Manifest domain.Manifest`, `Items map[int]*domain.Item`, `EnchantsByEffect map[int][]*domain.Enchant`, `Suffixes map[int]*domain.RandomSuffix`, `Talents map[domain.Class]*domain.TalentClass`, and `Capabilities domain.Capabilities`. `Load()` is `LoadFrom(foreverassets.FS)`. Return an error when the gzip stream is invalid, the manifest revision is empty, item count is zero, an enchant/suffix/talent reference cannot be indexed, or any class has zero talent trees. Do not substitute an empty catalog. Add `assets/forever/embed.go`:

```go
package foreverassets

import "embed"

//go:embed catalog.json.gz manifest.json capabilities.json talents/*.json
var FS embed.FS
```

- [ ] **Step 4: Run the load tests and confirm they pass**

Run: `go test ./internal/forever/catalog -run TestLoad -v`
Expected: PASS (both the embedded-catalog and corrupt-catalog cases).

- [ ] **Step 5: Write the failing meta test**

```go
func TestMeta(t *testing.T) {
	srv := newTestServer(t)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest("GET", "/api/meta", nil))
	if rec.Code != 200 { t.Fatalf("status %d", rec.Code) }
	var body struct {
		CatalogRevision string            `json:"catalogRevision"`
		Classes         []string          `json:"classes"`
		Specs           map[string][]string `json:"specs"`
	}
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body.CatalogRevision != "wowsims/forever@0cfb1a0:..." { t.Fatal("revision") }
	if len(body.Classes) != 9 { t.Fatalf("classes = %d", len(body.Classes)) }
}
```

- [ ] **Step 6: Run it and confirm it fails**

Run: `go test ./internal/forever/server -run TestMeta -v`
Expected: FAIL.

- [ ] **Step 7: Implement the server and binary**

`server.New` returns an `http.Handler` with `GET /api/meta`, a `GET /` that serves `cmd/forever-armory/webui/index.html` (the directory exists after Task 10; until then return a placeholder 503 when the bundle is absent so the binary still runs), and a `GET /assets/{path...}` guard that rejects traversal with `fs.ValidPath`. `main.go` parses `--addr` (default `127.0.0.1:0`), `--no-browser`, `--db` (default `./forever-armory.db`), forces loopback, prints the resolved URL, and shuts down on SIGINT.

- [ ] **Step 8: Run the meta test and confirm it passes**

Run: `go test ./internal/forever/server -run TestMeta -v`
Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add cmd/forever-armory internal/forever/catalog internal/forever/server assets/forever/embed.go
git commit -m "feat(forever): add server, catalog loader and meta endpoint"
```

### Task 4: Catalog search API

**Files:**
- Create: `internal/forever/catalog/search.go`
- Create: `internal/forever/server/handlers_catalog.go`
- Test: `internal/forever/catalog/search_test.go`, `internal/forever/server/handlers_catalog_test.go`

**Interfaces:**
- `catalog.SearchItems(cat *Catalog, f ItemFilter) ([]domain.Item, int)`
- `ItemFilter{Class domain.Class, Slot domain.Slot, Query string, Source string, Limit, Offset int}`
- `catalog.Compatibility(ch domain.Character, item *domain.Item, slot domain.Slot) (bool, string)` — class allowlist, armor/weapon/ranged proficiency, slot fit, faction and required profession. Returns a `snake_case` reason when false.
- `catalog.SearchEnchants(cat *Catalog, f EnchantFilter) []domain.Enchant`
- `catalog.TalentTrees(cat *Catalog, class domain.Class) (*domain.TalentClass, error)`
- Routes: `GET /api/items`, `GET /api/items/{id}`, `GET /api/enchants`, `GET /api/talents?class=`.
- Item rows in `/api/items` carry `compatible` and `reason` for the requested class and slot; the detail endpoint carries the same pair.

- [ ] **Step 1: Write the failing search test**

```go
func TestSearchItemsFiltersAndPaginates(t *testing.T) {
	cat := loadTestCatalog(t)
	got, total := SearchItems(cat, ItemFilter{Class: "paladin", Slot: "head", Limit: 1})
	if total < 2 { t.Fatalf("total = %d", total) }
	if len(got) != 1 { t.Fatalf("page = %d", len(got)) }
	for _, it := range got {
		if !slices.Contains(it.Slots, domain.SlotHead) { t.Fatal("slot leak") }
	}
	if _, n := SearchItems(cat, ItemFilter{Query: "brawler gloves"}); n != 1 { t.Fatalf("q total = %d", n) }
	if ok, _ := Compatibility(paladin(), cat.Items[plateHeadID], domain.SlotHead); !ok { t.Fatal("paladin plate head should fit") }
	if ok, reason := Compatibility(mage(), cat.Items[plateHeadID], domain.SlotHead); ok || reason != "incompatible_armor" { t.Fatalf("mage plate: %v %q", ok, reason) }
}
```

- [ ] **Step 2: Run it and confirm it fails**

Run: `go test ./internal/forever/catalog -run TestSearchItems -v`
Expected: FAIL.

- [ ] **Step 3: Implement deterministic search**

Case-insensitive substring match on `Name`; class filter rejects items whose non-empty `ClassAllowlist` excludes the class; slot filter matches `Item.Slots`; source filter matches normalized source kind; sort by `Ilvl` desc, then `ID` asc; apply `Limit`/`Offset` after filtering. Cap `Limit` at 100 (default 50). Return the total before pagination. `Compatibility` implements the static item rules once; the equipment validator reuses it rather than restating them.

- [ ] **Step 4: Run the search test and confirm it passes**

Run: `go test ./internal/forever/catalog -run TestSearchItems -v`
Expected: PASS.

- [ ] **Step 5: Write the failing handler test**

```go
func TestItemDetailUnknownReturns404(t *testing.T) {
	rec := get(t, srv, "/api/items/999999999")
	if rec.Code != 404 { t.Fatalf("status %d", rec.Code) }
	var e struct{ Error struct{ Code string `json:"code"` } `json:"error"` }
	json.Unmarshal(rec.Body.Bytes(), &e)
	if e.Error.Code != "not_found" { t.Fatalf("code %q", e.Error.Code) }
}

func TestTalentsRequiresKnownClass(t *testing.T) {
	if rec := get(t, srv, "/api/talents?class=paladin"); rec.Code != 200 { t.Fatalf("paladin %d", rec.Code) }
	if rec := get(t, srv, "/api/talents?class=bard"); rec.Code != 400 { t.Fatalf("bard %d", rec.Code) }
}
```

- [ ] **Step 6: Run it and confirm it fails**

Run: `go test ./internal/forever/server -run 'TestItemDetail|TestTalents' -v`
Expected: FAIL.

- [ ] **Step 7: Implement the handlers**

Wire the four routes through `server.go`. Decode no bodies. Map `domain.ErrNotFound` to `404 not_found`, bad query params to `400 invalid_query`, and unknown class to `400 unsupported_class`. Item detail includes `variants`, `suffixIds`, `sources` and stats so the picker can render the default without another call.

- [ ] **Step 8: Run the handler tests and confirm they pass**

Run: `go test ./internal/forever/server -run 'TestItemDetail|TestTalents' -v`
Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add internal/forever/catalog/search.go internal/forever/server/handlers_catalog.go internal/forever/server
git commit -m "feat(forever): serve catalog search, item detail, enchants and talents"
```

### Task 5: Equipment validation

**Files:**
- Create: `internal/forever/validate/equipment.go`
- Test: `internal/forever/validate/equipment_test.go`

**Interfaces:**
- `type Violation struct { Code, Severity, Message string; Slot domain.Slot }`
- `func Equipment(ch domain.Character, eq domain.Equipment, cat *catalog.Catalog) []Violation`
- Character input fields: `Class`, `Spec`, `Race`, `Level`, `Professions []string`.

- [ ] **Step 1: Write the failing table test**

```go
func TestEquipmentRules(t *testing.T) {
	cases := []struct {
		name string
		ch   domain.Character
		eq   domain.Equipment
		code string
	}{
		{"plate on mage", mage(), gear(720), "incompatible_armor"},
		{"sword on priest", priest(), gear(swordID), "incompatible_weapon"},
		{"shield in main hand", warrior(), gearIn("main_hand", shieldID), "incompatible_slot"},
		{"second unique copy", warrior(), twoCopies(uniqueID), "unique_conflict"},
		{"two limit-category items", warrior(), twoLimited(limitedIDs), "limit_category_conflict"},
		{"race/class mismatch", undeadPaladin(), gear(0), "race_class_mismatch"},
		{"off-hand one-hander without dw", paladin(), twoOneHands(), "dual_wield_unavailable"},
		{"two-hander clears off hand", warrior(), twoHandPlusOffHand(), "two_hand_conflict"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Equipment(tc.ch, tc.eq, testCatalog(t)); !hasCode(got, tc.code) {
				t.Fatalf("want %s, got %v", tc.code, codes(got))
			}
		})
	}
}
```

- [ ] **Step 2: Run it and confirm it fails**

Run: `go test ./internal/forever/validate -run TestEquipmentRules -v`
Expected: FAIL.

- [ ] **Step 3: Implement the rules in contract order**

For each occupied slot: resolve the item (unknown → `unknown_item`); call `catalog.Compatibility` for class allowlist, class/race availability, armor/weapon/ranged proficiency, slot eligibility, faction and required profession; then add uniqueness across the whole equipment map, `limit_category` counts across the whole map, off-hand one-hand requiring `Capabilities.DualWield[ch.Spec]`, and a two-handed main hand with an occupied off hand (`two_hand_conflict`). All messages name the item and slot. Severity is `error` for every rule in this task.

- [ ] **Step 4: Run the rule tests and confirm they pass**

Run: `go test ./internal/forever/validate -run TestEquipmentRules -v`
Expected: PASS.

- [ ] **Step 5: Add the supplementary check for empty required fields**

Test that an item with class allowlist `[paladin]` on a warrior returns `incompatible_class`, and that an item requiring `blacksmithing` on a character without it returns `incompatible_profession`.

- [ ] **Step 6: Run all validation tests**

Run: `go test ./internal/forever/validate -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/forever/validate/equipment.go internal/forever/validate/equipment_test.go
git commit -m "feat(forever): validate equipment against pinned capabilities"
```

### Task 6: Enchant validation

**Files:**
- Create: `internal/forever/catalog/enchant.go`
- Create: `internal/forever/validate/enchant.go`
- Test: `internal/forever/validate/enchant_test.go`

**Interfaces:**
- `func Enchant(ch domain.Character, slot domain.Slot, item *domain.Item, inst *domain.EnchantInstance, cat *catalog.Catalog) []Violation`
- `catalog.ResolveEnchant(cat, effectID int, slot domain.Slot) (*domain.Enchant, error)` returns `ErrAmbiguous` when more than one row matches effect + slot and the item/spell IDs do not disambiguate.

- [ ] **Step 1: Write the failing test**

```go
func TestEnchantResolution(t *testing.T) {
	if e, err := ResolveEnchant(cat, 8219, "neck"); err != nil || e.SpellID != 1249058 { t.Fatalf("neck: %v %v", e, err) }
	if _, err := ResolveEnchant(cat, 8219, "head"); !errors.Is(err, ErrAmbiguous) && !errors.Is(err, ErrNotFound) { t.Fatalf("head: %v", err) }
	if got := Enchant(paladin(), "head", helm, &domain.EnchantInstance{EffectID: 999999}, cat); !hasCode(got, "unknown_enchant") { t.Fatal("unknown") }
	if got := Enchant(paladin(), "head", helm, weaponOnlyEnchant, cat); !hasCode(got, "incompatible_enchant_slot") { t.Fatal("slot") }
}
```

- [ ] **Step 2: Run it and confirm it fails**

Run: `go test ./internal/forever/validate -run TestEnchant -v`
Expected: FAIL.

- [ ] **Step 3: Implement resolution and validation**

Load `EnchantsByEffect`; filter by the item type of the target slot (`type` plus `extra_types`); if multiple rows remain, try `item_id` then `spell_id`; if still multiple, return `ErrAmbiguous`. Validate class allowlist and required profession; an enchant on an empty slot is `enchant_without_item`.

- [ ] **Step 4: Run the test and confirm it passes**

Run: `go test ./internal/forever/validate -run TestEnchant -v`
Expected: PASS.

- [ ] **Step 5: Add the ambiguity-by-name test**

Pick a fixture effect with two rows for the same slot and distinct `item_id`; assert that supplying the matching `item_id` resolves and supplying neither returns `ErrAmbiguous`. Confirm the API code is `ambiguous_enchant`.

- [ ] **Step 6: Run all validation tests**

Run: `go test ./internal/forever/validate -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/forever/validate/enchant.go internal/forever/validate/enchant_test.go
git commit -m "feat(forever): resolve and validate enchants by effect and slot"
```

### Task 7: Talent validation

**Files:**
- Create: `internal/forever/validate/talents.go`
- Test: `internal/forever/validate/talents_test.go`

**Interfaces:**
- `func Talents(ch domain.Character, alloc domain.Talents, trees *domain.TalentClass) ([]Violation, domain.TalentIssues)`
- `domain.TalentIssues{Broken []string, TotalPoints int}`.
- `func TrySetPoints(trees *domain.TalentClass, alloc domain.Talents, id string, next int) (domain.Talents, []Violation)` — removal always allowed; addition only when the result has no broken talent at `id` and the budget is respected.

- [ ] **Step 1: Write the failing tests**

```go
func TestTalentBudgetAndTier(t *testing.T) {
	trees := paladinTrees(t)
	alloc := domain.Talents{"def:1": 5, "def:2": 5, "def:3": 5, "def:4": 5, "def:5": 5, "def:6": 5, "def:7": 5, "def:8": 5, "def:9": 5, "def:10": 5}
	_, issues := Talents(level60Paladin(), alloc, trees)
	if issues.TotalPoints != 50 { t.Fatalf("total = %d", issues.TotalPoints) }
	over := domain.Talents{"def:1": 51}
	if v, _ := Talents(level10Paladin(), over, trees); !hasCode(v, "talent_budget_exceeded") { t.Fatal("budget") }
	if v, _ := Talents(level60Paladin(), alloc, trees); len(v) != 0 { t.Fatalf("unexpected %v", v) }
}

func TestTierGateAndPrerequisiteAndBroken(t *testing.T) {
	v, _ := Talents(level60Paladin(), domain.Talents{"def:row1": 1}, trees)
	if !hasCode(v, "talent_tier_locked") { t.Fatal("tier") }
	v, _ = Talents(level60Paladin(), domain.Talents{"def:prereq": 1, "def:child": 1}, trees)
	if !hasCode(v, "talent_prereq_unmet") { t.Fatal("prereq") }
	if _, violations := TrySetPoints(trees, domain.Talents{"def:child": 1}, "def:child", 2); len(violations) == 0 { t.Fatal("addition over broken talent must be rejected") }
	if next, _ := TrySetPoints(trees, domain.Talents{"def:child": 1}, "def:child", 0); next["def:child"] != 0 { t.Fatal("removal must be allowed") }
}
```

- [ ] **Step 2: Run them and confirm they fail**

Run: `go test ./internal/forever/validate -run 'TestTalent|TestTier' -v`
Expected: FAIL.

- [ ] **Step 3: Implement the rules**

Reject unknown talent IDs (`unknown_talent`), points above `MaxPoints` (`talent_rank_exceeded`), totals above `level-9` (`talent_budget_exceeded`), a nonzero talent whose zero-based row needs `5*row` points in non-broken talents above it (`talent_tier_locked`), and a talent whose `PrereqID` is below max rank or broken (`talent_prereq_unmet`). A broken talent keeps its points, contributes nothing to higher rows, and breaks its children. `TrySetPoints` allows decreases unconditionally and rejects an increase that leaves `id` broken or over budget, returning typed violations.

- [ ] **Step 4: Run the talent tests and confirm they pass**

Run: `go test ./internal/forever/validate -run 'TestTalent|TestTier' -v`
Expected: PASS.

- [ ] **Step 5: Add the malformed-input test**

An allocation map with a key not present in the class trees returns `unknown_talent` and `TalentIssues` listing it; a negative point value returns `talent_rank_exceeded`. No clamping occurs.

- [ ] **Step 6: Run all validation tests**

Run: `go test ./internal/forever/validate -v`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/forever/validate/talents.go internal/forever/validate/talents_test.go
git commit -m "feat(forever): validate talent budget, tiers, prerequisites and broken points"
```

### Task 8: SQLite store and character/build service

**Files:**
- Create: `internal/forever/domain/character.go`
- Create: `internal/forever/store/store.go`
- Create: `internal/forever/store/sqlite.go`
- Create: `internal/forever/service/service.go`
- Test: `internal/forever/store/sqlite_test.go`, `internal/forever/service/service_test.go`

**Interfaces:**
- `domain.Character{ID, Name string; Class, Spec, Race string; Level int; Professions []string; Region, Realm string; CreatedAt, UpdatedAt time.Time}`
- `domain.Build{ID, CharacterID string; Kind string; Name string; CatalogRevision string; Equipment domain.Equipment; Talents domain.Talents; Source string; Version int; Published bool; CreatedAt, UpdatedAt time.Time}`
- `store.NewSQLite(path string) (*SQLite, error)` and the interface it implements:

```go
type Store interface {
	CreateCharacterWithCurrent(ctx context.Context, ch domain.Character, current domain.Build) error
	GetCharacter(ctx context.Context, id string) (domain.Character, error)
	ListCharacters(ctx context.Context) ([]domain.Character, error)
	UpdateCharacter(ctx context.Context, ch domain.Character) error
	GetBuild(ctx context.Context, id string) (domain.Build, error)
	ListBuilds(ctx context.Context, characterID string) ([]domain.Build, error)
	CreateBuild(ctx context.Context, b domain.Build) error
	UpdateBuild(ctx context.Context, b domain.Build, expectedVersion int) (domain.Build, error)
	DeleteBuild(ctx context.Context, id string) error
}
```

- `service.New(st Store, cat *catalog.Catalog) *Service`; methods `CreateCharacter`, `GetCharacter`, `ListCharacters`, `UpdateCharacter`, `CreateBuild`, `GetBuild`, `ListBuilds`, `UpdateBuild`, `CopyBuild`, `DeleteBuild`.
- `service.UpdateBuild` returns `(*domain.Build, []validate.Violation, error)`, `store.ErrVersionConflict` on mismatch.
- `service.UpdateBuildInput{Name *string; Equipment *domain.Equipment; Talents *domain.Talents; Version int}`; omitted pointer fields are left unchanged.
- `service.ValidateBuild(ch domain.Character, eq domain.Equipment, tal domain.Talents) ([]validate.Violation, error)` — pure; resolves the catalog and validates without writing.

- [ ] **Step 1: Write the failing store test**

```go
func TestBuildIsolationAndVersioning(t *testing.T) {
	st := newTestStore(t)
	ch := mustCreateCharacter(t, st)
	current := mustCreateBuild(t, st, ch.ID, "current", "")
	plan := mustCopyBuild(t, st, current.ID, "Pre-raid")
	updated := plan
	updated.Equipment = domain.Equipment{"head": {ItemID: 720}}
	if _, err := st.UpdateBuild(ctx, updated, plan.Version); err != nil { t.Fatal(err) }
	got, _ := st.GetBuild(ctx, current.ID)
	if len(got.Equipment) != 0 { t.Fatal("current mutated") }
	if _, err := st.UpdateBuild(ctx, updated, plan.Version); !errors.Is(err, store.ErrVersionConflict) { t.Fatalf("err = %v", err) }
}
```

- [ ] **Step 2: Run it and confirm it fails**

Run: `go test ./internal/forever/store -run TestBuildIsolation -v`
Expected: FAIL.

- [ ] **Step 3: Implement the schema and store**

Open `modernc.org/sqlite` through `database/sql`; set `MaxOpenConns(1)`; enable `PRAGMA foreign_keys=ON`. Create `characters`, `builds` with `version INTEGER NOT NULL DEFAULT 1` and `UNIQUE(character_id, kind) WHERE kind='current'` (partial unique index). JSON-encode `equipment` and `talents`. `CreateCharacterWithCurrent` inserts the character and its current build in one transaction, rolling back both on any failure. `UpdateBuild` runs `UPDATE ... SET version=version+1 WHERE id=? AND version=?` and returns `ErrVersionConflict` when zero rows change. `DeleteBuild` refuses `kind='current'` with `ErrCannotDeleteCurrent`.

- [ ] **Step 4: Run the store test and confirm it passes**

Run: `go test ./internal/forever/store -run TestBuildIsolation -v`
Expected: PASS.

- [ ] **Step 5: Write the failing service test**

```go
func TestServiceRejectsInvalidAndPreservesPriorBuild(t *testing.T) {
	svc := newTestService(t)
	ch := mustCreateCharacterSvc(t, svc, "paladin", "paladin_retribution", "human", 60)
	cur := mustCreateBuildSvc(t, svc, ch.ID, "current")
	before := cur.Equipment
	_, violations, err := svc.UpdateBuild(ctx, cur.ID, service.UpdateBuildInput{
		Equipment: &domain.Equipment{"head": {ItemID: plateOnMageID}},
		Version:   cur.Version,
	})
	if err == nil || !hasCode(violations, "incompatible_armor") { t.Fatal("expected validation failure") }
	after, _ := svc.GetBuild(ctx, cur.ID)
	if !reflect.DeepEqual(before, after.Equipment) { t.Fatal("failed save mutated build") }
}
```

- [ ] **Step 6: Run it and confirm it fails**

Run: `go test ./internal/forever/service -run TestServiceRejects -v`
Expected: FAIL.

- [ ] **Step 7: Implement the service**

CreateCharacter validates name (1–40 runes, trimmed), class, spec belongs to class, race allowed for class, level 1–60, and at most two distinct known professions. It calls `CreateCharacterWithCurrent` so the character and an empty `current` build are written in one transaction. CreateBuild accepts `kind: "plan"` only (the current is created once), a name, and an optional `copyFrom` build; copying deep-copies equipment and talents. UpdateBuild resolves the catalog revision, merges name/equipment/talents, validates, and only on zero errors writes. On a validation failure return the violations and leave the store untouched. CopyBuild duplicates any build as a plan with a new ID. UpdateCharacter validates identity/spec/professions, revalidates capability warnings, and never deletes equipment. ValidateBuild runs equipment and talent validation on the supplied state and returns violations without reading or writing the store.

- [ ] **Step 8: Run the service test and confirm it passes**

Run: `go test ./internal/forever/service -run TestServiceRejects -v`
Expected: PASS.

- [ ] **Step 9: Run store and service tests together**

Run: `go test ./internal/forever/store ./internal/forever/service -count=1`
Expected: PASS.

- [ ] **Step 10: Commit**

```bash
git add internal/forever/domain/character.go internal/forever/store internal/forever/service
git commit -m "feat(forever): add SQLite store and character/build service"
```

### Task 9: Character and build HTTP API

**Files:**
- Create: `internal/forever/server/handlers_characters.go`
- Modify: `internal/forever/server/server.go`
- Test: `internal/forever/server/handlers_characters_test.go`

**Interfaces:**
- Routes: `GET|POST /api/characters`, `GET|PATCH /api/characters/{id}`, `POST /api/characters/{id}/validate`, `GET|POST /api/characters/{id}/builds`, `GET|PATCH|DELETE /api/builds/{id}`, `POST /api/builds/{id}/copy`.
- Success bodies use the domain JSON tags. Errors use `{"error":{"code","message"}}`.

- [ ] **Step 1: Write the failing API test**

```go
func TestCreateCharacterAndPlan(t *testing.T) {
	rec := post(t, srv, "/api/characters", `{"name":"Elara","class":"paladin","spec":"paladin_retribution","race":"human","level":60}`)
	if rec.Code != 201 { t.Fatalf("status %d", rec.Code) }
	var ch domain.Character
	json.Unmarshal(rec.Body.Bytes(), &ch)
	if ch.ID == "" { t.Fatal("no id") }
	list := get(t, srv, "/api/characters/"+ch.ID+"/builds")
	if !strings.Contains(list.Body.String(), `"kind":"current"`) { t.Fatal("no current") }
	plan := post(t, srv, "/api/characters/"+ch.ID+"/builds", `{"name":"Pre-raid"}`)
	if plan.Code != 201 { t.Fatalf("plan %d: %s", plan.Code, plan.Body) }
}
```

- [ ] **Step 2: Run it and confirm it fails**

Run: `go test ./internal/forever/server -run TestCreateCharacterAndPlan -v`
Expected: FAIL.

- [ ] **Step 3: Implement the handlers**

Decode with `DisallowUnknownFields`, a 1 MiB body limit and one trailing-value check (copy the existing TBC `decodeJSON` shape). `POST /validate` decodes `{equipment, talents}`, calls `service.ValidateBuild`, and returns `200 {"violations":[...]}` — findings are data, not an HTTP error. Map domain errors: `ErrNotFound` → 404 `not_found`; `ErrVersionConflict` → 409 `version_conflict`; `ErrCannotDeleteCurrent` → 409 `cannot_delete_current`; validation errors on a write → 422 with `{"error":{...,"violations":[...]}}`; anything else → 500 `internal_error`.

- [ ] **Step 4: Run the API test and confirm it passes**

Run: `go test ./internal/forever/server -run TestCreateCharacterAndPlan -v`
Expected: PASS.

- [ ] **Step 5: Add conflict and atomicity tests**

PATCH a build with a stale `version` → 409 and no change. PATCH with an invalid plate-on-mage item → 422, violations include `incompatible_armor`, stored build unchanged. `POST /validate` with the same invalid equipment returns 200 with that violation and writes nothing. DELETE a plan → 204 and it disappears from the list. DELETE the current → 409 `cannot_delete_current`.

- [ ] **Step 6: Run all server tests**

Run: `go test ./internal/forever/server -count=1`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/forever/server/handlers_characters.go internal/forever/server/server.go internal/forever/server/handlers_characters_test.go
git commit -m "feat(forever): expose character and build API with typed errors"
```

---

## Part 2 — Editor UI

### Task 10: Svelte app scaffold, API client and embedded build

**Files:**
- Create: `ui-forever/package.json`, `vite.config.js`, `index.html`, `src/main.js`, `src/app.css`
- Create: `ui-forever/src/App.svelte`, `src/lib/api.js`, `src/lib/router.svelte.js`
- Create: `cmd/forever-armory/webui/` (build output)
- Test: `ui-forever/src/lib/router.test.js`, `ui-forever/src/lib/api.test.js`, `internal/forever/server/ui_test.go`

**Interfaces:**
- `api.get(path)`, `api.post(path, body)`, `api.patch(path, body)`, `api.del(path)` throw `ApiError{code,message,violations}`.
- `router.parse(hash) -> {name:'library'} | {name:'editor', id, tab, buildId}`; `router.href(...)` builds the same shapes.
- Routes: `#/`, `#/characters/<id>/<gear|stats|talents>?build=<buildId>`.

- [ ] **Step 1: Write the failing router test**

```js
test('parses editor routes', () => {
  assert.deepEqual(parse('#/characters/abc/talents?build=p1'), { name: 'editor', id: 'abc', tab: 'talents', buildId: 'p1' });
  assert.deepEqual(parse('#/'), { name: 'library' });
  assert.equal(href('editor', { id: 'abc', tab: 'gear', buildId: 'p1' }), '#/characters/abc/gear?build=p1');
});
```

- [ ] **Step 2: Run it and confirm it fails**

Run: `cd ui-forever && node --test src/lib/router.test.js`
Expected: FAIL.

- [ ] **Step 3: Implement the router, API client and app shell**

Unknown hashes redirect to `#/`. `ApiError` parses the server's error body and keeps `violations`. `App.svelte` switches between `CharacterLibrary` (Task 11) and `Editor` (Task 12); until those exist, render headings `Library`/`Editor` so the shell is testable. Copy the dark token set and typography from `ui-finder/src/app.css` into `ui-forever/src/app.css` (no imports across projects).

- [ ] **Step 4: Run the unit tests and confirm they pass**

Run: `cd ui-forever && node --test src/lib/router.test.js src/lib/api.test.js`
Expected: PASS.

- [ ] **Step 5: Configure Vite and the embedded bundle**

`vite.config.js` builds to `../cmd/forever-armory/webui` with `emptyOutDir: true`, base `./`. Add `//go:embed all:webui` in `cmd/forever-armory/webui.go` and serve `index.html` at `/` plus `/assets/{path...}` with `fs.ValidPath`. Add npm scripts `dev`, `build`, `test:unit`, `test:e2e`.

- [ ] **Step 6: Write the failing UI-serving test**

```go
func TestServesBuiltIndex(t *testing.T) {
	rec := get(t, srv, "/")
	if rec.Code != 200 { t.Fatalf("status %d", rec.Code) }
	if !strings.Contains(rec.Body.String(), `<div id="app">`) { t.Fatal("index") }
}
```

- [ ] **Step 7: Build and run the test**

Run: `cd ui-forever && npm install && npm run build && cd .. && go test ./internal/forever/server -run TestServesBuiltIndex -v`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add ui-forever cmd/forever-armory/webui cmd/forever-armory/webui.go internal/forever/server
git commit -m "feat(forever): scaffold Svelte app with router, API client and embedded build"
```

### Task 11: Character library and creation

**Files:**
- Create: `ui-forever/src/lib/CharacterLibrary.svelte`, `CreateCharacterDialog.svelte`, `stores.svelte.js`
- Create: `ui-forever/e2e/library.spec.js`
- Modify: `ui-forever/src/App.svelte`, `src/app.css`

**Interfaces:**
- `stores.svelte.js` exports `state = $state({characters: [], loading, error})` and `loadCharacters()`, `createCharacter(input)`.
- `CreateCharacterDialog` props: `open`, `onclose`, `oncreated`. It fetches `/api/meta` for class/spec/race options and enforces race/class compatibility client-side using the server-provided `classRaces`.

- [ ] **Step 1: Write the failing e2e test**

```js
test('creates a character and reopens it', async ({ page }) => {
  await page.goto(baseURL);
  await page.getByRole('button', { name: 'Create character' }).click();
  await page.getByLabel('Name').fill('Elara');
  await page.getByLabel('Class').selectOption('paladin');
  await page.getByLabel('Specialization').selectOption('paladin_retribution');
  await page.getByLabel('Race').selectOption('human');
  await page.getByRole('button', { name: 'Create' }).click();
  await expect(page.getByRole('heading', { name: 'Elara' })).toBeVisible();
  await page.reload();
  await expect(page.getByRole('heading', { name: 'Elara' })).toBeVisible();
});
```

- [ ] **Step 2: Run it and confirm it fails**

Run: `cd ui-forever && npx playwright test e2e/library.spec.js`
Expected: FAIL — no Create character button.

- [ ] **Step 3: Implement the library and dialog**

Empty state copy: "No characters yet. Create one to start planning." Cards show name, `Level 60 Human · Paladin`, and a `Reopen` link. The dialog selects class → spec, race and up to two optional professions from `/api/meta`, disables invalid races, defaults level 60, and shows server validation messages inline. On success navigate to `#/characters/<id>/gear`.

- [ ] **Step 4: Run the e2e test and confirm it passes**

Run: `cd ui-forever && npx playwright test e2e/library.spec.js`
Expected: PASS.

- [ ] **Step 5: Add the validation e2e test**

Submitting an empty name and an undead paladin shows `invalid_character_name` and `race_class_mismatch` messages, keeps the dialog open, and creates nothing.

- [ ] **Step 6: Run the library specs**

Run: `cd ui-forever && npx playwright test e2e/library.spec.js`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add ui-forever/src ui-forever/e2e/library.spec.js
git commit -m "feat(forever): character library and creation flow"
```

### Task 12: Editor shell, identity header and 17-slot gear grid

**Files:**
- Create: `ui-forever/src/lib/Editor.svelte`, `IdentityHeader.svelte`, `GearGrid.svelte`, `SlotCard.svelte`, `StatsPanel.svelte`
- Create: `ui-forever/e2e/editor.spec.js`
- Modify: `ui-forever/src/App.svelte`

**Interfaces:**
- `Editor` props: `characterId`, `tab`, `buildId`. It loads the character, its builds, selects `buildId` or the current build, and renders the tabs.
- `GearGrid` renders `domain.SlotOrder` in the approved eight-left / nine-right layout and emits `onopenslot(slot)`.
- `SlotCard` props: `slot`, `label`, `item` (nullable), `onopen`.

- [ ] **Step 1: Write the failing e2e test**

```js
test('shows all 17 slots and the empty current build', async ({ page, request }) => {
  const ch = await seedCharacter(request);
  await page.goto(`${baseURL}/#/characters/${ch.id}/gear`);
  await expect(page.getByRole('heading', { name: ch.name })).toBeVisible();
  await expect(page.getByTestId('gear-slot')).toHaveCount(17);
  await expect(page.getByText('Empty')).toHaveCount(17);
  await expect(page.getByRole('tab', { name: 'Talents' })).toBeVisible();
});
```

- [ ] **Step 2: Run it and confirm it fails**

Run: `cd ui-forever && npx playwright test e2e/editor.spec.js`
Expected: FAIL.

- [ ] **Step 3: Implement the shell**

Identity header shows the class-colored name, `Level 60 · Human · Paladin · Retribution`, professions when set, and an `Edit character` action opening a dialog for name, race, spec and professions (PATCH; a spec change shows the returned capability warnings without deleting equipment). The build switcher lists `Current` plus plans and offers `New plan` (empty or copy current), `Duplicate`, `Rename`, and `Delete` (confirmation; deleting Current is disabled). Structural actions write immediately and refresh the build list; equipment and talent edits stay local until Save. Tabs are Gear / Stats / Talents with `aria-selected`. Gear slots show item name, ilvl badge, quality color, and enchant line when present. Stats renders the explicit unavailable state from the spec: "Whole-character stats are not available yet. Item stats are shown in the picker." Never sum item stats.

- [ ] **Step 4: Run the e2e test and confirm it passes**

Run: `cd ui-forever && npx playwright test e2e/editor.spec.js`
Expected: PASS.

- [ ] **Step 5: Add the plan-lifecycle e2e test**

Create a plan copied from current, open it, assert Current remains empty in its own view; duplicate it, rename one, delete the other with confirmation, and assert the URL build id of the surviving plan is unchanged. Edit the character's spec to one whose class cannot use the equipped armor and assert the warning is shown and the equipment is still present.

- [ ] **Step 6: Run the editor specs**

Run: `cd ui-forever && npx playwright test e2e/editor.spec.js`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add ui-forever/src ui-forever/e2e/editor.spec.js
git commit -m "feat(forever): editor shell with identity header and 17-slot gear grid"
```

### Task 13: Item and enchant picker

**Files:**
- Create: `ui-forever/src/lib/ItemPicker.svelte`, `EnchantPicker.svelte`, `ItemTooltip.svelte`
- Create: `ui-forever/e2e/picker.spec.js`
- Modify: `ui-forever/src/lib/Editor.svelte`

**Interfaces:**
- `ItemPicker` props: `open`, `slot`, `character`, `onclose`, `onapply(itemInstance)`.
- `onapply` updates the editor's local equipment and triggers validation; it does not write to the server. Save is the only writer.
- The picker requests `/api/items?class=<class>&slot=<slot>&q=<query>&limit=50`; each row's `compatible` and `reason` decide whether Apply is enabled.

- [ ] **Step 1: Write the failing e2e test**

```js
test('filters by slot and applies an item', async ({ page, request }) => {
  const ch = await seedCharacter(request);
  await page.goto(`${baseURL}/#/characters/${ch.id}/gear`);
  await page.getByTestId('gear-slot-head').click();
  await page.getByPlaceholder('Search items').fill('bad hat');
  await expect(page.getByTestId('picker-row')).toHaveCount(1);
  await page.getByTestId('picker-row').first().click();
  await page.getByRole('button', { name: 'Apply to head' }).click();
  await expect(page.getByTestId('gear-slot-head')).toContainText('Bad Hat');
  await page.reload();
  await expect(page.getByTestId('gear-slot-head')).toContainText('Bad Hat');
});
```

- [ ] **Step 2: Run it and confirm it fails**

Run: `cd ui-forever && npx playwright test e2e/picker.spec.js`
Expected: FAIL.

- [ ] **Step 3: Implement the Items tab**

Rows show name, ilvl, armor/weapon type, source label and class fit, with a source filter populated from the returned source kinds. The detail panel shows the selected variant's stats, weapon damage/speed, set name, and an explicit `Level requirement: unknown` row (never a fabricated 60). When `variants.length > 0`, show a variant select and send `variantIlvl`; when `randomSuffixIds.length > 0`, show a suffix select. Incompatible rows are inspectable but Apply is disabled with the reason; unknown-source rows are labelled, not hidden.

- [ ] **Step 4: Run the e2e test and confirm it passes**

Run: `cd ui-forever && npx playwright test e2e/picker.spec.js`
Expected: PASS.

- [ ] **Step 5: Implement the Enchants tab**

The Enchants tab calls `/api/enchants?slot=<slot>&class=<class>` and, on Apply, updates the local item instance's enchant and triggers validation; nothing is written until Save. Switching tabs preserves the selected item. A weapon-only effect cannot be applied to an armor slot; Apply is disabled with `incompatible_enchant_slot`.

- [ ] **Step 6: Add the variant, incompatibility and retention e2e test**

Seed a build whose head item's stats exist only inside `scalingOptions`; assert the slot card and a reload show those stats, not zeros or another variant's values. Then pick a plate item for a mage plan, assert Apply is disabled and `incompatible_armor` is shown, and assert the existing head item is unchanged. Add an item with an unknown source and confirm it is shown but labelled.

- [ ] **Step 7: Run the picker specs**

Run: `cd ui-forever && npx playwright test e2e/picker.spec.js`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add ui-forever/src ui-forever/e2e/picker.spec.js
git commit -m "feat(forever): item and enchant picker with variant and suffix selection"
```

### Task 14: Talent editor

**Files:**
- Create: `ui-forever/src/lib/TalentEditor.svelte`, `TalentTree.svelte`, `TalentNode.svelte`
- Create: `ui-forever/e2e/talents.spec.js`
- Modify: `ui-forever/src/lib/Editor.svelte`

**Interfaces:**
- `TalentEditor` props: `character`, `build`, `onchange(talents)`. It loads `/api/talents?class=<class>` and renders three trees side by side at ≥1000px, tabs below.
- `TalentNode` props: `talent`, `points`, `state` (`available|allocated|maxed|locked|broken`), `onactivate`.
- Point changes update the editor's local talents and call `POST /api/characters/{id}/validate`; returned violations drive node states and `+` disabling. Nothing is written until Save.

- [ ] **Step 1: Write the failing e2e test**

```js
test('allocates points, locks lower tiers and confirms reset', async ({ page, request }) => {
  const ch = await seedCharacter(request);
  await page.goto(`${baseURL}/#/characters/${ch.id}/talents`);
  const first = page.getByTestId('talent-node').first();
  await first.click();
  await page.getByRole('button', { name: '+ Rank' }).click();
  await expect(page.getByTestId('point-total')).toContainText('1 / 51');
  await expect(page.getByTestId('talent-node').nth(4)).toHaveAttribute('data-state', 'locked');
  await page.getByRole('button', { name: 'Reset talents…' }).click();
  await page.getByRole('button', { name: 'Cancel' }).click();
  await expect(page.getByTestId('point-total')).toContainText('1 / 51');
});
```

- [ ] **Step 2: Run it and confirm it fails**

Run: `cd ui-forever && npx playwright test e2e/talents.spec.js`
Expected: FAIL.

- [ ] **Step 3: Implement the editor**

Point summary shows per-tree totals and `spent / (level-9)`. Clicking a node opens a details popover with `− Rank` / `+ Rank`. `+` is disabled when the server would return `talent_tier_locked`, `talent_prereq_unmet`, `talent_budget_exceeded` or `talent_rank_exceeded`; the reason is shown as text, not just a disabled control. Every change updates local state, calls the validate endpoint (debounced 150 ms), and invokes `onchange` so the editor marks the build dirty. Broken talents keep their points and get `data-state="broken"` with an explanatory legend. Reset opens a confirmation; on confirm it clears the local known allocations and revalidates. An ID present in a stored build but absent from the current trees (only possible after a future catalog change) renders read-only and blocks saving with `unknown_talent`; reset is disabled for that build.

- [ ] **Step 4: Run the e2e test and confirm it passes**

Run: `cd ui-forever && npx playwright test e2e/talents.spec.js`
Expected: PASS.

- [ ] **Step 5: Add the broken-talent e2e test**

Allocate a prerequisite and child, remove the prerequisite, assert the child keeps its points, is marked broken, and the total excludes it from unlocking further rows.

- [ ] **Step 6: Run the talent specs**

Run: `cd ui-forever && npx playwright test e2e/talents.spec.js`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add ui-forever/src ui-forever/e2e/talents.spec.js
git commit -m "feat(forever): talent editor with tiers, prerequisites and broken talents"
```

### Task 15: Save states, full-flow verification and responsive pass

**Files:**
- Create: `ui-forever/src/lib/SaveBar.svelte`, `ViolationList.svelte`
- Create: `ui-forever/e2e/full-flow.spec.js`
- Modify: `ui-forever/src/lib/Editor.svelte`, `src/app.css`, `playwright.config.js`

**Interfaces:**
- `SaveBar` props: `dirty`, `saving`, `version`, `onsave`. It renders "All changes saved" or "Unsaved changes" and disables Save while clean.
- `ViolationList` props: `violations`; renders code, message and slot, with `role="alert"`.

- [ ] **Step 1: Write the failing full-flow e2e test**

```js
test('edit, save, reload, and keep a failed save intact', async ({ page, request }) => {
  const ch = await seedCharacter(request);
  await page.goto(`${baseURL}/#/characters/${ch.id}/gear`);
  await page.getByTestId('gear-slot-head').click();
  await page.getByPlaceholder('Search items').fill('bad hat');
  await page.getByTestId('picker-row').first().click();
  await page.getByRole('button', { name: 'Apply to head' }).click();
  await expect(page.getByTestId('save-bar')).toContainText('Unsaved changes');
  await page.getByRole('button', { name: 'Save changes' }).click();
  await expect(page.getByTestId('save-bar')).toContainText('All changes saved');
  await page.reload();
  await expect(page.getByTestId('gear-slot-head')).toContainText('Bad Hat');
});
```

- [ ] **Step 2: Run it and confirm it fails**

Run: `cd ui-forever && npx playwright test e2e/full-flow.spec.js`
Expected: FAIL.

- [ ] **Step 3: Implement save state and error handling**

Track `dirty` on any local mutation; Save sends the full build with `version`. On 409 `version_conflict`, show "This build changed elsewhere. Reload to see the latest version." and keep local edits on screen. On 422, list violations, keep the picker/dialog open, and leave the saved build untouched. On success, replace the local build and set `dirty=false`. Add a `beforeunload` guard while dirty.

- [ ] **Step 4: Run the full-flow test and confirm it passes**

Run: `cd ui-forever && npx playwright test e2e/full-flow.spec.js`
Expected: PASS.

- [ ] **Step 5: Add the responsive and keyboard checks**

At 390px the gear grid is one column with no horizontal overflow, the talent editor shows one tree through tabs, the save bar stays visible, and Tab reaches every gear slot and talent node with a 3px focus outline. Run the same flow in a touch-enabled Playwright project (`hasTouch: true`) and tap a gear slot, a picker row and a talent node. Add these assertions to `full-flow.spec.js`.

- [ ] **Step 6: Run the whole suite**

Run: `go test ./internal/forever/... -count=1 && cd ui-forever && npm run test:unit && npm run build && npx playwright test`
Expected: all Go tests, unit tests, build and every Playwright spec pass.

- [ ] **Step 7: Rebuild and commit the embedded bundle**

Run: `cd ui-forever && npm run build`
```bash
git add ui-forever cmd/forever-armory/webui
git commit -m "feat(forever): save states, full-flow verification and responsive editor"
```
