# Project README design

## Goal

Make the repository's root README identify the local TBC Upgrade Finder before describing the bundled WoW TBC simulator. It must give users enough information to run the tool and find authoritative detail without duplicating `docs/upgrade-finder.md`.

## Audience

Users running the local browser application. Contributor material remains linked rather than expanded.

## README structure

1. **Title and purpose** — TBC Upgrade Finder ranks practical single-item DPS upgrades for a character using its imported wowsims individual-sim configuration.
2. **Quick start** — Run `rtk go run ./cmd/wowsimcli rank-upgrades`; explain that it opens a loopback-only local browser application.
3. **Use it** — Paste an accepted individual-sim export link, review the generated 17-slot armory, select filters and iteration budgets, then rank and inspect the report.
4. **Behavior and boundaries** — State that results are based on pinned wowsims simulator/item data; rankings cover single-item DPS upgrades only; character data and reports stay in memory for the process and are not persisted remotely.
5. **Documentation** — Link to the operator guide and the existing installation/development/add-a-sim/i18n documentation.
6. **Upstream attribution** — Credit `wowsims/tbc-new`, retain the upstream live-sims and support links, and reference `UPSTREAM.md` for the pinned revisions.

## Non-goals

- Do not duplicate the detailed validation, armory, result, or smoke-test material from `docs/upgrade-finder.md`.
- Do not alter command behavior, simulator code, packaging, or linked documentation.
- Do not add an additional dependency or README-specific tooling.

## Verification

Review the rendered Markdown structure and verify all local links resolve to the intended tracked files. No runtime behavior changes.
