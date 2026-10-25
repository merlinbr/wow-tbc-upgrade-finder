package catalog

import (
	"slices"
	"sort"
	"strings"

	"github.com/wowsims/tbc/internal/forever/domain"
)

// ItemFilter selects catalog items for the picker.
type ItemFilter struct {
	Class  domain.Class
	Slot   domain.Slot
	Query  string
	Source string
	Limit  int
	Offset int
}

// EnchantFilter selects catalog enchants for the picker.
type EnchantFilter struct {
	Class  domain.Class
	Slot   domain.Slot
	Query  string
	Limit  int
	Offset int
}

// ValidSlot reports whether slot is one of the seventeen canonical slots.
func ValidSlot(slot domain.Slot) bool {
	return slices.Contains(domain.SlotOrder, slot)
}

// SearchItems filters, sorts and paginates the main catalog. The total is
// returned before pagination.
func SearchItems(cat *Catalog, filter ItemFilter) ([]domain.Item, int) {
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	source := strings.ToLower(strings.TrimSpace(filter.Source))

	matched := make([]*domain.Item, 0, len(cat.ItemList))
	for _, item := range cat.ItemList {
		if filter.Class != "" && len(item.ClassAllowlist) > 0 &&
			!slices.Contains(item.ClassAllowlist, filter.Class) {
			continue
		}
		if filter.Slot != "" && !slices.Contains(item.Slots, filter.Slot) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(item.Name), query) {
			continue
		}
		if source != "" && !itemHasSource(item, source) {
			continue
		}
		matched = append(matched, item)
	}

	sort.SliceStable(matched, func(i, j int) bool {
		if matched[i].Ilvl != matched[j].Ilvl {
			return matched[i].Ilvl > matched[j].Ilvl
		}
		return matched[i].ID < matched[j].ID
	})

	page, _, _ := paginate(len(matched), filter.Limit, filter.Offset)
	items := make([]domain.Item, 0, len(page))
	for _, index := range page {
		items = append(items, *matched[index])
	}
	return items, len(matched)
}

// SearchEnchants filters, sorts and paginates enchants that admit the slot.
func SearchEnchants(cat *Catalog, filter EnchantFilter) []domain.Enchant {
	query := strings.ToLower(strings.TrimSpace(filter.Query))
	types := SlotItemTypes(filter.Slot)

	matched := make([]*domain.Enchant, 0, len(cat.EnchantList))
	for _, enchant := range cat.EnchantList {
		if filter.Slot != "" && !enchantAdmitsTypes(enchant, types) {
			continue
		}
		if filter.Class != "" && len(enchant.ClassAllowlist) > 0 &&
			!slices.Contains(enchant.ClassAllowlist, filter.Class) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(enchant.Name), query) {
			continue
		}
		matched = append(matched, enchant)
	}

	sort.SliceStable(matched, func(i, j int) bool {
		if matched[i].Lvl != matched[j].Lvl {
			return matched[i].Lvl > matched[j].Lvl
		}
		if matched[i].EffectID != matched[j].EffectID {
			return matched[i].EffectID < matched[j].EffectID
		}
		return matched[i].ItemID < matched[j].ItemID
	})

	page, _, _ := paginate(len(matched), filter.Limit, filter.Offset)
	enchants := make([]domain.Enchant, 0, len(page))
	for _, index := range page {
		enchants = append(enchants, *matched[index])
	}
	return enchants
}

// TalentTrees returns the pinned trees for a class.
func TalentTrees(cat *Catalog, class domain.Class) (*domain.TalentClass, error) {
	if class == "" {
		return nil, domain.ErrUnsupportedClass
	}
	trees, ok := cat.Talents[class]
	if !ok {
		return nil, domain.ErrUnsupportedClass
	}
	return trees, nil
}

