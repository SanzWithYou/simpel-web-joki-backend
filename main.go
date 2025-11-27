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
	log.Println("🔌 Connecting to database...")
	database.ConnectDB()
	log.Println("✅ Database connected")

	// Migrate models
	log.Println("🔄 Running database migrations...")
	if err := database.DB.AutoMigrate(
		&models.Order{},
		&models.CustomServiceRequest{},
	); err != nil {
		log.Fatal("❌ Gagal melakukan migrate database:", err)
	}
	log.Println("✅ Database migrations completed")

	// Init Fiber
	app := fiber.New(fiber.Config{
		AppName:      "Simpel Web Joki API",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	// ⚠️ PENTING: Health check HARUS sebelum middleware!
	// Leapcell health check endpoints - TANPA middleware apapun
	app.Get("/kaithhealthcheck", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	app.Get("/kaithheathcheck", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status": "ok",
		})
	})

	// Health check standard
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
		})
	})

	// Setup CORS
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "http://localhost:5173"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length",
		MaxAge:           86400,
	}))

	// Middleware - SETELAH health check
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

	// Test email endpoint (untuk debugging)
	app.Get("/test-email", func(c *fiber.Ctx) error {
		log.Println("🧪 Testing email...")

		err := utils.SendNewOrderNotificationEmail(
			999,
			"test-user",
			"Test Joki",
			"https://example.com/proof.jpg",
		)

		if err != nil {
			log.Printf("❌ Email test failed: %v", err)
			return c.JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}

		log.Println("✅ Email test succeeded")
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Email sent successfully",
		})
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🔥 Server starting on port %s", port)
	log.Printf("🌍 CORS allowed origins: %s", allowedOrigins)

	if err := app.Listen("0.0.0.0:" + port); err != nil {
		log.Fatal("❌ Gagal menjalankan server:", err)
	}
}
