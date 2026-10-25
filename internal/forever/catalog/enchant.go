package catalog

import (
	"fmt"

	"github.com/wowsims/tbc/internal/forever/domain"
)

// Enchant resolution errors. ErrIncompatibleSlot wraps ErrNotFound so callers
// that only distinguish "missing" from "ambiguous" still behave.
var (
	ErrNotFound         = domain.ErrNotFound
	ErrAmbiguous        = domain.ErrAmbiguous
	ErrIncompatibleSlot = fmt.Errorf("%w: enchant does not apply to this slot", domain.ErrNotFound)
)

// ResolveEnchant resolves an effect on a slot. When more than one row remains,
// the optional instance's item_id and then spell_id disambiguate it. It never
// guesses.
func ResolveEnchant(cat *Catalog, effectID int, slot domain.Slot, instance ...*domain.EnchantInstance) (*domain.Enchant, error) {
	rows := cat.EnchantsByEffect[effectID]
	if len(rows) == 0 {
		return nil, ErrNotFound
	}
	types := SlotItemTypes(slot)
	candidates := make([]*domain.Enchant, 0, len(rows))
	for _, row := range rows {
		if slot == "" || enchantAdmitsTypes(row, types) {
			candidates = append(candidates, row)
		}
	}
	if len(candidates) == 0 {
		return nil, ErrIncompatibleSlot
	}
	if len(candidates) > 1 && len(instance) > 0 && instance[0] != nil {
		hint := instance[0]
		if hint.ItemID != 0 {
			if filtered := filterEnchants(candidates, func(e *domain.Enchant) bool {
				return e.ItemID == hint.ItemID
			}); len(filtered) > 0 {
				candidates = filtered
			}
		}
		if len(candidates) > 1 && hint.SpellID != 0 {
			if filtered := filterEnchants(candidates, func(e *domain.Enchant) bool {
				return e.SpellID == hint.SpellID
			}); len(filtered) > 0 {
				candidates = filtered
			}
		}
	}
	if len(candidates) > 1 {
		return nil, ErrAmbiguous
	}
	return candidates[0], nil
}

func filterEnchants(enchantments []*domain.Enchant, keep func(*domain.Enchant) bool) []*domain.Enchant {
	filtered := make([]*domain.Enchant, 0, len(enchantments))
	for _, enchant := range enchantments {
		if keep(enchant) {
			filtered = append(filtered, enchant)
		}
	}
	return filtered
}
