# TBC upgrade finder design

## Status

Approved 2026-08-29. This document defines the first implementation slice only.

## Goal

Provide a locally run browser application that ranks the largest practical **single-item DPS upgrades** for a TBC character under the exact assumptions of an imported wowsims configuration.

The application starts from a wowsims export link, not character-profile data. Armory import, scheduled monitoring, MCP, pricing, and multi-item optimization are future work.

## Decisions

| Area | Decision | Reason |
| --- | --- | --- |
| Product surface | Local web application | Lets the user inspect assumptions, available content, and recommendations without accounts or hosting. |
| Simulation authority | The wowsims TBC engine | Item value depends on the full simulation configuration; an LLM and static stat weights cannot produce authoritative DPS deltas. |
| Configuration input | Exported wowsims individual-sim link | It includes the simulation-critical configuration that Armory cannot supply: talents, rotation, buffs, encounter, consumes, professions, and equipment. |
| Candidate scope | User-selected content | Recommendations must be obtainable, not merely best-in-slot in the entire database. |
| Candidate setup | Item plus explicit gem/enchant policy | A bare dropped item systematically underrates socketed and enchantable upgrades. |
| Ranking scope | One replacement at a time, DPS metric | Keeps the first optimizer understandable and verifies actual simulated outcomes. |
| Runtime | One local Go binary with a browser UI | The simulator is Go; no network service, credentials, accounts, or database are necessary. |

## Non-goals

- Character Armory import.
- Daily or event-driven monitoring.
- MCP transport or natural-language agent orchestration.
- Auction-house or acquisition-cost analysis.
- Multi-slot best-in-slot/loadout optimization.
- Tank survivability, threat, and healer-throughput recommendations.
- A separate static item-score or stat-weight calculator.

## Architecture

```text
Browser UI
  -> localhost Go server
       -> import and validation
       -> candidate construction and filtering
       -> ranking runner
            -> wowsims TBC engine and bundled item database
       -> result formatter
```

The server owns the simulation call and returns structured results. The UI only collects inputs and displays the report. No model participates in scoring.

The first implementation should use the upstream Go code directly, ideally by extending the existing command-line package with a `rank-upgrades` operation rather than starting a second simulation implementation. The repository already provides:

- a decoder from wowsims export links to `IndividualSimSettings` or `RaidSimSettings`;
- a runner that accepts a ProtoJSON `RaidSimRequest`;
- an item database whose item records include class restrictions, professions, phase, sources, weapon/hand constraints, gems, enchants, and unique-equipped constraints.

Pin the upstream revision and bundled database version in every run result. The public surface must visibly credit and link to wowsims as requested by its MIT-licensed repository README.

## Interfaces

### `RankUpgrades`

The deep module exposed to the UI:

```text
RankUpgrades(importedSettings, contentFilters, itemPolicy, simulationOptions)
  -> UpgradeReport | ValidationError
```

The interface guarantees that each reported candidate is compared with the same imported baseline and that all report assumptions are returned with results.

It does not expose simulator internals, candidate generation steps, process management, or per-iteration metrics to the UI.

### Input

**Imported settings**

A validated wowsims export link. Its decoded settings are the source of truth for player setup, rotation, raid/party context, encounter, buffs, consumes, professions, equipment, and user-selected simulation settings.

**Content filters**

The selectable item-source and availability constraints: expansion phase plus raids, dungeons, crafting, reputation, PvP, vendors, quests, and optionally source names. Candidates with absent source data are excluded by default and surfaced as an explicit `unknown source` count.

**Item policy**

A user-visible policy defining the legal gem choices by socket colour and gem quality, and standard enchants by eligible equipment slot. The report records the exact policy used. V1 does not claim to independently discover the globally optimal gem/enchant combination for every candidate.

**Simulation options**

DPS is the only v1 ranking metric. The user selects a practical iteration budget. The ranking runner uses a screening budget for the full candidate pool and a confirmation budget for finalists.

### Output

`UpgradeReport` contains:

