package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/azemoning/omni-5e/internal/domain"
)

// --- Monsters ---

func (s *Store) GetMonster(ctx context.Context, srdVersion, slug string) (*domain.Monster, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT m.id, m.slug, sv.version, m.name, m.size, m.type, m.alignment,
		       m.ac_value, m.ac_source, m.hp_avg, m.hp_formula, m.cr, m.xp, m.category,
		       m.speed, m.ability_scores, m.saving_throws, m.skills,
		       m.damage_resistances, m.damage_immunities, m.damage_vulnerabilities,
		       m.condition_immunities, m.senses, m.languages, m.environment,
		       m.traits, m.actions, m.bonus_actions, m.reactions, m.legendary_actions
		FROM monsters m JOIN srd_versions sv ON m.srd_version_id = sv.id
		WHERE sv.version = $1 AND m.slug = $2`, srdVersion, slug)

	m := &domain.Monster{}
	var speedJSON, abilityJSON, saveJSON, skillsJSON, resJSON, immJSON, vulnJSON, condJSON, sensesJSON, langJSON, envJSON []byte
	var traitsJSON, actionsJSON, bonusJSON, reactionsJSON, legendaryJSON []byte

	err := row.Scan(&m.ID, &m.Slug, &m.SRDVersion, &m.Name, &m.Size, &m.Type, &m.Alignment,
		&m.AC.Value, &m.AC.Source, &m.HP.Average, &m.HP.Formula, &m.CR, &m.XP, &m.Category,
		&speedJSON, &abilityJSON, &saveJSON, &skillsJSON, &resJSON, &immJSON, &vulnJSON, &condJSON, &sensesJSON, &langJSON, &envJSON,
		&traitsJSON, &actionsJSON, &bonusJSON, &reactionsJSON, &legendaryJSON)
	if err != nil {
		return nil, fmt.Errorf("monster %s/%s: %w", srdVersion, slug, err)
	}
	json.Unmarshal(speedJSON, &m.Speed)
	json.Unmarshal(abilityJSON, &m.AbilityScores)
	json.Unmarshal(saveJSON, &m.SavingThrows)
	json.Unmarshal(skillsJSON, &m.Skills)
	json.Unmarshal(resJSON, &m.DamageResistances)
	json.Unmarshal(immJSON, &m.DamageImmunities)
	json.Unmarshal(vulnJSON, &m.DamageVulnerabilities)
	json.Unmarshal(condJSON, &m.ConditionImmunities)
	json.Unmarshal(sensesJSON, &m.Senses)
	json.Unmarshal(langJSON, &m.Languages)
	json.Unmarshal(envJSON, &m.Environment)
	json.Unmarshal(traitsJSON, &m.Traits)
	json.Unmarshal(actionsJSON, &m.Actions)
	json.Unmarshal(bonusJSON, &m.BonusActions)
	json.Unmarshal(reactionsJSON, &m.Reactions)
	json.Unmarshal(legendaryJSON, &m.LegendaryActions)
	return m, nil
}

func (s *Store) ListMonsters(ctx context.Context, filter domain.MonsterFilter) (*domain.Page, error) {
	where, args, idx := "WHERE sv.version = $1", []any{filter.SRDVersion}, 2
	if filter.CRMin != nil {
		where += fmt.Sprintf(" AND m.cr >= $%d", idx)
		args, idx = append(args, *filter.CRMin), idx+1
	}
	if filter.CRMax != nil {
		where += fmt.Sprintf(" AND m.cr <= $%d", idx)
		args, idx = append(args, *filter.CRMax), idx+1
	}
	if filter.Type != "" {
		where += fmt.Sprintf(" AND m.type = $%d", idx)
		args, idx = append(args, filter.Type), idx+1
	}
	if filter.Size != "" {
		where += fmt.Sprintf(" AND m.size = $%d", idx)
		args, idx = append(args, filter.Size), idx+1
	}
	if filter.Category != "" {
		where += fmt.Sprintf(" AND m.category = $%d", idx)
		args, idx = append(args, filter.Category), idx+1
	}
	if filter.Query != "" {
		where += fmt.Sprintf(" AND m.search @@ plainto_tsquery('english', $%d)", idx)
		args, idx = append(args, filter.Query), idx+1
	}

	var total int
	s.pool.QueryRow(ctx, fmt.Sprintf("SELECT COUNT(*) FROM monsters m JOIN srd_versions sv ON m.srd_version_id = sv.id %s", where), args...).Scan(&total)

	rows, err := s.pool.Query(ctx, fmt.Sprintf(`
		SELECT m.id, m.slug, sv.version, m.name, m.size, m.type, m.alignment,
		       m.ac_value, m.ac_source, m.hp_avg, m.hp_formula, m.cr, m.xp, m.category
		FROM monsters m JOIN srd_versions sv ON m.srd_version_id = sv.id
		%s ORDER BY m.name LIMIT $%d OFFSET $%d`, where, idx, idx+1), append(args, filter.Limit, filter.Offset)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	monsters := []domain.Monster{}
	for rows.Next() {
		var m domain.Monster
		rows.Scan(&m.ID, &m.Slug, &m.SRDVersion, &m.Name, &m.Size, &m.Type, &m.Alignment,
			&m.AC.Value, &m.AC.Source, &m.HP.Average, &m.HP.Formula, &m.CR, &m.XP, &m.Category)
		monsters = append(monsters, m)
	}
	return &domain.Page{Items: monsters, Total: total, Limit: filter.Limit, Offset: filter.Offset, SRDVersion: filter.SRDVersion}, rows.Err()
}

func (s *Store) UpsertMonster(ctx context.Context, m *domain.Monster) error {
	speedJSON, _ := json.Marshal(m.Speed)
	abilityJSON, _ := json.Marshal(m.AbilityScores)
	saveJSON, _ := json.Marshal(m.SavingThrows)
	skillsJSON, _ := json.Marshal(m.Skills)
	resJSON, _ := json.Marshal(m.DamageResistances)
	immJSON, _ := json.Marshal(m.DamageImmunities)
	vulnJSON, _ := json.Marshal(m.DamageVulnerabilities)
	condJSON, _ := json.Marshal(m.ConditionImmunities)
	sensesJSON, _ := json.Marshal(m.Senses)
	langJSON, _ := json.Marshal(m.Languages)
	envJSON, _ := json.Marshal(m.Environment)
	traitsJSON, _ := json.Marshal(m.Traits)
	actionsJSON, _ := json.Marshal(m.Actions)
	bonusJSON, _ := json.Marshal(m.BonusActions)
	reactionsJSON, _ := json.Marshal(m.Reactions)
	legendaryJSON, _ := json.Marshal(m.LegendaryActions)

	_, err := s.pool.Exec(ctx, `
		INSERT INTO monsters (id, srd_version_id, slug, name, size, type, alignment,
			ac_value, ac_source, hp_avg, hp_formula, cr, xp, category,
			speed, ability_scores, saving_throws, skills,
			damage_resistances, damage_immunities, damage_vulnerabilities,
			condition_immunities, senses, languages, environment,
			traits, actions, bonus_actions, reactions, legendary_actions)
		VALUES ($1, (SELECT id FROM srd_versions WHERE version = $2), $3, $4, $5, $6, $7,
		        $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25,
		        $26, $27, $28, $29, $30)
		ON CONFLICT (srd_version_id, slug) DO UPDATE SET
			name = EXCLUDED.name, size = EXCLUDED.size, type = EXCLUDED.type,
			alignment = EXCLUDED.alignment, ac_value = EXCLUDED.ac_value,
			ac_source = EXCLUDED.ac_source, hp_avg = EXCLUDED.hp_avg,
			hp_formula = EXCLUDED.hp_formula, cr = EXCLUDED.cr, xp = EXCLUDED.xp,
			category = EXCLUDED.category, speed = EXCLUDED.speed,
			ability_scores = EXCLUDED.ability_scores, saving_throws = EXCLUDED.saving_throws,
			skills = EXCLUDED.skills, damage_resistances = EXCLUDED.damage_resistances,
			damage_immunities = EXCLUDED.damage_immunities,
			damage_vulnerabilities = EXCLUDED.damage_vulnerabilities,
			condition_immunities = EXCLUDED.condition_immunities,
			senses = EXCLUDED.senses, languages = EXCLUDED.languages,
			environment = EXCLUDED.environment, traits = EXCLUDED.traits,
			actions = EXCLUDED.actions, bonus_actions = EXCLUDED.bonus_actions,
			reactions = EXCLUDED.reactions, legendary_actions = EXCLUDED.legendary_actions,
			updated_at = NOW()`,
		m.ID, m.SRDVersion, m.Slug, m.Name, m.Size, m.Type, m.Alignment,
		m.AC.Value, m.AC.Source, m.HP.Average, m.HP.Formula, m.CR, m.XP, m.Category,
		speedJSON, abilityJSON, saveJSON, skillsJSON, resJSON, immJSON, vulnJSON, condJSON, sensesJSON, langJSON, envJSON,
		traitsJSON, actionsJSON, bonusJSON, reactionsJSON, legendaryJSON)
	return err
}

func (s *Store) UpsertMonsters(ctx context.Context, monsters []domain.Monster) error {
	for i := range monsters {
		if err := s.UpsertMonster(ctx, &monsters[i]); err != nil {
			return err
		}
	}
	return nil
}
