package upgrades

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/wowsims/tbc/assets/database"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	googleProto "google.golang.org/protobuf/proto"
)

// ParseLinkPayload extracts the raw uncompressed protobuf bytes from a wowsims link.
// It returns (payloadBytes, isRaidLink, error).
func ParseLinkPayload(link string) ([]byte, bool, error) {
	parts := strings.Split(link, "#")
	if len(parts) != 2 || parts[1] == "" {
		return nil, false, &ValidationError{
			Code:    "invalid_link",
			Message: "malformed export link: missing or invalid fragment",
		}
	}

	isRaid := strings.Contains(link, "/raid/")

	raw, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, false, &ValidationError{
			Code:    "invalid_link",
			Message: fmt.Sprintf("cannot decode base64 from link: %v", err),
		}
	}

	r, err := zlib.NewReader(bytes.NewReader(raw))
	if err != nil {
		return nil, false, &ValidationError{
			Code:    "invalid_link",
			Message: fmt.Sprintf("cannot create zlib reader: %v", err),
		}
	}
	defer r.Close()

	payload, err := io.ReadAll(r)
	if err != nil {
		return nil, false, &ValidationError{
			Code:    "invalid_link",
			Message: fmt.Sprintf("reading zlib data failed: %v", err),
		}
	}

	return payload, isRaid, nil
}

func formatPlayerSpec(player *proto.Player) string {
	if player == nil {
		return ""
	}
	switch player.GetSpec().(type) {
	case *proto.Player_BalanceDruid:
		return "BalanceDruid"
	case *proto.Player_FeralCatDruid:
		return "FeralCatDruid"
	case *proto.Player_FeralBearDruid:
		return "FeralBearDruid"
	case *proto.Player_RestorationDruid:
		return "RestorationDruid"
	case *proto.Player_Hunter:
		return "Hunter"
	case *proto.Player_Mage:
		return "Mage"
	case *proto.Player_HolyPaladin:
		return "HolyPaladin"
	case *proto.Player_ProtectionPaladin:
		return "ProtectionPaladin"
	case *proto.Player_RetributionPaladin:
		return "RetributionPaladin"
	case *proto.Player_Priest:
		return "Priest"
	case *proto.Player_Rogue:
		return "Rogue"
	case *proto.Player_ElementalShaman:
		return "ElementalShaman"
	case *proto.Player_EnhancementShaman:
		return "EnhancementShaman"
	case *proto.Player_RestorationShaman:
		return "RestorationShaman"
	case *proto.Player_Warlock:
		return "Warlock"
	case *proto.Player_DpsWarrior:
		return "DpsWarrior"
	case *proto.Player_ProtectionWarrior:
		return "ProtectionWarrior"
	default:
		return player.GetClass().String()
	}
}

