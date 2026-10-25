# Forever armory roadmap

Date: 2026-10-20
Status: Proposed delivery sequence and visual-design brief. Not an approved implementation plan or authorization to change the application.

Progress (2026-10-25): the [revised specification](../superpowers/specs/2026-10-01-forever-armory-design.md)
and the [upstream data and validation contract](../forever-upstream-data-contract.md) are written.
Five of the seven wireframes are accepted; the item picker and talent editor still await visual
review. Phase 1 exit is not yet met.

## Goal and starting point

Build a hosted personal Forever armory/planner: multiple characters, one actual Current build and independent named plans, guest or optional account ownership, and a permanent read-only URL showing the latest saved published builds.

Keep the existing Svelte visual direction and character-centered presentation. Use WoWSims Forever as a candidate upstream for pinned catalogs, structural talent data and selected framework-independent helpers. Do not fork its entire simulator or replace the existing TBC upgrade finder by default.

Sources:

- [Existing product specification](../superpowers/specs/2026-10-01-forever-armory-design.md)
- [WoWSims comparison and runtime findings](../forever-armory-wowsims-review.md)
- [Earlier data/import research](../forever-data-import-research.md)

These phases sequence the existing scope; they do not silently remove account access, lower-level planning, independent talents, backups or other specified requirements. Imports remain independently gated. Simulation, optimization, 3D and a searchable public directory remain excluded.

## Phase 1 — Revised design, wireframes and upstream data contract

### Product and visual design

Revise the existing specification with the comparison findings. Preserve the distinction between a complete build and upstream's separate gear/talent presets, and between a permanent hosted page and an editable simulator snapshot.

Produce wireframes for the previously undesigned character, editing and publishing flows. Start with the character editor as the visual anchor, then derive connected screens from the reviewed direction. Keep the work clearly labeled as design-only: illustrative names/items, no verified character totals, no implication that saves or authentication already work.

### Upstream adoption contract

Resolve concrete adoption questions rather than restarting a broad catalog search:

- Pin upstream commit, catalog/talent content hashes, and captured build/hotfix group provenance. Keep schema version separate from catalog revision.
- Define item IDs, enchant effect/spell/carrier-item IDs, talent identity and canonical slot mappings explicitly.
- Define version-bound talent allocations and migration or unsupported-reference behavior. Never reinterpret old positional allocations against a reordered live tree.
- Establish the intended hosted reuse scope, including numeric/structural facts separately from names, tooltip text, artwork/icons and inherited acquisition tables; retain applicable notices and attribution.
- Obtain required-character-level metadata and level-aware restrictions for the specified lower-level planning. Item level is not required character level. Do not promise lower-level legality from upstream's fixed-level-60 catalog alone.
- Define complete equipment/enchant/talent validation. Distinguish transient incompatible edits from valid supported builds; retain and flag existing incompatible choices rather than silently deleting them.
- Keep Forever payloads out of the TBC import/enrichment/stat pipeline. The comparison reproduced a recovered request panic when a live Forever export reached that path.
- Select real representative records for slot/class/level restrictions, weapon/offhand conflicts, uniqueness/categories, enchant fit, talent rank/tier/prerequisite/budget boundaries and version mismatches.

**Exit:** reviewed written design and wireframes; a concrete upstream data/validation contract with representative evidence, unresolved gaps explicitly identified, and an adequate reuse decision for the data intended for implementation. No full editor implementation plan may assume the catalog gate passed without that evidence.

## Phase 2 — Manual character planner

Implement against the reviewed domain and data contract, reusing existing Svelte presentation where applicable:

- Create/reopen characters with name, race, class, selected level and optional realm/region metadata.
- Begin manual characters with an empty Current build and an editable plan; never invent actual equipped gear.
- Distinguish Current from named Planned builds. Every build owns its equipment, enchants and talents.
- Create an empty plan, copy Current, duplicate, rename and delete plans using stable build identities.
- Add Forever item/enchant selection and talent editing, with verified compatibility and level-aware rules.
- Show explicit unsupported/partial statistics instead of TBC calculations or misleading item-stat sums.
- Preserve edits after failed saves and confirm destructive actions.
- Support keyboard and touch in build switching, picking and talent allocation.

**Exit:** usable manual editing against supported real data, with observable plan isolation and validation behavior. A local editor milestone is not the complete hosted product; durable ownership/publication remains Phase 3.

