package domain

// Slot is one of the seventeen canonical equipment slots. The order is ours;
// the mapping from the pinned upstream ItemType enum lives in the adapter.
type Slot string

const (
	SlotHead     Slot = "head"
	SlotNeck     Slot = "neck"
	SlotShoulder Slot = "shoulder"
	SlotBack     Slot = "back"
	SlotChest    Slot = "chest"
	SlotWrist    Slot = "wrist"
	SlotMainHand Slot = "main_hand"
	SlotOffHand  Slot = "off_hand"
	SlotHands    Slot = "hands"
	SlotWaist    Slot = "waist"
	SlotLegs     Slot = "legs"
	SlotFeet     Slot = "feet"
	SlotFinger1  Slot = "finger1"
	SlotFinger2  Slot = "finger2"
	SlotTrinket1 Slot = "trinket1"
	SlotTrinket2 Slot = "trinket2"
	SlotRanged   Slot = "ranged"
)

// SlotOrder is the canonical display order from the data contract.
var SlotOrder = []Slot{
	SlotHead,
	SlotNeck,
	SlotShoulder,
	SlotBack,
	SlotChest,
	SlotWrist,
	SlotMainHand,
	SlotOffHand,
	SlotHands,
	SlotWaist,
	SlotLegs,
	SlotFeet,
	SlotFinger1,
	SlotFinger2,
	SlotTrinket1,
	SlotTrinket2,
	SlotRanged,
}

// SlotLabel is the English display label for a canonical slot.
func SlotLabel(s Slot) string {
	labels := map[Slot]string{
		SlotHead:     "Head",
		SlotNeck:     "Neck",
		SlotShoulder: "Shoulder",
		SlotBack:     "Back",
		SlotChest:    "Chest",
		SlotWrist:    "Wrist",
		SlotMainHand: "Main Hand",
		SlotOffHand:  "Off Hand",
		SlotHands:    "Hands",
		SlotWaist:    "Waist",
		SlotLegs:     "Legs",
		SlotFeet:     "Feet",
		SlotFinger1:  "Finger 1",
		SlotFinger2:  "Finger 2",
		SlotTrinket1: "Trinket 1",
		SlotTrinket2: "Trinket 2",
		SlotRanged:   "Ranged",
	}
	if label, ok := labels[s]; ok {
		return label
	}
	return string(s)
}

// Class and Spec are string aliases so API payloads and catalog maps share one
// vocabulary without conversion noise.
type Class = string
type Spec = string
type Race = string

// Stat is one resolved stat value from the pinned catalog.
type Stat struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// ItemSource is one normalized UIItemSource arm.
type ItemSource struct {
	Kind   string         `json:"kind"`
	Detail map[string]any `json:"detail,omitempty"`
}

// Variant is an alternative scaling option carried by an item. The default
// variant (key "0") is flattened onto Item and is not repeated here.
type Variant struct {
	Ilvl            int    `json:"ilvl"`
	Stats           []Stat `json:"stats,omitempty"`
	WeaponDamageMin int    `json:"weaponDamageMin,omitempty"`
	WeaponDamageMax int    `json:"weaponDamageMax,omitempty"`
}

// Item is the normalized form of one upstream UIItem.
type Item struct {
	ID                   int          `json:"id"`
	Name                 string       `json:"name"`
	Icon                 string       `json:"icon,omitempty"`
	Type                 string       `json:"type"`
	TypeID               int          `json:"typeId"`
	ArmorType            string       `json:"armorType,omitempty"`
	ArmorTypeID          int          `json:"armorTypeId,omitempty"`
	WeaponType           string       `json:"weaponType,omitempty"`
	WeaponTypeID         int          `json:"weaponTypeId,omitempty"`
	HandType             string       `json:"handType,omitempty"`
	HandTypeID           int          `json:"handTypeId,omitempty"`
	RangedWeaponType     string       `json:"rangedWeaponType,omitempty"`
	RangedWeaponTypeID   int          `json:"rangedWeaponTypeId,omitempty"`
	Quality              string       `json:"quality"`
	QualityID            int          `json:"qualityId"`
	Phase                int          `json:"phase,omitempty"`
	Expansion            int          `json:"expansion,omitempty"`
	Ilvl                 int          `json:"ilvl"`
	Stats                []Stat       `json:"stats,omitempty"`
	WeaponSpeed          float64      `json:"weaponSpeed,omitempty"`
	WeaponDamageMin      int          `json:"weaponDamageMin,omitempty"`
	WeaponDamageMax      int          `json:"weaponDamageMax,omitempty"`
	Slots                []Slot       `json:"slots"`
	Unique               bool         `json:"unique,omitempty"`
	LimitCategory        int          `json:"limitCategory,omitempty"`
	ClassAllowlist       []string     `json:"classAllowlist,omitempty"`
	ClassAllowlistIDs    []int        `json:"classAllowlistIds,omitempty"`
	RequiredProfession   string       `json:"requiredProfession,omitempty"`
	RequiresProfession   int          `json:"requiredProfessionId,omitempty"`
	FactionRestriction   string       `json:"factionRestriction,omitempty"`
	FactionRestrictionID int          `json:"factionRestrictionId,omitempty"`
	SetID                int          `json:"setId,omitempty"`
	SetName              string       `json:"setName,omitempty"`
	NameDescription      string       `json:"nameDescription,omitempty"`
	SuffixIDs            []int        `json:"suffixIds,omitempty"`
	Sources              []ItemSource `json:"sources,omitempty"`
	Variants             []Variant    `json:"variants,omitempty"`
}

