package middleware

import (
	"backend/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Request logger
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Log request
		duration := time.Since(start)
		utils.LogRequest(c, duration, err)

		return err
	}
}

// Error handler
func ErrorHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Process request
		err := c.Next()

		if err != nil {
			// Log error
			utils.LogError(err)

			// Send error
			return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Internal server error", err)
		}

		return nil
	}
}
