package upgrades

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/wowsims/tbc/assets/database"
	"github.com/wowsims/tbc/sim/core"
	"github.com/wowsims/tbc/sim/core/proto"
	"github.com/wowsims/tbc/sim/core/simsignals"
)

// The pinned build commands omit the upstream `with_db` tag, so the simulator's
// global item/gem/enchant maps start empty. The upstream wasm path embeds the
// database in the request's player record; we do the same for the first run,
// after which every request resolves from the populated global maps.
var attachDatabaseOnce sync.Once

func attachSimDatabase(requestCopy *proto.RaidSimRequest) {
	attachDatabaseOnce.Do(func() {
		requestCopy.Raid.Parties[0].Players[0].Database = buildSimDatabase()
	})
}

func buildSimDatabase() *proto.SimDatabase {
	db := database.Load()
	simDB := &proto.SimDatabase{
		Items:                    make([]*proto.SimItem, len(db.Items)),
		Enchants:                 make([]*proto.SimEnchant, len(db.Enchants)),
		Gems:                     make([]*proto.SimGem, len(db.Gems)),
		ItemEffectRandPropPoints: make([]*proto.ItemEffectRandPropPoints, len(db.ItemEffectRandPropPoints)),
		RandomSuffixes:           make([]*proto.ItemRandomSuffix, len(db.RandomSuffixes)),
		Consumables:              make([]*proto.Consumable, len(db.Consumables)),
		SpellEffects:             make([]*proto.SpellEffect, len(db.SpellEffects)),
	}

	for i, item := range db.Items {
		simDB.Items[i] = &proto.SimItem{
			Id:               item.Id,
			Name:             item.Name,
			Type:             item.Type,
			ArmorType:        item.ArmorType,
			WeaponType:       item.WeaponType,
			HandType:         item.HandType,
			RangedWeaponType: item.RangedWeaponType,
			GemSockets:       item.GemSockets,
			SocketBonus:      item.SocketBonus,
			WeaponSpeed:      item.WeaponSpeed,
			QualityModifier:  item.QualityModifier,
			SetName:          item.SetName,
			SetId:            item.SetId,
			ScalingOptions:   item.ScalingOptions,
			ItemEffects:      item.ItemEffects,
		}
	}
	for i, suffix := range db.RandomSuffixes {
		simDB.RandomSuffixes[i] = &proto.ItemRandomSuffix{
			Id:    suffix.Id,
			Name:  suffix.Name,
			Stats: suffix.Stats,
		}
	}
	for i, enchant := range db.Enchants {
		simDB.Enchants[i] = &proto.SimEnchant{
			EffectId:       enchant.EffectId,
			Stats:          enchant.Stats,
			EnchantEffects: enchant.EnchantEffects,
			Name:           enchant.Name,
			Type:           enchant.Type,
		}
	}
	for i, gem := range db.Gems {
		simDB.Gems[i] = &proto.SimGem{
			Id:    gem.Id,
			Name:  gem.Name,
			Color: gem.Color,
			Stats: gem.Stats,
		}
	}
	for i, itemEffectRpp := range db.ItemEffectRandPropPoints {
		simDB.ItemEffectRandPropPoints[i] = &proto.ItemEffectRandPropPoints{
			Ilvl:           itemEffectRpp.Ilvl,
			RandPropPoints: itemEffectRpp.RandPropPoints,
		}
	}
	copy(simDB.Consumables, db.Consumables)
	copy(simDB.SpellEffects, db.SpellEffects)

	return simDB
}

type RealSimulator struct {
	requestCounter atomic.Uint64
}

func NewRealSimulator() *RealSimulator {
	return &RealSimulator{}
}

// Run executes one simulation through the wowsims engine.
// It copies the request, gives every run a unique request ID, forwards only
// completed/total iteration counts to onProgress, and extracts the imported
// character's player DPS distribution from the first party/first player.
func (s *RealSimulator) Run(ctx context.Context, request *proto.RaidSimRequest, onProgress func(completed, total int32)) (DPSResult, error) {
	if request == nil {
		return DPSResult{}, fmt.Errorf("nil sim request")
	}

	requestCopy := cloneMessage(request)
	attachSimDatabase(requestCopy)
	requestID := fmt.Sprintf("rank-upgrades-%d-%s", s.requestCounter.Add(1), uuid.NewString())

	reporter := make(chan *proto.ProgressMetrics, 10)
	core.RunRaidSimConcurrentAsync(requestCopy, reporter, requestID)

	iterations := requestCopy.SimOptions.GetIterations()

	for {
		select {
		case <-ctx.Done():
			simsignals.AbortById(requestID)
			// Drain the reporter channel until the simulator closes it so no
			// goroutine blocks on a full channel.
			for range reporter {
			}
			return DPSResult{}, ctx.Err()
		case msg, ok := <-reporter:
			if !ok {
				return DPSResult{}, fmt.Errorf("simulator stopped without producing a final result")
			}
			if msg == nil {
				continue
			}
			if final := msg.FinalRaidResult; final != nil {
				if err := final.GetError(); err != nil {
					return DPSResult{}, fmt.Errorf("%s", err.GetMessage())
				}
				return extractPlayerDPS(final, iterations)
			}
			if onProgress != nil {
				onProgress(msg.GetCompletedIterations(), msg.GetTotalIterations())
			}
		}
	}
}

func extractPlayerDPS(result *proto.RaidSimResult, iterations int32) (DPSResult, error) {
	parties := result.GetRaidMetrics().GetParties()
	if len(parties) == 0 {
		return DPSResult{}, fmt.Errorf("sim result contains no parties")
	}
	players := parties[0].GetPlayers()
	if len(players) == 0 {
		return DPSResult{}, fmt.Errorf("sim result contains no players")
	}
	dps := players[0].GetDps()
	return DPSResult{
		Average:    dps.GetAvg(),
		Stdev:      dps.GetStdev(),
		Iterations: iterations,
	}, nil
}
