# Project State — TBC Upgrade Finder

Updated: 2026-09-06

Authoritative operator instructions: `docs/upgrade-finder.md`.
Approved armory design: `docs/superpowers/specs/2026-08-29-svelte-armory-design.md`.
Approved implementation plans: `docs/superpowers/plans/2026-08-29-tbc-upgrade-finder.md`,
`docs/superpowers/plans/2026-08-29-svelte-armory.md`, and the
`docs/superpowers/plans/2026-08-31-*.md` follow-up plans.

## Current status

The Svelte armory migration is complete and committed. The local upgrade finder imports an individual-sim link, renders a server-enriched 17-slot armory review with a deterministic full-settings stat snapshot (the imported buffs, consumes, and talents, matching what ranking simulates), runs the existing ranking flow, displays the report, supports clipboard copy, and supports cancellation.

The same-day buffed-stat parity follow-up is implemented: the armory stat panels now show the fully raid-buffed stats from the link's settings instead of the unbuffed base-plus-gear snapshot, mirroring the wowsims site's stats panel.

The 2026-08-31 follow-up work is implemented and verified:

- The root README is a user-first landing page for the Upgrade Finder with upstream attribution.
- `.github/workflows/upgrade_finder.yml` runs the Go, build, UI, and Chromium verification contract on pushes/PRs.
- `Import` accepts current individual-sim exports that omit the optional simulation-settings message; the import API derives a bounded maximum phase from the simulated phase, then the highest equipped-item phase, and falls back to `5` only when neither is positive.

The armory redesign is implemented: character-centered layout with Gear/Stats/Talents tabs, WoWSims-style item cards (ilvl badge, rarity-colored names, enchant effect lines, sockets/gems), a labeled 3D placeholder stage with per-spec backdrop art, and a read-only talents view driven by the newly exposed talents string.

The equipped 3D character preview is implemented per
`docs/superpowers/plans/2026-09-06-equipped-character-preview.md`: the
feasibility experiment (milestone 1) succeeded — a correctly equipped TBC
character renders from a loopback app origin through a same-origin asset
transport with the ZAM/Wowhead live viewer + TBC content path. The mapping
resolver, ZAM adapter, stage integration and tests are in place, but the
integration is **disabled by default** (`--enable-3d` opt-in): the provider's
viewer/model assets are not a public SDK and their usage requires an
authorized arrangement the operator must confirm. Full account:
`docs/superpowers/experiments/2026-09-06-equipped-character-feasibility.md`.

## Working behavior

### Backend

