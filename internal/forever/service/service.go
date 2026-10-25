// Package service holds the character and build workflows, including the
// explicit-save validation boundary.
package service

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/wowsims/tbc/internal/forever/catalog"
	"github.com/wowsims/tbc/internal/forever/domain"
	"github.com/wowsims/tbc/internal/forever/store"
	"github.com/wowsims/tbc/internal/forever/validate"
)

// Service coordinates the store and the pinned catalog.
type Service struct {
	store   store.Store
	catalog *catalog.Catalog
}

// New builds a Service over a Store and the loaded catalog.
func New(st store.Store, cat *catalog.Catalog) *Service {
	return &Service{store: st, catalog: cat}
}

// ValidationError carries typed violations for an invalid request.
type ValidationError struct {
	Code       string
	Message    string
	Violations []validate.Violation
}

func (e *ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Code
}

// CreateCharacterInput is the character creation request.
type CreateCharacterInput struct {
	Name        string       `json:"name"`
	Class       domain.Class `json:"class"`
	Spec        domain.Spec  `json:"spec"`
	Race        domain.Race  `json:"race"`
	Level       int          `json:"level"`
	Professions []string     `json:"professions,omitempty"`
	Region      string       `json:"region,omitempty"`
	Realm       string       `json:"realm,omitempty"`
}

// CreateBuildInput creates a plan, optionally copying another build.
type CreateBuildInput struct {
	Name     string `json:"name"`
	CopyFrom string `json:"copyFrom,omitempty"`
}

// UpdateBuildInput merges a build. Omitted pointer fields are unchanged.
type UpdateBuildInput struct {
	Name      *string           `json:"name,omitempty"`
	Equipment *domain.Equipment `json:"equipment,omitempty"`
	Talents   *domain.Talents   `json:"talents,omitempty"`
	Version   int               `json:"version"`
}

// UpdateCharacterInput edits display identity and spec-level capabilities.
type UpdateCharacterInput struct {
	Name        *string      `json:"name,omitempty"`
	Spec        *domain.Spec `json:"spec,omitempty"`
	Race        *domain.Race `json:"race,omitempty"`
	Professions *[]string    `json:"professions,omitempty"`
	Region      *string      `json:"region,omitempty"`
	Realm       *string      `json:"realm,omitempty"`
}

// CreateCharacter validates identity and writes the character and its empty
// Current build in one transaction.
func (s *Service) CreateCharacter(ctx context.Context, input CreateCharacterInput) (domain.Character, error) {
	ch := domain.Character{
		ID:          uuid.NewString(),
		Name:        strings.TrimSpace(input.Name),
		Class:       input.Class,
		Spec:        input.Spec,
		Race:        input.Race,
		Level:       input.Level,
		Professions: normalizeProfessions(input.Professions),
		Region:      strings.TrimSpace(input.Region),
		Realm:       strings.TrimSpace(input.Realm),
	}
	if err := s.validateIdentity(ch); err != nil {
		return domain.Character{}, err
	}
	current := domain.Build{
		ID:              uuid.NewString(),
		CharacterID:     ch.ID,
		Kind:            "current",
		CatalogRevision: s.revision(),
		Equipment:       domain.Equipment{},
		Talents:         domain.Talents{},
		Source:          "manual",
	}
	if err := s.store.CreateCharacterWithCurrent(ctx, ch, current); err != nil {
		return domain.Character{}, err
	}
	return s.store.GetCharacter(ctx, ch.ID)
}

// GetCharacter returns one character.
func (s *Service) GetCharacter(ctx context.Context, id string) (domain.Character, error) {
	return s.store.GetCharacter(ctx, id)
}

// ListCharacters returns every character.
func (s *Service) ListCharacters(ctx context.Context) ([]domain.Character, error) {
	return s.store.ListCharacters(ctx)
}

// UpdateCharacter edits identity and re-reports capability warnings without
// deleting equipment.
func (s *Service) UpdateCharacter(ctx context.Context, id string, input UpdateCharacterInput) (domain.Character, []validate.Violation, error) {
	ch, err := s.store.GetCharacter(ctx, id)
	if err != nil {
		return domain.Character{}, nil, err
	}
	if input.Name != nil {
		ch.Name = strings.TrimSpace(*input.Name)
	}
	if input.Spec != nil {
		ch.Spec = *input.Spec
	}
	if input.Race != nil {
		ch.Race = *input.Race
	}
	if input.Professions != nil {
		ch.Professions = normalizeProfessions(*input.Professions)
	}
	if input.Region != nil {
		ch.Region = strings.TrimSpace(*input.Region)
	}
	if input.Realm != nil {
		ch.Realm = strings.TrimSpace(*input.Realm)
	}
	if err := s.validateIdentity(ch); err != nil {
		return domain.Character{}, nil, err
	}
	if err := s.store.UpdateCharacter(ctx, ch); err != nil {
		return domain.Character{}, nil, err
	}
	updated, err := s.store.GetCharacter(ctx, id)
	if err != nil {
		return domain.Character{}, nil, err
	}
	warnings, err := s.capabilityWarnings(ctx, updated)
	if err != nil {
		return domain.Character{}, nil, err
	}
	return updated, warnings, nil
}

