package database

import (
	"backend/models"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Koneksi DB
func ConnectDB() {
	// Load env
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ .env tidak ditemukan, menggunakan env sistem")
	}

	// Ambil env
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	sslMode := os.Getenv("DB_SSLMODE")

	// Format DSN untuk PostgreSQL
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, pass, name, sslMode,
	)

	// Connect DB
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Gagal koneksi ke database:", err)
	}

	fmt.Println("✅ Berhasil terkoneksi ke database!")

	// Migrasi model
	if err := DB.AutoMigrate(&models.Order{}); err != nil {
		log.Fatal("❌ Gagal migrate database:", err)
	}

	fmt.Println("🚀 Migrasi database selesai!")
}
