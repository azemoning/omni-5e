package service_test

import (
	"context"
	"testing"

	"github.com/azemoning/omni-5e/internal/domain"
	"github.com/azemoning/omni-5e/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock repositories ---

type mockSRDVersionRepo struct {
	versions []domain.SRDVersion
}

func (m *mockSRDVersionRepo) GetSRDVersion(_ context.Context, version string) (*domain.SRDVersion, error) {
	for _, v := range m.versions {
		if v.Version == version {
			return &v, nil
		}
	}
	return nil, nil
}

func (m *mockSRDVersionRepo) GetDefaultSRDVersion(_ context.Context) (*domain.SRDVersion, error) {
	for _, v := range m.versions {
		if v.IsDefault {
			return &v, nil
		}
	}
	return nil, nil
}

func (m *mockSRDVersionRepo) ListSRDVersions(_ context.Context) ([]domain.SRDVersion, error) {
	return m.versions, nil
}

func (m *mockSRDVersionRepo) UpsertSRDVersion(_ context.Context, v *domain.SRDVersion) error {
	m.versions = append(m.versions, *v)
	return nil
}

type mockSpellRepo struct {
	spells []domain.Spell
}

func (m *mockSpellRepo) GetSpell(_ context.Context, srdVersion, slug string) (*domain.Spell, error) {
	for _, s := range m.spells {
		if s.SRDVersion == srdVersion && s.Slug == slug {
			return &s, nil
		}
	}
	return nil, nil
}

func (m *mockSpellRepo) ListSpells(_ context.Context, filter domain.SpellFilter) (*domain.Page, error) {
	var items []domain.Spell
	for _, s := range m.spells {
		if s.SRDVersion == filter.SRDVersion {
			items = append(items, s)
		}
	}
	return &domain.Page{
		Items:      items,
		Total:      len(items),
		Limit:      filter.Limit,
		Offset:     filter.Offset,
		SRDVersion: filter.SRDVersion,
	}, nil
}

func (m *mockSpellRepo) UpsertSpell(_ context.Context, spell *domain.Spell) error {
	m.spells = append(m.spells, *spell)
	return nil
}

func (m *mockSpellRepo) UpsertSpells(_ context.Context, spells []domain.Spell) error {
	m.spells = append(m.spells, spells...)
	return nil
}

// Stub repos for unused interfaces
type stubRepo struct{}

func (s *stubRepo) GetMonster(context.Context, string, string) (*domain.Monster, error) {
	return nil, nil
}
func (s *stubRepo) ListMonsters(context.Context, domain.MonsterFilter) (*domain.Page, error) {
	return &domain.Page{}, nil
}
func (s *stubRepo) UpsertMonster(context.Context, *domain.Monster) error { return nil }
func (s *stubRepo) UpsertMonsters(context.Context, []domain.Monster) error { return nil }
func (s *stubRepo) GetClass(context.Context, string, string) (*domain.Class, error) {
	return nil, nil
}
func (s *stubRepo) ListClasses(context.Context, domain.ListParams) (*domain.Page, error) {
	return &domain.Page{}, nil
}
func (s *stubRepo) GetSubclasses(context.Context, string, string) ([]domain.Subclass, error) {
	return nil, nil
}
func (s *stubRepo) GetClassLevelTable(context.Context, string, string, int) (*domain.ClassLevelTableRow, error) {
	return nil, nil
}
func (s *stubRepo) UpsertClass(context.Context, *domain.Class) error { return nil }
func (s *stubRepo) GetSpecies(context.Context, string, string) (*domain.Species, error) {
	return nil, nil
}
func (s *stubRepo) ListSpecies(context.Context, domain.SpeciesFilter) (*domain.Page, error) {
	return &domain.Page{}, nil
}
func (s *stubRepo) UpsertSpecies(context.Context, *domain.Species) error { return nil }
func (s *stubRepo) GetBackground(context.Context, string, string) (*domain.Background, error) {
	return nil, nil
}
func (s *stubRepo) ListBackgrounds(context.Context, domain.ListParams) (*domain.Page, error) {
	return &domain.Page{}, nil
}
func (s *stubRepo) UpsertBackground(context.Context, *domain.Background) error { return nil }
func (s *stubRepo) GetFeat(context.Context, string, string) (*domain.Feat, error) {
	return nil, nil
}
func (s *stubRepo) ListFeats(context.Context, domain.FeatFilter) (*domain.Page, error) {
	return &domain.Page{}, nil
}
func (s *stubRepo) UpsertFeat(context.Context, *domain.Feat) error { return nil }
func (s *stubRepo) GetEquipment(context.Context, string, string) (*domain.Equipment, error) {
	return nil, nil
}
func (s *stubRepo) ListEquipment(context.Context, domain.EquipmentFilter) (*domain.Page, error) {
	return &domain.Page{}, nil
}
func (s *stubRepo) UpsertEquipment(context.Context, *domain.Equipment) error { return nil }
func (s *stubRepo) GetMagicItem(context.Context, string, string) (*domain.MagicItem, error) {
	return nil, nil
}
func (s *stubRepo) ListMagicItems(context.Context, domain.MagicItemFilter) (*domain.Page, error) {
	return &domain.Page{}, nil
}
func (s *stubRepo) UpsertMagicItem(context.Context, *domain.MagicItem) error { return nil }
func (s *stubRepo) GetCondition(context.Context, string, string) (*domain.Condition, error) {
	return nil, nil
}
func (s *stubRepo) ListConditions(context.Context, domain.ListParams) (*domain.Page, error) {
	return &domain.Page{}, nil
}
func (s *stubRepo) UpsertCondition(context.Context, *domain.Condition) error { return nil }
func (s *stubRepo) GetGlossaryTerm(context.Context, string, string) (*domain.GlossaryTerm, error) {
	return nil, nil
}
func (s *stubRepo) ListGlossaryTerms(context.Context, domain.GlossaryFilter) (*domain.Page, error) {
	return &domain.Page{}, nil
}
func (s *stubRepo) UpsertGlossaryTerm(context.Context, *domain.GlossaryTerm) error { return nil }
func (s *stubRepo) GetRuleSection(context.Context, string, string) (*domain.RuleSection, error) {
	return nil, nil
}
func (s *stubRepo) ListRuleSections(context.Context, domain.RuleSectionFilter) (*domain.Page, error) {
	return &domain.Page{}, nil
}
func (s *stubRepo) UpsertRuleSection(context.Context, *domain.RuleSection) error { return nil }
func (s *stubRepo) GetClassFeaturesByClass(context.Context, string, string) ([]domain.ClassFeature, error) {
	return nil, nil
}
func (s *stubRepo) GetClassFeaturesByClassAndLevel(context.Context, string, string, int) ([]domain.ClassFeature, error) {
	return nil, nil
}
func (s *stubRepo) UpsertClassFeature(context.Context, *domain.ClassFeature) error { return nil }
func (s *stubRepo) UpsertClassFeatures(context.Context, []domain.ClassFeature) error { return nil }
func (s *stubRepo) GetClassLevelTableByClass(context.Context, string, string) ([]domain.ClassLevelTableRow, error) {
	return nil, nil
}
func (s *stubRepo) GetClassLevelTableByClassAndLevel(context.Context, string, string, int) (*domain.ClassLevelTableRow, error) {
	return nil, nil
}
func (s *stubRepo) UpsertClassLevelTableRow(context.Context, *domain.ClassLevelTableRow) error {
	return nil
}
func (s *stubRepo) UpsertClassLevelTableRows(context.Context, []domain.ClassLevelTableRow) error {
	return nil
}

// --- Tests ---

func newTestService(srdVersions []domain.SRDVersion, spells []domain.Spell) *service.Service {
	srd := &mockSRDVersionRepo{versions: srdVersions}
	sp := &mockSpellRepo{spells: spells}
	stub := &stubRepo{}
	return service.New(srd, sp, stub, stub, stub, stub, stub, stub, stub, stub, stub, stub, stub, stub)
}

func TestListSRDVersions(t *testing.T) {
	versions := []domain.SRDVersion{
		{ID: uuid.New(), Version: "5.2.1", IsDefault: true},
	}
	svc := newTestService(versions, nil)

	result, err := svc.ListSRDVersions(context.Background())
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "5.2.1", result[0].Version)
}

