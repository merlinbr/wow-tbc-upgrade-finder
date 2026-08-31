package upgrades

import (
	"strings"

	"github.com/wowsims/tbc/sim/core/proto"
)

type Catalog struct {
	DB            *proto.UIDatabase
	Items         map[int32]*proto.UIItem
	RandomSuffixes map[int32]*proto.ItemRandomSuffix
	Gems          map[int32]*proto.UIGem
	Enchants      map[int32]*proto.UIEnchant
	EnchantsByID  map[int32]*proto.UIEnchant
	Zones         map[int32]*proto.UIZone
	NPCs          map[int32]*proto.UINPC
}

func NewCatalog(db *proto.UIDatabase) *Catalog {
	c := &Catalog{
		DB:            db,
		Items:         make(map[int32]*proto.UIItem, len(db.GetItems())),
		RandomSuffixes: make(map[int32]*proto.ItemRandomSuffix, len(db.GetRandomSuffixes())),
		Gems:          make(map[int32]*proto.UIGem, len(db.GetGems())),
		Enchants:      make(map[int32]*proto.UIEnchant, len(db.GetEnchants())),
		EnchantsByID:  make(map[int32]*proto.UIEnchant, len(db.GetEnchants())),
		Zones:         make(map[int32]*proto.UIZone, len(db.GetZones())),
		NPCs:          make(map[int32]*proto.UINPC, len(db.GetNpcs())),
	}

	for _, item := range db.GetItems() {
		c.Items[item.GetId()] = item
	}
	for _, suffix := range db.GetRandomSuffixes() {
		c.RandomSuffixes[suffix.GetId()] = suffix
	}
	for _, gem := range db.GetGems() {
		c.Gems[gem.GetId()] = gem
	}
	for _, ench := range db.GetEnchants() {
		c.Enchants[ench.GetEffectId()] = ench
		if ench.GetItemId() > 0 {
			c.EnchantsByID[ench.GetItemId()] = ench
		}
	}
	for _, zone := range db.GetZones() {
		c.Zones[zone.GetId()] = zone
	}
	for _, npc := range db.GetNpcs() {
		c.NPCs[npc.GetId()] = npc
	}

	return c
}

func (c *Catalog) ResolveSource(item *proto.UIItem) SourceSummary {
	if item == nil || len(item.GetSources()) == 0 {
		return SourceSummary{
			Kind: proto.SourceFilterOption_SourceUnknown,
			Name: "Unknown",
		}
	}

	src := item.GetSources()[0]
	if crafted := src.GetCrafted(); crafted != nil {
		profName := crafted.GetProfession().String()
		profName = strings.TrimPrefix(profName, "Profession")
		return SourceSummary{
			Kind:     proto.SourceFilterOption_SourceCrafting,
			Name:     profName,
			Category: "Crafted",
		}
	}
	if drop := src.GetDrop(); drop != nil {
		kind := proto.SourceFilterOption_SourceDungeon
		switch drop.GetDifficulty() {
		case proto.DungeonDifficulty_DifficultyRaid10, proto.DungeonDifficulty_DifficultyRaid25:
			kind = proto.SourceFilterOption_SourceRaid
		case proto.DungeonDifficulty_DifficultyRaid10H, proto.DungeonDifficulty_DifficultyRaid25H:
			kind = proto.SourceFilterOption_SourceRaidH
		case proto.DungeonDifficulty_DifficultyHeroic:
			kind = proto.SourceFilterOption_SourceDungeonH
		}

		var zoneName, npcName string
		if z, ok := c.Zones[drop.GetZoneId()]; ok {
			zoneName = z.GetName()
		}
		if n, ok := c.NPCs[drop.GetNpcId()]; ok {
			npcName = n.GetName()
		} else if drop.GetOtherName() != "" {
			npcName = drop.GetOtherName()
		}

		name := npcName
		if name == "" {
			name = zoneName
		}
		return SourceSummary{
			Kind:     kind,
			Name:     name,
			Zone:     zoneName,
			Category: drop.GetCategory(),
		}
	}
	if quest := src.GetQuest(); quest != nil {
		return SourceSummary{
			Kind:     proto.SourceFilterOption_SourceQuest,
			Name:     quest.GetName(),
			Category: "Quest",
		}
	}
	if sold := src.GetSoldBy(); sold != nil {
		var zoneName string
		if z, ok := c.Zones[sold.GetZoneId()]; ok {
			zoneName = z.GetName()
		}
		return SourceSummary{
			Kind:     proto.SourceFilterOption_SourceSoldBy,
			Name:     sold.GetNpcName(),
			Zone:     zoneName,
			Category: "Vendor",
		}
	}
	if rep := src.GetRep(); rep != nil {
		factionName := rep.GetRepFactionId().String()
		factionName = strings.TrimPrefix(factionName, "RepFaction")
		return SourceSummary{
			Kind:     proto.SourceFilterOption_SourceReputation,
			Name:     factionName,
			Category: rep.GetRepLevel().String(),
		}
	}

	return SourceSummary{
		Kind: proto.SourceFilterOption_SourceUnknown,
		Name: "Unknown",
	}
}

func (c *Catalog) ItemSummary(item *proto.UIItem, slot proto.ItemSlot) UIItemSummary {
	if item == nil {
		return UIItemSummary{}
	}
	return UIItemSummary{
		ID:      item.GetId(),
		Name:    item.GetName(),
		Icon:    item.GetIcon(),
		Quality: item.GetQuality(),
		Phase:   item.GetPhase(),
		Type:    item.GetType(),
		Slot:    slot,
	}
}
