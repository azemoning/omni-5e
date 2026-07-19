package service

import (
	"context"
	"fmt"

	"github.com/azemoning/omni-5e/internal/domain"
	"github.com/azemoning/omni-5e/internal/repository"
)

// Service orchestrates repositories and applies business rules.
type Service struct {
	srdVersions      repository.SRDVersionRepository
	spells           repository.SpellRepository
	monsters         repository.MonsterRepository
	classes          repository.ClassRepository
	species          repository.SpeciesRepository
	backgrounds      repository.BackgroundRepository
	feats            repository.FeatRepository
	equipment        repository.EquipmentRepository
	magicItems       repository.MagicItemRepository
	conditions       repository.ConditionRepository
	glossaryTerms    repository.GlossaryTermRepository
	ruleSections     repository.RuleSectionRepository
	classFeatures    repository.ClassFeatureRepository
	classLevelTables repository.ClassLevelTableRepository
}

// New creates a Service with the given repositories.
func New(
	srdVersions repository.SRDVersionRepository,
	spells repository.SpellRepository,
	monsters repository.MonsterRepository,
	classes repository.ClassRepository,
	species repository.SpeciesRepository,
	backgrounds repository.BackgroundRepository,
	feats repository.FeatRepository,
	equipment repository.EquipmentRepository,
	magicItems repository.MagicItemRepository,
	conditions repository.ConditionRepository,
	glossaryTerms repository.GlossaryTermRepository,
	ruleSections repository.RuleSectionRepository,
	classFeatures repository.ClassFeatureRepository,
	classLevelTables repository.ClassLevelTableRepository,
) *Service {
	return &Service{
		srdVersions: srdVersions, spells: spells, monsters: monsters, classes: classes,
		species: species, backgrounds: backgrounds, feats: feats, equipment: equipment,
		magicItems: magicItems, conditions: conditions, glossaryTerms: glossaryTerms,
		ruleSections: ruleSections, classFeatures: classFeatures, classLevelTables: classLevelTables,
	}
}

func (s *Service) resolveSRDVersion(ctx context.Context, version string) (string, error) {
	if version != "" {
		return version, nil
	}
	def, err := s.srdVersions.GetDefaultSRDVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("getting default SRD version: %w", err)
	}
	return def.Version, nil
}

// --- SRD Version ---
func (s *Service) ListSRDVersions(ctx context.Context) ([]domain.SRDVersion, error) {
	return s.srdVersions.ListSRDVersions(ctx)
}

func (s *Service) GetSRDVersion(ctx context.Context, version string) (*domain.SRDVersion, error) {
	return s.srdVersions.GetSRDVersion(ctx, version)
}

// --- Spells ---
func (s *Service) GetSpell(ctx context.Context, srdVersion, slug string) (*domain.Spell, error) {
	v, err := s.resolveSRDVersion(ctx, srdVersion)
	if err != nil {
		return nil, err
	}
	return s.spells.GetSpell(ctx, v, slug)
}

func (s *Service) ListSpells(ctx context.Context, filter domain.SpellFilter) (*domain.Page, error) {
	v, err := s.resolveSRDVersion(ctx, filter.SRDVersion)
	if err != nil {
		return nil, err
	}
	filter.SRDVersion = v
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	return s.spells.ListSpells(ctx, filter)
}

// --- Monsters ---
func (s *Service) GetMonster(ctx context.Context, srdVersion, slug string) (*domain.Monster, error) {
	v, err := s.resolveSRDVersion(ctx, srdVersion)
	if err != nil {
		return nil, err
	}
	return s.monsters.GetMonster(ctx, v, slug)
}

func (s *Service) ListMonsters(ctx context.Context, filter domain.MonsterFilter) (*domain.Page, error) {
	v, err := s.resolveSRDVersion(ctx, filter.SRDVersion)
	if err != nil {
		return nil, err
	}
	filter.SRDVersion = v
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	return s.monsters.ListMonsters(ctx, filter)
}

// --- Classes ---
func (s *Service) GetClass(ctx context.Context, srdVersion, slug string) (*domain.Class, error) {
	v, err := s.resolveSRDVersion(ctx, srdVersion)
	if err != nil {
		return nil, err
	}
	return s.classes.GetClass(ctx, v, slug)
}

func (s *Service) ListClasses(ctx context.Context, params domain.ListParams) (*domain.Page, error) {
	v, err := s.resolveSRDVersion(ctx, params.SRDVersion)
	if err != nil {
		return nil, err
	}
	params.SRDVersion = v
	if params.Limit <= 0 || params.Limit > 200 {
		params.Limit = 50
	}
	return s.classes.ListClasses(ctx, params)
}

func (s *Service) GetSubclasses(ctx context.Context, srdVersion, classSlug string) ([]domain.Subclass, error) {
	v, err := s.resolveSRDVersion(ctx, srdVersion)
	if err != nil {
		return nil, err
	}
	return s.classes.GetSubclasses(ctx, v, classSlug)
}

func (s *Service) GetClassLevelTable(ctx context.Context, srdVersion, classSlug string, level int) (*domain.ClassLevelTableRow, error) {
	v, err := s.resolveSRDVersion(ctx, srdVersion)
	if err != nil {
		return nil, err
	}
	return s.classes.GetClassLevelTable(ctx, v, classSlug, level)
}

