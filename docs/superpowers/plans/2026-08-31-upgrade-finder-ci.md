# Upgrade Finder CI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Automatically run the existing Upgrade Finder verification contract on pull requests and pushes targeting `master`.

**Architecture:** Add one standalone GitHub Actions workflow. A single Ubuntu job installs the pinned Go and Node runtimes, runs the focused backend checks and build, then builds the isolated Svelte UI and executes its existing local Playwright smoke test. The workflow has no dependencies on the upstream simulator's reusable workflows.

**Tech Stack:** GitHub Actions; Ubuntu; Go 1.25; Node 22; npm; Playwright Chromium.

## Global Constraints

- Create only `.github/workflows/upgrade_finder.yml`; do not modify application code, tests, package manifests, simulator workflows, or Playwright configuration.
- Trigger on `pull_request` and `push`, both restricted to `master`.
- Reuse `actions/checkout@v5`, `actions/setup-go@v6`, and `actions/setup-node@v5`, matching `.github/workflows/run_tests.yml`.
- Use one Ubuntu job with `timeout-minutes: 30`; no cache, matrix, artifact upload, release, or deployment integration.
- Run the exact focused commands documented in `docs/STATE.md` and `ui-finder/package.json`.

---

### Task 1: Add the Upgrade Finder verification workflow

**Files:**
- Create: `.github/workflows/upgrade_finder.yml`
- Reference: `.github/workflows/run_tests.yml:1-104`
- Reference: `docs/STATE.md:45-65`
- Reference: `ui-finder/package.json:1-17`
- Reference: `ui-finder/playwright.config.js:1-22`

**Interfaces:**
- Consumes: GitHub `pull_request` and `push` events for `master`.
- Produces: one `Upgrade Finder CI` GitHub Actions check that passes only after backend, CLI, UI-build, and Chromium browser-smoke verification all pass.

- [x] **Step 1: Create the workflow trigger and job shell**

Create `.github/workflows/upgrade_finder.yml` with the workflow name, both `master`-restricted triggers, and the single 30-minute Ubuntu job:

```yaml
name: Upgrade Finder CI

on:
  pull_request:
    branches: [master]
  push:
    branches: [master]

jobs:
  verify:
    name: Verify Upgrade Finder
    runs-on: ubuntu-latest
    timeout-minutes: 30
```

- [x] **Step 2: Configure source checkout and runtimes**

Add these steps beneath `jobs.verify`:

```yaml
    steps:
      - name: Checkout
        uses: actions/checkout@v5

      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version: 1.25.x

      - name: Set up Node
        uses: actions/setup-node@v5
        with:
          node-version: 22
```

- [x] **Step 3: Add the focused backend verification**

Add separate named workflow steps for each documented backend command, preserving `-count=1` so cached Go test results do not conceal failures:

```yaml
      - name: Test upgrade ranking
        run: go test ./cmd/wowsimcli/cmd/upgrades -count=1

      - name: Test CLI packages
        run: go test ./cmd/wowsimcli/cmd/... -count=1

      - name: Build CLI
        run: go build -o wowsimcli ./cmd/wowsimcli
```

- [x] **Step 4: Build and browser-test the UI**

Add UI-scoped commands with GitHub Actions `working-directory` rather than shell `cd`, then install the Chromium executable and Linux dependencies before running the existing browser test:

```yaml
      - name: Install UI dependencies
        working-directory: ui-finder
        run: npm ci

      - name: Build UI
        working-directory: ui-finder
        run: npm run build

      - name: Install Chromium
        working-directory: ui-finder
        run: npx playwright install --with-deps chromium

      - name: Run browser smoke test
        working-directory: ui-finder
        run: npm run test:e2e
```

- [x] **Step 5: Verify workflow behavior locally**

Run the workflow's commands in order from the repository root:

```bash
rtk go test ./cmd/wowsimcli/cmd/upgrades -count=1
rtk go test ./cmd/wowsimcli/cmd/... -count=1
rtk go build -o wowsimcli ./cmd/wowsimcli
cd ui-finder && rtk npm ci && rtk npm run build && rtk npm run test:e2e
```

Confirm the workflow does not need a remote URL: `ui-finder/playwright.config.js` starts `rtk go run ./cmd/wowsimcli rank-upgrades --addr 127.0.0.1:43123 --no-browser` as its `webServer`.

- [x] **Step 6: Commit the workflow**

```bash
rtk git add .github/workflows/upgrade_finder.yml
rtk git commit -m "ci: verify Upgrade Finder"
```