## Phase 3 — Hosted ownership and permanent sharing

Design the storage/authorization interfaces against the character/build contract; infrastructure decisions do not need to wait for every visual detail. Integrate with the manual planner to deliver:

- Durable server-side character/build storage; browser storage is not the authoritative published copy.
- High-entropy guest editing access independent of the public identifier; browser-remembered access, deliberate private recovery exchange, credential revocation/replacement and loss warnings.
- Optional maintained account authentication; atomic guest claim after proof of editing access, preserving the public URL and retiring guest credentials.
- Unpublished-by-default builds and explicit publication of Current and/or selected plans.
- A stable read-only character URL whose public projection includes only published content and updates after successful saves.
- Public build switching, unpublish/delete transitions and stable identity across renames/account attachment.
- Versioned character/all-plan export/import without credentials or automatic transfer of old ownership.
- Backups and an exercised storage restart/restore path.

**Exit:** the end-to-end milestone below works across owner and reader browsers, with authorization and durability evidence. Account, backup and security acceptance criteria remain part of the full phase even when an earlier guest-only milestone is demonstrated.

## Phase 4 — Verified imports

Attach each route independently, only after its own compatibility and data-use checks pass:

- **Addon to Current:** obtain a consented real equipped-character output with addon/client versions; map identity, level, slots, enchants and talents. Confirm replacement of populated Current.
- **Sixty text/file to a new plan:** resolve explicit slot and identifier mappings, rank-bearing talents, version assumptions and payload-use scope. Do not overwrite a plan by display name.
- **Native WoWSims interchange:** proposed addition to discuss in the design revision, not already-approved scope. Use an explicit Forever adapter and compatible catalog/talent version; never reuse the TBC parser as a shortcut.
- **Sixty URL:** remains unavailable without a provider-supported machine-export contract; a readable web page is not such a contract.

All imports parse, resolve and validate before changing saved data. Malformed input, missing references and incompatible versions must leave the previous build intact. Preserve source/import time separately from last edit time, and distinguish missing fields from explicitly empty allocations.

**Exit:** each enabled route has real compatible fixtures and successful/failing end-to-end evidence. Unsupported routes are not advertised as available.

## Phase 5 — Statistics with demonstrated coverage

Adopt only supported Forever character calculations, independently of combat simulation:

- Define the supported class/race/level/item/enchant/talent combinations and calculation context.
- Check base/racial/conversion behavior, then gear, talents and conditional/set effects against appropriate evidence.
- Identify unsupported effects and inherited/manual assumptions instead of treating catalog presence as modeled accuracy.
- Keep unsupported totals unavailable/partial, with a reason. Do not present upstream Alpha status as verification.

**Exit:** displayed whole-character totals and derived percentages have coverage/parity evidence. Broad statistics are not a prerequisite for a truthful item/talent planner.

## First end-to-end milestone

Create a character → populate Current manually → copy it to a named plan → change that plan's equipment and talents → save → publish selected builds → open the permanent URL in another browser.

Demonstrate:

1. Current and sibling plans remain unchanged when one plan is edited.
2. Invalid imports and failed saves preserve previous valid data and the user's unsaved work as specified.
3. Unpublished builds and credentials never reach the public reader; the public link cannot mutate the character.
4. A successful published-build save updates the same URL, while renaming does not change its identity.
5. The guest owner can recover access; revoked access fails.
6. Stored characters and stable identifiers survive service restart/restore.

This milestone connects the editor and hosted plans. It does not require addon/Sixty import, simulation, optimization or 3D.

## Initial wireframe set

Design objective: validate information hierarchy, navigation, ownership/publication clarity and responsive behavior before polished artwork. Use the existing dark armory/card/tab language as a reference, not a requirement to preserve TBC-specific data or controls.

| Screen / state | Decisions it must make visible |
| --- | --- |
| Character library and creation | Multiple characters, guest/account access, create/reopen actions, identity/level entry and the initially empty Current state |
| Owner character editor — desktop anchor | Identity header, Current/Planned switch, named-plan selector, Gear/Stats/Talents views, character-centered equipment layout, save status and private/public state |
| Item picker | Selected build/slot, search and applicable restrictions, source provenance, selection details, incompatible/unknown states and clear return to editing |
| Talent editor and narrow layout | Build-owned talent allocation, points/budget, dependency feedback and touch/keyboard interaction; no misleading TBC trees |
| Publishing panel and public page | Explicit per-build publication, public link distinct from recovery, read-only build selector, saved/imported timestamps, unavailable/unpublished state |
| Guest recovery/access panel | Private credential warning, restore/revoke/replace concepts, lost-access warning and optional account attachment; secrets never included in public Share |

