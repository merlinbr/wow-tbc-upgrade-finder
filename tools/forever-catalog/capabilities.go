package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/wowsims/tbc/internal/forever/domain"
)

var (
	tsClassEntryRe = regexp.MustCompile(`(?s)\[Class\.(\w+)\]:\s*\[(.*?)\]`)
	armorMemberRe  = regexp.MustCompile(`ArmorType\.(\w+)`)
	weaponMemberRe = regexp.MustCompile(`(?s)\{\s*weaponType:\s*WeaponType\.(\w+)\s*,\s*canUseTwoHand:\s*(true|false)\s*\}`)
	rangedMemberRe = regexp.MustCompile(`RangedWeaponType\.(\w+)`)
	raceMemberRe   = regexp.MustCompile(`Race\.(\w+)`)
	specMemberRe   = regexp.MustCompile(`Spec\.(\w+)`)
	dualWieldSetRe = regexp.MustCompile(`(?s)SPEC_CAN_DUAL_WIELD\s*=\s*new\s+Set<Spec>\(\s*\[(.*?)\]\s*\)`)
)

// dualWieldClassBySpec maps the pinned upstream Spec members onto the class
// whose every spec can dual-wield. Enhancement shaman is deliberately absent:
// the pinned set omits it and the contract leaves it open.
var dualWieldClassBySpec = map[string]domain.Class{
	"hunter":             "hunter",
	"rogue":              "rogue",
	"dps_warrior":        "warrior",
	"protection_warrior": "warrior",
}

