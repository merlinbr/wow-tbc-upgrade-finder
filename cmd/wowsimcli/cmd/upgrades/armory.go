package upgrades

import (
	"fmt"
	"strings"
	"sync"
	"unicode"

	"github.com/wowsims/tbc/assets/enchants"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/stats"
)

type armorySlot struct {
	slot  proto.ItemSlot
	name  string
	index int
}

var canonicalGearSlots = []armorySlot{
	{proto.ItemSlot_ItemSlotHead, "Head", int(proto.ItemSlot_ItemSlotHead)},
	{proto.ItemSlot_ItemSlotNeck, "Neck", int(proto.ItemSlot_ItemSlotNeck)},
	{proto.ItemSlot_ItemSlotShoulder, "Shoulder", int(proto.ItemSlot_ItemSlotShoulder)},
	{proto.ItemSlot_ItemSlotBack, "Back", int(proto.ItemSlot_ItemSlotBack)},
	{proto.ItemSlot_ItemSlotChest, "Chest", int(proto.ItemSlot_ItemSlotChest)},
	{proto.ItemSlot_ItemSlotWrist, "Wrist", int(proto.ItemSlot_ItemSlotWrist)},
	{proto.ItemSlot_ItemSlotMainHand, "Main Hand", int(proto.ItemSlot_ItemSlotMainHand)},
	{proto.ItemSlot_ItemSlotOffHand, "Off Hand", int(proto.ItemSlot_ItemSlotOffHand)},
	{proto.ItemSlot_ItemSlotHands, "Hands", int(proto.ItemSlot_ItemSlotHands)},
	{proto.ItemSlot_ItemSlotWaist, "Waist", int(proto.ItemSlot_ItemSlotWaist)},
	{proto.ItemSlot_ItemSlotLegs, "Legs", int(proto.ItemSlot_ItemSlotLegs)},
	{proto.ItemSlot_ItemSlotFeet, "Feet", int(proto.ItemSlot_ItemSlotFeet)},
	{proto.ItemSlot_ItemSlotFinger1, "Finger 1", int(proto.ItemSlot_ItemSlotFinger1)},
	{proto.ItemSlot_ItemSlotFinger2, "Finger 2", int(proto.ItemSlot_ItemSlotFinger2)},
	{proto.ItemSlot_ItemSlotTrinket1, "Trinket 1", int(proto.ItemSlot_ItemSlotTrinket1)},
	{proto.ItemSlot_ItemSlotTrinket2, "Trinket 2", int(proto.ItemSlot_ItemSlotTrinket2)},
	{proto.ItemSlot_ItemSlotRanged, "Ranged", int(proto.ItemSlot_ItemSlotRanged)},
}

var enchantDescriptions = sync.OnceValue(func() map[int32]string { return enchants.Descriptions() })

func statKey(name string) string {
	runes := []rune(name)
	var b strings.Builder
	for index, current := range runes {
		if index > 0 && unicode.IsUpper(current) {
			previous := runes[index-1]
			nextIsLower := index+1 < len(runes) && unicode.IsLower(runes[index+1])
			if unicode.IsLower(previous) || (unicode.IsUpper(previous) && nextIsLower) {
				b.WriteByte('_')
			}
		}
		b.WriteRune(unicode.ToLower(current))
	}
	return b.String()
}

func statMap(values []float64) map[string]float64 {
	return statValuesMap(stats.FromProtoArray(values))
}

func statValuesMap(values stats.Stats) map[string]float64 {
	result := make(map[string]float64)
	for stat := stats.Stat(0); int(stat) < stats.ProtoStatsLen; stat++ {
		if value := values[stat]; value != 0 {
			result[statKey(stat.StatName())] = value
		}
	}
	return result
}

func itemBaseStats(item *proto.UIItem) stats.Stats {
	if item != nil {
		if scaling, ok := item.GetScalingOptions()[0]; ok && scaling != nil {
			return stats.FromProtoMap(scaling.GetStats())
		}
		return stats.FromProtoArray(item.GetStats())
	}
	return stats.Stats{}
}

func itemRandPropPoints(item *proto.UIItem) int32 {
	if item != nil {
		if scaling, ok := item.GetScalingOptions()[0]; ok && scaling != nil {
			return scaling.GetRandPropPoints()
		}
		return item.GetRandPropPoints()
	}
	return 0
}

// itemIlvl reads the item's level from its own scaling bucket; the bundled
// database leaves UIItem.ilvl unset and stores the level per scaling option.
func itemIlvl(item *proto.UIItem) int32 {
	if scaling := item.GetScalingOptions()[0]; scaling != nil && scaling.GetIlvl() > 0 {
		return scaling.GetIlvl()
	}
	return item.GetIlvl()
}

func scaledStatMap(values []float64, randPropPoints int32) map[string]float64 {
	return statValuesMap(stats.FromProtoArray(values).Multiply(float64(randPropPoints) / 10000).Floor())
}

