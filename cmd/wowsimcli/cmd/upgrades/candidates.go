package upgrades

import (
	"sort"
	"strings"

	"github.com/wowsims/tbc/sim/core/proto"
)

func isHorde(race proto.Race) bool {
	switch race {
	case proto.Race_RaceOrc, proto.Race_RaceUndead, proto.Race_RaceTauren, proto.Race_RaceTroll, proto.Race_RaceBloodElf:
		return true
	default:
		return false
	}
}

func canWearArmor(class proto.Class, armorType proto.ArmorType) bool {
	switch armorType {
	case proto.ArmorType_ArmorTypeUnknown:
		return true
	case proto.ArmorType_ArmorTypeCloth:
		return true
	case proto.ArmorType_ArmorTypeLeather:
		switch class {
		case proto.Class_ClassRogue, proto.Class_ClassDruid, proto.Class_ClassHunter, proto.Class_ClassShaman, proto.Class_ClassPaladin, proto.Class_ClassWarrior:
			return true
		default:
			return false
		}
	case proto.ArmorType_ArmorTypeMail:
		switch class {
		case proto.Class_ClassHunter, proto.Class_ClassShaman, proto.Class_ClassPaladin, proto.Class_ClassWarrior:
			return true
		default:
			return false
		}
	case proto.ArmorType_ArmorTypePlate:
		switch class {
		case proto.Class_ClassPaladin, proto.Class_ClassWarrior:
			return true
		default:
			return false
		}
	default:
		return true
	}
}

func canUseWeapon(class proto.Class, weaponType proto.WeaponType) bool {
	switch weaponType {
	case proto.WeaponType_WeaponTypeUnknown:
		return true
	case proto.WeaponType_WeaponTypeDagger:
		return class != proto.Class_ClassPaladin
	case proto.WeaponType_WeaponTypeSword:
		switch class {
		case proto.Class_ClassMage, proto.Class_ClassWarlock, proto.Class_ClassRogue, proto.Class_ClassHunter, proto.Class_ClassPaladin, proto.Class_ClassWarrior:
			return true
		default:
			return false
		}
	case proto.WeaponType_WeaponTypeMace:
		switch class {
		case proto.Class_ClassPriest, proto.Class_ClassRogue, proto.Class_ClassDruid, proto.Class_ClassShaman, proto.Class_ClassPaladin, proto.Class_ClassWarrior:
			return true
		default:
			return false
		}
	case proto.WeaponType_WeaponTypeAxe:
		switch class {
		case proto.Class_ClassHunter, proto.Class_ClassShaman, proto.Class_ClassPaladin, proto.Class_ClassWarrior:
			return true
		default:
			return false
		}
	case proto.WeaponType_WeaponTypePolearm:
		switch class {
		case proto.Class_ClassDruid, proto.Class_ClassHunter, proto.Class_ClassPaladin, proto.Class_ClassWarrior:
			return true
		default:
			return false
		}
	case proto.WeaponType_WeaponTypeStaff:
		switch class {
		case proto.Class_ClassMage, proto.Class_ClassPriest, proto.Class_ClassWarlock, proto.Class_ClassDruid, proto.Class_ClassHunter, proto.Class_ClassShaman, proto.Class_ClassWarrior:
			return true
		default:
			return false
		}
	case proto.WeaponType_WeaponTypeFist:
		switch class {
		case proto.Class_ClassRogue, proto.Class_ClassHunter, proto.Class_ClassShaman, proto.Class_ClassWarrior:
			return true
		default:
			return false
		}
	case proto.WeaponType_WeaponTypeShield:
		switch class {
		case proto.Class_ClassPaladin, proto.Class_ClassShaman, proto.Class_ClassWarrior:
			return true
		default:
			return false
		}
	case proto.WeaponType_WeaponTypeOffHand:
		return true
	default:
		return true
	}
}

