package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/azemoning/omni-5e/internal/domain"
)

// --- Spells ---

func (s *Store) GetSpell(ctx context.Context, srdVersion, slug string) (*domain.Spell, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT sp.id, sp.slug, sv.version, sp.name, sp.level, sp.school, sp.casting_time,
		       sp.range, sp.duration, sp.concentration, sp.ritual, sp.components,
		       sp.description, sp.at_higher_levels
		FROM spells sp JOIN srd_versions sv ON sp.srd_version_id = sv.id
		WHERE sv.version = $1 AND sp.slug = $2`, srdVersion, slug)

	sp := &domain.Spell{}
	var compJSON []byte
	err := row.Scan(&sp.ID, &sp.Slug, &sp.SRDVersion, &sp.Name, &sp.Level, &sp.School,
		&sp.CastingTime, &sp.Range, &sp.Duration, &sp.Concentration, &sp.Ritual,
		&compJSON, &sp.Description, &sp.AtHigherLevels)
	if err != nil {
		return nil, fmt.Errorf("spell %s/%s: %w", srdVersion, slug, err)
	}
	json.Unmarshal(compJSON, &sp.Components)

	// Load class slugs
	rows, err := s.pool.Query(ctx, `SELECT class_slug FROM spell_classes WHERE spell_id = $1`, sp.ID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var cs string
			rows.Scan(&cs)
			sp.ClassSlugs = append(sp.ClassSlugs, cs)
		}
	}
	return sp, nil
}

func (s *Store) ListSpells(ctx context.Context, filter domain.SpellFilter) (*domain.Page, error) {
	return s.listSpells(ctx, filter)
}

func (s *Store) listSpells(ctx context.Context, filter domain.SpellFilter) (*domain.Page, error) {
	where, args, idx := "WHERE sv.version = $1", []any{filter.SRDVersion}, 2
	if filter.Class != "" {
		where += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM spell_classes sc WHERE sc.spell_id = sp.id AND sc.class_slug = $%d)", idx)
		args, idx = append(args, filter.Class), idx+1
	}
	if filter.Level != nil {
		where += fmt.Sprintf(" AND sp.level = $%d", idx)
		args, idx = append(args, *filter.Level), idx+1
	}
	if filter.School != "" {
		where += fmt.Sprintf(" AND sp.school = $%d", idx)
		args, idx = append(args, filter.School), idx+1
	}
	if filter.Ritual != nil {
		where += fmt.Sprintf(" AND sp.ritual = $%d", idx)
		args, idx = append(args, *filter.Ritual), idx+1
	}
	if filter.Concentration != nil {
		where += fmt.Sprintf(" AND sp.concentration = $%d", idx)
		args, idx = append(args, *filter.Concentration), idx+1
	}
	if filter.Query != "" {
		where += fmt.Sprintf(" AND sp.search @@ plainto_tsquery('english', $%d)", idx)
		args, idx = append(args, filter.Query), idx+1
	}

	var total int
	s.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM spells sp JOIN srd_versions sv ON sp.srd_version_id = sv.id %s", where), args...).Scan(&total)

	dataQ := fmt.Sprintf(`
		SELECT sp.id, sp.slug, sv.version, sp.name, sp.level, sp.school, sp.casting_time,
		       sp.range, sp.duration, sp.concentration, sp.ritual, sp.components,
		       sp.description, sp.at_higher_levels
		FROM spells sp JOIN srd_versions sv ON sp.srd_version_id = sv.id
		%s ORDER BY sp.level, sp.name LIMIT $%d OFFSET $%d`, where, idx, idx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := s.pool.Query(ctx, dataQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	spells := []domain.Spell{}
	for rows.Next() {
		var sp domain.Spell
		var compJSON []byte
		rows.Scan(&sp.ID, &sp.Slug, &sp.SRDVersion, &sp.Name, &sp.Level, &sp.School,
			&sp.CastingTime, &sp.Range, &sp.Duration, &sp.Concentration, &sp.Ritual,
			&compJSON, &sp.Description, &sp.AtHigherLevels)
		json.Unmarshal(compJSON, &sp.Components)
		spells = append(spells, sp)
	}
	return &domain.Page{Items: spells, Total: total, Limit: filter.Limit, Offset: filter.Offset, SRDVersion: filter.SRDVersion}, rows.Err()
}

func (s *Store) UpsertSpell(ctx context.Context, sp *domain.Spell) error {
	compJSON, _ := json.Marshal(sp.Components)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO spells (id, srd_version_id, slug, name, level, school, casting_time,
			range, duration, concentration, ritual, components, description, at_higher_levels)
		VALUES ($1, (SELECT id FROM srd_versions WHERE version = $2), $3, $4, $5, $6, $7,
		        $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (srd_version_id, slug) DO UPDATE SET
			name = EXCLUDED.name, level = EXCLUDED.level, school = EXCLUDED.school,
			casting_time = EXCLUDED.casting_time, range = EXCLUDED.range,
			duration = EXCLUDED.duration, concentration = EXCLUDED.concentration,
			ritual = EXCLUDED.ritual, components = EXCLUDED.components,
			description = EXCLUDED.description, at_higher_levels = EXCLUDED.at_higher_levels,
			updated_at = NOW()`,
		sp.ID, sp.SRDVersion, sp.Slug, sp.Name, sp.Level, sp.School,
		sp.CastingTime, sp.Range, sp.Duration, sp.Concentration, sp.Ritual,
		compJSON, sp.Description, sp.AtHigherLevels)
	return err
}

func (s *Store) UpsertSpells(ctx context.Context, spells []domain.Spell) error {
	for _, sp := range spells {
		if err := s.UpsertSpell(ctx, &sp); err != nil {
			return err
		}
		// Upsert class associations
		s.pool.Exec(ctx, `DELETE FROM spell_classes WHERE spell_id = $1`, sp.ID)
		for _, cs := range sp.ClassSlugs {
			s.pool.Exec(ctx, `INSERT INTO spell_classes (spell_id, class_slug) VALUES ($1, $2) ON CONFLICT DO NOTHING`, sp.ID, cs)
		}
	}
	return nil
}
