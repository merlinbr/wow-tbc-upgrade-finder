package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/wowsims/tbc/internal/forever/domain"
)

// RawTalents is the upstream talent input: one JSON document per class plus the
// spell icon lookup used to resolve talent icons.
type RawTalents struct {
	Classes    map[domain.Class][]byte
	SpellIcons map[int]string
}

type rawTalentLocation struct {
	RowIdx int `json:"rowIdx"`
	ColIdx int `json:"colIdx"`
}

type rawTalent struct {
	FieldName      string             `json:"fieldName"`
	FancyName      string             `json:"fancyName"`
	Name           string             `json:"name"`
	Location       rawTalentLocation  `json:"location"`
	SpellID        int                `json:"spellId"`
	DefinitionID   int                `json:"definitionId"`
	MaxPoints      int                `json:"maxPoints"`
	PrereqLocation *rawTalentLocation `json:"prereqLocation"`
}

type rawTalentTree struct {
	Name          string      `json:"name"`
	BackgroundURL string      `json:"backgroundUrl"`
	Talents       []rawTalent `json:"talents"`
}

// ProjectTalents normalizes the pinned class tree documents. Talent identity is
// definition-based where available, otherwise the documented location fallback.
func ProjectTalents(raw RawTalents, enums Enums) (map[domain.Class]*domain.TalentClass, error) {
	classes := make([]string, 0, len(raw.Classes))
	for class := range raw.Classes {
		classes = append(classes, class)
	}
	sort.Strings(classes)

	result := make(map[domain.Class]*domain.TalentClass, len(classes))
	for _, class := range classes {
		if _, ok := enums.Value("Class", class); !ok {
			return nil, fmt.Errorf("talent file for unknown class %q", class)
		}
		var rows []rawTalentTree
		if err := json.Unmarshal(raw.Classes[class], &rows); err != nil {
			return nil, fmt.Errorf("parse %s talents: %w", class, err)
		}
		talentClass := &domain.TalentClass{Class: class}
		for treeIndex, row := range rows {
			tree := domain.TalentTree{Name: row.Name, Background: row.BackgroundURL}
			byLocation := make(map[rawTalentLocation]string, len(row.Talents))
			for _, talent := range row.Talents {
				id := talentID(class, treeIndex, talent)
				if _, exists := byLocation[talent.Location]; exists {
					return nil, fmt.Errorf("%s tree %q: duplicate talent location %+v", class, row.Name, talent.Location)
				}
				byLocation[talent.Location] = id
				name := talent.FancyName
				if name == "" {
					name = talent.Name
				}
				tree.Talents = append(tree.Talents, domain.Talent{
					ID:        id,
					Name:      name,
					Icon:      raw.SpellIcons[talent.SpellID],
					SpellID:   talent.SpellID,
					MaxPoints: talent.MaxPoints,
					Row:       talent.Location.RowIdx,
					Col:       talent.Location.ColIdx,
				})
			}
			for talentIndex, talent := range row.Talents {
				if talent.PrereqLocation == nil {
					continue
				}
				prereqID, ok := byLocation[*talent.PrereqLocation]
				if !ok {
					return nil, fmt.Errorf(
						"%s tree %q talent %q: unresolved prerequisite location %+v",
						class, row.Name, talent.FancyName, *talent.PrereqLocation,
					)
				}
				tree.Talents[talentIndex].PrereqID = prereqID
			}
			talentClass.Trees = append(talentClass.Trees, tree)
		}
		if len(talentClass.Trees) == 0 {
			return nil, fmt.Errorf("%s talents: no trees", class)
		}
		result[class] = talentClass
	}
	return result, nil
}

func talentID(class domain.Class, treeIndex int, talent rawTalent) string {
	if talent.DefinitionID != 0 {
		return fmt.Sprintf("def:%d", talent.DefinitionID)
	}
	return fmt.Sprintf(
		"loc:%s:%d:%d,%d",
		class, treeIndex, talent.Location.RowIdx, talent.Location.ColIdx,
	)
}

// SpecsFromTrees derives one spec key per class tree: <class>_<tree-slug>.
func SpecsFromTrees(trees map[domain.Class]*domain.TalentClass) map[domain.Class][]domain.SpecInfo {
	specs := make(map[domain.Class][]domain.SpecInfo, len(trees))
	for class, talentClass := range trees {
		for _, tree := range talentClass.Trees {
			specs[class] = append(specs[class], domain.SpecInfo{
				Key:   domain.Spec(class + "_" + snakeCase(tree.Name)),
				Name:  tree.Name,
				Class: class,
			})
		}
	}
	return specs
}

// CountTalents totals every talent row across every class.
func CountTalents(trees map[domain.Class]*domain.TalentClass) int {
	total := 0
	for _, talentClass := range trees {
		for _, tree := range talentClass.Trees {
			total += len(tree.Talents)
		}
	}
	return total
}

type rawSpellIcons struct {
	SpellIcons []struct {
		ID   int    `json:"id"`
		Icon string `json:"icon"`
	} `json:"spellIcons"`
}

// ParseSpellIcons builds the spellId -> icon map from db.json.
func ParseSpellIcons(db []byte) (map[int]string, error) {
	var rows rawSpellIcons
	if err := json.Unmarshal(db, &rows); err != nil {
		return nil, fmt.Errorf("parse spell icons: %w", err)
	}
	icons := make(map[int]string, len(rows.SpellIcons))
	for _, row := range rows.SpellIcons {
		if row.ID == 0 || row.Icon == "" {
			continue
		}
		icons[row.ID] = row.Icon
	}
	return icons, nil
}
