# Forever manual editor — implementation handoff

Date: 2026-10-25
Status: Phase 1 complete and approved. Phase 2 plan written and committed. Implementation not started.
Branch at handoff: `docs/forever-armory-design` (latest design commit `8bb83bd24`, plus this
handoff commit).
Execution method chosen: **native inline** — one agent implements the plan task by task in a new session.

## Goal

Build the Phase 2 Forever armory: a local, single-user armory/planner where a character is created
by race/class/spec/level, owns one Current build and any number of named plans, and every equipment
and talent edit is validated against a pinned Forever catalog before it is saved.

Phase 2 does **not** include accounts, ownership, publication, public URLs, imports, statistics, 3D
or simulation. Those are Phase 3+ in the roadmap.

## Read these, in this order

1. `docs/superpowers/plans/2026-10-25-forever-manual-editor.md` — the plan to execute. 15 tasks,
   121 steps. This is the primary document.
2. `docs/forever-upstream-data-contract.md` — identifier domains, validation semantics, pinned
   hashes and open items. It governs where anything disagrees.
3. `docs/superpowers/specs/2026-10-01-forever-armory-design.md` — product scope, domain model,
   gates, verification contract.
4. `docs/forever-armory-wowsims-review.md` — why the upstream boundary exists and what was
   reproduced (the TBC import panic).
5. `docs/plans/2026-10-20-forever-armory-roadmap.md` — phase sequencing and accepted wireframes.
6. `docs/wireframes/` — the accepted visual baseline (seven screens). Use it as direction, not
   pixel spec.

## Decisions already made — do not re-litigate without asking

- **Code placement:** new `cmd/forever-armory`, `internal/forever/...`, `ui-forever/`, and
  `tools/forever-catalog` in this repo. Forever code imports nothing from `sim/`, `cmd/wowsimcli/`
  or `assets/database`.
- **Persistence:** SQLite (`modernc.org/sqlite`, already in `go.mod`) from Phase 2, behind a
  `Store` interface. Phase 3 adds ownership/publication on the same interface.
- **Save model:** explicit Save. Equipment/talent edits are local; `POST /api/characters/{id}/validate`
  returns inline feedback without writing. Only Save (PATCH) writes, with optimistic `version` and
  `409 version_conflict`.
- **Character model:** name, race, class, specialization, level (1–60), up to two optional
  professions, optional region/realm. Level 60 is the only supported equipment-planning level.
- **Catalog:** pinned `wowsims/forever@0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4`, normalized into
  committed assets by the generator, with the manifest hashes in the contract.
- **Talents:** stored by identity (`def:<definitionId>` or a location fallback), never position.
  Budget `level-9`; tier gate `5*row`; prerequisite at max rank; broken talents keep points.
- **Reuse posture:** documented risk, not a legal clearance. Keep the catalog behind the generator/
  loader seam and keep attribution.

## Repo state at handoff

The design branch is 11 commits ahead of `main`; all Forever work is documentation. Latest commits:

```text
8bb83bd24 docs: add Forever manual editor implementation plan
e680809c8 docs: add specialization to the Forever character model
ab8dcd288 docs: refresh talent wireframe captures and record Phase 1 sign-off
983691606 docs: approve gated native WoWSims import route
c24bf4caa docs: revise Forever armory spec and add upstream data contract
```

There are **pre-existing unrelated changes** in the working tree. They predate this work and must
not be staged, committed, reverted, formatted or "cleaned up" as part of Forever tasks:

```text
 M .gitignore
 M cmd/wowsimcli/cmd/upgrade_ui/assets/...        (embedded TBC bundle)
 M cmd/wowsimcli/cmd/upgrade_ui/index.html
 M docs/STATE.md
 M docs/superpowers/plans/2026-09-06-independent-character-preview.md
 M tools/character-assets/README.md
 M ui-finder/scripts/generate-talent-icons.mjs
?? .pi/
?? .superdesign/
```

Do not modify the TBC application or its assets at all.

## How to execute

1. Create the feature branch from the handoff commit:
   `git switch -c feat/forever-manual-editor`
2. Read the plan end to end before touching code. Then execute tasks 1–15 in order. Do not reorder
   tasks 1–3: later tasks consume the generated assets and the loader.
3. Follow each task's TDD steps literally: write the test, run it and see it fail, implement, run it
   and see it pass, commit with the message given. Keep the commits per task.
4. Task 1 and Task 2 need network access to `raw.githubusercontent.com` to download the pinned
   payload and capability TS. The generated assets are committed; runtime never fetches upstream.
5. If a test shows the plan is wrong, stop, fix the plan document in the same commit series, and
   record the divergence in this handoff. Do not silently invent an interface that later tasks do
   not know about.
6. Run the full verification at Task 15:
   `go test ./internal/forever/... ./tools/... -count=1`
   `cd ui-forever && npm run test:unit && npm run build && npx playwright test`
