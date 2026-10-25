# Forever Data and Import Feasibility Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Produce a reproducible, source-backed go/no-go decision for a manual Forever armory and its two optional import routes before building the hosted application.

**Architecture:** This phase produces an evidence packet, not a production application. Investigate catalog/rule coverage first, then independently verify addon and Sixty Upgrades import paths against that evidence. Use minimal permitted samples and explicit blocked outcomes rather than fabricated data or speculative integrations.

**Tech Stack:** Markdown, permitted provider exports, Python 3 standard library for artifact checks, existing Git tooling. No application dependencies, authentication provider, database, scraper framework, or simulator changes in this phase.

**Location:** `docs/plans/` follows the active workspace planning guard. The approved design remains in `docs/superpowers/specs/`.

## Global Constraints

Copied from `docs/superpowers/specs/2026-10-01-forever-armory-design.md`:

- "Blizzard character APIs are not a prerequisite."
- "Manual entry is the foundation. The two import routes are conditional release features: do not advertise them as available until their feasibility gates pass."
- "The picker and tooltips use the Forever catalog, never the bundled TBC database."
- "Do not reuse TBC racial, talent, buff, stat-conversion, or combat formulas as Forever formulas."
- "Do not bypass access controls or treat a JavaScript-rendered page as an API contract."
- "Do not replace the existing TBC upgrade agent or refactor its simulator into a generic multi-game framework as part of this feature."
- "Do not start a single large implementation plan that assumes Gates A and B have passed."

Additional execution constraints:

- Preserve all pre-existing working-tree changes. Never use `git add .`.
- No production frontend/backend changes and no network requests containing edit credentials.
- Cite primary sources for factual conclusions; search results are leads, not evidence.
- A negative finding is a valid deliverable. Record what was checked and what would resolve the block.
- Never commit client files, account cookies, private edit links, or exports without permission to retain them.
- Use the execution date and available game build as observations; do not assume this plan's date establishes beta/release status.

---

## Scope and delivery roadmap

This plan implements **Gate A and Gate B investigation**, not the full approved product.

| Phase | Deliverable | Dependency |
| --- | --- | --- |
| 1 — this plan | Data/rules evidence, import evidence, reproducible samples where permitted, go/no-go report | None |
| 2 — separate plan | Manual character editor, Current/Planned builds, named sets, independent talents, item picker | Sufficient catalog/rule evidence from Phase 1 |
| 3 — separate plan | Durable hosted storage, guest/account ownership, publication, stable multi-set share page | Approved storage/auth design and Phase 2 build contracts |
| 4 — separate importer plans | Addon and Sixty Upgrades integration | Each route independently passes its Phase 1 gate |

Phase 2 is not the release by itself: Phase 3 is necessary for the user's permanent Discord/bookmark links. Whole-character stats may remain explicitly unavailable if their independent coverage gate fails. A failed importer gate does not block manual entry.

Do not invent a production catalog schema, choose an auth/database service, or implement an exporter during this investigation. Record verified requirements so those decisions can be made in the next scoped plan.

## File map

**Create:** `docs/forever-data-import-research.md`

One primary-source research report containing: sources/rights, catalog and rules coverage, addon findings, Sixty Upgrades findings, status matrix, and next decisions. Follow the repo's existing `docs/*-research.md` convention.

**Create:** `docs/superpowers/experiments/2026-10-01-forever-data-import-feasibility.md`

A reproducible experiment log: environment, exact acquisition/inspection steps, fixture hashes, observed outcomes, and gate verdicts. This is not a second competing research report; it records procedures and results that support the report.

**Conditionally create, only when retention is permitted:**

- `docs/superpowers/experiments/forever-fixtures/catalog-sample.json` — small, source-backed evidence projection; not an application catalog.
- `docs/superpowers/experiments/forever-fixtures/addon-export.txt` — consented, credential-free actual Forever addon output.
- `docs/superpowers/experiments/forever-fixtures/sixty-export.txt` — consented, credential-free provider export text obtained through a supported route.

Absent or non-retainable fixtures are recorded as such in the experiment log. Do not create dummy exports to make the artifact list look complete. Screenshots alone cannot substitute for import fixtures.

**Read only:** approved spec, `README.md`, `docs/upgrade-finder.md`, `docs/blizzard-character-api-rights-research.md`, `ui-finder/src/lib/ArmoryView.svelte`, `TalentTrees.svelte`, `GearSlot.svelte`, and `cmd/wowsimcli/cmd/upgrades/import.go`.

