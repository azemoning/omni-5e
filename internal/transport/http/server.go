package http

import (
	"crypto/md5"
	_ "embed"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/azemoning/omni-5e/internal/config"
	"github.com/azemoning/omni-5e/internal/service"
	"github.com/azemoning/omni-5e/internal/transport/http/middleware"
	v1 "github.com/azemoning/omni-5e/internal/transport/http/v1"
	"github.com/rs/zerolog"
)

//go:embed openapi.yaml
var openapiSpec []byte

// Server holds the Fiber app and its dependencies.
type Server struct {
	app     *fiber.App
	cfg     *config.Config
	log     zerolog.Logger
	service *service.Service
}

// NewServer creates a new Fiber-based HTTP server.
func NewServer(cfg *config.Config, log zerolog.Logger, svc *service.Service) *Server {
	app := fiber.New(fiber.Config{
		AppName:      "omni-5e",
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		BodyLimit:    cfg.Server.MaxRequestBody,
		ErrorHandler: errorHandler(log),
	})

	s := &Server{
		app:     app,
		cfg:     cfg,
		log:     log,
		service: svc,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.app.Use(recover.New())
	s.app.Use(requestid.New())
	s.app.Use(cors.New())
	s.app.Use(middleware.Logging(s.log))
	s.app.Use(middleware.CacheControl())
	s.app.Use(limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
	}))
}

func (s *Server) setupRoutes() {
	// Health/readiness
	s.app.Get("/healthz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	s.app.Get("/readyz", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ready"})
	})

	// OpenAPI spec
	s.app.Get("/openapi.json", func(c fiber.Ctx) error {
		c.Set("Content-Type", "application/yaml")
		return c.Send(openapiSpec)
	})

	// Swagger UI
	s.app.Get("/docs", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(swaggerUIHTML)
	})

	// API v1
	v1Router := v1.NewRouter(s.service, s.log)
	v1Router.Register(s.app.Group("/api/v1"))

	// Root redirect to docs
	s.app.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().To("/docs")
	})
}

// ETagFor generates an ETag from content.
func ETagFor(content []byte) string {
	h := md5.Sum(content)
	return fmt.Sprintf(`"%x"`, h)
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)
	s.log.Info().Str("addr", addr).Msg("starting HTTP server")
	return s.app.Listen(addr)
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}

func errorHandler(log zerolog.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		log.Error().
			Err(err).
			Int("status", code).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Msg("request error")

		return c.Status(code).JSON(fiber.Map{
			"type":   "about:blank",
			"title":  http.StatusText(code),
			"status": code,
			"detail": err.Error(),
		})
	}
}

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>omni-5e API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
  <script>
    SwaggerUIBundle({ url: "/openapi.json", dom_id: "#swagger-ui", deepLinking: true })
  </script>
</body>
</html>`
