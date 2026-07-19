package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/azemoning/omni-5e/internal/repository"
)

// Store implements all repository interfaces using PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
}

// Verify interface compliance.
var (
	_ repository.SRDVersionRepository      = (*Store)(nil)
	_ repository.SpellRepository           = (*Store)(nil)
	_ repository.MonsterRepository         = (*Store)(nil)
	_ repository.ClassRepository           = (*Store)(nil)
	_ repository.SpeciesRepository         = (*Store)(nil)
	_ repository.BackgroundRepository      = (*Store)(nil)
	_ repository.FeatRepository            = (*Store)(nil)
	_ repository.EquipmentRepository       = (*Store)(nil)
	_ repository.MagicItemRepository       = (*Store)(nil)
	_ repository.ConditionRepository       = (*Store)(nil)
	_ repository.GlossaryTermRepository    = (*Store)(nil)
	_ repository.RuleSectionRepository     = (*Store)(nil)
	_ repository.ClassFeatureRepository    = (*Store)(nil)
	_ repository.ClassLevelTableRepository = (*Store)(nil)
)

// New creates a new Store with the given connection pool.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}
