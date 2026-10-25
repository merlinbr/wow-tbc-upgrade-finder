# Forever Armory — First-Version Design

Date: 2026-10-01
Revised: 2026-10-25
Status: Design approved in conversation; revised specification awaiting user review.

## Revision note (2026-10-25)

The October 20 review of `wowsims/forever` changed the starting point and corrected several
assumptions. This revision keeps the approved product goal and exclusions and records:

- **Selected upstream.** The catalog and talent structure come from `wowsims/forever` at commit
  `0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4`. Its React application, simulator, optimizer and
  saved-record format are not reused. The repository's MIT licence does not by itself settle
  rights in the client-derived payload; that reuse is a documented, accepted risk with a
  replaceable dataset adapter. Details and identifier rules are in the
  [data and validation contract](../../forever-upstream-data-contract.md).
- **Level promise narrowed.** The pinned catalog carries no required character level, so
  equipment legality below level 60 cannot be verified. Level 60 is the supported
  equipment-planning level for the first version.
- **Talent encoding.** Positional talent strings are an interchange format, not a validator.
  Allocations are stored by talent identity and bound to a catalog revision.
- **Statistics.** The upstream engine is a useful reference, not parity evidence. Whole-character
  stats remain explicitly unavailable until coverage is demonstrated.
- **Native WoWSims interchange** is approved as a gated entry route, inbound only, implemented
  after the manual editor.
- **Delivery shape.** Work is sequenced by the [roadmap](../../plans/2026-10-20-forever-armory-roadmap.md).
  Five wireframes are accepted; the item picker and talent editor await visual review.

This document remains a design artifact, not authorization to implement.

## Goal

A hosted personal armory and character planner for World of Warcraft: Forever.
One permanent character link lets a Discord community or bookmarked reader view
the latest saved Current build and multiple published plans, including independent
talents. Blizzard character APIs are not a prerequisite.

This is a design document, not authorization to start implementation. The first
version is an armory/planner, not a Forever combat simulator or searchable player
directory.

## Approved scope

- Hosted website; no local Go server required for users.
- Multiple characters, usable as a guest or through an optional account.
- Browser-remembered guest editing access and a separate private edit/recovery link.
- One Current build and multiple named plans per character.
- Independent equipment, enchants, and talents for every build.
- Manual creation plus addon, Sixty Upgrades and native WoWSims imports where compatibility is
  verified. Each import route is independently gated.
- Character-centered equipment editor with Current / Planned and Gear / Stats / Talents views.
- Level 60 as the supported equipment-planning level. Lower-level characters may be created and
  have talent budgets validated, but equipment legality below 60 is not claimed until a
  required-level source exists.
- Explicit publishing, a permanent read-only character URL, and published-set switching.
- Saved edits to published builds update the existing shared page.
- Export/import files for backup and portability.

## Domain model

### Character

A character has a stable internal identifier, name, race, class, level, and optional
region/realm metadata. Identity is shared across its builds. Names and realms do
not establish ownership and are not unique identifiers for manual characters.
Changing display identity must not change the permanent sharing URL.

### Current build

Current represents the owner's actual equipped gear and talents. It is empty until
populated manually or through a verified character import. Manual entry is labeled
as manual; imported entry records source and import time. Neither is advertised as
Blizzard-verified. Preserve the distinction between the last import time and the
last edit time.

### Planned build

A plan has a stable identifier, editable name, equipment, enchants, its own talent
allocation, and a publication flag. Plans share character identity but not talent
allocations or equipment with other builds.

Create a plan empty, copy Current, or duplicate another plan. Copies are independent:
a later Current import or plan edit never changes another build. Rename and delete
plans explicitly. Gear-set names such as Pre-raid, Phase 1, and PvP are ordinary
names, not hardcoded categories or automatic content schedules.

A manually created character starts with an empty Current build and an editable
plan; do not invent an equipped baseline.

### Ownership and publication

