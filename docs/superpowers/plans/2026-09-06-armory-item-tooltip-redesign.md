# Armory Item Tooltip Redesign Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox syntax for tracking.

**Status:** Proposed design and implementation sequence. Requested work is planning; no application implementation is included.

**Detailed implementation:** [Armory Item Tooltip Implementation Plan](2026-09-06-armory-item-tooltip-implementation.md) supersedes the high-level task sequence below with interfaces, code examples, tests, and integration checkpoints.

**Goal:** Give equipped-item tooltips the compact WoW presentation shown in reference images 4–7, including a small external item icon and the actual equipped gems.

**Architecture:** Keep imported equipment as the source of truth. Separate tooltip content formatting from its Svelte presentation and shared hover/focus positioning. Add only catalog-backed metadata to the armory response; richer reference details have a separate data-enrichment scope.

**Tech Stack:** Existing Go API, Svelte 5, Vite, Node test runner, Playwright. No remote tooltip service is required.

## Global constraints

- Preserve imported gear, gems, enchants, simulator stats, and ranking behavior.
- Render actual equipped gems, even when their colors do not match their sockets.
- Keep item base stats, random-suffix stats, gem effects, enchants, and socket bonuses distinct; never add them twice.
- Missing metadata must not become invented values. Item level is not required character level; character level is not item required level.
- Keep hover rendering immediate and local; use existing icon URLs with accessible image fallbacks.
- Treat the seven user screenshots as visual references, not sources for item database values.
- Preserve report summary tooltips while giving them the shared visual shell; full candidate tooltips are a separate scope because report data is only a summary.

## Recommended approach and alternatives

**Recommended: shared native tooltip with catalog-backed content.** Reproduce the hierarchy of images 4–5 and the dark panel/external icon of images 6–7. This supports imported gems directly and lets us control keyboard behavior, clipping, and fallbacks.

**Smaller alternative: restyle the existing component only.** Fastest way to add the two kinds of icons, but CSS-only positioning and duplicated tooltip instances remain, and richer content becomes harder to maintain.

**Larger alternative: full reference-content parity.** Include binding, level requirements, prices, proc descriptions, and complete set sections. This requires a verified metadata source and bundling work beyond the existing response. It should follow the core redesign rather than determine its first delivery.

## Proposed appearance

- Body approximately 320px wide on desktop, capped by available viewport width; allow long item names to wrap naturally.
- Separate 38px item icon aligned with the panel's upper-left corner, 5px outside its border. Include the icon in placement calculations. On very narrow viewports, move it into the header.
- Deep navy background around `#121526`, nearly opaque, with a thin muted gray border, 3px corners, and restrained shadow.
- Body text 13–14px, line height around 1.25; item name around 15px. Final values should be checked against the supplied screenshots.
- Quality-colored item name, gold item level, muted phase aligned right on one unbroken line. Prevent phase from squeezing the title into an excessively narrow column as in image 6.
- White base stats, bright green enchant/equip lines, muted inactive bonuses. Use modest 8px gaps between content groups, not large cards or dividers.
- Gem rows show 16px actual gem artwork followed by effect text. Use green effect text consistently; do not invent gem rarity coloring because the current gem response has no quality field.

Content order:

1. Item name / phase, then item level.
2. Slot on the left and armor or weapon type on the right.
3. Weapon damage and speed when applicable; armor, then primary stats in explicit order: Strength, Agility, Stamina, Intellect, Spirit. Preserve other base stats.
4. Applied enchant in green, for example `+6 All Stats` or `Mongoose`, without the current misleading `Equip:` prefix.
5. Actual socketed gem icons and effects, followed by the socket bonus in its supplied active/inactive state.
6. Catalog-backed class/profession restrictions and unique flag when available.
7. Secondary stat effects, using tooltip wording such as `Equip: Improves melee critical strike rating by 31.` Preserve distinctions between melee, ranged, and spell stats; only combine equal values when verified as one shared rating.
8. Available set name in gold. Full set membership and bonus text belong to the later enrichment scope.

Empty socket: display an outlined socket-color symbol and `Red Socket (empty)` or equivalent. A filled socket with a failed image still displays the gem's effect and accessible name; it must never look empty. Gem effect text falls back to its name when the stats map is empty. Non-stat meta-gem effects and activation requirements require additional metadata and must not be guessed.

## Current implementation evidence

