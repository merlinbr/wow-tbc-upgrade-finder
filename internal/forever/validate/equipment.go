// Package validate applies the pinned Forever validation rules.
package validate

import (
	"fmt"
	"slices"
	"strings"

	"github.com/wowsims/tbc/internal/forever/catalog"
	"github.com/wowsims/tbc/internal/forever/domain"
)

// Violation is one typed validation finding.
type Violation struct {
	Code     string      `json:"code"`
	Severity string      `json:"severity"`
	Message  string      `json:"message"`
	Slot     domain.Slot `json:"slot,omitempty"`
}

// Equipment validates a character's equipment against the pinned catalog.
func Equipment(ch domain.Character, eq domain.Equipment, cat *catalog.Catalog) []Violation {
	var violations []Violation

	// Race/class availability is a character-level rule: an empty build on an
	// invalid combination is still invalid.
	raceMismatch := false
	if ch.Class != "" && ch.Race != "" {
		if races, ok := cat.Capabilities.Races[ch.Class]; ok && !slices.Contains(races, ch.Race) {
			raceMismatch = true
			violations = append(violations, Violation{
				Code:     "race_class_mismatch",
				Severity: "error",
				Message: fmt.Sprintf(
					"%s is not an available race for %s.", displayName(ch.Race), displayName(ch.Class),
				),
			})
		}
	}

	itemCounts := make(map[int]int)
	limitCounts := make(map[int]int)
	for _, slot := range domain.SlotOrder {
		instance, ok := eq[slot]
		if !ok {
			continue
		}
		item := cat.Items[instance.ItemID]
		if item == nil {
			continue
		}
		itemCounts[item.ID]++
		if item.LimitCategory != 0 {
			limitCounts[item.LimitCategory]++
		}
	}

	for _, slot := range domain.SlotOrder {
		instance, ok := eq[slot]
		if !ok {
			continue
		}
		label := domain.SlotLabel(slot)
		item := cat.Items[instance.ItemID]
		if item == nil {
			violations = append(violations, Violation{
				Code:     "unknown_item",
				Severity: "error",
				Message: fmt.Sprintf(
					"Item %d in %s is not in the pinned catalog.", instance.ItemID, label,
				),
				Slot: slot,
			})
			continue
		}

		if allowed, reason := cat.Compatibility(ch, item, slot); !allowed {
			// The character-level check already reported a race/class mismatch.
			if reason != "race_class_mismatch" || !raceMismatch {
				violations = append(violations, Violation{
					Code:     reason,
					Severity: "error",
					Message:  fmt.Sprintf("%s cannot be equipped in %s (%s).", item.Name, label, reason),
					Slot:     slot,
				})
			}
		}

		if item.Unique && itemCounts[item.ID] > 1 {
			violations = append(violations, Violation{
				Code:     "unique_conflict",
				Severity: "error",
				Message: fmt.Sprintf(
					"%s is unique but is equipped %d times.", item.Name, itemCounts[item.ID],
				),
				Slot: slot,
			})
		}
		if item.LimitCategory != 0 && limitCounts[item.LimitCategory] > 1 {
			violations = append(violations, Violation{
				Code:     "limit_category_conflict",
				Severity: "error",
				Message: fmt.Sprintf(
					"%s shares limit category %d with another equipped item.",
					item.Name, item.LimitCategory,
				),
				Slot: slot,
			})
		}
		if slot == domain.SlotOffHand && item.HandType == "one_hand" &&
			!cat.Capabilities.DualWield[ch.Spec] {
			violations = append(violations, Violation{
				Code:     "dual_wield_unavailable",
				Severity: "error",
				Message: fmt.Sprintf(
					"%s is a one-handed off-hand weapon, but %s cannot dual-wield.",
					item.Name, displaySpec(ch),
				),
				Slot: slot,
			})
		}
	}

	if mainHand, ok := eq[domain.SlotMainHand]; ok {
		if _, offHandOccupied := eq[domain.SlotOffHand]; offHandOccupied {
			if item := cat.Items[mainHand.ItemID]; item != nil && item.HandType == "two_hand" {
				violations = append(violations, Violation{
					Code:     "two_hand_conflict",
					Severity: "error",
					Message: fmt.Sprintf(
						"%s is two-handed and cannot be used with an off-hand item.", item.Name,
					),
					Slot: domain.SlotOffHand,
				})
			}
		}
	}

	return violations
}

func displayName(value string) string {
	if value == "" {
		return ""
	}
	spaced := strings.ReplaceAll(value, "_", " ")
	return strings.ToUpper(spaced[:1]) + spaced[1:]
}

func displaySpec(ch domain.Character) string {
	if ch.Spec == "" {
		return displayName(ch.Class)
	}
	return displayName(ch.Spec)
}
