# Svelte Armory — Implementation Handoff

## Assignment

Implement the approved Svelte armory migration and character-review view end-to-end. The authoritative implementation plan is:

`docs/superpowers/plans/2026-08-29-svelte-armory.md`

Read that plan in full before editing. Execute its six tasks in order, preserve the declared interfaces, and check a task checkbox only after its listed verification passes.

The authoritative product design is:

`docs/superpowers/specs/2026-08-29-svelte-armory-design.md`

When a design detail and an implementation detail conflict, follow the design. When the simulator API differs from an expected implementation detail, adapt to the actual upstream API without weakening behavior, changing the ranking domain, or widening scope.

## Starting State

- The existing upgrade finder is a local Go binary with vanilla embedded assets in `cmd/wowsimcli/cmd/upgrade_ui/`.
- The existing import flow accepts an individual-sim link, returns a decoded summary, and passes the original link to the unchanged ranking job API.
- `cmd/wowsimcli/cmd/upgrades/` already owns import, catalog, policy, candidates, and ranking behavior. Keep armory domain logic there.
- `cmd/wowsimcli/cmd/upgrade_server.go` owns HTTP/job orchestration only. Do not put simulator calculations in handlers or browser code.
- The browser build source does not yet exist. Create it under `ui-finder/`; only generated output belongs under `cmd/wowsimcli/cmd/upgrade_ui/`.

## Non-Negotiable Product Rules

- Use Svelte 5 runes. Module-shared rune state is `ui-finder/src/lib/stores.svelte.js`.
- Commit generated Vite output in `cmd/wowsimcli/cmd/upgrade_ui/`. `rtk go build` and all Go tests must remain Node-free.
- Vite uses `base: './'`, emits normal nested `assets/` files, and resolves `outDir` to `../cmd/wowsimcli/cmd/upgrade_ui/` from `ui-finder`.
- Serve Vite output only through `GET /assets/{path...}`. Validate the wildcard with `fs.ValidPath`, map it below embedded `upgrade_ui/assets/`, preserve structured 404s, and never expose a directory listing or arbitrary embedded file.
- Keep `GET /`, all `/api/*` routes, loopback-only binding, 500 ms job polling, job cancellation behavior, report fields, and visible wowsims attribution intact.
- `POST /api/import` gains only successful-response fields `gear`, `stats`, and `derivedStats`.
- Before enrichment or ranking, reject every unsupported nonzero item, random-suffix, gem, and enchant-effect ID with a typed `incompatible_*` validation error. Do not render unsupported data as a placeholder or let it reach engine equipment construction.
- `Gear` contains every one of the 17 canonical slots in specified display order. Zero equipment IDs and zero gem IDs render ordinary empty slots/sockets.
- Resolve item, gem, enchant, and suffix metadata from the bundled `UIDatabase`; index suffixes in `Catalog`.
- Apply suffix stats with the same scaling as `core.ItemEquipmentBaseStats`. Determine socket-bonus activation with `core.ColorIntersects`; hybrids, prismatic gems, and meta gems must match engine behavior.
- Generate the armory total with `core.ComputeStats` on a clone with raid/party/individual buffs, consumes, and talents cleared. Expose the player's gear-stage stats. This runs no simulation iterations and is visibly labeled **unbuffed (base + gear)**.
- Raw stat keys are unprefixed snake_case `stats.StatName()` values. Derived percentage keys are exactly `melee_hit_percent`, `spell_hit_percent`, `melee_crit_percent`, `spell_crit_percent`, `ranged_hit_percent`, `ranged_crit_percent`, and `block_percent`.
- The browser renders server-provided armory data only; it performs no item lookup, stat conversion, socket rule, or simulator logic.
- Add no Go runtime dependency. Frontend build/test dependencies are isolated and locked under `ui-finder/package-lock.json`.

## Implementation Sequence

1. Extend catalog indexes and import validation for suffix, gem, and enchant-effect references.
2. Add `armory.go` with canonical slot enrichment, suffix/socket metadata, and sanitized engine stat snapshots plus focused Go contracts.
3. Add armory fields to the import response and replace one-segment asset serving with validated nested asset serving.
4. Create the isolated Svelte/Vite project and generate the committed embed bundle.
5. Port import, armory rendering, controls, polling, report, copying, and cancellation into focused Svelte components.
6. Add real-server Playwright smoke coverage and update operator documentation.

Keep cross-task contracts stable: `Catalog`, `ArmoryData`, `GearSlotData`, `EnrichArmory`, and the `gear`/`stats`/`derivedStats` response fields. Do not introduce a second data authority, a profile importer, a 3D model, tooltip system, buffs/talents UI, persistence, pricing, or ranking changes.

## Required Evidence Before Handoff Back

Run and report these exact checks after implementation:

```bash
rtk go test ./cmd/wowsimcli/cmd/upgrades -count=1
rtk go test ./cmd/wowsimcli/cmd/... -count=1
rtk go build -o wowsimcli ./cmd/wowsimcli
```

Then, from `ui-finder`:

```bash
rtk npm ci
rtk npm run build
rtk npm run test:e2e
```

Browser evidence must exercise the real local process and confirm:

1. importing the fixed individual-link fixture renders character details and all 17 armory slots;
2. gems, enchants, socket-bonus state, raw stats, derived stats, and the wowsims attribution are visible;
3. ranking shows real progress, then a report and working copy action;
4. canceling a fresh job clears the report and announces cancellation;
5. generated nested hashed assets load through the Go server.

## Completion Report Format

Return only:

- changed paths and their responsibility;
- test/build commands with observed results;
- browser-smoke evidence;
- unresolved simulator/API discrepancy, if one remains.

Do not return a partial implementation, scaffold, or follow-up list.
