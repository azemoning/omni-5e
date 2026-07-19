package dto

import "time"

// ResponseEnvelope is the standard response wrapper for list endpoints.
type ResponseEnvelope struct {
	Data  any            `json:"data"`
	Meta  ResponseMeta   `json:"meta"`
	Links ResponseLinks  `json:"links"`
}

// ResponseMeta contains pagination and version metadata.
type ResponseMeta struct {
	SRDVersion string `json:"srd_version"`
	Total      int    `json:"total"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

// ResponseLinks contains pagination links.
type ResponseLinks struct {
	Self string  `json:"self"`
	Next *string `json:"next"`
	Prev *string `json:"prev"`
}

// SingleResponse wraps a single resource.
type SingleResponse struct {
	Data any `json:"data"`
}

// ErrorResponse is an RFC 7807 problem detail.
type ErrorResponse struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// --- SRD Version DTOs ---

type SRDVersionResponse struct {
	Version     string     `json:"version"`
	ReleaseDate *time.Time `json:"release_date,omitempty"`
	SourceURL   string     `json:"source_url"`
	License     string     `json:"license"`
	IsDefault   bool       `json:"is_default"`
}

// --- Spell DTOs ---

type SpellResponse struct {
	Slug          string     `json:"slug"`
	URL           string     `json:"url"`
	SRDVersion    string     `json:"srd_version"`
	Name          string     `json:"name"`
	Level         int        `json:"level"`
	School        string     `json:"school"`
	CastingTime   string     `json:"casting_time"`
	Range         string     `json:"range"`
	Components    Components `json:"components"`
	Duration      string     `json:"duration"`
	Concentration bool       `json:"concentration"`
	Ritual        bool       `json:"ritual"`
	Description   string     `json:"description"`
	AtHigherLevels string    `json:"at_higher_levels,omitempty"`
	ClassSlugs    []string   `json:"class_slugs"`
}

type Components struct {
	Verbal       bool   `json:"verbal"`
	Somatic      bool   `json:"somatic"`
	Material     bool   `json:"material"`
	MaterialDetail string `json:"material_detail,omitempty"`
}

// --- Monster DTOs ---

type MonsterResponse struct {
	Slug                  string         `json:"slug"`
	URL                   string         `json:"url"`
	SRDVersion            string         `json:"srd_version"`
	Name                  string         `json:"name"`
	Size                  string         `json:"size"`
	Type                  string         `json:"type"`
	Alignment             string         `json:"alignment"`
	AC                    ACInfo         `json:"ac"`
	HP                    HPInfo         `json:"hp"`
	Speed                 map[string]int `json:"speed"`
	AbilityScores         AbilityScores  `json:"ability_scores"`
	SavingThrows          map[string]int `json:"saving_throws,omitempty"`
	Skills                map[string]int `json:"skills,omitempty"`
	DamageResistances     []string       `json:"damage_resistances,omitempty"`
	DamageImmunities      []string       `json:"damage_immunities,omitempty"`
	DamageVulnerabilities []string       `json:"damage_vulnerabilities,omitempty"`
	ConditionImmunities   []string       `json:"condition_immunities,omitempty"`
	Senses                map[string]any `json:"senses,omitempty"`
	Languages             []string       `json:"languages,omitempty"`
	CR                    float64        `json:"cr"`
	XP                    int            `json:"xp"`
	Traits                []NamedBlock   `json:"traits,omitempty"`
	Actions               []NamedBlock   `json:"actions,omitempty"`
	BonusActions          []NamedBlock   `json:"bonus_actions,omitempty"`
	Reactions             []NamedBlock   `json:"reactions,omitempty"`
	LegendaryActions      []NamedBlock   `json:"legendary_actions,omitempty"`
	Environment           []string       `json:"environment,omitempty"`
	Category              string         `json:"category"`
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

type NamedBlock struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// --- Class DTOs ---

type ClassResponse struct {
	Slug                     string       `json:"slug"`
	URL                      string       `json:"url"`
	SRDVersion               string       `json:"srd_version"`
	Name                     string       `json:"name"`
	HitDie                   int          `json:"hit_die"`
	PrimaryAbility           string       `json:"primary_ability"`
	SavingThrowProficiencies []string     `json:"saving_throw_proficiencies"`
	ArmorProficiencies       []string     `json:"armor_proficiencies"`
	WeaponProficiencies      []string     `json:"weapon_proficiencies"`
	Description              string       `json:"description"`
	Subclasses               []SubclassResponse `json:"subclasses,omitempty"`
}

type SubclassResponse struct {
	Slug        string `json:"slug"`
	URL         string `json:"url"`
	SRDVersion  string `json:"srd_version"`
	Name        string `json:"name"`
	ClassSlug   string `json:"class_slug"`
	Description string `json:"description"`
}

type ClassLevelTableRowResponse struct {
	ClassSlug        string            `json:"class_slug"`
	Level            int               `json:"level"`
	ProficiencyBonus int               `json:"proficiency_bonus"`
	FeaturesUnlocked []string          `json:"features_unlocked"`
	OtherColumns     map[string]string `json:"other_columns"`
}

// --- Species DTOs ---

type SpeciesResponse struct {
	Slug        string  `json:"slug"`
	URL         string  `json:"url"`
	SRDVersion  string  `json:"srd_version"`
	Name        string  `json:"name"`
	Size        string  `json:"size"`
	Speed       int     `json:"speed"`
	Traits      []Trait `json:"traits"`
	Description string  `json:"description"`
}

type Trait struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// --- Background DTOs ---

type BackgroundResponse struct {
	Slug                string         `json:"slug"`
	URL                 string         `json:"url"`
	SRDVersion          string         `json:"srd_version"`
	Name                string         `json:"name"`
	AbilityScoreOptions map[string]any `json:"ability_score_options,omitempty"`
	SkillProficiencies  []string       `json:"skill_proficiencies,omitempty"`
	GrantedFeatSlug     string         `json:"granted_feat_slug,omitempty"`
	Equipment           []string       `json:"equipment,omitempty"`
	Description         string         `json:"description"`
}

// --- Feat DTOs ---

type FeatResponse struct {
	Slug         string `json:"slug"`
	URL          string `json:"url"`
	SRDVersion   string `json:"srd_version"`
	Name         string `json:"name"`
	Category     string `json:"category,omitempty"`
	Prerequisite string `json:"prerequisite,omitempty"`
	Description  string `json:"description"`
	Repeatable   bool   `json:"repeatable"`
}

// --- Equipment DTOs ---

type EquipmentResponse struct {
	Slug        string         `json:"slug"`
	URL         string         `json:"url"`
	SRDVersion  string         `json:"srd_version"`
	Name        string         `json:"name"`
	Category    string         `json:"category"`
	Cost        string         `json:"cost"`
	Weight      float64        `json:"weight,omitempty"`
	Properties  map[string]any `json:"properties,omitempty"`
	Description string         `json:"description"`
}

// --- Magic Item DTOs ---

type MagicItemResponse struct {
	Slug                string `json:"slug"`
	URL                 string `json:"url"`
	SRDVersion          string `json:"srd_version"`
	Name                string `json:"name"`
	Rarity              string `json:"rarity"`
	RequiresAttunement  bool   `json:"requires_attunement"`
	Type                string `json:"type"`
	Description         string `json:"description"`
}

// --- Condition DTOs ---

type ConditionResponse struct {
	Slug        string `json:"slug"`
	URL         string `json:"url"`
	SRDVersion  string `json:"srd_version"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// --- Glossary Term DTOs ---

type GlossaryTermResponse struct {
	Slug        string `json:"slug"`
	URL         string `json:"url"`
	SRDVersion  string `json:"srd_version"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Definition  string `json:"definition"`
}

// --- Rule Section DTOs ---

type RuleSectionResponse struct {
	Slug        string   `json:"slug"`
	URL         string   `json:"url"`
	SRDVersion  string   `json:"srd_version"`
	Name        string   `json:"name"`
	SourceFile  string   `json:"source_file"`
	HeadingPath []string `json:"heading_path"`
	Body        string   `json:"body"`
}

// --- License DTOs ---

type LicenseResponse struct {
	License   string `json:"license"`
	SourceURL string `json:"source_url"`
	Version   string `json:"version"`
	Text      string `json:"text"`
}
