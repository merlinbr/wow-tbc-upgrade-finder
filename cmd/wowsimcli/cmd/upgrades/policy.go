package upgrades

import (
	"fmt"

	"github.com/wowsims/tbc/sim/core/proto"
)

func isLegalGemForSocket(gemColor proto.GemColor, socketColor proto.GemColor) bool {
	if socketColor == proto.GemColor_GemColorMeta {
		return gemColor == proto.GemColor_GemColorMeta
	}
	if gemColor == proto.GemColor_GemColorMeta {
		return false
	}
	switch socketColor {
	case proto.GemColor_GemColorRed:
		return gemColor == proto.GemColor_GemColorRed ||
			gemColor == proto.GemColor_GemColorOrange ||
			gemColor == proto.GemColor_GemColorPurple ||
			gemColor == proto.GemColor_GemColorPrismatic
	case proto.GemColor_GemColorBlue:
		return gemColor == proto.GemColor_GemColorBlue ||
			gemColor == proto.GemColor_GemColorGreen ||
			gemColor == proto.GemColor_GemColorPurple ||
			gemColor == proto.GemColor_GemColorPrismatic
	case proto.GemColor_GemColorYellow:
		return gemColor == proto.GemColor_GemColorYellow ||
			gemColor == proto.GemColor_GemColorOrange ||
			gemColor == proto.GemColor_GemColorGreen ||
			gemColor == proto.GemColor_GemColorPrismatic
	default:
		return gemColor != proto.GemColor_GemColorMeta
	}
}

// ApplyPolicy configures the candidate equipment replacement with declared gems and enchants.
func ApplyPolicy(candidate Candidate, policy ItemPolicy, catalog *Catalog) (Candidate, *PolicyError) {
	item, ok := catalog.Items[candidate.Item.ID]
	if !ok || item == nil {
		return candidate, &PolicyError{Reason: "item not found in catalog"}
	}

	gemSockets := item.GetGemSockets()
	appliedGemIDs := make([]int32, len(gemSockets))
	socketChoices := make([]int32, len(gemSockets))

	player := candidate.Request.GetRaid().GetParties()[0].GetPlayers()[0]

	for i, socketColor := range gemSockets {
		gemID, hasPolicy := policy.GemBySocket[socketColor]
		if hasPolicy && gemID > 0 {
			gem, ok := catalog.Gems[gemID]
			if !ok || gem == nil {
				return candidate, &PolicyError{Reason: fmt.Sprintf("configured gem ID %d not found in database", gemID)}
			}

			if !isLegalGemForSocket(gem.GetColor(), socketColor) {
				return candidate, &PolicyError{Reason: fmt.Sprintf("gem %d color %v cannot be socketed into %v socket", gemID, gem.GetColor(), socketColor)}
			}

			if policy.MaxGemQuality > proto.ItemQuality_ItemQualityJunk && gem.GetQuality() > policy.MaxGemQuality {
				return candidate, &PolicyError{Reason: fmt.Sprintf("gem %d quality %v exceeds max quality %v", gemID, gem.GetQuality(), policy.MaxGemQuality)}
			}

			if gem.GetRequiredProfession() != proto.Profession_ProfessionUnknown {
				if gem.GetRequiredProfession() != player.GetProfession1() && gem.GetRequiredProfession() != player.GetProfession2() {
					return candidate, &PolicyError{Reason: fmt.Sprintf("gem %d requires profession %v", gemID, gem.GetRequiredProfession())}
				}
			}

			if gem.GetUnique() {
				count := 0
				for slotIdx, eq := range player.GetEquipment().GetItems() {
					if proto.ItemSlot(slotIdx) == candidate.TargetSlot {
						continue
					}
					if eq != nil {
						for _, gid := range eq.GetGems() {
							if gid == gemID {
								count++
							}
						}
					}
				}
				for j := range i {
					if appliedGemIDs[j] == gemID {
						count++
					}
				}
				if count > 0 {
					return candidate, &PolicyError{Reason: fmt.Sprintf("unique gem %d is already equipped", gemID)}
				}
			}

			appliedGemIDs[i] = gemID
			socketChoices[i] = gemID
		} else {
			appliedGemIDs[i] = 0
			socketChoices[i] = 0
		}
	}

	var appliedEnchantID int32
	enchantID, hasEnchantPolicy := policy.EnchantByType[item.GetType()]
	if hasEnchantPolicy && enchantID > 0 {
		var enchant *proto.UIEnchant
		if e, ok := catalog.Enchants[enchantID]; ok {
			enchant = e
		} else if e, ok := catalog.EnchantsByID[enchantID]; ok {
			enchant = e
		}

		if enchant == nil {
			return candidate, &PolicyError{Reason: fmt.Sprintf("configured enchant ID %d not found in database", enchantID)}
		}

		typeMatches := (enchant.GetType() == item.GetType())
		if !typeMatches {
			for _, ext := range enchant.GetExtraTypes() {
				if ext == item.GetType() {
					typeMatches = true
					break
				}
			}
		}
		if !typeMatches {
			return candidate, &PolicyError{Reason: fmt.Sprintf("enchant %d type %v cannot be applied to item type %v", enchantID, enchant.GetType(), item.GetType())}
		}

		switch enchant.GetEnchantType() {
		case proto.EnchantType_EnchantTypeTwoHand:
			if item.GetHandType() != proto.HandType_HandTypeTwoHand {
				return candidate, &PolicyError{Reason: "enchant requires two-hand weapon"}
			}
		case proto.EnchantType_EnchantTypeShield:
			if item.GetWeaponType() != proto.WeaponType_WeaponTypeShield {
				return candidate, &PolicyError{Reason: "enchant requires shield"}
			}
		case proto.EnchantType_EnchantTypeOffHand:
			if item.GetWeaponType() != proto.WeaponType_WeaponTypeOffHand && item.GetHandType() != proto.HandType_HandTypeOffHand {
				return candidate, &PolicyError{Reason: "enchant requires off-hand item"}
			}
		}

		if len(enchant.GetClassAllowlist()) > 0 {
			allowed := false
			for _, c := range enchant.GetClassAllowlist() {
				if c == player.GetClass() {
					allowed = true
					break
				}
			}
			if !allowed {
				return candidate, &PolicyError{Reason: fmt.Sprintf("enchant %d not allowed for class %v", enchantID, player.GetClass())}
			}
		}

		if enchant.GetRequiredProfession() != proto.Profession_ProfessionUnknown {
			if enchant.GetRequiredProfession() != player.GetProfession1() && enchant.GetRequiredProfession() != player.GetProfession2() {
				return candidate, &PolicyError{Reason: fmt.Sprintf("enchant %d requires profession %v", enchantID, enchant.GetRequiredProfession())}
			}
		}

		appliedEnchantID = enchant.GetEffectId()
	}

	clonedReq := cloneMessage(candidate.Request)
	targetPlayer := clonedReq.GetRaid().GetParties()[0].GetPlayers()[0]
	targetItemSpec := targetPlayer.GetEquipment().GetItems()[int(candidate.TargetSlot)]
	targetItemSpec.Gems = appliedGemIDs
	targetItemSpec.Enchant = appliedEnchantID

	candidateCopy := candidate
	candidateCopy.Request = clonedReq
	candidateCopy.Applied = PolicyApplication{
		GemIDs:        appliedGemIDs,
		EnchantID:     appliedEnchantID,
		SocketChoices: socketChoices,
	}

	return candidateCopy, nil
}
