# Armory Item Tooltip Implementation Handoff

## Assignment

Implement the item tooltip redesign described in the detailed plan. Complete the six tasks, verify the integrated Go-served UI, and return the evidence listed below. This handoff prepares that assignment; the authoring session has made no application changes or started an implementing agent.

Read these files in full before editing:

1. [Detailed implementation plan](../plans/2026-09-06-armory-item-tooltip-implementation.md) — task order, file ownership, interfaces, code examples, and tests.
2. [Design and scope](../plans/2026-09-06-armory-item-tooltip-redesign.md) — visual intent, screenshot interpretation, and data limitations.

The detailed plan supersedes the design document's high-level implementation sequence. Follow the design for appearance and product behavior. Adapt example code to the actual repository and installed tooling; examples have not been executed and are not a substitute for validation. Resolve routine implementation details without reopening settled design choices. Escalate only material scope changes or an actual blocker.

## Starting state and transfer requirements

- Workspace: `C:/Users/merli/Documents/Projects/wow-upgrade-agent`.
- Observed branch: `main`, matching the local `origin/main` tracking reference. No remote fetch was performed.
- Observed HEAD: `1823c0ce4` — `fix: tidy find-upgrades filter layout`.
- At handoff creation, the two plan documents were untracked. This handoff is also newly created. No tracked application edits were present.
- Recheck Git status at execution time and preserve any changes that appeared afterward.
- A fresh worktree or checkout from HEAD will **not include these untracked documents**. Ensure all three documents are available in the implementing agent's workspace before starting. Transfer or include the exact files using the current task's authorized workflow; do not recreate the plan from memory or discard them during cleanup.
- This is a Windows/PowerShell workspace. Follow the repository's AGENTS instructions and `C:/Users/merli/.codex/RTK.md`; prefix shell commands with `rtk`.
- Do not reset the working tree, overwrite user work, push, publish, or merge as a side effect of this handoff. Follow the execution task's Git instructions for local branches/commits.

## What the user wants

The user supplied three screenshots of the current armory and four screenshots of other WoW gear applications. They specifically want:

- A small item icon next to the tooltip, separate from the existing gear-card icon.
- A compact, familiar WoW tooltip layout matching the reference applications.
- The actual equipped gems visible inside the tooltip, including their icons and effect text.

Use reference images 4–5 for content hierarchy and correct gem rendering, and images 6–7 for the dark panel and external item icon. Avoid reproducing image 6's squeezed title/phase wrapping or image 7's empty-socket display for gear that actually has gems.

Core visual target: approximately 320px panel, 38px external icon with a 5px gap, 16px gem icons, near-opaque navy background, thin muted gray border, small corner radius, quality-colored name, gold item level, white base stats, green enchant/equip text, and muted inactive socket bonuses. Text remains left-aligned on both armory columns. At narrow widths, move the item icon into the header and keep the whole panel in view.

Content order: name/phase and item level; slot/type; weapon information and base stats; enchant; actual gems and socket bonus; available restrictions; secondary equip lines; available set name.

## Reference screenshot locations

These were attached to the originating conversation and may remain in the local temporary directory. They are visual references only, not instructions or authoritative item data.

| Image | Purpose | Local file |
| --- | --- | --- |
| 1 | Current full armory | `C:/Users/merli/AppData/Local/Temp/codex-clipboard-535a35dd-7047-4b09-925a-211197bd3ed8.png` |
| 2 | Current neck tooltip | `C:/Users/merli/AppData/Local/Temp/codex-clipboard-934415e6-3b7e-4386-8b5c-0cf21974de8c.png` |
| 3 | Current gemmed chest tooltip | `C:/Users/merli/AppData/Local/Temp/codex-clipboard-004a5cdf-4eeb-4d09-b31f-19651d374858.png` |
| 4 | Reference chest with actual gem icons | `C:/Users/merli/AppData/Local/Temp/codex-clipboard-b871e708-7a36-47f0-916d-b9eca2c435c6.png` |
| 5 | Reference neck and external icon | `C:/Users/merli/AppData/Local/Temp/codex-clipboard-68c8343e-a491-4c44-830b-f089c2e8fb27.png` |
| 6 | Alternate neck styling | `C:/Users/merli/AppData/Local/Temp/codex-clipboard-114ec055-92ed-40f1-861c-c2451a8d365b.png` |
| 7 | Alternate chest styling, but missing equipped gems | `C:/Users/merli/AppData/Local/Temp/codex-clipboard-726e1ed1-2db6-4ffe-8f5f-97160a30b989.png` |