- `ui-finder/src/lib/ItemTooltip.svelte` renders a flat `Object.entries(stats)` list, colored socket dots, and an enchant at the bottom prefixed with `Equip:`. Only the report summary variant currently renders an item icon.
- `ui-finder/src/lib/GearSlot.svelte` already renders real gem icons on equipment cards. It creates separate full tooltips for the item icon and name; the name trigger is a span and is not keyboard focusable by default.
- `ui-finder/src/app.css` controls visibility with hover/focus selectors and flips tooltips only according to the left/right equipment column. It does not check viewport edges.
- `cmd/wowsimcli/cmd/upgrades/types.go` already exposes gem ID/name/icon/color/stats and socket-bonus activation. No gem-import redesign is necessary.
- `cmd/wowsimcli/cmd/upgrades/armory.go` uses base item stats and separately resolved suffix, gem, and enchant data. Keep that separation.
- `proto/ui.proto` already contains armor/weapon types, damage, speed, class restrictions, profession, unique flag, and set identifiers, but those fields are not all exposed through `GearSlotData`.
- Binding, durability, required level, price, and readable set-bonus descriptions are absent from the current armory response. `sim/core/item_sets.go` represents bonus behavior as functions, not display descriptions.

## Task 1: Normalize tooltip content and expose available metadata

**Files:** Modify `cmd/wowsimcli/cmd/upgrades/types.go`, `armory.go`, and `armory_test.go`; create `ui-finder/src/lib/itemTooltip.js` and `itemTooltip.test.js`.

**Interface:** `buildItemTooltip(item, variant = 'full')` returns `{ name, icon, quality, phase, ilvl, slotLabel, typeLabel, weaponLines, baseLines, enchantLine, sockets, socketBonus, restrictionLines, equipLines, setName }`. Each socket retains `{ color, gem }`; never infer gem identity from socket color. Summary mode returns only supported summary content.

- [ ] Extend `GearSlotData` with catalog-backed fields: `armorType`, `weaponType`, `handType`, `rangedWeaponType`, `weaponDamageMin`, `weaponDamageMax`, `weaponSpeed`, `classAllowlist`, `requiredProfession`, and `unique`. Resolve these from the existing UIItem getters without changing protobuf definitions or simulation requests.
- [ ] Add Go assertions for armor, weapon damage/speed, restrictions, populated gem icon/ID, empty sockets, and unchanged imported settings after enrichment. Use cloned catalog fixtures for deterministic assertions.
- [ ] Implement an explicit stat-order table and separate white base lines from green secondary-stat lines. Format armor without a leading plus. Unknown stat keys remain visible using the existing formatter instead of being discarded.
- [ ] Merge random-suffix stat contributions into item display lines once, while retaining the suffix label when supplied. Do not fold gem/enchant/socket-bonus stats into those lines.
- [ ] Add focused Node tests for stable ordering, preservation of unknown stats, suffix values counted once, enchant wording, filled versus empty sockets, and summary inputs with missing full fields.
- [ ] Run `rtk go test ./cmd/wowsimcli/cmd/upgrades -count=1` and `rtk node --test ui-finder/src/lib/itemTooltip.test.js`.

## Task 2: Build the reference-style content panel

**Files:** Modify `ui-finder/src/lib/ItemTooltip.svelte`; create `ui-finder/e2e/item-tooltip.spec.js`.

**Consumes:** `buildItemTooltip(item, variant)` from Task 1. **Produces:** A content panel with no trigger-relative positioning or independent open state.

- [ ] Implement the external item icon, constrained header, dense typography, content ordering, and dark border/background specified above.
- [ ] Render each actual gem's image and effect. Give each image the gem name as accessible text; preserve readable fallback text after an image error. Reset image failure state when item/gem identity changes on re-import.
- [ ] Use supplied socket-bonus activation and distinguish empty sockets from filled sockets with unavailable art.
- [ ] Add browser assertions against a controlled gemmed chest fixture: three sockets produce three actual gem images with expected icon URLs; the enchant is above the gems and is not prefixed `Equip:`; bonus state follows the payload.
- [ ] Cover a plain neck item, weapon, empty socket, broken icon, long item name, and a gem with only a name/no stat lines. Stub icon responses for deterministic screenshots and avoid relying on network availability.

## Task 3: Share accessible triggers and viewport positioning

**Files:** Create `ui-finder/src/lib/ItemTooltipTrigger.svelte` and `tooltipPosition.js`; modify `GearSlot.svelte`, `ReportView.svelte`, and `app.css`; extend `item-tooltip.spec.js`.

**Interface:** `ItemTooltipTrigger` accepts `{ item, variant = 'full', preferredSide = 'right', children }`, where `children` is a Svelte snippet receiving trigger handlers and the tooltip ID. It owns visibility, a unique ID, anchoring, and cleanup. `positionTooltip(anchorRect, panelSize, viewport, preferredSide)` returns `{ left, top }` for the complete icon-plus-panel footprint, keeping an 8px viewport margin.

