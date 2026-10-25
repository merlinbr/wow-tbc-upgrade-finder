package catalog

import (
	"slices"
	"testing"

	"github.com/wowsims/tbc/internal/forever/domain"
)

const plateHeadID = 23019 // Icebane Helmet, plate head, no class allowlist.

func loadTestCatalog(t *testing.T) *Catalog {
	t.Helper()
	cat, err := Load()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	return cat
}

func paladin() domain.Character {
	return domain.Character{
		Class: "paladin",
		Spec:  "paladin_retribution",
		Race:  "human",
		Level: 60,
	}
}

func mage() domain.Character {
	return domain.Character{
		Class: "mage",
		Spec:  "mage_frost",
		Race:  "gnome",
		Level: 60,
	}
}

func TestSearchItemsFiltersAndPaginates(t *testing.T) {
	cat := loadTestCatalog(t)

	got, total := SearchItems(cat, ItemFilter{Class: "paladin", Slot: "head", Limit: 1})
	if total < 2 {
		t.Fatalf("total = %d", total)
	}
	if len(got) != 1 {
		t.Fatalf("page = %d", len(got))
	}
	for _, item := range got {
		if !slices.Contains(item.Slots, domain.SlotHead) {
			t.Fatal("slot leak")
		}
	}
	if _, n := SearchItems(cat, ItemFilter{Query: "brawler gloves"}); n != 1 {
		t.Fatalf("q total = %d", n)
	}

	plateHead := cat.Items[plateHeadID]
	if plateHead == nil {
		t.Fatalf("item %d missing", plateHeadID)
	}
	if ok, reason := cat.Compatibility(paladin(), plateHead, domain.SlotHead); !ok {
		t.Fatalf("paladin plate head should fit: %s", reason)
	}
	if ok, reason := cat.Compatibility(mage(), plateHead, domain.SlotHead); ok || reason != "incompatible_armor" {
		t.Fatalf("mage plate: %v %q", ok, reason)
	}
}
