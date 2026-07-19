package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/azemoning/omni-5e/internal/domain"
)

// --- Classes ---

func (s *Store) GetClass(ctx context.Context, srdVersion, slug string) (*domain.Class, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT c.id, c.slug, sv.version, c.name, c.hit_die, c.primary_ability,
		       c.saving_throw_proficiencies, c.armor_proficiencies, c.weapon_proficiencies, c.description
		FROM classes c JOIN srd_versions sv ON c.srd_version_id = sv.id
		WHERE sv.version = $1 AND c.slug = $2`, srdVersion, slug)

	cl := &domain.Class{}
	var savesJSON, armorJSON, weaponsJSON []byte
	err := row.Scan(&cl.ID, &cl.Slug, &cl.SRDVersion, &cl.Name, &cl.HitDie, &cl.PrimaryAbility,
		&savesJSON, &armorJSON, &weaponsJSON, &cl.Description)
	if err != nil {
		return nil, fmt.Errorf("class %s/%s: %w", srdVersion, slug, err)
	}
	json.Unmarshal(savesJSON, &cl.SavingThrowProficiencies)
	json.Unmarshal(armorJSON, &cl.ArmorProficiencies)
	json.Unmarshal(weaponsJSON, &cl.WeaponProficiencies)

	// Load subclasses
	rows, err := s.pool.Query(ctx,
		`SELECT id, slug, sv.version, name, class_slug, description
		 FROM subclasses sc JOIN srd_versions sv ON sc.srd_version_id = sv.id
		 WHERE sv.version = $1 AND sc.class_slug = $2`, srdVersion, slug)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sc domain.Subclass
			rows.Scan(&sc.ID, &sc.Slug, &sc.SRDVersion, &sc.Name, &sc.ClassSlug, &sc.Description)
			cl.Subclasses = append(cl.Subclasses, sc)
		}
	}
	return cl, nil
}

func (s *Store) ListClasses(ctx context.Context, params domain.ListParams) (*domain.Page, error) {
	var total int
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM classes c JOIN srd_versions sv ON c.srd_version_id = sv.id WHERE sv.version = $1`, params.SRDVersion).Scan(&total)

	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.slug, sv.version, c.name, c.hit_die, c.primary_ability,
		       c.saving_throw_proficiencies, c.armor_proficiencies, c.weapon_proficiencies, c.description
		FROM classes c JOIN srd_versions sv ON c.srd_version_id = sv.id
		WHERE sv.version = $1 ORDER BY c.name LIMIT $2 OFFSET $3`,
		params.SRDVersion, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	classes := []domain.Class{}
	for rows.Next() {
		var cl domain.Class
		var savesJSON, armorJSON, weaponsJSON []byte
		rows.Scan(&cl.ID, &cl.Slug, &cl.SRDVersion, &cl.Name, &cl.HitDie, &cl.PrimaryAbility,
			&savesJSON, &armorJSON, &weaponsJSON, &cl.Description)
		json.Unmarshal(savesJSON, &cl.SavingThrowProficiencies)
		json.Unmarshal(armorJSON, &cl.ArmorProficiencies)
		json.Unmarshal(weaponsJSON, &cl.WeaponProficiencies)
		classes = append(classes, cl)
	}
	return &domain.Page{Items: classes, Total: total, Limit: params.Limit, Offset: params.Offset, SRDVersion: params.SRDVersion}, rows.Err()
}

func (s *Store) GetSubclasses(ctx context.Context, srdVersion, classSlug string) ([]domain.Subclass, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, slug, sv.version, name, class_slug, description
		FROM subclasses sc JOIN srd_versions sv ON sc.srd_version_id = sv.id
		WHERE sv.version = $1 AND sc.class_slug = $2`, srdVersion, classSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []domain.Subclass
	for rows.Next() {
		var sc domain.Subclass
		rows.Scan(&sc.ID, &sc.Slug, &sc.SRDVersion, &sc.Name, &sc.ClassSlug, &sc.Description)
		subs = append(subs, sc)
	}
	return subs, rows.Err()
}

func (s *Store) GetClassLevelTable(ctx context.Context, srdVersion, classSlug string, level int) (*domain.ClassLevelTableRow, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, class_slug, level, proficiency_bonus, features_unlocked, other_columns
		FROM class_level_tables clt JOIN srd_versions sv ON clt.srd_version_id = sv.id
		WHERE sv.version = $1 AND clt.class_slug = $2 AND clt.level = $3`, srdVersion, classSlug, level)

	r := &domain.ClassLevelTableRow{}
	var fJSON, oJSON []byte
	err := row.Scan(&r.ID, &r.ClassSlug, &r.Level, &r.ProficiencyBonus, &fJSON, &oJSON)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(fJSON, &r.FeaturesUnlocked)
	json.Unmarshal(oJSON, &r.OtherColumns)
	return r, nil
}

func (s *Store) UpsertClass(ctx context.Context, cl *domain.Class) error {
	savesJSON, _ := json.Marshal(cl.SavingThrowProficiencies)
	armorJSON, _ := json.Marshal(cl.ArmorProficiencies)
	weaponsJSON, _ := json.Marshal(cl.WeaponProficiencies)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO classes (id, srd_version_id, slug, name, hit_die, primary_ability,
			saving_throw_proficiencies, armor_proficiencies, weapon_proficiencies, description)
		VALUES ($1, (SELECT id FROM srd_versions WHERE version = $2), $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (srd_version_id, slug) DO UPDATE SET
			name = EXCLUDED.name, hit_die = EXCLUDED.hit_die, primary_ability = EXCLUDED.primary_ability,
			saving_throw_proficiencies = EXCLUDED.saving_throw_proficiencies,
			armor_proficiencies = EXCLUDED.armor_proficiencies,
			weapon_proficiencies = EXCLUDED.weapon_proficiencies,
			description = EXCLUDED.description, updated_at = NOW()`,
		cl.ID, cl.SRDVersion, cl.Slug, cl.Name, cl.HitDie, cl.PrimaryAbility,
		savesJSON, armorJSON, weaponsJSON, cl.Description)
	return err
}
