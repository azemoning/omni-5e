package v1

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/azemoning/omni-5e/internal/domain"
	"github.com/azemoning/omni-5e/internal/transport/http/v1/dto"
)

// --- Monsters ---

func (r *Router) listMonsters(c fiber.Ctx) error {
	filter := domain.MonsterFilter{
		ListParams: parseListParams(c),
		Type:       c.Query("type"),
		Size:       c.Query("size"),
		Environment: c.Query("environment"),
		Category:   c.Query("category"),
		Query:      c.Query("q"),
	}
	if v := c.Query("cr_min"); v != "" {
		cr, err := strconv.ParseFloat(v, 64)
		if err == nil {
			filter.CRMin = &cr
		}
	}
	if v := c.Query("cr_max"); v != "" {
		cr, err := strconv.ParseFloat(v, 64)
		if err == nil {
			filter.CRMax = &cr
		}
	}

	page, err := r.service.ListMonsters(c.Context(), filter)
	if err != nil {
		return errResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	monsters := page.Items.([]domain.Monster)
	resp := make([]dto.MonsterResponse, 0, len(monsters))
	for _, m := range monsters {
		resp = append(resp, monsterToDTO(&m))
	}

	return c.JSON(dto.ResponseEnvelope{
		Data: resp,
		Meta: dto.ResponseMeta{
			SRDVersion: page.SRDVersion,
			Total:      page.Total,
			Limit:      page.Limit,
			Offset:     page.Offset,
		},
		Links: buildLinks(c, page.Limit, page.Offset, page.Total),
	})
}

func (r *Router) getMonster(c fiber.Ctx) error {
	slug := c.Params("slug")
	monster, err := r.service.GetMonster(c.Context(), getSRDVersion(c), slug)
	if err != nil {
		return errResponse(c, fiber.StatusNotFound, "monster not found")
	}
	return c.JSON(dto.SingleResponse{Data: monsterToDTO(monster)})
}

func monsterToDTO(m *domain.Monster) dto.MonsterResponse {
	traits := namedBlocksToDTO(m.Traits)
	actions := namedBlocksToDTO(m.Actions)
	bonusActions := namedBlocksToDTO(m.BonusActions)
	reactions := namedBlocksToDTO(m.Reactions)
	legendaryActions := namedBlocksToDTO(m.LegendaryActions)

	return dto.MonsterResponse{
		Slug:      m.Slug,
		URL:       "/api/v1/monsters/" + m.Slug,
		SRDVersion: m.SRDVersion,
		Name:      m.Name,
		Size:      m.Size,
		Type:      m.Type,
		Alignment: m.Alignment,
		AC:        dto.ACInfo{Value: m.AC.Value, Source: m.AC.Source},
		HP:        dto.HPInfo{Average: m.HP.Average, Formula: m.HP.Formula},
		Speed:     m.Speed,
		AbilityScores: dto.AbilityScores{
			STR: m.AbilityScores.STR, DEX: m.AbilityScores.DEX,
			CON: m.AbilityScores.CON, INT: m.AbilityScores.INT,
			WIS: m.AbilityScores.WIS, CHA: m.AbilityScores.CHA,
		},
		SavingThrows:          m.SavingThrows,
		Skills:                m.Skills,
		DamageResistances:     m.DamageResistances,
		DamageImmunities:      m.DamageImmunities,
		DamageVulnerabilities: m.DamageVulnerabilities,
		ConditionImmunities:   m.ConditionImmunities,
		Senses:                m.Senses,
		Languages:             m.Languages,
		CR:                    m.CR,
		XP:                    m.XP,
		Traits:                traits,
		Actions:               actions,
		BonusActions:          bonusActions,
		Reactions:             reactions,
		LegendaryActions:      legendaryActions,
		Environment:           m.Environment,
		Category:              m.Category,
	}
}

func namedBlocksToDTO(blocks []domain.NamedBlock) []dto.NamedBlock {
	if blocks == nil {
		return nil
	}
	result := make([]dto.NamedBlock, len(blocks))
	for i, b := range blocks {
		result[i] = dto.NamedBlock{Name: b.Name, Description: b.Description}
	}
	return result
}
