# Project State — TBC Upgrade Finder

Updated: 2026-08-30

Authoritative operator instructions: `docs/upgrade-finder.md`.
Approved armory design: `docs/superpowers/specs/2026-08-29-svelte-armory-design.md`.
Approved implementation plan: `docs/superpowers/plans/2026-08-29-svelte-armory.md`.

## Current status

The Svelte armory migration is complete and committed. The local upgrade finder now imports an individual-sim link, renders a server-enriched 17-slot armory review with deterministic base-plus-gear stats, runs the existing ranking flow, displays the report, supports clipboard copy, and supports cancellation.

## Working behavior

### Backend

- `Import` accepts individual-sim links and rejects malformed, raid-sim, incompatible-version, and unsupported equipment references before enrichment or ranking.
- Unsupported nonzero item, random-suffix, gem, and enchant-effect IDs return typed `incompatible_*` validation errors. Main equipment and enabled item-swap records use the same validation path; empty item specs remain ordinary empties.
- `Catalog` indexes items, gems, enchant effects, zones, NPCs, and random suffixes from the bundled `UIDatabase`.
- `EnrichArmory` emits all 17 canonical slots in display order: Head, Neck, Shoulder, Back, Chest, Wrist, Main Hand, Off Hand, Hands, Waist, Legs, Feet, Finger 1, Finger 2, Trinket 1, Trinket 2, Ranged.
- Armory metadata includes item names/quality/icons/phases/set names, scaled random suffixes, gems, enchants, declared sockets, socket bonuses, and empty-slot/empty-socket state.
- Socket matching follows `core.ColorIntersects`, including hybrid, prismatic, and meta behavior. Suffix scaling follows `core.ItemEquipmentBaseStats`.
- Armory totals come from `core.ComputeStats` on a cloned setup with buffs, consumes, talents, bonus stats, and item swaps cleared. It does not run simulation iterations and is labeled **unbuffed (base + gear)**.
- Raw stat keys use unprefixed snake_case `stats.StatName()` values, including `mp5`. Derived keys are limited to `melee_hit_percent`, `spell_hit_percent`, `melee_crit_percent`, `spell_crit_percent`, `ranged_hit_percent`, `ranged_crit_percent`, and `block_percent`.

### HTTP server

- `POST /api/import` preserves existing successful fields and adds only `gear`, `stats`, and `derivedStats`.
- The server initializes one immutable catalog and passes it to import enrichment and the default ranking service.
- `GET /` serves the embedded Vite entry document.
- `GET /assets/{path...}` accepts only `fs.ValidPath` wildcards and reads only below embedded `upgrade_ui/assets/`. Unknown, empty, traversal, and nested-missing paths return structured `not_found` errors.
- Existing loopback-only binding, API routes, strict JSON decoding, 1 MiB body limit, 500 ms client polling contract, job statuses, report fields, and cancellation behavior remain intact.

### Frontend

- Source of truth is the isolated `ui-finder/` Svelte 5 + Vite project. Shared state is the module-scoped rune object in `src/lib/stores.svelte.js`.
- `ImportPanel`, `ArmoryView`, `GearSlot`, `StatPanels`, `RankingPanel`, and `ReportView` are focused components.
- The browser renders server-owned armory data. It performs no item lookup, simulator calculation, socket-rule evaluation, or stat conversion.
- Gear uses an eight-slot left column, an empty center track, and a nine-slot right column; narrow layouts collapse to one column without changing canonical order.
- Item icons use the ZAM CDN with quality-colored, slot-initial fallbacks. Socket colors and item qualities follow the protobuf enum values.
- Ranking preserves the original pasted link, validates screening/confirmation budgets, polls every 500 ms, retains completed reports, clears canceled/failed reports, and guards stale asynchronous actions.
- Reports display baseline/assumptions/exclusions/failures/confirmed upgrades, human-readable target slots and sources, per-upgrade assumptions, and JSON clipboard copy status.
- The generated Vite bundle is committed under `cmd/wowsimcli/cmd/upgrade_ui/`; Go builds and Go tests do not require Node.

## Verification

Final required checks passed:

```text
rtk go test ./cmd/wowsimcli/cmd/upgrades -count=1    39 passed
rtk go test ./cmd/wowsimcli/cmd/... -count=1         47 passed in 2 packages
rtk go build -o wowsimcli ./cmd/wowsimcli          passed
cd ui-finder && rtk npm ci                         40 packages installed
cd ui-finder && rtk npm run build                  Vite 7.3.6; 117 modules transformed; passed
cd ui-finder && rtk npm run test:e2e               1 Chromium test passed
```

Focused API coverage also passes with `rtk node --test src/lib/api.test.js` (`2 passed`). `npm ci` reports the environment's existing blocked optional `esbuild` install script; it does not prevent the build or smoke test.

The real-server Playwright smoke uses the fixed individual-link fixture and confirms:

1. `TestMage` imports successfully with exactly 17 `[data-slot]` gear elements.
2. Socket lines, enchant lines, raw stats, derived percentages, **unbuffed (base + gear)** labeling, and the wowsims attribution are visible.
3. A real ranking job shows queued/running progress, completes a report, and copies the JSON report successfully.
4. A second real job with longer budgets cancels successfully, announces `Ranking canceled.`, and leaves no report.
5. Generated nested hashed assets load through the Go server.

## Intentional boundaries

- No 3D model, talents tab, item hover tooltips, buffed-stat mode, persistence, pricing, accounts, remote server, or ranking-domain rewrite.
- The browser exposes the existing max-phase, unknown-source, screening, and confirmation controls. It does not search alternate gem/enchant policies or source-name filters.
- The armory snapshot is not a simulation result and excludes temporary buffs, consumes, talents, and alternate item-swap gear.

## Known follow-up risks

- Simulator database attachment remains process-wide through the existing `sync.Once` path; concurrent first jobs should be serialized or attachment should become per-job-safe.
- DELETE-failure recovery can briefly schedule duplicate polling timers while an already-fired poll request is in flight. Cancellation remains recoverable; this was recorded as a non-blocking review note.
- There is no CI workflow running the pinned Go, build, and browser checks automatically.
- The optional favicon remains a cosmetic 404.