// CreateBuild creates a plan, optionally copying another build.
func (s *Service) CreateBuild(ctx context.Context, characterID string, input CreateBuildInput) (domain.Build, error) {
	ch, err := s.store.GetCharacter(ctx, characterID)
	if err != nil {
		return domain.Build{}, err
	}
	name := strings.TrimSpace(input.Name)
	if err := validateBuildName(name); err != nil {
		return domain.Build{}, err
	}
	build := domain.Build{
		ID:              uuid.NewString(),
		CharacterID:     ch.ID,
		Kind:            "plan",
		Name:            name,
		CatalogRevision: s.revision(),
		Equipment:       domain.Equipment{},
		Talents:         domain.Talents{},
		Source:          "manual",
	}
	if input.CopyFrom != "" {
		source, err := s.store.GetBuild(ctx, input.CopyFrom)
		if err != nil {
			return domain.Build{}, err
		}
		if source.CharacterID != ch.ID {
			return domain.Build{}, domain.ErrNotFound
		}
		build.Equipment = cloneEquipment(source.Equipment)
		build.Talents = cloneTalents(source.Talents)
	}
	if err := s.store.CreateBuild(ctx, build); err != nil {
		return domain.Build{}, err
	}
	return s.store.GetBuild(ctx, build.ID)
}

// GetBuild returns one build.
func (s *Service) GetBuild(ctx context.Context, id string) (domain.Build, error) {
	return s.store.GetBuild(ctx, id)
}

// ListBuilds returns a character's builds, Current first.
func (s *Service) ListBuilds(ctx context.Context, characterID string) ([]domain.Build, error) {
	return s.store.ListBuilds(ctx, characterID)
}