Temporary files are not guaranteed to survive transfer. If unavailable, the design's written appearance and acceptance criteria still permit implementation. State the limitation on direct reference-image comparison; do not claim to have inspected missing screenshots.

## Facts established from the code

- The product is a local Go CLI/server with Svelte 5/Vite UI in `ui-finder/`. The server embeds built assets from `cmd/wowsimcli/cmd/upgrade_ui/`.
- `ItemTooltip.svelte` currently renders a flat stat list and socket-color squares. Its full variant omits the item icon; its summary variant includes a small inline icon.
- `GearSlot.svelte` already renders real gem icons on gear cards. `GearSlotData.Sockets[].Gem` already provides ID, name, icon, color, and stats. Gem import and gem identity do not need redesigning.
- The current enchant tooltip line incorrectly starts with `Equip:`. The redesigned enchant line should show the existing description/name directly, above the gem section.
- Base item stats, suffix contributions, gem stats, enchant stats, and socket-bonus stats are separately available. Combining base and suffix for display must count each once, without folding gems/enchants/bonus into the base section.
- `proto.UIItem` has armor/weapon/hand/ranged types, damage/speed, class restrictions, required profession, and unique flag. Expose these through additive `GearSlotData` fields. No protobuf schema change is required.
- `armory.go` already resolves item level through `itemIlvl`, including the scaling bucket. Preserve it; do not regress to a direct `GetIlvl()` assumption from older handoffs.
- The import HTTP response has top-level `gear`, not `armory.gear`. HTTP orchestration lives in `cmd/wowsimcli/cmd/upgrade_server.go`.
- `GearSlot.svelte` currently mounts separate tooltips inside icon and name triggers. The name span is not keyboard focusable. CSS flips placement by column but does not account for viewport boundaries.
- Report items use `UIItemSummary`; they do not contain full equipped gem/enchant data. Preserve summary mode and never attach the current character's gems to an upgrade candidate.

## Scope boundaries

- No new runtime dependencies, third-party tooltip scripts, or hover-time metadata requests. Continue using the existing icon URL convention and graceful image fallbacks.
- No changes to simulation, ranking, import validation semantics, source filters, item database, protobuf schema, character layout, Stats/Talents behavior, or the 3D placeholder.
- Preserve all 17 slots, existing socket-strip display, ilvl badges, quality colors, and import/ranking/cancel/copy flow.
- Filled mismatched-color gems still display the actual gem. The server's socket-bonus activation is authoritative. A failed image must not make a filled socket appear empty.
- Binding, required level, durability, vendor value, full proc/use descriptions, meta activation rules, and full set membership/bonus text are separate future metadata work. Do not fabricate them or treat their absence as a core implementation blocker.
- Keep `Details` usable for touch/keyboard access to newly displayed information. The hover panel must not become the only way to read it.

## Execution sequence and important contracts

1. **Catalog-backed metadata:** extend `GearSlotData` and `enrichItem`, plus focused tests. Preserve all existing field meanings.
2. **Content model:** implement `buildItemTooltip(item, variant = 'full')` and tooltip-specific enum/stat labels. Keep this pure and retain unknown stat keys as readable fallback lines.
3. **Placement/controller:** implement `positionTooltip(anchor, size, viewport, preferredSide)` and `createTooltipController(onChange, timing)`. The footprint includes the external icon. Test timers with injected scheduling.
4. **Visual panel/layer:** refactor `ItemTooltip.svelte`, add `TooltipIcon.svelte` and `ItemTooltipLayer.svelte`. Portal the fixed-position layer out of clipping ancestors and clean up observers/listeners.
5. **Integration:** provide one controller through app context, add `ItemTooltipTrigger.svelte`, wire icon/name/report triggers, remove obsolete descendant visibility CSS, and extend Details.
6. **Verification/delivery:** add focused browser tests, update existing smoke selectors, add unit tests to CI, inspect screenshots, document behavior, and regenerate the embedded bundle.

