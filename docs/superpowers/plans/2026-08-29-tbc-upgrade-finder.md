# TBC Upgrade Finder Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship a local browser application that imports one wowsims TBC individual-sim link and returns confirmed, obtainable, single-item DPS upgrades under that exact simulation configuration.

**Architecture:** Make this repository a pinned fork of `wowsims/tbc-new`, then add a `rank-upgrades` subcommand to its existing `wowsimcli` binary. A Go upgrade service decodes and canonicalizes the link, constructs legal equipment replacements from the bundled `UIDatabase`, runs the existing simulator at screening and confirmation budgets, and returns an immutable `UpgradeReport`. A small embedded HTML/CSS/vanilla-JS client talks only to local JSON endpoints; jobs and reports live only in process memory.

**Tech Stack:** Go 1.25.0/toolchain 1.25.4, upstream wowsims TBC Go simulator and protobuf database, Cobra, `net/http`, `embed`, browser-native JavaScript/CSS.

## Global Constraints

- Start from upstream revision `88fb853466a391e731e12de012f6707a11e94446`; record that revision and the bundled database blob revision `84555ed6e3ddf19edca22204892a2366e1a177da` in every report.
- Accept only wowsims **individual-sim** export links. Reject malformed, raid-sim, incompatible API-version, and unknown-item links before any simulation starts.
- Use the wowsims engine and the bundled database as the sole scoring and item-data authority. Do not add static stat weights, LLM scoring, remote APIs, accounts, persistence, or a database server.
- Rank one candidate equipment replacement at a time by DPS only. Never return a partial ranking as a final/confirmed report.
- Source metadata is required for recommendations. Exclude and count candidates without source metadata by default.
- Preserve the imported baseline byte-for-byte. Every candidate starts from a deep copy of that baseline.
- Apply only the user-declared gem/enchant policy and only where it is legal; do not search for globally optimal gem or enchant combinations.
- A failed candidate is isolated and reported with its reason. Cancellation stops outstanding work and leaves the job without a report.
- The browser page must visibly link to `https://github.com/wowsims/tbc-new` and state that wowsims supplies the simulator and item database, as requested by its MIT-licensed README.
- Use no new runtime dependencies. The existing Cobra/protobuf dependencies and Go standard library are sufficient.

---

## Locked File Structure

The current repository contains only planning material; the implementation begins by importing the upstream source tree at the pinned revision. The product code then has these focused additions:

| Path | Responsibility |
| --- | --- |
| `cmd/wowsimcli/cmd/rank_upgrades.go` | Cobra command, server construction, and browser launch only. |
| `cmd/wowsimcli/cmd/upgrade_server.go` | Local HTTP routes, job lifecycle, JSON encoding, and no domain logic. |
| `cmd/wowsimcli/cmd/upgrade_ui/index.html` | Accessible form and report layout. |
| `cmd/wowsimcli/cmd/upgrade_ui/app.js` | Browser-side import, polling, cancellation, rendering, and copy actions. |
| `cmd/wowsimcli/cmd/upgrade_ui/app.css` | Small responsive styles for the embedded page. |
| `cmd/wowsimcli/cmd/upgrades/types.go` | Stable request, report, filter, policy, revision, and error types. |
| `cmd/wowsimcli/cmd/upgrades/import.go` | Pure exported-link decoding, API compatibility checking, and canonical individual-request conversion. |
| `cmd/wowsimcli/cmd/upgrades/catalog.go` | Bundled `UIDatabase` indexing and source-name resolution. |
| `cmd/wowsimcli/cmd/upgrades/candidates.go` | Filtering, compatibility checks, legal slot variants, and immutable equipment mutations. |
| `cmd/wowsimcli/cmd/upgrades/policy.go` | Gem/enchant policy validation and legal candidate configuration. |
| `cmd/wowsimcli/cmd/upgrades/service.go` | `RankUpgrades` orchestration, screening/confirmation selection, ranking, and cancellation. |
| `cmd/wowsimcli/cmd/upgrades/simulator.go` | Narrow adapter over `core.RunRaidSimConcurrentAsync`, progress, and player-DPS extraction. |
| `cmd/wowsimcli/cmd/upgrades/*_test.go` | Deterministic contract tests using compact in-memory catalog/settings fixtures and a fake simulator. |
| `cmd/wowsimcli/cmd/upgrade_server_test.go` | HTTP/job contract tests with the real import service and fake rank service. |
| `docs/upgrade-finder.md` | Local run instructions, input boundaries, non-goals, attribution, and smoke-test procedure. |

### Task 1: Establish the pinned upstream application base

**Files:**
- Create/replace: upstream source tree from `https://github.com/wowsims/tbc-new` at commit `88fb853466a391e731e12de012f6707a11e94446`
- Preserve: `docs/superpowers/specs/2026-08-29-tbc-upgrade-finder-design.md`
- Create: `UPSTREAM.md`
- Modify: `cmd/wowsimcli/cmd/root.go`
- Modify: `cmd/wowsimcli/cli_main.go`
- Test: `cmd/wowsimcli/cmd/rank_upgrades_command_test.go`

**Interfaces:**
- Consumes: upstream `cmd/wowsimcli` root command, `sim.RegisterAll()`, `assets/database.Load()`, and `core.RunRaidSimConcurrentAsync`.
- Produces: `wowsimcli rank-upgrades [--addr 127.0.0.1:0] [--no-browser]`, a single local executable mode that owns the HTTP server.

