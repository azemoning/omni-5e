package v1

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/azemoning/omni-5e/internal/domain"
	"github.com/azemoning/omni-5e/internal/transport/http/v1/dto"
)

// --- Classes ---

func (r *Router) listClasses(c fiber.Ctx) error {
	params := parseListParams(c)
	page, err := r.service.ListClasses(c.Context(), params)
	if err != nil {
		return errResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	classes := page.Items.([]domain.Class)
	resp := make([]dto.ClassResponse, 0, len(classes))
	for _, cl := range classes {
		resp = append(resp, classToDTO(&cl))
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

func (r *Router) getClass(c fiber.Ctx) error {
	slug := c.Params("slug")
	class, err := r.service.GetClass(c.Context(), getSRDVersion(c), slug)
	if err != nil {
		return errResponse(c, fiber.StatusNotFound, "class not found")
	}
	return c.JSON(dto.SingleResponse{Data: classToDTO(class)})
}

func (r *Router) getSubclasses(c fiber.Ctx) error {
	classSlug := c.Params("slug")
	subclasses, err := r.service.GetSubclasses(c.Context(), getSRDVersion(c), classSlug)
	if err != nil {
		return errResponse(c, fiber.StatusInternalServerError, err.Error())
	}

	resp := make([]dto.SubclassResponse, 0, len(subclasses))
	for _, sc := range subclasses {
		resp = append(resp, subclassToDTO(&sc))
	}
	return c.JSON(fiber.Map{"data": resp})
}

func (r *Router) getClassLevel(c fiber.Ctx) error {
	classSlug := c.Params("slug")
	level, err := strconv.Atoi(c.Params("level"))
	if err != nil || level < 1 || level > 20 {
		return errResponse(c, fiber.StatusBadRequest, "level must be 1-20")
	}

	row, err := r.service.GetClassLevelTable(c.Context(), getSRDVersion(c), classSlug, level)
	if err != nil {
		return errResponse(c, fiber.StatusNotFound, "level table row not found")
	}
	return c.JSON(dto.SingleResponse{Data: dto.ClassLevelTableRowResponse{
		ClassSlug:        row.ClassSlug,
		Level:            row.Level,
		ProficiencyBonus: row.ProficiencyBonus,
		FeaturesUnlocked: row.FeaturesUnlocked,
		OtherColumns:     row.OtherColumns,
	}})
}

func classToDTO(cl *domain.Class) dto.ClassResponse {
	subclasses := make([]dto.SubclassResponse, 0, len(cl.Subclasses))
	for _, sc := range cl.Subclasses {
		subclasses = append(subclasses, subclassToDTO(&sc))
	}
	return dto.ClassResponse{
		Slug:                     cl.Slug,
		URL:                      "/api/v1/classes/" + cl.Slug,
		SRDVersion:               cl.SRDVersion,
		Name:                     cl.Name,
		HitDie:                   cl.HitDie,
		PrimaryAbility:           cl.PrimaryAbility,
		SavingThrowProficiencies: cl.SavingThrowProficiencies,
		ArmorProficiencies:       cl.ArmorProficiencies,
		WeaponProficiencies:      cl.WeaponProficiencies,
		Description:              cl.Description,
		Subclasses:               subclasses,
	}
}

func subclassToDTO(sc *domain.Subclass) dto.SubclassResponse {
	return dto.SubclassResponse{
		Slug:        sc.Slug,
		URL:         "/api/v1/classes/" + sc.ClassSlug + "/subclasses/" + sc.Slug,
		SRDVersion:  sc.SRDVersion,
		Name:        sc.Name,
		ClassSlug:   sc.ClassSlug,
		Description: sc.Description,
	}
}
