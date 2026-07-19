package v1

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/azemoning/omni-5e/internal/domain"
	"github.com/azemoning/omni-5e/internal/transport/http/v1/dto"
)

// --- Meta endpoints ---

func (r *Router) listSRDVersions(c fiber.Ctx) error {
	versions, err := r.service.ListSRDVersions(c.Context())
	if err != nil {
		return errResponse(c, fiber.StatusInternalServerError, err.Error())
	}
	resp := make([]dto.SRDVersionResponse, len(versions))
	for i, v := range versions {
		resp[i] = dto.SRDVersionResponse{
			Version: v.Version, ReleaseDate: v.ReleaseDate,
			SourceURL: v.SourceURL, License: v.License, IsDefault: v.IsDefault,
		}
	}
	return c.JSON(fiber.Map{"data": resp})
}

func (r *Router) getLicense(c fiber.Ctx) error {
	// Returns CC BY 4.0 attribution for Wizards of the Coast
	return c.JSON(dto.LicenseResponse{
		License:   "CC-BY-4.0",
		SourceURL: "https://github.com/downfallx/dnd-5e-srd-markdown",
		Version:   "5.2.1",
		Text:      "This work includes material taken from the System Reference Document 5.2.1 by Wizards of the Coast LLC, available at https://www.dndbeyond.com/resources/1781-system-reference-document-5-2-1. The SRD 5.2.1 is licensed under the Creative Commons Attribution 4.0 International License, available at https://creativecommons.org/licenses/by/4.0/legalcode.",
	})
}

func (r *Router) getOpenAPI(c fiber.Ctx) error {
	// TODO: serve actual OpenAPI spec
	return c.JSON(fiber.Map{"info": "OpenAPI spec placeholder"})
}

// --- Helpers ---

func parseListParams(c fiber.Ctx) domain.ListParams {
	params := domain.ListParams{
		Limit:      50,
		Offset:     0,
		SRDVersion: getSRDVersion(c),
		SortBy:     c.Query("sort_by"),
		SortOrder:  c.Query("sort_order"),
	}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			params.Limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			params.Offset = n
		}
	}
	return params
}

func buildLinks(c fiber.Ctx, limit, offset, total int) dto.ResponseLinks {
	self := fmt.Sprintf("%s?limit=%d&offset=%d", c.Path(), limit, offset)
	var next, prev *string
	if offset+limit < total {
		n := fmt.Sprintf("%s?limit=%d&offset=%d", c.Path(), limit, offset+limit)
		next = &n
	}
	if offset > 0 {
		p := fmt.Sprintf("%s?limit=%d&offset=%d", c.Path(), limit, max(0, offset-limit))
		prev = &p
	}
	return dto.ResponseLinks{Self: self, Next: next, Prev: prev}
}

func errResponse(c fiber.Ctx, status int, detail string) error {
	return c.Status(status).JSON(dto.ErrorResponse{
		Type:   "about:blank",
		Title:  fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Status: status,
		Detail: detail,
	})
}
