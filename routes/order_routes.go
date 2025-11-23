package routes

import (
	"backend/handlers"
	"backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func SetupOrderRoutes(
	app *fiber.App,
	handler *handlers.OrderHandler,
	createOrderLimiter middleware.RateLimitConfig,
) {
	api := app.Group("/api")

	// Order routes
	api.Post("/orders", middleware.RateLimit(createOrderLimiter), handler.CreateOrderWithFile)
	api.Get("/orders", handler.GetAllOrders)
	api.Get("/orders/:id", handler.GetOrder)
	api.Delete("/orders/:id", handler.DeleteOrder)
}
