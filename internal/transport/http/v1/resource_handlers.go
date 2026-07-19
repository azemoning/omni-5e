package v1

import (
	"github.com/gofiber/fiber/v3"
	"github.com/azemoning/omni-5e/internal/domain"
	"github.com/azemoning/omni-5e/internal/transport/http/v1/dto"
)

// --- Species ---

func (r *Router) listSpecies(c fiber.Ctx) error {
	filter := domain.SpeciesFilter{
		ListParams: parseListParams(c),
		Size:       c.Query("size"),
	}
	page, err := r.service.ListSpecies(c.Context(), filter)
	if err != nil {
		return errResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	species := page.Items.([]domain.Species)
	resp := make([]dto.SpeciesResponse, 0, len(species))
	for _, s := range species {
		resp = append(resp, speciesToDTO(&s))
	}
	return c.JSON(dto.ResponseEnvelope{
		Data: resp,
		Meta: dto.ResponseMeta{SRDVersion: page.SRDVersion, Total: page.Total, Limit: page.Limit, Offset: page.Offset},
		Links: buildLinks(c, page.Limit, page.Offset, page.Total),
	})
}

func (r *Router) getSpecies(c fiber.Ctx) error {
	sp, err := r.service.GetSpecies(c.Context(), getSRDVersion(c), c.Params("slug"))
	if err != nil {
		return errResponse(c, fiber.StatusNotFound, "species not found")
	}
	return c.JSON(dto.SingleResponse{Data: speciesToDTO(sp)})
}

func speciesToDTO(s *domain.Species) dto.SpeciesResponse {
	traits := make([]dto.Trait, len(s.Traits))
	for i, t := range s.Traits {
		traits[i] = dto.Trait{Name: t.Name, Description: t.Description}
	}
	return dto.SpeciesResponse{
		Slug: s.Slug, URL: "/api/v1/species/" + s.Slug, SRDVersion: s.SRDVersion,
		Name: s.Name, Size: s.Size, Speed: s.Speed, Traits: traits, Description: s.Description,
	}
}

// --- Backgrounds ---

func (r *Router) listBackgrounds(c fiber.Ctx) error {
	params := parseListParams(c)
	page, err := r.service.ListBackgrounds(c.Context(), params)
	if err != nil {
		return errResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	bgs := page.Items.([]domain.Background)
	resp := make([]dto.BackgroundResponse, 0, len(bgs))
	for _, b := range bgs {
		resp = append(resp, backgroundToDTO(&b))
	}
	return c.JSON(dto.ResponseEnvelope{
		Data: resp,
		Meta: dto.ResponseMeta{SRDVersion: page.SRDVersion, Total: page.Total, Limit: page.Limit, Offset: page.Offset},
		Links: buildLinks(c, page.Limit, page.Offset, page.Total),
	})
}

func (r *Router) getBackground(c fiber.Ctx) error {
	bg, err := r.service.GetBackground(c.Context(), getSRDVersion(c), c.Params("slug"))
	if err != nil {
		return errResponse(c, fiber.StatusNotFound, "background not found")
	}
	return c.JSON(dto.SingleResponse{Data: backgroundToDTO(bg)})
}

func backgroundToDTO(b *domain.Background) dto.BackgroundResponse {
	return dto.BackgroundResponse{
		Slug: b.Slug, URL: "/api/v1/backgrounds/" + b.Slug, SRDVersion: b.SRDVersion,
		Name: b.Name, AbilityScoreOptions: b.AbilityScoreOptions,
		SkillProficiencies: b.SkillProficiencies, GrantedFeatSlug: b.GrantedFeatSlug,
		Equipment: b.Equipment, Description: b.Description,
	}
}

// --- Feats ---

func (r *Router) listFeats(c fiber.Ctx) error {
	filter := domain.FeatFilter{ListParams: parseListParams(c), Category: c.Query("category")}
	page, err := r.service.ListFeats(c.Context(), filter)
	if err != nil {
		return errResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	feats := page.Items.([]domain.Feat)
	resp := make([]dto.FeatResponse, 0, len(feats))
	for _, f := range feats {
		resp = append(resp, featToDTO(&f))
	}
	return c.JSON(dto.ResponseEnvelope{
		Data: resp,
		Meta: dto.ResponseMeta{SRDVersion: page.SRDVersion, Total: page.Total, Limit: page.Limit, Offset: page.Offset},
		Links: buildLinks(c, page.Limit, page.Offset, page.Total),
	})
}

func (r *Router) getFeat(c fiber.Ctx) error {
	feat, err := r.service.GetFeat(c.Context(), getSRDVersion(c), c.Params("slug"))
	if err != nil {
		return errResponse(c, fiber.StatusNotFound, "feat not found")
	}
	return c.JSON(dto.SingleResponse{Data: featToDTO(feat)})
}

func featToDTO(f *domain.Feat) dto.FeatResponse {
	return dto.FeatResponse{
		Slug: f.Slug, URL: "/api/v1/feats/" + f.Slug, SRDVersion: f.SRDVersion,
		Name: f.Name, Category: f.Category, Prerequisite: f.Prerequisite,
		Description: f.Description, Repeatable: f.Repeatable,
	}
}

// --- Equipment ---

func (r *Router) listEquipment(c fiber.Ctx) error {
	filter := domain.EquipmentFilter{ListParams: parseListParams(c), Category: c.Query("category")}
	page, err := r.service.ListEquipment(c.Context(), filter)
	if err != nil {
		return errResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	items := page.Items.([]domain.Equipment)
	resp := make([]dto.EquipmentResponse, 0, len(items))
	for _, e := range items {
		resp = append(resp, equipmentToDTO(&e))
	}
	return c.JSON(dto.ResponseEnvelope{
		Data: resp,
		Meta: dto.ResponseMeta{SRDVersion: page.SRDVersion, Total: page.Total, Limit: page.Limit, Offset: page.Offset},
		Links: buildLinks(c, page.Limit, page.Offset, page.Total),
	})
}

func (r *Router) getEquipment(c fiber.Ctx) error {
	eq, err := r.service.GetEquipment(c.Context(), getSRDVersion(c), c.Params("slug"))
	if err != nil {
		return errResponse(c, fiber.StatusNotFound, "equipment not found")
	}
	return c.JSON(dto.SingleResponse{Data: equipmentToDTO(eq)})
}

func equipmentToDTO(e *domain.Equipment) dto.EquipmentResponse {
	return dto.EquipmentResponse{
		Slug: e.Slug, URL: "/api/v1/equipment/" + e.Slug, SRDVersion: e.SRDVersion,
		Name: e.Name, Category: e.Category, Cost: e.Cost, Weight: e.Weight,
		Properties: e.Properties, Description: e.Description,
	}
}

// --- Magic Items ---

func (r *Router) listMagicItems(c fiber.Ctx) error {
	filter := domain.MagicItemFilter{ListParams: parseListParams(c), Rarity: c.Query("rarity")}
	if v := c.Query("attunement"); v != "" {
		b := v == "true"
		filter.Attunement = &b
	}
	page, err := r.service.ListMagicItems(c.Context(), filter)
	if err != nil {
		return errResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	items := page.Items.([]domain.MagicItem)
	resp := make([]dto.MagicItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, magicItemToDTO(&item))
	}
	return c.JSON(dto.ResponseEnvelope{
		Data: resp,
		Meta: dto.ResponseMeta{SRDVersion: page.SRDVersion, Total: page.Total, Limit: page.Limit, Offset: page.Offset},
		Links: buildLinks(c, page.Limit, page.Offset, page.Total),
	})
}

func (r *Router) getMagicItem(c fiber.Ctx) error {
	item, err := r.service.GetMagicItem(c.Context(), getSRDVersion(c), c.Params("slug"))
	if err != nil {
		return errResponse(c, fiber.StatusNotFound, "magic item not found")
	}
	return c.JSON(dto.SingleResponse{Data: magicItemToDTO(item)})
}

func magicItemToDTO(m *domain.MagicItem) dto.MagicItemResponse {
	return dto.MagicItemResponse{
		Slug: m.Slug, URL: "/api/v1/magic-items/" + m.Slug, SRDVersion: m.SRDVersion,
		Name: m.Name, Rarity: m.Rarity, RequiresAttunement: m.RequiresAttunement,
		Type: m.Type, Description: m.Description,
	}
}

// --- Conditions ---

func (r *Router) listConditions(c fiber.Ctx) error {
	params := parseListParams(c)
	page, err := r.service.ListConditions(c.Context(), params)
	if err != nil {
		return errResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	conds := page.Items.([]domain.Condition)
	resp := make([]dto.ConditionResponse, 0, len(conds))
	for _, cond := range conds {
		resp = append(resp, conditionToDTO(&cond))
	}
	return c.JSON(dto.ResponseEnvelope{
		Data: resp,
		Meta: dto.ResponseMeta{SRDVersion: page.SRDVersion, Total: page.Total, Limit: page.Limit, Offset: page.Offset},
		Links: buildLinks(c, page.Limit, page.Offset, page.Total),
	})
}

func (r *Router) getCondition(c fiber.Ctx) error {
	cond, err := r.service.GetCondition(c.Context(), getSRDVersion(c), c.Params("slug"))
	if err != nil {
		return errResponse(c, fiber.StatusNotFound, "condition not found")
	}
	return c.JSON(dto.SingleResponse{Data: conditionToDTO(cond)})
}

func conditionToDTO(c *domain.Condition) dto.ConditionResponse {
	return dto.ConditionResponse{
		Slug: c.Slug, URL: "/api/v1/conditions/" + c.Slug, SRDVersion: c.SRDVersion,
		Name: c.Name, Description: c.Description,
	}
}

// --- Glossary Terms ---

func (r *Router) listGlossaryTerms(c fiber.Ctx) error {
	filter := domain.GlossaryFilter{ListParams: parseListParams(c), Category: c.Query("category")}
	page, err := r.service.ListGlossaryTerms(c.Context(), filter)
	if err != nil {
		return errResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	terms := page.Items.([]domain.GlossaryTerm)
	resp := make([]dto.GlossaryTermResponse, 0, len(terms))
	for _, t := range terms {
		resp = append(resp, glossaryTermToDTO(&t))
	}
	return c.JSON(dto.ResponseEnvelope{
		Data: resp,
		Meta: dto.ResponseMeta{SRDVersion: page.SRDVersion, Total: page.Total, Limit: page.Limit, Offset: page.Offset},
		Links: buildLinks(c, page.Limit, page.Offset, page.Total),
	})
}

func (r *Router) getGlossaryTerm(c fiber.Ctx) error {
	term, err := r.service.GetGlossaryTerm(c.Context(), getSRDVersion(c), c.Params("slug"))
	if err != nil {
		return errResponse(c, fiber.StatusNotFound, "glossary term not found")
	}
	return c.JSON(dto.SingleResponse{Data: glossaryTermToDTO(term)})
}

func glossaryTermToDTO(t *domain.GlossaryTerm) dto.GlossaryTermResponse {
	return dto.GlossaryTermResponse{
		Slug: t.Slug, URL: "/api/v1/glossary/" + t.Slug, SRDVersion: t.SRDVersion,
		Name: t.Name, Category: t.Category, Definition: t.Definition,
	}
}

// --- Rule Sections ---

func (r *Router) listRuleSections(c fiber.Ctx) error {
	filter := domain.RuleSectionFilter{ListParams: parseListParams(c), SourceFile: c.Query("source_file")}
	page, err := r.service.ListRuleSections(c.Context(), filter)
	if err != nil {
		return errResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	sections := page.Items.([]domain.RuleSection)
	resp := make([]dto.RuleSectionResponse, 0, len(sections))
	for _, s := range sections {
		resp = append(resp, ruleSectionToDTO(&s))
	}
	return c.JSON(dto.ResponseEnvelope{
		Data: resp,
		Meta: dto.ResponseMeta{SRDVersion: page.SRDVersion, Total: page.Total, Limit: page.Limit, Offset: page.Offset},
		Links: buildLinks(c, page.Limit, page.Offset, page.Total),
	})
}

func (r *Router) getRuleSection(c fiber.Ctx) error {
	section, err := r.service.GetRuleSection(c.Context(), getSRDVersion(c), c.Params("slug"))
	if err != nil {
		return errResponse(c, fiber.StatusNotFound, "rule section not found")
	}
	return c.JSON(dto.SingleResponse{Data: ruleSectionToDTO(section)})
}

func ruleSectionToDTO(s *domain.RuleSection) dto.RuleSectionResponse {
	return dto.RuleSectionResponse{
		Slug: s.Slug, URL: "/api/v1/rules/" + s.Slug, SRDVersion: s.SRDVersion,
		Name: s.Name, SourceFile: s.SourceFile, HeadingPath: s.HeadingPath, Body: s.Body,
	}
}