func canUseRanged(class proto.Class, rangedType proto.RangedWeaponType) bool {
	switch rangedType {
	case proto.RangedWeaponType_RangedWeaponTypeUnknown:
		return true
	case proto.RangedWeaponType_RangedWeaponTypeBow, proto.RangedWeaponType_RangedWeaponTypeCrossbow, proto.RangedWeaponType_RangedWeaponTypeGun:
		switch class {
		case proto.Class_ClassRogue, proto.Class_ClassHunter, proto.Class_ClassWarrior:
			return true
		default:
			return false
		}
	case proto.RangedWeaponType_RangedWeaponTypeWand:
		switch class {
		case proto.Class_ClassMage, proto.Class_ClassPriest, proto.Class_ClassWarlock:
			return true
		default:
			return false
		}
	case proto.RangedWeaponType_RangedWeaponTypeThrown:
		switch class {
		case proto.Class_ClassRogue, proto.Class_ClassWarrior:
			return true
		default:
			return false
		}
	case proto.RangedWeaponType_RangedWeaponTypeIdol:
		return class == proto.Class_ClassDruid
	case proto.RangedWeaponType_RangedWeaponTypeLibram:
		return class == proto.Class_ClassPaladin
	case proto.RangedWeaponType_RangedWeaponTypeTotem:
		return class == proto.Class_ClassShaman
	default:
		return false
	}
}

func canDualWield(class proto.Class) bool {
	switch class {
	case proto.Class_ClassWarrior, proto.Class_ClassRogue, proto.Class_ClassHunter, proto.Class_ClassShaman:
		return true
	default:
		return false
	}
}

func getEligibleSlots(item *proto.UIItem, class proto.Class) []proto.ItemSlot {
	switch item.GetType() {
	case proto.ItemType_ItemTypeHead:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotHead}
	case proto.ItemType_ItemTypeNeck:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotNeck}
	case proto.ItemType_ItemTypeShoulder:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotShoulder}
	case proto.ItemType_ItemTypeBack:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotBack}
	case proto.ItemType_ItemTypeChest:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotChest}
	case proto.ItemType_ItemTypeWrist:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotWrist}
	case proto.ItemType_ItemTypeHands:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotHands}
	case proto.ItemType_ItemTypeWaist:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotWaist}
	case proto.ItemType_ItemTypeLegs:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotLegs}
	case proto.ItemType_ItemTypeFeet:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotFeet}
	case proto.ItemType_ItemTypeFinger:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotFinger1, proto.ItemSlot_ItemSlotFinger2}
	case proto.ItemType_ItemTypeTrinket:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotTrinket1, proto.ItemSlot_ItemSlotTrinket2}
	case proto.ItemType_ItemTypeRanged:
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotRanged}
	case proto.ItemType_ItemTypeWeapon:
		if item.GetWeaponType() == proto.WeaponType_WeaponTypeShield || item.GetWeaponType() == proto.WeaponType_WeaponTypeOffHand || item.GetHandType() == proto.HandType_HandTypeOffHand {
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotOffHand}
		}
		if item.GetHandType() == proto.HandType_HandTypeMainHand || item.GetHandType() == proto.HandType_HandTypeTwoHand {
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}
		}
		// OneHand weapon
		if canDualWield(class) {
			return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand, proto.ItemSlot_ItemSlotOffHand}
		}
		return []proto.ItemSlot{proto.ItemSlot_ItemSlotMainHand}
	default:
		return nil
	}
}

func normalizeFilters(f ContentFilters) ContentFilters {
	res := f
	if len(res.SourceKinds) > 0 {
		seen := make(map[proto.SourceFilterOption]bool)
		var unique []proto.SourceFilterOption
		for _, k := range res.SourceKinds {
			if !seen[k] {
				seen[k] = true
				unique = append(unique, k)
			}
		}
		sort.Slice(unique, func(i, j int) bool { return unique[i] < unique[j] })
		res.SourceKinds = unique
	}
	if len(res.SourceNames) > 0 {
		seen := make(map[string]bool)
		var unique []string
		for _, n := range res.SourceNames {
			trimmed := strings.TrimSpace(n)
			if trimmed != "" && !seen[trimmed] {
				seen[trimmed] = true
				unique = append(unique, trimmed)
			}
		}
		sort.Strings(unique)
		res.SourceNames = unique
	}
	return res
}

