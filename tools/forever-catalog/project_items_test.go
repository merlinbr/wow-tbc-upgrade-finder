package main

import (
	"testing"

	"github.com/wowsims/tbc/internal/forever/domain"
)

func findItem(items []domain.Item, id int) domain.Item {
	for _, item := range items {
		if item.ID == id {
			return item
		}
	}
	return domain.Item{}
}

func statValue(stats []domain.Stat, name string) int {
	for _, stat := range stats {
		if stat.Name == name {
			return stat.Value
		}
	}
	return 0
}

func testEnums(t *testing.T) Enums {
	t.Helper()
	enums, err := ParseEnums(string(mustRead(t, "testdata/proto.txt")))
	if err != nil {
		t.Fatal(err)
	}
	return enums
}

func TestProjectItems(t *testing.T) {
	enums := testEnums(t)
	raw := RawPayload{
		DB:         mustRead(t, "testdata/db.json"),
		LeftoverDB: mustRead(t, "testdata/leftover_db.json"),
	}
	cat, err := ProjectItems(raw, enums)
	if err != nil {
		t.Fatal(err)
	}

	item := findItem(cat.Items, 720)
	if item.ID == 0 {
		t.Fatal("item 720 missing")
	}
	if item.Ilvl != 28 {
		t.Fatalf("ilvl = %d", item.Ilvl)
	}
	if got := statValue(item.Stats, "strength"); got != 8 {
		t.Fatalf("strength = %v", got)
	}
	if got := statValue(item.Stats, "stamina"); got != 7 {
		t.Fatalf("stamina = %v", got)
	}
	if got := statValue(item.Stats, "armor"); got != 62 {
		t.Fatalf("armor = %v", got)
	}
	if len(item.Slots) != 1 || item.Slots[0] != domain.SlotHands {
		t.Fatalf("slots = %v", item.Slots)
	}
	if len(item.Variants) != 1 || item.Variants[0].Ilvl != 34 {
		t.Fatalf("variants = %+v", item.Variants)
	}
	if got := statValue(item.Variants[0].Stats, "strength"); got != 10 {
		t.Fatalf("variant strength = %v", got)
	}

	if got := cat.Enchants[0].EffectID; got != 8219 {
		t.Fatalf("enchant effect = %d", got)
	}
	if got := cat.Enchants[0].SpellID; got != 1249058 {
		t.Fatalf("enchant spell = %d", got)
	}
	if got := cat.Suffixes[0].ID; got != 12718 {
		t.Fatal("suffix id")
	}
	if got := statValue(cat.Suffixes[0].Stats, "strength"); got != 9 {
		t.Fatalf("suffix strength = %v", got)
	}

	// The default variant is flattened onto the item; other variants are kept.
	flat := findItem(cat.Items, 721)
	if flat.Ilvl != 60 {
		t.Fatalf("flat ilvl = %d", flat.Ilvl)
	}
	if flat.RequiredProfession != "blacksmithing" {
		t.Fatalf("profession = %q", flat.RequiredProfession)
	}
	if len(flat.ClassAllowlist) != 1 || flat.ClassAllowlist[0] != "paladin" {
		t.Fatalf("class allowlist = %v", flat.ClassAllowlist)
	}
	if len(flat.Slots) != 1 || flat.Slots[0] != domain.SlotHead {
		t.Fatalf("flat slots = %v", flat.Slots)
	}

	// Shirts/tabbards (item type 0) are not 17-slot items.
	if got := findItem(cat.Items, 722); got.ID != 0 {
		t.Fatal("type-0 item was not dropped")
	}
	if len(cat.LeftoverItems) != 1 || cat.LeftoverItems[0].ID != 9001 {
		t.Fatalf("leftover items = %+v", cat.LeftoverItems)
	}
	if len(cat.LeftoverEnchants) != 1 || cat.LeftoverEnchants[0].EffectID != 15 {
		t.Fatalf("leftover enchants = %+v", cat.LeftoverEnchants)
	}
}
