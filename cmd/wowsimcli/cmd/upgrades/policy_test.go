package upgrades

import (
	"bytes"
	"slices"
	"testing"

	"github.com/wowsims/tbc/sim/core/proto"
)

func fixturePolicyCatalog() *Catalog {
	cat := fixtureCatalog()

	// Socketed Head Item (Red, Blue, Yellow sockets)
	cat.Items[1010] = &proto.UIItem{
		Id:         1010,
		Name:       "Socketed Helm",
		Type:       proto.ItemType_ItemTypeHead,
		Phase:      1,
		GemSockets: []proto.GemColor{proto.GemColor_GemColorRed, proto.GemColor_GemColorBlue, proto.GemColor_GemColorYellow},
		Sources: []*proto.UIItemSource{
			{
				Source: &proto.UIItemSource_Drop{
					Drop: &proto.DropSource{Difficulty: proto.DungeonDifficulty_DifficultyRaid25},
				},
			},
		},
	}

	// Gems
	cat.Gems[2001] = &proto.UIGem{
		Id:      2001,
		Name:    "Bold Crimson Spinel",
		Color:   proto.GemColor_GemColorRed,
		Quality: proto.ItemQuality_ItemQualityEpic,
		Phase:   1,
	}
	cat.Gems[2002] = &proto.UIGem{
		Id:                 2002,
		Name:               "JC Only Gem",
		Color:              proto.GemColor_GemColorRed,
		Quality:            proto.ItemQuality_ItemQualityLegendary, // Over quality and requires JC
		RequiredProfession: proto.Profession_Jewelcrafting,
		Phase:              1,
	}
	cat.Gems[2003] = &proto.UIGem{
		Id:      2003,
		Name:    "Solid Empyrean Sapphire",
		Color:   proto.GemColor_GemColorBlue,
		Quality: proto.ItemQuality_ItemQualityRare,
		Phase:   1,
	}
	cat.Gems[2004] = &proto.UIGem{
		Id:      2004,
		Name:    "Chaotic Skyfire Diamond",
		Color:   proto.GemColor_GemColorMeta,
		Quality: proto.ItemQuality_ItemQualityRare,
		Phase:   1,
	}

	// Enchants
	cat.Enchants[3001] = &proto.UIEnchant{
		EffectId: 3001,
		Name:     "Glyph of Power",
		Type:     proto.ItemType_ItemTypeHead,
		Phase:    1,
	}
	cat.Enchants[3002] = &proto.UIEnchant{
		EffectId:       3002,
		Name:           "Rogue Head Enchant",
		Type:           proto.ItemType_ItemTypeHead,
		ClassAllowlist: []proto.Class{proto.Class_ClassRogue}, // Ineligible for Mage
		Phase:          1,
	}

	return cat
}

func fixtureSocketedCandidate() Candidate {
	imported := mustImportFixture(nil)
	cat := fixturePolicyCatalog()
	item := cat.Items[1010]

	clonedSettings := cloneMessage(imported.Settings)
	clonedEquip := cloneMessage(clonedSettings.GetPlayer().GetEquipment())
	clonedEquip.Items[int(proto.ItemSlot_ItemSlotHead)] = &proto.ItemSpec{Id: 1010}
	clonedSettings.GetPlayer().Equipment = clonedEquip

	req := (&ImportedSettings{Settings: clonedSettings}).NewRequest(0)

	return Candidate{
		Item:       cat.ItemSummary(item, proto.ItemSlot_ItemSlotHead),
		TargetSlot: proto.ItemSlot_ItemSlotHead,
		Request:    req,
		Source:     cat.ResolveSource(item),
	}
}

func fixturePolicy() ItemPolicy {
	return ItemPolicy{
		GemBySocket: map[proto.GemColor]int32{
			proto.GemColor_GemColorRed: 2001,
		},
		MaxGemQuality: proto.ItemQuality_ItemQualityEpic,
		EnchantByType: map[proto.ItemType]int32{
			proto.ItemType_ItemTypeHead: 3001,
		},
	}
}

func mustApplyPolicy(t *testing.T, candidate Candidate, policy ItemPolicy) Candidate {
	t.Helper()
	cat := fixturePolicyCatalog()
	applied, err := ApplyPolicy(candidate, policy, cat)
	if err != nil {
		t.Fatalf("ApplyPolicy failed: %v", err)
	}
	return applied
}

func TestApplyPolicyUsesOnlyLegalSocketChoices(t *testing.T) {
	got := mustApplyPolicy(t, fixtureSocketedCandidate(), fixturePolicy())
	if !slices.Equal(got.Applied.GemIDs, []int32{2001, 0, 0}) {
		t.Fatalf("gems = %v, want [2001 0 0]", got.Applied.GemIDs)
	}
	if got.Applied.EnchantID != 3001 {
		t.Fatalf("enchant ID = %d, want 3001", got.Applied.EnchantID)
	}
}

func TestApplyPolicyRejectsOverQualityAndProfessionGems(t *testing.T) {
	cat := fixturePolicyCatalog()
	_, err := ApplyPolicy(fixtureSocketedCandidate(), ItemPolicy{
		GemBySocket: map[proto.GemColor]int32{
			proto.GemColor_GemColorRed: 2002,
		},
		MaxGemQuality: proto.ItemQuality_ItemQualityEpic,
	}, cat)
	if err == nil {
		t.Fatal("over-quality or profession gem was accepted")
	}
}

func TestApplyPolicyRejectsIneligibleEnchant(t *testing.T) {
	cat := fixturePolicyCatalog()
	_, err := ApplyPolicy(fixtureSocketedCandidate(), ItemPolicy{
		EnchantByType: map[proto.ItemType]int32{
			proto.ItemType_ItemTypeHead: 3002,
		},
	}, cat)
	if err == nil {
		t.Fatal("ineligible enchant was accepted")
	}
}

func TestApplyPolicyDoesNotChangeBaselineEquipment(t *testing.T) {
	imported := fixtureImported(t)
	before := mustMarshal(t, imported.Settings)
	_ = mustApplyPolicy(t, fixtureSocketedCandidate(), fixturePolicy())
	if after := mustMarshal(t, imported.Settings); !bytes.Equal(before, after) {
		t.Fatal("policy mutated baseline")
	}
}
