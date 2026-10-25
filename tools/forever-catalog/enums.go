package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// requiredEnums are the enum tables the projection and capabilities parser
// depend on. A missing table is a hard generator error: silently fall back to
// numeric keys would let TBC-era assumptions leak in.
var requiredEnums = []string{
	"Stat",
	"PseudoStat",
	"ItemType",
	"ArmorType",
	"WeaponType",
	"HandType",
	"RangedWeaponType",
	"Class",
	"Race",
	"ItemSlot",
	"ItemQuality",
	"Profession",
	"Faction",
	"EnchantType",
	"EnchantProcKind",
	"Spec",
}

// Enums holds the pinned enum tables parsed from proto/common.proto.
type Enums struct {
	byValue map[string]map[int]string
	byKey   map[string]map[string]int
}

var (
	enumBlockRe  = regexp.MustCompile(`(?ms)enum\s+([A-Za-z_][A-Za-z0-9_]*)\s*\{(.*?)\}`)
	enumMemberRe = regexp.MustCompile(`(?m)^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(-?[0-9]+)\s*;`)
	acronymRe    = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
	wordRe       = regexp.MustCompile(`([a-z0-9])([A-Z])`)
	nonWordRunRe = regexp.MustCompile(`[^A-Za-z0-9]+`)
)

// ParseEnums extracts the required enum tables from the pinned proto text.
func ParseEnums(protoText string) (Enums, error) {
	e := Enums{
		byValue: make(map[string]map[int]string),
		byKey:   make(map[string]map[string]int),
	}
	for _, block := range enumBlockRe.FindAllStringSubmatch(protoText, -1) {
		name, body := block[1], block[2]
		values := make(map[int]string)
		keys := make(map[string]int)
		for _, member := range enumMemberRe.FindAllStringSubmatch(body, -1) {
			memberName, rawValue := member[1], member[2]
			value, err := strconv.Atoi(rawValue)
			if err != nil {
				return Enums{}, fmt.Errorf("parse enum %s member %s: %w", name, memberName, err)
			}
			key := enumKey(name, memberName)
			if key == "" {
				continue
			}
			values[value] = key
			keys[key] = value
		}
		e.byValue[name] = values
		e.byKey[name] = keys
	}
	for _, name := range requiredEnums {
		values, ok := e.byValue[name]
		if !ok || len(values) == 0 {
			return Enums{}, fmt.Errorf("pinned proto is missing enum %s", name)
		}
		// Reject a table whose members all stripped to an empty key: that
		// would silently break every projection downstream.
		if len(e.byKey[name]) == 0 {
			return Enums{}, fmt.Errorf("pinned enum %s has no usable members", name)
		}
	}
	return e, nil
}

// enumKey strips the enum-name prefix and converts the member to snake_case.
func enumKey(enumName, memberName string) string {
	name := memberName
	if strings.HasPrefix(name, enumName) {
		name = strings.TrimPrefix(name, enumName)
	}
	return snakeCase(name)
}

func snakeCase(s string) string {
	s = acronymRe.ReplaceAllString(s, "${1}_${2}")
	s = wordRe.ReplaceAllString(s, "${1}_${2}")
	s = nonWordRunRe.ReplaceAllString(s, "_")
	return strings.ToLower(strings.Trim(s, "_"))
}

// Name resolves a numeric enum value to its catalog key.
func (e Enums) Name(enumName string, value int) (string, bool) {
	values, ok := e.byValue[enumName]
	if !ok {
		return "", false
	}
	name, ok := values[value]
	return name, ok
}

// Value resolves a catalog enum key to its numeric value.
func (e Enums) Value(enumName, key string) (int, bool) {
	keys, ok := e.byKey[enumName]
	if !ok {
		return 0, false
	}
	value, ok := keys[key]
	return value, ok
}

// Keys returns every catalog key of an enum, sorted for deterministic output.
func (e Enums) Keys(enumName string) []string {
	values, ok := e.byValue[enumName]
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(values))
	for _, key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (e Enums) fallback(prefix string, value int) string {
	return fmt.Sprintf("%s_%d", prefix, value)
}

func (e Enums) StatName(v int) string {
	if name, ok := e.Name("Stat", v); ok {
		return name
	}
	return e.fallback("stat", v)
}

func (e Enums) PseudoStatName(v int) string {
	if name, ok := e.Name("PseudoStat", v); ok {
		return name
	}
	return e.fallback("pseudo_stat", v)
}

func (e Enums) SlotName(v int) string {
	if name, ok := e.Name("ItemSlot", v); ok {
		return name
	}
	return e.fallback("slot", v)
}

func (e Enums) ItemTypeName(v int) string {
	if name, ok := e.Name("ItemType", v); ok {
		return name
	}
	return e.fallback("item_type", v)
}

func (e Enums) ArmorName(v int) string {
	if name, ok := e.Name("ArmorType", v); ok {
		return name
	}
	return e.fallback("armor_type", v)
}

func (e Enums) WeaponName(v int) string {
	if name, ok := e.Name("WeaponType", v); ok {
		return name
	}
	return e.fallback("weapon_type", v)
}

func (e Enums) HandName(v int) string {
	if name, ok := e.Name("HandType", v); ok {
		return name
	}
	return e.fallback("hand_type", v)
}

func (e Enums) RangedName(v int) string {
	if name, ok := e.Name("RangedWeaponType", v); ok {
		return name
	}
	return e.fallback("ranged_weapon_type", v)
}

func (e Enums) ClassName(v int) string {
	if name, ok := e.Name("Class", v); ok {
		return name
	}
	return e.fallback("class", v)
}

func (e Enums) RaceName(v int) string {
	if name, ok := e.Name("Race", v); ok {
		return name
	}
	return e.fallback("race", v)
}

func (e Enums) QualityName(v int) string {
	if name, ok := e.Name("ItemQuality", v); ok {
		return name
	}
	return e.fallback("quality", v)
}

func (e Enums) ProfessionName(v int) string {
	if name, ok := e.Name("Profession", v); ok {
		return name
	}
	return e.fallback("profession", v)
}

func (e Enums) FactionName(v int) string {
	if name, ok := e.Name("Faction", v); ok {
		return name
	}
	return e.fallback("faction", v)
}

func (e Enums) EnchantTypeName(v int) string {
	if name, ok := e.Name("EnchantType", v); ok {
		return name
	}
	return e.fallback("enchant_type", v)
}

func (e Enums) ProcKindName(v int) string {
	if name, ok := e.Name("EnchantProcKind", v); ok {
		return name
	}
	return e.fallback("proc_kind", v)
}
