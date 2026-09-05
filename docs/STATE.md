# Project State — TBC Upgrade Finder

Updated: 2026-09-05

Authoritative operator instructions: `docs/upgrade-finder.md`.
Approved armory design: `docs/superpowers/specs/2026-08-29-svelte-armory-design.md`.
Approved implementation plans: `docs/superpowers/plans/2026-08-29-tbc-upgrade-finder.md`,
`docs/superpowers/plans/2026-08-29-svelte-armory.md`, and the
`docs/superpowers/plans/2026-08-31-*.md` follow-up plans.

## Current status

The Svelte armory migration is complete and committed. The local upgrade finder imports an individual-sim link, renders a server-enriched 17-slot armory review with deterministic base-plus-gear stats, runs the existing ranking flow, displays the report, supports clipboard copy, and supports cancellation.

The 2026-08-31 follow-up work is implemented and verified:

- The root README is a user-first landing page for the Upgrade Finder with upstream attribution.
- `.github/workflows/upgrade_finder.yml` runs the Go, build, UI, and Chromium verification contract on pushes/PRs.
- `Import` accepts current individual-sim exports that omit the optional simulation-settings message; the import API defaults the maximum phase to `5` when the export has no positive phase.

## Working behavior

### Backend

- `Import` accepts individual-sim links and rejects malformed, raid-sim, incompatible-version, and unsupported equipment references before enrichment or ranking.
- Missing optional simulation settings are accepted: nil-safe protobuf getters yield phase `0`, iterations `0`, and non-fixed seed. Exports with a positive phase retain it.
- Unsupported nonzero item, random-suffix, gem, and enchant-effect IDs return typed `incompatible_*` validation errors. Main equipment and enabled item-swap records use the same validation path; empty item specs remain ordinary empties.
- `Catalog` indexes items, gems, enchant effects, zones, NPCs, and random suffixes from the bundled `UIDatabase`.
- `EnrichArmory` emits all 17 canonical slots in display order: Head, Neck, Shoulder, Back, Chest, Wrist, Main Hand, Off Hand, Hands, Waist, Legs, Feet, Finger 1, Finger 2, Trinket 1, Trinket 2, Ranged.
- Armory metadata includes item names/quality/icons/phases/set names, scaled random suffixes, gems, enchants, declared sockets, socket bonuses, and empty-slot/empty-socket state.
- Socket matching follows `core.ColorIntersects`, including hybrid, prismatic, and meta behavior. Suffix scaling follows `core.ItemEquipmentBaseStats`.
- Armory totals come from `core.ComputeStats` on a cloned setup with buffs, consumes, talents, bonus stats, and item swaps cleared. It does not run simulation iterations and is labeled **unbuffed (base + gear)**.
- Raw stat keys use unprefixed snake_case `stats.StatName()` values, including `mp5`. Derived keys are limited to `melee_hit_percent`, `spell_hit_percent`, `melee_crit_percent`, `spell_crit_percent`, `ranged_hit_percent`, `ranged_crit_percent`, and `block_percent`.

### HTTP server

- `POST /api/import` preserves existing successful fields and adds only `gear`, `stats`, and `derivedStats`.
- The import response defaults `maxPhase` to `5` when the imported phase is less than `1`; exports with a positive phase retain it.
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

### Repo and CI

- `.github/workflows/upgrade_finder.yml` is independent of the upstream simulator's reusable `run_tests.yml`: one Ubuntu job installs Go `1.25.x` and Node `22`, runs `go test ./cmd/wowsimcli/cmd/upgrades -count=1`, `go test ./cmd/wowsimcli/cmd/... -count=1`, `go build`, `npm ci`, `npm run build`, installs Chromium, and runs `npm run test:e2e`.

## Verification

Final required checks passed (2026-09-05):

```text
go test ./cmd/wowsimcli/cmd/upgrades -count=1    ok
go test ./cmd/wowsimcli/cmd -count=1             ok
go build -o wowsimcli ./cmd/wowsimcli             passed
cd ui-finder && npm run build                     Vite 7.3.6; 117 modules transformed; passed
cd ui-finder && npm run test:e2e                  1 Chromium test passed (9.0s)
rtk node --test src/lib/api.test.js               2 passed
```

Real-application import check: the rebuilt binary was started with
`rank-upgrades --addr 127.0.0.1:43199 --no-browser`, the retribution fixture
link was posted to `/api/import`, and the response carried class
`ClassPaladin`, spec `RetributionPaladin`, `defaults.maxPhase` `5`, and all 17
gear slots.

## Intentional boundaries

- No 3D model, talents tab, item hover tooltips, buffed-stat mode, persistence, pricing, accounts, remote server, or ranking-domain rewrite.
- The browser exposes the existing max-phase, unknown-source, screening, and confirmation controls. It does not search alternate gem/enchant policies or source-name filters.
- The armory snapshot is not a simulation result and excludes temporary buffs, consumes, talents, and alternate item-swap gear.

## Known follow-up risks

- The CI workflow triggers are pinned to `master`, while the `origin` remote's only branch is `main`; the workflow has not yet run on GitHub. Resolve the branch/trigger mismatch before relying on CI (tracked as open item).
- Simulator database attachment remains process-wide through the existing `sync.Once` path; concurrent first jobs should be serialized or attachment should become per-job-safe.
- DELETE-failure recovery can briefly schedule duplicate polling timers while an already-fired poll request is in flight. Cancellation remains recoverable; this was recorded as a non-blocking review note.
- The optional favicon remains a cosmetic 404.