Ownership permits private reads and mutations. Publication grants read-only access
to selected content. These are separate permissions; knowing a public character
identifier never grants editing access.

## User experience

### Character library

Owners can create and reopen multiple characters and see their named plans. Guest
access is remembered in the creating browser. Account users can access owned
characters from another device by signing in.

### Armory editor

- Current / Planned tabs distinguish actual equipment from targets.
- Planned has a named-set selector and create, duplicate, rename, and delete actions.
- Gear / Stats / Talents views operate on the selected build.
- Selecting an equipment slot opens an item picker with text search, equipment-compatibility
  restrictions, and source filtering where source data exists. Character-level restrictions are
  shown only where the catalog carries them.
- The picker and tooltips use the Forever catalog, never the bundled TBC database.
- Equipment and enchant edits affect only the selected build.
- Talents can be allocated independently per plan; Current can be populated manually
  or by import. Enforce verified class/tree/rank/prerequisite/point-budget rules.
- Controls work with keyboard and touch, with clear focus and active-state labels.
- Destructive actions require confirmation. Failed saves retain the user's edits
  and clearly report that the shared page has not been updated.

### Levels and statistics

Level 60 is the supported equipment-planning level, not a universal character restriction. Store
the selected level and derive the talent budget from it (one point per level from 10, so 51 at
60). Apply item requirements only where the catalog carries them: the pinned Forever catalog has
no required-character-level field, so equipment legality below 60 is shown as unverified rather
than inferred from item level or expansion labels. Existing incompatible choices must be flagged
rather than silently removed or treated as valid.

Show individual catalog item information where verified. Whole-character stat totals
and derived percentages appear only for combinations covered by verified Forever
rules. If coverage is missing, show an explicit unavailable/partial state and name
what is missing; do not present item-stat sums as full character totals. Whole-character
totals are not a prerequisite for shipping a truthful item and talent planner.

Do not reuse TBC racial, talent, buff, stat-conversion, or combat formulas as Forever
formulas. No DPS estimates or simulation-based upgrade claims in this version.

## Entry points and import behavior

1. Manual: choose identity and level, then edit equipment and talents.
2. Addon: import actual equipped gear/talents into Current through a verified format.
3. Sixty Upgrades: import a planned build through a verified supported representation.
4. Native WoWSims interchange: open a Forever simulator link or JSON payload as a new plan through
   an explicit adapter bound to one native protocol version and one catalog revision. Implemented
   in Phase 4, after the manual editor, inbound only.

Manual entry is the foundation. Routes 2-4 are conditional release features: do not advertise
them as available until their own feasibility gates pass. Route 4 was proposed by the October 20
review and is approved only as a gated Phase 4 route; its snapshot link is a payload, not a hosted
record, so opening one never creates ownership or publication.
A Sixty Upgrades URL is not assumed to be an accessible export API. Existing addon
Forever support is not assumed. An addon route may require a minimal exporter if
no compatible, permitted exporter exists; that decision follows investigation.

Imports are validated against the correct Forever catalog/version. Parse and
validate before applying changes. On malformed input or unsupported references,
report the problem without overwriting the existing build or silently substituting
TBC data. Replacing a populated Current build requires confirmation. Importing a
planned set creates a separate plan rather than overwriting another plan by name.

Record available source/version metadata. Equipment data alone does not imply that
appearance, enchants, or talents were included; missing fields remain explicit.

## Sharing and publishing

Characters and builds are unpublished by default. The owner explicitly publishes
Current and/or selected plans. Publishing at least one build enables the stable
character URL. Public readers can switch only between published builds and see
the selected build's gear and talents.

Once a build is published, successful saves update what readers see on their next
load or refresh. Real-time push updates and a separate draft revision of an already
published build are not required. Duplicate a build into an unpublished plan when
experimenting privately.

Unpublishing a build removes it from the public response. With no published builds,
the character link shows a generic unavailable page; the identifier remains reserved
so republishing restores the same link. Deleting the character invalidates public
access. Private builds and editing credentials must never be sent to public clients,
including hidden dropdown entries, embedded page data, or export responses.

