# Forever armory: data and import feasibility

Checked 2026-10-02 UTC. Research only; no application changes, authenticated Blizzard calls, client extraction, provider messages, or proprietary catalog copies committed.

## Question

Can the approved hosted Forever armory obtain permitted, versioned item/talent data and import actual or planned builds without depending on a Blizzard character API?

**Initial result:** community talent structure is accessible under an explicit attribution licence. A distributable Forever item catalog and both import routes are not yet verified. The sources have different versions and scopes; their availability must not be collapsed into a single ready verdict.

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

Task 2 has not yet completed. Accessible data and licence evidence do not establish full catalog accuracy, class/race coverage, equipment legality, talent-validation behavior, or character-stat formulas. No TBC rules were substituted.

Known version pitfall: the Talents Forever upstream currently describes client-derived beta data, whereas wowforever.space explicitly retains a September 15 demo snapshot. The repository build marker `70178` differs from the `70170` sites. Never combine these silently.

## Addon import

Task 3 has not yet completed. No real Forever addon export or client was provided. The CurseForge project fetch is denied. Addon policy and search results are not a successful runtime export test.

## Sixty Upgrades import

Task 3 has not yet completed. Public homepage rendered successfully and explicitly says “Set up multiple characters and gear sets.” Terms rendered successfully. The supplied share-link data/export has not yet been inspected in a browser; an HTTP 200 shell is not a machine export.

## Capability verdicts

| Capability | Status after Task 1 | Evidence / missing condition |
| --- | --- | --- |
| Permitted talent structure research | partial | Upstream CC BY contribution and accessible JSON; inspect actual structure/validation next |
| Hosted Forever item catalog | blocked | No confirmed applicable reuse grant and versioned usable acquisition path |
| Addon import | blocked | No verified project/client version or real export fixture |
| Sixty Upgrades import | blocked | Export representation and cross-site reuse scope not verified |
| Whole-character stats | blocked | No verified complete Forever formula coverage |

## Next decisions

Continue the bounded catalog/talent inspection and separate import investigations. If they remain unsupported, report their exact missing inputs rather than claiming a ready editor. Stable hosted ownership/sharing is still part of the product roadmap, but this phase does not implement it.