// armoryComputeRequest builds the full-settings snapshot ranking will simulate:
// buffs, consumes, talents, and bonus stats from the imported baseline are
// retained so the armory stats match both WoWSims and the ranking run.
func armoryComputeRequest(imported *ImportedSettings) *proto.ComputeStatsRequest {
	raid, encounter := imported.raidAndEncounter()
	raid.Parties[0].Players[0].Database = buildSimDatabase()
	return &proto.ComputeStatsRequest{Raid: raid, Encounter: encounter}
}

func enrichGem(gem *proto.UIGem) *GemData {
	if gem == nil {
		return nil
	}
	return &GemData{
		ID:    gem.GetId(),
		Name:  gem.GetName(),
		Icon:  gem.GetIcon(),
		Color: gem.GetColor(),
		Stats: statMap(gem.GetStats()),
	}
}

func enrichItem(spec *proto.ItemSpec, item *proto.UIItem, catalog *Catalog) (GearSlotData, error) {
	data := GearSlotData{Stats: make(map[string]float64), Sockets: []SocketData{}, SocketBonus: SocketBonusData{Stats: make(map[string]float64)}}
	if item == nil || spec == nil || spec.GetId() == 0 {
		return data, nil
	}

	data.ItemID = item.GetId()
	data.ItemName = item.GetName()
	data.Quality = item.GetQuality()
	data.Icon = item.GetIcon()
	data.Phase = item.GetPhase()
	data.SetName = item.GetSetName()
	data.Ilvl = itemIlvl(item)
	data.Stats = statValuesMap(itemBaseStats(item))

	if suffixID := spec.GetRandomSuffix(); suffixID != 0 {
		suffix := catalog.RandomSuffixes[suffixID]
		if suffix == nil {
			return GearSlotData{}, fmt.Errorf("random suffix %d not found in catalog", suffixID)
		}
		data.RandomSuffix = &RandomSuffixData{
			ID:    suffix.GetId(),
			Name:  suffix.GetName(),
			Stats: scaledStatMap(suffix.GetStats(), itemRandPropPoints(item)),
		}
	}

	for socketIndex, color := range item.GetGemSockets() {
		socket := SocketData{Color: color}
		if socketIndex < len(spec.GetGems()) {
			gemID := spec.GetGems()[socketIndex]
			if gemID != 0 {
				gem := catalog.Gems[gemID]
				if gem == nil {
					return GearSlotData{}, fmt.Errorf("gem %d not found in catalog", gemID)
				}
				socket.Gem = enrichGem(gem)
			}
		}
		data.Sockets = append(data.Sockets, socket)
	}

	data.SocketBonus.Stats = statMap(item.GetSocketBonus())
	data.SocketBonus.Active = len(data.Sockets) > 0 && len(data.Sockets) <= len(spec.GetGems())
	for index, socket := range data.Sockets {
		if socket.Gem == nil || !core.ColorIntersects(socket.Color, socket.Gem.Color) {
			data.SocketBonus.Active = false
			break
		}
		if index >= len(spec.GetGems()) || spec.GetGems()[index] == 0 {
			data.SocketBonus.Active = false
			break
		}
	}

	if enchantID := spec.GetEnchant(); enchantID != 0 {
		enchant := catalog.Enchants[enchantID]
		if enchant == nil {
			return GearSlotData{}, fmt.Errorf("enchant %d not found in catalog", enchantID)
		}
		description := enchantDescriptions()[enchant.GetEffectId()]
		if description == "" {
			description = enchant.GetName()
		}
		data.Enchant = &EnchantData{
			ID:          enchant.GetEffectId(),
			Name:        enchant.GetName(),
			Icon:        enchant.GetIcon(),
			Description: description,
			Stats:       statMap(enchant.GetStats()),
		}
	}

	return data, nil
}

// panelDebuffStats mirrors the wowsims site's CharacterStats.getDebuffStats
// (ui/core/components/character_stats.tsx): the stats panel adds these
// target-debuff benefits on top of the engine's finalStats, so the armory must
// do the same to match the site's displayed numbers.
// ponytail: expose-weakness uptime only gates the contribution (truthiness),
// matching the site's display simplification; it is not scaled by uptime.
func panelDebuffStats(debuffs *proto.Debuffs) (stats.Stats, map[proto.PseudoStat]float64) {
	extra := stats.Stats{}
	pseudo := map[proto.PseudoStat]float64{}
	if debuffs == nil {
		return extra, pseudo
	}
	if debuffs.GetFaerieFire() == proto.TristateEffect_TristateEffectImproved {
		pseudo[proto.PseudoStat_PseudoStatMeleeHitPercent] += 3
		pseudo[proto.PseudoStat_PseudoStatRangedHitPercent] += 3
	}
	if debuffs.GetImprovedSealOfTheCrusader() != proto.TristateEffect_TristateEffectMissing {
		pseudo[proto.PseudoStat_PseudoStatMeleeCritPercent] += 3
		pseudo[proto.PseudoStat_PseudoStatRangedCritPercent] += 3
		pseudo[proto.PseudoStat_PseudoStatSpellCritPercent] += 3
	}
	if debuffs.GetExposeWeaknessUptime() != 0 && debuffs.GetExposeWeaknessHunterAgility() != 0 {
		extra[stats.AttackPower] += debuffs.GetExposeWeaknessHunterAgility() * 0.25
		extra[stats.RangedAttackPower] += debuffs.GetExposeWeaknessHunterAgility() * 0.25
	}
	switch debuffs.GetHuntersMark() {
	case proto.TristateEffect_TristateEffectImproved:
		extra[stats.AttackPower] += 110
		extra[stats.RangedAttackPower] += 440
	case proto.TristateEffect_TristateEffectMissing:
	default:
		extra[stats.RangedAttackPower] += 440
	}
	return extra, pseudo
}

