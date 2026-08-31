package upgrades

import (
	"bytes"
	"sort"
	"slices"
	"testing"

	"github.com/wowsims/tbc/sim/core/proto"
)

func fixtureCatalog() *Catalog {
	db := &proto.UIDatabase{
		Items: []*proto.UIItem{
			{
				Id:    1001,
				Name:  "Raid Robe",
				Type:  proto.ItemType_ItemTypeChest,
				Phase: 1,
				Sources: []*proto.UIItemSource{
					{
						Source: &proto.UIItemSource_Drop{
							Drop: &proto.DropSource{
								Difficulty: proto.DungeonDifficulty_DifficultyRaid25,
								ZoneId:     1,
								NpcId:      1,
							},
						},
					},
				},
			},
			{
				Id:      1002,
				Name:    "Unknown Item",
				Type:    proto.ItemType_ItemTypeChest,
				Phase:   1,
				Sources: []*proto.UIItemSource{}, // No sources
			},
			{
				Id:                 1003,
				Name:               "LW Belt",
				Type:               proto.ItemType_ItemTypeWaist,
				Phase:              1,
				RequiredProfession: proto.Profession_Leatherworking, // Player is Tailor/Enchanter
				Sources: []*proto.UIItemSource{
					{
						Source: &proto.UIItemSource_Crafted{
							Crafted: &proto.CraftedSource{
								Profession: proto.Profession_Leatherworking,
							},
						},
					},
				},
			},
			{
				Id:             1004,
				Name:           "Rogue Boots",
				Type:           proto.ItemType_ItemTypeFeet,
				Phase:          1,
				ClassAllowlist: []proto.Class{proto.Class_ClassRogue}, // Player is Mage
				Sources: []*proto.UIItemSource{
					{
						Source: &proto.UIItemSource_Drop{
							Drop: &proto.DropSource{Difficulty: proto.DungeonDifficulty_DifficultyRaid25},
						},
					},
				},
			},
			{
				Id:      1005,
				Name:    "Unique Equipped Item",
				Type:    proto.ItemType_ItemTypeChest,
				Phase:   1,
				Unique:  true,
				Sources: []*proto.UIItemSource{
					{
						Source: &proto.UIItemSource_Drop{
							Drop: &proto.DropSource{Difficulty: proto.DungeonDifficulty_DifficultyRaid25},
						},
					},
				},
			},
			{
				Id:                 1006,
				Name:               "Alliance Staff",
				Type:               proto.ItemType_ItemTypeWeapon,
				WeaponType:         proto.WeaponType_WeaponTypeStaff,
				HandType:           proto.HandType_HandTypeTwoHand,
				Phase:              1,
				FactionRestriction: proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY, // Player is Troll (Horde)
				Sources: []*proto.UIItemSource{
					{
						Source: &proto.UIItemSource_Drop{
							Drop: &proto.DropSource{Difficulty: proto.DungeonDifficulty_DifficultyRaid25},
						},
					},
				},
			},
			{
				Id:    1007,
				Name:  "Test Ring",
				Type:  proto.ItemType_ItemTypeFinger,
				Phase: 1,
				Sources: []*proto.UIItemSource{
					{
						Source: &proto.UIItemSource_Drop{
							Drop: &proto.DropSource{Difficulty: proto.DungeonDifficulty_DifficultyRaid25},
						},
					},
				},
			},
			{
				Id:         1008,
				Name:       "Two-Hand Staff",
				Type:       proto.ItemType_ItemTypeWeapon,
				WeaponType: proto.WeaponType_WeaponTypeStaff,
				HandType:   proto.HandType_HandTypeTwoHand,
				Phase:      1,
				Sources: []*proto.UIItemSource{
					{
						Source: &proto.UIItemSource_Drop{
							Drop: &proto.DropSource{Difficulty: proto.DungeonDifficulty_DifficultyRaid25},
						},
					},
				},
			},
			{
				Id:         1009,
				Name:       "Test Off-Hand",
				Type:       proto.ItemType_ItemTypeWeapon,
				WeaponType: proto.WeaponType_WeaponTypeOffHand,
				HandType:   proto.HandType_HandTypeOffHand,
				Phase:      1,
				Sources: []*proto.UIItemSource{
					{
						Source: &proto.UIItemSource_Drop{
							Drop: &proto.DropSource{Difficulty: proto.DungeonDifficulty_DifficultyRaid25},
						},
					},
				},
			},
		},
		Zones: []*proto.UIZone{
			{Id: 1, Name: "Test Raid"},
		},
		Npcs: []*proto.UINPC{
			{Id: 1, Name: "Test Boss", ZoneId: 1},
		},
	}

	return NewCatalog(db)
}

