# Talent tree redesign research

Date: 2026-09-07. Research and proposal only; no application implementation.

## Recommendation

Keep the current read-only imported build and replace the Svelte talent presentation, not its data model. Use three portrait tree panels, original TBC backgrounds, explicitly positioned icon nodes, numerical rank badges, and the existing tooltip lifecycle. Native CSS Grid covers the layout; a talent editor or graph library does not belong in this change.

The supplied image is the visual reference, not a new gameplay-data source. Its useful features are the three illustrated panels, sparse four-column grids, compact headers, desaturated unallocated icons, lime allocated borders, and bottom-right ranks. Do not infer its website, build-switching implementation, or asset license from the image.

Proposed design: [talent-tree-redesign-design.md](superpowers/specs/2026-09-07-talent-tree-redesign-design.md). Proposed execution sequence: [talent-tree-redesign.md](superpowers/plans/2026-09-07-talent-tree-redesign.md).

## Current implementation and observed baseline

The actual path is `App.svelte` → `ArmoryView.svelte` → the Talents tab → `TalentTrees.svelte`.

- [`ArmoryView.svelte:73–75`](../ui-finder/src/lib/ArmoryView.svelte) supplies only `character.class` and the imported `talentsString`.
- [`TalentTrees.svelte:3–51`](../ui-finder/src/lib/TalentTrees.svelte) already imports all nine class JSON files, decodes three tree strings, builds a coordinate map and materializes empty grid cells.
- Its current node is a non-focusable `div` with a name, rank pips and native `title`, not an icon (`:65–75`).
- [`app.css:249–261`](../ui-finder/src/app.css) defines no outer grid columns. Each tree therefore fills a separate row. Opaque, flexible-width text cells obscure the art; `background-size: cover` crops it to the wide panel shape.
- [`talents.js`](../ui-finder/src/lib/talents.js) owns the existing digit decoding and totals. The talent's original array index is the serialization contract; coordinates are presentation data, not rank-string indices.
- [`upgrade_server.go:224–286`](../cmd/wowsimcli/cmd/upgrade_server.go) returns one `talentsString`, not named saved builds, a secondary build, or an active-build index.

### Browser experiment actually run

Started the existing `wowsimcli.exe` API and the current Vite source UI. Imported the real `retribution_no_settings_link.txt` fixture through the normal form, then opened Talents. No ranking job was started.

| Observation | Result |
| --- | --- |
| Desktop viewport | 1280 × 1000 CSS px |
| Current talent area | 1150 × 1781.5 CSS px |
| Tree panels | Three vertically stacked panels, each approximately 1150 × 583.2 px |
| Imported totals | Holy 5, Protection 11, Retribution 45 |
| Talent-node images | Zero; names and pips only |
| Narrow viewport | 390 × 844 CSS px; document width remained 390 px |
| Narrow panels | 332 px wide; approximately 620–647 px tall, with wrapped talent names |

Saved evidence: [desktop talent area](talent-tree-before-desktop.png) and [mobile viewport](talent-tree-before-mobile.png). These are **before** screenshots, not redesign mockups.

Browser image-load checks succeeded for the three Paladin background URLs (382, 383, 381), each 300 × 331 pixels, and `ability_golemthunderclap.jpg`, 56 × 56 pixels. This establishes present availability for those four resources only, not all remote images or permission to redistribute them.

## Dataset audit actually run

An in-memory JavaScript audit read all nine `ui/core/talents/trees/*.json` files and joined first-rank spell IDs against **`assets/database/db.json.spellIcons`**. Do not join against item IDs: item and spell numeric IDs overlap.