// UpdateBuild merges edits, validates the result, and writes only when valid.
func (s *Service) UpdateBuild(ctx context.Context, id string, input UpdateBuildInput) (*domain.Build, []validate.Violation, error) {
	build, err := s.store.GetBuild(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	ch, err := s.store.GetCharacter(ctx, build.CharacterID)
	if err != nil {
		return nil, nil, err
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if err := validateBuildName(name); err != nil {
			return nil, nil, err
		}
		build.Name = name
	}
	if input.Equipment != nil {
		build.Equipment = cloneEquipment(*input.Equipment)
	}
	if input.Talents != nil {
		build.Talents = cloneTalents(*input.Talents)
	}

	violations, err := s.ValidateBuild(ch, build.Equipment, build.Talents)
	if err != nil {
		return nil, nil, err
	}
	if len(violations) > 0 {
		return nil, violations, &ValidationError{
			Code:       "invalid_build",
			Message:    "The build did not pass validation.",
			Violations: violations,
		}
	}

	build.CatalogRevision = s.revision()
	saved, err := s.store.UpdateBuild(ctx, build, input.Version)
	if err != nil {
		return nil, nil, err
	}
	return &saved, nil, nil
}

// CopyBuild duplicates any build as an independent plan.
func (s *Service) CopyBuild(ctx context.Context, id string) (domain.Build, error) {
	source, err := s.store.GetBuild(ctx, id)
	if err != nil {
		return domain.Build{}, err
	}
	name := strings.TrimSpace(source.Name)
	if name == "" {
		name = "Current"
	}
	name = truncateRunes(name+" copy", 40)
	build := domain.Build{
		ID:              uuid.NewString(),
		CharacterID:     source.CharacterID,
		Kind:            "plan",
		Name:            name,
		CatalogRevision: s.revision(),
		Equipment:       cloneEquipment(source.Equipment),
		Talents:         cloneTalents(source.Talents),
		Source:          source.Source,
	}
	if err := s.store.CreateBuild(ctx, build); err != nil {
		return domain.Build{}, err
	}
	return s.store.GetBuild(ctx, build.ID)
}

// DeleteBuild deletes a plan. The Current build is refused by the store.
func (s *Service) DeleteBuild(ctx context.Context, id string) error {
	return s.store.DeleteBuild(ctx, id)
}

// ValidateBuild validates an equipment and talent state without reading or
// writing the store.
func (s *Service) ValidateBuild(ch domain.Character, eq domain.Equipment, tal domain.Talents) ([]validate.Violation, error) {
	trees, ok := s.catalog.Talents[ch.Class]
	if !ok {
		return nil, domain.ErrUnsupportedClass
	}
	violations := validate.Equipment(ch, eq, s.catalog)
	for _, slot := range domain.SlotOrder {
		instance, present := eq[slot]
		if !present {
			continue
		}
		item := s.catalog.Items[instance.ItemID]
		violations = append(violations, validate.Enchant(ch, slot, item, instance.Enchant, s.catalog)...)
	}
	talentViolations, _ := validate.Talents(ch, tal, trees)
	violations = append(violations, talentViolations...)
	if violations == nil {
		violations = []validate.Violation{}
	}
	return violations, nil
}

func (s *Service) validateIdentity(ch domain.Character) *ValidationError {
	var violations []validate.Violation

	name := strings.TrimSpace(ch.Name)
	if count := utf8.RuneCountInString(name); count < 1 || count > 40 {
		violations = append(violations, validate.Violation{
			Code:     "invalid_character_name",
			Severity: "error",
			Message:  "Name must be between 1 and 40 characters.",
		})
	}

	specs, knownClass := s.catalog.Capabilities.Specs[ch.Class]
	if !knownClass {
		violations = append(violations, validate.Violation{
			Code:     "unsupported_class",
			Severity: "error",
			Message:  fmt.Sprintf("Class %q is not in the pinned catalog.", ch.Class),
		})
	} else {
		found := false
		for _, spec := range specs {
			if spec.Key == ch.Spec {
				found = true
				break
			}
		}
		if !found {
			violations = append(violations, validate.Violation{
				Code:     "unsupported_spec",
				Severity: "error",
				Message:  fmt.Sprintf("Specialization %q does not belong to %s.", ch.Spec, ch.Class),
			})
		}
	}
	if knownClass {
		races := s.catalog.Capabilities.Races[ch.Class]
		if !slices.Contains(races, ch.Race) {
			violations = append(violations, validate.Violation{
				Code:     "race_class_mismatch",
				Severity: "error",
				Message:  fmt.Sprintf("%s is not an available race for %s.", displayName(ch.Race), displayName(ch.Class)),
			})
		}
	}
	if ch.Level < 1 || ch.Level > 60 {
		violations = append(violations, validate.Violation{
			Code:     "invalid_level",
			Severity: "error",
			Message:  "Level must be between 1 and 60.",
		})
	}
	professions := normalizeProfessions(ch.Professions)
	if len(professions) > 2 {
		violations = append(violations, validate.Violation{
			Code:     "too_many_professions",
			Severity: "error",
			Message:  "A character can have at most two professions.",
		})
	}
	for _, profession := range professions {
		if !slices.Contains(s.catalog.Capabilities.Professions, profession) {
			violations = append(violations, validate.Violation{
				Code:     "invalid_profession",
				Severity: "error",
				Message:  fmt.Sprintf("Profession %q is not in the pinned catalog.", profession),
			})
		}
	}
	if len(violations) == 0 {
		return nil
	}
	return &ValidationError{
		Code:       violations[0].Code,
		Message:    violations[0].Message,
		Violations: violations,
	}
}

func (s *Service) capabilityWarnings(ctx context.Context, ch domain.Character) ([]validate.Violation, error) {
	builds, err := s.store.ListBuilds(ctx, ch.ID)
	if err != nil {
		return nil, err
	}
	warnings := make([]validate.Violation, 0)
	for _, build := range builds {
		for _, violation := range validate.Equipment(ch, build.Equipment, s.catalog) {
			violation.Severity = "warning"
			warnings = append(warnings, violation)
		}
	}
	return warnings, nil
}

func (s *Service) revision() string {
	if s.catalog == nil {
		return ""
	}
	return s.catalog.Revision()
}

func validateBuildName(name string) error {
	if count := utf8.RuneCountInString(name); count < 1 || count > 40 {
		return &ValidationError{
			Code:    "invalid_build_name",
			Message: "Build name must be between 1 and 40 characters.",
			Violations: []validate.Violation{{
				Code:     "invalid_build_name",
				Severity: "error",
				Message:  "Build name must be between 1 and 40 characters.",
			}},
		}
	}
	return nil
}

func normalizeProfessions(professions []string) []string {
	out := make([]string, 0, len(professions))
	for _, profession := range professions {
		profession = strings.TrimSpace(profession)
		if profession == "" || slices.Contains(out, profession) {
			continue
		}
		out = append(out, profession)
	}
	return out
}

func cloneEquipment(equipment domain.Equipment) domain.Equipment {
	cloned := make(domain.Equipment, len(equipment))
	for slot, instance := range equipment {
		if instance.Enchant != nil {
			enchant := *instance.Enchant
			instance.Enchant = &enchant
		}
		cloned[slot] = instance
	}
	return cloned
}

func cloneTalents(talents domain.Talents) domain.Talents {
	cloned := make(domain.Talents, len(talents))
	for id, points := range talents {
		cloned[id] = points
	}
	return cloned
}

func truncateRunes(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

func displayName(value string) string {
	if value == "" {
		return ""
	}
	spaced := strings.ReplaceAll(value, "_", " ")
	return strings.ToUpper(spaced[:1]) + spaced[1:]
}
