package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"github.com/wowsims/tbc/internal/forever/domain"
)

// SQLite is the Phase 2 Store implementation.
type SQLite struct {
	db *sql.DB
}

// NewSQLite opens the database, enables foreign keys and creates the schema.
func NewSQLite(path string) (*SQLite, error) {
	if path == "" {
		path = ":memory:"
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}
	store := &SQLite{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

// Close releases the database handle.
func (s *SQLite) Close() error {
	return s.db.Close()
}

func (s *SQLite) migrate() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS characters (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			class TEXT NOT NULL,
			spec TEXT NOT NULL,
			race TEXT NOT NULL,
			level INTEGER NOT NULL,
			professions TEXT NOT NULL,
			region TEXT NOT NULL DEFAULT '',
			realm TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS builds (
			id TEXT PRIMARY KEY,
			character_id TEXT NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
			kind TEXT NOT NULL,
			name TEXT NOT NULL DEFAULT '',
			catalog_revision TEXT NOT NULL,
			equipment TEXT NOT NULL,
			talents TEXT NOT NULL,
			source TEXT NOT NULL,
			version INTEGER NOT NULL DEFAULT 1,
			published INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS builds_one_current
			ON builds(character_id) WHERE kind = 'current'`,
		`CREATE INDEX IF NOT EXISTS builds_by_character ON builds(character_id)`,
	}
	for _, statement := range statements {
		if _, err := s.db.Exec(statement); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

func (s *SQLite) CreateCharacterWithCurrent(ctx context.Context, ch domain.Character, current domain.Build) error {
	if ch.ID == "" {
		ch.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if ch.CreatedAt.IsZero() {
		ch.CreatedAt = now
	}
	ch.UpdatedAt = now
	professions, err := json.Marshal(nonNilStrings(ch.Professions))
	if err != nil {
		return fmt.Errorf("encode professions: %w", err)
	}

	current.CharacterID = ch.ID
	if current.ID == "" {
		current.ID = uuid.NewString()
	}
	current.Kind = "current"
	current.Version = 1
	if current.Source == "" {
		current.Source = "manual"
	}
	if current.Equipment == nil {
		current.Equipment = domain.Equipment{}
	}
	if current.Talents == nil {
		current.Talents = domain.Talents{}
	}
	if current.CreatedAt.IsZero() {
		current.CreatedAt = now
	}
	current.UpdatedAt = now
	equipment, err := encodeEquipment(current.Equipment)
	if err != nil {
		return err
	}
	talents, err := encodeTalents(current.Talents)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin character transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO characters
			(id, name, class, spec, race, level, professions, region, realm, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		ch.ID, ch.Name, ch.Class, ch.Spec, ch.Race, ch.Level, string(professions),
		ch.Region, ch.Realm, formatTime(ch.CreatedAt), formatTime(ch.UpdatedAt),
	); err != nil {
		return fmt.Errorf("insert character: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO builds
			(id, character_id, kind, name, catalog_revision, equipment, talents, source,
			 version, published, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		current.ID, current.CharacterID, current.Kind, current.Name, current.CatalogRevision,
		equipment, talents, current.Source, current.Version, boolInt(current.Published),
		formatTime(current.CreatedAt), formatTime(current.UpdatedAt),
	); err != nil {
		return fmt.Errorf("insert current build: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit character: %w", err)
	}
	return nil
}

func (s *SQLite) GetCharacter(ctx context.Context, id string) (domain.Character, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, class, spec, race, level, professions, region, realm, created_at, updated_at
		 FROM characters WHERE id = ?`, id)
	return scanCharacter(row)
}

func (s *SQLite) ListCharacters(ctx context.Context) ([]domain.Character, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, class, spec, race, level, professions, region, realm, created_at, updated_at
		 FROM characters ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	characters := make([]domain.Character, 0)
	for rows.Next() {
		ch, err := scanCharacter(rows)
		if err != nil {
			return nil, err
		}
		characters = append(characters, ch)
	}
	return characters, rows.Err()
}

func (s *SQLite) UpdateCharacter(ctx context.Context, ch domain.Character) error {
	professions, err := json.Marshal(nonNilStrings(ch.Professions))
	if err != nil {
		return fmt.Errorf("encode professions: %w", err)
	}
	updatedAt := time.Now().UTC()
	result, err := s.db.ExecContext(ctx,
		`UPDATE characters
		 SET name = ?, class = ?, spec = ?, race = ?, level = ?, professions = ?,
		     region = ?, realm = ?, updated_at = ?
		 WHERE id = ?`,
		ch.Name, ch.Class, ch.Spec, ch.Race, ch.Level, string(professions),
		ch.Region, ch.Realm, formatTime(updatedAt), ch.ID,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLite) GetBuild(ctx context.Context, id string) (domain.Build, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, character_id, kind, name, catalog_revision, equipment, talents, source,
		        version, published, created_at, updated_at
		 FROM builds WHERE id = ?`, id)
	return scanBuild(row)
}

func (s *SQLite) ListBuilds(ctx context.Context, characterID string) ([]domain.Build, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, character_id, kind, name, catalog_revision, equipment, talents, source,
		        version, published, created_at, updated_at
		 FROM builds WHERE character_id = ?
		 ORDER BY CASE kind WHEN 'current' THEN 0 ELSE 1 END, created_at, id`, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	builds := make([]domain.Build, 0)
	for rows.Next() {
		build, err := scanBuild(rows)
		if err != nil {
			return nil, err
		}
		builds = append(builds, build)
	}
	return builds, rows.Err()
}

func (s *SQLite) CreateBuild(ctx context.Context, b domain.Build) error {
	if b.ID == "" {
		b.ID = uuid.NewString()
	}
	if b.Kind == "" {
		b.Kind = "plan"
	}
	if b.Source == "" {
		b.Source = "manual"
	}
	if b.Equipment == nil {
		b.Equipment = domain.Equipment{}
	}
	if b.Talents == nil {
		b.Talents = domain.Talents{}
	}
	now := time.Now().UTC()
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
	b.UpdatedAt = now
	b.Version = 1
	equipment, err := encodeEquipment(b.Equipment)
	if err != nil {
		return err
	}
	talents, err := encodeTalents(b.Talents)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO builds
			(id, character_id, kind, name, catalog_revision, equipment, talents, source,
			 version, published, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		b.ID, b.CharacterID, b.Kind, b.Name, b.CatalogRevision, equipment, talents,
		b.Source, b.Version, boolInt(b.Published), formatTime(b.CreatedAt), formatTime(b.UpdatedAt),
	)
	return err
}

func (s *SQLite) UpdateBuild(ctx context.Context, b domain.Build, expectedVersion int) (domain.Build, error) {
	equipment, err := encodeEquipment(b.Equipment)
	if err != nil {
		return domain.Build{}, err
	}
	talents, err := encodeTalents(b.Talents)
	if err != nil {
		return domain.Build{}, err
	}
	result, err := s.db.ExecContext(ctx,
		`UPDATE builds
		 SET name = ?, catalog_revision = ?, equipment = ?, talents = ?, source = ?,
		     published = ?, version = version + 1, updated_at = ?
		 WHERE id = ? AND version = ?`,
		b.Name, b.CatalogRevision, equipment, talents, b.Source, boolInt(b.Published),
		formatTime(time.Now().UTC()), b.ID, expectedVersion,
	)
	if err != nil {
		return domain.Build{}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return domain.Build{}, err
	}
	if affected == 0 {
		var exists int
		err := s.db.QueryRowContext(ctx, `SELECT 1 FROM builds WHERE id = ?`, b.ID).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Build{}, ErrNotFound
		}
		if err != nil {
			return domain.Build{}, err
		}
		return domain.Build{}, ErrVersionConflict
	}
	return s.GetBuild(ctx, b.ID)
}

func (s *SQLite) DeleteBuild(ctx context.Context, id string) error {
	var kind string
	err := s.db.QueryRowContext(ctx, `SELECT kind FROM builds WHERE id = ?`, id).Scan(&kind)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if kind == "current" {
		return ErrCannotDeleteCurrent
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM builds WHERE id = ?`, id); err != nil {
		return err
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanCharacter(row scanner) (domain.Character, error) {
	var ch domain.Character
	var professions, createdAt, updatedAt string
	err := row.Scan(
		&ch.ID, &ch.Name, &ch.Class, &ch.Spec, &ch.Race, &ch.Level, &professions,
		&ch.Region, &ch.Realm, &createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Character{}, ErrNotFound
	}
	if err != nil {
		return domain.Character{}, err
	}
	if err := json.Unmarshal([]byte(professions), &ch.Professions); err != nil {
		return domain.Character{}, fmt.Errorf("decode professions: %w", err)
	}
	ch.CreatedAt = parseTime(createdAt)
	ch.UpdatedAt = parseTime(updatedAt)
	return ch, nil
}

func scanBuild(row scanner) (domain.Build, error) {
	var build domain.Build
	var equipment, talents, createdAt, updatedAt string
	var published int
	err := row.Scan(
		&build.ID, &build.CharacterID, &build.Kind, &build.Name, &build.CatalogRevision,
		&equipment, &talents, &build.Source, &build.Version, &published,
		&createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Build{}, ErrNotFound
	}
	if err != nil {
		return domain.Build{}, err
	}
	decodedEquipment, err := decodeEquipment(equipment)
	if err != nil {
		return domain.Build{}, err
	}
	decodedTalents, err := decodeTalents(talents)
	if err != nil {
		return domain.Build{}, err
	}
	build.Equipment = decodedEquipment
	build.Talents = decodedTalents
	build.Published = published != 0
	build.CreatedAt = parseTime(createdAt)
	build.UpdatedAt = parseTime(updatedAt)
	return build, nil
}

func encodeEquipment(equipment domain.Equipment) (string, error) {
	if equipment == nil {
		equipment = domain.Equipment{}
	}
	data, err := json.Marshal(equipment)
	if err != nil {
		return "", fmt.Errorf("encode equipment: %w", err)
	}
	return string(data), nil
}

func decodeEquipment(raw string) (domain.Equipment, error) {
	equipment := make(domain.Equipment)
	if raw == "" || raw == "null" {
		return equipment, nil
	}
	if err := json.Unmarshal([]byte(raw), &equipment); err != nil {
		return nil, fmt.Errorf("decode equipment: %w", err)
	}
	if equipment == nil {
		equipment = make(domain.Equipment)
	}
	return equipment, nil
}

func encodeTalents(talents domain.Talents) (string, error) {
	if talents == nil {
		talents = domain.Talents{}
	}
	data, err := json.Marshal(talents)
	if err != nil {
		return "", fmt.Errorf("encode talents: %w", err)
	}
	return string(data), nil
}

func decodeTalents(raw string) (domain.Talents, error) {
	talents := make(domain.Talents)
	if raw == "" || raw == "null" {
		return talents, nil
	}
	if err := json.Unmarshal([]byte(raw), &talents); err != nil {
		return nil, fmt.Errorf("decode talents: %w", err)
	}
	if talents == nil {
		talents = make(domain.Talents)
	}
	return talents, nil
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
