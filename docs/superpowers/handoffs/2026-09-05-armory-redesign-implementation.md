# Armory Redesign — Implementation Handoff

## Assignment

Implement the approved armory redesign end-to-end. The authoritative implementation plan is:

`docs/superpowers/plans/2026-09-05-armory-redesign.md`

Read that plan in full before editing. Execute its five tasks in order, preserve the declared interfaces, and check a task checkbox only after its listed verification passes.

Authoritative design and research:

- `docs/superpowers/specs/2026-09-05-armory-redesign-design.md` — approved design.
- `docs/armory-redesign-research.md` — why the 3D model is a labeled placeholder (gated on renderer permission and verified browser asset access; do not attempt to ship a viewer).

When a design detail and a plan detail conflict, follow the design. When the real code differs from an expected snippet (naming, line positions), adapt to the actual code without weakening behavior or widening scope.

## Starting State

- Working tree is clean at commit `2af64bd87` on `main`. Implement on top of it; do not rebase or merge unrelated work.
- The finder is a local Go binary (`cmd/wowsimcli`) plus an isolated Svelte 5 + Vite UI in `ui-finder/`. The built UI bundle is committed under `cmd/wowsimcli/cmd/upgrade_ui/`.
- `cmd/wowsimcli/cmd/upgrades/` owns import, armory enrichment, catalog, and ranking. `cmd/wowsimcli/cmd/upgrade_server.go` owns HTTP/job orchestration only.
- The current armory renders a two-column gear list with an empty center, stat panels inline, no tabs, and no talents view. Item enrichment lives in `armory.go` (`enrichItem`, `canonicalGearSlots`); response types in `types.go`; the import response in `upgrade_server.go`.
- Bundled reusable data: `assets/enchants/descriptions.json` (enchant effect text keyed by effect ID), `ui/core/talents/trees/*.json` (nine class talent trees), `assets/img/*.jpg` (per-spec backdrop art), `assets/database/db.bin` (embedded via `assets/database/loader.go`).
- Host is Windows; git CRLF warnings are normal. Run commands with the repo's `rtk` wrapper.

## Non-Negotiable Product Rules

- No new frontend or Go dependencies. No changes to ranking, simulation, validation codes, or job lifecycle.
- Import response changes are additive only: `GearSlotData.ilvl`, `EnchantData.description`, top-level `talentsString`. Nothing existing is removed or renamed.
- Enchant descriptions resolve server-side by effect ID from `assets/enchants/descriptions.json` (embedded like `assets/database/loader.go` embeds `db.bin`), falling back to the enchant name. Never invent stats for proc-only enchants (e.g. `Mongoose` keeps its name).
- `ilvl` comes from the enriched `UIItem.GetIlvl()`. Do not derive it from phase or item ID.
- The browser stays a renderer: no item lookup, stat math, socket rules, or simulator logic in `ui-finder/`.
- Keep the Playwright contract: `data-region="armory-view"`, `data-slot`, the `{slotName} sockets` aria-label, and the `TestMage` heading survive the redesign. Update e2e assertions only where the redesign intentionally changes them (enchant line, tabs).
- Preserve all 17 canonical slots. Gear columns and the weapons strip are keyed by explicit slot names (`Head`, `Neck`, `Shoulder`, `Back`, `Chest`, `Wrist`, `Hands` | `Waist`, `Legs`, `Feet`, `Finger 1`, `Finger 2`, `Trinket 1`, `Trinket 2` | `Main Hand`, `Off Hand`, `Ranged`) — never by array slices.
- Stats tab keeps the exact server data and the **raid buffed (link settings)** label. Talents tab is read-only from the imported `talentsString`.
- The 3D stage is a labeled placeholder with per-spec backdrop art and a disabled `Activate 3D` control. Do not load any viewer, script, or model asset.
- After every UI change, run `cd ui-finder && rtk npm run build` and commit the regenerated `cmd/wowsimcli/cmd/upgrade_ui/` bundle with the source. `rtk go build` and Go tests must remain Node-free.
- Commit per task with the messages in the plan (`feat:` / `docs:`).

## Implementation Sequence

1. **Backend display fields** — `assets/enchants` loader + test; `GearSlotData.Ilvl`; `EnchantData.Description`; `importResponse.TalentsString`; extend armory and server tests.
2. **Talent decoder and tree view** — `talents.js` + node test; `TalentTrees.svelte` read-only rendering of the bundled trees.
3. **WoWSims-style gear card** — `GearSlot.svelte` rewrite (ilvl badge, rarity-colored name, enchant effect line, socket strip, details disclosure); CSS; e2e assertion updates; rebuild bundle.
4. **Character-centered layout with tabs** — `CharacterHeader.svelte`, `CharacterStage.svelte`, `ArmoryView.svelte` rewrite (tabs Gear/Stats/Talents, stage, weapons strip); CSS incl. responsive rules; e2e tab assertions; rebuild bundle.
5. **Docs, full verification, manual smoke** — update `docs/upgrade-finder.md` and `docs/STATE.md`; run the full suite; manual smoke at wide and narrow widths.

Keep cross-task contracts stable: `enchants.Descriptions()`, `GearSlotData.Ilvl`, `EnchantData.Description`, response field `talentsString`, `decodeTalentsString`/`rankAt`/`treePoints`, `GearSlot` `side` prop, `TalentTrees` `{ class, talentsString }` props.

## Required Evidence Before Handoff Back

Run and report these exact checks:

```bash
rtk go test ./assets/enchants -count=1
rtk go test ./cmd/wowsimcli/cmd/upgrades -count=1
rtk go test ./cmd/wowsimcli/cmd -count=1
rtk go build -o wowsimcli ./cmd/wowsimcli
```

Then, from `ui-finder`:

```bash
rtk npm run build
rtk node --test src/lib/talents.test.js
rtk npm run test:e2e
```

Browser evidence must exercise the real local process (`rtk go run ./cmd/wowsimcli rank-upgrades --addr 127.0.0.1:43199 --no-browser`, then open the printed URL):

1. Import `cmd/wowsimcli/cmd/upgrades/testdata/retribution_no_settings_link.txt`: compact header, 17 slot cards across two columns plus a three-card weapons strip, ilvl badges, rarity-colored names, enchant effect lines, socket strips.
2. Stats tab shows **raid buffed (link settings)** raw and derived panels.
3. Talents tab shows three Paladin trees with allocated rank pips.
4. Narrowing below 640 px collapses gear to one readable column with the stage compacted.
5. A minimal ranking run (phase/screening/confirmation `1`) still shows progress, report, copy, and cancel behavior.

## Completion Report Format

Return only:

- changed paths and their responsibility;
- test/build commands with observed results;
- browser-smoke evidence;
- any unresolved discrepancy between the plan and the real code, and how it was resolved.

Do not return a partial implementation, scaffold, stubs, or a follow-up list. The `Activate 3D` button is deliberately disabled; that is the designed end state, not unfinished work.