| Class | Trees in serialized order | Node counts |
| --- | --- | --- |
| Druid | Balance / Feral Combat / Restoration | 21 / 21 / 20 |
| Hunter | Beast Mastery / Marksmanship / Survival | 21 / 20 / 23 |
| Mage | Arcane / Fire / Frost | 23 / 22 / 22 |
| Paladin | Holy / Protection / Retribution | 20 / 22 / 22 |
| Priest | Discipline / Holy / Shadow | 22 / 21 / 21 |
| Rogue | Assassination / Combat / Subtlety | 21 / 24 / 22 |
| Shaman | Elemental / Enhancement / Restoration | 20 / 21 / 20 |
| Warlock | Affliction / Demonology / Destruction | 21 / 22 / 21 |
| Warrior | Arms / Fury / Protection | 23 / 21 / 22 |

Results:

- **27 trees, 579 talents, 85 explicit prerequisite references.**
- Every tree spans **nine rows and four columns**, has a background URL, and currently orders its talents strictly by row then column.
- **579/579 first-rank spell IDs have nonempty icon labels.** All explicitly listed rank IDs also resolve in `spellIcons`; omitted/inferred higher rank IDs were not audited.
- A compact first-rank-ID → icon-label projection serializes to **19,723 bytes**. The source `db.json` is **3,102,050 bytes**. These are UTF-8 source sizes, not compressed bundle sizes or network measurements.
- Talent fields are `fieldName`, `fancyName`, `location`, `spellIds`, `maxPoints`, and optional `prereqLocation`. There is no talent description or tree-header icon field.

### Reproducible coverage experiment

Run from the repository root with Node; this reads data only and writes no files:

```sh
node --input-type=module -e "import { readFileSync } from 'node:fs'; const read = p => JSON.parse(readFileSync(p, 'utf8')); const classes = ['druid','hunter','mage','paladin','priest','rogue','shaman','warlock','warrior']; const trees = classes.flatMap(c => read('ui/core/talents/trees/' + c + '.json')); const nodes = trees.flatMap(t => t.talents); const db = read('assets/database/db.json'); const icons = new Map(db.spellIcons.map(s => [s.id, s.icon])); const missing = nodes.filter(t => !icons.get(t.spellIds[0])); const projection = Object.fromEntries(nodes.map(t => [t.spellIds[0], icons.get(t.spellIds[0])])); console.log({trees:trees.length,nodes:nodes.length,missingFirstIcons:missing.length,projectionBytes:Buffer.byteLength(JSON.stringify(projection))}); if (missing.length) process.exitCode = 1;"
```

The browser observations and dataset audit are the proof for this investigation. No build, lint or test suite was needed for unchanged application code.

## Reuse the data, not the legacy editor

The legacy route is:

`raw trees → rank spell ID → ActionId.fill → Database.getSpellIconData → spellIcons → ZAM image URL`.

Evidence:

- [`talents_picker.tsx:550–574`](../ui/core/talents/talents_picker.tsx): uses rank one for an unallocated node, otherwise the selected rank; delegates to ActionId.
- [`action_id.ts:829–842`](../ui/core/proto_utils/action_id.ts): icon URL is `https://wow.zamimg.com/images/wow/icons/large/{icon}.jpg`; metadata delegates to Database.
- [`database.ts:26–59,169–170,383–424`](../ui/core/proto_utils/database.ts): loads and indexes the whole UIDatabase; missing spell metadata can trigger an external tooltip request.
- [`gen_db/main.go:558–633`](../tools/database/gen_db/main.go): includes talent spell IDs in generated icon data.
- [`paladin.json:6–14`](../ui/core/talents/trees/paladin.json): Divine Strength is rank ID 20262, row 0, column 1, maximum five points. [`db.json:10879`](../assets/database/db.json) supplies icon `ability_golemthunderclap`.

**Recommended data route:** a small checked-in generated `talentIcons.json`, with a deterministic Node generator and a freshness check. Read all source data during generation/checking, not in the browser. Keep raw tree JSON as the sole authority for geometry, names and maximum ranks. No new Go fields, runtime metadata calls, icon service or whole-database client import.

