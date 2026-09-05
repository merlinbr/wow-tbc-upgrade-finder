package upgrades

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/wowsims/tbc/assets/database"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

func cloneImportedSettings(t *testing.T) (*ImportedSettings, *Catalog, []byte) {
	t.Helper()
	original := mustImportFixture(t)
	clone := *original
	clone.Settings = cloneMessage(original.Settings)
	before := mustMarshal(t, clone.Settings)
	return &clone, NewCatalog(database.Load()), before
}

func TestEnrichArmoryStatKeysPreserveAcronyms(t *testing.T) {
	if got, want := statKey(stats.MP5.StatName()), "mp5"; got != want {
		t.Fatalf("stat key = %q, want %q", got, want)
	}
}

func TestEnrichArmoryReturnsCanonicalSlotsAndResolvedMetadata(t *testing.T) {
	armory, err := EnrichArmory(mustImportFixture(t), NewCatalog(database.Load()))
	if err != nil {
		t.Fatal(err)
	}
	if len(armory.Gear) != 17 {
		t.Fatalf("gear slots = %d, want 17", len(armory.Gear))
	}
	if armory.Gear[0].Slot != proto.ItemSlot_ItemSlotHead || armory.Gear[6].Slot != proto.ItemSlot_ItemSlotMainHand || armory.Gear[7].Slot != proto.ItemSlot_ItemSlotOffHand || armory.Gear[8].Slot != proto.ItemSlot_ItemSlotHands {
		t.Fatalf("unexpected display order: %#v", armory.Gear)
	}
	if armory.Gear[0].ItemID != 24266 || armory.Gear[8].ItemID != 29078 || armory.Gear[16].ItemID != 28673 {
		t.Fatalf("unexpected resolved item order: head=%d hands=%d ranged=%d", armory.Gear[0].ItemID, armory.Gear[8].ItemID, armory.Gear[16].ItemID)
	}
	if armory.Gear[0].ItemID == 0 || armory.Gear[0].ItemName == "" || len(armory.Gear[0].Stats) == 0 {
		t.Fatalf("head metadata was not resolved: %#v", armory.Gear[0])
	}
	if armory.Gear[0].RandomSuffix != nil {
		t.Fatal("unexpected random suffix on fixture head")
	}
}

func TestEnrichArmoryUsesEngineSocketMatchingAndPreservesImport(t *testing.T) {
	imported, catalog, before := cloneImportedSettings(t)
	var target *proto.ItemSpec
	var targetGem *proto.UIGem
	for index, spec := range imported.Settings.Player.Equipment.Items {
		if spec == nil {
			continue
		}
		item := catalog.Items[spec.GetId()]
		if item == nil || len(item.GetGemSockets()) != 1 {
			continue
		}
		for _, gem := range catalog.DB.GetGems() {
			if (gem.GetColor() == proto.GemColor_GemColorPurple || gem.GetColor() == proto.GemColor_GemColorPrismatic) && core.ColorIntersects(item.GetGemSockets()[0], gem.GetColor()) {
				target = imported.Settings.Player.Equipment.Items[index]
				targetGem = gem
				break
			}
		}
		if target != nil {
			break
		}
	}
	if target == nil {
		t.Fatal("fixture database has no one-socket item compatible with a purple or prismatic gem")
	}
	target.Gems = []int32{targetGem.GetId()}
	before = mustMarshal(t, imported.Settings)
	armory, err := EnrichArmory(imported, catalog)
	if err != nil {
		t.Fatal(err)
	}
	var slot GearSlotData
	for _, candidate := range armory.Gear {
		if candidate.ItemID == target.GetId() {
			slot = candidate
			break
		}
	}
	if len(slot.Sockets) != 1 || slot.Sockets[0].Gem == nil {
		t.Fatalf("socket metadata = %#v", slot.Sockets)
	}
	wantActive := core.ColorIntersects(slot.Sockets[0].Color, targetGem.GetColor())
	if slot.SocketBonus.Active != wantActive {
		t.Fatalf("socket bonus active = %v, want %v", slot.SocketBonus.Active, wantActive)
	}
	if !bytes.Equal(before, mustMarshal(t, imported.Settings)) {
		t.Fatal("enrichment mutated imported settings")
	}
}

