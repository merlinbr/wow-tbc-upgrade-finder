# Optional Simulation Settings Import Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Import and rank valid current individual-sim exports that omit the optional simulation-settings protobuf message.

**Architecture:** Keep `IndividualSimSettings.settings` optional in `upgrades.Import`; protobuf zero values already represent absent phase, iterations, and random seed. At the HTTP boundary, replace an absent/non-positive phase with the UI's established default phase `5`, so the rank request stays bounded.

**Tech Stack:** Go, protobuf, Cobra HTTP server, Go standard-library testing.

## Global Constraints

- Preserve all existing typed validation for malformed links, raid links, player, encounter, and equipment.
- Do not add dependencies, a diagnostic endpoint, or a protocol migration.
- Retain positive exported simulation-settings values unchanged.
- Use the exact supplied API-v14 Retribution Paladin link as the regression fixture.

---

### Task 1: Accept optional simulation settings

**Files:**
- Create: `cmd/wowsimcli/cmd/upgrades/testdata/retribution_no_settings_link.txt`
- Modify: `cmd/wowsimcli/cmd/upgrades/import.go:140-145,249-259`
- Modify: `cmd/wowsimcli/cmd/upgrades/import_test.go`

**Interfaces:**
- Consumes: `proto.IndividualSimSettings.GetSettings() *proto.SimSettings`; protobuf getters return zero values for a nil message.
- Produces: `Import(link string) (*ImportedSettings, error)` succeeds when `settings` is absent; `CharacterSummary` reports zero-valued phase/iterations and `false` fixed-seed state.

- [x] **Step 1: Add the failing production-export fixture and focused import test**

Create `retribution_no_settings_link.txt` containing exactly the supplied URL. Add this test next to `TestImportDecodesFixedIndividualLink`:

```go
func TestImportAcceptsExportWithoutSimSettings(t *testing.T) {
    settings := readFixture(t, "retribution_no_settings_link.txt")

    imported, err := Import(settings)
    if err != nil {
        t.Fatalf("Import() error = %v", err)
    }
    if got, want := imported.Character.Class, proto.Class_ClassPaladin.String(); got != want {
        t.Fatalf("class = %q, want %q", got, want)
    }
    if got, want := imported.Character.Spec, "RetributionPaladin"; got != want {
        t.Fatalf("spec = %q, want %q", got, want)
    }
    if imported.Character.Phase != 0 || imported.Character.Iterations != 0 || imported.Character.FixedRngSeed {
        t.Fatalf("summary = %+v, want zero-valued missing simulation settings", imported.Character)
    }
}
```

- [x] **Step 2: Run the focused test to verify the current rejection**

Run: `rtk go test ./cmd/wowsimcli/cmd/upgrades -run '^TestImportAcceptsExportWithoutSimSettings$' -count=1`

Expected: FAIL with `missing simulation settings in export`.

- [x] **Step 3: Remove the invalid required-message check and use nil-safe protobuf getters**

Delete the `if settings.Settings == nil` validation block in `Import`. Replace the direct summary field reads with the generated protobuf getters:

```go
Phase:         settings.GetSettings().GetPhase(),
Iterations:    settings.GetSettings().GetIterations(),
FixedRngSeed:  settings.GetSettings().GetFixedRngSeed() != 0,
```

This leaves player, encounter, equipment, version, and database-reference validation untouched.

- [x] **Step 4: Re-run the focused import test**

Run: `rtk go test ./cmd/wowsimcli/cmd/upgrades -run '^TestImportAcceptsExportWithoutSimSettings$' -count=1`

Expected: PASS.

- [x] **Step 5: Commit the import behavior**

```bash
rtk git add cmd/wowsimcli/cmd/upgrades/import.go cmd/wowsimcli/cmd/upgrades/import_test.go cmd/wowsimcli/cmd/upgrades/testdata/retribution_no_settings_link.txt
rtk git commit -m "fix: accept exports without sim settings"
```

### Task 2: Default a missing exported phase at the HTTP boundary

**Files:**
- Modify: `cmd/wowsimcli/cmd/upgrade_server.go:220-252`
- Modify: `cmd/wowsimcli/cmd/upgrade_server_test.go`

**Interfaces:**
- Consumes: `imported.Character.Phase`, which is `0` when the optional exported `SimSettings` message is absent.
- Produces: `POST /api/import` returns `defaults.maxPhase: 5` for that export and preserves a positive imported phase.

- [x] **Step 1: Add the failing import-endpoint regression test**

Add a helper that reads `upgrades/testdata/retribution_no_settings_link.txt`, then add:

```go
func TestImportDefaultsPhaseWhenExportOmitsSimSettings(t *testing.T) {
    server := newTestServer(t, &fakeRanker{})
    response := postJSON(t, server.URL+"/api/import", map[string]string{
        "link": optionalSettingsFixtureLink(t),
    })
    if response.StatusCode != http.StatusOK {
        t.Fatalf("status = %d, want 200", response.StatusCode)
    }
    defaults, ok := decodeBody(t, response)["defaults"].(map[string]any)
    if !ok || defaults["maxPhase"] != float64(5) {
        t.Fatalf("defaults = %#v, want maxPhase 5", defaults)
    }
}
```

- [x] **Step 2: Run the endpoint test to verify the current failure**

Run: `rtk go test ./cmd/wowsimcli/cmd -run '^TestImportDefaultsPhaseWhenExportOmitsSimSettings$' -count=1`

Expected: FAIL because import returns HTTP 400.

- [x] **Step 3: Apply the bounded default before constructing `importResponse`**

In `handleImport`, after armory enrichment and before `writeJSON`, add:

```go
maxPhase := imported.Character.Phase
if maxPhase < 1 {
    maxPhase = 5
}
```

Set `Defaults.MaxPhase` to `maxPhase` instead of `imported.Character.Phase`.

- [x] **Step 4: Re-run the endpoint regression test**

Run: `rtk go test ./cmd/wowsimcli/cmd -run '^TestImportDefaultsPhaseWhenExportOmitsSimSettings$' -count=1`

Expected: PASS with HTTP 200 and `defaults.maxPhase` equal to `5`.

- [x] **Step 5: Commit the HTTP default**

```bash
rtk git add cmd/wowsimcli/cmd/upgrade_server.go cmd/wowsimcli/cmd/upgrade_server_test.go
rtk git commit -m "fix: default phase for current exports"
```

### Task 3: Verify the regression through the real application

**Files:**
- Modify: none.

**Interfaces:**
- Consumes: the production fixture and the updated import endpoint.
- Produces: the running upgrade finder accepts the supplied URL and renders the imported Ret Paladin armory.

- [x] **Step 1: Run the focused server and upgrades packages**

Run: `rtk go test ./cmd/wowsimcli/cmd/upgrades ./cmd/wowsimcli/cmd -count=1`

Expected: PASS.

- [x] **Step 2: Build the executable**

Run: `rtk go build -o wowsimcli ./cmd/wowsimcli`

Expected: exit status `0`.

- [x] **Step 3: Restart the local service and import the fixture through its browser UI**

Run the rebuilt `wowsimcli rank-upgrades`, paste the fixture URL, click **Import settings**, and verify the Ret Paladin header, 17 armory slots, and maximum phase `5` appear without an error.

- [x] **Step 4: Commit the implementation plan if it changed during execution**

```bash
rtk git add docs/superpowers/plans/2026-08-31-optional-sim-settings-import.md
rtk git commit -m "docs: plan optional sim settings import"
```