func fixtureImported(t *testing.T) *ImportedSettings {
	t.Helper()
	imported := mustImportFixture(t)
	// Equip unique item 1005 in chest to test unique constraint
	imported.Settings.Player.Equipment.Items[int(proto.ItemSlot_ItemSlotChest)].Id = 1005
	return imported
}

func mustBuild(t *testing.T, filters ContentFilters) BuildResult {
	t.Helper()
	imported := fixtureImported(t)
	return mustBuildFrom(t, imported, filters)
}

func mustBuildFrom(t *testing.T, imported *ImportedSettings, filters ContentFilters) BuildResult {
	t.Helper()
	catalog := fixtureCatalog()
	result, err := BuildCandidates(imported, filters, ItemPolicy{}, catalog)
	if err != nil {
		t.Fatalf("BuildCandidates failed: %v", err)
	}
	return result
}

func containsID(candidates []Candidate, id int32) bool {
	for _, c := range candidates {
		if c.Item.ID == id {
			return true
		}
	}
	return false
}

func targetSlotsFor(candidates []Candidate, id int32) []proto.ItemSlot {
	var slots []proto.ItemSlot
	for _, c := range candidates {
		if c.Item.ID == id {
			slots = append(slots, c.TargetSlot)
		}
	}
	return slots
}

func containsSlot(displaced []UIItemSummary, slot proto.ItemSlot) bool {
	for _, d := range displaced {
		if d.Slot == slot {
			return true
		}
	}
	return false
}

func candidateByIDAndSlot(t *testing.T, candidates []Candidate, id int32, slot proto.ItemSlot) Candidate {
	t.Helper()
	for _, c := range candidates {
		if c.Item.ID == id && c.TargetSlot == slot {
			return c
		}
	}
	t.Fatalf("candidate %d at slot %v not found", id, slot)
	return Candidate{}
}

func TestBuildCandidatesExcludesUnknownSourceByDefault(t *testing.T) {
	result := mustBuild(t, ContentFilters{})
	if result.Excluded.UnknownSource != 1 || containsID(result.Candidates, 1002) {
		t.Fatalf("unknown source handling = %#v", result)
	}
}

func TestBuildCandidatesPreservesBaselineBytes(t *testing.T) {
	imported := fixtureImported(t)
	before := mustMarshal(t, imported.Settings)
	_ = mustBuildFrom(t, imported, ContentFilters{})
	if after := mustMarshal(t, imported.Settings); !bytes.Equal(before, after) {
		t.Fatal("BuildCandidates mutated baseline")
	}
}

func TestBuildCandidatesRejectsClassProfessionUniqueAndFaction(t *testing.T) {
	result := mustBuild(t, ContentFilters{IncludeUnknown: true})
	for _, id := range []int32{1003, 1004, 1005, 1006} {
		if containsID(result.Candidates, id) {
			t.Fatalf("ineligible item %d was proposed", id)
		}
	}
}

func TestBuildCandidatesEvaluatesBothRingSlots(t *testing.T) {
	if got := targetSlotsFor(mustBuild(t, ContentFilters{}).Candidates, 1007); !slices.Equal(got, []proto.ItemSlot{proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2}) {
		t.Fatalf("ring slots = %v", got)
	}
}

func TestBuildCandidatesModelsTwoHandOffHandConflict(t *testing.T) {
	candidate := candidateByIDAndSlot(t, mustBuild(t, ContentFilters{}).Candidates, 1008, proto.ItemSlot_ItemSlotMainHand)
	if !containsSlot(candidate.Displaced, proto.ItemSlot_ItemSlotOffHand) {
		t.Fatal("two-hand candidate retained off-hand")
	}
}

func candidateIDs(candidates []Candidate) []int32 {
	seen := make(map[int32]bool)
	var ids []int32
	for _, c := range candidates {
		if !seen[c.Item.ID] {
			seen[c.Item.ID] = true
			ids = append(ids, c.Item.ID)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func TestBuildCandidatesFiltersExpectedIDs(t *testing.T) {
	// Filter permits only the fixture raid source kind.
	result := mustBuild(t, ContentFilters{SourceKinds: []proto.SourceFilterOption{proto.SourceFilterOption_SourceRaid}})
	if got := candidateIDs(result.Candidates); !slices.Equal(got, []int32{1001, 1007, 1008, 1009}) {
		t.Fatalf("candidate IDs = %v, want [1001 1007 1008 1009]", got)
	}
}
