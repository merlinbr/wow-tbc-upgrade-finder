# Talent tree redesign design

## Status and goal

**Proposed; awaiting user review.** No application changes are authorized by this document alone.

Make the existing Talents area resemble the supplied reference: three tall illustrated trees, sparse square talent icons, lime allocation borders and small numerical rank badges. Preserve the single imported build, its exact ranks, and the existing Gear/Stats/Talents navigation.

Evidence: [research and before screenshots](../../talent-tree-redesign-research.md). Execution proposal: [implementation plan](../plans/2026-09-07-talent-tree-redesign.md).

## Recommended composition

1. Keep the character header and armory navigation unchanged.
2. Place one compact, centered **Imported build** summary above the trees. Show a static **Read-only** badge and the three tree totals in canonical order. For the verified Paladin fixture: `5 / 11 / 45`, with accessible labels naming each tree. This is not a selectable build or a live region.
3. On a wide talent container, display all three trees side by side. Each panel has a dark translucent header with a representative tree icon, tree name, and allocated total.
4. Let the illustrated background fill the full tree body. The cells themselves are transparent except for the square node artwork/frame and badge; there are no visible empty-cell boxes.
5. Preserve exactly four columns and nine tiers from the source definitions, including all intentional gaps.

### Geometry and art

- Center the board at a maximum width of **936 CSS px**.
- Three-column layout requires at least **888 px of available talent-container width**, not viewport width. Use a CSS container query; no JavaScript size observer.
- At that width: three panels at least 280 px wide, separated by 24 px gaps; 52 px square nodes, 12 px row/column gaps, approximately 16 px horizontal body padding. Header target height is 56 px. Nine rows produce a roughly 650 px tall panel, close to the reference.
- Below the threshold: one centered panel per row, maximum width 360 px; keep canonical tree order. Use 44 px nodes and 10 px gaps by default, which fit inside the existing armory padding at 320 px viewport width. Do not squeeze three trees into one row, create mobile tree tabs, or change rank positions.
- The 300 × 331 sampled backgrounds intentionally stretch to the portrait body using `background-size: 100% 100%`, matching the legacy talent renderer's convention. Add a restrained dark overlay and inset edge vignette. Do not use `cover`, which would crop the illustrations heavily in this geometry.
- Reuse the existing remote TBC background URLs. When an image fails, keep a dark gradient backing and all node information visible. No new asset mirroring or redistribution is part of this proposal.

These dimensions are implementation targets; contrast and overlap must be verified against actual rendered artwork and zoom, not assumed from hex colors.

## Node appearance and interaction

| State | Visual treatment | Meaning |
| --- | --- | --- |
| Zero rank | Desaturated, dimmed artwork; restrained readable frame; `0/M` badge | No points allocated; still inspectable |
| Partial rank | Full-color artwork; lime border; `N/M` badge | Some points allocated |
| Maximum rank | Full-color artwork; lime border with a slightly stronger inner highlight; `M/M` badge | Maximum rank reached |
| Keyboard focus | Separate high-contrast offset outline | Inspection focus, independent of allocation |

Use a lime near `#c7ef31` as the initial allocation color; keep it local to talents rather than changing global button/accent tokens. Badge backgrounds are opaque dark surfaces. Do not dim the entire node: the rank, focus ring and accessible controls remain readable. Numerical ranks communicate state without color.

Each node is a native `button type="button"` whose accessible name includes the talent name and actual rank/max. Its sole action is **inspect**, never allocate. Rank-zero nodes are not disabled and allocation is not represented by `aria-pressed`.

- Pointer hover opens details using the existing approximately 120 ms delay.
- Keyboard focus opens the same details immediately. Enter/Space and click/tap explicitly open inspection, including reopening after Escape.
- The popup remains open while the trigger has focus or while the pointer is over trigger/popup. Tap inspection therefore persists during scrolling while focus remains on the node.
- Escape dismisses; moving focus away or leaving the hover region follows the existing close lifecycle. Switching tabs or re-importing releases the previous popup and pending timers.
- No interactive controls are placed inside a `role="tooltip"` popup. Keep focus on the node.

Details contain the talent's `fancyName`, tree name, actual rank/max, and `Unallocated` / `Allocated` / `Maximum rank` state. Full spell-effect descriptions are unavailable locally and are not fabricated. No Wowhead tooltip script or per-node metadata request.

Use existing `TooltipIcon.svelte` for node/header artwork, wrapped decoratively when the parent button/header already provides the name. Its existing image-error fallback avoids broken-image placeholders. A tree's capstone icon is representative header artwork, not a claim to be a canonical specialization icon; this explicitly differs from exact reference header artwork.

## Data and module interfaces

### Talent viewer

`TalentTrees.svelte` keeps its current interface:

```text
{ class: playerClass, talentsString }
```

It still consumes the nine raw tree JSON files and `decodeTalentsString`, `rankAt`, `treePoints`. Add only the generated first-rank spell-ID/icon map. Keep those helpers' decoding semantics unchanged.