Profiles are unlisted initially, with no player search, leaderboard, or directory.
Anyone with the public link can view published content. Unlisted is not confidential;
avoid indexing shared profiles and explain link access clearly.

## Guest and account access

Guest editing uses a high-entropy credential independent of the public identifier.
The browser remembers access. A deliberately generated private edit/recovery link
can restore access to that character in another browser. Treat it as a password:
warn against posting it, keep it out of the public Share action, and allow revocation
and replacement. Guest credentials must not grant access to other owners' records.

Recovery/edit-link secrets must not appear in public URLs, analytics, or ordinary
request logs. Exchange recovery credentials through an appropriately protected flow
rather than placing long-lived edit secrets in query strings. Use HTTPS, secure
credential storage, server-side authorization on every private operation, bounded
inputs, abuse controls, and safe rendering of user-authored names.

Without a retained private link, remembered browser access, or an attached account,
guest editing access may be irrecoverable. Display that warning before relying on
browser storage alone. Guests still have file export/import for backup.

An authenticated account may claim a guest character only after proving guest editing
access. Claiming is atomic, preserves the public URL, and invalidates the prior guest
credentials. Account ownership is application ownership, not proof of in-game identity.
Choose a maintained authentication solution during implementation planning; do not
invent a password or account-recovery system.

## Architecture and existing-project reuse

Use a browser frontend, a hosted service, and durable server-side character storage.
The service owns access control, validated saves/imports, and the public projection.
Browser storage remembers guest access and preferences; it is not the sole copy of
published characters. Backups and a documented restoration path are necessary for hosted data.

### Selected upstream (2026-10-25)

The Forever catalog and talent structure come from `wowsims/forever` at commit
`0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4`, pinned by payload hashes in the
[data and validation contract](../../forever-upstream-data-contract.md). Candidate reuse is
limited to that data and to framework-independent model helpers: item-search normalization,
compatibility predicates, talent tree structure, and the talent-string codec. The React
application, simulator, optimizer, rotations, EP and combat calculations are not reused, and no
Forever payload may enter the TBC import, enrichment, or stat path.

The MIT licence covers the repository's software and documentation. It does not by itself resolve
rights in client-derived or third-party data, so this reuse is a documented, accepted risk rather
than a clearance. The contract keeps the catalog behind an adapter, preserves attribution, and
defines a replacement path so a licensed or self-generated dataset can be substituted without
changing the domain, validation rules, or UI contracts.

Reuse the current Svelte armory layout, equipment-card/tooltip patterns, and applicable
UI interactions. Separate ownership/import data from presentation. The TBC 17-slot projection
happens to match the observed Forever slot count, but its enum values, talent metadata, import
protobufs and engine-backed stats are not automatically valid for Forever. The contract fixes
canonical slots, identifier domains and validation rules before implementation.

Do not replace the existing TBC upgrade agent or refactor its simulator into a generic
multi-game framework as part of this feature. Decide deployment and code placement
in implementation planning after the feasibility checks. There is no need to select
a new frontend framework to describe or validate this product.

## Feasibility gates and delivery boundaries

### Gate A: Forever catalog and rules

Identify usable, permitted, versioned item/enchant/talent data, source metadata, slot
rules, and talent validation rules. Check actual data rather than relying on search
summaries. Required evidence is a representative dataset and valid/invalid examples
for manual gearing and talent allocation. Whole-character stat support has its own
coverage evidence and can remain explicitly unavailable.

### Gate B: Import formats

Obtain real addon and Sixty Upgrades fixtures, document how they are obtained lawfully,
and verify identity, gear, enchant, and talent mapping. Each importer is independently
gated. If one is unavailable, manual creation remains usable and that route is not
advertised. Do not bypass access controls or treat a JavaScript-rendered page as an
API contract.

### Gate C: Hosted ownership and stable sharing