func TestEnrichArmoryScalesRandomSuffixAndMatchesEngineSnapshot(t *testing.T) {
	imported, catalog, before := cloneImportedSettings(t)
	target := imported.Settings.Player.Equipment.Items[0]
	if target == nil || target.GetId() == 0 {
		t.Fatal("fixture has no equipped item for suffix fixture")
	}
	targetItem := cloneMessage(catalog.Items[target.GetId()])
	if targetItem == nil {
		t.Fatalf("fixture item %d is missing from catalog", target.GetId())
	}
	var sourceSuffix *proto.ItemRandomSuffix
	for _, suffix := range catalog.DB.GetRandomSuffixes() {
		if suffix != nil && suffix.GetId() != 0 && len(suffix.GetStats()) != 0 {
			sourceSuffix = suffix
			break
		}
	}
	if sourceSuffix == nil {
		t.Fatal("database has no nonempty random suffix fixture")
	}
	targetSuffix := cloneMessage(sourceSuffix)
	targetItem.RandomSuffixOptions = []int32{targetSuffix.GetId()}
	catalog.Items[target.GetId()] = targetItem
	catalog.RandomSuffixes[targetSuffix.GetId()] = targetSuffix
	target.RandomSuffix = targetSuffix.GetId()
	before = mustMarshal(t, imported.Settings)
	armory, err := EnrichArmory(imported, catalog)
	if err != nil {
		t.Fatal(err)
	}
	var got *RandomSuffixData
	for _, slot := range armory.Gear {
		if slot.ItemID == target.GetId() {
			got = slot.RandomSuffix
			break
		}
	}
	if got == nil {
		t.Fatal("enriched suffix is nil")
	}
	want := statValuesMap(stats.FromProtoArray(targetSuffix.GetStats()).Multiply(float64(itemRandPropPoints(targetItem)) / 10000).Floor())
	if !reflect.DeepEqual(got.Stats, want) {
		t.Fatalf("suffix stats = %#v, want %#v", got.Stats, want)
	}
	if !bytes.Equal(before, mustMarshal(t, imported.Settings)) {
		t.Fatal("enrichment mutated imported settings")
	}
}

func TestPanelDebuffStatsMirrorsSite(t *testing.T) {
	debuffs := &proto.Debuffs{
		FaerieFire:                  proto.TristateEffect_TristateEffectImproved,
		ImprovedSealOfTheCrusader:   proto.TristateEffect_TristateEffectImproved,
		HuntersMark:                 proto.TristateEffect_TristateEffectImproved,
		ExposeWeaknessUptime:        0.9,
		ExposeWeaknessHunterAgility: 1150,
	}
	extra, pseudo := panelDebuffStats(debuffs)
	if got, want := extra[stats.AttackPower], 1150*0.25+110; got != want {
		t.Fatalf("attack power = %v, want %v", got, want)
	}
	if got, want := extra[stats.RangedAttackPower], 1150*0.25+440; got != want {
		t.Fatalf("ranged attack power = %v, want %v", got, want)
	}
	for _, entry := range []struct {
		pseudo proto.PseudoStat
		want   float64
	}{
		{proto.PseudoStat_PseudoStatMeleeCritPercent, 3},
		{proto.PseudoStat_PseudoStatRangedCritPercent, 3},
		{proto.PseudoStat_PseudoStatSpellCritPercent, 3},
		{proto.PseudoStat_PseudoStatMeleeHitPercent, 3},
		{proto.PseudoStat_PseudoStatRangedHitPercent, 3},
	} {
		if got := pseudo[entry.pseudo]; got != entry.want {
			t.Fatalf("pseudo %v = %v, want %v", entry.pseudo, got, entry.want)
		}
	}
}

