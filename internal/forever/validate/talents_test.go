package validate

import (
	"fmt"
	"slices"
	"testing"

	"github.com/wowsims/tbc/internal/forever/domain"
)

func paladinTrees(t *testing.T) *domain.TalentClass {
	t.Helper()
	talents := make([]domain.Talent, 0, 13)
	for index := 1; index <= 10; index++ {
		talents = append(talents, domain.Talent{
			ID:        fmt.Sprintf("def:%d", index),
			Name:      fmt.Sprintf("Talent %d", index),
			MaxPoints: 5,
			Row:       0,
			Col:       index - 1,
		})
	}
	talents = append(talents,
		domain.Talent{ID: "def:row1", Name: "Row One", MaxPoints: 5, Row: 1, Col: 0},
		domain.Talent{ID: "def:prereq", Name: "Prereq", MaxPoints: 5, Row: 0, Col: 10},
		domain.Talent{
			ID: "def:child", Name: "Child", MaxPoints: 5, Row: 1, Col: 1,
			PrereqID: "def:prereq",
		},
	)
	return &domain.TalentClass{
		Class: "paladin",
		Trees: []domain.TalentTree{{Name: "Holy", Talents: talents}},
	}
}

func level60Paladin() domain.Character {
	return domain.Character{Class: "paladin", Spec: "paladin_retribution", Race: "human", Level: 60}
}

func level10Paladin() domain.Character {
	return domain.Character{Class: "paladin", Spec: "paladin_retribution", Race: "human", Level: 10}
}

func TestTalentBudgetAndTier(t *testing.T) {
	trees := paladinTrees(t)
	alloc := domain.Talents{
		"def:1": 5, "def:2": 5, "def:3": 5, "def:4": 5, "def:5": 5,
		"def:6": 5, "def:7": 5, "def:8": 5, "def:9": 5, "def:10": 5,
	}
	_, issues := Talents(level60Paladin(), alloc, trees)
	if issues.TotalPoints != 50 {
		t.Fatalf("total = %d", issues.TotalPoints)
	}
	over := domain.Talents{"def:1": 51}
	if violations, _ := Talents(level10Paladin(), over, trees); !hasCode(violations, "talent_budget_exceeded") {
		t.Fatalf("budget: %v", codes(violations))
	}
	if violations, _ := Talents(level60Paladin(), alloc, trees); len(violations) != 0 {
		t.Fatalf("unexpected %v", codes(violations))
	}
}

func TestTierGateAndPrerequisiteAndBroken(t *testing.T) {
	trees := paladinTrees(t)

	violations, _ := Talents(level60Paladin(), domain.Talents{"def:row1": 1}, trees)
	if !hasCode(violations, "talent_tier_locked") {
		t.Fatalf("tier: %v", codes(violations))
	}

	violations, issues := Talents(
		level60Paladin(), domain.Talents{"def:prereq": 1, "def:child": 1}, trees,
	)
	if !hasCode(violations, "talent_prereq_unmet") {
		t.Fatalf("prereq: %v", codes(violations))
	}
	if !slices.Contains(issues.Broken, "def:child") {
		t.Fatalf("broken = %v", issues.Broken)
	}

	if _, violations := TrySetPoints(trees, domain.Talents{"def:child": 1}, "def:child", 2); len(violations) == 0 {
		t.Fatal("addition over broken talent must be rejected")
	}
	if next, _ := TrySetPoints(trees, domain.Talents{"def:child": 1}, "def:child", 0); next["def:child"] != 0 {
		t.Fatal("removal must be allowed")
	}
}

func TestTalentMalformedInput(t *testing.T) {
	trees := paladinTrees(t)

	violations, issues := Talents(level60Paladin(), domain.Talents{"def:missing": 1}, trees)
	if !hasCode(violations, "unknown_talent") {
		t.Fatalf("unknown: %v", codes(violations))
	}
	if !slices.Contains(issues.Broken, "def:missing") {
		t.Fatalf("broken = %v", issues.Broken)
	}

	violations, _ = Talents(level60Paladin(), domain.Talents{"def:1": -1}, trees)
	if !hasCode(violations, "talent_rank_exceeded") {
		t.Fatalf("negative: %v", codes(violations))
	}
	violations, _ = Talents(level60Paladin(), domain.Talents{"def:1": 6}, trees)
	if !hasCode(violations, "talent_rank_exceeded") {
		t.Fatalf("over max: %v", codes(violations))
	}
}