// Enchant is the normalized form of one upstream UIEnchant.
type Enchant struct {
	EffectID             int      `json:"effectId"`
	ItemID               int      `json:"itemId,omitempty"`
	SpellID              int      `json:"spellId,omitempty"`
	Name                 string   `json:"name"`
	Icon                 string   `json:"icon,omitempty"`
	Type                 string   `json:"type"`
	TypeID               int      `json:"typeId"`
	ExtraTypes           []string `json:"extraTypes,omitempty"`
	ExtraTypeIDs         []int    `json:"extraTypeIds,omitempty"`
	Stats                []Stat   `json:"stats,omitempty"`
	PseudoStats          []Stat   `json:"pseudoStats,omitempty"`
	Quality              string   `json:"quality"`
	QualityID            int      `json:"qualityId"`
	Lvl                  int      `json:"lvl,omitempty"`
	ClassAllowlist       []string `json:"classAllowlist,omitempty"`
	ClassAllowlistIDs    []int    `json:"classAllowlistIds,omitempty"`
	RequiredProfession   string   `json:"requiredProfession,omitempty"`
	RequiredProfessionID int      `json:"requiredProfessionId,omitempty"`
	EnchantType          string   `json:"enchantType,omitempty"`
	EnchantTypeID        int      `json:"enchantTypeId,omitempty"`
	ProcKind             string   `json:"procKind,omitempty"`
	ProcKindID           int      `json:"procKindId,omitempty"`
	WeaponDamage         int      `json:"weaponDamage,omitempty"`
}

// RandomSuffix is the normalized form of one upstream random suffix.
type RandomSuffix struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Stats []Stat `json:"stats,omitempty"`
}

// Catalog is the generator's output. Leftover rows are kept as a separate
// provenance partition; they are not part of the main picker by default.
type Catalog struct {
	Revision         string         `json:"revision"`
	Items            []Item         `json:"items"`
	Enchants         []Enchant      `json:"enchants"`
	Suffixes         []RandomSuffix `json:"suffixes"`
	LeftoverItems    []Item         `json:"leftoverItems,omitempty"`
	LeftoverEnchants []Enchant      `json:"leftoverEnchants,omitempty"`
}

// Talent is one talent identity in a tree. ID is either def:<definitionId> or
// the documented location fallback loc:<class>:<treeIndex>:<row>,<col>.
type Talent struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Icon      string `json:"icon,omitempty"`
	SpellID   int    `json:"spellId,omitempty"`
	MaxPoints int    `json:"maxPoints"`
	Row       int    `json:"row"`
	Col       int    `json:"col"`
	PrereqID  string `json:"prereqId,omitempty"`
}

// TalentTree is one of the three trees of a class.
type TalentTree struct {
	Name       string   `json:"name"`
	Background string   `json:"background,omitempty"`
	Talents    []Talent `json:"talents"`
}

// TalentClass is the whole talent structure for one class.
type TalentClass struct {
	Class Class        `json:"class"`
	Trees []TalentTree `json:"trees"`
}

// SpecInfo is one playable specialization, derived from a class tree.
type SpecInfo struct {
	Key   Spec   `json:"key"`
	Name  string `json:"name"`
	Class Class  `json:"class"`
}

// ClassInfo is the class-level metadata exposed by /api/meta.
type ClassInfo struct {
	Class Class      `json:"class"`
	Name  string     `json:"name"`
	Specs []SpecInfo `json:"specs"`
}

// WeaponEligibility is one class weapon proficiency.
type WeaponEligibility struct {
	WeaponType    string `json:"weaponType"`
	CanUseTwoHand bool   `json:"canUseTwoHand"`
}

// Capabilities is the pinned class/spec capability data used by validation.
type Capabilities struct {
	Classes   []ClassInfo                   `json:"classes"`
	Specs     map[Class][]SpecInfo          `json:"specs"`
	Armor     map[Class][]string            `json:"armor"`
	Weapons   map[Class][]WeaponEligibility `json:"weapons"`
	Ranged    map[Class][]string            `json:"ranged"`
	Races     map[Class][]string            `json:"races"`
	DualWield map[Spec]bool                 `json:"dualWield"`
}

// Manifest pins one generated catalog build. It carries no timestamp so
// regeneration is byte-stable.
type Manifest struct {
	CatalogRevision string         `json:"catalogRevision"`
	SourceCommit    string         `json:"sourceCommit"`
	SHA256          ManifestHashes `json:"sha256"`
	Counts          ManifestCounts `json:"counts"`
}

// ManifestHashes records the raw payload hashes the revision was built from.
type ManifestHashes struct {
	DB         string `json:"db"`
	LeftoverDB string `json:"leftoverDb"`
}

// ManifestCounts records source and eligible row counts.
type ManifestCounts struct {
	ItemsTotal       int `json:"itemsTotal"`
	ItemsEligible    int `json:"itemsEligible"`
	ItemsLeftover    int `json:"itemsLeftover"`
	Enchants         int `json:"enchants"`
	LeftoverEnchants int `json:"leftoverEnchants"`
	Talents          int `json:"talents"`
	Classes          int `json:"classes"`
	Suffixes         int `json:"suffixes"`
}