- [x] **Step 1: Import the upstream tree without creating a second simulator module**

Make the repository a fork whose source parent is exactly commit `88fb853466a391e731e12de012f6707a11e94446`; keep that upstream commit reachable as `upstream/master`. Retain upstream `go.mod`, `LICENSE`, generated protobufs, `assets/database/db.bin`, and the existing `cmd/wowsimcli` binary. Do not use a Go `replace`, a copied subset of simulator packages, or a separate simulation process: those approaches either lose the generated database or duplicate simulator ownership.

- [x] **Step 2: Record provenance next to the source pin**

Create `UPSTREAM.md` containing the source repository, immutable simulator revision, database blob revision, and update rule:

```markdown
# Upstream provenance

- Simulator source: https://github.com/wowsims/tbc-new
- Simulator revision: `88fb853466a391e731e12de012f6707a11e94446`
- Bundled database blob: `assets/database/db.bin` at `84555ed6e3ddf19edca22204892a2366e1a177da`

Upgrade the simulator and database together. Update this file, the report revision constants, fixtures, and every expected report when changing this pin.
```

- [x] **Step 3: Write the failing command-registration test**

Add a test that builds the root command through the existing `cmd.Execute` construction seam (extract `newRootCommand(version string) *cobra.Command` from `root.go` if necessary) and asserts the new command is discoverable:

```go
func TestRankUpgradesCommandIsRegistered(t *testing.T) {
    command, _, err := newRootCommand("test").Find([]string{"rank-upgrades"})
    if err != nil {
        t.Fatal(err)
    }
    if command.Name() != "rank-upgrades" {
        t.Fatalf("command = %q, want rank-upgrades", command.Name())
    }
}
```

- [x] **Step 4: Add the command shell and keep initialization singular**

Register `rankUpgradesCmd` from `root.go`; define `--addr` (default `127.0.0.1:0`) and `--no-browser` on that command. Let `cli_main.go` continue to call `sim.RegisterAll()` once before `cmd.Execute(Version)`. The command calls `newUpgradeServer(Version).Serve(ctx, addr, openBrowser)` and must print the resolved `http://127.0.0.1:<port>/` URL. It must neither start a remote listener nor initialize a second simulator/database.

- [x] **Step 5: Run the focused command test**

Run: `rtk go test ./cmd/wowsimcli/cmd -run TestRankUpgradesCommandIsRegistered -count=1`

Expected: PASS.

- [x] **Step 6: Commit the upstream base and command entry point**

```bash
git add .
git commit -m "feat: add pinned upgrade finder command"
```

### Task 2: Decode links into an immutable canonical baseline

**Files:**
- Create: `cmd/wowsimcli/cmd/upgrades/types.go`
- Create: `cmd/wowsimcli/cmd/upgrades/import.go`
- Create: `cmd/wowsimcli/cmd/upgrades/import_test.go`
- Modify: `cmd/wowsimcli/cmd/decode_link.go`

**Interfaces:**
- Consumes: a wowsims individual export URL and `proto.IndividualSimSettings` fields.
- Produces: `func Import(link string) (*ImportedSettings, error)` and `func (s *ImportedSettings) NewRequest(iterations int32) *proto.RaidSimRequest`.
- Error contract: typed `ValidationError{Code, Message}` where `Code` is one of `invalid_link`, `unsupported_link`, `incompatible_settings`, or `incompatible_item`; callers render `Message` and never run a job after this error.

- [x] **Step 1: Define the boundary types and deterministic copy helper**

In `types.go`, import `googleProto "google.golang.org/protobuf/proto"` alongside the generated `sim/core/proto` package, keep the UI-facing request separate from protobufs, and make every report field JSON-ready:

```go
type ImportedSettings struct {
    Link             string
    Settings         *proto.IndividualSimSettings
    SettingsDigest   string
    Character        CharacterSummary
    SimulatorVersion string
    DatabaseVersion  string
}

type ValidationError struct { Code, Message string }
func (e *ValidationError) Error() string { return e.Message }

func cloneMessage[M googleProto.Message](m M) M { return googleProto.Clone(m).(M) }
```

Use `proto.MarshalOptions{Deterministic: true}` plus SHA-256 for `SettingsDigest`; do not retain or mutate the caller's message.

- [x] **Step 2: Write decoder contract tests before implementation**

Create a compact `IndividualSimSettings` fixture, encode it using the same zlib + standard-base64 fragment format used by wowsims, and commit the resulting fixed URL in `testdata/fixed_individual_link.txt`. Assert all of the following:

```go
func TestImportDecodesFixedIndividualLink(t *testing.T) {
    imported, err := Import(readFixture(t, "fixed_individual_link.txt"))
    if err != nil { t.Fatal(err) }
    if got, want := imported.Character.Class, proto.Class_ClassMage.String(); got != want {
        t.Fatalf("class = %q, want %q", got, want)
    }
}

func TestImportRejectsRaidAndMalformedLinks(t *testing.T) {
    for _, link := range []string{"https://wowsims.com/tbc/mage/", "https://wowsims.com/tbc/mage/#!", "https://wowsims.com/tbc/raid/#" + fixedFragment} {
        if _, err := Import(link); err == nil {
            t.Fatalf("Import(%q) succeeded", link)
        }
    }
}

func TestNewRequestDoesNotMutateImportedSettings(t *testing.T) {
    imported := mustImportFixture(t)
    before := mustMarshal(t, imported.Settings)
    _ = imported.NewRequest(1234)
    if after := mustMarshal(t, imported.Settings); !bytes.Equal(before, after) {
        t.Fatal("NewRequest mutated imported settings")
    }
}
```

