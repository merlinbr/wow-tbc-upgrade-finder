package upgrades

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"runtime"
	"sort"
	"sync"

	"github.com/wowsims/tbc/assets/database"
	"github.com/wowsims/tbc/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

const (
	finalistTargetCount = 20
	finalistHardCap     = 50
	// Two-sided 95% confidence multiplier.
	z95 = 1.96
)

type Service struct {
	simulator Simulator
	catalog   *Catalog
}

func NewRankService(sim Simulator, catalog *Catalog) *Service {
	if catalog == nil {
		catalog = NewCatalog(database.Load())
	}
	return &Service{simulator: sim, catalog: catalog}
}

type simJob struct {
	req       *proto.RaidSimRequest
	candidate *Candidate // nil for baseline runs
}

// withIterations returns a copy of the request with the stage's iteration
// count substituted. Candidate requests are built once with a zero placeholder.
func withIterations(req *proto.RaidSimRequest, iterations int32) *proto.RaidSimRequest {
	if req == nil {
		return nil
	}
	copied := cloneMessage(req)
	if copied.SimOptions == nil {
		copied.SimOptions = &proto.SimOptions{}
	}
	copied.SimOptions.Iterations = iterations
	return copied
}

func (s *Service) RankUpgrades(ctx context.Context, request RankRequest, progress func(Progress)) (*UpgradeReport, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if request.Imported == nil || request.Imported.Settings == nil {
		return nil, &ValidationError{Code: "incompatible_settings", Message: "missing imported settings"}
	}
	if request.Options.ScreeningIterations <= 0 {
		return nil, fmt.Errorf("screeningIterations must be greater than 0")
	}
	if request.Options.ConfirmationIterations < request.Options.ScreeningIterations {
		return nil, fmt.Errorf("confirmationIterations must be greater than or equal to screeningIterations")
	}

	build, err := BuildCandidates(request.Imported, request.Filters, request.Policy, s.catalog)
	if err != nil {
		return nil, err
	}

	// Apply the declared policy; policy failures exclude the candidate.
	var valid []Candidate
	for _, c := range build.Candidates {
		applied, policyErr := ApplyPolicy(c, request.Policy, s.catalog)
		if policyErr != nil {
			build.Excluded.Policy++
			build.Excluded.Reasons["policy"]++
			continue
		}
		valid = append(valid, applied)
	}

	fingerprint, assumptions := computeFingerprint(request)
	tracker := newProgressTracker(progress)

	// Stage 1: baseline at screening budget.
	baselineScreeningReq := withIterations(request.Imported.NewRequest(request.Options.ScreeningIterations), request.Options.ScreeningIterations)
	baselineScreening, err := s.runSingle(ctx, baselineScreeningReq, "screening", int32(len(valid)+1), tracker)
	if err != nil {
		return nil, err
	}

	// Stage 2: every legal candidate at screening budget.
	screenJobs := make([]*proto.RaidSimRequest, len(valid))
	for i := range valid {
		screenJobs[i] = withIterations(valid[i].Request, request.Options.ScreeningIterations)
	}
	screenResults, screenErrs, err := s.runJobs(ctx, screenJobs, "screening", int32(len(valid)+1), tracker)
	if err != nil {
		return nil, err
	}

	// Collect failed candidates from screening.
	failedByKey := make(map[string]FailedCandidate)
	var screened []scoredCandidate
	for i := range screenJobs {
		if screenErrs[i] != nil {
			failedByKey[failedKey(valid[i])] = FailedCandidate{
				Item:       valid[i].Item,
				TargetSlot: valid[i].TargetSlot,
				Reason:     screenErrs[i].Error(),
			}
			continue
		}
		delta, se := deltaInterval(baselineScreening, screenResults[i])
		screened = append(screened, scoredCandidate{
			candidate: valid[i],
			delta:     delta,
			se:        se,
			low:       delta - z95*se,
			high:      delta + z95*se,
		})
	}

	sortScreened(screened)

	// Two-pass selection: first 20 plus candidates whose 95% interval overlaps
	// the 20th candidate's interval, capped at 50 total.
	finalistCount := min(finalistTargetCount, len(screened))
	capTruncated := false
	if len(screened) > finalistTargetCount {
		cutoff := screened[finalistTargetCount-1]
		for i := finalistTargetCount; i < len(screened); i++ {
			if finalistCount >= finalistHardCap {
				if intervalsOverlap(screened[i], cutoff) {
					capTruncated = true
				}
				break
			}
			if !intervalsOverlap(screened[i], cutoff) {
				break
			}
			finalistCount++
		}
	}
	finalists := screened[:finalistCount]

	// Stage 3: baseline at confirmation budget.
	baselineConfirmReq := withIterations(request.Imported.NewRequest(request.Options.ConfirmationIterations), request.Options.ConfirmationIterations)
	baselineConfirmed, err := s.runSingle(ctx, baselineConfirmReq, "confirmation", int32(len(finalists)+1), tracker)
	if err != nil {
		return nil, err
	}

	// Stage 4: finalists at confirmation budget.
	confirmJobs := make([]*proto.RaidSimRequest, len(finalists))
	for i := range finalists {
		confirmJobs[i] = withIterations(finalists[i].candidate.Request, request.Options.ConfirmationIterations)
	}
	confirmResults, confirmErrs, err := s.runJobs(ctx, confirmJobs, "confirmation", int32(len(finalists)+1), tracker)
	if err != nil {
		return nil, err
	}

	var confirmed []scoredCandidate
	for i := range confirmJobs {
		if confirmErrs[i] != nil {
			failedByKey[failedKey(finalists[i].candidate)] = FailedCandidate{
				Item:       finalists[i].candidate.Item,
				TargetSlot: finalists[i].candidate.TargetSlot,
				Reason:     confirmErrs[i].Error(),
			}
			continue
		}
		delta, se := deltaInterval(baselineConfirmed, confirmResults[i])
		confirmed = append(confirmed, scoredCandidate{
			candidate: finalists[i].candidate,
			delta:     delta,
			se:        se,
			low:       delta - z95*se,
			high:      delta + z95*se,
		})
	}

	sortScreened(confirmed)

	// Candidates whose 95% intervals overlap share the tooCloseToCall marker.
	for i := 1; i < len(confirmed); i++ {
		if intervalsOverlap(confirmed[i], confirmed[i-1]) {
			confirmed[i].tooClose = true
			confirmed[i-1].tooClose = true
		}
	}

	failed := make([]FailedCandidate, 0, len(failedByKey))
	for _, f := range failedByKey {
		failed = append(failed, f)
	}

	report := &UpgradeReport{
		Baseline: BaselineSummary{
			Dps:        baselineConfirmed.Average,
			DpsStdev:   baselineConfirmed.Stdev,
			Iterations: baselineConfirmed.Iterations,
		},
		Character:              request.Imported.Character,
		Confirmed:              makeConfirmed(confirmed, baselineConfirmed, request.Options, assumptions),
		Excluded:               build.Excluded,
		Failed:                 failed,
		Assumptions:            assumptions,
		AssumptionsFingerprint: fingerprint,
		SimulatorRevision:      SimulatorRevision,
		DatabaseRevision:       DatabaseRevision,
		CapTruncatedTieRegion:  capTruncated,
	}

	return report, nil
}

