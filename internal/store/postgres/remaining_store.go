package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/azemoning/omni-5e/internal/domain"
)

// --- Feats ---

func (s *Store) GetFeat(ctx context.Context, srdVersion, slug string) (*domain.Feat, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT f.id, f.slug, sv.version, f.name, f.category, f.prerequisite, f.description, f.repeatable
		FROM feats f JOIN srd_versions sv ON f.srd_version_id = sv.id
		WHERE sv.version = $1 AND f.slug = $2`, srdVersion, slug)
	f := &domain.Feat{}
	err := row.Scan(&f.ID, &f.Slug, &f.SRDVersion, &f.Name, &f.Category, &f.Prerequisite, &f.Description, &f.Repeatable)
	return f, err
}

func (s *Store) ListFeats(ctx context.Context, filter domain.FeatFilter) (*domain.Page, error) {
	where, args, idx := "WHERE sv.version = $1", []any{filter.SRDVersion}, 2
	if filter.Category != "" {
		where += fmt.Sprintf(" AND f.category = $%d", idx)
		args, idx = append(args, filter.Category), idx+1
	}
	var total int
	s.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM feats f JOIN srd_versions sv ON f.srd_version_id = sv.id %s", where), args...).Scan(&total)

	rows, err := s.pool.Query(ctx, fmt.Sprintf(`
		SELECT f.id, f.slug, sv.version, f.name, f.category, f.prerequisite, f.description, f.repeatable
		FROM feats f JOIN srd_versions sv ON f.srd_version_id = sv.id
		%s ORDER BY f.name LIMIT $%d OFFSET $%d`, where, idx, idx+1), append(args, filter.Limit, filter.Offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.Feat{}
	for rows.Next() {
		var f domain.Feat
		rows.Scan(&f.ID, &f.Slug, &f.SRDVersion, &f.Name, &f.Category, &f.Prerequisite, &f.Description, &f.Repeatable)
		items = append(items, f)
	}
	return &domain.Page{Items: items, Total: total, Limit: filter.Limit, Offset: filter.Offset, SRDVersion: filter.SRDVersion}, rows.Err()
}

func (s *Store) UpsertFeat(ctx context.Context, f *domain.Feat) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO feats (id, srd_version_id, slug, name, category, prerequisite, description, repeatable)
		VALUES ($1, (SELECT id FROM srd_versions WHERE version = $2), $3, $4, $5, $6, $7, $8)
		ON CONFLICT (srd_version_id, slug) DO UPDATE SET
			name = EXCLUDED.name, category = EXCLUDED.category, prerequisite = EXCLUDED.prerequisite,
			description = EXCLUDED.description, repeatable = EXCLUDED.repeatable, updated_at = NOW()`,
		f.ID, f.SRDVersion, f.Slug, f.Name, f.Category, f.Prerequisite, f.Description, f.Repeatable)
	return err
}

// --- Equipment ---

func (s *Store) GetEquipment(ctx context.Context, srdVersion, slug string) (*domain.Equipment, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT e.id, e.slug, sv.version, e.name, e.category, e.cost, e.weight, e.properties, e.description
		FROM equipment e JOIN srd_versions sv ON e.srd_version_id = sv.id
		WHERE sv.version = $1 AND e.slug = $2`, srdVersion, slug)
	e := &domain.Equipment{}
	var propsJSON []byte
	err := row.Scan(&e.ID, &e.Slug, &e.SRDVersion, &e.Name, &e.Category, &e.Cost, &e.Weight, &propsJSON, &e.Description)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(propsJSON, &e.Properties)
	return e, nil
}

func (s *Store) ListEquipment(ctx context.Context, filter domain.EquipmentFilter) (*domain.Page, error) {
	where, args, idx := "WHERE sv.version = $1", []any{filter.SRDVersion}, 2
	if filter.Category != "" {
		where += fmt.Sprintf(" AND e.category = $%d", idx)
		args, idx = append(args, filter.Category), idx+1
	}
	var total int
	s.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM equipment e JOIN srd_versions sv ON e.srd_version_id = sv.id %s", where), args...).Scan(&total)

	rows, err := s.pool.Query(ctx, fmt.Sprintf(`
		SELECT e.id, e.slug, sv.version, e.name, e.category, e.cost, e.weight, e.properties, e.description
		FROM equipment e JOIN srd_versions sv ON e.srd_version_id = sv.id
		%s ORDER BY e.name LIMIT $%d OFFSET $%d`, where, idx, idx+1), append(args, filter.Limit, filter.Offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.Equipment{}
	for rows.Next() {
		var e domain.Equipment
		var propsJSON []byte
		rows.Scan(&e.ID, &e.Slug, &e.SRDVersion, &e.Name, &e.Category, &e.Cost, &e.Weight, &propsJSON, &e.Description)
		json.Unmarshal(propsJSON, &e.Properties)
		items = append(items, e)
	}
	return &domain.Page{Items: items, Total: total, Limit: filter.Limit, Offset: filter.Offset, SRDVersion: filter.SRDVersion}, rows.Err()
}

func (s *Store) UpsertEquipment(ctx context.Context, e *domain.Equipment) error {
	propsJSON, _ := json.Marshal(e.Properties)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO equipment (id, srd_version_id, slug, name, category, cost, weight, properties, description)
		VALUES ($1, (SELECT id FROM srd_versions WHERE version = $2), $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (srd_version_id, slug) DO UPDATE SET
			name = EXCLUDED.name, category = EXCLUDED.category, cost = EXCLUDED.cost,
			weight = EXCLUDED.weight, properties = EXCLUDED.properties,
			description = EXCLUDED.description, updated_at = NOW()`,
		e.ID, e.SRDVersion, e.Slug, e.Name, e.Category, e.Cost, e.Weight, propsJSON, e.Description)
	return err
}