The fixture must contain a supported API version, one player, equipment, encounter, party buffs, raid buffs, debuffs, tanks, and a deterministic RNG seed. It must be generated once by a test-only `encodeIndividualLink` helper and then stored as a literal fixture; the decoder test reads the literal rather than testing encode/decode round-trip only.

- [x] **Step 3: Implement side-effect-free export-link import**

Move the decode mechanics from `decode_link.go` into `upgrades.Import`: split exactly once at `#`, standard-base64 decode, zlib inflate, select `IndividualSimSettings` only for a non-raid URL, and `proto.Unmarshal`. Validate before returning: non-nil `Settings`, `Player`, `Encounter`, and `Settings.Settings`; supported API version; all equipped IDs exist in `database.Load().Items`; and the player has an equipment spec. Return the specified `ValidationError` instead of logging or exiting.

Build `RaidSimRequest` from the imported individual setting exactly as the individual wowsims UI does: one party containing the imported player and party buffs; imported raid buffs, debuffs, tanks, target dummies, and encounter; `SimTypeIndividual`; and sim options copied from imported `iterations`/`fixed_rng_seed`, with only the requested iteration count substituted. Preserve every other imported rotation, talent, consume, profession, equipment, and encounter value.

- [x] **Step 4: Make the existing CLI decoder reuse the import parser**

Change `decodelink` to call the shared low-level parser and print the decoded protobuf as ProtoJSON. It remains a one-argument inspection command; it must not silently accept a raid link in `rank-upgrades` merely because `decodelink` can display it.

- [x] **Step 5: Run import contracts**

Run: `rtk go test ./cmd/wowsimcli/cmd/upgrades -run 'TestImport|TestNewRequest' -count=1`

Expected: PASS; malformed/raid fixtures return a typed validation error and the baseline protobuf bytes match before/after conversion.

- [x] **Step 6: Commit the import boundary**

```bash
git add cmd/wowsimcli/cmd/decode_link.go cmd/wowsimcli/cmd/upgrades
git commit -m "feat: import individual wowsims settings"
```

### Task 3: Build an obtainable, legal candidate set without touching baseline gear

**Files:**
- Create: `cmd/wowsimcli/cmd/upgrades/catalog.go`
- Create: `cmd/wowsimcli/cmd/upgrades/candidates.go`
- Create: `cmd/wowsimcli/cmd/upgrades/candidates_test.go`
- Modify: `cmd/wowsimcli/cmd/upgrades/types.go`

**Interfaces:**
- Consumes: `ImportedSettings`, `ContentFilters`, `ItemPolicy`, and `assets/database.Load()`.
- Produces: `func BuildCandidates(imported *ImportedSettings, filters ContentFilters, policy ItemPolicy, catalog *Catalog) (BuildResult, error)` where `BuildResult` has `Candidates []Candidate`, `Excluded ExclusionSummary`, and no simulation result.
- `Candidate` contains `Item UIItemSummary`, `TargetSlot proto.ItemSlot`, `Displaced []UIItemSummary`, `Request *proto.RaidSimRequest`, `Applied PolicyApplication`, and `Source SourceSummary`.

- [x] **Step 1: Define filters, policy references, and exclusions as explicit data**

Add serializable types; use IDs/enums rather than UI labels as the source of truth:

```go
type ContentFilters struct {
    MaxPhase        int32                `json:"maxPhase"`
    SourceKinds     []proto.SourceFilterOption `json:"sourceKinds"`
    SourceNames     []string             `json:"sourceNames"`
    IncludeUnknown  bool                 `json:"includeUnknown"`
}

type ItemPolicy struct {
    GemBySocket map[proto.GemColor]int32 `json:"gemBySocket"`
    MaxGemQuality proto.ItemQuality      `json:"maxGemQuality"`
    EnchantByType map[proto.ItemType]int32 `json:"enchantByType"`
}

type ExclusionSummary struct {
    UnknownSource int            `json:"unknownSource"`
    Source        int            `json:"source"`
    Compatibility int            `json:"compatibility"`
    Policy        int            `json:"policy"`
    Reasons       map[string]int `json:"reasons"`
}
```

Normalize `SourceKinds`/`SourceNames` by sorting and deduplicating before hashing or filtering. Keep unknown-source exclusion separate from source-filter exclusion so the UI can state why candidates were not evaluated.

- [x] **Step 2: Write catalog and compatibility tests against tiny fixture data**