func BuildCandidates(imported *ImportedSettings, filters ContentFilters, policy ItemPolicy, catalog *Catalog) (BuildResult, error) {
	filters = normalizeFilters(filters)
	excluded := ExclusionSummary{
		Reasons: make(map[string]int),
	}

	player := imported.Settings.GetPlayer()
	class := player.GetClass()
	race := player.GetRace()
	hordePlayer := isHorde(race)
	prof1 := player.GetProfession1()
	prof2 := player.GetProfession2()

	// Map of baseline equipped item specs by slot
	equippedSlots := make([]*proto.ItemSpec, 17)
	if player.GetEquipment() != nil {
		for i, it := range player.GetEquipment().GetItems() {
			if i < len(equippedSlots) {
				equippedSlots[i] = it
			}
		}
	}

	var candidates []Candidate

	// Sort item IDs for deterministic iteration
	var itemIDs []int32
	for id := range catalog.Items {
		itemIDs = append(itemIDs, id)
	}
	sort.Slice(itemIDs, func(i, j int) bool { return itemIDs[i] < itemIDs[j] })

	for _, id := range itemIDs {
		item := catalog.Items[id]

		// 1. Source check
		sources := item.GetSources()
		if len(sources) == 0 {
			if !filters.IncludeUnknown {
				excluded.UnknownSource++
				excluded.Reasons["unknown_source"]++
				continue
			}
		} else {
			// Phase check
			if filters.MaxPhase > 0 && item.GetPhase() > filters.MaxPhase {
				excluded.Source++
				excluded.Reasons["source_phase"]++
				continue
			}

			resolvedSource := catalog.ResolveSource(item)

			// Source kinds check
			if len(filters.SourceKinds) > 0 {
				matched := false
				for _, k := range filters.SourceKinds {
					if resolvedSource.Kind == k {
						matched = true
						break
					}
				}
				if !matched {
					excluded.Source++
					excluded.Reasons["source_kind"]++
					continue
				}
			}

			// Source names check
			if len(filters.SourceNames) > 0 {
				matched := false
				for _, name := range filters.SourceNames {
					if strings.EqualFold(resolvedSource.Name, name) ||
						strings.EqualFold(resolvedSource.Zone, name) ||
						(resolvedSource.Name != "" && strings.Contains(strings.ToLower(resolvedSource.Name), strings.ToLower(name))) ||
						(resolvedSource.Zone != "" && strings.Contains(strings.ToLower(resolvedSource.Zone), strings.ToLower(name))) {
						matched = true
						break
					}
				}
				if !matched {
					excluded.Source++
					excluded.Reasons["source_name"]++
					continue
				}
			}
		}

		// 2. Class check
		if len(item.GetClassAllowlist()) > 0 {
			allowed := false
			for _, c := range item.GetClassAllowlist() {
				if c == class {
					allowed = true
					break
				}
			}
			if !allowed {
				excluded.Compatibility++
				excluded.Reasons["incompatible_class"]++
				continue
			}
		}

		// 3. Faction check
		switch item.GetFactionRestriction() {
		case proto.UIItem_FACTION_RESTRICTION_ALLIANCE_ONLY:
			if hordePlayer {
				excluded.Compatibility++
				excluded.Reasons["incompatible_faction"]++
				continue
			}
		case proto.UIItem_FACTION_RESTRICTION_HORDE_ONLY:
			if !hordePlayer {
				excluded.Compatibility++
				excluded.Reasons["incompatible_faction"]++
				continue
			}
		}

		// 4. Profession check
		if reqProf := item.GetRequiredProfession(); reqProf != proto.Profession_ProfessionUnknown {
			if reqProf != prof1 && reqProf != prof2 {
				excluded.Compatibility++
				excluded.Reasons["incompatible_profession"]++
				continue
			}
		}

		// 5. Unique check across all equipped gear
		if item.GetUnique() {
			alreadyEquipped := false
			for _, eq := range equippedSlots {
				if eq != nil && eq.GetId() == item.GetId() {
					alreadyEquipped = true
					break
				}
			}
			if alreadyEquipped {
				excluded.Compatibility++
				excluded.Reasons["unique_equipped_conflict"]++
				continue
			}
		}

		// 6. Armor proficiency check
		if item.GetType() != proto.ItemType_ItemTypeWeapon &&
			item.GetType() != proto.ItemType_ItemTypeRanged &&
			item.GetType() != proto.ItemType_ItemTypeFinger &&
			item.GetType() != proto.ItemType_ItemTypeTrinket &&
			item.GetType() != proto.ItemType_ItemTypeNeck &&
			item.GetType() != proto.ItemType_ItemTypeBack {
			if !canWearArmor(class, item.GetArmorType()) {
				excluded.Compatibility++
				excluded.Reasons["incompatible_armor"]++
				continue
			}
		}

		// 7. Weapon proficiency check
		if item.GetType() == proto.ItemType_ItemTypeWeapon {
			if !canUseWeapon(class, item.GetWeaponType()) {
				excluded.Compatibility++
				excluded.Reasons["incompatible_weapon"]++
				continue
			}
		}

		// 8. Ranged proficiency check
		if item.GetType() == proto.ItemType_ItemTypeRanged {
			if !canUseRanged(class, item.GetRangedWeaponType()) {
				excluded.Compatibility++
				excluded.Reasons["incompatible_ranged"]++
				continue
			}
		}

		// 9. Generate candidate variants for each legal slot
		slots := getEligibleSlots(item, class)
		if len(slots) == 0 {
			excluded.Compatibility++
			excluded.Reasons["incompatible_slot"]++
			continue
		}

		for _, targetSlot := range slots {
			// Skip if already wearing this exact item in targetSlot
			if eq := equippedSlots[int(targetSlot)]; eq != nil && eq.GetId() == item.GetId() {
				continue
			}

			// Check limit_category constraint across other equipped slots
			conflict := false
			if item.GetLimitCategory() > 0 {
				for s, eq := range equippedSlots {
					if eq == nil || eq.GetId() == 0 {
						continue
					}
					if proto.ItemSlot(s) == targetSlot {
						continue
					}
					if item.GetHandType() == proto.HandType_HandTypeTwoHand && proto.ItemSlot(s) == proto.ItemSlot_ItemSlotOffHand {
						continue
					}
					if eqItem, ok := catalog.Items[eq.GetId()]; ok {
						if eqItem.GetLimitCategory() == item.GetLimitCategory() {
							conflict = true
							break
						}
					}
				}
			}
			if conflict {
				excluded.Compatibility++
				excluded.Reasons["limit_category_conflict"]++
				continue
			}

			// If equipping OffHand, check if current MainHand is 2H
			if targetSlot == proto.ItemSlot_ItemSlotOffHand {
				mh := equippedSlots[int(proto.ItemSlot_ItemSlotMainHand)]
				if mh != nil && mh.GetId() > 0 {
					if mhItem, ok := catalog.Items[mh.GetId()]; ok && mhItem.GetHandType() == proto.HandType_HandTypeTwoHand {
						excluded.Compatibility++
						excluded.Reasons["two_hand_equipped_conflict"]++
						continue
					}
				}
			}

			// Displaced items
			var displaced []UIItemSummary
			if oldSpec := equippedSlots[int(targetSlot)]; oldSpec != nil && oldSpec.GetId() > 0 {
				if oldItem, ok := catalog.Items[oldSpec.GetId()]; ok {
					displaced = append(displaced, catalog.ItemSummary(oldItem, targetSlot))
				} else {
					displaced = append(displaced, UIItemSummary{
						ID:   oldSpec.GetId(),
						Slot: targetSlot,
					})
				}
			}
			if item.GetHandType() == proto.HandType_HandTypeTwoHand && targetSlot == proto.ItemSlot_ItemSlotMainHand {
				if oldOffHand := equippedSlots[int(proto.ItemSlot_ItemSlotOffHand)]; oldOffHand != nil && oldOffHand.GetId() > 0 {
					if oldItem, ok := catalog.Items[oldOffHand.GetId()]; ok {
						displaced = append(displaced, catalog.ItemSummary(oldItem, proto.ItemSlot_ItemSlotOffHand))
					} else {
						displaced = append(displaced, UIItemSummary{
							ID:   oldOffHand.GetId(),
							Slot: proto.ItemSlot_ItemSlotOffHand,
						})
					}
				}
			}

			// Create cloned request with candidate replacement
			clonedSettings := cloneMessage(imported.Settings)
			clonedEquip := cloneMessage(clonedSettings.GetPlayer().GetEquipment())
			for len(clonedEquip.Items) < 17 {
				clonedEquip.Items = append(clonedEquip.Items, &proto.ItemSpec{})
			}

			newItemSpec := &proto.ItemSpec{
				Id: item.GetId(),
			}
			clonedEquip.Items[int(targetSlot)] = newItemSpec

			if item.GetHandType() == proto.HandType_HandTypeTwoHand && targetSlot == proto.ItemSlot_ItemSlotMainHand {
				clonedEquip.Items[int(proto.ItemSlot_ItemSlotOffHand)] = &proto.ItemSpec{}
			}
			clonedSettings.GetPlayer().Equipment = clonedEquip

			tempImported := &ImportedSettings{
				Settings: clonedSettings,
			}
			req := tempImported.NewRequest(0)

			candidate := Candidate{
				Item:       catalog.ItemSummary(item, targetSlot),
				TargetSlot: targetSlot,
				Displaced:  displaced,
				Request:    req,
				Source:     catalog.ResolveSource(item),
			}

			candidates = append(candidates, candidate)
		}
	}

	return BuildResult{
		Candidates: candidates,
		Excluded:   excluded,
	}, nil
}
