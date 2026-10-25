package validate

import (
	"fmt"
	"slices"
	"sort"

	"github.com/wowsims/tbc/internal/forever/domain"
)

type locatedTalent struct {
	talent domain.Talent
	tree   int
}

type talentIndex map[string]locatedTalent

func newTalentIndex(trees *domain.TalentClass) talentIndex {
	index := make(talentIndex)
	if trees == nil {
		return index
	}
	for treeIndex, tree := range trees.Trees {
		for _, talent := range tree.Talents {
			if _, exists := index[talent.ID]; exists {
				continue
			}
			index[talent.ID] = locatedTalent{talent: talent, tree: treeIndex}
		}
	}
	return index
}

func talentBudget(level int) int {
	if level <= 9 {
		return 0
	}
	return level - 9
}

// Talents validates a stored allocation and reports broken talents. It never
// clamps: invalid points stay in the map and are flagged.
func Talents(ch domain.Character, alloc domain.Talents, trees *domain.TalentClass) ([]Violation, domain.TalentIssues) {
	index := newTalentIndex(trees)
	issues := domain.TalentIssues{Broken: make([]string, 0)}
	var violations []Violation

	for id, points := range alloc {
		located, known := index[id]
		if !known {
			violations = append(violations, Violation{
				Code:     "unknown_talent",
				Severity: "error",
				Message:  fmt.Sprintf("Talent %q is not in the pinned trees.", id),
				TalentID: id,
			})
			issues.Broken = append(issues.Broken, id)
			continue
		}
		if points != 0 {
			issues.TotalPoints += points
		}
		if points < 0 || points > located.talent.MaxPoints {
			violations = append(violations, Violation{
				Code:     "talent_rank_exceeded",
				Severity: "error",
				Message: fmt.Sprintf(
					"%s has %d points; the maximum is %d.",
					located.talent.Name, points, located.talent.MaxPoints,
				),
				TalentID: id,
			})
		}
	}

	if budget := talentBudget(ch.Level); issues.TotalPoints > budget {
		violations = append(violations, Violation{
			Code:     "talent_budget_exceeded",
			Severity: "error",
			Message: fmt.Sprintf(
				"The allocation spends %d points; %d are allowed at level %d.",
				issues.TotalPoints, budget, ch.Level,
			),
		})
	}

	evaluator := newTalentEvaluator(index, alloc)
	for id, points := range alloc {
		located, known := index[id]
		if !known || points <= 0 {
			continue
		}
		result := evaluator.evaluate(id)
		if result.tierFail {
			violations = append(violations, Violation{
				Code:     "talent_tier_locked",
				Severity: "error",
				Message: fmt.Sprintf(
					"%s needs %d points in higher rows of its tree.",
					located.talent.Name, 5*located.talent.Row,
				),
				TalentID: id,
			})
		}
		if result.prereqFail {
			message := fmt.Sprintf("%s has an unmet prerequisite.", located.talent.Name)
			if prereq, ok := index[located.talent.PrereqID]; ok {
				message = fmt.Sprintf(
					"%s requires %s at maximum rank and unbroken.",
					located.talent.Name, prereq.talent.Name,
				)
			}
			violations = append(violations, Violation{
				Code:     "talent_prereq_unmet",
				Severity: "error",
				Message:  message,
				TalentID: id,
			})
		}
		if result.broken {
			issues.Broken = append(issues.Broken, id)
		}
	}

	sort.Strings(issues.Broken)
	issues.Broken = slices.Compact(issues.Broken)
	return violations, issues
}

// TrySetPoints applies a point change at the level-60 budget. Callers that need
// the character's actual budget use TrySetPointsAtLevel.
func TrySetPoints(trees *domain.TalentClass, alloc domain.Talents, id string, next int) (domain.Talents, []Violation) {
	return TrySetPointsAtLevel(trees, alloc, id, next, 60)
}

