package upgrades

import (
	"context"
	"testing"

	"github.com/wowsims/tbc/sim"
)

func init() {
	sim.RegisterAll()
}

// TestRealSimulatorRunsFixtureBaseline exercises the real wowsims engine end
// to end through the adapter: import fixture, convert, attach DB, run 50
// iterations, and expect a positive player DPS.
func TestRealSimulatorRunsFixtureBaseline(t *testing.T) {
	imported := mustImportFixture(t)
	request := withIterations(imported.NewRequest(50), 50)

	sim := NewRealSimulator()
	result, err := sim.Run(context.Background(), request, nil)
	if err != nil {
		t.Fatalf("simulator run failed: %v", err)
	}
	if result.Average <= 0 {
		t.Fatalf("player DPS = %v, want > 0 (stdev %v, iterations %d)", result.Average, result.Stdev, result.Iterations)
	}
	t.Logf("player DPS = %.1f ± %.1f over %d iterations", result.Average, result.Stdev, result.Iterations)
}