func (s *Service) runSingle(ctx context.Context, req *proto.RaidSimRequest, stage string, total int32, tracker *progressTracker) (DPSResult, error) {
	if err := ctx.Err(); err != nil {
		return DPSResult{}, err
	}
	result, err := s.simulator.Run(ctx, req, nil)
	tracker.complete(stage, total)
	return result, err
}

// runJobs executes jobs with a bounded worker count, checking the context
// before scheduling each candidate.
func (s *Service) runJobs(ctx context.Context, reqs []*proto.RaidSimRequest, stage string, total int32, tracker *progressTracker) ([]DPSResult, []error, error) {
	results := make([]DPSResult, len(reqs))
	errs := make([]error, len(reqs))
	if len(reqs) == 0 {
		return results, errs, nil
	}

	workers := min(runtime.GOMAXPROCS(0), 8)
	workers = min(workers, len(reqs))
	if workers < 1 {
		workers = 1
	}

	jobsCh := make(chan int)
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobsCh {
				results[idx], errs[idx] = s.simulator.Run(ctx, reqs[idx], nil)
				tracker.complete(stage, total)
			}
		}()
	}

	canceled := false
	for i := range reqs {
		if err := ctx.Err(); err != nil {
			canceled = true
			break
		}
		jobsCh <- i
	}
	close(jobsCh)
	wg.Wait()

	if canceled || ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}
	return results, errs, nil
}

type progressTracker struct {
	mu        sync.Mutex
	completed map[string]int32
	fn        func(Progress)
}

func newProgressTracker(fn func(Progress)) *progressTracker {
	return &progressTracker{completed: make(map[string]int32), fn: fn}
}