Construct a test `proto.UIDatabase` with: one legal raid item, one unknown-source item, a profession-gated item, a class-gated item, a unique item already equipped, two ring candidates, one two-hand weapon, one off-hand, and a faction-gated item. Exercise the externally visible contract:
```go

func TestBuildCandidatesExcludesUnknownSourceByDefault(t *testing.T) {
    result := mustBuild(t, ContentFilters{})
    if result.Excluded.UnknownSource != 1 || containsID(result.Candidates, 1002) {
        t.Fatalf("unknown source handling = %#v", result)
    }
}

func TestBuildCandidatesPreservesBaselineBytes(t *testing.T) {
    imported := fixtureImported(t)
    before := mustMarshal(t, imported.Settings)
    _ = mustBuildFrom(t, imported, ContentFilters{})
    if after := mustMarshal(t, imported.Settings); !bytes.Equal(before, after) {
        t.Fatal("BuildCandidates mutated baseline")
    }
}

func TestBuildCandidatesRejectsClassProfessionUniqueAndFaction(t *testing.T) {
    result := mustBuild(t, ContentFilters{IncludeUnknown: true})
    for _, id := range []int32{1003, 1004, 1005, 1006} {
        if containsID(result.Candidates, id) { t.Fatalf("ineligible item %d was proposed", id) }
    }
}

func TestBuildCandidatesEvaluatesBothRingSlots(t *testing.T) {
    if got := targetSlotsFor(mustBuild(t, ContentFilters{}).Candidates, 1007); !slices.Equal(got, []proto.ItemSlot{proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2}) {
        t.Fatalf("ring slots = %v", got)
    }
}

func TestBuildCandidatesModelsTwoHandOffHandConflict(t *testing.T) {
    candidate := candidateByIDAndSlot(t, mustBuild(t, ContentFilters{}).Candidates, 1008, proto.ItemSlot_ItemSlotMainHand)
    if !containsSlot(candidate.Displaced, proto.ItemSlot_ItemSlotOffHand) { t.Fatal("two-hand candidate retained off-hand") }
}
```

Use a deterministic `baselineBytes := mustMarshal(t, imported.Settings)` assertion after every build; no test may accept a mutation merely because a later clone hides it.

- [x] **Step 3: Load and index the bundled complete UI database once**

Create `NewCatalog(db *proto.UIDatabase) *Catalog` around `assets/database.Load()`. Index `UIItem`, `UIGem`, `UIEnchant`, NPCs, zones, and source names by ID. Resolve item source text from the database's crafted/drop/quest/vendor/reputation source variants; preserve phase and source-kind enum. The catalog is immutable after construction and shared by all jobs.

- [x] **Step 4: Implement filtering and all legal slot variants**

For each TBC `UIItem`, first require source metadata unless `IncludeUnknown` is true, then apply phase, source-kind, and optional normalized source-name filters. Check class allowlist, faction restriction, both imported professions, item type, armor proficiency, hand type/weapon requirements, and `limit_category` occupancy against the candidate equipment set.

Generate every legal replacement position. Evaluate finger and trinket items in both slots. For one-hand weapons create main-hand and off-hand variants when both are legal; for two-hand/main-hand items create a main-hand variant that also clears the displaced off-hand; for off-hand-only items create only an off-hand variant. Reject variants that leave invalid equipment or exceed unique/limit constraints. Reuse the upstream database-registration/conversion path to make the candidate item and selected gem/enchant records available to the cloned simulator request. Build each variant from `cloneMessage(imported.Settings)` and convert that clone into its own request—never edit the baseline or reuse an equipment slice.

- [x] **Step 5: Run candidate contracts**

Run: `rtk go test ./cmd/wowsimcli/cmd/upgrades -run TestBuildCandidates -count=1`

Expected: PASS; source/compatibility counts are correct, rings produce two variants, and all baseline byte-comparison assertions pass.

- [x] **Step 6: Commit candidate construction**

```bash
git add cmd/wowsimcli/cmd/upgrades
git commit -m "feat: construct legal upgrade candidates"
```

### Task 4: Apply only legal declared gems and enchants

**Files:**
- Create: `cmd/wowsimcli/cmd/upgrades/policy.go`
- Create: `cmd/wowsimcli/cmd/upgrades/policy_test.go`
- Modify: `cmd/wowsimcli/cmd/upgrades/candidates.go`
- Modify: `cmd/wowsimcli/cmd/upgrades/types.go`

**Interfaces:**
- Consumes: a candidate item, target slot, player class/professions, and `ItemPolicy`.
- Produces: `func ApplyPolicy(candidate Candidate, policy ItemPolicy, catalog *Catalog) (Candidate, *PolicyError)` and `PolicyApplication{GemIDs, EnchantID, SocketChoices}` included in every result.
- Rejects: missing selected database records, wrong gem colour, gem quality above policy, unique-gem conflict, profession/class-ineligible gems or enchants, and enchants in an ineligible slot/type.

- [x] **Step 1: Write policy contract tests first**

