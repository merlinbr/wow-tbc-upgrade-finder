// Package store persists Forever characters and builds.
package store

import (
	"context"
	"errors"

	"github.com/wowsims/tbc/internal/forever/domain"
)

// Store is the persistence boundary. Phase 3 adds ownership and publication on
// this interface without changing the domain.
type Store interface {
	CreateCharacterWithCurrent(ctx context.Context, ch domain.Character, current domain.Build) error
	GetCharacter(ctx context.Context, id string) (domain.Character, error)
	ListCharacters(ctx context.Context) ([]domain.Character, error)
	UpdateCharacter(ctx context.Context, ch domain.Character) error
	GetBuild(ctx context.Context, id string) (domain.Build, error)
	ListBuilds(ctx context.Context, characterID string) ([]domain.Build, error)
	CreateBuild(ctx context.Context, b domain.Build) error
	UpdateBuild(ctx context.Context, b domain.Build, expectedVersion int) (domain.Build, error)
	DeleteBuild(ctx context.Context, id string) error
}

// Store errors shared with the service and server layers.
var (
	// ErrNotFound means the record does not exist.
	ErrNotFound = domain.ErrNotFound
	// ErrVersionConflict means the build was written by another edit.
	ErrVersionConflict = errors.New("forever: build version conflict")
	// ErrCannotDeleteCurrent means the Current build is permanent.
	ErrCannotDeleteCurrent = errors.New("forever: cannot delete the current build")
)