func EnrichArmory(imported *ImportedSettings, catalog *Catalog) (*ArmoryData, error) {
	if imported == nil || imported.Settings == nil || imported.Settings.Player == nil || imported.Settings.Encounter == nil {
		return nil, fmt.Errorf("imported settings are incomplete")
	}
	if catalog == nil {
		return nil, fmt.Errorf("catalog is nil")
	}

	result := core.ComputeStats(armoryComputeRequest(imported))
	if result == nil {
		return nil, fmt.Errorf("compute stats returned no result")
	}
	if result.GetErrorResult() != "" {
		return nil, fmt.Errorf("compute stats failed: %s", result.GetErrorResult())
	}
	raidStats := result.GetRaidStats()
	if raidStats == nil || len(raidStats.GetParties()) != 1 {
		return nil, fmt.Errorf("compute stats did not return exactly one player gear snapshot")
	}
	party := raidStats.GetParties()[0]
	if party == nil {
		return nil, fmt.Errorf("compute stats did not return exactly one player gear snapshot")
	}
	var playerStats *proto.PlayerStats
	for _, candidate := range party.GetPlayers() {
		if candidate == nil {
			continue
		}
		if playerStats != nil {
			return nil, fmt.Errorf("compute stats did not return exactly one player gear snapshot")
		}
		playerStats = candidate
	}
	if playerStats == nil || playerStats.GetFinalStats() == nil {
		return nil, fmt.Errorf("compute stats did not return exactly one player gear snapshot")
	}
	finalStats := playerStats.GetFinalStats()
	extra, pseudoDebuffs := panelDebuffStats(imported.Settings.GetDebuffs())
	final := stats.FromProtoArray(finalStats.GetStats()).Add(extra)
	pseudo := append([]float64(nil), finalStats.GetPseudoStats()...)
	for pseudoStat, value := range pseudoDebuffs {
		pseudo[int(pseudoStat)] += value
	}
	armory := &ArmoryData{Gear: make([]GearSlotData, 0, len(canonicalGearSlots)), Stats: statValuesMap(final), DerivedStats: make(map[string]float64)}
	for _, entry := range []struct {
		key    string
		pseudo proto.PseudoStat
	}{
		{"melee_hit_percent", proto.PseudoStat_PseudoStatMeleeHitPercent},
		{"spell_hit_percent", proto.PseudoStat_PseudoStatSpellHitPercent},
		{"melee_crit_percent", proto.PseudoStat_PseudoStatMeleeCritPercent},
		{"spell_crit_percent", proto.PseudoStat_PseudoStatSpellCritPercent},
		{"ranged_hit_percent", proto.PseudoStat_PseudoStatRangedHitPercent},
		{"ranged_crit_percent", proto.PseudoStat_PseudoStatRangedCritPercent},
		{"block_percent", proto.PseudoStat_PseudoStatBlockPercent},
	} {
		index := int(entry.pseudo)
		if index >= len(pseudo) {
			return nil, fmt.Errorf("compute stats returned malformed pseudo-stat snapshot")
		}
		armory.DerivedStats[entry.key] = pseudo[index]
	}

	equipment := imported.Settings.Player.GetEquipment()
	for _, slot := range canonicalGearSlots {
		data := GearSlotData{Slot: slot.slot, SlotName: slot.name, Stats: make(map[string]float64), Sockets: []SocketData{}, SocketBonus: SocketBonusData{Stats: make(map[string]float64)}}
		if equipment != nil && slot.index < len(equipment.GetItems()) {
			spec := equipment.GetItems()[slot.index]
			if spec != nil && spec.GetId() != 0 {
				item := catalog.Items[spec.GetId()]
				if item == nil {
					return nil, fmt.Errorf("item %d not found in catalog", spec.GetId())
				}
				var err error
				data, err = enrichItem(spec, item, catalog)
				if err != nil {
					return nil, err
				}
				data.Slot = slot.slot
				data.SlotName = slot.name
			}
		}
		armory.Gear = append(armory.Gear, data)
	}
	return armory, nil
}
