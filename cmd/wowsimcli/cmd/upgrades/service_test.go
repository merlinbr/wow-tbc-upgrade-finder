package upgrades

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/wowsims/tbc/sim/core/proto"
)

// baselineChestID is the chest item equipped by the fixed individual link
// fixture; used by fakes to key scripted DPS results.
const baselineChestID = 29077

type fakeSimulator struct {
	mu        sync.Mutex
	results   map[int32]DPSResult // keyed by chest-slot item ID
	failures  map[int32]string    // chest item ID -> error
	fallback  DPSResult
	blocking  bool
	callCount map[int32]int
}

func newFakeSimulator() *fakeSimulator {
	return &fakeSimulator{
		results:   make(map[int32]DPSResult),
		failures:  make(map[int32]string),
		callCount: make(map[int32]int),
	}
}

func (f *fakeSimulator) Run(ctx context.Context, request *proto.RaidSimRequest, onProgress func(completed, total int32)) (DPSResult, error) {
	f.mu.Lock()
	iterations := request.GetSimOptions().GetIterations()
	f.callCount[iterations]++
	f.mu.Unlock()

	if f.blocking {
		<-ctx.Done()
		return DPSResult{}, ctx.Err()
	}

	chestID := request.GetRaid().GetParties()[0].GetPlayers()[0].GetEquipment().GetItems()[int(proto.ItemSlot_ItemSlotChest)].GetId()
	if reason, ok := f.failures[chestID]; ok {
		return DPSResult{}, fmt.Errorf("%s", reason)
	}
	if result, ok := f.results[chestID]; ok {
		return result, nil
	}
	if f.fallback.Average != 0 {
		return f.fallback, nil
	}
	return DPSResult{}, fmt.Errorf("no scripted result for chest item %d", chestID)
}

func (f *fakeSimulator) callsAt(iterations int32) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.callCount[iterations]
}

func fixtureOptions() SimulationOptions {
	return SimulationOptions{
		ScreeningIterations:    100,
		ConfirmationIterations: 200,
	}
}

// fixtureRankCatalog returns a catalog with 25 legal chest items (2001-2025).
func fixtureRankCatalog() *Catalog {
	items := make([]*proto.UIItem, 0, 25)
	for i := range 25 {
		id := int32(2001 + i)
		items = append(items, &proto.UIItem{
			Id:    id,
			Name:  fmt.Sprintf("Chest %d", id),
			Type:  proto.ItemType_ItemTypeChest,
			Phase: 1,
			Sources: []*proto.UIItemSource{
				{
					Source: &proto.UIItemSource_Drop{
						Drop: &proto.DropSource{Difficulty: proto.DungeonDifficulty_DifficultyRaid25},
					},
				},
			},
		})
	}
	return NewCatalog(&proto.UIDatabase{Items: items})
}

func newService(sim Simulator) *Service {
	return NewRankService(sim, fixtureRankCatalog())
}

func fixtureRankRequest(t *testing.T) RankRequest {
	t.Helper()
	return RankRequest{
		Imported: mustImportFixture(t),
		Filters:  ContentFilters{},
		Policy:   ItemPolicy{},
		Options:  fixtureOptions(),
	}
}

// screeningResults scripts: baseline 1000 DPS; candidates 2001-2025 get
// descending deltas 100..76 except candidates 2020/2021 which tie at 81 so the
// 20th candidate's tie region extends one candidate past the top-20 cutoff.
func screeningResults(baseline float64, first float64, tieAtCutoff float64) func(*fakeSimulator) {
	return func(f *fakeSimulator) {
		f.results[baselineChestID] = DPSResult{Average: baseline, Stdev: 0, Iterations: 100}
		for i := range 25 {
			id := int32(2001 + i)
			delta := first - float64(i)
			if id >= 2020 && id <= 2021 {
				delta = tieAtCutoff
			}
			f.results[id] = DPSResult{Average: baseline + delta, Stdev: 0, Iterations: 100}
		}
	}
}

func applyScript(f *fakeSimulator, script func(*fakeSimulator)) *fakeSimulator {
	script(f)
	return f
}

func mustRank(t *testing.T, sim Simulator) *UpgradeReport {
	t.Helper()
	report, err := newService(sim).RankUpgrades(context.Background(), fixtureRankRequest(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func TestRankUpgradesScreensThenConfirmsTopAndCloseCandidates(t *testing.T) {
	fake := applyScript(newFakeSimulator(), screeningResults(1000, 100, 81))
	_, err := newService(fake).RankUpgrades(context.Background(), fixtureRankRequest(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := fake.callsAt(fixtureOptions().ConfirmationIterations); got != 22 {
		t.Fatalf("confirmation calls = %d, want baseline + 21 finalists", got)
	}
}

func TestRankUpgradesMarksOverlappingIntervalsTooCloseToCall(t *testing.T) {
	fake := newFakeSimulator()
	fake.results[baselineChestID] = DPSResult{Average: 1000, Stdev: 10, Iterations: 100}
	// All candidates tie with a wide interval: every confirmed pair overlaps.
	fake.fallback = DPSResult{Average: 1050, Stdev: 10, Iterations: 100}

	report := mustRank(t, fake)
	if len(report.Confirmed) < 2 {
		t.Fatalf("confirmed = %d, want at least 2", len(report.Confirmed))
	}
	if !report.Confirmed[0].TooCloseToCall || !report.Confirmed[1].TooCloseToCall {
		t.Fatal("overlap was ranked precisely")
	}
}

func TestRankUpgradesIsolatesCandidateFailure(t *testing.T) {
	fake := newFakeSimulator()
	fake.results[baselineChestID] = DPSResult{Average: 1000, Stdev: 5, Iterations: 100}
	fake.fallback = DPSResult{Average: 1010, Stdev: 5, Iterations: 100}
	fake.failures[2005] = "candidate exploded"

	report := mustRank(t, fake)
	if len(report.Failed) != 1 || len(report.Confirmed) == 0 {
		t.Fatalf("failed=%d confirmed=%d", len(report.Failed), len(report.Confirmed))
	}
}

func TestRankUpgradesCancellationReturnsNoReport(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	blocking := &fakeSimulator{blocking: true}
	report, err := newService(blocking).RankUpgrades(ctx, fixtureRankRequest(t), nil)
	if report != nil || !errors.Is(err, context.Canceled) {
		t.Fatalf("report=%v err=%v", report, err)
	}
}

func TestRankUpgradesIncludesStableFingerprintAndRevisions(t *testing.T) {
	fake := newFakeSimulator()
	fake.results[baselineChestID] = DPSResult{Average: 1000, Stdev: 5, Iterations: 100}
	fake.fallback = DPSResult{Average: 1010, Stdev: 5, Iterations: 100}

	first := mustRank(t, fake)
	second := mustRank(t, fake)
	if first.AssumptionsFingerprint != second.AssumptionsFingerprint || first.SimulatorRevision == "" || first.DatabaseRevision == "" {
		t.Fatal("report provenance is unstable or absent")
	}
	if first.AssumptionsFingerprint == "" {
		t.Fatal("fingerprint is empty")
	}
}
