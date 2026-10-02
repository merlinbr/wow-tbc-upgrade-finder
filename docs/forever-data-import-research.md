# Forever armory: data and import feasibility

Checked 2026-10-02 UTC. Research only; no application changes, authenticated Blizzard calls, client extraction, provider messages, or proprietary catalog copies committed.

## Question

Can the approved hosted Forever armory obtain permitted, versioned item/talent data and import actual or planned builds without depending on a Blizzard character API?

**Final research verdict:** talent structure and user-facing talent/gear export formats have useful verified evidence. The user also supplied a nonempty Sixty talent export and a friend-reported addon-generated talent link that renders correctly. **Do not start the full real-data editor yet:** a permitted, sufficiently complete versioned item catalog, identifier/version mapping and full gear-addon runtime evidence remain unresolved. Research is complete; production readiness is not.

Experiment procedures and outcomes: [`superpowers/experiments/2026-10-01-forever-data-import-feasibility.md`](superpowers/experiments/2026-10-01-forever-data-import-feasibility.md). The filename follows the approved plan date; observations were made October 2.

## Sources and permissions

All rows below were checked October 2, 2026. `partial` means useful evidence, not production authorization. This is a published-permission assessment, not jurisdiction-specific legal advice.

| Source / owner | Observed version | Access and coverage | Local inspection / hosted reuse verdict |
| --- | --- | --- | --- |
| [Talents Forever / Chris Baldwin](https://talentsforever.com/about), [terms](https://talentsforever.com/terms), [JSON](https://talentsforever.com/data.json) | JSON generated `2026-10-01`; homepage identifies beta `1.60.1.70170` | HTTP 200; nine class keys; talents, spellbooks, racials, class abilities and Legacy | **partial**: own work of reading/checking/arranging data is CC BY 4.0, allowing copying/adaptation/commercial use with a credit link. Spell names, tooltip text, icons and artwork are expressly excluded. Numeric/structural evidence can be retained with attribution; do not redistribute the whole mixed-content file as if wholly licensed. |
| [Talents Forever licence deed](https://creativecommons.org/licenses/by/4.0/) / Creative Commons | CC BY 4.0 International | Readable licence terms | Supports sharing/adaptation with appropriate credit, licence link and indication of changes; does not grant rights the licensor does not hold. |
| [wowforever.space sources](https://wowforever.space/sources/), [versions](https://wowforever.space/data-updates/) / site operator | Fixed `demo-2026-09-15-5a452c300ec1`, nine classes / 27 trees / 470 talents | Readable, attributed adaptation of Talents Forever | **partial**: clearly labels pre-beta footage snapshot, unknown ranks and lack of independent footage verification. It is not the current upstream beta snapshot. Blizzard visual/text material is excluded from the community licence. Prefer original provider for new work. |
| [Forever Codex sources](https://forever-codex.com/sources/), [about](https://forever-codex.com/about/) / site operator | Beta `1.60.1.70170`; generation date October 2 | HTTP 200; client/hotfix-derived item/talent information; some Classic-derived acquisition sources | **blocked for our hosted catalog**: no express site-wide data redistribution grant identified in the inspected pages. Source credits identify Blizzard ownership; MIT reader software and GPL Classic world data do not license the entire Forever item collection. No catalog retained. |
| [ElliotWood/Forever](https://github.com/ElliotWood/Forever), pinned [LICENSE](https://raw.githubusercontent.com/ElliotWood/Forever/1e0474d94dc0ce7b38b4bf15f09416d20f797688/LICENSE), [data sources](https://raw.githubusercontent.com/ElliotWood/Forever/1e0474d94dc0ce7b38b4bf15f09416d20f797688/docs/DATA_SOURCES.md) / repository contributors | `master` at `1e0474d94dc0ce7b38b4bf15f09416d20f797688`; build marker `1.60.1.70178` | GitHub metadata and selected raw files HTTP 200; no full clone | **partial for code, blocked for catalog reuse**: MIT explicitly covers software and associated documentation, with notice retention. README requests a visible upstream link. Data provenance includes Wowhead scraping/client tables and inherited data. No separate permission resolving third-party catalog rights was established. Do not assume code licensing or the repository's availability authorizes rehosting all inputs. |
| [Sixty Upgrades Forever terms](https://sixtyupgrades.com/forever/terms) / Sixty Upgrades | No displayed terms revision observed; Forever UI inspected October 2 | HTTP 200 application shell; rendered terms obtained in a fresh unsigned-in browser | **blocked for mirroring catalog/materials**: §2 grants temporary personal, non-commercial transitory viewing and forbids copying, public display, transferring/mirroring and reverse engineering site software. User export interoperability needs a separate scope decision; these terms are not an identified hosted-data licence. |
| [Wowhead Forever](https://www.wowhead.com/forever) / Wowhead | No usable build/version verified | One direct request with an identifying research User-Agent returned HTML HTTP 200; readable extractor reported dynamic rendering. A later ordinary urllib request returned 403 and was not bypassed | **blocked**: no catalog export or applicable redistribution grant verified. Guessed `/terms-of-service` returned 404; no positive rights conclusion from that failure. |
| [Sixty Upgrades Exporter project](https://www.curseforge.com/wow/addons/sixty-upgrades-exporter) / publisher not independently verified here | No version/client support independently verified | HTTP 403 from direct and content-fetch requests; did not bypass | **blocked**: search summaries about Forever support/license remain leads, not primary evidence. |
| [Blizzard Developer API Terms](https://www.blizzard.com/en-us/legal/a2989b50-5f16-43b1-abec-2ae17cc09dd6/blizzard-developer-api-terms-of-use) / Blizzard | Live terms page inspected October 2 | Readable public legal page; no registered-app/API calls | **conditional route, no Forever data verified**: §2 allows registered applications to distribute API data to end users for personal use under restrictions. §§2.13, 2.15 and 2.18 require attribution/non-endorsement, a privacy policy and maximum 30-day retention/refresh; §2.16 restricts extraction outside authorized API use. This does not license client extraction or third-party catalogs. |
| [Blizzard addon policy](https://eu.forums.blizzard.com/en/wow/t/wow-user-interface-add-on-development-policy/1642) / Blizzard | Published policy including 2009 updates; inspected October 2 | Readable official forum post | **partial**: recognizes addons under conditions including free distribution, visible code, no in-game ads, and ToU/EULA compliance. It does not establish the format, Forever runtime support, or rights in a third-party addon. |

### Rights distinctions

- Viewing a public site is not permission to mirror its dataset.
- A community licence can license the author's arrangement/verification work without licensing Blizzard text or art.
- Game-file provenance and an open-source extractor are not a redistribution grant.
- User-selected identifiers and allocations need not require copying a provider's whole catalog; exported payload content and applicable terms must be checked before implementing that route.
- An API-backed approach has its own registration, retention, privacy, branding and deletion requirements. No specific Forever namespace was confirmed; a dynamic developer-page fetch does not prove absence of one.

### Contact routes, not inquiries sent

[Talents Forever contact](https://talentsforever.com/contact) provides an ideas message box and Reddit/X contact; its terms already permit reuse of its own contribution with attribution. Sixty Upgrades' rendered footer exposes a Discord community entry point; confirm an authorized owner/contact before requesting export/cross-site reuse clarification. Forever Codex's about page says its public reporting channel is not open yet; no verified licensing contact was found there. The Blizzard addon policy identifies `WoWUI@blizzard.com` for addon-policy questions, not general catalog redistribution grants. No messages were sent.

## Catalog and rules

### Talent structure: usable evidence, not a complete game-data clearance

The current [upstream JSON](https://talentsforever.com/data.json) has **467 class talents across nine classes and 27 trees**, not the mirror's old 470-record snapshot. Class tree counts observed directly:

| Class | Talents per tree |
| --- | --- |
| Warrior | 17 / 18 / 18 |
| Paladin | 17 / 16 / 17 |
| Hunter | 16 / 16 / 18 |
| Rogue | 17 / 17 / 19 |
| Priest | 18 / 17 / 18 |
| Shaman | 16 / 18 / 16 |
| Mage | 18 / 17 / 19 |
| Warlock | 17 / 19 / 16 |
| Druid | 16 / 20 / 16 |

Observed paths: `talents.<Class>.trees[].talents[]`; `row`/`col` are one-based, `max` is the cap, `req` refers to a same-tree talent name, and `desc` contains rank text. Auditing all 467 records found positive integer caps, no duplicate grid positions, 71 resolvable prerequisite references, and no `complete`/description-count gaps. Those are source-consistency checks, not independent game verification. The export's explanatory README still mentions estimated/unconfirmed ranks; do not use that wording as proof every current field is independently confirmed.

No canonical talent/spell identifier was present in the class-talent record keys. Our [17-record numeric projection](superpowers/experiments/forever-fixtures/catalog-sample.json) uses clearly labeled **snapshot locators**, not persistent game IDs. A production planner needs an explicit versioned identity/migration policy; never reinterpret an old allocation using a reordered live array. Race data contains five entries per faction and `classes` fields, but race identity, faction variants and full compatibility were not independently client-tested.

The [Warrior calculator](https://talentsforever.com/warrior) documents rows opening every five tree points and prerequisites requiring maximum rank. Its UI was exercised in isolated browser contexts: **8/8 observations passed**, covering level-60/30 budgets, rank ceilings, row unlocking, prerequisite rejection/acceptance, and refusing a second point at level 10. These verify provider UI behavior, not our nonexistent validator or a game-server rule. Exact cases and a runnable recipe are in the experiment log.

The retained projection includes only numeric/structural fields, author attribution, licence link and the changes made; it omits spell names, tooltip text, icons and art. This grants no broader clearance for a production name/tooltip/icon presentation.

### Equipment, source labels and statistics: blocked

No candidate passed both permitted hosted reuse and usable versioned equipment acquisition. Consequently armor, weapon/off-hand conflicts, ring/trinket legality, enchants, equipment slot inventory, and lower-level item restrictions could not be fixture-tested. Their acceptance cases remain requirements, not passing observations. The existing app's 17 statistical slots and TBC enchant IDs are not assumed valid for Forever.

[Forever Codex's source page](https://forever-codex.com/sources/) says 8,189 of 31,821 item IDs lack data, mostly level-50–60 equipment; its Classic-derived loot/vendor data is expressly not guaranteed Forever-equivalent. These provider-reported limits directly matter to our level-60-first scope. Do not turn Classic loot rates into Forever source filters.

No complete base/racial/talent/level formula set was independently verified. Item descriptions, provider stat displays and a simulator fork are not sufficient proof of whole-character stat parity. Full character totals and derived percentages remain blocked; talent budgets have narrower UI evidence only.

Known version pitfall: the upstream describes client-derived beta data, wowforever.space retains a September 15 demo snapshot, and the repository build marker `70178` differs from the `70170` sites. Never combine these silently.

## Addon import

### Existing exporter source: inspected, Forever runtime unverified

The candidate repository is [wowsims/exporter](https://github.com/wowsims/exporter), not the guessed `wowsims/WoWSimsExporter` (404). Inspected commit **`ad9e903e3997b99f66a2122a4a91b555d40c4ece`**, whose [MIT licence](https://github.com/wowsims/exporter/blob/ad9e903e3997b99f66a2122a4a91b555d40c4ece/LICENSE) allows software reuse with notices. Selected code was read from a temporary tool-managed clone, not copied into this project.

[Shared.lua](https://github.com/wowsims/exporter/blob/ad9e903e3997b99f66a2122a4a91b555d40c4ece/Shared.lua) recognizes Era/SoD/TBC/Wrath/Cata/Mists. The [.toc](https://github.com/wowsims/exporter/blob/ad9e903e3997b99f66a2122a4a91b555d40c4ece/WowSimsExporter.toc) lists interfaces `50504, 50503, 40402, 38001, 20505, 11508`, with no explicit Forever interface/branch. This does not prove it cannot run in Forever under an existing project flag; it means no dedicated support or runtime correctness was verified.

| Target | Source path / encoding from inspected code | Representative source behavior / missing-field limit |
| --- | --- | --- |
| Name | `name` | `UnitFullName`; no real exported value obtained |
| Race / class | `race`, `class` | Race from game GUID info; class lowercased; new faction-specific race handling not verified |
| Level | `level` | `UnitLevel`; no real Forever output obtained |
| Realm / region | `realm`; no region field in Character structure | Optional realm from `UnitFullName`; region is absent, not inferred |
| Gear | `gear.items[index].id` | EquipmentSpec's fixed 17-slot order; preserve empty-slot positions rather than compacting arrays |
| Enchant | `gear.items[index].enchant` | Numeric enchant component parsed from the item link; actual Forever identifier semantics remain unverified |
| Talents | `talents` | Hyphen-separated positional rank digits; `extras.lua` orders old `GetTalentInfo` entries by tier/column; Forever API compatibility and versioned tree mapping unverified |
| Version | `version` | Addon metadata version, **not a game/client build**; no explicit Forever catalog revision in the Character structure |

Mapping evidence: [Character.lua](https://github.com/wowsims/exporter/blob/ad9e903e3997b99f66a2122a4a91b555d40c4ece/ExportStructures/Character.lua), [EquipmentSpec.lua](https://github.com/wowsims/exporter/blob/ad9e903e3997b99f66a2122a4a91b555d40c4ece/ExportStructures/EquipmentSpec.lua), [ItemSpec.lua](https://github.com/wowsims/exporter/blob/ad9e903e3997b99f66a2122a4a91b555d40c4ece/ExportStructures/ItemSpec.lua), [Misc.lua](https://github.com/wowsims/exporter/blob/ad9e903e3997b99f66a2122a4a91b555d40c4ece/ExportStructures/Misc.lua), [extras.lua](https://github.com/wowsims/exporter/blob/ad9e903e3997b99f66a2122a4a91b555d40c4ece/extras.lua), and [output encoding](https://github.com/wowsims/exporter/blob/ad9e903e3997b99f66a2122a4a91b555d40c4ece/WowSimsExporter.lua). This is a code-derived format, not a real export fixture.

The Sixty Upgrades Exporter CurseForge project still returns 403; its actual licence/client support could not be independently inspected. [Talents Forever's actual addon page](https://talentsforever.com/ideas) documents beta installation and talent-plan link sharing, but no complete gear-export format was verified there. Guessed `/addon` returned 404; that was not evidence the addon does not exist.

### Friend-reported Talents Forever addon sharing

The user supplied [this Shaman build](https://talentsforever.com/shaman/60/050433002-054032031005102251--RUVXYS1cbefSghBEFDI-5) and explicitly confirmed that their beta-playing friend generated it **from the addon**, not the website. Record this as **user-reported runtime evidence**, not our own in-game test.

Independently opening that link in a fresh browser showed level 60, **17 Elemental / 34 Enhancement / 0 Restoration**, 51 displayed allocated ranks, zero remaining points after the counters settled, and a leveling-order section. The URL's extra suffix was not decoded or assigned a guessed meaning. Early reads of animated counters briefly showed nonzero remaining points; a condition-based wait reached the consistent final state.

This demonstrates a site-compatible talent-plan artifact and strengthens the talent-only addon path. It does **not** prove that the friend's character is level 60, that the plan equals currently learned talents, that a specific addon version/API was tested by us, or that equipped gear/enchants/race/realm can be exported.

**Gate B full character addon remains blocked; talent-only sharing is partial.** A real equipped-character export plus addon version and `GetBuildInfo()` output is an optional future input, not required to complete this research handoff. No new addon was written or installed.

## Sixty Upgrades import

### User-facing JSON export: technically verified

The [supplied public set](https://sixtyupgrades.com/forever/set/vH9HQmj6tzxi7fajuHmu8p) rendered HTTP 200 in an isolated unsigned-in browser. Its **Export** action opened JSON with Copy to Clipboard and Download controls. No site software was reverse-engineered, private editor accessed, or endpoint invented.

The JSON decoded successfully in memory and contained **17 unique named statistical equipment slots**. The sample was a level-60 Priest/Dwarf planned set. A single observed export is not a general equipment ruleset or proof of all-class coverage. Shirt/tabard placeholders were visible in the UI but absent from this export's item records.

| Target | Observed JSON source path / value | Missing-field / interpretation behavior for future import |
| --- | --- | --- |
| Plan name | `name` | Use only as a display name, never identify/overwrite an existing plan by name |
| Character name | `character.name` | Present; treat as user-supplied identity, not ownership proof |
| Race / class / level | `character.race = DWARF`, `character.gameClass = PRIEST`, `character.level = 60` | Explicit mapping needed; not TBC enum/protobuf values |
| Faction | `character.faction = ALLIANCE` | Present; full race/faction compatibility still needs verified rules |
| Realm / region | No fields in observed `character` | Leave absent; never infer from language or link |
| Equipment | `items[].slot`, `items[].id`; example `HEAD`, `260212` | Named slots, not positional sim equipment; verify IDs against a permitted Forever catalog |
| Enchant | `items[].enchant.id = 8219`, `.spellId = 1249058` for observed neck record | Two distinct identifier domains; do not substitute the spell ID for an enchant effect or assume TBC equality |
| Talent allocation | Initial browser export omitted `talents`; later user export contains `talents[].id`, `.rank`, `.spellId`, `.name` | Distinguish absent field, explicitly empty array, and nonempty allocation. Identifier domains and legality need validation; never infer allocation from the export option alone |
| Game/source | `links.set` has `/forever/set/…`; `phase = 1` | Route is a source hint, not authenticated proof; phase is not a schema/game build |
| Version | No explicit schema or client/catalog version observed | Require a documented source/version compatibility decision; do not guess from level or item names |
| Stats / EP / buffs | `stats`, `points`, `buffs`, `consumables` present | Outside first-version import calculation scope; provider totals are not our verified character stats |

Observed slot labels: `HEAD`, `NECK`, `SHOULDERS`, `CHEST`, `WAIST`, `LEGS`, `FEET`, `WRISTS`, `HANDS`, `FINGER_1`, `FINGER_2`, `TRINKET_1`, `TRINKET_2`, `BACK`, `MAIN_HAND`, `OFF_HAND`, `RANGED`. Keep the provider enum mapping distinct from display labels and future canonical game slots.

A truncated copy failed native JSON parsing as expected. Unknown-item and wrong-game mutations can remain syntactically valid JSON; rejecting them requires our future schema/catalog validator, which was not implemented or tested. Initial absent talents were a partial export, not evidence that the format cannot export talents.

### User-provided nonempty talent export

At the checkpoint the user first provided an explicitly empty `talents: []`, then a nonempty export after assigning points. Its **18 unique talent IDs total 48 allocated points**, leaving three against the observed level-60 budget. Example: `id = 105848`, `rank = 3`, `spellId = 14523`. It also includes `links.talents` pointing to a Forever Priest calculator page. This verifies the supplied nonempty record shape; it does not establish whether `id` is a provider ID, game trait ID or another domain, or prove rank/prerequisite legality.

The user example supplies the previously missing format evidence. No need to decode the sharing URL merely to obtain allocations when these JSON records exist. Phase is still not a catalog version; a source link is still not authenticated origin. No full pasted payload, names, stat tables or raw exporter file were committed.

### Import verdicts remain separate

- **Text/file import: partial.** Public UI export plus supplied gear/enchant/talent JSON shapes are verified. Schema/version compatibility, permitted hosted payload scope and identifier/catalog validation remain unresolved. No raw export file was retained because retention/cross-site scope was not cleared.
- **Link import: blocked.** A human browser can view the page; no documented, permitted machine-export URL/API contract was established. Do not build a server scraper based on that observation.
- **End-to-end import: blocked.** A nonempty talent example is now available; missing examples are no longer the only issue. Resolve permitted catalog acquisition and ID/version validation before implementing imports.

The [terms](https://sixtyupgrades.com/forever/terms) constrain copying/mirroring website materials. That is not a legal conclusion that all user-entered factual build data is forbidden to transfer; obtain a scoped permission/interoperability decision instead of pretending the broad terms grant catalog reuse. Verified community contact entry point: `https://discord.gg/xUnQtQB`; no message sent.

## Capability verdicts

Final statuses below concern production capabilities unless explicitly labeled research. All blocked/partial rows have a concrete next action; no gate is inferred from another row.

| Capability | Status | Evidence / coverage limits | Next action |
| --- | --- | --- | --- |
| Manual item selection | blocked | Public catalogs exist, but permitted hosted acquisition and sufficient level-60 coverage were not established | Obtain scoped catalog permission/access and a versioned representative dataset |
| Enchants | blocked | Export exposes distinct enchant and spell IDs; no permitted authoritative mapping verified | Verify identifier domains, restrictions and effects against that dataset |
| Independent talent planning | partial | CC BY structural evidence, nine class trees, eight provider UI observations, nonempty JSON and an addon-reported share link | Decide versioned talent identity and text/icon presentation rights; validate classes/prerequisites without TBC fallbacks |
| Lower-level restrictions | partial | Talent budgets observed at levels 10/30/60; equipment/race restrictions not verified | Test actual item/slot/class rules and talent budgets for supported levels |
| Source filters | blocked | Source metadata includes provider-labeled Classic-derived loot that may differ from Forever | Acquire attributable Forever sources and preserve unknown/Classic provenance labels |
| Whole-character stats | blocked | No complete independently checked base/racial/talent/level formula coverage | Keep totals unavailable until coverage/parity is demonstrated |
| Addon import: full character | blocked | MIT candidate source inspected; Talents Forever sharing is friend-reported and browser-observed, not gear export | Obtain an equipped-character output with client/addon versions or approve a separate minimal-exporter experiment |
| Sixty text/file import | partial | Public gear JSON and user-supplied nonempty talent record shape verified | Resolve compatibility/rights scope and validate all reference domains before applying a build |
| Sixty link import | blocked | Public browser viewing works, but a permitted machine-export contract was not established | Obtain provider-supported API/export access; do not scrape an editor |
| Numeric talent research | ready | Attributed 17-record projection; source audit and runnable probes | Preserve licence/changes notices; do not promote this to full game-data clearance |
| Talent-only addon share link | partial | Friend reports addon origin; Shaman link opens to a consistent 17/34/0 plan | Check addon/client versions and versioned allocation mapping when this route is implemented |
| Hosted ownership/stable sharing | outside phase | No server, account or publication implementation tested | Separate Gate C plan remains required for permanent latest-edit links |

## Next decisions

**Go:** the bounded feasibility investigation is complete; the product remains technically plausible. Nothing else is required from the user merely to finalize this report.

**No-go for the full editor now:** Gate A lacks a permitted, sufficiently complete, versioned item/enchant dataset. This is a missing-evidence/rights decision, not a claim that the product is impossible or that all game facts are categorically prohibited from use. Both full import routes remain short of end-to-end readiness.

**Recommended next scoped work:** catalog access/permission and identifier mapping. Confirm an authorized provider/contact and request a grant covering a free hosted planner, caching/updates, names/tooltips/icons separately, and user-selected gear/talent interoperability. No inquiry has been sent. Verify a representative armor item, weapon/off-hand conflict, ring, enchant, and level-60 coverage against the chosen version. Map Sixty `id` versus `spellId` and community snapshot positions before deciding an import/build schema.

Only after that evidence is sufficient should the next manual-editor plan include Current/Planned, independent named builds, supported class/level coverage and the item picker. Leave unverified character totals unavailable and omit unverified import routes. A synthetic mockup is not a substitute for passing the real-data gate.

The hosted ownership/publication plan remains a required subsequent milestone: guest/account ownership, recovery/revocation, durable storage and selected published builds under permanent URLs. Browser-local storage alone does not satisfy the Discord/bookmark requirement. Future beta exports are useful optional inputs, not a demand that the friend install more addons or provide account/private details.

No application implementation, merge, push or provider contact is authorized by this research completion.