func TestGetSpell(t *testing.T) {
	versions := []domain.SRDVersion{
		{ID: uuid.New(), Version: "5.2.1", IsDefault: true},
	}
	spells := []domain.Spell{
		{
			BaseEntity: domain.BaseEntity{
				ID:         uuid.New(),
				Slug:       "fireball",
				SRDVersion: "5.2.1",
				Name:       "Fireball",
			},
			Level:  3,
			School: "evocation",
		},
	}
	svc := newTestService(versions, spells)

	// Test explicit version
	spell, err := svc.GetSpell(context.Background(), "5.2.1", "fireball")
	require.NoError(t, err)
	assert.NotNil(t, spell)
	assert.Equal(t, "Fireball", spell.Name)
	assert.Equal(t, 3, spell.Level)

	// Test default version
	spell, err = svc.GetSpell(context.Background(), "", "fireball")
	require.NoError(t, err)
	assert.NotNil(t, spell)

	// Test not found
	spell, err = svc.GetSpell(context.Background(), "5.2.1", "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, spell)
}

func TestListSpells(t *testing.T) {
	versions := []domain.SRDVersion{
		{ID: uuid.New(), Version: "5.2.1", IsDefault: true},
	}
	spells := []domain.Spell{
		{BaseEntity: domain.BaseEntity{Slug: "fireball", SRDVersion: "5.2.1", Name: "Fireball"}, Level: 3},
		{BaseEntity: domain.BaseEntity{Slug: "wish", SRDVersion: "5.2.1", Name: "Wish"}, Level: 9},
	}
	svc := newTestService(versions, spells)

	page, err := svc.ListSpells(context.Background(), domain.SpellFilter{
		ListParams: domain.ListParams{Limit: 50, SRDVersion: "5.2.1"},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, page.Total)
}

func TestListSpellsDefaultVersion(t *testing.T) {
	versions := []domain.SRDVersion{
		{ID: uuid.New(), Version: "5.2.1", IsDefault: true},
	}
	spells := []domain.Spell{
		{BaseEntity: domain.BaseEntity{Slug: "fireball", SRDVersion: "5.2.1", Name: "Fireball"}},
	}
	svc := newTestService(versions, spells)

	page, err := svc.ListSpells(context.Background(), domain.SpellFilter{
		ListParams: domain.ListParams{Limit: 50},
	})
	require.NoError(t, err)
	assert.Equal(t, 1, page.Total)
}

func TestListSpellsClampLimit(t *testing.T) {
	versions := []domain.SRDVersion{
		{ID: uuid.New(), Version: "5.2.1", IsDefault: true},
	}
	svc := newTestService(versions, nil)

	// Limit > 200 should be clamped to50
	page, err := svc.ListSpells(context.Background(), domain.SpellFilter{
		ListParams: domain.ListParams{Limit: 999, SRDVersion: "5.2.1"},
	})
	require.NoError(t, err)
	assert.Equal(t, 50, page.Limit)
}
