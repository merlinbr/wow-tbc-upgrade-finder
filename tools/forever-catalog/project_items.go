package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/wowsims/tbc/internal/forever/domain"
)

type rawDatabase struct {
	Items          []rawItem    `json:"items"`
	Enchants       []rawEnchant `json:"enchants"`
	RandomSuffixes []rawSuffix  `json:"randomSuffixes"`
}

type rawScaling struct {
	RandPropPoints  int            `json:"randPropPoints"`
	Stats           map[string]int `json:"stats"`
	Ilvl            int            `json:"ilvl"`
	WeaponDamageMin int            `json:"weaponDamageMin"`
	WeaponDamageMax int            `json:"weaponDamageMax"`
	UnroundedDps    float64        `json:"unroundedDps"`
}

type rawItem struct {
	ID                  int                          `json:"id"`
	Name                string                       `json:"name"`
	Icon                string                       `json:"icon"`
	Type                int                          `json:"type"`
	ArmorType           int                          `json:"armorType"`
	WeaponType          int                          `json:"weaponType"`
	HandType            int                          `json:"handType"`
	RangedWeaponType    int                          `json:"rangedWeaponType"`
	Quality             int                          `json:"quality"`
	Phase               int                          `json:"phase"`
	Expansion           int                          `json:"expansion"`
	ScalingOptions      map[string]rawScaling        `json:"scalingOptions"`
	Stats               map[string]int               `json:"stats"`
	Ilvl                int                          `json:"ilvl"`
	WeaponSpeed         float64                      `json:"weaponSpeed"`
	RandomSuffixOptions []int                        `json:"randomSuffixOptions"`
	ClassAllowlist      []int                        `json:"classAllowlist"`
	RequiredProfession  int                          `json:"requiredProfession"`
	FactionRestriction  int                          `json:"factionRestriction"`
	SetID               int                          `json:"setId"`
	SetName             string                       `json:"setName"`
	NameDescription     string                       `json:"nameDescription"`
	Unique              bool                         `json:"unique"`
	LimitCategory       int                          `json:"limitCategory"`
	Sources             []map[string]json.RawMessage `json:"sources"`
}

type rawEnchant struct {
	EffectID           int    `json:"effectId"`
	ItemID             int    `json:"itemId"`
	SpellID            int    `json:"spellId"`
	Name               string `json:"name"`
	Icon               string `json:"icon"`
	Type               int    `json:"type"`
	ExtraTypes         []int  `json:"extraTypes"`
	Stats              []int  `json:"stats"`
	PseudoStats        []int  `json:"pseudoStats"`
	Quality            int    `json:"quality"`
	Lvl                int    `json:"lvl"`
	ClassAllowlist     []int  `json:"classAllowlist"`
	RequiredProfession int    `json:"requiredProfession"`
	EnchantType        int    `json:"enchantType"`
	ProcKind           int    `json:"procKind"`
	WeaponDamage       int    `json:"weaponDamage"`
}

type rawSuffix struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Stats []int  `json:"stats"`
}

// RawCounts records the source row counts before any eligibility filter.
type RawCounts struct {
	ItemsTotal       int
	ItemsLeftover    int
	Enchants         int
	LeftoverEnchants int
	Suffixes         int
}

// CountRaw parses both payloads and returns their row counts.
func CountRaw(raw RawPayload) (RawCounts, error) {
	var main rawDatabase
	if err := json.Unmarshal(raw.DB, &main); err != nil {
		return RawCounts{}, fmt.Errorf("parse db.json: %w", err)
	}
	counts := RawCounts{
		ItemsTotal:    len(main.Items),
		Enchants:      len(main.Enchants),
		Suffixes:      len(main.RandomSuffixes),
		ItemsLeftover: 0,
	}
	if len(raw.LeftoverDB) > 0 {
		var leftover rawDatabase
		if err := json.Unmarshal(raw.LeftoverDB, &leftover); err != nil {
			return RawCounts{}, fmt.Errorf("parse leftover_db.json: %w", err)
		}
		counts.ItemsLeftover = len(leftover.Items)
		counts.LeftoverEnchants = len(leftover.Enchants)
	}
	return counts, nil
}