- [ ] Replace CSS-only visibility with shared hover/focus state. Open after about 120ms on mouse hover, immediately on keyboard focus, and close after about 120ms when both anchor and panel are left.
- [ ] Make item-name triggers keyboard focusable with visible focus styling, and connect each trigger using `aria-describedby` to its tooltip. Avoid placing interactive controls inside the tooltip. Ensure only one tooltip is visible when moving between item icon, name, and neighboring items.
- [ ] Support Escape dismissal and retain that dismissal until the trigger is left or focus changes, avoiding immediate reopening. Keep the panel open while hovered so text can be inspected.
- [ ] Render the floating panel outside clipping ancestors, using fixed coordinates. Prefer the column's inward side, flip if necessary, then clamp horizontally and vertically. Recompute for resize, scrolling ancestors, and content/image size changes; remove listeners/observers on close/unmount.
- [ ] For oversized content, constrain height to the viewport and allow pointer scrolling. Keep full item details available through the existing disclosure so the tooltip is not the sole access path. On touch, preserve access through Details.
- [ ] Update existing armory/report tests to locate the visible tooltip by its accessible relationship rather than expecting it to be a descendant of the trigger.
- [ ] Verify all viewport edges, both armory columns, bottom weapons, scrolling, 200% zoom, keyboard focus, Escape, rapid switching, and re-import after image failure. Test the entire icon-plus-panel bounding box stays in the viewport.

## Task 4: Verify visual parity and rebuild the embedded UI

**Files:** `ui-finder/e2e/item-tooltip.spec.js`, `.github/workflows/upgrade_finder.yml`, `docs/upgrade-finder.md`, and generated `cmd/wowsimcli/cmd/upgrade_ui/` assets.

- [ ] Compare a gemmed chest and plain neck against images 4–5 for hierarchy and images 6–7 for panel/icon placement. Inspect the actual gem art, title wrapping, left alignment on mirrored slots, border contrast, and section spacing.
- [ ] Capture deterministic browser screenshots at 1280×900 and 390×844, including a near-edge tooltip. Add focused visual baselines for the neck and gemmed chest rather than snapshotting the entire ranking workflow.
- [ ] Add the new Node test file to CI alongside existing UI tests; do not rely on the unrelated root TypeScript check as validation of this Svelte app.
- [ ] From the repository root, run `rtk node --test ui-finder/src/lib/itemTooltip.test.js ui-finder/src/lib/labels.test.js ui-finder/src/lib/api.test.js ui-finder/src/lib/talents.test.js` and `rtk go test ./cmd/wowsimcli/cmd/... -count=1`.
- [ ] From `ui-finder`, install dependencies if absent, then run `rtk npm run build` followed by `rtk npm run test:e2e`. Build first because Playwright starts the Go server with the embedded UI.
- [ ] Update the tooltip documentation and include the rebuilt assets with the implementation. Review the final diff for unrelated changes.

## Separate follow-up: complete reference metadata

This is the remaining work for full information parity with images 4 and 7, not a promise that the current catalog already supports it.

1. Audit existing bundled inputs for verified TBC binding, required level, vendor value, maximum durability, item effect text, gem non-stat effects/requirements, set pieces, and set-bonus descriptions. Record coverage and provenance by item/gem/set ID. If unavailable, select a versioned source before expanding the import pipeline.
2. Package available text/metadata locally and expose optional structured fields through the armory response. No fetching third-party HTML at hover time and no inferred binding or level values.
3. Build set membership using verified set identities and the current imported equipment. Highlight equipped pieces and active thresholds; do not equate the number of catalog rows with the true set size.
4. Render readable equip/use effects and set sections without duplicating numeric stat lines already shown. Distinguish maximum durability from current durability, which the import does not provide; do not display an invented `165 / 165`.
5. Preserve missing-field fallbacks and label incomplete descriptive coverage in the expanded Details view when relevant. Show meta activation only if evaluated from verified requirements and the imported gems.

## Acceptance criteria for the core redesign

- Hovering or focusing an equipped item shows a small item icon beside a compact WoW-style panel.
- The tooltip displays the imported gems' actual artwork and effects, including mismatched-color gems; genuinely empty sockets remain visibly empty.
- Enchants, base stats, secondary effects, and socket bonuses have separate, predictable positions and correct labels.
- Names wrap without crushing the phase label; right-column text remains left-aligned inside the panel.
- No edge clips the icon or panel; mouse, keyboard, Escape, and re-import behavior are covered.
- Summary report tooltips still work and never imply candidate gems that are absent from their payload.
- The Go-served embedded build contains the redesign and the targeted tests pass.
