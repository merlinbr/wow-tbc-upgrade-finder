package store

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/wowsims/tbc/internal/forever/domain"
)

var ctx = context.Background()

func newTestStore(t *testing.T) *SQLite {
	t.Helper()
	st, err := NewSQLite(":memory:")
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func mustCreateCharacter(t *testing.T, st *SQLite) domain.Character {
	t.Helper()
	ch := domain.Character{
		ID:    uuid.NewString(),
		Name:  "Elara",
		Class: "paladin",
		Spec:  "paladin_retribution",
		Race:  "human",
		Level: 60,
	}
	current := domain.Build{
		ID:              uuid.NewString(),
		CharacterID:     ch.ID,
		Kind:            "current",
		CatalogRevision: "test-revision",
		Source:          "manual",
	}
	if err := st.CreateCharacterWithCurrent(ctx, ch, current); err != nil {
		t.Fatalf("create character: %v", err)
	}
	got, err := st.GetCharacter(ctx, ch.ID)
	if err != nil {
		t.Fatalf("get character: %v", err)
	}
	return got
}

func mustCreateBuild(t *testing.T, st *SQLite, characterID, kind, name string) domain.Build {
	t.Helper()
	if kind == "current" {
		builds, err := st.ListBuilds(ctx, characterID)
		if err != nil {
			t.Fatalf("list builds: %v", err)
		}
		for _, build := range builds {
			if build.Kind == "current" {
				return build
			}
		}
		t.Fatal("current build missing")
	}
	build := domain.Build{
		ID:              uuid.NewString(),
		CharacterID:     characterID,
		Kind:            kind,
		Name:            name,
		CatalogRevision: "test-revision",
		Source:          "manual",
	}
	if err := st.CreateBuild(ctx, build); err != nil {
		t.Fatalf("create build: %v", err)
	}
	got, err := st.GetBuild(ctx, build.ID)
	if err != nil {
		t.Fatalf("get build: %v", err)
	}
	return got
}

func mustCopyBuild(t *testing.T, st *SQLite, sourceID, name string) domain.Build {
	t.Helper()
	source, err := st.GetBuild(ctx, sourceID)
	if err != nil {
		t.Fatalf("get source build: %v", err)
	}
	source.ID = uuid.NewString()
	source.Kind = "plan"
	source.Name = name
	source.Version = 1
	if err := st.CreateBuild(ctx, source); err != nil {
		t.Fatalf("create copied build: %v", err)
	}
	got, err := st.GetBuild(ctx, source.ID)
	if err != nil {
		t.Fatalf("get copied build: %v", err)
	}
	return got
}

func TestBuildIsolationAndVersioning(t *testing.T) {
	st := newTestStore(t)
	ch := mustCreateCharacter(t, st)
	current := mustCreateBuild(t, st, ch.ID, "current", "")
	plan := mustCopyBuild(t, st, current.ID, "Pre-raid")

	updated := plan
	updated.Equipment = domain.Equipment{domain.SlotHead: {ItemID: 720}}
	if _, err := st.UpdateBuild(ctx, updated, plan.Version); err != nil {
		t.Fatal(err)
	}

	got, err := st.GetBuild(ctx, current.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Equipment) != 0 {
		t.Fatal("current mutated")
	}
	if _, err := st.UpdateBuild(ctx, updated, plan.Version); !errors.Is(err, ErrVersionConflict) {
		t.Fatalf("err = %v", err)
	}
}
