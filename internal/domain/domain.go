package domain

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// SRDVersion represents an ingested SRD content version.
type SRDVersion struct {
	ID          uuid.UUID `json:"id"`
	Version     string    `json:"version"`
	ReleaseDate *time.Time `json:"release_date,omitempty"`
	SourceURL   string    `json:"source_url"`
	License     string    `json:"license"`
	IsDefault   bool      `json:"is_default"`
}

// NamedBlock is a reusable {Name, Description} pair for monster traits/actions/etc.
type NamedBlock struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Trait is used by Species for racial traits.
type Trait struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// --- Species (Race) ---

type Species struct {
	BaseEntity
	Size        string  `json:"size"`
	Speed       int     `json:"speed"`
	Traits      []Trait `json:"traits"`
	Description string  `json:"description"`
}

// --- Background ---

type Background struct {
	BaseEntity
	AbilityScoreOptions  map[string]any `json:"ability_score_options,omitempty"`
	SkillProficiencies   []string       `json:"skill_proficiencies,omitempty"`
	GrantedFeatSlug      string         `json:"granted_feat_slug,omitempty"`
	Equipment            []string       `json:"equipment,omitempty"`
	Description          string         `json:"description"`
}

// --- Class ---

type Class struct {
	BaseEntity
	HitDie                   int        `json:"hit_die"`
	PrimaryAbility           string     `json:"primary_ability"`
	SavingThrowProficiencies []string   `json:"saving_throw_proficiencies"`
	ArmorProficiencies       []string   `json:"armor_proficiencies"`
	WeaponProficiencies      []string   `json:"weapon_proficiencies"`
	Description              string     `json:"description"`
	Subclasses               []Subclass `json:"subclasses,omitempty"`
}

// --- Subclass ---

type Subclass struct {
	BaseEntity
	ClassSlug   string `json:"class_slug"`
	Description string `json:"description"`
}

// --- ClassFeature ---

type ClassFeature struct {
	BaseEntity
	ClassSlug    string         `json:"class_slug"`
	SubclassSlug sql.NullString `json:"subclass_slug,omitempty"`
	Level        int            `json:"level"`
	Description  string         `json:"description"`
}

// --- ClassLevelTableRow ---

type ClassLevelTableRow struct {
	ID               uuid.UUID          `json:"id"`
	ClassSlug        string             `json:"class_slug"`
	SRDVersion       string             `json:"srd_version"`
	Level            int                `json:"level"`
	ProficiencyBonus int                `json:"proficiency_bonus"`
	FeaturesUnlocked []string           `json:"features_unlocked"`
	OtherColumns     map[string]string  `json:"other_columns"`
}

// --- Feat ---

type Feat struct {
	BaseEntity
	Category      string `json:"category,omitempty"`
	Prerequisite  string `json:"prerequisite,omitempty"`
	Description   string `json:"description"`
	Repeatable    bool   `json:"repeatable"`
}

// --- Spell ---

type Spell struct {
	BaseEntity
	Level         int      `json:"level"`
	School        string   `json:"school"`
	CastingTime   string   `json:"casting_time"`
	Range         string   `json:"range"`
	Components    Components `json:"components"`
	Duration      string   `json:"duration"`
	Concentration bool     `json:"concentration"`
	Ritual        bool     `json:"ritual"`
	Description   string   `json:"description"`
	AtHigherLevels string  `json:"at_higher_levels,omitempty"`
	ClassSlugs    []string `json:"class_slugs"`
}

// Components holds spell component info.
type Components struct {
	Verbal    bool   `json:"verbal"`
	Somatic   bool   `json:"somatic"`
	Material  bool   `json:"material"`
	MaterialDetail string `json:"material_detail,omitempty"`
}

// --- Equipment ---

type Equipment struct {
	BaseEntity
	Category    string         `json:"category"` // weapon|armor|tool|gear|mount|vehicle
	Cost        string         `json:"cost"`
	Weight      float64        `json:"weight,omitempty"`
	Properties  map[string]any `json:"properties,omitempty"`
	Description string         `json:"description"`
}

// --- MagicItem ---

type MagicItem struct {
	BaseEntity
	Rarity            string `json:"rarity"`
	RequiresAttunement bool  `json:"requires_attunement"`
	Type              string `json:"type"`
	Description       string `json:"description"`
}

// --- Monster ---

