package v1

import (
	"github.com/gofiber/fiber/v3"
	"github.com/azemoning/omni-5e/internal/service"
	"github.com/rs/zerolog"
)

// Router holds the v1 API route handlers.
type Router struct {
	service *service.Service
	log     zerolog.Logger
}

// NewRouter creates a new v1 API router.
func NewRouter(svc *service.Service, log zerolog.Logger) *Router {
	return &Router{service: svc, log: log}
}

// Register mounts all v1 routes on the given fiber.Group.
func (r *Router) Register(app fiber.Router) {
	// SRD version resolution middleware
	app.Use(r.srdVersionMiddleware())

	// Meta endpoints
	app.Get("/meta/srd-versions", r.listSRDVersions)
	app.Get("/license", r.getLicense)
	app.Get("/openapi.json", r.getOpenAPI)

	// Spells
	app.Get("/spells", r.listSpells)
	app.Get("/spells/:slug", r.getSpell)

	// Monsters
	app.Get("/monsters", r.listMonsters)
	app.Get("/monsters/:slug", r.getMonster)

	// Classes
	app.Get("/classes", r.listClasses)
	app.Get("/classes/:slug", r.getClass)
	app.Get("/classes/:slug/subclasses", r.getSubclasses)
	app.Get("/classes/:slug/levels/:level", r.getClassLevel)

	// Species (with /races redirect)
	app.Get("/species", r.listSpecies)
	app.Get("/species/:slug", r.getSpecies)
	app.Get("/races", func(c fiber.Ctx) error {
		return c.Redirect().To("/api/v1/species?" + c.Request().URI().QueryArgs().String())
	})
	app.Get("/races/:slug", func(c fiber.Ctx) error {
		return c.Redirect().To("/api/v1/species/" + c.Params("slug"))
	})

	// Backgrounds
	app.Get("/backgrounds", r.listBackgrounds)
	app.Get("/backgrounds/:slug", r.getBackground)

	// Feats
	app.Get("/feats", r.listFeats)
	app.Get("/feats/:slug", r.getFeat)

	// Equipment
	app.Get("/equipment", r.listEquipment)
	app.Get("/equipment/:slug", r.getEquipment)

	// Magic Items
	app.Get("/magic-items", r.listMagicItems)
	app.Get("/magic-items/:slug", r.getMagicItem)

	// Conditions
	app.Get("/conditions", r.listConditions)
	app.Get("/conditions/:slug", r.getCondition)

	// Glossary
	app.Get("/glossary", r.listGlossaryTerms)
	app.Get("/glossary/:slug", r.getGlossaryTerm)

	// Rules
	app.Get("/rules", r.listRuleSections)
	app.Get("/rules/:slug", r.getRuleSection)
}

// srdVersionMiddleware extracts srd_version from query param or header
// and stores it in locals.
func (r *Router) srdVersionMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		version := c.Query("srd_version")
		if version == "" {
			version = c.Get("X-SRD-Version")
		}
		c.Locals("srd_version", version)
		return c.Next()
	}
}

func getSRDVersion(c fiber.Ctx) string {
	v, _ := c.Locals("srd_version").(string)
	return v
}