Show save-failure/unsaved states and destructive confirmations as focused annotations or overlays rather than inventing separate navigation for every state. Mobile should preserve the same build/publication semantics, not merely shrink desktop columns.

Existing visual references, all TBC implementation evidence rather than Forever mockups:

- [Equipment layout and tooltip](../armory-tooltip-neck.png)
- [Desktop talent presentation](../talent-tree-after-desktop.png)
- [Mobile talent presentation](../talent-tree-after-mobile.png)
- [Mobile equipment tooltip](../armory-tooltip-mobile.png)

### First wireframe artifact — owner editor

Created a single responsive owner-editor wireframe with desktop and narrow-screen captures. Planned / Pre-raid / Gear is the illustrated state; Current, talent editing, the library, picker and access/publication panels are review affordances, not separately implemented screens.

- [Desktop screenshot](../wireframes/forever-armory-owner-editor-desktop.png)
- [Mobile screenshot](../wireframes/forever-armory-owner-editor-mobile.png)
- [Editable standalone HTML](../wireframes/forever-armory-owner-editor.html)
- [Browser preview](https://p.superdesign.dev/draft/75ecb954-c480-4d85-aefa-bfdb094b1943)
- [Design canvas](https://superdesign.dev/teams/c2f2a0e9-0779-4902-9e75-310455e112ad/projects/0cd92272-a4d0-45d1-9e62-7b319c414266)

The service blocked model generation because the team's credits were exhausted. The wireframe was authored directly and imported without generation credits; draft version 1 was retrieved successfully. No real game imagery or private data was uploaded.

Verification: rendered both the local HTML and imported browser preview at 1440px and 390px widths. Both retained all 17 equipment slots, and neither had horizontal page overflow. Controls are explicitly static design affordances, not working saves/authentication/publication. Full-page screenshots are saved locally above. No production application code changed.

Baseline accepted by the user on 2026-10-20: “Its fine we can use this as the base.” Use this owner-editor hierarchy and visual direction for subsequent wireframes; do not restart style exploration. This accepts the visual baseline, not the complete product specification or application implementation.

### Connected wireframe pair — character library and public profile

Screen briefs approved by the user on 2026-10-20: “yes sounds fine.” Rendered character-library and public-profile layouts subsequently accepted: “looks good to me.” Preserve these layouts and the owner-editor visual baseline for subsequent wireframes. This accepts the design artifacts, not application implementation or the complete product specification.

| Screen | Standalone HTML | Desktop capture | Mobile capture | Browser preview |
| --- | --- | --- | --- | --- |
| Character library | [HTML](../wireframes/forever-armory-character-library.html) | [Desktop](../wireframes/forever-armory-character-library-desktop.png) | [Mobile](../wireframes/forever-armory-character-library-mobile.png) | [Preview](https://p.superdesign.dev/draft/60303dc2-f756-4021-8dbb-86e86101a44f) |
| Public profile | [HTML](../wireframes/forever-armory-public-profile.html) | [Desktop](../wireframes/forever-armory-public-profile-desktop.png) | [Mobile](../wireframes/forever-armory-public-profile-mobile.png) | [Preview](https://p.superdesign.dev/draft/fdd9dc78-0020-4aed-9738-c7ffd5ec3ea6) |

The library retains the editor's compact navigation and dark card language. It shows two owned characters, named private plans, create/reopen affordances, an explicitly empty Current for a new level-12 character, and a separate browser-access warning/private-recovery/account entry panel. Creation remains an annotation, not a working form or an invented equipped build.

The public profile retains the character header and all 17 equipment slots. Its only build options are published Current and Pre-raid; Pre-raid / Gear is illustrated. Gear/Stats/Talents are static viewing affordances, equipment cards are not edit buttons, and no owner controls or private plan names appear. Last saved is distinct from source and last imported; this manual sample explicitly says “Not imported.” Statistics remain unavailable pending verification. A focused annotation covers the generic unavailable state when no builds are published.

Both pages were authored directly and imported into the existing design canvas without model-generation credits. Retrieved draft version 1 for each; the saved remote HTML exactly matched its local standalone file. Only the two illustrative HTML files were uploaded, without credentials, game artwork or real character exports.

Verification: rendered local HTML and imported previews at 1440px and 390px widths, with no horizontal page overflow. Both public layouts retained 17 slots, only the two published options, and no owner-action controls. The public build selector showed a visible 3px keyboard-focus outline. Four local full-page captures are linked above. Controls are static design affordances, not working build switching, persistence, authentication, publication or import.

No production application code or accepted owner-editor artifact changed during this pair.

### Access wireframe pair — publishing and guest recovery

The user approved this pair's scope on 2026-10-20: “okay lets do this.” Both artifacts follow the accepted editor/library/public-profile language. The recovery revision's visual acceptance is recorded below; these design artifacts do not authorize application implementation.

| Screen | Standalone HTML | Desktop capture | Mobile capture | Browser preview |
| --- | --- | --- | --- | --- |
| Publishing panel | [HTML](../wireframes/forever-armory-publishing-panel.html) | [Desktop](../wireframes/forever-armory-publishing-panel-desktop.png) | [Mobile](../wireframes/forever-armory-publishing-panel-mobile.png) | [Preview](https://p.superdesign.dev/draft/cb4e88f8-6cf8-44a8-ac24-d786da2d36cf) |
| Guest access and recovery | [HTML](../wireframes/forever-armory-guest-access.html) | [Desktop](../wireframes/forever-armory-guest-access-desktop.png) | [Mobile](../wireframes/forever-armory-guest-access-mobile.png) | [Preview](https://p.superdesign.dev/draft/0bfcc4eb-5ff4-48df-9e25-2b9fe8059460) |

The publishing panel illustrates Current and Pre-raid as published and Holy, later as private, explicitly a different sample sharing state from the earlier owner views. Each build has its own publish/unpublish affordance. The public-link card uses a reserved `.example` URL, not a real hosted profile. Copy, preview, and publication controls are static. Save-versus-publish guidance, private experimentation, unpublish confirmation, the no-published-builds transition and failure preservation are visible; no private recovery credential is included.

The guest-access panel now uses three interactive tabs: **Recovery** (default), **Manage access**, and **Account**. A short private-editing warning remains above the tabs. Recovery keeps the browser-loss warning, deliberate private-link generation, masked restore input and irrecoverable-access guidance visible. Manage access keeps replacement/revocation warnings beside their actions. Account retains proof-of-editing-access requirements, stable public URL and retirement of guest credentials. Technical requirements, error scenarios and backup/ownership distinctions are retained in collapsed Design notes, not deleted.

The user approved trying this tabbed revision after finding the original page too dense, then accepted the result: “so much better now with the 3 tabs.” Keep **Recovery / Manage access / Account** as the recovery layout baseline. Recovery is draft version 2 under the same preview URL; version 1 remains in Superdesign history. Publishing stays unchanged at version 1. Only tabs and note disclosures work, not authentication, credential generation, publication or account attachment.

Verification: all three tabs were exercised locally and in the imported preview at 1440px and 390px widths. Exactly one panel is visible at a time, with no horizontal overflow. Design notes open and close; a synthetic restore value survives tab switching. Local keyboard checks covered Left/Right wrapping, Home/End, selected state and a 3px focus outline; keyboard switching and note expansion also worked in the imported preview. No local browser script errors were reported. Remote version-2 HTML exactly matches the local file. Only the revised illustrative HTML was uploaded, without credentials or real character data; no production code changed.

The Recovery desktop/mobile captures in the table are refreshed. Additional tab captures:

| Tab | Desktop | Mobile |
| --- | --- | --- |
| Manage access | [Desktop](../wireframes/forever-armory-guest-access-manage-desktop.png) | [Mobile](../wireframes/forever-armory-guest-access-manage-mobile.png) |
| Account | [Desktop](../wireframes/forever-armory-guest-access-account-desktop.png) | [Mobile](../wireframes/forever-armory-guest-access-account-mobile.png) |

### Final wireframe pair — item/enchant picker and talent editor

The user approved the screen briefs: “yes do it.” These drafts follow the accepted dark armory/card language, tabbed navigation and collapsed Design notes. Both show Elara's private Planned / Pre-raid context and await visual review; this approval does not authorize production editing or verified game rules.

| Screen | Standalone HTML | Desktop capture | Mobile capture | Browser preview |
| --- | --- | --- | --- | --- |
| Item/enchant picker — Items | [HTML](../wireframes/forever-armory-item-picker.html) | [Desktop](../wireframes/forever-armory-item-picker-desktop.png) | [Mobile](../wireframes/forever-armory-item-picker-mobile.png) | [Preview](https://p.superdesign.dev/draft/50ce446a-d1eb-43e0-8097-51e76aaa9875) |
| Talent editor — Tree A | [HTML](../wireframes/forever-armory-talent-editor.html) | [Desktop](../wireframes/forever-armory-talent-editor-desktop.png) | [Mobile](../wireframes/forever-armory-talent-editor-mobile.png) | [Preview](https://p.superdesign.dev/draft/718fb8c7-e29b-4999-b208-8beaf5a41ae0) |

Additional tab captures:

- Enchants: [desktop](../wireframes/forever-armory-item-picker-enchants-desktop.png), [mobile](../wireframes/forever-armory-item-picker-enchants-mobile.png).
- Tree B: [desktop](../wireframes/forever-armory-talent-editor-b-desktop.png), [mobile](../wireframes/forever-armory-talent-editor-b-mobile.png).
- Tree C: [desktop](../wireframes/forever-armory-talent-editor-c-desktop.png), [mobile](../wireframes/forever-armory-talent-editor-c-mobile.png).

The picker includes working Items/Enchants tabs, search, source/demo-compatibility filters, result selection and a no-results/reset state. Selected details distinguish item identity from enchant effects and required character level from item level. Synthetic fit/incompatible/unknown examples are explicitly unverified; incompatible and unknown new choices cannot enable Apply. An existing unknown Head choice is retained and explained. Per-tab selections survive navigation, hidden-tab rendering does not override the active action state, and filtered-out selections are not actionable. Apply and return are illustrative affordances, not equipment mutation or a save.

The talent editor shows one neutral sample tree at a time, with sticky demo-point accounting, selected-node details, rank controls and synthetic prerequisite/cap/budget feedback. Its 9-point demo budget is not a real level-60/class budget. Tree selections and ranks persist when switching; unmet prerequisites retain and flag existing preview ranks rather than removing them. A separate unknown allocation stays flagged and outside the synthetic tally. Reset requires confirmation, cancel retains ranks, and confirmed reset clears only supported Pre-raid demo ranks while preserving the unknown fixture and current tree. Current/sibling builds and storage are outside this prototype. No TBC tree, game statistic or real Forever validation is used.

Verification: rendered all five tabs locally and in imported previews at 1440px and 390px widths, with exactly one panel visible and no horizontal page overflow. Ten full-page captures are linked above. Local browser exercises covered picker fit/incompatible/unknown states, combined filters, empty results/reset, per-tab selection and keyboard navigation; talent node selection, prerequisite lock/unlock, retained incompatible allocation, rank cap, cross-tree demo budget and confirmed/canceled reset. The mobile point summary remained visible at the rank controls. Keyboard focus showed a 3px outline; no local browser script errors were reported.

Imported-preview checks also exercised picker empty/reset/unknown states and keyboard tabs, plus talent rank editing, canceled reset and keyboard tree switching. Both directly authored drafts were imported without model-generation credits and retrieved at version 1; remote HTML exactly matched the local files. Only the two illustrative HTML files were uploaded, without credentials, game artwork or real character exports. No production application code or accepted wireframe changed.

This supplies draft artifacts for the last two screen types in the initial wireframe set. Their visual review, the revised written specification and the upstream data/validation contract remain separate gates; the Phase 1 exit is not implied by having these mockups.

## Planning and approval boundaries

The roadmap is a proposed sequence, not a task-by-task implementation contract. Review the visual direction and updated written specification before producing the corresponding implementation plans. Split plans around the data boundary/manual editor, hosted ownership/publication and independently gated import routes; retain explicit interfaces and complete product acceptance criteria.

Mockups can proceed as design artifacts using clearly illustrative content. They do not establish real-data permission, validated game rules, working persistence or production readiness. Keep the existing TBC application and unrelated work unchanged during design.
