package repository

import (
	"context"

	"github.com/azemoning/omni-5e/internal/domain"
)

// SRDVersionRepository manages SRD version metadata.
type SRDVersionRepository interface {
	GetSRDVersion(ctx context.Context, version string) (*domain.SRDVersion, error)
	GetDefaultSRDVersion(ctx context.Context) (*domain.SRDVersion, error)
	ListSRDVersions(ctx context.Context) ([]domain.SRDVersion, error)
	UpsertSRDVersion(ctx context.Context, v *domain.SRDVersion) error
}

// SpellRepository manages spell data.
type SpellRepository interface {
	GetSpell(ctx context.Context, srdVersion, slug string) (*domain.Spell, error)
	ListSpells(ctx context.Context, filter domain.SpellFilter) (*domain.Page, error)
	UpsertSpell(ctx context.Context, spell *domain.Spell) error
	UpsertSpells(ctx context.Context, spells []domain.Spell) error
}

// MonsterRepository manages monster data.
type MonsterRepository interface {
	GetMonster(ctx context.Context, srdVersion, slug string) (*domain.Monster, error)
	ListMonsters(ctx context.Context, filter domain.MonsterFilter) (*domain.Page, error)
	UpsertMonster(ctx context.Context, monster *domain.Monster) error
	UpsertMonsters(ctx context.Context, monsters []domain.Monster) error
}

// ClassRepository manages class data.
type ClassRepository interface {
	GetClass(ctx context.Context, srdVersion, slug string) (*domain.Class, error)
	ListClasses(ctx context.Context, params domain.ListParams) (*domain.Page, error)
	GetSubclasses(ctx context.Context, srdVersion, classSlug string) ([]domain.Subclass, error)
	GetClassLevelTable(ctx context.Context, srdVersion, classSlug string, level int) (*domain.ClassLevelTableRow, error)
	UpsertClass(ctx context.Context, class *domain.Class) error
}

// SpeciesRepository manages species data.
type SpeciesRepository interface {
	GetSpecies(ctx context.Context, srdVersion, slug string) (*domain.Species, error)
	ListSpecies(ctx context.Context, filter domain.SpeciesFilter) (*domain.Page, error)
	UpsertSpecies(ctx context.Context, species *domain.Species) error
}

// BackgroundRepository manages background data.
type BackgroundRepository interface {
	GetBackground(ctx context.Context, srdVersion, slug string) (*domain.Background, error)
	ListBackgrounds(ctx context.Context, params domain.ListParams) (*domain.Page, error)
	UpsertBackground(ctx context.Context, bg *domain.Background) error
}

// FeatRepository manages feat data.
type FeatRepository interface {
	GetFeat(ctx context.Context, srdVersion, slug string) (*domain.Feat, error)
	ListFeats(ctx context.Context, filter domain.FeatFilter) (*domain.Page, error)
	UpsertFeat(ctx context.Context, feat *domain.Feat) error
}

// EquipmentRepository manages equipment data.
type EquipmentRepository interface {
	GetEquipment(ctx context.Context, srdVersion, slug string) (*domain.Equipment, error)
	ListEquipment(ctx context.Context, filter domain.EquipmentFilter) (*domain.Page, error)
	UpsertEquipment(ctx context.Context, equip *domain.Equipment) error
}

// MagicItemRepository manages magic item data.
type MagicItemRepository interface {
	GetMagicItem(ctx context.Context, srdVersion, slug string) (*domain.MagicItem, error)
	ListMagicItems(ctx context.Context, filter domain.MagicItemFilter) (*domain.Page, error)
	UpsertMagicItem(ctx context.Context, item *domain.MagicItem) error
}

// ConditionRepository manages condition data.
type ConditionRepository interface {
	GetCondition(ctx context.Context, srdVersion, slug string) (*domain.Condition, error)
	ListConditions(ctx context.Context, params domain.ListParams) (*domain.Page, error)
	UpsertCondition(ctx context.Context, cond *domain.Condition) error
}

// GlossaryTermRepository manages glossary term data.
type GlossaryTermRepository interface {
	GetGlossaryTerm(ctx context.Context, srdVersion, slug string) (*domain.GlossaryTerm, error)
	ListGlossaryTerms(ctx context.Context, filter domain.GlossaryFilter) (*domain.Page, error)
	UpsertGlossaryTerm(ctx context.Context, term *domain.GlossaryTerm) error
}

// RuleSectionRepository manages rule section data.
type RuleSectionRepository interface {
	GetRuleSection(ctx context.Context, srdVersion, slug string) (*domain.RuleSection, error)
	ListRuleSections(ctx context.Context, filter domain.RuleSectionFilter) (*domain.Page, error)
	UpsertRuleSection(ctx context.Context, section *domain.RuleSection) error
}

// ClassFeatureRepository manages class feature data.
type ClassFeatureRepository interface {
	GetClassFeaturesByClass(ctx context.Context, srdVersion, classSlug string) ([]domain.ClassFeature, error)
	GetClassFeaturesByClassAndLevel(ctx context.Context, srdVersion, classSlug string, level int) ([]domain.ClassFeature, error)
	UpsertClassFeature(ctx context.Context, feature *domain.ClassFeature) error
	UpsertClassFeatures(ctx context.Context, features []domain.ClassFeature) error
}

// ClassLevelTableRepository manages class level table data.
type ClassLevelTableRepository interface {
	GetClassLevelTableByClass(ctx context.Context, srdVersion, classSlug string) ([]domain.ClassLevelTableRow, error)
	GetClassLevelTableByClassAndLevel(ctx context.Context, srdVersion, classSlug string, level int) (*domain.ClassLevelTableRow, error)
	UpsertClassLevelTableRow(ctx context.Context, row *domain.ClassLevelTableRow) error
	UpsertClassLevelTableRows(ctx context.Context, rows []domain.ClassLevelTableRow) error
}
