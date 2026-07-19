-- +goose Up
-- SQL in this section is executed when the migration is applied.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE srd_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version TEXT UNIQUE NOT NULL,
    release_date DATE,
    source_url TEXT NOT NULL,
    license TEXT NOT NULL DEFAULT 'CC-BY-4.0',
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE spells (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    srd_version_id UUID NOT NULL REFERENCES srd_versions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    level SMALLINT NOT NULL,
    school TEXT NOT NULL,
    casting_time TEXT,
    range TEXT,
    duration TEXT,
    concentration BOOLEAN NOT NULL DEFAULT false,
    ritual BOOLEAN NOT NULL DEFAULT false,
    components JSONB NOT NULL DEFAULT '{}',
    description TEXT NOT NULL,
    at_higher_levels TEXT,
    search TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', name || ' ' || description)) STORED,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (srd_version_id, slug)
);
CREATE INDEX spells_search_idx ON spells USING GIN (search);
CREATE INDEX spells_level_school_idx ON spells (srd_version_id, level, school);

CREATE TABLE spell_classes (
    spell_id UUID REFERENCES spells(id) ON DELETE CASCADE,
    class_slug TEXT NOT NULL,
    PRIMARY KEY (spell_id, class_slug)
);

CREATE TABLE monsters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    srd_version_id UUID NOT NULL REFERENCES srd_versions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    size TEXT,
    type TEXT,
    alignment TEXT,
    ac_value SMALLINT,
    ac_source TEXT,
    hp_avg INTEGER,
    hp_formula TEXT,
    cr NUMERIC(4,2),
    xp INTEGER,
    category TEXT NOT NULL DEFAULT 'monster',
    speed JSONB,
    ability_scores JSONB,
    saving_throws JSONB,
    skills JSONB,
    damage_resistances JSONB,
    damage_immunities JSONB,
    damage_vulnerabilities JSONB,
    condition_immunities JSONB,
    senses JSONB,
    languages JSONB,
    environment JSONB,
    traits JSONB,
    actions JSONB,
    bonus_actions JSONB,
    reactions JSONB,
    legendary_actions JSONB,
    search TSVECTOR GENERATED ALWAYS AS (to_tsvector('english', name)) STORED,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (srd_version_id, slug)
);
CREATE INDEX monsters_cr_idx ON monsters (srd_version_id, cr);
CREATE INDEX monsters_resist_gin ON monsters USING GIN (damage_resistances);
CREATE INDEX monsters_search_idx ON monsters USING GIN (search);

CREATE TABLE classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    srd_version_id UUID NOT NULL REFERENCES srd_versions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    hit_die SMALLINT NOT NULL,
    primary_ability TEXT,
    saving_throw_proficiencies JSONB,
    armor_proficiencies JSONB,
    weapon_proficiencies JSONB,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (srd_version_id, slug)
);

CREATE TABLE subclasses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    srd_version_id UUID NOT NULL REFERENCES srd_versions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    class_slug TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (srd_version_id, slug)
);
CREATE INDEX subclasses_class_idx ON subclasses (srd_version_id, class_slug);

CREATE TABLE class_features (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    srd_version_id UUID NOT NULL REFERENCES srd_versions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    class_slug TEXT NOT NULL,
    subclass_slug TEXT,
    level SMALLINT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (srd_version_id, slug)
);
CREATE INDEX class_features_class_level_idx ON class_features (srd_version_id, class_slug, level);

CREATE TABLE class_level_tables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    srd_version_id UUID NOT NULL REFERENCES srd_versions(id),
    class_slug TEXT NOT NULL,
    level SMALLINT NOT NULL,
    proficiency_bonus SMALLINT NOT NULL,
    features_unlocked JSONB,
    other_columns JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (srd_version_id, class_slug, level)
);

CREATE TABLE species (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    srd_version_id UUID NOT NULL REFERENCES srd_versions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    size TEXT,
    speed SMALLINT,
    traits JSONB,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (srd_version_id, slug)
);

CREATE TABLE backgrounds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    srd_version_id UUID NOT NULL REFERENCES srd_versions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    ability_score_options JSONB,
    skill_proficiencies JSONB,
    granted_feat_slug TEXT,
    equipment JSONB,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (srd_version_id, slug)
);

CREATE TABLE feats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    srd_version_id UUID NOT NULL REFERENCES srd_versions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    category TEXT,
    prerequisite TEXT,
    description TEXT NOT NULL,
    repeatable BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (srd_version_id, slug)
);

CREATE TABLE equipment (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    srd_version_id UUID NOT NULL REFERENCES srd_versions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    category TEXT NOT NULL,
    cost TEXT,
    weight NUMERIC(8,2),
    properties JSONB,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (srd_version_id, slug)
);
CREATE INDEX equipment_category_idx ON equipment (srd_version_id, category);

CREATE TABLE magic_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    srd_version_id UUID NOT NULL REFERENCES srd_versions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    rarity TEXT NOT NULL,
    requires_attunement BOOLEAN NOT NULL DEFAULT false,
    type TEXT,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (srd_version_id, slug)
);
CREATE INDEX magic_items_rarity_idx ON magic_items (srd_version_id, rarity);

CREATE TABLE conditions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    srd_version_id UUID NOT NULL REFERENCES srd_versions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (srd_version_id, slug)
);

CREATE TABLE glossary_terms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    srd_version_id UUID NOT NULL REFERENCES srd_versions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    category TEXT,
    definition TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (srd_version_id, slug)
);

CREATE TABLE rule_sections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    srd_version_id UUID NOT NULL REFERENCES srd_versions(id),
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    source_file TEXT,
    heading_path JSONB,
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (srd_version_id, slug)
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE IF EXISTS rule_sections;
DROP TABLE IF EXISTS glossary_terms;
DROP TABLE IF EXISTS conditions;
DROP TABLE IF EXISTS magic_items;
DROP TABLE IF EXISTS equipment;
DROP TABLE IF EXISTS feats;
DROP TABLE IF EXISTS backgrounds;
DROP TABLE IF EXISTS species;
DROP TABLE IF EXISTS class_level_tables;
DROP TABLE IF EXISTS class_features;
DROP TABLE IF EXISTS subclasses;
DROP TABLE IF EXISTS classes;
DROP TABLE IF EXISTS spell_classes;
DROP TABLE IF EXISTS monsters;
DROP TABLE IF EXISTS spells;
DROP TABLE IF EXISTS srd_versions;
