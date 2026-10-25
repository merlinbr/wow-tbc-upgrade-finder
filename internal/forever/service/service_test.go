package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/wowsims/tbc/internal/forever/catalog"
	"github.com/wowsims/tbc/internal/forever/domain"
	"github.com/wowsims/tbc/internal/forever/store"
	"github.com/wowsims/tbc/internal/forever/validate"
)

var ctx = context.Background()

const (
	plateHeadID    = 23019
	leatherHandsID = 720
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	st, err := store.NewSQLite(":memory:")
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	return New(st, cat)
}

func mustCreateCharacterSvc(t *testing.T, svc *Service, class, spec, race string, level int) domain.Character {
	t.Helper()
	ch, err := svc.CreateCharacter(ctx, CreateCharacterInput{
		Name:  "Elara",
		Class: class,
		Spec:  spec,
		Race:  race,
		Level: level,
	})
	if err != nil {
		t.Fatalf("create character: %v", err)
	}
	return ch
}

func mustCreateBuildSvc(t *testing.T, svc *Service, characterID, kind string) domain.Build {
	t.Helper()
	if kind == "current" {
		builds, err := svc.ListBuilds(ctx, characterID)
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
	build, err := svc.CreateBuild(ctx, characterID, CreateBuildInput{Name: "Pre-raid"})
	if err != nil {
		t.Fatalf("create build: %v", err)
	}
	return build
}

func hasCode(violations []validate.Violation, code string) bool {
	for _, violation := range violations {
		if violation.Code == code {
			return true
		}
	}
	return false
}

func TestServiceRejectsInvalidAndPreservesPriorBuild(t *testing.T) {
	svc := newTestService(t)
	ch := mustCreateCharacterSvc(t, svc, "mage", "mage_frost", "gnome", 60)
	current := mustCreateBuildSvc(t, svc, ch.ID, "current")
	before := current.Equipment

	_, violations, err := svc.UpdateBuild(ctx, current.ID, UpdateBuildInput{
		Equipment: &domain.Equipment{domain.SlotHead: {ItemID: plateHeadID}},
		Version:   current.Version,
	})
	if err == nil || !hasCode(violations, "incompatible_armor") {
		t.Fatalf("expected validation failure, err=%v violations=%v", err, violations)
	}

	after, err := svc.GetBuild(ctx, current.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after.Equipment) {
		t.Fatal("failed save mutated build")
	}
}

func TestServiceCopyIsIndependent(t *testing.T) {
	svc := newTestService(t)
	ch := mustCreateCharacterSvc(t, svc, "paladin", "paladin_retribution", "human", 60)
	current := mustCreateBuildSvc(t, svc, ch.ID, "current")

	plan, err := svc.CopyBuild(ctx, current.ID)
	if err != nil {
		t.Fatalf("copy build: %v", err)
	}
	saved, violations, err := svc.UpdateBuild(ctx, plan.ID, UpdateBuildInput{
		Equipment: &domain.Equipment{domain.SlotHead: {ItemID: plateHeadID}},
		Version:   plan.Version,
	})
	if err != nil {
		t.Fatalf("save plan: %v (violations %v)", err, violations)
	}
	if saved.Equipment[domain.SlotHead].ItemID != plateHeadID {
		t.Fatalf("saved equipment = %+v", saved.Equipment)
	}

	reloadedCurrent, err := svc.GetBuild(ctx, current.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloadedCurrent.Equipment) != 0 {
		t.Fatal("copy mutated current")
	}
}

func TestServiceCreateCharacterValidation(t *testing.T) {
	svc := newTestService(t)

	if _, err := svc.CreateCharacter(ctx, CreateCharacterInput{
		Name: "", Class: "paladin", Spec: "paladin_retribution", Race: "human", Level: 60,
	}); err == nil {
		t.Fatal("empty name accepted")
	}
	if _, err := svc.CreateCharacter(ctx, CreateCharacterInput{
		Name: "Orcadin", Class: "paladin", Spec: "paladin_retribution", Race: "orc", Level: 60,
	}); err == nil {
		t.Fatal("orc paladin accepted")
	}
	if _, err := svc.CreateCharacter(ctx, CreateCharacterInput{
		Name: "Elara", Class: "paladin", Spec: "mage_frost", Race: "human", Level: 60,
	}); err == nil {
		t.Fatal("foreign spec accepted")
	}
	if _, err := svc.CreateCharacter(ctx, CreateCharacterInput{
		Name: "Elara", Class: "paladin", Spec: "paladin_retribution", Race: "human", Level: 61,
	}); err == nil {
		t.Fatal("level 61 accepted")
	}
}