// --- Magic Items ---

func (s *Store) GetMagicItem(ctx context.Context, srdVersion, slug string) (*domain.MagicItem, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT mi.id, mi.slug, sv.version, mi.name, mi.rarity, mi.requires_attunement, mi.type, mi.description
		FROM magic_items mi JOIN srd_versions sv ON mi.srd_version_id = sv.id
		WHERE sv.version = $1 AND mi.slug = $2`, srdVersion, slug)
	mi := &domain.MagicItem{}
	err := row.Scan(&mi.ID, &mi.Slug, &mi.SRDVersion, &mi.Name, &mi.Rarity, &mi.RequiresAttunement, &mi.Type, &mi.Description)
	return mi, err
}

func (s *Store) ListMagicItems(ctx context.Context, filter domain.MagicItemFilter) (*domain.Page, error) {
	where, args, idx := "WHERE sv.version = $1", []any{filter.SRDVersion}, 2
	if filter.Rarity != "" {
		where += fmt.Sprintf(" AND mi.rarity = $%d", idx)
		args, idx = append(args, filter.Rarity), idx+1
	}
	if filter.Attunement != nil {
		where += fmt.Sprintf(" AND mi.requires_attunement = $%d", idx)
		args, idx = append(args, *filter.Attunement), idx+1
	}
	var total int
	s.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM magic_items mi JOIN srd_versions sv ON mi.srd_version_id = sv.id %s", where), args...).Scan(&total)

	rows, err := s.pool.Query(ctx, fmt.Sprintf(`
		SELECT mi.id, mi.slug, sv.version, mi.name, mi.rarity, mi.requires_attunement, mi.type, mi.description
		FROM magic_items mi JOIN srd_versions sv ON mi.srd_version_id = sv.id
		%s ORDER BY mi.name LIMIT $%d OFFSET $%d`, where, idx, idx+1), append(args, filter.Limit, filter.Offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.MagicItem{}
	for rows.Next() {
		var mi domain.MagicItem
		rows.Scan(&mi.ID, &mi.Slug, &mi.SRDVersion, &mi.Name, &mi.Rarity, &mi.RequiresAttunement, &mi.Type, &mi.Description)
		items = append(items, mi)
	}
	return &domain.Page{Items: items, Total: total, Limit: filter.Limit, Offset: filter.Offset, SRDVersion: filter.SRDVersion}, rows.Err()
}

func (s *Store) UpsertMagicItem(ctx context.Context, mi *domain.MagicItem) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO magic_items (id, srd_version_id, slug, name, rarity, requires_attunement, type, description)
		VALUES ($1, (SELECT id FROM srd_versions WHERE version = $2), $3, $4, $5, $6, $7, $8)
		ON CONFLICT (srd_version_id, slug) DO UPDATE SET
			name = EXCLUDED.name, rarity = EXCLUDED.rarity,
			requires_attunement = EXCLUDED.requires_attunement, type = EXCLUDED.type,
			description = EXCLUDED.description, updated_at = NOW()`,
		mi.ID, mi.SRDVersion, mi.Slug, mi.Name, mi.Rarity, mi.RequiresAttunement, mi.Type, mi.Description)
	return err
}