func TestEnrichArmoryStatsMatchFullSettingsEngineSnapshot(t *testing.T) {
	imported := mustImportFixture(t)
	imported.Settings.Player.Buffs = cloneMessage(core.FullIndividualBuffs)
	imported.Settings.Player.Consumables = &proto.ConsumesSpec{ScrollInt: true}
	imported.Settings.Player.TalentsString = "032002-2302320032120231221-03"
	imported.Settings.Player.EnableItemSwap = true
	imported.Settings.Player.ItemSwap = &proto.ItemSwap{
		Items: cloneMessage(imported.Settings.Player.Equipment).GetItems(),
	}
	before := mustMarshal(t, imported.Settings)

	armory, err := EnrichArmory(imported, NewCatalog(database.Load()))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, mustMarshal(t, imported.Settings)) {
		t.Fatal("enrichment mutated imported settings")
	}

	expectedPlayer := cloneMessage(imported.Settings.Player)
	expectedPlayer.Database = buildSimDatabase()
	expectedResult := core.ComputeStats(&proto.ComputeStatsRequest{
		Raid: &proto.Raid{
			Parties: []*proto.Party{{Players: []*proto.Player{expectedPlayer}, Buffs: cloneOrEmpty(imported.Settings.PartyBuffs, &proto.PartyBuffs{})}},
			Buffs:   cloneOrEmpty(imported.Settings.RaidBuffs, &proto.RaidBuffs{}),
			Debuffs: cloneOrEmpty(imported.Settings.Debuffs, &proto.Debuffs{}),
		},
		Encounter: cloneOrEmpty(imported.Settings.Encounter, &proto.Encounter{}),
	})
	if expectedResult == nil || expectedResult.GetRaidStats() == nil || len(expectedResult.GetRaidStats().GetParties()) != 1 {
		t.Fatal("independent engine snapshot was malformed")
	}
	expectedParty := expectedResult.GetRaidStats().GetParties()[0]
	if expectedParty == nil || len(expectedParty.GetPlayers()) == 0 || expectedParty.GetPlayers()[0] == nil || expectedParty.GetPlayers()[0].GetFinalStats() == nil {
		t.Fatal("independent engine snapshot had no final stats")
	}
	gear := expectedParty.GetPlayers()[0].GetFinalStats()
	if want := statMap(gear.GetStats()); !reflect.DeepEqual(armory.Stats, want) {
		t.Fatalf("stats = %#v, want %#v", armory.Stats, want)
	}
	pseudo := gear.GetPseudoStats()
	for _, entry := range []struct {
		key   string
		index proto.PseudoStat
	}{
		{"melee_hit_percent", proto.PseudoStat_PseudoStatMeleeHitPercent},
		{"spell_hit_percent", proto.PseudoStat_PseudoStatSpellHitPercent},
		{"melee_crit_percent", proto.PseudoStat_PseudoStatMeleeCritPercent},
		{"spell_crit_percent", proto.PseudoStat_PseudoStatSpellCritPercent},
		{"ranged_hit_percent", proto.PseudoStat_PseudoStatRangedHitPercent},
		{"ranged_crit_percent", proto.PseudoStat_PseudoStatRangedCritPercent},
		{"block_percent", proto.PseudoStat_PseudoStatBlockPercent},
	} {
		if got, want := armory.DerivedStats[entry.key], pseudo[entry.index]; got != want {
			t.Fatalf("derived stat %s = %v, want %v", entry.key, got, want)
		}
	}
}
func TestEnrichArmoryExposesIlvlAndEnchantDescription(t *testing.T) {
	imported, err := Import(readFixture(t, "retribution_no_settings_link.txt"))
	if err != nil {
		t.Fatal(err)
	}
	armory, err := EnrichArmory(imported, NewCatalog(database.Load()))
	if err != nil {
		t.Fatal(err)
	}
	for _, slot := range armory.Gear {
		if slot.ItemID == 0 {
			continue
		}
		if slot.Ilvl <= 0 {
			t.Fatalf("slot %s item %d has ilvl %d, want > 0", slot.SlotName, slot.ItemID, slot.Ilvl)
		}
	}
	found := false
	for _, slot := range armory.Gear {
		if slot.Enchant == nil {
			continue
		}
		found = true
		if slot.Enchant.Description == "" {
			t.Fatalf("enchant %s has empty description", slot.Enchant.Name)
		}
	}
	if !found {
		t.Fatal("fixture has no enchanted slot to assert description")
	}
}
