# Upgrade Finder CI design

## Goal

Run the existing Upgrade Finder verification contract automatically before pull requests merge into `master` and after every push to `master`.

## Scope

Create one dedicated GitHub Actions workflow at `.github/workflows/upgrade_finder.yml`. It is independent of the upstream simulator's reusable `run_tests.yml`, so release/deploy workflows do not gain Upgrade Finder browser-test work.

## Triggers

Run on:

```yaml
pull_request:
  branches: [master]
push:
  branches: [master]
```

## Job

One Ubuntu job, with a 30-minute timeout, runs the checks in the same order as the documented local verification:

1. Checkout the triggering revision with `actions/checkout@v5`.
2. Install Go `1.25.x` with `actions/setup-go@v6`.
3. Run `go test ./cmd/wowsimcli/cmd/upgrades -count=1`.
4. Run `go test ./cmd/wowsimcli/cmd/... -count=1`.
5. Run `go build -o wowsimcli ./cmd/wowsimcli`.
6. Install Node `22` with `actions/setup-node@v5`.
7. In `ui-finder`, run `npm ci` then `npm run build`.
8. Install Chromium and Linux dependencies with `npx playwright install --with-deps chromium`.
9. In `ui-finder`, run `npm run test:e2e`.

The Playwright test config starts the local CLI itself at a fixed loopback address, so the workflow needs no external service, deployed URL, credentials, artifact upload, or data persistence.

## Constraints

- Reuse the repository's existing GitHub Action major versions and Go/Node version policy.
- No cache, matrix, parallel jobs, browser-report artifact, release, or deployment changes. Add those only if CI runtime or debugging data makes them necessary.
- Do not change application code, tests, package manifests, simulator workflows, or the existing Playwright configuration.
- Keep the workflow name and job labels specific to the TBC Upgrade Finder so it is distinct in the GitHub Actions UI.

## Verification

Validate YAML syntax and compare every command against `docs/STATE.md`, `ui-finder/package.json`, and `ui-finder/playwright.config.js`. GitHub Actions performs the full behavioral verification on the first pull request or `master` push after the workflow is published.