Use catalog fixture items with red/blue/meta sockets and eligible/ineligible enchant types:
```go

func TestApplyPolicyUsesOnlyLegalSocketChoices(t *testing.T) {
    got := mustApplyPolicy(t, fixtureSocketedCandidate(), fixturePolicy())
    if !slices.Equal(got.Applied.GemIDs, []int32{2001, 0, 0}) { t.Fatalf("gems = %v", got.Applied.GemIDs) }
}

func TestApplyPolicyRejectsOverQualityAndProfessionGems(t *testing.T) {
    _, err := ApplyPolicy(fixtureSocketedCandidate(), ItemPolicy{GemBySocket: map[proto.GemColor]int32{proto.GemColor_GemColorRed: 2002}}, fixtureCatalog())
    if err == nil { t.Fatal("over-quality or profession gem was accepted") }
}

func TestApplyPolicyRejectsIneligibleEnchant(t *testing.T) {
    _, err := ApplyPolicy(fixtureSocketedCandidate(), ItemPolicy{EnchantByType: map[proto.ItemType]int32{proto.ItemType_ItemTypeHead: 3002}}, fixtureCatalog())
    if err == nil { t.Fatal("ineligible enchant was accepted") }
}

func TestApplyPolicyDoesNotChangeBaselineEquipment(t *testing.T) {
    imported := fixtureImported(t)
    before := mustMarshal(t, imported.Settings)
    _ = mustApplyPolicy(t, fixtureSocketedCandidate(), fixturePolicy())
    if after := mustMarshal(t, imported.Settings); !bytes.Equal(before, after) { t.Fatal("policy mutated baseline") }
}
```

Also assert that the candidate report contains the selected gem IDs, enchant effect ID, and the policy values used; a reader must be able to distinguish an ungemed item from a policy-configured item.

- [x] **Step 2: Implement deterministic policy application**

For each candidate socket, choose only the configured `GemBySocket` record matching that socket colour or a legal prismatic interaction, at or below `MaxGemQuality`; preserve an empty socket when the policy has no legal selection. Enforce meta socket and unique-gem constraints using the full candidate equipment, not just the new item. Select an enchant only from `EnchantByType` when the selected `UIEnchant` permits the target item's type/extra type, class, profession, and phase. Store the result in the copied `ItemSpec` only after all policy checks pass.

Policy failure excludes that candidate under `policy`; it is not a simulator failure and it must not fall back to a different gem/enchant.

- [x] **Step 3: Run policy contracts**

Run: `rtk go test ./cmd/wowsimcli/cmd/upgrades -run TestApplyPolicy -count=1`

Expected: PASS; policy is visible in output, illegal combinations are absent, and baseline bytes remain identical.

- [x] **Step 4: Commit policy application**

```bash
git add cmd/wowsimcli/cmd/upgrades
git commit -m "feat: apply declared item policies"
```

### Task 5: Simulate, confirm, and rank with honest uncertainty

**Files:**
- Create: `cmd/wowsimcli/cmd/upgrades/simulator.go`
- Create: `cmd/wowsimcli/cmd/upgrades/service.go`
- Create: `cmd/wowsimcli/cmd/upgrades/service_test.go`
- Modify: `cmd/wowsimcli/cmd/upgrades/types.go`

**Interfaces:**
- Consumes: `RankRequest{Imported, Filters, Policy, Options}` and `context.Context`.
- Produces: `func (s *Service) RankUpgrades(ctx context.Context, request RankRequest, progress func(Progress)) (*UpgradeReport, error)`.
- `SimulationOptions` has user-selected `ScreeningIterations`, `ConfirmationIterations`, and server-validated values `ScreeningIterations > 0`, `ConfirmationIterations >= ScreeningIterations`.
- `UpgradeReport` contains baseline DPS/metadata, confirmed candidates, excluded counts, failed simulations, exact request assumptions, SHA-256 fingerprint, simulator/database revisions, and no internal simulator progress structures.

- [x] **Step 1: Write the service tests around a fake simulator**

Define a small adapter so ranking tests never run a full TBC simulation:

```go
type Simulator interface {
    Run(ctx context.Context, request *proto.RaidSimRequest, onProgress func(completed, total int32)) (DPSResult, error)
}

type DPSResult struct { Average, Stdev float64; Iterations int32 }
```

Test these required transitions with scripted `DPSResult` values:
```go

func TestRankUpgradesScreensThenConfirmsTopAndCloseCandidates(t *testing.T) {
    fake := newFakeSimulator(screeningResults(25, 20, 19.9))
    _, err := newService(fake).RankUpgrades(context.Background(), fixtureRankRequest(), nil)
    if err != nil { t.Fatal(err) }
    if got := fake.callsAt(fixtureOptions().ConfirmationIterations); got != 22 { t.Fatalf("confirmation calls = %d, want baseline + 21 finalists", got) }
}

func TestRankUpgradesMarksOverlappingIntervalsTooCloseToCall(t *testing.T) {
    report := mustRank(t, newFakeSimulator(overlappingResults()))
    if !report.Confirmed[0].TooCloseToCall || !report.Confirmed[1].TooCloseToCall { t.Fatal("overlap was ranked precisely") }
}

func TestRankUpgradesIsolatesCandidateFailure(t *testing.T) {
    report := mustRank(t, newFakeSimulator(oneFailedCandidate()))
    if len(report.Failed) != 1 || len(report.Confirmed) == 0 { t.Fatalf("failed=%d confirmed=%d", len(report.Failed), len(report.Confirmed)) }
}

func TestRankUpgradesCancellationReturnsNoReport(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background()); cancel()
    report, err := newService(newBlockingSimulator()).RankUpgrades(ctx, fixtureRankRequest(), nil)
    if report != nil || !errors.Is(err, context.Canceled) { t.Fatalf("report=%v err=%v", report, err) }
}

func TestRankUpgradesIncludesStableFingerprintAndRevisions(t *testing.T) {
    first, second := mustRank(t, newFakeSimulator(stableResults())), mustRank(t, newFakeSimulator(stableResults()))
    if first.AssumptionsFingerprint != second.AssumptionsFingerprint || first.SimulatorRevision == "" || first.DatabaseRevision == "" { t.Fatal("report provenance is unstable or absent") }
}
```