// --- Species ---
func (s *Service) GetSpecies(ctx context.Context, srdVersion, slug string) (*domain.Species, error) {
	v, err := s.resolveSRDVersion(ctx, srdVersion)
	if err != nil {
		return nil, err
	}
	return s.species.GetSpecies(ctx, v, slug)
}

func (s *Service) ListSpecies(ctx context.Context, filter domain.SpeciesFilter) (*domain.Page, error) {
	v, err := s.resolveSRDVersion(ctx, filter.SRDVersion)
	if err != nil {
		return nil, err
	}
	filter.SRDVersion = v
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	return s.species.ListSpecies(ctx, filter)
}

// --- Backgrounds ---
func (s *Service) GetBackground(ctx context.Context, srdVersion, slug string) (*domain.Background, error) {
	v, err := s.resolveSRDVersion(ctx, srdVersion)
	if err != nil {
		return nil, err
	}
	return s.backgrounds.GetBackground(ctx, v, slug)
}

func (s *Service) ListBackgrounds(ctx context.Context, params domain.ListParams) (*domain.Page, error) {
	v, err := s.resolveSRDVersion(ctx, params.SRDVersion)
	if err != nil {
		return nil, err
	}
	params.SRDVersion = v
	if params.Limit <= 0 || params.Limit > 200 {
		params.Limit = 50
	}
	return s.backgrounds.ListBackgrounds(ctx, params)
}

// --- Feats ---
func (s *Service) GetFeat(ctx context.Context, srdVersion, slug string) (*domain.Feat, error) {
	v, err := s.resolveSRDVersion(ctx, srdVersion)
	if err != nil {
		return nil, err
	}
	return s.feats.GetFeat(ctx, v, slug)
}

func (s *Service) ListFeats(ctx context.Context, filter domain.FeatFilter) (*domain.Page, error) {
	v, err := s.resolveSRDVersion(ctx, filter.SRDVersion)
	if err != nil {
		return nil, err
	}
	filter.SRDVersion = v
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	return s.feats.ListFeats(ctx, filter)
}

// --- Equipment ---
func (s *Service) GetEquipment(ctx context.Context, srdVersion, slug string) (*domain.Equipment, error) {
	v, err := s.resolveSRDVersion(ctx, srdVersion)
	if err != nil {
		return nil, err
	}
	return s.equipment.GetEquipment(ctx, v, slug)
}

func (s *Service) ListEquipment(ctx context.Context, filter domain.EquipmentFilter) (*domain.Page, error) {
	v, err := s.resolveSRDVersion(ctx, filter.SRDVersion)
	if err != nil {
		return nil, err
	}
	filter.SRDVersion = v
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	return s.equipment.ListEquipment(ctx, filter)
}

// --- Magic Items ---
func (s *Service) GetMagicItem(ctx context.Context, srdVersion, slug string) (*domain.MagicItem, error) {
	v, err := s.resolveSRDVersion(ctx, srdVersion)
	if err != nil {
		return nil, err
	}
	return s.magicItems.GetMagicItem(ctx, v, slug)
}

func (s *Service) ListMagicItems(ctx context.Context, filter domain.MagicItemFilter) (*domain.Page, error) {
	v, err := s.resolveSRDVersion(ctx, filter.SRDVersion)
	if err != nil {
		return nil, err
	}
	filter.SRDVersion = v
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	return s.magicItems.ListMagicItems(ctx, filter)
}

// --- Conditions ---
func (s *Service) GetCondition(ctx context.Context, srdVersion, slug string) (*domain.Condition, error) {
	v, err := s.resolveSRDVersion(ctx, srdVersion)
	if err != nil {
		return nil, err
	}
	return s.conditions.GetCondition(ctx, v, slug)
}

func (s *Service) ListConditions(ctx context.Context, params domain.ListParams) (*domain.Page, error) {
	v, err := s.resolveSRDVersion(ctx, params.SRDVersion)
	if err != nil {
		return nil, err
	}
	params.SRDVersion = v
	if params.Limit <= 0 || params.Limit > 200 {
		params.Limit = 50
	}
	return s.conditions.ListConditions(ctx, params)
}

// --- Glossary Terms ---
func (s *Service) GetGlossaryTerm(ctx context.Context, srdVersion, slug string) (*domain.GlossaryTerm, error) {
	v, err := s.resolveSRDVersion(ctx, srdVersion)
	if err != nil {
		return nil, err
	}
	return s.glossaryTerms.GetGlossaryTerm(ctx, v, slug)
}

func (s *Service) ListGlossaryTerms(ctx context.Context, filter domain.GlossaryFilter) (*domain.Page, error) {
	v, err := s.resolveSRDVersion(ctx, filter.SRDVersion)
	if err != nil {
		return nil, err
	}
	filter.SRDVersion = v
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	return s.glossaryTerms.ListGlossaryTerms(ctx, filter)
}

// --- Rule Sections ---
func (s *Service) GetRuleSection(ctx context.Context, srdVersion, slug string) (*domain.RuleSection, error) {
	v, err := s.resolveSRDVersion(ctx, srdVersion)
	if err != nil {
		return nil, err
	}
	return s.ruleSections.GetRuleSection(ctx, v, slug)
}

func (s *Service) ListRuleSections(ctx context.Context, filter domain.RuleSectionFilter) (*domain.Page, error) {
	v, err := s.resolveSRDVersion(ctx, filter.SRDVersion)
	if err != nil {
		return nil, err
	}
	filter.SRDVersion = v
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	return s.ruleSections.ListRuleSections(ctx, filter)
}
