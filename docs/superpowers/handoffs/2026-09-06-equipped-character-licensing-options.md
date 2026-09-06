# Equipped Character License / Alternatives Investigation Handoff

## Assignment

Resolve the authorization question blocking the equipped 3D character
preview, or identify an alternative renderer/asset path that is legally
usable, and return a researched verdict with sources. **Research only: no
application code changes** from this investigation. The implementation is
complete, tested, and gated; only the rights question stands between it and
being enabled by default.

Read these files in full before anything else:

1. [Feasibility experiment account](../experiments/2026-09-06-equipped-character-feasibility.md) — what was built, what was verified, the exact transport and integration contract.
2. [Implementation plan](../plans/2026-09-06-equipped-character-preview.md) — milestone gates and acceptance criteria (milestone 1 is done; the authorization gate is the remaining one).
3. [Reference-sites research](../../3d-reference-sites-research.md) — ZAM proxy pattern, Sixty Upgrades pipeline, per-site rights evidence.
4. [Armory redesign research](../../armory-redesign-research.md) — the original permission analysis (ZAM terms sections 4/8, Blizzard API terms, alternatives).
5. [Upgrade Finder docs](../../upgrade-finder.md), section "Equipped 3D character preview" and [STATE.md](../../STATE.md) boundaries — the operator-facing gating statements.

The last four are tracked in commit `14bcbdfe5`; the first four files are also
tracked. Recheck Git status and HEAD at execution time; the workspace is
`C:/Users/merli/Documents/Projects/wow-upgrade-agent`, branch `main`, origin
`https://github.com/merlinbr/wow-tbc-upgrade-finder.git`.

## What the user wants

The 3D preview works and ships behind `--enable-3d` (opt-in). The user wants
an answer to: *can we actually use the ZAM/Wowhead viewer + model assets (and
on what terms), or is there a legally usable alternative?* They want to hand
this question to a dedicated agent — the deliverable is the answer and a
recommendation, not code.

## State of the feature (so the agent does not redo it)

- Renderer: ZAM/Wowhead **live** viewer distribution
  (`https://wow.zamimg.com/modelviewer/live/viewer/viewer.min.js`,
  266,932 bytes) driving the **TBC content tree**
  (`https://wow.zamimg.com/modelviewer/tbc/...`). The dedicated
  `tbc/viewer/viewer.min.js` distribution requests `.mo3` files that do not
  exist in the TBC tree and is **not** compatible.
- Verified in Chromium from a loopback origin: equipped characters render
  (mesh/skin/texture/animation), drag rotation, resize, and destroy all work.
  Same-origin transport is required — direct browser-to-`wow.zamimg.com`
  fetches fail CORS. The Go server proxies the fixed path
  (`/visuals/zam/modelviewer/tbc/**`), gated by `--enable-3d`.
- Browser integration requires jQuery 1.x (vendored, MIT), `window.WH`
  shims, and handling the constructor's thenable; the character lives in
  `viewer.renderer.actors`; `actor.setAnimPaused` is a no-op on this
  revision (pause is implemented as destroy/recreate); `renderer.distance`
  (fit margin ×1.4) and reset-to-captured-angles are needed for correct
  framing.
- Item display mapping comes from `https://www.wowhead.com/tbc/item={id}&xml`
  (server-side; gives `<icon displayId>` + `<inventorySlot id>`), with a
  chest/robe ambiguity resolved by probing the ZAM meta tree
  (`meta/armor/5/...` vs `meta/armor/20/...`), and positional weapon slots
  (main hand 21, off hand 22, ranged 15). Relics/jewelry have no display and
  are reported as an honest partial preview.
- Code surface of interest if the agent needs it: `ui-finder/src/lib/characterViewerAdapter.js` (provider adapter — the seam where an alternative provider would plug in), `characterPreview.js` (pure mapping), `cmd/wowsimcli/cmd/upgrade_visuals.go` (proxy + resolver, flag-gated).

## The blocker, precisely

- The ZAM viewer and model assets are not a public SDK; the published
  `viewer.min.js.LICENSE.txt` URL 404s, and ZAM's terms
  (https://corp.fanbyte.com/legal/terms, sections 4 and 8 per the earlier
  research) restrict copying/distribution and display/mirroring/framing
  without consent. No supported third-party SDK grant was established.
- The app already hotlinks ZAM **icon images** (existing behavior, not a
  precedent for full model/mesh/texture redistribution).
- Nothing in the existing research constitutes a license. The prior research
  explicitly rejected building transport to evade provider restrictions.

## Options to investigate (with what is already known)

