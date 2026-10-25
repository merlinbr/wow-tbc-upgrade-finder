package domain

import "time"

// Character is the shared identity of every build. Class, Spec and Race are
// catalog keys from the pinned capabilities data.
type Character struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Class       Class     `json:"class"`
	Spec        Spec      `json:"spec"`
	Race        Race      `json:"race"`
	Level       int       `json:"level"`
	Professions []string  `json:"professions,omitempty"`
	Region      string    `json:"region,omitempty"`
	Realm       string    `json:"realm,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	UpdatedAt   time.Time `json:"updatedAt,omitempty"`
}

// EnchantInstance identifies an applied enchant by effect plus optional
// item/spell cross-references. The slot is the equipment slot it sits on.
type EnchantInstance struct {
	EffectID int `json:"effectId"`
	ItemID   int `json:"itemId,omitempty"`
	SpellID  int `json:"spellId,omitempty"`
}

// ItemInstance is an equipped item, not a bare catalog ID.
type ItemInstance struct {
	ItemID         int              `json:"itemId"`
	VariantIlvl    int              `json:"variantIlvl,omitempty"`
	RandomSuffixID int              `json:"randomSuffixId,omitempty"`
	Enchant        *EnchantInstance `json:"enchant,omitempty"`
}

// Equipment is the seventeen-slot map; absent slots are empty.
type Equipment map[Slot]ItemInstance

// Talents allocates points by talent identity, never by position.
type Talents map[string]int

// Build is one Current or Plan record.
type Build struct {
	ID              string    `json:"id"`
	CharacterID     string    `json:"characterId"`
	Kind            string    `json:"kind"`
	Name            string    `json:"name,omitempty"`
	CatalogRevision string    `json:"catalogRevision"`
	Equipment       Equipment `json:"equipment"`
	Talents         Talents   `json:"talents"`
	Source          string    `json:"source"`
	Version         int       `json:"version"`
	Published       bool      `json:"published"`
	CreatedAt       time.Time `json:"createdAt,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt,omitempty"`
}