// Import parses and validates a wowsims individual export link into an immutable ImportedSettings.
func Import(link string) (*ImportedSettings, error) {
	payload, isRaid, err := ParseLinkPayload(link)
	if err != nil {
		return nil, err
	}

	if isRaid {
		return nil, &ValidationError{
			Code:    "unsupported_link",
			Message: "raid sim export links are not supported; please export from an individual sim",
		}
	}

	settings := &proto.IndividualSimSettings{}
	if err := googleProto.Unmarshal(payload, settings); err != nil {
		return nil, &ValidationError{
			Code:    "invalid_link",
			Message: fmt.Sprintf("cannot unmarshal individual settings proto: %v", err),
		}
	}

	if settings.Player == nil {
		return nil, &ValidationError{
			Code:    "incompatible_settings",
			Message: "missing player configuration in exported settings",
		}
	}

	if settings.Encounter == nil {
		return nil, &ValidationError{
			Code:    "incompatible_settings",
			Message: "missing encounter configuration in exported settings",
		}
	}


	if settings.Player.Equipment == nil || len(settings.Player.Equipment.Items) == 0 {
		return nil, &ValidationError{
			Code:    "incompatible_settings",
			Message: "missing player equipment in export",
		}
	}

	currentVersion := core.GetCurrentProtoVersion()
	if settings.ApiVersion > currentVersion {
		return nil, &ValidationError{
			Code:    "incompatible_settings",
			Message: fmt.Sprintf("unsupported API version %d (max supported is %d)", settings.ApiVersion, currentVersion),
		}
	}

	db := database.Load()
	knownItems := make(map[int32]bool, len(db.GetItems()))
	knownSuffixes := make(map[int32]bool, len(db.GetRandomSuffixes()))
	knownGems := make(map[int32]bool, len(db.GetGems()))
	knownEnchants := make(map[int32]bool, len(db.GetEnchants()))
	for _, item := range db.GetItems() {
		knownItems[item.GetId()] = true
	}
	for _, suffix := range db.GetRandomSuffixes() {
		knownSuffixes[suffix.GetId()] = true
	}
	for _, gem := range db.GetGems() {
		knownGems[gem.GetId()] = true
	}
	for _, ench := range db.GetEnchants() {
		knownEnchants[ench.GetEffectId()] = true
	}

	validateItemSpec := func(itemSpec *proto.ItemSpec) error {
		if !knownItems[itemSpec.Id] {
			return &ValidationError{
				Code:    "incompatible_item",
				Message: fmt.Sprintf("equipped item ID %d not found in item database", itemSpec.Id),
			}
		}
		if itemSpec.GetRandomSuffix() != 0 && !knownSuffixes[itemSpec.GetRandomSuffix()] {
			return &ValidationError{
				Code:    "incompatible_random_suffix",
				Message: fmt.Sprintf("equipped random suffix ID %d not found in item database", itemSpec.GetRandomSuffix()),
			}
		}
		if itemSpec.GetEnchant() != 0 && !knownEnchants[itemSpec.GetEnchant()] {
			return &ValidationError{
				Code:    "incompatible_enchant",
				Message: fmt.Sprintf("equipped enchant effect ID %d not found in item database", itemSpec.GetEnchant()),
			}
		}
		for _, gemID := range itemSpec.GetGems() {
			if gemID != 0 && !knownGems[gemID] {
				return &ValidationError{
					Code:    "incompatible_gem",
					Message: fmt.Sprintf("equipped gem ID %d not found in item database", gemID),
				}
			}
		}
		return nil
	}

	equippedCount := 0
	for _, itemSpec := range settings.Player.Equipment.Items {
		if itemSpec != nil && itemSpec.Id > 0 {
			if err := validateItemSpec(itemSpec); err != nil {
				return nil, err
			}
			equippedCount++
		}
	}
	if settings.Player.ItemSwap != nil {
		for _, itemSpec := range settings.Player.ItemSwap.Items {
			if itemSpec != nil && itemSpec.Id > 0 {
				if err := validateItemSpec(itemSpec); err != nil {
					return nil, err
				}
			}
		}
	}

	deterministicBytes, err := googleProto.MarshalOptions{Deterministic: true}.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal settings deterministically: %w", err)
	}
	hash := sha256.Sum256(deterministicBytes)
	digest := hex.EncodeToString(hash[:])

	var professions []proto.Profession
	if settings.Player.Profession1 != proto.Profession_ProfessionUnknown {
		professions = append(professions, settings.Player.Profession1)
	}
	if settings.Player.Profession2 != proto.Profession_ProfessionUnknown {
		professions = append(professions, settings.Player.Profession2)
	}

	var encounterTargets int
	if settings.Encounter != nil {
		encounterTargets = len(settings.Encounter.Targets)
	}

	summary := CharacterSummary{
		Name:             settings.Player.Name,
		Class:            settings.Player.Class.String(),
		Spec:             formatPlayerSpec(settings.Player),
		Race:             settings.Player.Race.String(),
		EquippedItems:    equippedCount,
		Professions:      professions,
		Phase:            settings.GetSettings().GetPhase(),
		Iterations:       settings.GetSettings().GetIterations(),
		FixedRngSeed:     settings.GetSettings().GetFixedRngSeed() != 0,
		EncounterTargets: encounterTargets,
	}

	return &ImportedSettings{
		Link:             link,
		Settings:         settings,
		SettingsDigest:   digest,
		Character:        summary,
		SimulatorVersion: SimulatorRevision,
		DatabaseVersion:  DatabaseRevision,
	}, nil
}

// NewRequest converts the immutable imported baseline into a new canonical RaidSimRequest.
func (s *ImportedSettings) NewRequest(iterations int32) *proto.RaidSimRequest {
	playerClone := cloneMessage(s.Settings.Player)
	var partyBuffsClone *proto.PartyBuffs
	if s.Settings.PartyBuffs != nil {
		partyBuffsClone = cloneMessage(s.Settings.PartyBuffs)
	} else {
		partyBuffsClone = &proto.PartyBuffs{}
	}

	party := &proto.Party{
		Players: []*proto.Player{playerClone},
		Buffs:   partyBuffsClone,
	}

	var raidBuffsClone *proto.RaidBuffs
	if s.Settings.RaidBuffs != nil {
		raidBuffsClone = cloneMessage(s.Settings.RaidBuffs)
	} else {
		raidBuffsClone = &proto.RaidBuffs{}
	}

	var debuffsClone *proto.Debuffs
	if s.Settings.Debuffs != nil {
		debuffsClone = cloneMessage(s.Settings.Debuffs)
	} else {
		debuffsClone = &proto.Debuffs{}
	}

	var encounterClone *proto.Encounter
	if s.Settings.Encounter != nil {
		encounterClone = cloneMessage(s.Settings.Encounter)
	} else {
		encounterClone = &proto.Encounter{}
	}

	var tanks []*proto.UnitReference
	for _, tank := range s.Settings.Tanks {
		tanks = append(tanks, cloneMessage(tank))
	}

	var randomSeed int64
	if s.Settings.Settings != nil {
		randomSeed = s.Settings.Settings.FixedRngSeed
	}

	return &proto.RaidSimRequest{
		Type: proto.SimType_SimTypeIndividual,
		Raid: &proto.Raid{
			Parties:       []*proto.Party{party},
			Buffs:         raidBuffsClone,
			Debuffs:       debuffsClone,
			Tanks:         tanks,
			TargetDummies: s.Settings.TargetDummies,
		},
		Encounter: encounterClone,
		SimOptions: &proto.SimOptions{
			Iterations:          iterations,
			RandomSeed:          randomSeed,
			DebugFirstIteration: true,
		},
	}
}