## Task 1: Establish primary sources and acquisition rights

**Files:** Create the research report and experiment log.

**Interfaces:**
- Consumes: approved spec, provider links, existing rights research.
- Produces: source table with `source URL`, `owner`, `version/build`, `access method`, `retention/redistribution evidence`, `coverage`, `verdict`, and `checked date`.
- Verdict values throughout this packet: `ready`, `partial`, or `blocked`. `partial` is not permission to ship an unsupported capability.

- [ ] **Step 1: Record the baseline.** Run `git status --short` and `git branch --show-current`; record pre-existing changes and use an isolated execution worktree as required by the execution skill. Read the approved spec and the read-only files above.
- [ ] **Step 2: Create report sections.** Use these headings: `Question`, `Sources and permissions`, `Catalog and rules`, `Addon import`, `Sixty Upgrades import`, `Capability verdicts`, and `Next decisions`. Begin verdicts as `blocked` with the factual reason that acquisition/verification has not occurred yet.
- [ ] **Step 3: Investigate provider-owned material.** Start with `https://www.wowhead.com/forever`, `https://sixtyupgrades.com/forever`, and Blizzard's official Forever pages. Find provider export documentation, repository licenses, data licenses, and applicable terms. Follow first-party references; do not assume permission from public availability or an open-source tool license.
- [ ] **Step 4: Bound the source search.** Inspect the supplied sites, official documentation, and publicly documented/openly licensed alternatives. If rights or access remain unclear, record the exact missing permission/endpoint and a provider-contact route if available. Do not bypass an HTTP 401/403 or inspect a private editor using someone else's credentials.
- [ ] **Step 5: Log reproducible access results.** Record URL, fetch/browser method, checked time, HTTP result where available, and whether the result is actual data, documentation, or an application shell. Do not repeat the earlier unverified simulator-repository search claim as a fact.
- [ ] **Step 6: Review and commit only the two evidence documents.** Every positive rights conclusion must cite the actual license/terms/permission passage and distinguish local processing from hosted retention and redistribution. Unresolved rights remain `blocked`.

```bash
git diff --check
git add -- docs/forever-data-import-research.md docs/superpowers/experiments/2026-10-01-forever-data-import-feasibility.md
git diff --cached --name-only
git commit -m "docs: establish Forever data sources and access evidence"
```

**Acceptance:** The report identifies what is genuinely accessible and what may be retained/served. An evidence-backed blocked result passes the task; it does not pass Gate A.

## Task 2: Verify catalog, slot, and talent-rule coverage

**Files:** Update both evidence documents; conditionally create `catalog-sample.json`.

**Interfaces:**
- Consumes: Task 1's permitted sources and version identities.
- Produces: per-capability coverage and evidence references for items, enchants, slot legality, races/classes, talent structure/rules, source metadata, and whole-character stats.
- A retained sample has top-level fields `sourceUrl`, `version`, `retrievedAt`, and `records`. Each record has `kind`, `id`, `evidenceUrl`, and `observed`; `observed` holds the actual verified fields. This is an evidence wrapper, not a production import format.

- [ ] **Step 1: Choose the smallest lawful candidate dataset.** If Task 1 found none, record Gate A as blocked and proceed to documenting the import routes; do not manufacture a catalog. A pre-beta/incomplete source must carry its actual confidence/build labels.
- [ ] **Step 2: Trace representative equipment.** Inspect one armor item, one weapon/two-hand conflict, one ring or trinket, and one enchant where available. Record stable identifiers, name, quality, slot, level/class restrictions, stats, source, and missing fields. Verify which equipment slots Forever actually supports, including how cosmetic slots differ from stat equipment.
- [ ] **Step 3: Trace representative talents.** Inspect one full class tree and at least one prerequisite, rank ceiling, row-unlock condition, and level-derived point budget. Cross-check the same rules at level 60 and one supported lower level. Record whether encoding is documented or was mapped by direct observations.
- [ ] **Step 4: Write legal/illegal examples.** For each verified rule record `input`, `expected allowed/rejected`, `rule citation`, and `observed outcome`. Include too-high item level requirement, incompatible equipment slot, two-handed/off-hand conflict, talent over-rank, unmet prerequisite, and exceeded point budget. These are fixture-backed acceptance cases for the future validator, not claims that production validation already exists.
- [ ] **Step 5: Audit breadth separately from examples.** Record supported classes/races, level range, and approximate catalog completeness. One observed class or item cannot establish all-class coverage. Mark unsupported breadth explicitly; an all-class release needs sufficient data for all selectable classes.
- [ ] **Step 6: Assess stats independently.** Identify which base/racial/talent/level formulas are actually available. If only item stat lines are known, mark full character totals blocked; do not add TBC formulas or call a sum of gear stats a character total.
- [ ] **Step 7: If a sample can be retained, create and check its wrapper.** Populate only real observations and source references. Use this standard-library check; a missing required value fails rather than being guessed:

