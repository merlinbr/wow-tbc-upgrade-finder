# Project README Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the root README introduce and explain the local TBC Upgrade Finder while retaining clear attribution to its bundled WoW TBC simulator upstream.

**Architecture:** Replace the upstream-only README copy with a concise, user-first landing page. It describes the local command and workflow at summary level, then links to `docs/upgrade-finder.md` as the canonical operator guide. Upstream provenance stays visible but becomes supporting context rather than the page's primary identity.

**Tech Stack:** Markdown; Go CLI command documentation; wowsims/tbc-new upstream links.

## Global Constraints

- Modify only `README.md`; do not change application behavior, build tooling, simulator code, or linked documentation.
- Keep `docs/upgrade-finder.md` authoritative for detailed validation, armory, result, and smoke-test instructions.
- State that the command binds locally and keeps character data/reports only in process memory.
- Retain a visible `wowsims/tbc-new` attribution and link `UPSTREAM.md` for pinned simulator/database revisions.
- Add no dependencies or README-specific tooling.

---

### Task 1: Replace the upstream-only landing page

**Files:**
- Modify: `README.md:1-30`
- Reference: `docs/upgrade-finder.md:1-112`
- Reference: `UPSTREAM.md:1-7`

**Interfaces:**
- Consumes: `wowsimcli rank-upgrades` as the user-facing local command.
- Produces: a user-first repository landing page with working relative documentation links and upstream attribution.

- [x] **Step 1: Replace the title and introduction**

Replace `# WoW The Burning Crusade Classic Simulator` and its upstream-only introduction with a `# TBC Upgrade Finder` heading and this scope statement:

```markdown
A local browser application that ranks practical single-item DPS upgrades for one The Burning Crusade character. It imports an individual-sim configuration and evaluates candidates with the bundled [wowsims/tbc-new](https://github.com/wowsims/tbc-new) simulator and item data.
```

- [x] **Step 2: Add a quick-start and usage summary**

Add a `## Quick start` section containing:

```markdown
```bash
rtk go run ./cmd/wowsimcli rank-upgrades
```

The command starts a loopback-only local server and opens the application in your browser. Add `--no-browser` when you only need the URL printed to the terminal.
```

Follow it with a `## Use it` ordered list: paste an individual-sim wowsims export link, review the 17-slot armory, choose filters and iteration budgets, then start ranking and inspect the report.

- [x] **Step 3: Add boundaries and documentation links**

Add a concise `## What it does` section stating that rankings cover practical single-item DPS upgrades, imported baselines are unchanged, and character data/jobs/reports are in-process only with no remote persistence.

Add `## Documentation` with these links:

```markdown
- [Upgrade Finder guide](docs/upgrade-finder.md)
- [Installation Guide](docs/installation.md)
- [Development Commands](docs/commands.md)
- [Adding a New Sim](docs/adding_sim.md)
- [Internationalization](docs/i18n_guide.md)
```

- [x] **Step 4: Retain upstream attribution and provenance**

Add `## Upstream simulator` with a concise description that the project bundles the TBC simulator and data from `wowsims/tbc-new`; point readers to [`UPSTREAM.md`](UPSTREAM.md) for pinned revisions. Keep the existing live-sims and Patreon links. Keep the existing MIT license and visible-attribution request, phrased as upstream attribution rather than the Upgrade Finder's sole project description.

- [x] **Step 5: Verify rendered links and factual claims**

Confirm every local Markdown destination exists: `docs/upgrade-finder.md`, `docs/installation.md`, `docs/commands.md`, `docs/adding_sim.md`, `docs/i18n_guide.md`, and `UPSTREAM.md`. Compare the quick-start command, loopback-only behavior, input type, workflow, persistence statement, and single-item DPS scope against `docs/upgrade-finder.md`.

- [x] **Step 6: Commit the README update**

```bash
rtk git add README.md
rtk git commit -m "docs: describe TBC Upgrade Finder"
```