- [x] **Step 2: Implement the narrow real simulator adapter**

Implement `Simulator.Run` with `core.RunRaidSimConcurrentAsync(request, reporter, requestID)`. Forward only completed/total iteration counts to `onProgress`; wait for `FinalRaidResult`; return its error message as a Go error; and read the imported character's player DPS distribution from the first party/first player. Copy the request and give every run a unique request ID. On `ctx.Done()`, call `simsignals.AbortById(requestID)`, drain its reporter channel, and return `ctx.Err()`.

Do not expose `ProgressMetrics`, simulator channels, or core request IDs through the HTTP/UI contract.

- [x] **Step 3: Implement two-pass selection and confidence intervals**

Run the unchanged baseline at screening iterations, then every legal candidate at screening iterations. Let a candidate delta be `candidate.Average - baseline.Average`; calculate each delta standard error as:

```text
sqrt((baseline.Stdev^2 / baseline.Iterations) + (candidate.Stdev^2 / candidate.Iterations))
```

Use the two-sided 95% interval `delta ± 1.96 * standardError`. Sort screening candidates by descending delta. Confirm the first 20 candidates plus candidates whose 95% interval overlaps the 20th candidate's interval, capped at 50 total; the report records the cap when it truncates a tie region. Re-run the baseline at confirmation iterations and re-run only that finalist set at confirmation iterations. Sort the confirmed deltas; candidates with overlapping 95% intervals must share a `tooCloseToCall: true` marker and the UI must not assign them distinct ordinal ranks. Do not include screened-only candidates in `Confirmed`.

Run candidates with a bounded worker count `min(runtime.GOMAXPROCS(0), 8)`, returning work through channels and checking context before scheduling each candidate. This is intentionally a simple local CPU bound; add a `// ponytail:` comment only if profiling proves a different scheduler is needed.

- [x] **Step 4: Fingerprint all material assumptions**

Canonicalize the deterministic imported-settings protobuf bytes and a JSON structure containing normalized filters, normalized policy, options, simulator revision, and database revision. SHA-256 that byte sequence and set `AssumptionsFingerprint` to lower-case hexadecimal. Include the same structured assumptions alongside the hash so a copied report is interpretable without re-importing the link.

- [x] **Step 5: Run service contracts**

Run: `rtk go test ./cmd/wowsimcli/cmd/upgrades -run TestRankUpgrades -count=1`

Expected: PASS; only finalists receive confirmation, overlaps are marked, one failure does not erase successes, and cancellation returns no report.

- [x] **Step 6: Commit the ranker**

```bash
git add cmd/wowsimcli/cmd/upgrades
git commit -m "feat: rank confirmed single-item upgrades"
```

### Task 6: Expose the ranker through a local, cancellable JSON server

**Files:**
- Create: `cmd/wowsimcli/cmd/upgrade_server.go`
- Create: `cmd/wowsimcli/cmd/upgrade_server_test.go`
- Modify: `cmd/wowsimcli/cmd/rank_upgrades.go`

**Interfaces:**
- Consumes: `upgrades.Service` and local HTTP requests.
- Produces:
  - `POST /api/import` with a JSON body whose required `link` property is a wowsims individual export URL → decoded character/settings summary, defaults, and no simulation.
  - `POST /api/jobs` with `RankJobInput{Link, Filters, Policy, Options}` JSON → `202` and a generated job ID; the server re-imports `Link` and constructs the non-JSON `RankRequest` itself.
  - `GET /api/jobs/{id}` → `queued|running|completed|failed|canceled`, progress, validation/error payload, and report only when `completed`.
  - `DELETE /api/jobs/{id}` → cancellation acknowledgement; a canceled job never has `report`.
  - `GET /` and `/assets/*` → embedded UI only.

- [x] **Step 1: Write server contract tests with a fake rank service**

Exercise endpoint behavior with `httptest.NewServer` and a fake that blocks until its context is cancelled:
```go

func TestImportReturnsSummaryWithoutStartingJob(t *testing.T) {
    response := postJSON(t, server.URL+"/api/import", fixtureImportBody())
    if response.StatusCode != http.StatusOK || fake.Runs() != 0 { t.Fatalf("status=%d runs=%d", response.StatusCode, fake.Runs()) }
}

func TestCreateJobPollsCompletedReport(t *testing.T) {
    id := createJob(t, server.URL, fixtureRankBody())
    if job := waitForJob(t, server.URL, id); job.Status != "completed" || job.Report == nil { t.Fatalf("job = %#v", job) }
}

func TestInvalidImportIsSpecificAndDoesNotCreateJob(t *testing.T) {
    response := postJSON(t, server.URL+"/api/import", map[string]string{"link": "not-a-link"})
    if response.StatusCode != http.StatusBadRequest || decodeError(t, response).Code != "invalid_link" { t.Fatal("invalid link did not return validation error") }
}

func TestCancelJobReturnsCanceledWithoutPartialReport(t *testing.T) {
    id := createJob(t, server.URL, fixtureRankBody())
    deleteJob(t, server.URL, id)
    if job := waitForJob(t, server.URL, id); job.Status != "canceled" || job.Report != nil { t.Fatalf("job = %#v", job) }
}
```