// ProjectItems normalizes the raw upstream rows into the committed catalog.
func ProjectItems(raw RawPayload, enums Enums) (domain.Catalog, error) {
	var main rawDatabase
	if err := json.Unmarshal(raw.DB, &main); err != nil {
		return domain.Catalog{}, fmt.Errorf("parse db.json: %w", err)
	}

	cat := domain.Catalog{Revision: raw.Revision()}
	for _, row := range main.Items {
		item := projectItem(row, enums)
		// Shirts, tabards and other non-statistical rows have no item type.
		if item.TypeID == 0 {
			continue
		}
		cat.Items = append(cat.Items, item)
	}
	for _, row := range main.Enchants {
		cat.Enchants = append(cat.Enchants, projectEnchant(row, enums))
	}
	for _, row := range main.RandomSuffixes {
		cat.Suffixes = append(cat.Suffixes, projectSuffix(row, enums))
	}

	if len(raw.LeftoverDB) > 0 {
		var leftover rawDatabase
		if err := json.Unmarshal(raw.LeftoverDB, &leftover); err != nil {
			return domain.Catalog{}, fmt.Errorf("parse leftover_db.json: %w", err)
		}
		for _, row := range leftover.Items {
			cat.LeftoverItems = append(cat.LeftoverItems, projectItem(row, enums))
		}
		for _, row := range leftover.Enchants {
			cat.LeftoverEnchants = append(cat.LeftoverEnchants, projectEnchant(row, enums))
		}
	}
	return cat, nil
}

func projectItem(row rawItem, enums Enums) domain.Item {
	item := domain.Item{
		ID:              row.ID,
		Name:            row.Name,
		Icon:            row.Icon,
		Type:            enums.ItemTypeName(row.Type),
		TypeID:          row.Type,
		Quality:         enums.QualityName(row.Quality),
		QualityID:       row.Quality,
		Phase:           row.Phase,
		Expansion:       row.Expansion,
		WeaponSpeed:     row.WeaponSpeed,
		Unique:          row.Unique,
		LimitCategory:   row.LimitCategory,
		SetID:           row.SetID,
		SetName:         row.SetName,
		NameDescription: row.NameDescription,
		SuffixIDs:       row.RandomSuffixOptions,
	}
	if row.ArmorType != 0 {
		item.ArmorType = enums.ArmorName(row.ArmorType)
		item.ArmorTypeID = row.ArmorType
	}
	if row.WeaponType != 0 {
		item.WeaponType = enums.WeaponName(row.WeaponType)
		item.WeaponTypeID = row.WeaponType
	}
	if row.HandType != 0 {
		item.HandType = enums.HandName(row.HandType)
		item.HandTypeID = row.HandType
	}
	if row.RangedWeaponType != 0 {
		item.RangedWeaponType = enums.RangedName(row.RangedWeaponType)
		item.RangedWeaponTypeID = row.RangedWeaponType
	}
	for _, class := range row.ClassAllowlist {
		item.ClassAllowlistIDs = append(item.ClassAllowlistIDs, class)
		item.ClassAllowlist = append(item.ClassAllowlist, enums.ClassName(class))
	}
	if row.RequiredProfession != 0 {
		item.RequiresProfession = row.RequiredProfession
		item.RequiredProfession = enums.ProfessionName(row.RequiredProfession)
	}
	if row.FactionRestriction != 0 {
		item.FactionRestrictionID = row.FactionRestriction
		item.FactionRestriction = enums.FactionName(row.FactionRestriction)
	}

	// Stats mostly live inside scalingOptions; the default variant (key "0")
	// is the flat view. Never substitute zeros for a missing variant.
	if def, ok := row.ScalingOptions["0"]; ok {
		item.Ilvl = def.Ilvl
		item.Stats = statsFromMap(def.Stats, enums)
		item.WeaponDamageMin = def.WeaponDamageMin
		item.WeaponDamageMax = def.WeaponDamageMax
	}
	if item.Ilvl == 0 {
		item.Ilvl = row.Ilvl
	}
	if len(item.Stats) == 0 {
		item.Stats = statsFromMap(row.Stats, enums)
	}

	for _, key := range sortedScalingKeys(row.ScalingOptions) {
		if key == "0" {
			continue
		}
		variantRow := row.ScalingOptions[key]
		ilvl := variantRow.Ilvl
		if ilvl == 0 {
			if parsed, err := strconv.Atoi(key); err == nil {
				ilvl = parsed
			}
		}
		item.Variants = append(item.Variants, domain.Variant{
			Ilvl:            ilvl,
			Stats:           statsFromMap(variantRow.Stats, enums),
			WeaponDamageMin: variantRow.WeaponDamageMin,
			WeaponDamageMax: variantRow.WeaponDamageMax,
		})
	}
	sort.Slice(item.Variants, func(i, j int) bool { return item.Variants[i].Ilvl < item.Variants[j].Ilvl })

	item.Slots = deriveSlots(row, enums)
	item.Sources = projectSources(row.Sources)
	return item
}

