package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/azemoning/omni-5e/internal/domain"
)

// --- Species ---

func (s *Store) GetSpecies(ctx context.Context, srdVersion, slug string) (*domain.Species, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT sp.id, sp.slug, sv.version, sp.name, sp.size, sp.speed, sp.traits, sp.description
		FROM species sp JOIN srd_versions sv ON sp.srd_version_id = sv.id
		WHERE sv.version = $1 AND sp.slug = $2`, srdVersion, slug)
	sp := &domain.Species{}
	var traitsJSON []byte
	err := row.Scan(&sp.ID, &sp.Slug, &sp.SRDVersion, &sp.Name, &sp.Size, &sp.Speed, &traitsJSON, &sp.Description)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(traitsJSON, &sp.Traits)
	return sp, nil
}

func (s *Store) ListSpecies(ctx context.Context, filter domain.SpeciesFilter) (*domain.Page, error) {
	where, args, idx := "WHERE sv.version = $1", []any{filter.SRDVersion}, 2
	if filter.Size != "" {
		where += fmt.Sprintf(" AND sp.size = $%d", idx)
		args, idx = append(args, filter.Size), idx+1
	}
	var total int
	s.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM species sp JOIN srd_versions sv ON sp.srd_version_id = sv.id %s", where), args...).Scan(&total)

	rows, err := s.pool.Query(ctx, fmt.Sprintf(`
		SELECT sp.id, sp.slug, sv.version, sp.name, sp.size, sp.speed, sp.traits, sp.description
		FROM species sp JOIN srd_versions sv ON sp.srd_version_id = sv.id
		%s ORDER BY sp.name LIMIT $%d OFFSET $%d`, where, idx, idx+1), append(args, filter.Limit, filter.Offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.Species{}
	for rows.Next() {
		var sp domain.Species
		var traitsJSON []byte
		rows.Scan(&sp.ID, &sp.Slug, &sp.SRDVersion, &sp.Name, &sp.Size, &sp.Speed, &traitsJSON, &sp.Description)
		json.Unmarshal(traitsJSON, &sp.Traits)
		items = append(items, sp)
	}
	return &domain.Page{Items: items, Total: total, Limit: filter.Limit, Offset: filter.Offset, SRDVersion: filter.SRDVersion}, rows.Err()
}

func (s *Store) UpsertSpecies(ctx context.Context, sp *domain.Species) error {
	traitsJSON, _ := json.Marshal(sp.Traits)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO species (id, srd_version_id, slug, name, size, speed, traits, description)
		VALUES ($1, (SELECT id FROM srd_versions WHERE version = $2), $3, $4, $5, $6, $7, $8)
		ON CONFLICT (srd_version_id, slug) DO UPDATE SET
			name = EXCLUDED.name, size = EXCLUDED.size, speed = EXCLUDED.speed,
			traits = EXCLUDED.traits, description = EXCLUDED.description, updated_at = NOW()`,
		sp.ID, sp.SRDVersion, sp.Slug, sp.Name, sp.Size, sp.Speed, traitsJSON, sp.Description)
	return err
}

// --- Backgrounds ---

func (s *Store) GetBackground(ctx context.Context, srdVersion, slug string) (*domain.Background, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT bg.id, bg.slug, sv.version, bg.name, bg.ability_score_options,
		       bg.skill_proficiencies, bg.granted_feat_slug, bg.equipment, bg.description
		FROM backgrounds bg JOIN srd_versions sv ON bg.srd_version_id = sv.id
		WHERE sv.version = $1 AND bg.slug = $2`, srdVersion, slug)
	bg := &domain.Background{}
	var asoJSON, spJSON, eqJSON []byte
	err := row.Scan(&bg.ID, &bg.Slug, &bg.SRDVersion, &bg.Name, &asoJSON, &spJSON, &bg.GrantedFeatSlug, &eqJSON, &bg.Description)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(asoJSON, &bg.AbilityScoreOptions)
	json.Unmarshal(spJSON, &bg.SkillProficiencies)
	json.Unmarshal(eqJSON, &bg.Equipment)
	return bg, nil
}

func (s *Store) ListBackgrounds(ctx context.Context, params domain.ListParams) (*domain.Page, error) {
	var total int
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM backgrounds bg JOIN srd_versions sv ON bg.srd_version_id = sv.id WHERE sv.version = $1`, params.SRDVersion).Scan(&total)

	rows, err := s.pool.Query(ctx, `
		SELECT bg.id, bg.slug, sv.version, bg.name, bg.ability_score_options,
		       bg.skill_proficiencies, bg.granted_feat_slug, bg.equipment, bg.description
		FROM backgrounds bg JOIN srd_versions sv ON bg.srd_version_id = sv.id
		WHERE sv.version = $1 ORDER BY bg.name LIMIT $2 OFFSET $3`,
		params.SRDVersion, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.Background{}
	for rows.Next() {
		var bg domain.Background
		var asoJSON, spJSON, eqJSON []byte
		rows.Scan(&bg.ID, &bg.Slug, &bg.SRDVersion, &bg.Name, &asoJSON, &spJSON, &bg.GrantedFeatSlug, &eqJSON, &bg.Description)
		json.Unmarshal(asoJSON, &bg.AbilityScoreOptions)
		json.Unmarshal(spJSON, &bg.SkillProficiencies)
		json.Unmarshal(eqJSON, &bg.Equipment)
		items = append(items, bg)
	}
	return &domain.Page{Items: items, Total: total, Limit: params.Limit, Offset: params.Offset, SRDVersion: params.SRDVersion}, rows.Err()
}

func (s *Store) UpsertBackground(ctx context.Context, bg *domain.Background) error {
	asoJSON, _ := json.Marshal(bg.AbilityScoreOptions)
	spJSON, _ := json.Marshal(bg.SkillProficiencies)
	eqJSON, _ := json.Marshal(bg.Equipment)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO backgrounds (id, srd_version_id, slug, name, ability_score_options,
			skill_proficiencies, granted_feat_slug, equipment, description)
		VALUES ($1, (SELECT id FROM srd_versions WHERE version = $2), $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (srd_version_id, slug) DO UPDATE SET
			name = EXCLUDED.name, ability_score_options = EXCLUDED.ability_score_options,
			skill_proficiencies = EXCLUDED.skill_proficiencies,
			granted_feat_slug = EXCLUDED.granted_feat_slug, equipment = EXCLUDED.equipment,
			description = EXCLUDED.description, updated_at = NOW()`,
		bg.ID, bg.SRDVersion, bg.Slug, bg.Name, asoJSON, spJSON, bg.GrantedFeatSlug, eqJSON, bg.Description)
	return err
}