// --- Conditions ---

func (s *Store) GetCondition(ctx context.Context, srdVersion, slug string) (*domain.Condition, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT c.id, c.slug, sv.version, c.name, c.description
		FROM conditions c JOIN srd_versions sv ON c.srd_version_id = sv.id
		WHERE sv.version = $1 AND c.slug = $2`, srdVersion, slug)
	c := &domain.Condition{}
	err := row.Scan(&c.ID, &c.Slug, &c.SRDVersion, &c.Name, &c.Description)
	return c, err
}

func (s *Store) ListConditions(ctx context.Context, params domain.ListParams) (*domain.Page, error) {
	var total int
	s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM conditions c JOIN srd_versions sv ON c.srd_version_id = sv.id WHERE sv.version = $1`, params.SRDVersion).Scan(&total)

	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.slug, sv.version, c.name, c.description
		FROM conditions c JOIN srd_versions sv ON c.srd_version_id = sv.id
		WHERE sv.version = $1 ORDER BY c.name LIMIT $2 OFFSET $3`,
		params.SRDVersion, params.Limit, params.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.Condition{}
	for rows.Next() {
		var c domain.Condition
		rows.Scan(&c.ID, &c.Slug, &c.SRDVersion, &c.Name, &c.Description)
		items = append(items, c)
	}
	return &domain.Page{Items: items, Total: total, Limit: params.Limit, Offset: params.Offset, SRDVersion: params.SRDVersion}, rows.Err()
}

func (s *Store) UpsertCondition(ctx context.Context, c *domain.Condition) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO conditions (id, srd_version_id, slug, name, description)
		VALUES ($1, (SELECT id FROM srd_versions WHERE version = $2), $3, $4, $5)
		ON CONFLICT (srd_version_id, slug) DO UPDATE SET
			name = EXCLUDED.name, description = EXCLUDED.description, updated_at = NOW()`,
		c.ID, c.SRDVersion, c.Slug, c.Name, c.Description)
	return err
}

// --- Glossary Terms ---

func (s *Store) GetGlossaryTerm(ctx context.Context, srdVersion, slug string) (*domain.GlossaryTerm, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT g.id, g.slug, sv.version, g.name, g.category, g.definition
		FROM glossary_terms g JOIN srd_versions sv ON g.srd_version_id = sv.id
		WHERE sv.version = $1 AND g.slug = $2`, srdVersion, slug)
	g := &domain.GlossaryTerm{}
	err := row.Scan(&g.ID, &g.Slug, &g.SRDVersion, &g.Name, &g.Category, &g.Definition)
	return g, err
}

func (s *Store) ListGlossaryTerms(ctx context.Context, filter domain.GlossaryFilter) (*domain.Page, error) {
	where, args, idx := "WHERE sv.version = $1", []any{filter.SRDVersion}, 2
	if filter.Category != "" {
		where += fmt.Sprintf(" AND g.category = $%d", idx)
		args, idx = append(args, filter.Category), idx+1
	}
	var total int
	s.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM glossary_terms g JOIN srd_versions sv ON g.srd_version_id = sv.id %s", where), args...).Scan(&total)

	rows, err := s.pool.Query(ctx, fmt.Sprintf(`
		SELECT g.id, g.slug, sv.version, g.name, g.category, g.definition
		FROM glossary_terms g JOIN srd_versions sv ON g.srd_version_id = sv.id
		%s ORDER BY g.name LIMIT $%d OFFSET $%d`, where, idx, idx+1), append(args, filter.Limit, filter.Offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.GlossaryTerm{}
	for rows.Next() {
		var g domain.GlossaryTerm
		rows.Scan(&g.ID, &g.Slug, &g.SRDVersion, &g.Name, &g.Category, &g.Definition)
		items = append(items, g)
	}
	return &domain.Page{Items: items, Total: total, Limit: filter.Limit, Offset: filter.Offset, SRDVersion: filter.SRDVersion}, rows.Err()
}

func (s *Store) UpsertGlossaryTerm(ctx context.Context, g *domain.GlossaryTerm) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO glossary_terms (id, srd_version_id, slug, name, category, definition)
		VALUES ($1, (SELECT id FROM srd_versions WHERE version = $2), $3, $4, $5, $6)
		ON CONFLICT (srd_version_id, slug) DO UPDATE SET
			name = EXCLUDED.name, category = EXCLUDED.category,
			definition = EXCLUDED.definition, updated_at = NOW()`,
		g.ID, g.SRDVersion, g.Slug, g.Name, g.Category, g.Definition)
	return err
}