```bash
python - <<'PY'
import json
from pathlib import Path
p = Path('docs/superpowers/experiments/forever-fixtures/catalog-sample.json')
s = json.loads(p.read_text(encoding='utf-8'))
assert all(s.get(k) for k in ('sourceUrl', 'version', 'retrievedAt', 'records'))
assert isinstance(s['records'], list)
for record in s['records']:
    assert record['kind'] in ('item', 'enchant', 'talent', 'rule')
    assert record['id'] is not None
    assert record['evidenceUrl'].startswith('https://')
    assert isinstance(record['observed'], dict) and record['observed']
print(f"Evidence wrapper valid: {len(s['records'])} records")
PY
```

This check validates artifact structure, not game correctness or rights. Record the inspection evidence and coverage verdict regardless of whether a sample is retainable.

- [ ] **Step 8: Commit the verified documents and only the permitted sample, if created.** Run `git diff --check`, stage the two exact document paths, and stage the exact sample path only if present and reviewed. Inspect `git diff --cached --name-only` before committing with `docs: verify Forever catalog and talent coverage`.

**Acceptance:** Gate A has an evidence-backed verdict and explicit coverage limits. There is no ready verdict for manual gearing/talents if data rights, acquisition, identifiers, or core validation rules are still unverified. Source labels and full-character stats may be separately unavailable without being invented.

## Task 3: Verify addon and Sixty Upgrades imports independently

**Files:** Update both documents; conditionally create `addon-export.txt` and `sixty-export.txt`.

**Interfaces:**
- Consumes: verified source/version/reference coverage from Tasks 1–2.
- Produces: one mapping table per route: `target field`, `actual source path/encoding`, `representative value`, `version`, `missing-field behavior`, and `evidence`.
- Target fields: character name/race/class/level, optional region/realm, slot/item identifier, enchant identifier, talent allocation, and export/source version. Authentication and ownership are never imported.

- [ ] **Step 1: Inspect existing exporters.** Find primary documentation/source code for candidate addons, including the Sixty Upgrades exporter. Verify actual Forever client support, exported fields, format version, and license. A TBC export or a search snippet does not demonstrate Forever support.
- [ ] **Step 2: Obtain one real addon export lawfully.** Use an available supported client with the owner's consent, or request an export from the user if no client is available. Never install/provision client data without authorization. If a fixture cannot be obtained, mark this route blocked with the exact required input and continue; do not claim the format is verified from documentation alone.
- [ ] **Step 3: Inspect the public Sixty Upgrades set.** Start from `https://sixtyupgrades.com/forever/set/vH9HQmj6tzxi7fajuHmu8p`. Check public UI actions and provider documentation. Prefer user-produced Export output over undocumented endpoint dependencies. Record whether a share link is readable, whether a supported machine export exists, and which fields it preserves.
- [ ] **Step 4: Separate link import from pasted-file/text import.** If provider-generated text works but URL retrieval is undocumented, blocked, or permission-unclear, mark text/file import partial/ready as evidence warrants and URL import blocked. Do not advertise a working paste-a-link route based on another route's success.
- [ ] **Step 5: Map actual exports.** For each obtained fixture, follow the documented decoder or provider-owned implementation and record exact decoding steps in the experiment log. Match its gear/enchant/talent references to Task 2's Forever evidence. Missing fields remain missing; export metadata must not overwrite application ownership or publication settings.
- [ ] **Step 6: Record rejection and loss cases.** Document a truncated payload, unknown item identifier, wrong game/version, and absent talents. Record observed decoding/provider behavior separately from the required behavior of our future importer. The future importer rejects malformed, unknown-reference, and wrong-version input; valid partial exports preserve explicitly missing fields. Do not claim that our not-yet-written importer passed these cases. No Current/plan replacement happens in this investigation.
- [ ] **Step 7: Sanitize before retention.** Remove credentials, account identifiers not needed for mapping, and private URLs. If sanitation changes bytes, record that fact and preserve enough information to reproduce the format checks. If retaining the fixture is not permitted, record a lawful reproduction procedure without committing it.
- [ ] **Step 8: Verify byte identity for retained fixtures.** Run this check and paste the printed hashes into the experiment log; absence is a reported condition, not an implied success:

