# TBC Upgrade Finder — Implementation Handoff

## Assignment

Implement the approved TBC upgrade finder end-to-end. The authoritative implementation plan is:

`docs/superpowers/plans/2026-08-29-tbc-upgrade-finder.md`

Read that plan in full before editing. Execute its eight tasks in order, preserve their interfaces, and check each checkbox only after its listed verification passes.

The approved product design is:

`docs/superpowers/specs/2026-08-29-tbc-upgrade-finder-design.md`

When a design detail and an implementation detail conflict, follow the design. If a required upstream API differs from the plan's expected shape at the pinned revision, adapt the implementation to the upstream API without weakening the behavior or widening scope.

## Starting State

- This repository currently has planning/specification material only; it is not yet an application.
- Seed the implementation from `https://github.com/wowsims/tbc-new` at commit `88fb853466a391e731e12de012f6707a11e94446`.
- Preserve the existing `docs/superpowers/specs/` and `docs/superpowers/plans/` files while establishing the upstream source base.
- The upstream database blob revision is `84555ed6e3ddf19edca22204892a2366e1a177da`.
- The resulting application is one local Go binary: `wowsimcli rank-upgrades`.

## Non-Negotiable Product Rules

- Accept wowsims **individual-sim** export links only. Validate malformed, raid, incompatible-settings, and unknown-item links before starting simulations.
- Use the pinned wowsims simulator and bundled item database as the only scoring/data authority.
- Rank only practical single-item DPS replacements under the imported configuration. No profile import, scheduler, MCP, pricing, persistence, static weights, or multi-item optimizer.
- Copy imported settings for every candidate. Tests must demonstrate that the imported baseline remains byte-identical.
- Default to excluding items without source metadata; surface their count.
- Enforce class, faction, profession, phase, equipment, ring, hand/weapon, and unique-equipped constraints before simulation.
- Apply only the submitted legal gem/enchant policy. Do not silently substitute or optimize alternatives.
- Use screening and confirmation simulation budgets. Confirm the bounded finalist set; mark overlapping 95% intervals as `tooCloseToCall` rather than pretending they have a precise order.
- Isolate candidate simulation failures. Cancellation must leave no partial report presented as final.
- Bind only to loopback, retain jobs in memory only, and return no cached character data.
- Make the wowsims attribution link visibly present in the browser UI and include simulator/database revisions plus an assumptions fingerprint in every report.
- Add no runtime dependency: use the upstream dependencies and Go standard library.

## Implementation Sequence

1. Import and pin upstream, then register `rank-upgrades` in the existing Cobra CLI.
2. Build the side-effect-free import boundary and canonical `RaidSimRequest` conversion.
3. Build catalog indexing and legal candidate construction before adding any simulation behavior.
4. Add gem/enchant policy validation and policy application.
5. Add the simulator adapter, two-pass ranking, uncertainty, progress, failure isolation, cancellation, and fingerprinting.
6. Add loopback HTTP job endpoints and embedded assets.
7. Build the browser workflow against those endpoints only; do not put simulator logic in JavaScript.
8. Complete the contract matrix, documentation, and real browser smoke test.

Keep the domain module at `cmd/wowsimcli/cmd/upgrades/`. Keep `upgrade_server.go` as HTTP/job orchestration only. Do not leak protobuf simulator channels, request IDs, candidate construction steps, or per-iteration metrics into the browser API.

## Required Evidence Before Handoff Back

Run and report these exact checks after the implementation is complete:

```bash
rtk go test ./cmd/wowsimcli/cmd/upgrades -count=1
rtk go test ./cmd/wowsimcli/cmd/... -count=1
rtk go build -o wowsimcli ./cmd/wowsimcli
```

Then run the built binary with `rank-upgrades --no-browser`, use the fixed individual-link fixture and small known source filter, and browser-drive the local page. Evidence must confirm:

1. imported character/settings summary appears before ranking;
2. progress advances while the real simulator runs;
3. the completed report is non-empty;
4. each result exposes source, policy configuration, DPS delta, interval, iterations, and assumptions;
5. report provenance (simulator revision, database revision, fingerprint) and wowsims attribution are visible;
6. copying produces parseable JSON;
7. canceling a fresh job produces no partial report.

## Completion Report Format

Return only:

- changed paths and their responsibility;
- test/build commands with observed results;
- browser-smoke evidence;
- unresolved upstream/API discrepancy, if one remains.

Do not return a partial implementation, a scaffold, or a list of follow-up tasks.