- [x] **Step 2: Implement one in-memory job manager**

Use a mutex-protected `map[string]*job` with a `context.CancelFunc` per job and `crypto/rand`/hex job IDs. Store only the request summary, status, progress counts, final report, and error. A `POST /api/jobs` imports the submitted link and validates its filters, policy, and options synchronously, then starts one goroutine that calls `RankUpgrades`; progress updates are counters only. Cancel removes queued work where possible, signals the running context, and clears any report. Set `Cache-Control: no-store` on API responses; do not write reports, links, or characters to disk or browser storage.

Bind the listener only to loopback. Reject non-JSON bodies, unknown fields, oversized bodies (1 MiB), unsupported methods, and paths outside the explicit route table with structured errors.

- [x] **Step 3: Register static assets and browser behavior**

Embed `upgrade_ui/*` with `embed.FS`; serve `index.html` from `/` and assets under `/assets/` with content types and no directory listing. `rank-upgrades` opens the resolved local URL through the existing `github.com/pkg/browser` dependency unless `--no-browser` was selected. Graceful command shutdown cancels every active job and calls `http.Server.Shutdown`.

- [x] **Step 4: Run server contracts**

Run: `rtk go test ./cmd/wowsimcli/cmd -run 'TestImport|TestCreateJob|TestInvalidImport|TestCancelJob' -count=1`

Expected: PASS; import does not simulate, and canceled jobs expose no partial result.

- [x] **Step 5: Commit the HTTP boundary**

```bash
git add cmd/wowsimcli/cmd/rank_upgrades.go cmd/wowsimcli/cmd/upgrade_server.go cmd/wowsimcli/cmd/upgrade_server_test.go
git commit -m "feat: serve local upgrade ranking jobs"
```

### Task 7: Build the browser flow and visible report assumptions

**Files:**
- Create: `cmd/wowsimcli/cmd/upgrade_ui/index.html`
- Create: `cmd/wowsimcli/cmd/upgrade_ui/app.js`
- Create: `cmd/wowsimcli/cmd/upgrade_ui/app.css`
- Modify: `cmd/wowsimcli/cmd/upgrade_server_test.go`

**Interfaces:**
- Consumes: the exact local HTTP API from Task 6.
- Produces: an accessible, no-build browser UI with import, decoded-summary review, filters/policy/options selection, progress/cancellation, report display, and report-copying.

- [x] **Step 1: Add static-page route coverage before UI behavior**

Extend the server test to assert that `/` returns the form landmarks and that `/assets/app.js` is served:

```go
func TestUpgradePageAndAssetsAreServed(t *testing.T) {
    // GET / includes <main>, import form, and wowsims attribution URL.
    // GET /assets/app.js returns JavaScript content.
}
```

- [x] **Step 2: Implement the import-and-review form**

Build semantic HTML with a labelled URL input, an `Import settings` button, live error region, and disabled ranking controls until `/api/import` succeeds. Render decoded player name/class/spec, equipped-item count, profession pair, phase, iterations, encounter, and settings digest before allowing a run. Include phase, source kind, optional source-name inputs; socket-colour gem selections; max gem quality; slot/type enchant selections; and positive screening/confirmation iteration inputs. Validate simple empty/non-positive input in the browser, but rely on the server for every authority decision.

- [x] **Step 3: Implement job progress, cancellation, and report rendering**

On successful job creation, poll `GET /api/jobs/{id}` every 500 ms only while status is `queued` or `running`; show completed/total candidate runs and a `Cancel ranking` button. Stop polling on all terminal states. For `completed`, render baseline DPS, revisions, fingerprint, normalized filters/policy/options, excluded counts, failed candidates/reasons, and a table of item, target slot, displaced item(s), phase/source, gems/enchant, DPS delta, relative gain, iterations, 95% interval, and material assumptions. Render `tooCloseToCall` candidates with the shared label `Too close to call`; do not render numeric rank positions inside that group.

Add `Copy assumptions and result` using `navigator.clipboard.writeText(JSON.stringify(report, null, 2))`; show success/failure in the live region. On cancellation, clear table rows and display only the canceled status—never stale data from a prior job.

- [x] **Step 4: Add attribution and non-goal copy**

Place this visible footer text and link in `index.html`:

```html
<p>
  Simulation and item data provided by
  <a href="https://github.com/wowsims/tbc-new">wowsims/tbc-new</a>.
  Rankings are local single-item DPS comparisons, not acquisition pricing or multi-item optimization.
</p>
```

- [x] **Step 5: Run static-route tests**

Run: `rtk go test ./cmd/wowsimcli/cmd -run TestUpgradePageAndAssetsAreServed -count=1`

Expected: PASS; the local page and JavaScript are reachable and the attribution is present.

- [x] **Step 6: Commit the browser UI**

```bash
git add cmd/wowsimcli/cmd/upgrade_ui cmd/wowsimcli/cmd/upgrade_server_test.go
git commit -m "feat: add upgrade finder browser UI"
```

### Task 8: Lock contract coverage, run the real smoke test, and document operation