Use the first-rank icon regardless of allocation. This is a deliberate presentation decision, not a replacement for rank-specific spell IDs. Explicit spell IDs can be nonsequential or nonmonotonic: Spiritual Focus is `[20205,20206,20207,20209,20208]` (`paladin.json:26–34`). Do not calculate rank links by adding to the first ID.

Do not import the legacy class wrappers, factory, renderer, Player, ActionId or Database. Those modules bring mutable talent normalization, editor rules, protobuf/player state, unrelated indexes and tooltip dependencies that this viewer does not need.

## Art and header feasibility

All 27 tree backgrounds are already referenced as remote `/images/wow/talents/backgrounds/tbc/{id}.jpg` URLs. `assets/talent_calc/` contains arrow PNGs, not bundled tree illustrations or node icons.

The sampled backgrounds are nearly square; the desired panels are portrait. Reuse the legacy presentation convention [`ui/scss/core/talents/_talents_picker.scss:156–162`](../ui/scss/core/talents/_talents_picker.scss): **`background-size: 100% 100%` plus an inset vignette**. This deliberately stretches the illustration to the tree instead of cropping it with `cover`. The reference is stylistic, not a photographic aspect-ratio requirement. Keep the full art and test the resulting visual treatment in the browser.

There is no complete canonical tree-header icon map. The legacy header image is commented out; its spec-index lookup is unsafe across classes (`talents_picker.tsx:203,290–292`). Druid has four simulator specs but three trees; Mage and several other classes have one simulator spec but three trees.

**Proposal:** use each tree's existing capstone talent icon as a representative header image, plus its explicit text name. All 27 capstone icons resolve in the measured data. These are representative artwork, **not claims about canonical specialization icons**. This avoids inventing or maintaining another 27-entry metadata catalog. Paladin examples: Divine Illumination / Avenger's Shield / Crusader Strike. The title remains Holy / Protection / Retribution.

## Details and accessibility

The current item tooltips already implement delayed pointer opening, immediate focus opening, hoverable panels, Escape dismissal, viewport positioning and cleanup on re-import/unmount. Reuse that behavior through a small shared tooltip interface; keep item and talent content renderers separate.

Exact existing files: [`tooltipController.js`](../ui-finder/src/lib/tooltipController.js), [`tooltipPosition.js`](../ui-finder/src/lib/tooltipPosition.js), [`ItemTooltipTrigger.svelte`](../ui-finder/src/lib/ItemTooltipTrigger.svelte), [`ItemTooltipLayer.svelte`](../ui-finder/src/lib/ItemTooltipLayer.svelte). Existing consumers are `App.svelte`, `GearSlot.svelte`, and `ReportView.svelte`.

Research constraints:

