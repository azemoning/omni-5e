package middleware

import "github.com/gofiber/fiber/v3"

// Auth is a no-op passthrough middleware slot for future auth.
func Auth() fiber.Handler {
	return func(c fiber.Ctx) error {
		return c.Next()
	}
}
