package validate

import (
	"testing"

	"github.com/wowsims/tbc/internal/forever/catalog"
	"github.com/wowsims/tbc/internal/forever/domain"
)

const (
	brawlerGlovesID  = 720
	swordID          = 23467
	shieldID         = 18826
	uniqueTrinketID  = 23047
	limitedOneID     = 18706
	limitedTwoID     = 19024
	oneHandSwordID   = 12584
	twoHandSwordID   = 18876
	twoHandShieldID  = 18825
	paladinOnlyHands = 246055
	tailoringRobeID  = 14152
)

func testCatalog(t *testing.T) *catalog.Catalog {
	t.Helper()
	cat, err := catalog.Load()
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}
	return cat
}

func warrior() domain.Character {
	return domain.Character{
		Class: "warrior", Spec: "warrior_fury", Race: "human", Level: 60,
	}
}

func priest() domain.Character {
	return domain.Character{
		Class: "priest", Spec: "priest_shadow", Race: "human", Level: 60,
	}
}

func paladin() domain.Character {
	return domain.Character{
		Class: "paladin", Spec: "paladin_retribution", Race: "human", Level: 60,
	}
}

func mage() domain.Character {
	return domain.Character{
		Class: "mage", Spec: "mage_frost", Race: "gnome", Level: 60,
	}
}

func orcPaladin() domain.Character {
	return domain.Character{
		Class: "paladin", Spec: "paladin_retribution", Race: "orc", Level: 60,
	}
}

func gear(id int) domain.Equipment {
	return domain.Equipment{domain.SlotHead: {ItemID: id}}
}

func gearIn(slot domain.Slot, id int) domain.Equipment {
	return domain.Equipment{slot: {ItemID: id}}
}

func twoCopies(id int) domain.Equipment {
	return domain.Equipment{
		domain.SlotTrinket1: {ItemID: id},
		domain.SlotTrinket2: {ItemID: id},
	}
}

func twoLimited(ids []int) domain.Equipment {
	return domain.Equipment{
		domain.SlotTrinket1: {ItemID: ids[0]},
		domain.SlotTrinket2: {ItemID: ids[1]},
	}
}

func twoOneHands() domain.Equipment {
	return domain.Equipment{
		domain.SlotMainHand: {ItemID: oneHandSwordID},
		domain.SlotOffHand:  {ItemID: oneHandSwordID},
	}
}

func twoHandPlusOffHand() domain.Equipment {
	return domain.Equipment{
		domain.SlotMainHand: {ItemID: twoHandSwordID},
		domain.SlotOffHand:  {ItemID: twoHandShieldID},
	}
}

func hasCode(violations []Violation, code string) bool {
	for _, violation := range violations {
		if violation.Code == code {
			return true
		}
	}
	return false
}

func codes(violations []Violation) []string {
	out := make([]string, 0, len(violations))
	for _, violation := range violations {
		out = append(out, violation.Code)
	}
	return out
}

func TestEquipmentRules(t *testing.T) {
	cases := []struct {
		name string
		ch   domain.Character
		eq   domain.Equipment
		code string
	}{
		{"plate on mage", mage(), gear(brawlerGlovesID), "incompatible_armor"},
		{"sword on priest", priest(), gearIn(domain.SlotMainHand, swordID), "incompatible_weapon"},
		{"shield in main hand", warrior(), gearIn(domain.SlotMainHand, shieldID), "incompatible_slot"},
		{"second unique copy", warrior(), twoCopies(uniqueTrinketID), "unique_conflict"},
		{"two limit-category items", warrior(), twoLimited([]int{limitedOneID, limitedTwoID}), "limit_category_conflict"},
		{"race/class mismatch", orcPaladin(), domain.Equipment{}, "race_class_mismatch"},
		{"off-hand one-hander without dw", paladin(), twoOneHands(), "dual_wield_unavailable"},
		{"two-hander clears off hand", warrior(), twoHandPlusOffHand(), "two_hand_conflict"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Equipment(tc.ch, tc.eq, testCatalog(t)); !hasCode(got, tc.code) {
				t.Fatalf("want %s, got %v", tc.code, codes(got))
			}
		})
	}
}

func TestEquipmentSupplementaryRules(t *testing.T) {
	cat := testCatalog(t)

	if got := Equipment(warrior(), gearIn(domain.SlotHands, paladinOnlyHands), cat); !hasCode(got, "incompatible_class") {
		t.Fatalf("paladin-only on warrior: %v", codes(got))
	}
	if got := Equipment(mage(), gearIn(domain.SlotChest, tailoringRobeID), cat); !hasCode(got, "incompatible_profession") {
		t.Fatalf("tailoring robe on non-tailor: %v", codes(got))
	}
	if got := Equipment(warrior(), gearIn(domain.SlotHead, 999999999), cat); !hasCode(got, "unknown_item") {
		t.Fatalf("unknown item: %v", codes(got))
	}
}
