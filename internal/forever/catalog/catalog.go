// Package catalog loads the embedded, pinned Forever assets and exposes the
// indexes the API and validators need. It never fetches upstream.
package catalog

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"sort"

	foreverassets "github.com/wowsims/tbc/assets/forever"
	"github.com/wowsims/tbc/internal/forever/domain"
)

// Catalog is the loaded, indexed view of the generated Forever assets.
type Catalog struct {
	Manifest         domain.Manifest
	Items            map[int]*domain.Item
	ItemList         []*domain.Item
	EnchantsByEffect map[int][]*domain.Enchant
	EnchantList      []*domain.Enchant
	Suffixes         map[int]*domain.RandomSuffix
	Talents          map[domain.Class]*domain.TalentClass
	Capabilities     domain.Capabilities
}

// Revision returns the pinned catalog revision of the loaded assets.
func (c *Catalog) Revision() string {
	return c.Manifest.CatalogRevision
}

// Load reads the embedded catalog.
func Load() (*Catalog, error) {
	return LoadFrom(foreverassets.FS)
}

// LoadFrom reads a catalog from any filesystem. A corrupt, partial or
// internally inconsistent asset set is an error; it never yields an empty
// catalog.
func LoadFrom(fsys fs.FS) (*Catalog, error) {
	rawCatalog, err := fs.ReadFile(fsys, "catalog.json.gz")
	if err != nil {
		return nil, fmt.Errorf("read catalog.json.gz: %w", err)
	}
	reader, err := gzip.NewReader(bytes.NewReader(rawCatalog))
	if err != nil {
		return nil, fmt.Errorf("catalog.json.gz is not gzip: %w", err)
	}
	var data domain.Catalog
	if err := json.NewDecoder(reader).Decode(&data); err != nil {
		return nil, fmt.Errorf("decode catalog.json.gz: %w", err)
	}
	_ = reader.Close()
	if len(data.Items) == 0 {
		return nil, errors.New("catalog has no items")
	}

	manifest, err := readJSON[domain.Manifest](fsys, "manifest.json")
	if err != nil {
		return nil, err
	}
	if manifest.CatalogRevision == "" {
		return nil, errors.New("catalog manifest has an empty revision")
	}
	capabilities, err := readJSON[domain.Capabilities](fsys, "capabilities.json")
	if err != nil {
		return nil, err
	}
	if len(capabilities.Specs) == 0 {
		return nil, errors.New("catalog capabilities have no specs")
	}

	cat := &Catalog{
		Manifest:         manifest,
		Items:            make(map[int]*domain.Item, len(data.Items)),
		ItemList:         make([]*domain.Item, 0, len(data.Items)),
		EnchantsByEffect: make(map[int][]*domain.Enchant, len(data.Enchants)),
		EnchantList:      make([]*domain.Enchant, 0, len(data.Enchants)),
		Suffixes:         make(map[int]*domain.RandomSuffix, len(data.Suffixes)),
		Talents:          make(map[domain.Class]*domain.TalentClass, manifest.Counts.Classes),
		Capabilities:     capabilities,
	}

	for index := range data.Items {
		item := &data.Items[index]
		if _, duplicate := cat.Items[item.ID]; duplicate {
			return nil, fmt.Errorf("duplicate item id %d", item.ID)
		}
		cat.Items[item.ID] = item
		cat.ItemList = append(cat.ItemList, item)
	}
	for index := range data.Suffixes {
		suffix := &data.Suffixes[index]
		if _, duplicate := cat.Suffixes[suffix.ID]; duplicate {
			return nil, fmt.Errorf("duplicate random suffix id %d", suffix.ID)
		}
		cat.Suffixes[suffix.ID] = suffix
	}
	for index := range data.Enchants {
		enchant := &data.Enchants[index]
		cat.EnchantsByEffect[enchant.EffectID] = append(cat.EnchantsByEffect[enchant.EffectID], enchant)
		cat.EnchantList = append(cat.EnchantList, enchant)
	}
	for _, item := range cat.ItemList {
		for _, suffixID := range item.SuffixIDs {
			if _, ok := cat.Suffixes[suffixID]; !ok {
				return nil, fmt.Errorf("item %d references unknown random suffix %d", item.ID, suffixID)
			}
		}
	}

	talentFiles, err := fs.Glob(fsys, "talents/*.json")
	if err != nil {
		return nil, fmt.Errorf("glob talents: %w", err)
	}
	if len(talentFiles) == 0 {
		return nil, errors.New("catalog has no talent trees")
	}
	sort.Strings(talentFiles)
	talentCount := 0
	for _, path := range talentFiles {
		talentClass, err := readJSON[domain.TalentClass](fsys, path)
		if err != nil {
			return nil, err
		}
		if talentClass.Class == "" {
			return nil, fmt.Errorf("%s: empty class", path)
		}
		if _, duplicate := cat.Talents[talentClass.Class]; duplicate {
			return nil, fmt.Errorf("%s: duplicate class %s", path, talentClass.Class)
		}
		if len(talentClass.Trees) == 0 {
			return nil, fmt.Errorf("%s: class %s has no talent trees", path, talentClass.Class)
		}
		talentIDs := make(map[string]bool)
		for _, tree := range talentClass.Trees {
			for _, talent := range tree.Talents {
				if talent.ID == "" {
					return nil, fmt.Errorf("%s: %s tree %q has a talent without an id", path, talentClass.Class, tree.Name)
				}
				if talentIDs[talent.ID] {
					return nil, fmt.Errorf("%s: duplicate talent id %q", path, talent.ID)
				}
				talentIDs[talent.ID] = true
				talentCount++
			}
		}
		for _, tree := range talentClass.Trees {
			for _, talent := range tree.Talents {
				if talent.PrereqID != "" && !talentIDs[talent.PrereqID] {
					return nil, fmt.Errorf(
						"%s: talent %q references unknown prerequisite %q",
						path, talent.ID, talent.PrereqID,
					)
				}
			}
		}
		cat.Talents[talentClass.Class] = &talentClass
	}

	if err := cat.checkCounts(talentCount); err != nil {
		return nil, err
	}
	return cat, nil
}

func (c *Catalog) checkCounts(talentCount int) error {
	counts := c.Manifest.Counts
	checks := []struct {
		name   string
		got    int
		want   int
		assert bool
	}{
		{"items", len(c.Items), counts.ItemsEligible, counts.ItemsEligible != 0},
		{"enchants", len(c.EnchantList), counts.Enchants, counts.Enchants != 0},
		{"suffixes", len(c.Suffixes), counts.Suffixes, counts.Suffixes != 0},
		{"talents", talentCount, counts.Talents, counts.Talents != 0},
		{"classes", len(c.Talents), counts.Classes, counts.Classes != 0},
	}
	for _, check := range checks {
		if check.assert && check.got != check.want {
			return fmt.Errorf("catalog %s = %d, manifest says %d", check.name, check.got, check.want)
		}
	}
	return nil
}

func readJSON[T any](fsys fs.FS, path string) (T, error) {
	var value T
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return value, fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &value); err != nil {
		return value, fmt.Errorf("decode %s: %w", path, err)
	}
	return value, nil
}