**Files:**
- Create: `cmd/wowsimcli/cmd/upgrades/testdata/fixed_individual_link.txt`
- Create: `cmd/wowsimcli/cmd/upgrades/testdata/fixed_import_summary.json`
- Create: `docs/upgrade-finder.md`
- Modify: `cmd/wowsimcli/cmd/upgrades/import_test.go`
- Modify: `cmd/wowsimcli/cmd/upgrades/candidates_test.go`
- Modify: `cmd/wowsimcli/cmd/upgrades/policy_test.go`
- Modify: `cmd/wowsimcli/cmd/upgrades/service_test.go`

**Interfaces:**
- Consumes: the completed local binary, immutable fixed fixture, and a small known source filter.
- Produces: reproducible contract coverage and a documented manual browser smoke procedure proving the real local surface.

- [x] **Step 1: Complete the required contract matrix**

Confirm the focused tests collectively enforce every spec verification point:

| Contract | Test |
| --- | --- |
| Fixed export link decodes to expected settings | `TestImportDecodesFixedIndividualLink` compares to `fixed_import_summary.json`. |
| Filter yields expected candidate IDs | `TestBuildCandidatesFiltersExpectedIDs`. |
| Baseline never mutates | `TestNewRequestDoesNotMutateImportedSettings`, `TestBuildCandidatesPreservesBaselineBytes`, and `TestApplyPolicyDoesNotChangeBaselineEquipment`. |
| Class/profession/unique/ring/weapon/source rules | the named candidate tests from Task 3. |
| Policy applies only legally | the policy tests from Task 4. |
| Every result carries fingerprint and revisions | `TestRankUpgradesIncludesStableFingerprintAndRevisions`. |
| Failed candidates and cancellation stay isolated | `TestRankUpgradesIsolatesCandidateFailure` and `TestRankUpgradesCancellationReturnsNoReport`. |

Add `TestBuildCandidatesFiltersExpectedIDs` with a fixture filter that permits only the fixture raid source and assert a sorted exact ID list, for example:

```go
if got := candidateIDs(result.Candidates); !slices.Equal(got, []int32{1001, 1002}) {
    t.Fatalf("candidate IDs = %v, want [1001 1002]", got)
}
```

Use a local `slices.Equal` comparison instead of adding `go-cmp` if the fork does not already depend on it.

- [x] **Step 2: Document exact local usage and boundaries**

Write `docs/upgrade-finder.md` with these commands and behavior:

```bash
rtk go build -o wowsimcli ./cmd/wowsimcli
./wowsimcli rank-upgrades
```

Document the five UI steps (paste/import/review/filter-and-policy/run/copy), link constraints, loopback-only operation, the simulator/database revisions in results, uncertainty interpretation, cancellation semantics, and each explicitly deferred non-goal: Armory import, scheduled monitoring, MCP, pricing, multi-item optimization, non-DPS metrics, accounts, and remote persistence. Include the same required wowsims attribution link.

- [x] **Step 3: Run all deterministic Go contracts**

Run: `rtk go test ./cmd/wowsimcli/cmd/... -count=1`

Expected: PASS.

- [x] **Step 4: Execute the real browser smoke test**

Build and run the binary with `--no-browser`, open the printed loopback URL in a browser, paste `fixed_individual_link.txt`, select the fixture's small source set and legal policy, and run with the fixture's bounded iteration values. Verify all of these on the rendered page:

1. decoded character/settings summary appears before ranking;
2. progress changes while the real simulator runs;
3. the completed report has at least one candidate;
4. that candidate shows its expected source plus gems/enchant, DPS delta, interval, iterations, and assumptions;
5. simulator/database revisions, assumptions fingerprint, and visible wowsims link are present;
6. `Copy assumptions and result` places parseable report JSON on the clipboard.

Use browser automation against the running local page for this step and save its screenshot/output as the implementation evidence. Do not replace this with an HTTP-only test.

- [x] **Step 5: Commit verification and documentation**

```bash
git add cmd/wowsimcli/cmd/upgrades/testdata cmd/wowsimcli/cmd/upgrades docs/upgrade-finder.md
git commit -m "test: verify local upgrade finder"
```

## Plan Self-Review

- **Spec coverage:** Tasks 1–2 cover the one-binary/local/wowsims-import foundation. Tasks 3–4 cover content filtering, unknown source handling, compatibility, slot conflicts, immutable baseline copies, and declared policy. Task 5 covers baseline/candidate simulation, two-pass confirmation, uncertainty, isolated failures, cancellation, revisions, and fingerprinting. Tasks 6–7 cover the required local UI flow, progress, copyable results, and attribution. Task 8 maps every listed contract and browser smoke requirement to a named test or concrete exercise.
- **Intentional exclusions:** No Armory/profile seam, monitoring, MCP, price data, persistence, static scores, or multi-item optimizer is introduced.
- **Consistency:** `ImportedSettings`, `ContentFilters`, `ItemPolicy`, `SimulationOptions`, `RankRequest`, `UpgradeReport`, and `RankUpgrades` are the sole cross-task contracts. Every later task consumes those names unchanged.
- **Placeholder scan:** No implementation step defers a required behavior. The fixed test URL is generated from the committed controlled protobuf fixture and committed as a literal before its decoder assertion, so tests remain independent of upstream hosting.