func projectEnchant(row rawEnchant, enums Enums) domain.Enchant {
	enchant := domain.Enchant{
		EffectID:      row.EffectID,
		ItemID:        row.ItemID,
		SpellID:       row.SpellID,
		Name:          row.Name,
		Icon:          row.Icon,
		Type:          enums.ItemTypeName(row.Type),
		TypeID:        row.Type,
		Quality:       enums.QualityName(row.Quality),
		QualityID:     row.Quality,
		Lvl:           row.Lvl,
		EnchantType:   enums.EnchantTypeName(row.EnchantType),
		EnchantTypeID: row.EnchantType,
		ProcKind:      enums.ProcKindName(row.ProcKind),
		ProcKindID:    row.ProcKind,
		WeaponDamage:  row.WeaponDamage,
	}
	for _, t := range row.ExtraTypes {
		enchant.ExtraTypeIDs = append(enchant.ExtraTypeIDs, t)
		enchant.ExtraTypes = append(enchant.ExtraTypes, enums.ItemTypeName(t))
	}
	enchant.Stats = statsFromArray(row.Stats, enums.StatName)
	enchant.PseudoStats = statsFromArray(row.PseudoStats, enums.PseudoStatName)
	for _, class := range row.ClassAllowlist {
		enchant.ClassAllowlistIDs = append(enchant.ClassAllowlistIDs, class)
		enchant.ClassAllowlist = append(enchant.ClassAllowlist, enums.ClassName(class))
	}
	if row.RequiredProfession != 0 {
		enchant.RequiredProfessionID = row.RequiredProfession
		enchant.RequiredProfession = enums.ProfessionName(row.RequiredProfession)
	}
	return enchant
}

func projectSuffix(row rawSuffix, enums Enums) domain.RandomSuffix {
	return domain.RandomSuffix{
		ID:    row.ID,
		Name:  row.Name,
		Stats: statsFromArray(row.Stats, enums.StatName),
	}
}

func deriveSlots(row rawItem, enums Enums) []domain.Slot {
	switch enums.ItemTypeName(row.Type) {
	case "head":
		return []domain.Slot{domain.SlotHead}
	case "neck":
		return []domain.Slot{domain.SlotNeck}
	case "shoulder":
		return []domain.Slot{domain.SlotShoulder}
	case "back":
		return []domain.Slot{domain.SlotBack}
	case "chest":
		return []domain.Slot{domain.SlotChest}
	case "wrist":
		return []domain.Slot{domain.SlotWrist}
	case "hands":
		return []domain.Slot{domain.SlotHands}
	case "waist":
		return []domain.Slot{domain.SlotWaist}
	case "legs":
		return []domain.Slot{domain.SlotLegs}
	case "feet":
		return []domain.Slot{domain.SlotFeet}
	case "finger":
		return []domain.Slot{domain.SlotFinger1, domain.SlotFinger2}
	case "trinket":
		return []domain.Slot{domain.SlotTrinket1, domain.SlotTrinket2}
	case "ranged":
		return []domain.Slot{domain.SlotRanged}
	case "weapon":
		weapon := enums.WeaponName(row.WeaponType)
		hand := enums.HandName(row.HandType)
		if weapon == "shield" || weapon == "off_hand" || hand == "off_hand" {
			return []domain.Slot{domain.SlotOffHand}
		}
		if hand == "one_hand" {
			return []domain.Slot{domain.SlotMainHand, domain.SlotOffHand}
		}
		return []domain.Slot{domain.SlotMainHand}
	default:
		return nil
	}
}

func projectSources(rows []map[string]json.RawMessage) []domain.ItemSource {
	var sources []domain.ItemSource
	for _, row := range rows {
		for kind, raw := range row {
			source := domain.ItemSource{Kind: kind}
			var detail map[string]any
			if err := json.Unmarshal(raw, &detail); err == nil {
				source.Detail = detail
			}
			sources = append(sources, source)
		}
	}
	return sources
}

func statsFromMap(values map[string]int, enums Enums) []domain.Stat {
	if len(values) == 0 {
		return nil
	}
	keys := make([]int, 0, len(values))
	for key := range values {
		index, err := strconv.Atoi(key)
		if err != nil {
			continue
		}
		keys = append(keys, index)
	}
	sort.Ints(keys)
	stats := make([]domain.Stat, 0, len(keys))
	for _, index := range keys {
		stats = append(stats, domain.Stat{
			Name:  enums.StatName(index),
			Value: values[strconv.Itoa(index)],
		})
	}
	return stats
}

func statsFromArray(values []int, name func(int) string) []domain.Stat {
	var stats []domain.Stat
	for index, value := range values {
		if value == 0 {
			continue
		}
		stats = append(stats, domain.Stat{Name: name(index), Value: value})
	}
	return stats
}

func sortedScalingKeys(options map[string]rawScaling) []string {
	keys := make([]string, 0, len(options))
	for key := range options {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		left, leftErr := strconv.Atoi(keys[i])
		right, rightErr := strconv.Atoi(keys[j])
		if leftErr == nil && rightErr == nil {
			return left < right
		}
		return keys[i] < keys[j]
	})
	return keys
}