7. Manual smoke: `go run ./cmd/forever-armory --db /tmp/forever-smoke.db`, create a paladin, equip
   an item, allocate a talent, save, restart, confirm it persisted.

## Gotchas discovered while planning

- **Stats live inside `scalingOptions`.** Most catalog rows have no top-level `stats`/`ilvl`; the
  default variant is key `"0"`. Flatten it, and keep other variants separately. Do not default to
  zeros.
- **Numeric enums.** `db.json` stores `type: 7`, `armorType: 2`, `weaponType: 9`, `handType: 2`,
  and stats as `{"27": 62}`. Map them through the enums parsed from the pinned `common.proto`;
  never hardcode TBC values. `Class` has numeric gaps (Shaman 7, Mage 8, Warlock 9, Druid 11).
- **Shirts/tabbards have no `type`.** Drop entries with item type 0; they are not 17-slot items.
- **Spec-level dual-wield.** The pinned `SPEC_CAN_DUAL_WIELD` covers hunter, rogue, DPS warrior and
  protection warrior. The plan applies that as every spec of those classes. Enhancement shaman is
  an open contract item (#9) — do not guess a talent-based rule.
- **No required character level in the catalog.** Never show or validate one; show `unknown`.
  Level 60 is the supported equipment level.
- **No gems/sockets** in the pinned proto (fields reserved). Do not add socket UI.
- **All main items are phase 1**, so there is no phase filter.
- **Enchant identity is effect + slot**, with item/spell as cross-references. An ambiguous match is
  an error (`ambiguous_enchant`), never a pick.
- **Talent strings are positional and clamp.** Our decoder must validate, not clamp. Preserve
  unknown IDs rather than deleting them.
- **Explicit save, not auto-save.** The wireframe has "Unsaved changes" and "Save changes". Do not
  make edits write immediately; structural actions (create/rename/delete/duplicate) are the only
  immediate writes.
- **Embedded UI.** `ui-forever` builds to `cmd/forever-armory/webui`; the Go binary embeds it. Run
  `npm run build` before any Go test that serves the index, and commit the bundle where the plan
  says to.

## Implementation divergences

Recorded as implementation proceeds. The plan is corrected in the same commit series when a test
proves it wrong.

1. **Pinned payload hashes were CRLF-normalized (Task 1).** The contract recorded
   `9fa8335e…` / `eeb3ff77…`. Those are the SHA-256 values of the files after converting LF to
   CRLF. The bytes actually served by `raw.githubusercontent.com` hash to `e78be48c…` /
   `6de2991a…`. The generator hashes fetched bytes directly, so the contract, plan and review were
   corrected to the raw-byte values. `catalog_revision` therefore ends in the corrected hashes.
2. **Undead paladins are legal (Task 5).** The pinned `CLASS_RACES` lists human, dwarf and undead
   for paladin, so the plan's `undeadPaladin` mismatch fixture was wrong. Validation reads the
   pinned table; the plan's Task 5 and Task 11 fixtures now use an orc paladin instead.
3. **Spec changes cannot change armor legality (Task 12).** Every spec of a class shares the
   class armor table, and the pinned data's only spec-level rule is dual-wield (which covers whole
   classes). The Task 12 capability-warning e2e therefore changes race to make an equipped
   faction-restricted item incompatible. Equipment is preserved and the warning is shown.
4. **The Task 13 picker test contradicted explicit save.** Its snippet reloaded after Apply, which
   would require an immediate write. The spec now asserts the local slot update; Task 15's
   full-flow test covers save plus reload.
5. **Pinned items have only `scalingOptions` key `"0"` (Task 13).** No non-default variants exist in
   the payload, so the picker's variant select is dormant. The test asserts the flattened default
   variant's ilvl and stats rather than another variant's values.
6. **Enchants endpoint gained `includeIncompatible=1` (Task 13).** The picker must show a weapon-only
   effect on an armor slot as disabled with `incompatible_enchant_slot`, but the default
   `/api/enchants?slot=` response filters those rows out. The opt-in flag returns them with
   `compatible`/`reason`; default behavior is unchanged.

## Definition of done for Phase 2

- All 15 plan tasks complete, each with its tests and commit.
- All Go tests, UI unit tests, production build and Playwright specs pass.
- Manual smoke passes, including persistence across a server restart.
- The roadmap's Phase 2 exit holds: usable manual editing against the supported real data, with
  observable plan isolation and validation behavior.
- No TBC files changed; no unrelated working-tree changes staged.

## Opening prompt for the new session

```text
Read docs/superpowers/handoffs/2026-10-25-forever-manual-editor.md, then execute
docs/superpowers/plans/2026-10-25-forever-manual-editor.md using native inline implementation.

Create the branch feat/forever-manual-editor from the current HEAD, then work task by task:
write the failing test first, run it, implement, run it, and commit with the message in the task.
Read the spec and the data contract before starting. Do not modify the TBC application or stage
the unrelated dirty files listed in the handoff.
```
