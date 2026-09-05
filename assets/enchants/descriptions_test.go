package enchants

import "testing"

func TestDescriptionsResolveKnownEffects(t *testing.T) {
	descriptions := Descriptions()
	if descriptions[2673] != "Mongoose" {
		t.Fatalf("descriptions[2673] = %q, want %q", descriptions[2673], "Mongoose")
	}
	if descriptions[3003] != "+34 Attack Power and +16 Hit Rating" {
		t.Fatalf("descriptions[3003] = %q, want %q", descriptions[3003], "+34 Attack Power and +16 Hit Rating")
	}
}
