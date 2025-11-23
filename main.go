package main

import (
	"backend/database"
	"backend/handlers"
	"backend/middleware"
	"backend/models"
	"backend/routes"
	"backend/utils"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Generate key
	generateKey := flag.Bool("generate-key", false, "Generate AES Encryption Key")
	flag.Parse()

	if *generateKey {
		key, err := utils.GenerateAESKey()
		if err != nil {
			fmt.Println("❌ Error:", err)
			os.Exit(1)
		}

		fmt.Println("✅ Generated AES-256 Key:")
		fmt.Println(key)
		return
	}

	// Connect DB
	database.ConnectDB()

	// Migrate models
	if err := database.DB.AutoMigrate(
		&models.Order{},
		&models.CustomServiceRequest{},
	); err != nil {
		log.Fatal("❌ Gagal melakukan migrate database:", err)
	}

	// Init Fiber
	app := fiber.New(fiber.Config{
		AppName:      "Simpel Web Joki API",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	// Setup CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length",
		MaxAge:           86400,
	}))

	// Middleware
	app.Use(recover.New())
	app.Use(middleware.Logger())
	app.Use(middleware.ErrorHandler())

	// Rate limit
	createOrderConfig := middleware.GetRateLimitConfig("CREATE_ORDER", 2, 1*time.Minute)
	customServiceRequestConfig := middleware.GetRateLimitConfig("CUSTOM_SERVICE_REQUEST", 1, 5*time.Minute)

	// Handlers
	orderHandler := handlers.NewOrderHandler(database.DB)
	fileHandler := handlers.NewFileHandler()
	customServiceRequestHandler := handlers.NewCustomServiceRequestHandler(database.DB)

	// Routes
	routes.SetupOrderRoutes(app, orderHandler, createOrderConfig)
	routes.SetupFileRoutes(app, fileHandler)
	routes.SetupCustomServiceRequestRoutes(app, customServiceRequestHandler, customServiceRequestConfig)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return utils.SuccessResponse(c, fiber.StatusOK, "API is running", nil)
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🔥 Server berjalan di port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("❌ Gagal menjalankan server:", err)
	}
}