Iterate each tree's real talent array directly, preserving original indices. Place each node at `location.rowIdx + 1` / `location.colIdx + 1`. Remove the coordinate Map, dense `cells` construction, null placeholders, visible names and pip markup that the old renderer required. There is no separate graph/view-model framework.

All 27 current arrays are already row-major. Do not sort the serialized array or derive rank offsets from coordinates. Tree totals are from the existing decoder, not a new talent budget calculator. Preserve current behavior for missing/unknown data and do not silently clamp imported ranks under a visual refactor.

### Static icons

New files:

- `ui-finder/scripts/generate-talent-icons.mjs`: deterministic projection from the nine raw tree files and `assets/database/db.json.spellIcons`; supports `--check` without writing.
- `ui-finder/src/lib/talentIcons.json`: checked-in generated spell-ID → icon-label map.

Missing source icons cause generation/check failure with the spell ID and talent name. Runtime image failure still uses the existing fallback. Regenerate the projection when pinned tree/database data changes; do not ship the 3 MB database to obtain approximately 20 KB of labels.

### Shared tooltip seam

The controller and positioning algorithm already solve the hard lifecycle behavior. Generalize the existing item-only trigger/layer rather than duplicating them:

```text
TooltipTrigger props:
  content = { kind: 'item', item, variant: 'full' | 'summary' }
          | { kind: 'talent', talent: { name, treeName, points, maxPoints } }
  preferredSide = 'right' | 'left'
  children = existing trigger event snippet, plus onactivate

Active record:
  { owner, id, content, preferredSide, anchor }
```

- Rename `ItemTooltipTrigger.svelte` → `TooltipTrigger.svelte` and `ItemTooltipLayer.svelte` → `TooltipLayer.svelte`; rename the shared context constant to `TOOLTIP_CONTEXT`.
- Keep `createTooltipController` and `positionTooltip` behavior; neither needs a talent-aware state machine.
- `TooltipLayer` selects `ItemTooltip` or a small new `TalentTooltip.svelte` by `content.kind`. Item formatting remains item-specific; do not encode talents as fake items or add an extensible renderer registry.
- Migrate all callers in `App.svelte`, `GearSlot.svelte` and `ReportView.svelte`; no obsolete files, compatibility aliases or duplicate tooltip layers.
- Existing item tooltip IDs may retain their current format if useful; IDs must be unique and `aria-describedby` must resolve when open. Exact ID text is not an observable requirement.

This is a real two-content seam, not a speculative tooltip framework. Keep header/board markup in `TalentTrees.svelte`; do not split tree/card wrappers that only pass props through.

## Explicit non-goals

- Talent allocation/editing/reset, legal-build validation, save/export, presets, alternate/dual builds, comparison or baseline mutation.
- New Go API fields, import schema changes, simulation/ranking changes, protobuf changes or new dependencies.
- Full spell descriptions, effect inference or runtime talent metadata requests.
- Connector arrows, a graph library, canvas, pan/zoom, a carousel or mobile-only tree navigation.
- A new canonical 27-tree header-icon catalog or asset scraping/mirroring.
- Unrelated character header, Gear/Stats tab or ranking redesign.

## Acceptance and verification

1. Import the existing Retribution fixture through the real form. All 64 Paladin talents appear in their original cells; totals remain 5 / 11 / 45. Clicking, tapping, hovering, focusing and keyboard activation never alter ranks, imported settings or start ranking.
2. At 1280 px viewport width, all three portrait panels share a row. At 978 px (the reference width), verify the actual container meets the 888 px threshold and the composition fits. At 390 and 320 px, trees stack without page horizontal overflow or clipped focus rings/badges. Verify 400% zoom-equivalent reflow.
3. Rank-zero, partial and full nodes remain visually distinct with readable numerical badges. Measure normal text at 4.5:1 and necessary focus/state boundaries at 3:1 against their real backing; test image failure too.
4. Mouse, keyboard and touch all expose identical supported detail content. Check Escape, pointer travel into the popup, explicit reactivation after Escape, viewport edges, scrolling, re-import and Talents → Gear cleanup.
5. Re-import the existing Mage fixture while inspection is open: no stale Paladin content, icons or pending popup. Gear and report item tooltips retain their content and lifecycle after the shared rename.
6. The generator's `--check` covers every first-rank talent icon and detects stale/missing metadata. Preserve existing rank-decoder tests; add only lifecycle/ordering regressions that can fail on plausible bugs.
7. Run targeted unit/E2E checks and a Vite production build after implementation; use a real browser for visual proof. Rebuild the Go embedded UI binary only after the final frontend build when checking the packaged surface.

## Review gate

Approve or amend this direction before implementation. The default is a faithful **read-only icon-tree viewer**, not a new talent calculator. Research and draft planning are complete; visual fidelity of the redesign remains to be proven when it is built.