func (p *progressTracker) complete(stage string, total int32) {
	if p == nil || p.fn == nil {
		return
	}
	p.mu.Lock()
	p.completed[stage]++
	completed := p.completed[stage]
	p.mu.Unlock()
	p.fn(Progress{Stage: stage, Completed: completed, Total: total})
}

type scoredCandidate struct {
	candidate Candidate
	delta     float64
	se        float64
	low       float64
	high      float64
	tooClose  bool
}

func deltaInterval(base DPSResult, candidate DPSResult) (delta, se float64) {
	baseIters := math.Max(float64(base.Iterations), 1)
	candIters := math.Max(float64(candidate.Iterations), 1)
	delta = candidate.Average - base.Average
	variance := base.Stdev*base.Stdev/baseIters + candidate.Stdev*candidate.Stdev/candIters
	se = math.Sqrt(variance)
	return delta, se
}

func intervalsOverlap(a, b scoredCandidate) bool {
	return a.low <= b.high && a.high >= b.low
}

func sortScreened(list []scoredCandidate) {
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].delta != list[j].delta {
			return list[i].delta > list[j].delta
		}
		return list[i].candidate.Item.ID < list[j].candidate.Item.ID
	})
}

func failedKey(c Candidate) string {
	return fmt.Sprintf("%d:%d", c.Item.ID, c.TargetSlot)
}

func makeConfirmed(list []scoredCandidate, baseline DPSResult, options SimulationOptions, assumptions ReportAssumptions) []ConfirmedUpgrade {
	out := make([]ConfirmedUpgrade, 0, len(list))
	for i, sc := range list {
		relativeGain := 0.0
		if baseline.Average != 0 {
			relativeGain = sc.delta / baseline.Average * 100
		}
		out = append(out, ConfirmedUpgrade{
			Rank:                 i + 1,
			Item:                 sc.candidate.Item,
			TargetSlot:           sc.candidate.TargetSlot,
			Displaced:            sc.candidate.Displaced,
			Source:               sc.candidate.Source,
			Applied:              sc.candidate.Applied,
			DpsDelta:             sc.delta,
			RelativeGainPercent:  relativeGain,
			StandardError:        sc.se,
			ConfidenceInterval95: [2]float64{sc.low, sc.high},
			Iterations:           options.ConfirmationIterations,
			TooCloseToCall:       sc.tooClose,
			Assumptions:          assumptions,
		})
	}
	return out
}

// computeFingerprint hashes the deterministic imported-settings protobuf bytes
// together with a canonical JSON structure of the normalized filters, policy,
// options, and pinned revisions.
func computeFingerprint(request RankRequest) (string, ReportAssumptions) {
	normalized := normalizeFilters(request.Filters)

	gemBySocket := make(map[string]int32, len(request.Policy.GemBySocket))
	for color, id := range request.Policy.GemBySocket {
		gemBySocket[color.String()] = id
	}
	enchantByType := make(map[string]int32, len(request.Policy.EnchantByType))
	for itemType, id := range request.Policy.EnchantByType {
		enchantByType[itemType.String()] = id
	}

	sourceKinds := make([]string, 0, len(normalized.SourceKinds))
	for _, kind := range normalized.SourceKinds {
		sourceKinds = append(sourceKinds, kind.String())
	}
	sourceNames := append([]string(nil), normalized.SourceNames...)
	sort.Strings(sourceNames)

	assumptions := ReportAssumptions{
		LinkDigest:             request.Imported.SettingsDigest,
		MaxPhase:               normalized.MaxPhase,
		SourceKinds:            sourceKinds,
		SourceNames:            sourceNames,
		IncludeUnknown:         normalized.IncludeUnknown,
		MaxGemQuality:          request.Policy.MaxGemQuality.String(),
		GemBySocket:            gemBySocket,
		EnchantByType:          enchantByType,
		ScreeningIterations:    request.Options.ScreeningIterations,
		ConfirmationIterations: request.Options.ConfirmationIterations,
	}

	assumptionsJSON, err := json.Marshal(assumptions)
	if err != nil {
		assumptionsJSON = []byte{}
	}
	settingsBytes, err := googleProto.MarshalOptions{Deterministic: true}.Marshal(request.Imported.Settings)
	if err != nil {
		settingsBytes = nil
	}

	hash := sha256.New()
	hash.Write(settingsBytes)
	hash.Write(assumptionsJSON)
	hash.Write([]byte(SimulatorRevision))
	hash.Write([]byte(DatabaseRevision))

	return hex.EncodeToString(hash.Sum(nil)), assumptions
}
