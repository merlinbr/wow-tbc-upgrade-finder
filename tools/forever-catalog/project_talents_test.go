package main

import (
	"testing"

	"github.com/wowsims/tbc/internal/forever/domain"
)

func TestProjectTalents(t *testing.T) {
	rawTrees := []byte(`[
		{
			"name": "Holy",
			"backgroundUrl": "https://example.test/holy.jpg",
			"talents": [
				{
					"fieldName": "alpha",
					"fancyName": "Alpha",
					"location": { "rowIdx": 0, "colIdx": 0 },
					"spellId": 1,
					"maxPoints": 5
				},
				{
					"fieldName": "beta",
					"fancyName": "Beta",
					"location": { "rowIdx": 0, "colIdx": 1 },
					"spellId": 2,
					"definitionId": 135378,
					"maxPoints": 3
				},
				{
					"fieldName": "gamma",
					"fancyName": "Gamma",
					"location": { "rowIdx": 2, "colIdx": 0 },
					"spellId": 3,
					"definitionId": 135379,
					"maxPoints": 1,
					"prereqLocation": { "rowIdx": 0, "colIdx": 1 }
				}
			]
		}
	]`)
	raw := RawTalents{
		Classes:    map[domain.Class][]byte{"paladin": rawTrees},
		SpellIcons: map[int]string{2: "spell_holy_beta"},
	}

	trees, err := ProjectTalents(raw, testEnums(t))
	if err != nil {
		t.Fatal(err)
	}
	paladin := trees["paladin"]
	if paladin == nil || len(paladin.Trees) != 1 {
		t.Fatalf("paladin trees = %+v", paladin)
	}
	talents := paladin.Trees[0].Talents
	if len(talents) != 3 {
		t.Fatalf("talents = %d", len(talents))
	}

	tal := talents[1]
	if tal.ID != "def:135378" {
		t.Fatalf("id = %q", tal.ID)
	}
	if tal.MaxPoints != 3 {
		t.Fatalf("max = %d", tal.MaxPoints)
	}
	if tal.Name != "Beta" {
		t.Fatalf("name = %q", tal.Name)
	}
	if tal.Icon != "spell_holy_beta" {
		t.Fatalf("icon = %q", tal.Icon)
	}
	if talents[2].PrereqID != tal.ID {
		t.Fatalf("prereq = %q, want %q", talents[2].PrereqID, tal.ID)
	}
	if talents[0].ID != "loc:paladin:0:0,0" {
		t.Fatalf("fallback id = %q", talents[0].ID)
	}
	if talents[2].Row != 2 || talents[2].Col != 0 {
		t.Fatalf("location = %d,%d", talents[2].Row, talents[2].Col)
	}
	if paladin.Trees[0].Name != "Holy" || paladin.Trees[0].Background != "https://example.test/holy.jpg" {
		t.Fatalf("tree = %+v", paladin.Trees[0])
	}
}