// ParseCapabilities reads the four generated capability tables and the
// spec-level dual-wield set. Every member must translate to a pinned enum.
func ParseCapabilities(capsTS string, enums Enums, specs map[domain.Class][]domain.SpecInfo) (domain.Capabilities, error) {
	caps := domain.Capabilities{
		Specs:     specs,
		Armor:     make(map[domain.Class][]string),
		Weapons:   make(map[domain.Class][]domain.WeaponEligibility),
		Ranged:    make(map[domain.Class][]string),
		Races:     make(map[domain.Class][]string),
		DualWield: make(map[domain.Spec]bool),
	}

	armorBody, err := tsBlock(capsTS, "CLASS_ARMOR_TYPES")
	if err != nil {
		return domain.Capabilities{}, err
	}
	for _, entry := range tsClassEntries(armorBody) {
		class, err := capabilityKey(enums, "Class", entry.Member)
		if err != nil {
			return domain.Capabilities{}, err
		}
		for _, match := range armorMemberRe.FindAllStringSubmatch(entry.Body, -1) {
			key, err := capabilityKey(enums, "ArmorType", match[1])
			if err != nil {
				return domain.Capabilities{}, err
			}
			caps.Armor[class] = append(caps.Armor[class], key)
		}
	}

	weaponBody, err := tsBlock(capsTS, "CLASS_WEAPON_TYPES")
	if err != nil {
		return domain.Capabilities{}, err
	}
	for _, entry := range tsClassEntries(weaponBody) {
		class, err := capabilityKey(enums, "Class", entry.Member)
		if err != nil {
			return domain.Capabilities{}, err
		}
		for _, match := range weaponMemberRe.FindAllStringSubmatch(entry.Body, -1) {
			key, err := capabilityKey(enums, "WeaponType", match[1])
			if err != nil {
				return domain.Capabilities{}, err
			}
			caps.Weapons[class] = append(caps.Weapons[class], domain.WeaponEligibility{
				WeaponType:    key,
				CanUseTwoHand: match[2] == "true",
			})
		}
	}

	rangedBody, err := tsBlock(capsTS, "CLASS_RANGED_WEAPON_TYPES")
	if err != nil {
		return domain.Capabilities{}, err
	}
	for _, entry := range tsClassEntries(rangedBody) {
		class, err := capabilityKey(enums, "Class", entry.Member)
		if err != nil {
			return domain.Capabilities{}, err
		}
		for _, match := range rangedMemberRe.FindAllStringSubmatch(entry.Body, -1) {
			key, err := capabilityKey(enums, "RangedWeaponType", match[1])
			if err != nil {
				return domain.Capabilities{}, err
			}
			caps.Ranged[class] = append(caps.Ranged[class], key)
		}
	}

	raceBody, err := tsBlock(capsTS, "CLASS_RACES")
	if err != nil {
		return domain.Capabilities{}, err
	}
	for _, entry := range tsClassEntries(raceBody) {
		class, err := capabilityKey(enums, "Class", entry.Member)
		if err != nil {
			return domain.Capabilities{}, err
		}
		for _, match := range raceMemberRe.FindAllStringSubmatch(entry.Body, -1) {
			key, err := capabilityKey(enums, "Race", match[1])
			if err != nil {
				return domain.Capabilities{}, err
			}
			caps.Races[class] = append(caps.Races[class], key)
		}
	}

	dualWieldMatch := dualWieldSetRe.FindStringSubmatch(capsTS)
	if dualWieldMatch == nil {
		return domain.Capabilities{}, fmt.Errorf("capabilities: SPEC_CAN_DUAL_WIELD missing")
	}
	dualWieldClasses := make(map[domain.Class]bool)
	for _, match := range specMemberRe.FindAllStringSubmatch(dualWieldMatch[1], -1) {
		specKey, err := capabilityKey(enums, "Spec", match[1])
		if err != nil {
			return domain.Capabilities{}, err
		}
		class, ok := dualWieldClassBySpec[specKey]
		if !ok {
			return domain.Capabilities{}, fmt.Errorf(
				"capabilities: SPEC_CAN_DUAL_WIELD spec %q has no class mapping", specKey,
			)
		}
		dualWieldClasses[class] = true
	}
	for class := range dualWieldClasses {
		classSpecs, ok := specs[class]
		if !ok || len(classSpecs) == 0 {
			return domain.Capabilities{}, fmt.Errorf(
				"capabilities: dual-wield class %q has no specs", class,
			)
		}
		for _, spec := range classSpecs {
			caps.DualWield[spec.Key] = true
		}
	}

	for _, key := range enums.Keys("Profession") {
		if key == "unknown" {
			continue
		}
		caps.Professions = append(caps.Professions, key)
	}

	classKeys := make([]string, 0, len(specs))
	for class := range specs {
		classKeys = append(classKeys, class)
	}
	sort.Strings(classKeys)
	for _, class := range classKeys {
		caps.Classes = append(caps.Classes, domain.ClassInfo{
			Class: class,
			Name:  displayClassName(class),
			Specs: specs[class],
		})
	}
	return caps, nil
}

type tsClassEntry struct {
	Member string
	Body   string
}

func tsClassEntries(body string) []tsClassEntry {
	var entries []tsClassEntry
	for _, match := range tsClassEntryRe.FindAllStringSubmatch(body, -1) {
		entries = append(entries, tsClassEntry{Member: match[1], Body: match[2]})
	}
	return entries
}

func tsBlock(ts, name string) (string, error) {
	pattern := regexp.MustCompile(`(?s)` + regexp.QuoteMeta(name) + `\s*(?::[^=]*)?=\s*\{(.*?)\n\};`)
	match := pattern.FindStringSubmatch(ts)
	if match == nil {
		return "", fmt.Errorf("capabilities: block %s missing", name)
	}
	return match[1], nil
}

func capabilityKey(enums Enums, enumName, qualified string) (string, error) {
	member := qualified
	if index := strings.LastIndex(member, "."); index >= 0 {
		member = member[index+1:]
	}
	key := enumKey(enumName, member)
	if key == "" {
		return "", fmt.Errorf("capabilities: empty member for enum %s", enumName)
	}
	if _, ok := enums.Value(enumName, key); !ok {
		return "", fmt.Errorf(
			"capabilities: %s member %q does not map to pinned enum %s", enumName, qualified, enumName,
		)
	}
	return key, nil
}

func displayClassName(class domain.Class) string {
	if class == "" {
		return ""
	}
	return strings.ToUpper(class[:1]) + class[1:]
}