```bash
python - <<'PY'
import hashlib
from pathlib import Path
root = Path('docs/superpowers/experiments/forever-fixtures')
for name in ('catalog-sample.json', 'addon-export.txt', 'sixty-export.txt'):
    p = root / name
    if not p.exists():
        print(f'{name}: not retained; see experiment log')
        continue
    data = p.read_bytes()
    assert data, f'{name}: empty fixture'
    assert len(data) <= 1048576, f'{name}: exceeds investigation size limit'
    print(f'{name}: {len(data)} bytes; sha256={hashlib.sha256(data).hexdigest()}')
PY
```

The 1 MiB limit is an investigation safety limit, not a decided production request contract. Hashes establish sample identity, not import correctness.

- [ ] **Step 9: Review and commit only evidence and approved fixtures.** Run `git diff --check`, explicitly stage the two documents and any reviewed retained export paths, inspect the staged file list, and commit with `docs: verify Forever character import routes`.

**Acceptance:** Both routes have independent verdicts with actual mapping evidence or explicit missing inputs. Writing a new exporter is a proposed next task, never silently included here.

## Task 4: Produce the go/no-go handoff

**Files:** Finalize both evidence documents. No application changes.

**Interfaces:**
- Consumes: the actual evidence from Tasks 1–3.
- Produces: a decision matrix and a narrowly scoped recommendation for the next implementation plan.

- [ ] **Step 1: Fill the capability matrix.** Use rows `manual item selection`, `enchants`, `independent talent planning`, `lower-level restrictions`, `source filters`, `whole-character stats`, `addon import`, `Sixty text/file import`, and `Sixty link import`. Every row contains `status`, `evidence`, `coverage limits`, and `next action`.
- [ ] **Step 2: Apply the stop rules.** If no permitted versioned catalog or no sufficient talent rules exist, block the corresponding real-data editor work. Recommend the specific provider permission, user fixture, or authoritative data needed. Do not relabel an unsupported production feature as a demo to pass the gate.
- [ ] **Step 3: Specify the next plan's boundary.** If manual item/talent evidence is sufficient, recommend the manual Current/Planned editor as the next scoped plan. List exact verified fields, the selectable coverage, unavailable stats, and import routes to omit. Keep hosted ownership/sharing as a required subsequent milestone, not an optional replacement with local-only storage.
- [ ] **Step 4: Check the packet.** Confirm every positive capability has a citation and observed result, every retained fixture has rights/consent and a hash, and every unsupported route has a resolution action. Check that no secret or private editor URL is present. Re-run the applicable sample checks above and `git diff --check`.
- [ ] **Step 5: Commit the final verdict.** Explicitly stage the two documents and any corrected permitted samples, review `git diff --cached --name-only`, and commit with `docs: record Forever armory feasibility verdict`.
- [ ] **Step 6: Ask the user to approve the next scope.** Summarize what is ready, what is blocked, and what inputs/permissions are needed. Do not start the manual editor or hosted service merely because this investigation finished.

**Acceptance:** A second engineer can reproduce the evidence, understand the limits, and decide whether manual editor work is justified without assuming API availability or TBC compatibility.

## Spec coverage and deliberate exclusions

- Catalog/slot/race/class/talent/lower-level requirements: Tasks 1–2.
- Addon/Sixty imports, provenance, missing references, and permissions: Tasks 1 and 3.
- Gate verdicts, preservation of the TBC app, and next scoped decisions: Task 4.
- Current/Planned editing and independent plans: Phase 2, contingent on the evidence.
- Guest/account access, recovery/revocation, backups, protected publication, and stable updated links: Phase 3; not implemented or claimed by this plan.
- Actual importer code: Phase 4, independently gated.
- Simulation, 3D rendering, search directories, and verified in-game ownership remain deferred as in the spec.

No frontend/Go test run proves a research conclusion. Since this phase edits only evidence documents and permitted samples, verification is source/rights review, reproduced observations, artifact checks, and clean scoped diffs. Production tests belong in the subsequent implementation plans.