Demonstrate guest creation/recovery/revocation, account claim, durable saves, independent
plans, and public projections that expose only published builds. Confirm that edits,
renames, and account attachment preserve the sharing identifier.

### Gate status at this revision

| Gate | Status | Basis |
| --- | --- | --- |
| A: catalog and rules | **Partial** | Pinned candidate selected and contract written; representative identifiers and talent rules verified. Required character level and stat parity remain open, and reuse is an accepted risk rather than a clearance. |
| B: import formats | **Partial / blocked** | Native interchange is approved but gated; addon has no real Forever fixture; Sixty gear shape is verified but its talent ID domain is unresolved; the URL route is blocked. |
| C: ownership and sharing | **Not started** | Phase 3 of the roadmap. |

These boundaries can become separate implementation plans: data/import feasibility,
manual armory/planner, and hosted ownership/sharing. Importers attach only once verified.
Do not start a single large implementation plan that assumes Gates A and B have passed.

## Verification contract

Leave runnable checks for the implemented behaviors using the project's chosen tools:

- Copying, editing, importing, and deleting a plan never mutate sibling builds.
- Independent talents respect verified rules and the character level.
- Invalid imports and failed saves preserve the previous valid data.
- Incompatible/unknown catalog references are reported, not silently substituted.
- Guests cannot access another owner's private data or mutate through a public link.
- Revoked credentials fail; guest-to-account claim preserves URL and retires guest access.
- Public responses contain only published builds and never editing secrets.
- Saving a published build updates the same public URL; renaming does not change it.
- Unpublishing/deleting removes public content as specified.
- Export/import preserves supported character/build data and rejects credential injection;
  backups do not contain authentication secrets or automatically confer old ownership.
- Keyboard/touch flows cover item selection, build switching, and talent allocation.
- A hosted storage restart/restore retains characters and stable identifiers.

## Deferred

Searchable public armory, verified in-game ownership, Blizzard API synchronization,
combat simulation, upgrade rankings, automated gear optimization, 3D characters,
buffs/consumables planning, pricing, multi-user editing, revision history, live push
updates, and public community discovery. Add these only under their own requirements.

## Evidence and limitations

- Existing project: `README.md`, `docs/upgrade-finder.md`, `docs/STATE.md`,
  `cmd/wowsimcli/cmd/upgrades/import.go`, `simulator.go`, and
  `ui-finder/src/lib/ArmoryView.svelte` establish that the current app is a local
  browser UI backed by a TBC-specific Go simulator and in-memory imports/reports.
- The [October 20 review](../../forever-armory-wowsims-review.md) compared this design against
  `wowsims/forever` at commit `0cfb1a027cd4f73ae42e1db9b78d10ed6532a1f4`. It reproduced the TBC
  import panic on a real Forever payload and recorded catalog counts, hashes, identifier domains
  and reuse risks.
- The [data and validation contract](../../forever-upstream-data-contract.md) fixes the pinned
  manifest, identifier domains, validation rules, import-route status and open items.
- Blizzard's Forever panel recap describes new class/talent/content/race changes:
  https://worldofwarcraft.blizzard.com/en-us/news/24303862/world-of-warcraft-forever-whats-next-panel-recap
- User-provided Sixty Upgrades screenshots show a signed-out character library with
  multiple sets, a set selector in an editor, and a shared Current armory view without
  a visible set selector. They do not establish its guest authorization/storage
  implementation or exhaust all of its sharing capabilities.
- The provided Forever set link returned an application shell during inspection;
  its equipment and export capabilities were not verified:
  https://sixtyupgrades.com/forever/set/vH9HQmj6tzxi7fajuHmu8p
- A Forever-specific Blizzard profile API was not confirmed in this exploration.
  Lack of confirmation is not proof that no such API exists.
- Search surfaced https://wowforeversim.com/, whose fetched landing page labeled its
  data pre-beta/confidence-rated. No simulator implementation or accuracy was verified.

The proposed value is a cohesive character-centric editing and stable sharing
experience, not a claim that other planners lack multiple gear sets.
