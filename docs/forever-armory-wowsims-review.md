# Forever armory versus WoWSims Forever

Reviewed: 2026-10-20. Research and design review only; no application implementation or design approval implied.

## Verdict

**Keep the hosted armory goal, but change the starting point: WoWSims Forever is now a substantial technical upstream, not merely a simulator lead.** It already supplies a working level-60 gear/talent editor, catalogs, rule implementations, native snapshot/file interchange, and active simulations for some specs. Building those from an empty data model would duplicate considerable work.

**It does not replace the proposed armory.** Our distinguishing requirements are character ownership, Current versus complete independent plans, durable hosted saves, publication controls, and one permanent read-only URL showing the latest saved published builds. The inspected upstream implementation is spec-centered, browser-local, and shares editable snapshots.

The earlier catalog verdict needs a dated reassessment, not erasure: an inspectable technical candidate now exists, but complete coverage, strict validation, immutable catalog binding, lower-level support, and the intended hosted-content reuse scope are not established merely by the upstream site's existence.

## Review scope and evidence boundaries

- Local branch: `docs/forever-armory-design`, HEAD `8c570a11e`.
- Comparison baseline: local `main`, `ceff87d41`; inspected `git log main..HEAD` and `git diff main...HEAD` file inventory. This is a branch-capability comparison, not a review of a nonexistent Forever implementation diff.
- Upstream: [`wowsims/forever` at `0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4`](https://github.com/wowsims/forever/tree/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4). Inspected a separate shallow checkout outside this project. The commit updates items for client build `1.60.1.70450` with October 20 hotfixes.
- Live surfaces: [landing page](https://www.wowsims.com/forever/), [Retribution Paladin](https://www.wowsims.com/forever/paladin/retribution/), and [Holy Paladin](https://www.wowsims.com/forever/paladin/holy/), exercised in managed Chromium without user account access.
- Source and deployment are separate evidence: no claim that the live deployment exactly matches the pinned commit. Runtime observations are identified below; other implementation claims refer to that commit.
- Existing unrelated dirty assets, state/preview documentation, icon-generator changes and `.pi/` were not changed. No branch checkout, merge, commit, push, provider contact, account operation, or client extraction was performed.
- No upstream build, test suite, or numerical/game-parity audit was run. Existing tests and workflow declarations were inspected, not counted as passing verification.

## 1. What our branch actually has

The branch adds six commits and five files relative to `main`, with no application-code changes:

| Artifact | Delivered | Still open |
| --- | --- | --- |
| [Armory specification](superpowers/specs/2026-10-01-forever-armory-design.md) | Product/domain requirements, publication and ownership semantics, gates and acceptance criteria | Written spec status still says awaiting user review; no manual-editor or hosted-armory implementation plan |
| [Feasibility plan](plans/2026-10-01-forever-data-import-feasibility.md) | Scoped data/import research procedure | This is not an implementation plan for the complete armory |
| [Research report](forever-data-import-research.md) | October 2 source/licence investigation, Sixty export shape, addon/talent evidence and explicit limits | Conclusions need the new upstream evidence in this review applied before implementation planning |
| [Experiment log](superpowers/experiments/2026-10-01-forever-data-import-feasibility.md) | Reproducible research recipes, recorded provider UI observations and provenance | Does not implement our validators, imports, storage, or permissions |
| [Numeric talent projection](superpowers/experiments/forever-fixtures/catalog-sample.json) | Attributed 17-record structural Warrior sample | Not a complete item catalog, production all-class talent dataset, or stable identity scheme |

The inherited application is useful UI infrastructure, not a partly implemented Forever armory:

- Svelte 5/Vite frontend, character header, equipment cards/tooltips, Gear/Stats/Talents views.
- A TBC-only import/enrichment/ranking backend, with transient in-memory jobs and reports.
- Read-only imported talents using TBC tree JSON; no item-selection editor, independent build model, character library, ownership, publishing, or durable character storage in these paths.
- Existing 3D work is separate, gated work and explicitly deferred by the Forever design. It should not become a dependency for this armory.

Evidence: local [`ArmoryView.svelte:8–25,39–76`](../ui-finder/src/lib/ArmoryView.svelte), [`TalentTrees.svelte:6–25,45–52`](../ui-finder/src/lib/TalentTrees.svelte), [`stores.svelte.js:4–11,43–49`](../ui-finder/src/lib/stores.svelte.js), [`upgrade_server.go:57–77,133–174`](../cmd/wowsimcli/cmd/upgrade_server.go), and [`upgrades/import.go:13–16,104–200`](../cmd/wowsimcli/cmd/upgrades/import.go).

## 2. Requirement-by-requirement comparison

“Available upstream” below means observed source capability unless marked live. It does not mean the capability has been integrated here or independently verified against the game.

| Our design requirement | Our implementation today | Upstream provides | Remaining armory work |
| --- | --- | --- | --- |
| Hosted browser application, no local server for readers | Local TBC application | Public static site; browser WASM, plus optional Go simulator server | Deploy our UI and durable character/ownership service |
| Character-centered visual gear view | Reusable read-only Svelte view | Working slot editor, item cards and stat sidebar; live exercised | Preserve our character UX; add editing against Forever data |
| Forever equipment, enchants and tooltips | No production Forever catalog | Main/leftover catalogs, typed slots, effects, names/icons and external Wowhead presentation | Curated/versioned projection, content-use assessment, coverage and restriction checks |
| Item search and source filters | No manual picker | Text search, slot/compatibility, phase, source, faction, item-level and weapon filters | Adapt model/UI; distinguish missing or inherited source metadata |
| Multiple characters, including same class/spec | No library | One active player per spec page and spec-keyed browser state | Stable character IDs, identity fields and character library |
| Current actual build versus plans | No model | One working simulator configuration and presets | Current role, import/edit provenance, initial empty Current, explicit replacement confirmation |
| Complete independent named plans | No model | Separate named saved gear and saved talents | A plan must own equipment, enchants and talents together; stable build IDs and copy/rename/delete semantics |
| Independent talent editing | Imported TBC talents read-only | Full three-tree editor, add-point guards and broken-build highlighting; live preset change exercised | Version-bound identity, complete legality checks and build ownership |
| Level 60 default with basic lower-level planning | TBC-derived presentation | Fixed level-60 player and fixed talent budget; some low-level items | Store selected level, retain item requirements, dynamic budgets and level-aware capabilities; do not silently narrow design |
| Verified whole-character stats only | TBC engine totals | Forever engine-derived stats and extensive formulas, with known gaps | Coverage/parity decision per supported combination; explicit unavailable/partial state elsewhere |
| Manual character creation | Import-only TBC entry | Manual gearing within a spec/race simulator | Name/race/class/level creation flow and validated persistent builds |
| Addon actual-character import | No Forever importer | Gear/race/profession parser; third-party talent strings intentionally not applied | Real Forever fixture/client/addon evidence, identity/level mapping, strict validation and atomic Current update |
| Sixty planned-build import | Shape research only | Pasted/uploaded JSON gear/race parser, but drops talents and suffixes | Explicit slots, talent crosswalk, source/version decision, all-or-nothing validation and creation of a separate plan |
| Sixty URL import | Not implemented | No provider-supported machine-export contract found in inspected importer | Remains independently gated; do not scrape based on a visible page |
| Guest editing remembered in browser | No ownership model | Browser-local settings and saved records | Server credentials separate from public IDs, secure session storage and authorization |
| Private recovery link, rotation/revocation | None | Not found in inspected ownership/server paths | Protected credential exchange, recovery/revocation flows and loss warning |
| Optional account and guest claim | None | Not found in inspected ownership/server paths | Maintained authentication integration, proof of guest access and atomic claim |
| Permanent URL showing latest saved builds | None | URL-fragment snapshot; live verified | Durable character lookup whose published projection updates at the same identifier |
| Read-only public access; selected published builds | None | An editable simulator after opening a snapshot | Server-enforced public projection, unpublished-by-default, build switching limited to published content |
| Unpublish/delete, keep URL stable across renames/claim | None | No corresponding hosted record lifecycle found | Implement all specified transitions; never send private builds or credentials to public clients |
| Export/import backup | Reports only, not character backups | JSON for current simulator categories and native snapshot links | Versioned whole-character/all-plan portable files, validation, no auth secrets or transferred ownership |
| Durable restart/restore | In-memory jobs | Stateless simulation service; static site and browser caches | Persistent character database, backups and proven restoration |
| Keyboard/touch; confirmations; failed-save retention | Some existing tooltip/focus interactions | Several picker/talent interactions | Exercise our complete flows; save errors and destructive confirmations are not supplied by local presets |
| Simulation, rankings, optimizer | Existing TBC feature; deferred for Forever | Active Alpha simulations, batch and optimization facilities | No need to add these to the armory scope; optional future interoperability is a separate decision |

Sources: [upstream UI architecture](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/README.md), [saved gear](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/features/gear/components/SavedGear/utils.ts#L10-L33), [saved-record storage](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/ui-kit/hooks/useSavedData.ts#L25-L72), [persistence](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/sim/state/persistence.ts#L48-L109), [import finalization](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/features/import-export/importers/finish_individual_import.ts#L19-L63), [Sixty parser](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/features/import-export/importers/sixty_upgrades.tsx#L23-L105), [server](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/sim/web/main.go#L100-L160), and [stateless deployment](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/docker-compose.yml#L11-L38).

## 3. What data and game support now exist

### Catalog inventory

Counts are read-only aggregations of the pinned payloads, not claims of game completeness or observed passing tests.

| Dataset / coverage | Observed |
| --- | --- |
| Main catalog | 8,599 equipment rows; 246 enchant rows, representing 213 distinct enchant effect IDs |
| Leftover catalog | 1,186 equipment rows; 20 enchant rows |
| Talent structures | 465 talents, nine classes, 27 trees, 72 emitted prerequisite locations |
| Main catalog source coverage | 2,119 items with source entries; 6,480 without |
| Main item-level range and phase | Item levels 1–100; all main items phase 1 |
| Statistical slots | 17; no modeled shirt/tabard slots |
| Playable race capability matrix | Ten choices including High Order and Windshaper Skyborne; retained Blood Elf enum is not playable capability evidence |

“Leftover” is the generator's non-simmable partition, not proof those items are unavailable in game. Browser lookup can load it for missing imports; a normal picker/catalog export is not necessarily the union. Source filtering supports more categories than the committed data actually populates. All-phase-1 and inherited expansion labels are not safe client-version/game discriminators.

Sources: [main catalog](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/assets/database/db.json), [leftovers](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/assets/database/leftover_db.json), [schema](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/proto/ui.proto#L14-L109), [filtering](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/tools/database/gen_db/main.go#L477-L539), [capabilities](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/sim/core/character_constants.go#L111-L185), and [talent files](https://github.com/wowsims/forever/tree/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/sim/talents/trees).

### Concrete crosswalk with our earlier Sixty examples

| Earlier example | Current upstream evidence | Remaining issue |
| --- | --- | --- |
| Item `260212` | Bad Hat; head/cloth; item level 56 | This is a concrete item-ID match, not proof of required character level or complete item coverage |
| Enchant `8219`, spell `1249058` | Enchant Necklace – Healing Power; effect ID 8219, item ID 249503, spell ID 1249058; neck | Effect, spell and carrier-item IDs are distinct domains; effect ID alone does not uniquely identify all catalog enchant rows |
| Talent `105848`, spell `14523`, rank 3 | Silent Resolve has spell ID 14523, definition ID 135378, maximum rank 3 | The Sixty ID is not that definition ID; its domain/crosswalk is still unresolved |

Sources: [Bad Hat](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/assets/database/db.json#L6686), [enchant](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/assets/database/db.json#L8993), [Silent Resolve](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/sim/talents/trees/priest.json#L40-L48), and [enchant identity warning](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/proto/ui.proto#L187-L223).

### Versioning: useful provenance, not a durable save contract

Decoded the committed gzip-formatted `assets/db_inputs/spell_store_inputs.json` in memory. All captured groups identify build `1.60.1.70450`, but their hotfix pushes differ:

| Captured groups | Hotfix push |
| --- | --- |
| Items, Mage, shared | 113005 |
| Druid, Hunter, Paladin, Priest, racials, Rogue, Shaman, Warlock, Warrior | 112887 |

This illustrates why a single build string is insufficient. Partial group updates are intentional. The public `UIDatabase` and talent JSON themselves contain no immutable catalog revision. Native proto/API version is separate again.

Observed payload hashes:

```text
assets/database/db.json
  sha256 9fa8335e7e97f824373ad8225b6c2d82e1a2915b0e71a6f05e2312add442e9e3
assets/database/leftover_db.json
  sha256 eeb3ff7788226b9739e470c27d8a9a4035dbae5b5ad5f5d781483eeef0ce4685
```

A durable armory save must bind to an immutable catalog/talent snapshot and define migration or unsupported-reference handling. Never reinterpret an old positional talent string against reordered live trees. Our old research's 467-talent provider snapshot is not this 465-talent dataset.

Sources: [captured inputs](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/assets/db_inputs/spell_store_inputs.json), [partial-update contract](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/docs/spell_data/partial-updates.md#L65-L90), [talent codec](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/sim/talents/talents_string.ts#L1-L30), and [API version policy](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/proto/common.proto#L8-L26).

### Levels, validation and statistics

These are material mismatches, not reasons to discard the upstream:

- **Fixed level:** player level and talent budget are tied to 60. Required character level is used while extracting eligible items but is not carried into the published equipment records. Item level and enchant `lvl` must not be mistaken for a character/application-level requirement.
- **Equipment conflicts:** useful class/slot/weapon/enchant rules exist, but setters automatically clear conflicting weapons/unique items. Our existing-incompatible-choice requirement needs preserve-and-flag semantics rather than copying these setters blindly. Item race restrictions and category-count semantics need further checks.
- **Talent legality:** add-point guards enforce tier/prerequisite/budget rules; removing points may deliberately create broken talents, which remain allocated and are still applied by the simulator. Positional decoding clamps ranks and converts malformed characters rather than serving as a strict import validator.
- **Trait fidelity:** generation simplifies client traits into three grids, selects one choice entry and one incoming prerequisite, and drops certain duplicate candidates. Client-derived does not establish full TraitCond/choice/multi-parent fidelity.
- **Stats:** substantial Forever-specific base-stat/conversion work exists, including client-derived data and manually observed race offsets. Some vanilla base-crit values, manual attributes, MoP-era tooltip scaling, unsupported effects and estimated proc rates remain. Engine output is a useful candidate/reference, not automatic independent parity proof.

Sources: [fixed mechanics](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/sim/constants/mechanics.ts#L8-L9), [item extraction](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/tools/database/tables.go#L93-L160), [gear conflicts](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/sim/proto/gear.ts#L81-L130), [talent point guard](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/features/talents/model/can_set_points.ts#L1-L29), [trait extraction](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/tools/database/tables.go#L1462-L1790), [base-stat parser](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/tools/base_stats_parser.py#L8-L48), and [base-stat caveats](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/sim/core/base_stats.go#L15-L245).

## 4. Important review findings

### High: existing TBC importer is not a safe Forever adapter

**Reproduced against current local source**, not inferred solely from a prebuilt binary:

```text
go run ./cmd/wowsimcli rank-upgrades --addr 127.0.0.1:43272 --no-browser
POST /api/import with existing TBC fixture -> 200, 17 gear slots
POST /api/import with actual live Forever Ret snapshot -> ECONNRESET
server: panic, nil pointer dereference
  NewRetributionPaladin -> ComputeStats -> EnrichArmory -> handleImport
```

The HTTP server recovers the request panic; this was not an observed whole-process exit. The captured Forever export had API version 2 and shared legacy item 12640. Our current proto version is 14; an upper-bound check accepts the lower number. Item presence and compatible top-level wire tags do not establish identical spec/game semantics.

Decision: keep Forever import/catalog/rules outside the existing TBC `Import → EnrichArmory → ComputeStats` path. A future integration needs explicit game/protocol discrimination, validated conversion and no accidental TBC calculation. This review does not implement a panic catch or otherwise mask the failure.

Local anchors: [`import.go:118–200`](../cmd/wowsimcli/cmd/upgrades/import.go), [`proto/common.proto:18–30`](../proto/common.proto), [`upgrade_server.go:239–253`](../cmd/wowsimcli/cmd/upgrade_server.go), and [`armory.go:250–264`](../cmd/wowsimcli/cmd/upgrades/armory.go).

### High: import menus overstate talent compatibility if read literally

The live Sixty/Add-on dialogs say they import talents. The pinned implementation does not:

- Shared third-party finalization applies race/gear/professions but never sets imported talents.
- Sixty counts supplied talent records, then passes an empty talent string and warns afterward.
- Sixty also drops suffixes; unknown items are skipped, and an unknown enchant skips the entire item. The remaining gear can replace the working gear before missing-ID feedback.
- Native JSON/hash import is different and can restore native talent strings.

The Sixty source comment claims ranks are unavailable, but our earlier user-provided payload contains `talents[].rank`. The source comment is not proof of that format limitation; the implementation's omission is the supported finding.

Do not copy permissive partial application into our atomic import contract. Parse → resolve IDs/versions → validate → present diagnostics/confirmation → apply to the intended build, with previous data unchanged on failure.

Sources: [shared finalization](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/features/import-export/importers/finish_individual_import.ts#L19-L63), [Sixty parser](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/features/import-export/importers/sixty_upgrades.tsx#L47-L103), [native JSON](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/features/import-export/importers/json.ts), and [our real nonempty talent-format evidence](forever-data-import-research.md#user-provided-nonempty-talent-export).

### High: stable public records and complete plans are still ours to build

`SavedGearSet` holds equipment and bonus stats; `SavedTalents` holds a talent string. Named browser records are keyed by display name, not stable build identity. Loading gear does not load talents. Live smoke confirmed this separation.

Link export serializes selected current settings, compresses them and puts them in the URL fragment. It does not save or identify a server character. Opening the link imports into an editable local sim, clears the hash and can preserve local settings for categories omitted from the link. It is valuable interoperability, not an ownership or publication layer.

Sources: [save schemas](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/proto/ui.proto#L347-L417), [link encoder](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/features/import-export/exporters/link.ts#L13-L35), and [load order](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/sim/state/persistence.ts#L48-L109).

### Medium: a stateless simulator server is not an armory backend

Upstream does have a Go backend, static `UIDatabase`, IndexedDB caches/history and SQLite-based generation tooling. “No backend/database exists” would be inaccurate. What was not found in the inspected UI state/features, schemas, proto, simulation server and deployment paths is character ownership CRUD, account sessions, guest recovery/claim, per-build publication, durable public character records or character backup/restore.

Also, `schemas/character.schema.json` describes localized character strings, not an armory aggregate; `talents.schema.json` is not an allocation validator. These names should not drive architecture assumptions.

Sources: [server routes/state](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/sim/web/main.go#L100-L160), [stateless deployment](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/docker-compose.yml#L11-L38), and [schemas](https://github.com/wowsims/forever/tree/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/schemas).

## 5. Reuse and architecture decision

The upstream frontend is React 19/Zustand with React-specific table, virtualization and UI dependencies; ours is Svelte 5. Its TSX components are not drop-in Svelte components. Pure models/data/codecs are more plausible reuse seams than `Player`/`Sim` wholesale.

| Approach | Benefit | Cost / mismatch | Recommendation |
| --- | --- | --- | --- |
| Hosted Svelte armory with pinned upstream data/model adapters | Preserves approved UX/domain and existing UI investment; keeps simulator maintenance upstream | We implement editing surfaces and hosted behavior; selected pure rules still need adaptation and validation | **Recommended for the stated product** |
| Fork upstream gearplanner and add hosted ownership/builds | Maximum initial editor reuse; upstream already has explicit gearplanner mode | React product fork, extensive app/state/persistence changes, simulator-oriented envelope and update/merge burden; changes existing Svelte-reuse direction | Viable alternative only after an explicit design change |
| Thin saved-profile viewer around imported WoWSims snapshots | Smallest new editor surface; lets WoWSims own editing | Does not satisfy approved manual editing, complete plan management or reliable field coverage | Scope reduction, not an equivalent implementation |

Retain `Character → Current / named Plans → equipment + enchants + talents` as our domain. Give it stable IDs, explicit source/catalog revisions and separate ownership/publication state. Treat simulator payloads as external interchange, never as the authoritative character/ownership schema.

Concrete upstream adaptation candidates:

- `ui/features/gear/model/item_search.ts`: normalization and source matching.
- `ui/sim/proto/items.ts`: slot, class, weapon and enchant compatibility candidates.
- `ui/sim/player/classes/capabilities_auto_gen.ts`: maintained capability inputs, subject to level/game verification.
- `ui/sim/talents/{config,talents_string,broken_talents}.ts` and `ui/features/talents/model/{can_set_points,tree_graph}.ts`: structural data, graph/layout and rule candidates; add strict full-allocation validation and selected-level support.
- `ui/features/import-export/importers/`: field/ID mapping references, but replace partial-apply finalization with our contract.
- Native link/JSON exporters: optional WoWSims interoperability, not our public URL or backup format.

Do not automatically port ranking, optimizer, rotation, raid buffs, 3D or combat simulation into the armory. Their existence does not change the approved exclusions. Whole-character stats can remain explicitly unavailable pending coverage evidence.

Sources: [dependencies](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/package.json#L31-L61), [search helper](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/features/gear/model/item_search.ts#L8-L51), [compatibility helpers](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/sim/proto/items.ts), and [talent model](https://github.com/wowsims/forever/tree/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/sim/talents).

## 6. Software licence, data provenance and reuse risk

The root MIT licence expressly supports software reuse with notice retention; README requests a visible upstream link. That is a credible reuse route, not a reason to rebuild all software from scratch.

The actual pipeline reads beta client/CDN tables and hotfix caches, merges overrides and AtlasLoot-derived acquisition metadata, and uses external icons/artwork. Its third-party notices distinguish software licences from definition-data licences. No blanket catalog-content clearance document was identified in this snapshot. The notices' statement about no TACT keys also differs from the extractor README's key/decryption instructions, so do not treat notices as a complete current extraction-policy account.

Decision still required: document the intended hosted reuse of factual/numeric/structural data separately from game text, names, icons/artwork, inherited source tables and provider-hosted assets. Preserve licence/attribution obligations for reused code and licensed contributions. MIT is neither automatic clearance for everything nor evidence of a categorical prohibition on the armory. No legal advice or provider permission is asserted here.

Sources: [MIT](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/LICENSE), [README](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/README.md#L7-L9), [extractor notices](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/tools/db2tool/NOTICES.md), [extractor inputs](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/tools/db2tool/README.md#L22-L74), [generator configuration](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/tools/database/generator-settings.json), and [AtlasLoot source](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/tools/database/atlasloot.go#L15-L24).

## 7. Upstream availability and open work

Landing-page links are not proof all destinations have active simulations. Pinned source declares:

| Status | Entries |
| --- | --- |
| Alpha simulator: 12 | Balance Druid, Feral Cat, Hunter, Mage, Protection/Ret Paladin, DPS Priest, Rogue, Elemental/Enhancement Shaman, Warlock, DPS Warrior |
| GearPlanner: 4 | Restoration Druid, Holy Paladin, Healer Priest, Restoration Shaman |
| Unlaunched: 2 | Feral Bear, Protection Warrior; disabled outside development mode |

Live Holy Paladin explicitly says gear planning only, gear/enchants/talents/optimizer work, and no simulation is planned. Live Ret completed a small simulation. Neither observation proves combat/stat parity.

Source: [spec registry](https://github.com/wowsims/forever/tree/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/sim/player/specs), [launch gate](https://github.com/wowsims/forever/blob/0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4/ui/app/individual_sim_ui.tsx#L118-L122), and [live Holy planner](https://www.wowsims.com/forever/paladin/holy/).

Relevant open tracker items inspected on the review date; reports/PR descriptions are not independent proof and unmerged work is not credited as delivered:

| Item | Implication |
| --- | --- |
| [PR #556: dev-only Armory import](https://github.com/wowsims/forever/pull/556) | Open, not in reviewed master/live importer. Local Cloudflare Worker proposal; real Forever namespace/name mapping/responses unresolved in its description. Does not establish a production Blizzard API route or hosted armory |
| [PR #555: several talent setups in batch](https://github.com/wowsims/forever/pull/555) | Open batch feature, not a complete independently owned/published build model |
| [Issue #57: unmodeled enchant shapes](https://github.com/wowsims/forever/issues/57) | Catalog presence must not imply complete modeled enchant effects |
| [Issue #58: stat-weight gaps](https://github.com/wowsims/forever/issues/58) | Distinguish optimizer/EP coverage from verified armory statistics |
| [Issue #107: Gnome Eureka validation](https://github.com/wowsims/forever/issues/107) | Some racial interactions still request real-game validation |
| [PR #607: Wisepaw speed correction](https://github.com/wowsims/forever/pull/607) | Open proposed correction reports client-versus-game disagreement; client extraction alone is not complete runtime truth |
| [PR #502: initial healing simulation](https://github.com/wowsims/forever/pull/502) | Open work must not be treated as current launched healer simulation |

The official exporter repository inspected still lists Era/SoD/TBC/Wrath/Cata/Mists detection and no explicit Forever branch/interface. That does not prove it cannot run under a shared client flag, but it also does not supply a real compatible equipped-character fixture. The live Addon button is parser availability, not runtime validation.

Sources: [exporter compatibility](https://github.com/wowsims/exporter/blob/master/Shared.lua), [TOC](https://github.com/wowsims/exporter/blob/master/WowSimsExporter.toc), and [open upstream tracker](https://github.com/wowsims/forever/issues). Exporter links are mutable observations, unlike the pinned Forever source links.

## 8. Gate reassessment and concrete next work

| Gate | New assessment | What must happen next |
| --- | --- | --- |
| A: catalog/rules | **Substantial technical candidate demonstrated; partial, not passed** | Adopt scoped content-use policy; pin manifest with commit/hash/group provenance; audit representative items, enchant domains, trait fidelity and restrictions; obtain required-level fields for promised lower-level support |
| B: native WoWSims interchange | **New useful candidate**, not previously a specified entry route | Decide whether to add it to the design; implement a distinct game/schema adapter, catalog binding and atomic validation; preserve native talents only with compatible versioned trees |
| B: addon | **Parser exists; full compatible character import not established** | Obtain consented real output with addon/client versions; verify identity, level, slots, enchants and talent mapping; do not apply foreign positional trees |
| B: Sixty text/file | **Gear parser and representative ID matches improve feasibility; complete import still partial** | Resolve talent/slot/version crosswalk and payload-use scope, use explicit ranks where valid, reject unsupported data before changing a build |
| B: Sixty URL | **Still gated** | Provider-supported machine-export contract; no invented scraper/API |
| C: ownership/sharing | **Not supplied by upstream and not implemented locally** | Separate hosted plan for durable characters, authorization, guest recovery/revocation, account claim, publication projection and restart/restore |
| Statistics | **Useful Forever implementation exists; independent coverage still unverified** | Define supported combinations and prove parity; keep unverified totals unavailable rather than reusing TBC or silently summing items |

Recommended order, without authorizing implementation:

1. Review/revise the written armory specification with these findings. Keep the core hosted goal and exclusions; explicitly record the selected upstream-reuse boundary. A native WoWSims entry route is a proposed addition, not already approved scope.
2. Complete a narrow upstream-adoption/data contract: immutable manifest, rights/provenance decision, normalized IDs/slots, strict import/validation behavior and required lower-level fields. Do not repeat a broad search as though no usable catalog candidate exists.
3. Write the manual-editor implementation plan using our complete build model, Svelte shell and selected upstream data/model adaptations. Keep independently gated import routes unavailable until their checks pass.
4. Write and execute the hosted ownership/publication plan. It can be designed against the domain contract without pretending catalogs solve authentication; a browser-only prototype does not complete the product.
5. Attach verified importers and, only when independently supported, whole-character stats. Do not make simulation/3D adoption a prerequisite.

Acceptance criteria still needing implementation include every plan-isolation transition, selected-level talent validity, invalid-import rollback, failed-save retention, cross-owner denial, public-link mutation denial, credential revocation/claim, private-field exclusion, stable URL across edits/renames/claim, unpublish/delete behavior, secret-free portable backups, keyboard/touch flows and durable restart/restore. Their absence is expected in a design branch, not evidence that research delivered a broken implementation.

## 9. Exercised verification and limits

| Scenario | Observed result |
| --- | --- |
| Live landing page | Nine class groups and public spec links; broad Alpha labeling does not distinguish every route's launch mode |
| Ret equipment editing | Opened Head picker, searched Lionheart, equipped item 12640; displayed item and stat changes |
| Named gear save | Saved `Review A`; entry and spec-scoped `__savedGear__` storage survived reload |
| Gear versus talents | Saved gear while at 16/0/35, changed preset to 23/0/28, loaded saved gear; talents stayed 23/0/28 |
| Snapshot sharing | Exported link, opened in another tab, removed original working helmet, reopened the exact captured link in a fresh tab; helmet returned from snapshot. Destination remained editable and hash was consumed |
| Link reload nuance | Reloading the consumed `#` address restored browser state, not the captured payload; re-adding hash to the same loaded page did not trigger import in this probe. Fresh document load of original link did |
| Import menus/dialogs | Opened Sixty and Addon dialogs; recorded text and options. Did not claim successful full-character/talent import |
| Native JSON | Parsed live export in memory: API version 2, equipment ID 12640 and a native talent string; no explicit nondefault player-level field in observed payload |
| Live Ret simulation | Completed a 10-iteration run and displayed a DPS/TPS result; smoke only, not a numerical accuracy result |
| Holy route | Explicit Gear Planner message and no simulation controls |
| Current local source | `go run` launched; existing TBC fixture imported with 200 and 17 slots; actual Forever snapshot caused recovered request panic/connection reset |
| Current embedded local UI | Imported TBC fixture and observed Gear/Stats/Talents, updated identity header and `Imported build / Read-only` talent presentation |
| Pinned data | Counted catalogs; computed hashes; decoded captured per-group build/hotfix provenance in memory |

An initial smoke also used a preexisting executable whose UI was older than the source. It was not used as current-branch UI evidence; the current-source run above supersedes it. Browser automation encountered modal-transition/overlay timing failures before completing interactions; these are not classified as product defects.

Managed tabs and review-launched local services were closed afterward. No real addon/client export was generated, no accounts tested, no provider authorizations obtained, no upstream tests run, and no full game/stat parity or all-class usability is claimed. Product code and the existing historical design/research files are unchanged; this report is the review deliverable.
