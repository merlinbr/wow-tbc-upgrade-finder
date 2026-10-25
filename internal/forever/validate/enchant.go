package validate

import (
	"errors"
	"fmt"
	"slices"

	"github.com/wowsims/tbc/internal/forever/catalog"
	"github.com/wowsims/tbc/internal/forever/domain"
)

// Enchant validates one applied enchant on a slot.
func Enchant(ch domain.Character, slot domain.Slot, item *domain.Item, inst *domain.EnchantInstance, cat *catalog.Catalog) []Violation {
	if inst == nil {
		return nil
	}
	label := domain.SlotLabel(slot)
	if item == nil {
		return []Violation{{
			Code:     "enchant_without_item",
			Severity: "error",
			Message:  fmt.Sprintf("An enchant cannot be applied to empty slot %s.", label),
			Slot:     slot,
		}}
	}

	resolved, err := catalog.ResolveEnchant(cat, inst.EffectID, slot, inst)
	if err != nil {
		code := "unknown_enchant"
		message := fmt.Sprintf("Enchant effect %d is not in the pinned catalog.", inst.EffectID)
		switch {
		case errors.Is(err, catalog.ErrAmbiguous):
			code = "ambiguous_enchant"
			message = fmt.Sprintf(
				"Enchant effect %d matches more than one row for %s; supply the item or spell id.",
				inst.EffectID, label,
			)
		case errors.Is(err, catalog.ErrIncompatibleSlot):
			code = "incompatible_enchant_slot"
			message = fmt.Sprintf("Enchant effect %d cannot be applied to %s.", inst.EffectID, label)
		}
		return []Violation{{Code: code, Severity: "error", Message: message, Slot: slot}}
	}

	var violations []Violation
	if len(resolved.ClassAllowlist) > 0 && !slices.Contains(resolved.ClassAllowlist, ch.Class) {
		violations = append(violations, Violation{
			Code:     "incompatible_enchant_class",
			Severity: "error",
			Message: fmt.Sprintf(
				"%s cannot be used by %s.", resolved.Name, displayName(ch.Class),
			),
			Slot: slot,
		})
	}
	if resolved.RequiredProfession != "" && !slices.Contains(ch.Professions, resolved.RequiredProfession) {
		violations = append(violations, Violation{
			Code:     "incompatible_enchant_profession",
			Severity: "error",
			Message: fmt.Sprintf(
				"%s requires %s.", resolved.Name, displayName(resolved.RequiredProfession),
			),
			Slot: slot,
		})
	}
	return violations
}