type Monster struct {
	BaseEntity
	Size                 string         `json:"size"`
	Type                 string         `json:"type"`
	Alignment            string         `json:"alignment"`
	AC                   ACInfo         `json:"ac"`
	HP                   HPInfo         `json:"hp"`
	Speed                map[string]int `json:"speed"`
	AbilityScores        AbilityScores  `json:"ability_scores"`
	SavingThrows         map[string]int `json:"saving_throws,omitempty"`
	Skills               map[string]int `json:"skills,omitempty"`
	DamageResistances    []string       `json:"damage_resistances,omitempty"`
	DamageImmunities     []string       `json:"damage_immunities,omitempty"`
	DamageVulnerabilities []string      `json:"damage_vulnerabilities,omitempty"`
	ConditionImmunities  []string       `json:"condition_immunities,omitempty"`
	Senses               map[string]any `json:"senses,omitempty"`
	Languages            []string       `json:"languages,omitempty"`
	CR                   float64        `json:"cr"`
	XP                   int            `json:"xp"`
	Traits               []NamedBlock   `json:"traits,omitempty"`
	Actions              []NamedBlock   `json:"actions,omitempty"`
	BonusActions         []NamedBlock   `json:"bonus_actions,omitempty"`
	Reactions            []NamedBlock   `json:"reactions,omitempty"`
	LegendaryActions     []NamedBlock   `json:"legendary_actions,omitempty"`
	Environment          []string       `json:"environment,omitempty"`
	Category             string         `json:"category"` // monster|animal
}

type ACInfo struct {
	Value  int    `json:"value"`
	Source string `json:"source,omitempty"`
}

type HPInfo struct {
	Average int    `json:"average"`
	Formula string `json:"formula,omitempty"`
}

type AbilityScores struct {
	STR int `json:"str"`
	DEX int `json:"dex"`
	CON int `json:"con"`
	INT int `json:"int"`
	WIS int `json:"wis"`
	CHA int `json:"cha"`
}

// --- Condition ---

type Condition struct {
	BaseEntity
	Description string `json:"description"`
}

// --- GlossaryTerm ---

type GlossaryTerm struct {
	BaseEntity
	Category   string `json:"category"`
	Definition string `json:"definition"`
}

// --- RuleSection ---

type RuleSection struct {
	BaseEntity
	SourceFile  string   `json:"source_file"`
	HeadingPath []string `json:"heading_path"`
	Body        string   `json:"body"`
}

// --- Pagination ---

type Page struct {
	Items      any   `json:"items"`
	Total      int   `json:"total"`
	Limit      int   `json:"limit"`
	Offset     int   `json:"offset"`
	SRDVersion string `json:"srd_version"`
}

// --- ListParams ---

type ListParams struct {
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
	SRDVersion string `json:"srd_version"`
	SortBy     string `json:"sort_by,omitempty"`
	SortOrder  string `json:"sort_order,omitempty"`
}

// --- FilterParams for specific resources ---

type SpellFilter struct {
	ListParams
	Class         string `json:"class,omitempty"`
	Level         *int   `json:"level,omitempty"`
	School        string `json:"school,omitempty"`
	Ritual        *bool  `json:"ritual,omitempty"`
	Concentration *bool  `json:"concentration,omitempty"`
	Query         string `json:"q,omitempty"`
}

type MonsterFilter struct {
	ListParams
	CRMin       *float64 `json:"cr_min,omitempty"`
	CRMax       *float64 `json:"cr_max,omitempty"`
	Type        string   `json:"type,omitempty"`
	Size        string   `json:"size,omitempty"`
	Environment string   `json:"environment,omitempty"`
	Category    string   `json:"category,omitempty"`
	Query       string   `json:"q,omitempty"`
}

type EquipmentFilter struct {
	ListParams
	Category string `json:"category,omitempty"`
}

type MagicItemFilter struct {
	ListParams
	Rarity      string `json:"rarity,omitempty"`
	Attunement  *bool  `json:"attunement,omitempty"`
}

type FeatFilter struct {
	ListParams
	Category string `json:"category,omitempty"`
}

type SpeciesFilter struct {
	ListParams
	Size string `json:"size,omitempty"`
}

type GlossaryFilter struct {
	ListParams
	Category string `json:"category,omitempty"`
}

type RuleSectionFilter struct {
	ListParams
	SourceFile string `json:"source_file,omitempty"`
}
