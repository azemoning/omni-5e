package ingest

import (
	"context"

	"github.com/azemoning/omni-5e/internal/domain"
)

// ContentParser is the interface each SRD version adapter must implement.
// One method per entity family.
type ContentParser interface {
	Version() string
	ParseSpells(ctx context.Context, sourceDir string) ([]domain.Spell, error)
	ParseMonsters(ctx context.Context, sourceDir string) ([]domain.Monster, error)
	ParseClasses(ctx context.Context, sourceDir string) ([]domain.Class, error)
	ParseClassFeatures(ctx context.Context, sourceDir string) ([]domain.ClassFeature, error)
	ParseClassLevelTables(ctx context.Context, sourceDir string) ([]domain.ClassLevelTableRow, error)
	ParseSubclasses(ctx context.Context, sourceDir string) ([]domain.Subclass, error)
	ParseSpecies(ctx context.Context, sourceDir string) ([]domain.Species, error)
	ParseBackgrounds(ctx context.Context, sourceDir string) ([]domain.Background, error)
	ParseFeats(ctx context.Context, sourceDir string) ([]domain.Feat, error)
	ParseEquipment(ctx context.Context, sourceDir string) ([]domain.Equipment, error)
	ParseMagicItems(ctx context.Context, sourceDir string) ([]domain.MagicItem, error)
	ParseConditions(ctx context.Context, sourceDir string) ([]domain.Condition, error)
	ParseGlossaryTerms(ctx context.Context, sourceDir string) ([]domain.GlossaryTerm, error)
	ParseRuleSections(ctx context.Context, sourceDir string) ([]domain.RuleSection, error)
}

// Registry maps version strings to ContentParser implementations.
var Registry = map[string]ContentParser{}

// Register adds a ContentParser to the global registry.
func Register(parser ContentParser) {
	Registry[parser.Version()] = parser
}

// Get returns the ContentParser for the given version, or nil if not found.
func Get(version string) ContentParser {
	return Registry[version]
}