// --- Rule Sections ---

func (s *Store) GetRuleSection(ctx context.Context, srdVersion, slug string) (*domain.RuleSection, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT r.id, r.slug, sv.version, r.name, r.source_file, r.heading_path, r.body
		FROM rule_sections r JOIN srd_versions sv ON r.srd_version_id = sv.id
		WHERE sv.version = $1 AND r.slug = $2`, srdVersion, slug)
	r := &domain.RuleSection{}
	var hpJSON []byte
	err := row.Scan(&r.ID, &r.Slug, &r.SRDVersion, &r.Name, &r.SourceFile, &hpJSON, &r.Body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(hpJSON, &r.HeadingPath)
	return r, nil
}

func (s *Store) ListRuleSections(ctx context.Context, filter domain.RuleSectionFilter) (*domain.Page, error) {
	where, args, idx := "WHERE sv.version = $1", []any{filter.SRDVersion}, 2
	if filter.SourceFile != "" {
		where += fmt.Sprintf(" AND r.source_file = $%d", idx)
		args, idx = append(args, filter.SourceFile), idx+1
	}
	var total int
	s.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM rule_sections r JOIN srd_versions sv ON r.srd_version_id = sv.id %s", where), args...).Scan(&total)

	rows, err := s.pool.Query(ctx, fmt.Sprintf(`
		SELECT r.id, r.slug, sv.version, r.name, r.source_file, r.heading_path, r.body
		FROM rule_sections r JOIN srd_versions sv ON r.srd_version_id = sv.id
		%s ORDER BY r.name LIMIT $%d OFFSET $%d`, where, idx, idx+1), append(args, filter.Limit, filter.Offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.RuleSection{}
	for rows.Next() {
		var r domain.RuleSection
		var hpJSON []byte
		rows.Scan(&r.ID, &r.Slug, &r.SRDVersion, &r.Name, &r.SourceFile, &hpJSON, &r.Body)
		json.Unmarshal(hpJSON, &r.HeadingPath)
		items = append(items, r)
	}
	return &domain.Page{Items: items, Total: total, Limit: filter.Limit, Offset: filter.Offset, SRDVersion: filter.SRDVersion}, rows.Err()
}

func (s *Store) UpsertRuleSection(ctx context.Context, r *domain.RuleSection) error {
	hpJSON, _ := json.Marshal(r.HeadingPath)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO rule_sections (id, srd_version_id, slug, name, source_file, heading_path, body)
		VALUES ($1, (SELECT id FROM srd_versions WHERE version = $2), $3, $4, $5, $6, $7)
		ON CONFLICT (srd_version_id, slug) DO UPDATE SET
			name = EXCLUDED.name, source_file = EXCLUDED.source_file,
			heading_path = EXCLUDED.heading_path, body = EXCLUDED.body, updated_at = NOW()`,
		r.ID, r.SRDVersion, r.Slug, r.Name, r.SourceFile, hpJSON, r.Body)
	return err
}

// --- Class Features ---

func (s *Store) GetClassFeaturesByClass(ctx context.Context, srdVersion, classSlug string) ([]domain.ClassFeature, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT cf.id, cf.slug, sv.version, cf.name, cf.class_slug, cf.subclass_slug, cf.level, cf.description
		FROM class_features cf JOIN srd_versions sv ON cf.srd_version_id = sv.id
		WHERE sv.version = $1 AND cf.class_slug = $2 ORDER BY cf.level, cf.name`, srdVersion, classSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []domain.ClassFeature
	for rows.Next() {
		var f domain.ClassFeature
		rows.Scan(&f.ID, &f.Slug, &f.SRDVersion, &f.Name, &f.ClassSlug, &f.SubclassSlug, &f.Level, &f.Description)
		features = append(features, f)
	}
	return features, rows.Err()
}

func (s *Store) GetClassFeaturesByClassAndLevel(ctx context.Context, srdVersion, classSlug string, level int) ([]domain.ClassFeature, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT cf.id, cf.slug, sv.version, cf.name, cf.class_slug, cf.subclass_slug, cf.level, cf.description
		FROM class_features cf JOIN srd_versions sv ON cf.srd_version_id = sv.id
		WHERE sv.version = $1 AND cf.class_slug = $2 AND cf.level = $3 ORDER BY cf.name`, srdVersion, classSlug, level)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var features []domain.ClassFeature
	for rows.Next() {
		var f domain.ClassFeature
		rows.Scan(&f.ID, &f.Slug, &f.SRDVersion, &f.Name, &f.ClassSlug, &f.SubclassSlug, &f.Level, &f.Description)
		features = append(features, f)
	}
	return features, rows.Err()
}

