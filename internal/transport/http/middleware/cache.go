package middleware

import "github.com/gofiber/fiber/v3"

// CacheControl adds Cache-Control and ETag headers to GET responses.
func CacheControl() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Only add caching headers to GET requests
		if c.Method() != "GET" {
			return c.Next()
		}

		err := c.Next()
		if err != nil {
			return err
		}

		// Reference data is near-static per SRD version — cache aggressively
		status := c.Response().StatusCode()
		if status == 200 {
			c.Set("Cache-Control", "public, max-age=3600, stale-while-revalidate=86400")
		}

		return nil
	}
}