- `Import` accepts individual-sim links and rejects malformed, raid-sim, incompatible-version, and unsupported equipment references before enrichment or ranking.
- Missing optional simulation settings are accepted: nil-safe protobuf getters yield phase `0`, iterations `0`, and non-fixed seed. Exports with a positive phase retain it.
- Unsupported nonzero item, random-suffix, gem, and enchant-effect IDs return typed `incompatible_*` validation errors. Main equipment and enabled item-swap records use the same validation path; empty item specs remain ordinary empties.
- `Catalog` indexes items, gems, enchant effects, zones, NPCs, and random suffixes from the bundled `UIDatabase`.
- `EnrichArmory` emits all 17 canonical slots in display order: Head, Neck, Shoulder, Back, Chest, Wrist, Main Hand, Off Hand, Hands, Waist, Legs, Feet, Finger 1, Finger 2, Trinket 1, Trinket 2, Ranged.
- Armory metadata includes item names/quality/icons/phases/set names, scaled random suffixes, gems, enchants, declared sockets, socket bonuses, and empty-slot/empty-socket state.
- Socket matching follows `core.ColorIntersects`, including hybrid, prismatic, and meta behavior. Suffix scaling follows `core.ItemEquipmentBaseStats`.
- Armory totals come from `core.ComputeStats` on the imported baseline as-is via the same raid construction ranking uses (`ImportedSettings.raidAndEncounter`): raid, party, and individual buffs, consumes, talents, and bonus stats are retained. The displayed panel reads the engine's `finalStats` snapshot and adds the site's target-debuff contributions (Improved Faerie Fire hit, Improved Seal of the Crusader crit, Hunter's Mark, Expose Weakness) exactly like the wowsims site's stats panel. It does not run simulation iterations and is labeled **raid buffed (link settings)**.
- Raw stat keys use unprefixed snake_case `stats.StatName()` values, including `mp5`. Derived keys are limited to `melee_hit_percent`, `spell_hit_percent`, `melee_crit_percent`, `spell_crit_percent`, `ranged_hit_percent`, `ranged_crit_percent`, and `block_percent`.

### HTTP server

- `POST /api/import` preserves existing successful fields and adds only `gear`, `stats`, and `derivedStats`.
- The import response defaults `maxPhase` to the exported phase when positive, otherwise to the highest phase among equipped items, and falls back to `5` only if neither is available.
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
- UI iteration uses the Vite dev server with HMR (`ui-finder/` `npm run dev`, proxying `/api` to the Go server on port `43123`); the embedded bundle only updates on `npm run build` plus Go restart. Workflow: `docs/upgrade-finder.md` "UI development (hot reload)".

### Repo and CI

- `.github/workflows/upgrade_finder.yml` is a standalone upgrade-finder workflow: one Ubuntu job installs Go `1.25.x`, Protoc, `protoc-gen-go`, and Node `22`; generates the gitignored proto stubs with `make sim/core/proto/api.pb.go binary_dist/dist.go`; runs `go test ./cmd/wowsimcli/cmd/upgrades -count=1`, `go test ./cmd/wowsimcli/cmd/... -count=1`, `go build`, `npm ci`, `npm run build`; installs Chromium; and runs `npm run test:e2e`. Triggers on pushes and pull requests targeting `main`.
- The inherited upstream simulator workflows (`run_tests.yml`, `deploy.yml`, `release.yml`, `scheduled_release.yml`) were removed from the fork; they target upstream publishing infrastructure and cannot pass on a forked repository.

## Verification

Current checkout verification (2026-09-06): Go upgrade tests, CLI package tests,
Go build, the Vite production build, all UI unit tests, and the Chromium E2E
suite passed. The 3D-preview additions added 9 mapping unit tests, 9 Go
resolver/route tests, and 2 E2E tests (default-build unavailable state and
fake-adapter activation flow). Real-provider smoke (this checkout): the
embedded binary with `--enable-3d` rendered the imported Retribution paladin
in Chromium through the same-origin proxy (57 `/visuals/zam/` asset requests,
1 resolve request, zero errors), with the expected "Partial preview: Ranged
(no visible model)" for the relic, and verified rotate, pause/resume, gear-tab
cleanup, and re-import remount behavior.

Earlier milestone checks (2026-09-05):

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
`ClassPaladin`, spec `RetributionPaladin`, `defaults.maxPhase` `2` (highest equipped-item phase), and all 17
gear slots.

## Intentional boundaries

- The 3D character preview is opt-in via `--enable-3d` (off by default; the
  provider arrangement is an operator gate). The stage still never blocks
  import, ranking, or the assumptions fingerprint; preview settings stay
  local. Exact in-game cosmetics (face, hair matching a live character),
  shirt/tabard inputs, and ranked-upgrade visual previews remain future work.
- The viewer's own `setAnimPaused` is a no-op on the current distribution, so
  the stage pause/visibility control destroys and lazily recreates the viewer
  (HTTP-cached assets); there is no full renderer pause API.
- No interactive 3D model by default, talent editing, persistence, pricing,
  accounts, remote server, or ranking-domain rewrite. Item hover/focus
  tooltips are implemented and replace the former per-slot Details
  disclosure.
- The browser exposes the existing max-phase, unknown-source, screening, and
  confirmation controls. It does not search alternate gem/enchant policies or
  source-name filters.
- The armory snapshot is not a simulation result; it shows the imported buffs,
  consumes, and talents but never runs iterations or applies fight-time item
  swaps.

## Known follow-up risks

- The 3D provider integration is authorized-arrangement-gated: flipping
  `--enable-3d` on requires confirming the ZAM/Wowhead usage arrangement
  before production use; the transport and a request-origin inventory are
  documented in the feasibility account.
- The repo workflows were retargeted from `master` to `main` to match the `origin` remote's default branch; the Upgrade Finder CI validated end to end after the proto-generation fix (2026-09-05).
- Simulator database attachment remains process-wide through the existing `sync.Once` path; concurrent first jobs should be serialized or attachment should become per-job-safe.
- DELETE-failure recovery can briefly schedule duplicate polling timers while an already-fired poll request is in flight. Cancellation remains recoverable; this was recorded as a non-blocking review note.
- The optional favicon remains a cosmetic 404.
- The unresolved-detail-consent/performance measurement items from the plan (slow network, offline mode, WebGL-unavailable handling, frame/load measurements) were verified only as far as the real-provider smoke; the remaining measurements belong before shipping the integration.