func (s *Store) UpsertClassFeature(ctx context.Context, f *domain.ClassFeature) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO class_features (id, srd_version_id, slug, name, class_slug, subclass_slug, level, description)
		VALUES ($1, (SELECT id FROM srd_versions WHERE version = $2), $3, $4, $5, $6, $7, $8)
		ON CONFLICT (srd_version_id, slug) DO UPDATE SET
			name = EXCLUDED.name, class_slug = EXCLUDED.class_slug,
			subclass_slug = EXCLUDED.subclass_slug, level = EXCLUDED.level,
			description = EXCLUDED.description, updated_at = NOW()`,
		f.ID, f.SRDVersion, f.Slug, f.Name, f.ClassSlug, f.SubclassSlug, f.Level, f.Description)
	return err
}

func (s *Store) UpsertClassFeatures(ctx context.Context, features []domain.ClassFeature) error {
	for i := range features {
		if err := s.UpsertClassFeature(ctx, &features[i]); err != nil {
			return err
		}
	}
	return nil
}

// --- Class Level Tables ---

func (s *Store) GetClassLevelTableByClass(ctx context.Context, srdVersion, classSlug string) ([]domain.ClassLevelTableRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT clt.id, clt.class_slug, clt.level, clt.proficiency_bonus, clt.features_unlocked, clt.other_columns
		FROM class_level_tables clt JOIN srd_versions sv ON clt.srd_version_id = sv.id
		WHERE sv.version = $1 AND clt.class_slug = $2 ORDER BY clt.level`, srdVersion, classSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.ClassLevelTableRow
	for rows.Next() {
		var r domain.ClassLevelTableRow
		var fJSON, oJSON []byte
		rows.Scan(&r.ID, &r.ClassSlug, &r.Level, &r.ProficiencyBonus, &fJSON, &oJSON)
		json.Unmarshal(fJSON, &r.FeaturesUnlocked)
		json.Unmarshal(oJSON, &r.OtherColumns)
		result = append(result, r)
	}
	return result, rows.Err()
}

func (s *Store) GetClassLevelTableByClassAndLevel(ctx context.Context, srdVersion, classSlug string, level int) (*domain.ClassLevelTableRow, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT clt.id, clt.class_slug, clt.level, clt.proficiency_bonus, clt.features_unlocked, clt.other_columns
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

func (s *Store) UpsertClassLevelTableRow(ctx context.Context, r *domain.ClassLevelTableRow) error {
	fJSON, _ := json.Marshal(r.FeaturesUnlocked)
	oJSON, _ := json.Marshal(r.OtherColumns)
	_, err := s.pool.Exec(ctx, `
		INSERT INTO class_level_tables (id, srd_version_id, class_slug, level, proficiency_bonus, features_unlocked, other_columns)
		VALUES ($1, (SELECT id FROM srd_versions WHERE version = $2), $3, $4, $5, $6, $7)
		ON CONFLICT (srd_version_id, class_slug, level) DO UPDATE SET
			proficiency_bonus = EXCLUDED.proficiency_bonus,
			features_unlocked = EXCLUDED.features_unlocked,
			other_columns = EXCLUDED.other_columns, updated_at = NOW()`,
		r.ID, r.SRDVersion, r.ClassSlug, r.Level, r.ProficiencyBonus, fJSON, oJSON)
	return err
}

func (s *Store) UpsertClassLevelTableRows(ctx context.Context, rows []domain.ClassLevelTableRow) error {
	for i := range rows {
		if err := s.UpsertClassLevelTableRow(ctx, &rows[i]); err != nil {
			return err
		}
	}
	return nil
}