1. **Explicit arrangement with Wowhead/ZAM.** Determine who to contact
   (Fanbyte/ZAM legal), what permission wording is needed for: (a) loading
   the viewer script and TBC model assets from `wow.zamimg.com` from a
   local, non-commercial, open-source app via a loopback proxy; (b) whether
   any existing embed/partner mechanisms exist (e.g., Wowhead's dressing
   room iframe/embed), and on what terms. Draft the inquiry if that is the
   recommended path.
2. **Blizzard first-party profile APIs (official, documented).** The Classic
   profile reference lists `/appearance`, `/equipment`, `/character-media`
   and the `profile-classicann-{region}` namespace for Burning Crusade
   Classic Anniversary. This gives a **static full-body image** of a named
   character (plan: an accepted static fallback is permitted only as a
   separately accepted compromise, never presented as 3D). Needs
   operator-owned credentials (never in browser/binary); a local Go config is
   possible. Verify: does `character-media` return a full-body render for
   Anniversary, and what terms attach to it? Note: it cannot show local
   hypothetical replacement gear, and live gear must never replace imported
   gear.
3. **Alternative hosted renderers with permissive terms.** Search for
   TBC-capable character renderers or viewers with explicit embedding
   grants, open providers, or data licenses. Known candidates with problems:
   Sixty Upgrades (site-bundled code + Cognito-signed GraphQL; their use
   license restricts copying/modification, public display and mirroring;
   wildcard CORS is not a reuse license; not a public SDK), Miorey
   `wow-model-viewer` (WotLK/Retail lineage, external globals, TBC unproven),
   Google `model-viewer` (needs assembled glTF — no WoW assembly pipeline),
   WMVx (desktop exporter, TBC 2.4.3, FBX, particle/texture-animation
   limits; an export tool, not a browser component).
4. **Self-hosted conversion (wow.export / WMVx pipelines).** Technically the
   largest alternative: converted geometry, textures, animation from Blizzard
   client files, self-hosted, potentially offline. The question is rights:
   is conversion for a local non-commercial tool within Blizzard's personal
   use terms, and what about redistributing converted assets or relying on
   user-supplied client data? Investigate documented positions, not
   assumptions.
5. **Deferral alternative.** Keep the gate exactly as is (documented operator
   gate) — only if the investigation definitively shows no path, state that
   with the evidence.

For every option give: status now (usable today / needs authorization /
not viable), exact requirements (who to ask, credentials, terms to sign),
the evidence and source, and the honest product outcome (interactive
equipped 3D vs static image vs nothing). Cite primary sources (legal pages,
official API docs, source code, READMEs) — follow each claim to its source;
no secondary write-ups as proof.

## Constraints

- Do not change the shipped gating, write proxy workarounds to evade
  restrictions, or treat any fetching mechanism as permission.
- Do not recommend embedding credentials in the distributed binary or
  browser.
- Do not recommend replacing the imported (simulation) gear with live
  character gear, or representing a static image as an interactive 3D
  preview.
- No application code changes, no commits, no pushes. The verdict report is
  the deliverable.
- If a new provider path is recommended, note how it maps onto the existing
  adapter seam (`characterViewerAdapter.js` factory surface: `create`, 
  `rotate`, `resetView`, `resize`, `destroy`) so the Svelte/Go layers need
  no redesign.

## Return report

- Verdict per option (table preferred): usable now / needs authorization /
  not viable + why, with sources.
- The single recommended path and the exact next action (contact, credentials
  setup, or provision), including what the user must decide.
- Anything the previous research got wrong or left unresolved.
- A short "what the code would need if we switch providers" note, scoped to
  the adapter seam, without implementing it.

## Copy-paste assignment

> Investigate the license/authorization blocker for the equipped 3D
> character preview in `docs/superpowers/handoffs/2026-09-06-equipped-character-licensing-options.md`
> (workspace `C:/Users/merli/Documents/Projects/wow-upgrade-agent`, HEAD
> `14bcbdfe5`). Read the handoff and the four referenced documents in full:
> feasibility experiment, implementation plan, reference-sites research,
> armory redesign research, plus the upgrade-finder/STATE gating statements.
> Answer whether the ZAM/Wowhead viewer + TBC assets can be used and on what
> terms, or identify a legally usable alternative (including Blizzard's
> official Classic profile APIs for a documented static fallback, hosted
> providers, and self-hosted conversion paths), with primary-source evidence.
> Research only — no code changes, no commits. Return the per-option verdicts,
> a single recommendation with the exact next action, and a provider-switch
> note scoped to the existing adapter seam.
