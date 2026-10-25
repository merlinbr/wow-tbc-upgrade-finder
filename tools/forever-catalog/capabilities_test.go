package main

import (
	"slices"
	"strings"
	"testing"

	"github.com/wowsims/tbc/internal/forever/domain"
)

func capabilitySpecs() map[domain.Class][]domain.SpecInfo {
	return map[domain.Class][]domain.SpecInfo{
		"paladin": {
			{Key: "paladin_holy", Name: "Holy", Class: "paladin"},
			{Key: "paladin_protection", Name: "Protection", Class: "paladin"},
			{Key: "paladin_retribution", Name: "Retribution", Class: "paladin"},
		},
		"warrior": {
			{Key: "warrior_arms", Name: "Arms", Class: "warrior"},
			{Key: "warrior_fury", Name: "Fury", Class: "warrior"},
			{Key: "warrior_protection", Name: "Protection", Class: "warrior"},
		},
		"mage": {
			{Key: "mage_arcane", Name: "Arcane", Class: "mage"},
		},
	}
}

func TestParseCapabilities(t *testing.T) {
	capsTS := string(mustRead(t, "testdata/capabilities.txt"))
	caps, err := ParseCapabilities(capsTS, testEnums(t), capabilitySpecs())
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Contains(caps.Armor["paladin"], "plate") {
		t.Fatal("paladin plate")
	}
	if !slices.Contains(caps.Armor["mage"], "cloth") {
		t.Fatal("mage cloth")
	}
	if !slices.Contains(caps.Weapons["paladin"], domain.WeaponEligibility{WeaponType: "sword", CanUseTwoHand: true}) {
		t.Fatalf("paladin sword: %+v", caps.Weapons["paladin"])
	}
	if !slices.Contains(caps.Ranged["mage"], "wand") {
		t.Fatalf("mage wand: %+v", caps.Ranged["mage"])
	}
	if !slices.Contains(caps.Ranged["warrior"], "bow") {
		t.Fatalf("warrior bow: %+v", caps.Ranged["warrior"])
	}
	if !slices.Contains(caps.Races["paladin"], "human") {
		t.Fatalf("paladin human: %+v", caps.Races["paladin"])
	}
	if caps.DualWield["paladin_retribution"] {
		t.Fatal("paladin dw")
	}
	if !caps.DualWield["warrior_fury"] {
		t.Fatal("warrior dw")
	}
	if len(caps.Classes) != 3 {
		t.Fatalf("classes = %+v", caps.Classes)
	}
	if caps.Classes[0].Class != "mage" || caps.Classes[0].Name != "Mage" {
		t.Fatalf("class order = %+v", caps.Classes)
	}
	if len(caps.Specs["paladin"]) != 3 {
		t.Fatalf("specs = %+v", caps.Specs["paladin"])
	}

	// Unknown members must fail the generator rather than silently disappear.
	broken := strings.Replace(capsTS, "ArmorType.ArmorTypePlate", "ArmorType.ArmorTypeSpandex", 1)
	if _, err := ParseCapabilities(broken, testEnums(t), capabilitySpecs()); err == nil {
		t.Fatal("unknown armor member accepted")
	}
	if _, err := ParseCapabilities("export const nothing = 1;", testEnums(t), capabilitySpecs()); err == nil {
		t.Fatal("missing blocks accepted")
	}
}
