package main

import (
	"encoding/json"
	"os"

	"github.com/wowsims/tbc/internal/forever/domain"
)

// BuildManifest assembles the pinned manifest. No timestamp is written, so a
// regeneration from the same commit is byte-stable.
func BuildManifest(raw RawPayload, catalog domain.Catalog, counts RawCounts, talentCount, classCount int) domain.Manifest {
	return domain.Manifest{
		CatalogRevision: raw.Revision(),
		SourceCommit:    pinnedCommit,
		SHA256: domain.ManifestHashes{
			DB:         raw.DBHash,
			LeftoverDB: raw.LeftoverHash,
		},
		Counts: domain.ManifestCounts{
			ItemsTotal:       counts.ItemsTotal,
			ItemsEligible:    len(catalog.Items),
			ItemsLeftover:    counts.ItemsLeftover,
			Enchants:         counts.Enchants,
			LeftoverEnchants: counts.LeftoverEnchants,
			Talents:          talentCount,
			Classes:          classCount,
			Suffixes:         counts.Suffixes,
		},
	}
}

// writeJSON writes indented JSON with a trailing newline.
func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
