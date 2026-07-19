package v1

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/azemoning/omni-5e/internal/domain"
	"github.com/azemoning/omni-5e/internal/transport/http/v1/dto"
)

// --- Spells ---

func (r *Router) listSpells(c fiber.Ctx) error {
	filter := domain.SpellFilter{
		ListParams: parseListParams(c),
		Class:      c.Query("class"),
		School:     c.Query("school"),
		Query:      c.Query("q"),
	}
	if v := c.Query("level"); v != "" {
		level, err := strconv.Atoi(v)
		if err == nil {
			filter.Level = &level
		}
	}
	if v := c.Query("ritual"); v != "" {
		b := v == "true"
		filter.Ritual = &b
	}
	if v := c.Query("concentration"); v != "" {
		b := v == "true"
		filter.Concentration = &b
	}

	page, err := r.service.ListSpells(c.Context(), filter)
	if err != nil {
		return errResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	resp := make([]dto.SpellResponse, 0, len(page.Items.([]domain.Spell)))
	for _, s := range page.Items.([]domain.Spell) {
		resp = append(resp, spellToDTO(&s))
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

func (r *Router) getSpell(c fiber.Ctx) error {
	slug := c.Params("slug")
	spell, err := r.service.GetSpell(c.Context(), getSRDVersion(c), slug)
	if err != nil {
		return errResponse(c, fiber.StatusNotFound, "spell not found")
	}
	return c.JSON(dto.SingleResponse{Data: spellToDTO(spell)})
}

func spellToDTO(s *domain.Spell) dto.SpellResponse {
	return dto.SpellResponse{
		Slug:          s.Slug,
		URL:           "/api/v1/spells/" + s.Slug,
		SRDVersion:    s.SRDVersion,
		Name:          s.Name,
		Level:         s.Level,
		School:        s.School,
		CastingTime:   s.CastingTime,
		Range:         s.Range,
		Components: dto.Components{
			Verbal:         s.Components.Verbal,
			Somatic:        s.Components.Somatic,
			Material:       s.Components.Material,
			MaterialDetail: s.Components.MaterialDetail,
		},
		Duration:       s.Duration,
		Concentration:  s.Concentration,
		Ritual:         s.Ritual,
		Description:    s.Description,
		AtHigherLevels: s.AtHigherLevels,
		ClassSlugs:     s.ClassSlugs,
	}
}