- [CSS Grid line-based placement](https://developer.mozilla.org/en-US/docs/Web/CSS/Guides/Grid_layout/Line-based_placement) directly supports sparse cells. Add one to the stored zero-based coordinates; do not create empty DOM placeholders or densely auto-pack nodes.
- [W3C CSS Grid reflow technique C32](https://www.w3.org/WAI/WCAG22/Techniques/css/C32) supports a CSS-only responsive layout. Preserve visual and DOM reading order.
- [Svelte style directives](https://svelte.dev/docs/svelte/style) support dynamic row/column placement and CSS custom properties in the existing framework.
- Native inspection buttons must remain operable at rank zero. [W3C button pattern](https://www.w3.org/WAI/ARIA/apg/patterns/button/). Allocation is not an operable toggle: no `aria-pressed`, no disabling unallocated talents.
- [WCAG Content on Hover or Focus](https://www.w3.org/WAI/WCAG22/Understanding/content-on-hover-or-focus.html) requires dismissible, hoverable and persistent additional content. Use keyboard parity, Escape and explicit tap activation; native `title` alone is insufficient.
- [WCAG Use of Color](https://www.w3.org/WAI/WCAG22/Understanding/use-of-color.html): numerical badges must carry state independently of lime borders.
- [Text contrast](https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html): normal-sized rank and tooltip text needs 4.5:1 contrast. [Non-text contrast](https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast.html): meaningful state/focus indicators need 3:1. Dim the image, not the entire node/badge/focus ring.
- [WCAG Reflow](https://www.w3.org/WAI/WCAG22/Understanding/reflow.html): target no page-level horizontal scrolling at 320 CSS px equivalent width. Stack trees rather than shrinking three panels onto a phone.
- [WCAG Target Size Minimum](https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum.html): AA minimum is 24 × 24 CSS px; proposed touch targets are at least 44 × 44 px.
- [W3C tabs](https://www.w3.org/WAI/ARIA/apg/patterns/tabs/) represent switchable panels, not static build metadata. The new imported-build chip is not a tab or button. Existing Gear/Stats/Talents navigation remains separate.

Full talent-effect descriptions are **not available** in the current tree data, `IconData` (`proto/ui.proto:232–238`) or import response. Proposed inspection content is name, tree, actual rank/max and allocation state. Do not synthesize effect text, scrape third-party calculators or add per-node tooltip requests for this visual redesign.

## Scope and data integrity

The existing documented product contract is a read-only imported talent viewer: [`docs/upgrade-finder.md:65–100`](upgrade-finder.md), [`2026-09-05 armory design`](superpowers/specs/2026-09-05-armory-redesign-design.md).

`character.spec` describes the simulator configuration, not reliably the largest tree. Mage may be `Mage` rather than `Fire`, for example (`cmd/wowsimcli/cmd/upgrades/import.go:60–100`). Use **Imported build** rather than manufacturing a dominant-tree specialization or a second build. Keep totals from the imported string; do not introduce editing, validation, clamping or talent-budget rules under a styling change.

The pinned legacy UI declares 61 player talent points (`ui/core/talents/talents_picker.tsx:14`), and the raw trees have nine tiers. Preserve those source definitions; no later-expansion tree conversion. Blizzard's [2021 Burning Crusade Classic Deep Dive](https://news.blizzard.com/en-us/article/23625673/world-of-warcraft-burning-crusade-classic-deep-dive) is useful edition context, not a complete machine-readable talent schema.

## Approaches considered

| Approach | Benefit | Cost / limitation | Decision |
| --- | --- | --- | --- |
| CSS-only reskin | Smallest change; fixes stacked layout and art treatment | Still text/pip tiles, no icons or meaningful inspection; misses the reference's main visual language | Insufficient alone |
| Targeted Svelte icon viewer | Reuses complete data, native Grid, current tooltip behavior; keeps imports/ranking unchanged | Needs a compact icon projection and a real second tooltip content type | **Recommended** |
| Embed/rebuild a talent calculator | Editing, prerequisite enforcement and saved builds could be added | Imports a much larger state model and dependencies; creates a baseline-mutation contract the user did not request | Reject for this scope |

No connector graph is required by the supplied reference. Keep prerequisite metadata intact but omit arrow rendering; add it only if dependency inspection becomes a requested feature. No mobile tree tabs, carousel, pan/zoom, new art pipeline or full spell-description service.

## Provenance limits

[`UPSTREAM.md`](../UPSTREAM.md) records the pinned simulator/database revisions. Repository source has an [MIT license](../LICENSE); that is not an asset-specific license for Blizzard artwork or third-party CDN content. Existing public URLs establish a technical route, not authorization to mirror or redistribute images. [U.S. Copyright Office FAQ](https://www.copyright.gov/help/faq/faq-fairuse.html) explains that online availability does not remove copyright protection. Preserve attribution/provenance; resolve rights before any new asset mirroring or redistribution. This investigation does not make a legal determination.

## Review conclusion

The layout and icons are feasible without backend changes or new dependencies. The material product choice is whether this remains a read-only imported-build display; this proposal preserves the existing contract. Approve the proposed design before application edits. Full spell descriptions, alternate builds and canonical header-icon curation are excluded explicitly, not represented as existing capabilities.
