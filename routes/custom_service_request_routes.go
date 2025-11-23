package routes

import (
	"backend/handlers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupCustomServiceRequestRoutes(app *fiber.App, handler *handlers.CustomServiceRequestHandler, rateLimitConfig middleware.RateLimitConfig) {
	customServiceRequest := app.Group("/api/custom-service-requests")

	customServiceRequest.Use(middleware.RateLimit(rateLimitConfig))

	customServiceRequest.Post("/", handler.CreateCustomServiceRequest)
	customServiceRequest.Delete("/:id", handler.DeleteCustomServiceRequest)
}