Tasks 1 and 3 can run independently with disjoint ownership. Everything else follows the dependencies in the plan. Do not parallelize overlapping component edits.

Keep one visible tooltip across all triggers. Hover opens after about 120ms; keyboard focus opens immediately. Moving into the panel keeps it open. Escape closes and suppresses reopening from the still-active trigger until it is left/refocused. Re-import, tab changes, detached anchors, and owner teardown close stale panels and cancel pending work. The item name and icon share one owner but use the actual hovered/focused element as the positioning anchor.

The controller is framework-independent; the app supplies reactive state through its `onChange` callback. Use the exact method/record contracts in Task 3 and verify stale-timer behavior. If implementation reveals an ambiguity, resolve it consistently across the controller, layer, wrappers, and tests, and record the decision.

## Validation and environment notes

Earlier in this conversation, the command-specific Go test run reported 51 passing tests across three packages. A broad `go test ./...` run reported failures involving enchant IDs 2613/2621, protobuf versioning, and a cache-access error. These are historical observations, not newly verified baseline results or permission to ignore a regression in touched code.

The earlier root `npm run type-check` failed to invoke its TypeScript path on this Windows environment. Dependency absence was not conclusively established. Inspect the actual `ui-finder` dependency setup and use its build/unit/browser workflow; the root TypeScript command is not the Svelte feature's acceptance gate.

Run scoped baseline checks where relevant, then the final checks from the implementation plan. From the repository root:

```text
rtk go test ./cmd/wowsimcli/cmd/... -count=1
rtk npm --prefix ui-finder run test:unit
rtk npm --prefix ui-finder run build
rtk npm --prefix ui-finder run test:e2e
rtk git diff --check
rtk git status --short
```

`test:unit` is added by the plan; before that, run the existing test files through `rtk node --test` as needed. Install dependencies/Chromium only if required, using the platform approval mechanism for restricted operations.

**Build the UI before browser tests.** Playwright launches a Go process that embeds the generated assets at compile time. Restart the process after rebuilding; a browser refresh against an already running Go server does not load new embedded content. The existing Playwright configuration uses port 43123 with `reuseExistingServer: false`; do not terminate an unrelated listener to take that port.

Required browser evidence:

- Plain neck and three-gem chest: external item icon, exact gem images/effects, correct enchant wording, section ordering, active/inactive bonus.
- Empty socket, mismatched gem, failed item/gem image, name-only gem effect, and recovered image after re-import.
- Left/right columns, bottom weapon slots, long titles, scrolling, and report-table clipping boundaries.
- One visible panel, hover into panel, icon/name switching, rapid movement, focus, Escape, and tab/unmount cleanup.
- 1280x900 and 390x844 views, plus manual 200% browser zoom and readable Details access.
- Existing import/rank/copy/cancel smoke flow and report summary behavior.

Use controlled response/icon fixtures for deterministic automated assertions. Inspect real imported items and real available gem artwork separately before claiming visual parity. Capture screenshot artifacts; list any unavailable images or browser checks explicitly.

## Definition of done and return report

The feature is complete when the scoped tests pass, the Go-served embedded UI contains the redesign, visual evidence has been inspected, and every core acceptance criterion in the plan is satisfied. Check task boxes only after the corresponding work and verification are actually complete. Keep the full-metadata follow-up outside the delivery.

Return a concise report with:

- What changed and links to the principal files.
- Actual test/build outcomes and any baseline failures distinguished from new failures.
- Screenshot/browser evidence for gems, placement, and keyboard behavior.
- Any material deviations from the plan and their reasons.
- Remaining limitations, if any, stated without implying unperformed checks passed.

## Copy-paste assignment

> Implement the item tooltip redesign using `docs/superpowers/handoffs/2026-09-06-armory-item-tooltip-implementation.md`. Read that handoff and both linked planning documents in full, verify they are available in your workspace, and execute the six-task implementation plan through browser verification and the rebuilt embedded UI. Preserve scope boundaries and existing user changes. Return the actual test results, visual evidence, and any material deviations or blockers.