- baseline DPS and simulation metadata;
- a ranked list of confirmed candidates;
- for each candidate: item, target slot, displaced item, source/phase, applied gems and enchant, DPS delta, relative gain, confidence/uncertainty, iterations, and material assumptions;
- candidates excluded for compatibility or source reasons, summarized without claiming they were evaluated;
- an assumptions fingerprint covering imported settings, filters, policy, simulation options, simulator revision, and database revision.

## Candidate construction

For each item in the bundled database:

1. Apply content filters.
2. Validate class, faction, profession, phase, item type, armor proficiency, hand/weapon requirements, and equip limits.
3. Determine every legal replacement slot. Rings are evaluated in both ring slots. Two-hand/off-hand and main-hand conflicts are handled as equipment changes, not assumed away.
4. Create a fresh copy of the imported settings and replace only the candidate equipment position.
5. Apply the declared gem/enchant policy only where legal.
6. Run the actual simulator against the unchanged encounter, rotation, and raid assumptions.

Set-bonus loss, cap effects, proc behavior, and rotational thresholds are intentionally evaluated by the simulation result, not approximated from item stats.

## Ranking and uncertainty

A candidate is an estimated DPS delta against the baseline, not an absolute item rating.

1. Simulate the baseline and all legal candidates at the screening budget.
2. Retain a bounded top set, including candidates statistically close to the cutoff.
3. Re-run that set at the confirmation budget.
4. Sort confirmed deltas and mark overlaps as `too close to call` when the chosen uncertainty interval overlaps.

The report must not display a precise order for statistically indistinguishable candidates. Failed candidate simulations are isolated, recorded with the failure reason, and omitted from the ranked list; they must not invalidate results from successful candidates.

## UI flow

1. Paste a wowsims individual-sim export link.
2. Inspect decoded character and settings summary before running.
3. Select available content and the gem/enchant policy.
4. Start ranking and view progress.
5. Read top upgrades with source, configuration, delta, uncertainty, and assumptions.
6. Copy the exact assumptions/result data for later comparison.

The UI does not need account management, saved character records, or remote persistence in v1. Local browser storage is optional convenience only; a pasted link remains sufficient.

## Error handling

- Invalid or unsupported links produce a specific import error and no simulation runs.
- A settings version or item ID unsupported by the pinned simulator revision is reported as incompatible.
- Candidate construction never mutates the baseline configuration.
- Unknown source metadata prevents a candidate from being recommended by default.
- A single failed candidate run is listed separately and does not abort the ranking job.
- Cancellation stops outstanding work and returns no partial ranking as final/confirmed.

## Verification

### Contract tests

- Decode a fixed exported sim link into the expected settings.
- A known filter produces an expected candidate ID set.
- Candidate construction preserves the original baseline settings byte-for-byte.
- Invalid class, profession, unique-equipped, ring, weapon, and source combinations are rejected.
- The configured gem/enchant policy is applied only where legal.
- Every result includes its assumptions fingerprint and simulator/database revisions.

### Smoke test

Run the local application, paste a fixed wowsims link, select a small known content set, and confirm the browser displays a non-empty ranked report with the expected source and assumptions fields.

## Deferred extensions

### Armory importer

Add only after confirming that Blizzard's current profile endpoints cover the relevant Classic character fields. It converts profile equipment, gems, enchants, and identity into simulator equipment. It cannot infer rotation, raid composition, encounter, buffs, or consumes; the user must still import or choose those settings.

Do not create a profile-source seam now. When Armory import exists, add the converter at the input edge alongside export-link decoding.

### Monitoring

Persist an accepted import snapshot, filter configuration, and item policy. A scheduler reruns the same `RankUpgrades` request and notifies only for a materially significant confirmed new result.

### MCP

After the deterministic backend is stable, expose it through a thin MCP tool:

```text
rank_upgrades(sim_link, content_filters, item_policy)
```

The AI client may explain or refine a request but must use the ranking result as the source of truth.

## Sources

- https://github.com/wowsims/tbc-new
- https://github.com/wowsims/tbc-new/blob/master/cmd/wowsimcli/cmd/decode_link.go
- https://github.com/wowsims/tbc-new/blob/master/cmd/wowsimcli/cmd/basic_sim.go
- https://github.com/wowsims/tbc-new/blob/master/proto/ui.proto
