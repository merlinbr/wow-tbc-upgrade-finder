package validate

import (
	"errors"
	"testing"

	"github.com/wowsims/tbc/internal/forever/catalog"
	"github.com/wowsims/tbc/internal/forever/domain"
)

const (
	helmID            = 23019
	weaponOnlyEffect  = 241
	ambiguousEffect   = 1897
	ambiguousItemID   = 273489
	ambiguousSpellID  = 13529
	neckEnchantEffect = 8219
)

func TestEnchantResolution(t *testing.T) {
	cat := testCatalog(t)
	helm := cat.Items[helmID]
	if helm == nil {
		t.Fatalf("item %d missing", helmID)
	}

	if e, err := catalog.ResolveEnchant(cat, neckEnchantEffect, domain.SlotNeck); err != nil || e.SpellID != 1249058 {
		t.Fatalf("neck: %v %v", e, err)
	}
	if _, err := catalog.ResolveEnchant(cat, neckEnchantEffect, domain.SlotHead); !errors.Is(err, catalog.ErrAmbiguous) && !errors.Is(err, catalog.ErrNotFound) {
		t.Fatalf("head: %v", err)
	}

	if got := Enchant(paladin(), domain.SlotHead, helm, &domain.EnchantInstance{EffectID: 999999}, cat); !hasCode(got, "unknown_enchant") {
		t.Fatalf("unknown: %v", codes(got))
	}
	if got := Enchant(paladin(), domain.SlotHead, helm, &domain.EnchantInstance{EffectID: weaponOnlyEffect}, cat); !hasCode(got, "incompatible_enchant_slot") {
		t.Fatalf("slot: %v", codes(got))
	}
	if got := Enchant(paladin(), domain.SlotHead, nil, &domain.EnchantInstance{EffectID: neckEnchantEffect}, cat); !hasCode(got, "enchant_without_item") {
		t.Fatalf("empty slot: %v", codes(got))
	}
}

func TestEnchantAmbiguityByName(t *testing.T) {
	cat := testCatalog(t)

	if _, err := catalog.ResolveEnchant(cat, ambiguousEffect, domain.SlotMainHand); !errors.Is(err, catalog.ErrAmbiguous) {
		t.Fatalf("ambiguous: %v", err)
	}
	resolved, err := catalog.ResolveEnchant(
		cat, ambiguousEffect, domain.SlotMainHand,
		&domain.EnchantInstance{EffectID: ambiguousEffect, ItemID: ambiguousItemID},
	)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.ItemID != ambiguousItemID || resolved.SpellID != ambiguousSpellID {
		t.Fatalf("resolved = %+v", resolved)
	}

	sword := cat.Items[oneHandSwordID]
	if sword == nil {
		t.Fatalf("item %d missing", oneHandSwordID)
	}
	if got := Enchant(warrior(), domain.SlotMainHand, sword, &domain.EnchantInstance{EffectID: ambiguousEffect}, cat); !hasCode(got, "ambiguous_enchant") {
		t.Fatalf("ambiguous enchant: %v", codes(got))
	}
}