// Compatibility applies the static item rules for one character and slot. It is
// the single implementation the equipment validator reuses.
func (c *Catalog) Compatibility(ch domain.Character, item *domain.Item, slot domain.Slot) (bool, string) {
	if item == nil {
		return false, "unknown_item"
	}
	if ch.Class != "" && len(item.ClassAllowlist) > 0 &&
		!slices.Contains(item.ClassAllowlist, ch.Class) {
		return false, "incompatible_class"
	}
	if ch.Class != "" && ch.Race != "" {
		if races, ok := c.Capabilities.Races[ch.Class]; ok && !slices.Contains(races, ch.Race) {
			return false, "race_class_mismatch"
		}
	}
	if ch.Class != "" && item.ArmorType != "" &&
		!slices.Contains(c.Capabilities.Armor[ch.Class], item.ArmorType) {
		return false, "incompatible_armor"
	}
	if ch.Class != "" && item.WeaponType != "" {
		eligible := false
		for _, weapon := range c.Capabilities.Weapons[ch.Class] {
			if weapon.WeaponType != item.WeaponType {
				continue
			}
			eligible = !(item.HandType == "two_hand" && !weapon.CanUseTwoHand)
			break
		}
		if !eligible {
			return false, "incompatible_weapon"
		}
	}
	if ch.Class != "" && item.RangedWeaponType != "" &&
		!slices.Contains(c.Capabilities.Ranged[ch.Class], item.RangedWeaponType) {
		return false, "incompatible_ranged"
	}
	if slot != "" && !slices.Contains(item.Slots, slot) {
		return false, "incompatible_slot"
	}
	if item.FactionRestriction != "" && ch.Race != "" {
		if faction := raceFaction(ch.Race); faction != "" && faction != item.FactionRestriction {
			return false, "incompatible_faction"
		}
	}
	if item.RequiredProfession != "" && !slices.Contains(ch.Professions, item.RequiredProfession) {
		return false, "incompatible_profession"
	}
	return true, ""
}

// SlotItemTypes maps a canonical slot to the item types an enchant applies to.
func SlotItemTypes(slot domain.Slot) []string {
	switch slot {
	case domain.SlotHead:
		return []string{"head"}
	case domain.SlotNeck:
		return []string{"neck"}
	case domain.SlotShoulder:
		return []string{"shoulder"}
	case domain.SlotBack:
		return []string{"back"}
	case domain.SlotChest:
		return []string{"chest"}
	case domain.SlotWrist:
		return []string{"wrist"}
	case domain.SlotHands:
		return []string{"hands"}
	case domain.SlotWaist:
		return []string{"waist"}
	case domain.SlotLegs:
		return []string{"legs"}
	case domain.SlotFeet:
		return []string{"feet"}
	case domain.SlotFinger1, domain.SlotFinger2:
		return []string{"finger"}
	case domain.SlotTrinket1, domain.SlotTrinket2:
		return []string{"trinket"}
	case domain.SlotMainHand, domain.SlotOffHand:
		return []string{"weapon"}
	case domain.SlotRanged:
		return []string{"ranged"}
	default:
		return nil
	}
}

// raceFaction maps the standard races to their faction. Custom Skyborne races
// are left unknown rather than guessed.
func raceFaction(race domain.Race) string {
	switch race {
	case "human", "dwarf", "night_elf", "gnome":
		return "alliance"
	case "orc", "undead", "tauren", "troll":
		return "horde"
	default:
		return ""
	}
}

func itemHasSource(item *domain.Item, kind string) bool {
	for _, source := range item.Sources {
		if strings.EqualFold(source.Kind, kind) {
			return true
		}
	}
	return false
}

func enchantAdmitsTypes(enchant *domain.Enchant, types []string) bool {
	for _, itemType := range types {
		if enchant.Type == itemType {
			return true
		}
		if slices.Contains(enchant.ExtraTypes, itemType) {
			return true
		}
	}
	return false
}

// paginate clamps the limit and offset and returns the selected indexes plus
// the effective limit and offset.
func paginate(total, limit, offset int) ([]int, int, int) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	indexes := make([]int, 0, end-offset)
	for index := offset; index < end; index++ {
		indexes = append(indexes, index)
	}
	return indexes, limit, offset
}