// TrySetPointsAtLevel allows any decrease and rejects an increase that would
// leave the talent broken or exceed the level budget.
func TrySetPointsAtLevel(trees *domain.TalentClass, alloc domain.Talents, id string, next, level int) (domain.Talents, []Violation) {
	index := newTalentIndex(trees)
	updated := cloneTalents(alloc)

	located, known := index[id]
	if !known {
		if next > 0 {
			return alloc, []Violation{{
				Code:     "unknown_talent",
				Severity: "error",
				Message:  fmt.Sprintf("Talent %q is not in the pinned trees.", id),
				TalentID: id,
			}}
		}
		delete(updated, id)
		return updated, nil
	}
	if next < 0 || next > located.talent.MaxPoints {
		return alloc, []Violation{{
			Code:     "talent_rank_exceeded",
			Severity: "error",
			Message: fmt.Sprintf(
				"%s has a maximum of %d points.", located.talent.Name, located.talent.MaxPoints,
			),
			TalentID: id,
		}}
	}

	current := updated[id]
	if next <= current {
		if next == 0 {
			delete(updated, id)
		} else {
			updated[id] = next
		}
		return updated, nil
	}

	updated[id] = next
	evaluator := newTalentEvaluator(index, updated)
	result := evaluator.evaluate(id)
	if result.tierFail {
		return alloc, []Violation{{
			Code:     "talent_tier_locked",
			Severity: "error",
			Message: fmt.Sprintf(
				"%s needs %d points in higher rows of its tree.",
				located.talent.Name, 5*located.talent.Row,
			),
			TalentID: id,
		}}
	}
	if result.prereqFail {
		return alloc, []Violation{{
			Code:     "talent_prereq_unmet",
			Severity: "error",
			Message:  fmt.Sprintf("%s has an unmet prerequisite.", located.talent.Name),
			TalentID: id,
		}}
	}
	if result.broken {
		return alloc, []Violation{{
			Code:     "talent_prereq_unmet",
			Severity: "error",
			Message:  fmt.Sprintf("%s would be broken.", located.talent.Name),
			TalentID: id,
		}}
	}

	total := 0
	for _, points := range updated {
		if points > 0 {
			total += points
		}
	}
	if budget := talentBudget(level); total > budget {
		return alloc, []Violation{{
			Code:     "talent_budget_exceeded",
			Severity: "error",
			Message: fmt.Sprintf(
				"The allocation would spend %d points; %d are allowed at level %d.",
				total, budget, level,
			),
		}}
	}
	return updated, nil
}

func cloneTalents(alloc domain.Talents) domain.Talents {
	updated := make(domain.Talents, len(alloc))
	for id, points := range alloc {
		updated[id] = points
	}
	return updated
}

type talentEvalResult struct {
	broken     bool
	tierFail   bool
	prereqFail bool
}

type talentEvaluator struct {
	index    talentIndex
	alloc    domain.Talents
	memo     map[string]talentEvalResult
	visiting map[string]bool
}

func newTalentEvaluator(index talentIndex, alloc domain.Talents) *talentEvaluator {
	return &talentEvaluator{
		index:    index,
		alloc:    alloc,
		memo:     make(map[string]talentEvalResult),
		visiting: make(map[string]bool),
	}
}

func (e *talentEvaluator) points(id string) int {
	located, ok := e.index[id]
	if !ok {
		return 0
	}
	points := e.alloc[id]
	if points < 0 {
		return 0
	}
	if points > located.talent.MaxPoints {
		return located.talent.MaxPoints
	}
	return points
}

func (e *talentEvaluator) evaluate(id string) talentEvalResult {
	if result, ok := e.memo[id]; ok {
		return result
	}
	if e.visiting[id] {
		return talentEvalResult{broken: true, prereqFail: true}
	}
	located, known := e.index[id]
	if !known {
		return talentEvalResult{}
	}

	e.visiting[id] = true
	defer delete(e.visiting, id)

	var result talentEvalResult
	if e.points(id) > 0 {
		if located.talent.Row > 0 {
			pointsAbove := 0
			for otherID, other := range e.index {
				if other.tree != located.tree || other.talent.Row >= located.talent.Row {
					continue
				}
				points := e.points(otherID)
				if points <= 0 || e.evaluate(otherID).broken {
					continue
				}
				pointsAbove += points
			}
			if pointsAbove < 5*located.talent.Row {
				result.tierFail = true
			}
		}
		if located.talent.PrereqID != "" {
			prereq, ok := e.index[located.talent.PrereqID]
			if !ok || e.points(located.talent.PrereqID) < prereq.talent.MaxPoints ||
				e.evaluate(located.talent.PrereqID).broken {
				result.prereqFail = true
			}
		}
		result.broken = result.tierFail || result.prereqFail
	}

	e.memo[id] = result
	return result
}
