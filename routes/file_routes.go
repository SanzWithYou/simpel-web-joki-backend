package routes

import (
	"backend/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupFileRoutes(app *fiber.App, handler *handlers.FileHandler) {
	// File routes
	api := app.Group("/api")

	// Serve file
	api.Get("/files/*", handler.ServeFile)

	// Get file info
	api.Get("/files/info/*", handler.GetFileInfo)
}
